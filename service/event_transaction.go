package service

import (
	"assist-tix/config"
	"assist-tix/database"
	"assist-tix/domain"
	"assist-tix/dto"
	"assist-tix/helper"
	"assist-tix/lib"
	"assist-tix/model"
	"assist-tix/repository"
	"bytes"
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	mrand "math/rand"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

type EventTransactionService interface {
	CreateEventTransaction(ctx context.Context, eventId, ticketCategoryId string, req dto.CreateEventTransaction) (res dto.EventTransactionResponse, err error)
	PaylabsVASnap(ctx *gin.Context) (err error)
	CallbackVASnap(ctx *gin.Context, req dto.SnapCallbackPaymentRequest) (err error)
}

type EventTransactionServiceImpl struct {
	DB                       *database.WrapDB
	Env                      *config.EnvironmentVariable
	EventRepo                repository.EventRepository
	EventSettingRepo         repository.EventSettingsRepository
	EventTicketCategoryRepo  repository.EventTicketCategoryRepository
	EventTransactionRepo     repository.EventTransactionRepository
	EventTransactionItemRepo repository.EventTransactionItemRepository
	EventSeatmapBookRepo     repository.EventSeatmapBookRepository
	VenueSectorRepo          repository.VenueSectorRepository
}

func NewEventTransactionService(
	db *database.WrapDB,
	env *config.EnvironmentVariable,
	eventRepo repository.EventRepository,
	eventSettingRepo repository.EventSettingsRepository,
	eventTicketCategoryRepo repository.EventTicketCategoryRepository,
	eventTransactionRepo repository.EventTransactionRepository,
	eventTransactionItemRepo repository.EventTransactionItemRepository,
	eventSeatmapBookRepo repository.EventSeatmapBookRepository,
	venueSectorRepo repository.VenueSectorRepository,
) EventTransactionService {
	return &EventTransactionServiceImpl{
		DB:                       db,
		Env:                      env,
		EventRepo:                eventRepo,
		EventSettingRepo:         eventSettingRepo,
		EventTicketCategoryRepo:  eventTicketCategoryRepo,
		EventTransactionRepo:     eventTransactionRepo,
		EventTransactionItemRepo: eventTransactionItemRepo,
		EventSeatmapBookRepo:     eventSeatmapBookRepo,
		VenueSectorRepo:          venueSectorRepo,
	}
}

func (s *EventTransactionServiceImpl) CreateEventTransaction(ctx context.Context, eventId, ticketCategoryId string, req dto.CreateEventTransaction) (res dto.EventTransactionResponse, err error) {
	log.Info().Str("eventId", eventId).Str("ticketCategoryId", ticketCategoryId).Str("paymentMethod", req.PaymentMethod).Msg("create event transaction")
	tx, err := s.DB.Postgres.Begin(ctx)
	if err != nil {
		return res, err
	}
	defer tx.Rollback(ctx)

	log.Info().Msg("validate event by id")
	event, err := s.EventRepo.FindById(ctx, tx, eventId)
	if err != nil {
		return
	}

	if event.IsSaleActive {
		now := time.Now()
		if now.After(event.EndSaleAt.Time) {
			err = &lib.ErrorEventSaleAlreadyOver
			return
		} else if !(now.After(event.StartSaleAt.Time) && now.Before(event.EndSaleAt.Time)) {
			err = &lib.ErrorEventSaleIsNotStartedYet
			return
		}
	} else {
		err = &lib.ErrorEventSaleIsPaused
		return
	}

	log.Info().Msg("find event settings by event id")
	settings, err := s.EventSettingRepo.FindByEventId(ctx, tx, eventId)
	if err != nil {
		return
	}

	log.Info().Msg("mapping event settings")
	eventSettings := lib.MapEventSettings(settings)
	log.Info().Interface("Settings", eventSettings).Msg("Event settings")

	log.Info().Str("eventId", eventId).Str("ticketCategoryId", ticketCategoryId).Msg("find ticket category by id and event id")
	ticketCategory, err := s.EventTicketCategoryRepo.FindByIdAndEventId(ctx, tx, eventId, ticketCategoryId)
	if err != nil {
		return
	}

	log.Info().Str("venueSectorId", ticketCategory.VenueSectorId).Msg("find venue by venue sector id")
	venueSector, err := s.VenueSectorRepo.FindById(ctx, tx, ticketCategory.VenueSectorId)
	if err != nil {
		return
	}

	buyCount := len(req.Items)
	log.Info().Int("count", buyCount).Int("MaxAdultTicketPerTransaction", eventSettings.MaxAdultTicketPerTransaction).Msg("buy items")
	if buyCount > eventSettings.MaxAdultTicketPerTransaction {
		err = &lib.ErrorPurchaseQuantityExceedTheLimit
		return
	}

	log.Info().Int("publicStock", ticketCategory.PublicStock).Msg("checking is user capable to buy by their buy count")
	if ticketCategory.PublicStock < 0 || buyCount > ticketCategory.PublicStock {
		err = &lib.ErrorTicketIsOutOfStock
		return
	}

	log.Info().Msg("update stock public ticket by ticket category id")
	err = s.EventTicketCategoryRepo.BuyPublicTicketById(ctx, tx, eventId, ticketCategoryId, buyCount)
	if err != nil {
		return
	}

	now := time.Now()
	expiryInvoice := now.Add(s.Env.Transaction.ExpirationDuration)
	invoiceNumber := helper.GenerateInvoiceNumber()
	log.Info().Str("InvoiceNumber", invoiceNumber).Msg("generated invoice number")

	transaction := model.EventTransaction{
		FullName:    req.FullName,
		Email:       req.FullName,
		PhoneNumber: req.PhoneNumber,

		InvoiceNumber: invoiceNumber,
		Status:        lib.PaymentStatusPending,

		PaymentMethod:    req.PaymentMethod,
		PaymentChannel:   lib.PaymentChannelPaylabs,
		PaymentExpiredAt: expiryInvoice,
	}

	if venueSector.HasSeatmap {
		log.Info().Msg("venueSector in ticket category has seatmap")
		var seatParams []domain.SeatmapParam
		for _, val := range req.Items {
			seatParams = append(seatParams, domain.SeatmapParam{
				SeatRow:    val.SeatRow,
				SeatColumn: val.SeatColumn,
			})
		}

		// Checking choosen seat is in available status
		log.Info().Msg("checking choosen seat is in available status")
		sectorSeatmap, sectorSeatmapErr := s.EventTicketCategoryRepo.FindSeatmapStatusByEventSectorId(ctx, tx, eventId, ticketCategory.VenueSectorId, seatParams)
		if sectorSeatmapErr != nil {
			return
		}

		for _, val := range req.Items {
			seat, ok := sectorSeatmap[helper.ConvertRowColumnKey(val.SeatRow, val.SeatColumn)]
			if !ok {
				err = &lib.ErrorBookedSeatNotFound
				return
			} else {
				switch seat.Status {
				case lib.SeatmapStatusUnavailable:
					err = &lib.ErrorSeatIsAlreadyBooked
					return
				case lib.SeatmapStatusDisable:
					err = &lib.ErrorFailedToBookSeat
					return
				}
			}
		}

		// Checking seat is already booked by try to insert
		log.Info().Msg("checking seat is already booked by try to insert")
		err = s.EventSeatmapBookRepo.CreateSeatBook(ctx, tx, eventId, ticketCategory.VenueSectorId, seatParams)
		if err != nil {
			return
		}
	}

	// TODO: Checking bulk garuda id
	if eventSettings.GarudaIdVerification {
		log.Info().Msg("validation garuda id")
		// Verify garuda id
	}

	// Calculate price
	transaction.TotalPrice = ticketCategory.Price * len(req.Items)
	taxPerTransaction := (eventSettings.TaxPercentage / 100) * float64(transaction.TotalPrice)
	transaction.TotalTax = int(taxPerTransaction)
	log.Info().Int("TotalPrice", transaction.TotalPrice).Float64("TaxaPerTransaction", taxPerTransaction).Int("TotalTax", transaction.TotalTax).Msg("calculate price")

	var totalAdminFee int
	if eventSettings.AdminFeePercentage > 0 {
		totalAdminFee = int(eventSettings.AdminFeePercentage/100) * transaction.TotalPrice
	} else {
		totalAdminFee = eventSettings.AdminFee
	}

	transaction.AdminFeePercentage = float32(eventSettings.AdminFeePercentage)
	transaction.TotalAdminFee = totalAdminFee
	log.Info().Int("TotalAdminFee", totalAdminFee).Float32("AdminFeePercentage", transaction.AdminFeePercentage).Msg("calculate admin fee")

	transaction.GrandTotal = transaction.TotalPrice + transaction.TotalTax + transaction.TotalAdminFee
	log.Info().Int("GrandTotal", transaction.GrandTotal).Msg("got grand total price")

	log.Info().Msg("create transaction to database")
	transactionRes, err := s.EventTransactionRepo.CreateTransaction(ctx, tx, transaction)
	if err != nil {
		return
	}

	transaction.ID = transactionRes.ID

	var transactionItems []model.EventTransactionItem
	for _, item := range req.Items {
		transactionItems = append(transactionItems, model.EventTransactionItem{
			TransactionID:         transaction.ID,
			TicketCategoryID:      ticketCategoryId,
			Quantity:              1,
			SeatRow:               item.SeatRow,
			SeatColumn:            item.SeatColumn,
			AdditionalInformation: sql.NullString{String: item.AdditionalInformation},
			TotalPrice:            ticketCategory.Price,
		})
	}
	log.Info().Str("transactionId", transaction.ID).Int("count", len(transactionItems)).Msg("create transaction item")

	// TODO: Add item name, email phone number
	log.Info().Msg("insert transaction item")
	err = s.EventTransactionItemRepo.CreateTransactionItems(ctx, tx, transactionItems)
	if err != nil {
		return
	}

	err = tx.Commit(ctx)
	if err != nil {
		return
	}

	res = dto.EventTransactionResponse{
		InvoiceNumber:      invoiceNumber,
		PaymentMethod:      req.PaymentMethod,
		TotalPrice:         transaction.TotalPrice,
		TaxPercentage:      transaction.TaxPercentage,
		TotalTax:           transaction.TotalTax,
		AdminFeePercentage: transaction.AdminFeePercentage,
		TotalAdminFee:      transaction.TotalAdminFee,
		GrandTotal:         transaction.GrandTotal,
		ExpiredAt:          transaction.PaymentExpiredAt,
		CreatedAt:          transaction.CreatedAt,
	}

	log.Info().Msg("success create transaction")

	return
}

// static Eventtransaction without any business logic
func (s *EventTransactionServiceImpl) PaylabsVASnap(ctx *gin.Context) (err error) {
	date := time.Now().Format("2006-01-02T15:04:05.999+07:00")
	merchantId := s.Env.Paylabs.AccountID[len(s.Env.Paylabs.AccountID)-6:]
	partnerServiceId := s.Env.Paylabs.AccountID[:8]
	idRequest := fmt.Sprintf("%d", mrand.Intn(9999999-1111)+1111)
	// Generate a random 20-digit customer number as a string
	var customerNo string
	for i := 0; i < 20; i++ {
		digit := mrand.Intn(10)
		customerNo += fmt.Sprintf("%d", digit)
	}
	privateKeyPEM := s.Env.Paylabs.PrivateKey // Private key in PEM format
	payload := dto.VirtualAccountSnapRequest{
		PartnerServiceID:    partnerServiceId,        // 8 characters
		CustomerNo:          customerNo,              // Fixed 20-digit value
		VirtualAccountNo:    customerNo + merchantId, // 28-digit composite value
		VirtualAccountName:  "john doe",              // Payer name
		VirtualAccountEmail: "john.doe@example.com",
		VirtualAccountPhone: "6281234567890", // Mobile phone number in Indonesian format
		TrxID:               idRequest,       // Merchant transaction number
		TotalAmount: dto.Amount{
			Value:    "10000.00", // Amount with 2 decimal
			Currency: "IDR",      // Fixed currency
		},
		AdditionalInfo: dto.AdditionalInfo{
			PaymentType: "MandiriVA", // Payment type
		},
		ExpiredDate: "2025-12-31T23:59:59+07:00", // ISO-8601 formatted expiration
	}

	log.Info().Msgf("Creating event transaction with ID: %s", idRequest)
	// VA

	jsonData, err := json.Marshal(payload)
	if err != nil {
		log.Error().Err(err)
		return
	}
	log.Info().Msgf("JSON Payload: %s", jsonData)

	// Hash the JSON body
	shaJson := sha256.Sum256(jsonData)

	signature := helper.GenerateSnapSignature(shaJson, date, privateKeyPEM)
	log.Info().Msgf("Payload: %x", shaJson)
	// Create HTTP headers
	headers := map[string]string{
		"X-TIMESTAMP":   date,
		"X-SIGNATURE":   signature,
		"X-PARTNER-ID":  merchantId,
		"X-EXTERNAL-ID": idRequest,
		"X-IP-ADDRESS":  ctx.ClientIP(),
		"Content-Type":  "application/json",
	}
	log.Info().Msgf("Headers: %v", headers)

	// Send HTTP request
	url := s.Env.Paylabs.BaseUrl + "/api/v1.0/transfer-va/create-va"
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		log.Error().Err(err)
		return
	}
	for key, value := range headers {
		req.Header.Set(key, value)
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Error().Err(err).Msg("Failed to send request to Paylabs")
		return
	}
	defer resp.Body.Close()

	// Decode response
	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		log.Error().Err(err)
		return
	}

	// Print response
	log.Info().Msgf("Response: %v", response)
	return
}

func (s *EventTransactionServiceImpl) CallbackVASnap(ctx *gin.Context, req dto.SnapCallbackPaymentRequest) (err error) {
	log.Info().Msg("Processing Paylabs VA snap callback")
	header := map[string]interface{}{}
	for key, value := range ctx.Request.Header {
		header[key] = value
	}
	log.Info().Msgf("Headers: %v", header)

	rawPayload := ctx.GetString("rawPayload")
	var buf bytes.Buffer
	json.Compact(&buf, []byte(rawPayload))
	log.Info().Msgf("Raw Payload: %s", buf.String())
	log.Info().Msgf("Request URL: %v", req)
	isValid := helper.IsValidPaylabsRequest(ctx, "/transfer-va/payment", buf.String(), s.Env.Paylabs.PublicKey)
	if !isValid {
		return errors.New("invalid signature")
	}
	return
}

func (s *EventTransactionServiceImpl) CallbackVA(ctx *gin.Context, req dto.PaylabsVACallbackRequest) (err error) {
	stringifyPayload, err := json.Marshal(req)
	if err != nil {
		log.Error().Err(err).Msg("Failed to marshal callback request")
		return
	}
	isValid := helper.IsValidPaylabsRequest(ctx, ctx.Request.URL.Path, string(stringifyPayload), s.Env.Paylabs.PublicKey)
	if !isValid {
		return errors.New("invalid signature")
	}
	return
}
