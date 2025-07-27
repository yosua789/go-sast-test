package service

import (
	"assist-tix/config"
	"assist-tix/database"
	"assist-tix/domain"
	"assist-tix/dto"
	"assist-tix/helper"
	"assist-tix/internal/job"
	"assist-tix/internal/usecase"
	"assist-tix/lib"
	"assist-tix/model"
	"assist-tix/repository"
	"bytes"
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	mrand "math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

type EventTransactionService interface {
	CreateEventTransaction(ctx *gin.Context, eventId, ticketCategoryId string, req dto.CreateEventTransaction) (res dto.EventTransactionResponse, err error)
	PaylabsVASnap(ctx *gin.Context) (err error)
	CallbackVASnap(ctx *gin.Context, req dto.SnapCallbackPaymentRequest) (err error)
}

type EventTransactionServiceImpl struct {
	DB                            *database.WrapDB
	Env                           *config.EnvironmentVariable
	EventRepo                     repository.EventRepository
	EventSettingRepo              repository.EventSettingsRepository
	EventTicketCategoryRepo       repository.EventTicketCategoryRepository
	EventTransactionRepo          repository.EventTransactionRepository
	EventTransactionItemRepo      repository.EventTransactionItemRepository
	EventSeatmapBookRepo          repository.EventSeatmapBookRepository
	EventTransactionGarudaIDRepo  repository.EventTransactionGarudaIDRepository
	EventOrderInformationBookRepo repository.EventOrderInformationBookRepository
	VenueSectorRepo               repository.VenueSectorRepository

	CheckStatusTransactionJob job.CheckStatusTransactionJob

	TransactionUseCase usecase.TransactionUsecase
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
	EventOrderInformationBookRepo repository.EventOrderInformationBookRepository,
	venueSectorRepo repository.VenueSectorRepository,
	eventTransactionGarudaIDRepo repository.EventTransactionGarudaIDRepository,

	checkStatusTransactionJob job.CheckStatusTransactionJob,

	transactionUseCase usecase.TransactionUsecase,
) EventTransactionService {
	return &EventTransactionServiceImpl{
		DB:                            db,
		Env:                           env,
		EventRepo:                     eventRepo,
		EventSettingRepo:              eventSettingRepo,
		EventTicketCategoryRepo:       eventTicketCategoryRepo,
		EventTransactionRepo:          eventTransactionRepo,
		EventTransactionItemRepo:      eventTransactionItemRepo,
		EventSeatmapBookRepo:          eventSeatmapBookRepo,
		EventOrderInformationBookRepo: EventOrderInformationBookRepo,
		VenueSectorRepo:               venueSectorRepo,
		EventTransactionGarudaIDRepo:  eventTransactionGarudaIDRepo,

		CheckStatusTransactionJob: checkStatusTransactionJob,

		TransactionUseCase: transactionUseCase,
	}
}

func (s *EventTransactionServiceImpl) CreateEventTransaction(ctx *gin.Context, eventId, ticketCategoryId string, req dto.CreateEventTransaction) (res dto.EventTransactionResponse, err error) {
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

	if event.PublishStatus != lib.EventPublishStatusPublished {
		err = &lib.ErrorEventNotFound
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

	log.Info().Str("eventId", eventId).Msg("validate email is booked in the event")
	orderInformationBookId, err := s.EventOrderInformationBookRepo.CreateOrderInformation(ctx, tx, eventId, req.Email, req.Fullname)
	if err != nil {
		return
	}

	log.Info().Interface("SettingsRaw", settings).Msg("mapping event settings")
	eventSettings := lib.MapEventSettings(settings)
	log.Info().Interface("Settings", eventSettings).Msg("Event settings")

	log.Info().Str("eventId", eventId).Str("ticketCategoryId", ticketCategoryId).Msg("find ticket category by id and event id")
	ticketCategory, err := s.EventTicketCategoryRepo.FindByIdAndEventId(ctx, tx, eventId, ticketCategoryId)
	if err != nil {
		return
	}

	log.Info().Str("venueSectorId", ticketCategory.VenueSectorId).Msg("find venue by venue sector id")
	venueSector, err := s.VenueSectorRepo.FindVenueSectorById(ctx, tx, ticketCategory.VenueSectorId)
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
		Fullname: req.Fullname, // request fullname ?
		Email:    req.Email,
		// PhoneNumber: req.PhoneNumber, // request phone number ?

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
					// err = &lib.ErrorFailedToBookSeat
					err = &lib.ErrorBookedSeatNotFound
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
		// Verify garuda id
		for i, item := range req.Items {

			// check internal database whether garuda id is hold
			_, garudaIdErr := s.EventTransactionGarudaIDRepo.GetEventGarudaID(ctx, tx, eventId, item.GarudaID)
			if garudaIdErr == nil {
				return res, &lib.ErrorGarudaIDAlreadyUsed
			}

			// When error isn't TixError
			var tixErr *lib.TIXError
			if !errors.As(garudaIdErr, &tixErr) {
				log.Error().Err(err).Msg("error validate hold garuda id")
				err = garudaIdErr
				return res, err
			}

			// When garuda id not found in event books
			if tixErr == &lib.ErrorGarudaIDNotFound {

				// Verify garuda id Validity  by external service
				externalResp, errExternal := helper.VerifyUserGarudaIDByID(s.Env.GarudaID.BaseUrl, item.GarudaID)
				if errExternal != nil {
					log.Error().Err(errExternal).Msg("failed to verify garuda id")
					err = &lib.ErrorGetGarudaID
					return
				}

				if externalResp != nil && !externalResp.Success {
					switch externalResp.ErrorCode {
					case 40401:
						err = &lib.ErrorGarudaIDNotFound
						return
					case 42205:
						err = &lib.ErrorGarudaIDBlacklisted
						return
					case 40909:
						err = &lib.ErrorGarudaIDInvalid
						return
					case 40910:
						err = &lib.ErrorGarudaIDRejected
						return
					case 50001:
						err = &lib.ErrorGetGarudaID
						return
					}
				}
				//  append garuda id to transaction item
				req.Items[i].GarudaID = item.GarudaID
				req.Items[i].FullName = externalResp.Data.Name
				req.Items[i].Email = externalResp.Data.Email
				req.Items[i].PhoneNumber = externalResp.Data.PhoneNumber

				log.Info().Interface("externalResp", externalResp).Msg("garuda id validation response")
			}
		}
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
	transactionRes, err := s.EventTransactionRepo.CreateTransaction(ctx, tx, eventId, ticketCategoryId, transaction)
	if err != nil {
		return
	}

	transaction.ID = transactionRes.ID
	transaction.CreatedAt = transactionRes.CreatedAt

	// Update order information book to set transactionId
	err = s.EventOrderInformationBookRepo.UpdateTransactionIdByID(ctx, tx, orderInformationBookId, transaction.ID)
	if err != nil {
		return
	}

	var transactionItems []model.EventTransactionItem
	for _, item := range req.Items {
		var garudaId sql.NullString = helper.ToSQLString(item.GarudaID)

		var fullName sql.NullString = helper.ToSQLString(item.FullName)
		var email sql.NullString = helper.ToSQLString(item.Email)
		var phoneNumber sql.NullString = helper.ToSQLString(item.PhoneNumber)

		transactionItems = append(transactionItems, model.EventTransactionItem{
			TransactionID: transaction.ID,
			// TicketCategoryID:      ticketCategoryId,

			Quantity: 1,

			SeatRow:    item.SeatRow,
			SeatColumn: item.SeatColumn,

			GarudaID:    garudaId,
			Fullname:    fullName,
			Email:       email,
			PhoneNumber: phoneNumber,

			AdditionalInformation: sql.NullString{String: item.AdditionalInformation},
			TotalPrice:            ticketCategory.Price,

			CreatedAt: transaction.CreatedAt,
		})
	}
	log.Info().Str("transactionId", transaction.ID).Int("count", len(transactionItems)).Msg("create transaction item")

	// TODO: Add item name, email phone number ->done on validating garuda id
	log.Info().Msg("insert transaction item")
	err = s.EventTransactionItemRepo.CreateTransactionItems(ctx, tx, transactionItems)
	if err != nil {
		return
	}

	//  VA SNAP Init
	expiredDate := expiryInvoice.Format("2006-01-02T15:04:05+07:00")
	date := time.Now().Format("2006-01-02T15:04:05.999+07:00")
	merchantId := s.Env.Paylabs.AccountID[len(s.Env.Paylabs.AccountID)-6:]
	partnerServiceId := s.Env.Paylabs.AccountID[:8]
	idRequest := transaction.ID
	// Generate a random 20-digit customer number as a string

	privateKeyPEM := s.Env.Paylabs.PrivateKey                     // Private key in PEM format
	totalPriceStr := strconv.Itoa(transaction.GrandTotal) + ".00" // Amount with 2 decimal
	payload := dto.VirtualAccountSnapRequest{
		PartnerServiceID:    partnerServiceId,                 // 8 characters
		CustomerNo:          transaction.ID[:20],              // Fixed 20-digit value
		VirtualAccountNo:    transaction.ID[:20] + merchantId, // 28-digit composite value
		VirtualAccountName:  req.Fullname,                     // Payer name
		VirtualAccountEmail: req.Email,
		// VirtualAccountPhone: req.Items[0].PhoneNumber,  // Mobile phone number in Indonesian format
		TrxID: transaction.InvoiceNumber, // Merchant transaction number
		TotalAmount: dto.Amount{
			Value:    totalPriceStr, // Amount with 2 decimal
			Currency: "IDR",         // Fixed currency
		},
		AdditionalInfo: dto.AdditionalInfo{
			PaymentType: req.PaymentMethod, // Payment type
		},

		ExpiredDate: expiredDate, // ISO-8601 formatted expiration
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
	paylabsReq, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		log.Error().Err(err)
		err = &lib.ErrorTransactionPaylabs
		return
	}
	for key, value := range headers {
		paylabsReq.Header.Set(key, value)
	}
	client := &http.Client{}
	resp, err := client.Do(paylabsReq)
	if err != nil {
		log.Error().Err(err).Msg("Failed to send request to Paylabs")
		err = &lib.ErrorTransactionPaylabs
		return
	}
	defer resp.Body.Close()
	log.Info().Msgf("Response Status: %s", resp.Status)
	// Decode response
	var responsePaylabs map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&responsePaylabs)
	if err != nil {
		log.Error().Err(err).Msg("Failed to decode response from Paylabs")
		err = &lib.ErrorTransactionPaylabs
		return
	}
	log.Info().Interface("Resp", responsePaylabs).Msgf("response from paylabs")

	var virtualAccountData = responsePaylabs["virtualAccountData"].(map[string]interface{})
	paylabsVaNumber, _ := virtualAccountData["virtualAccountNo"].(string)
	transaction.VANumber = paylabsVaNumber

	err = s.EventTransactionRepo.UpdateVANo(ctx, tx, transaction.ID, paylabsVaNumber)
	if err != nil {
		log.Error().Err(err).Msg("Failed to update VA number")
		err = &lib.ErrorTransactionPaylabs
		return
	}
	var EventTransactionGarudaID dto.BulkGarudaIDRequest
	EventTransactionGarudaID.EventID = eventId
	EventTransactionGarudaID.GarudaIDs = make([]string, 0, len(req.Items))
	for _, item := range req.Items {
		EventTransactionGarudaID.GarudaIDs = append(EventTransactionGarudaID.GarudaIDs, item.GarudaID)
	}
	err = s.EventTransactionGarudaIDRepo.CreateBatch(ctx, tx, EventTransactionGarudaID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create batch garuda id")
		err = &lib.ErrorInternalServer
		return
	}
	err = tx.Commit(ctx)
	if err != nil {
		return
	}

	// Kickin check status transaction
	err = s.CheckStatusTransactionJob.EnqueueCheckTransaction(ctx, transaction.ID, s.Env.Transaction.ExpirationDuration)
	if err != nil {
		log.Error().Err(err).Str("TransactionId", transaction.ID).Msg("failed to kick job check status transaction")
		return
	}

	// Send email send bill job
	err = s.TransactionUseCase.SendBill(ctx, req.Email, req.Fullname, len(transactionItems), event, transaction, ticketCategory, venueSector)
	if err != nil {
		log.Warn().Err(err).Msg("error send bill to email")
		return
	}

	// TODO ADD JWT
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
	tx, err := s.DB.Postgres.Begin(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to begin transaction")
		return err
	}
	defer tx.Rollback(ctx)
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
	//  actual callback processing
	transactionData, err := s.EventTransactionRepo.FindByInvoiceNumber(ctx, tx, *req.TrxId)
	if err != nil {
		log.Error().Err(err).Msg("Failed to find transaction by invoice number")
		return
	}
	if transactionData.ID == "" {
		log.Error().Msg("Transaction not found")
		return &lib.ErrorInvoiceIDNotFound
	}
	markResult, err := s.EventTransactionRepo.MarkTransactionAsSuccess(ctx, tx, transactionData.ID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to mark transaction as success")
		return
	}
	// sent email to users

	log.Info().Msgf("Transaction marked as success: %v", markResult)

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
