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
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

type EventTransactionService interface {
	CreateEventTransaction(ctx *gin.Context, eventId, ticketCategoryId string, req dto.CreateEventTransaction) (res dto.EventTransactionResponse, err error)
	paylabsVASnap(ctx *gin.Context, transaction model.EventTransaction) (vaNo string, err error)
	paylabsQris(ctx *gin.Context, transaction model.EventTransaction, productName string) (barcode string, err error)
	CallbackVASnap(ctx *gin.Context, req dto.SnapCallbackPaymentRequest) (err error)
	ValidateEmailIsAlreadyBook(ctx *gin.Context, eventId, email string) (err error)
	GetAvailablePaymentMethods(ctx *gin.Context, eventId string) (res []dto.EventGrouppedPaymentMethodsResponse, err error)
	FindById(ctx context.Context, transactionID string) (res dto.OrderDetails, err error)
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
	EventTicketRepo               repository.EventTicketRepository
	VenueSectorRepo               repository.VenueSectorRepository
	PaymentMethodRepo             repository.PaymentMethodRepository

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
	eventTicketRepo repository.EventTicketRepository,
	paymentMethodRepo repository.PaymentMethodRepository,

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
		PaymentMethodRepo:             paymentMethodRepo,
		EventTicketRepo:               eventTicketRepo,

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

	paymentMethod, err := s.PaymentMethodRepo.ValidatePaymentCodeIsActive(ctx, tx, req.PaymentMethod)
	if err != nil {
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
	expiryOrder := now.Add(s.Env.Transaction.ExpirationDuration)
	orderNumber := helper.GeneraeteOrderNumber()
	log.Info().Str("OrderNumber", orderNumber).Msg("generated order number")

	transaction := model.EventTransaction{
		Fullname: req.Fullname, // request fullname ?
		Email:    req.Email,
		// PhoneNumber: req.PhoneNumber, // request phone number ?

		OrderNumber: orderNumber,
		Status:      lib.PaymentStatusPending,

		PaymentMethod:    req.PaymentMethod,
		PaymentChannel:   lib.PaymentChannelPaylabs,
		PaymentExpiredAt: expiryOrder,
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
	usedGarudaID := make(map[string]interface{})

	// TODO: Checking bulk garuda id
	if eventSettings.GarudaIdVerification {
		// Verify garuda id
		for i, item := range req.Items {

			if item.GarudaID == "" {
				log.Error().Msg("GarudaID is required")
				return res, &lib.ErrorBadRequest
			}
			if _, ok := usedGarudaID[item.GarudaID]; ok {
				log.Warn().Str("GarudaID", item.GarudaID).Msg("Duplicate GarudaID on payload")
				return res, &lib.ErrorDuplicateGarudaIDPayload
			}

			usedGarudaID[item.GarudaID] = struct{}{}
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
	} else {
		for _, item := range req.Items {
			if !helper.IsValidEmail(item.Email) || !helper.ValidatePhoneNumber(item.PhoneNumber) || !helper.IsValidUsername(item.FullName) {
				log.Error().Msg("Invalid email, phone number or full name")
				return res, &lib.ErrorBadRequest
			}
		}
		// check if it already exist
	}

	// Calculate price
	transaction.TotalPrice = ticketCategory.Price * len(req.Items)
	additionalFees, err := s.EventSettingRepo.FindAdditionalFee(ctx, nil, eventId)
	if err != nil {
		log.Error().Err(err).Msg("failed to find additional fees for event")
		return
	}
	var totalTaxPercentage float64
	var totalAdminFeePercentage float64
	for _, fee := range additionalFees {
		if fee.IsTax {
			if fee.IsPercentage {
				totalTaxPercentage += fee.Value
				transaction.TotalTax += int(float64(transaction.TotalPrice) * fee.Value / 100)
			} else {
				transaction.TotalTax += int(fee.Value)
			}
		} else {
			if fee.IsPercentage {
				totalAdminFeePercentage += fee.Value
				transaction.TotalAdminFee += int(float64(transaction.TotalPrice) * fee.Value / 100)
			} else {
				transaction.TotalAdminFee += int(fee.Value)
			}
		}
	}
	if len(additionalFees) > 0 {
		var additionalFeesDetails string
		additionalFeeStr, errMarshal := json.Marshal(additionalFees)
		if errMarshal != nil {
			log.Error().Err(errMarshal).Msg("failed to marshal additional fees")
			return res, &lib.ErrorInternalServer
		}
		additionalFeesDetails = string(additionalFeeStr)
		transaction.AdditionalFeeDetails = additionalFeesDetails
	}
	transaction.AdminFeePercentage = float32(totalAdminFeePercentage)
	transaction.TaxPercentage = float32(totalTaxPercentage)
	transaction.GrandTotal = transaction.TotalPrice + transaction.TotalTax + transaction.TotalAdminFee
	// transaction.AdminFeePercentage = float32(eventSettings.AdminFeePercentage)
	// log.Info().Int("TotalAdminFee", totalAdminFee).Float32("AdminFeePercentage", transaction.AdminFeePercentage).Msg("calculate admin fee")

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
	var paymentAdditionalInformation string
	// mapping paylabs va snap or qris
	if helper.IsVA(transaction.PaymentMethod) {
		log.Info().Msg("payment method is VA, calling paylabs snap")
		var errPaylabs error
		paymentAdditionalInformation, errPaylabs = s.paylabsVASnap(ctx, transaction)
		if errPaylabs != nil {
			log.Error().Err(errPaylabs).Msg("failed to get paylabs va number")
			err = &lib.ErrorTransactionPaylabs
			return
		}
		log.Info().Str("paymentAdditionalInformation", paymentAdditionalInformation).Msg("got payment additional information")
	} else if helper.IsQRIS(transaction.PaymentMethod) {
		var errPaylabs error
		log.Info().Msg("payment method is QRIS, calling paylabs qris")
		paymentAdditionalInformation, errPaylabs = s.paylabsQris(ctx, transaction, event.Name+" - "+ticketCategory.Name)
		if errPaylabs != nil {
			log.Error().Err(errPaylabs).Msg("failed to get paylabs qris barcode")
			err = &lib.ErrorTransactionPaylabs
			return
		}

	}
	transaction.PaymentAdditionalInfo = paymentAdditionalInformation

	err = s.EventTransactionRepo.UpdatePaymentAdditionalInformation(ctx, tx, transaction.ID, paymentAdditionalInformation)
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
	err = s.TransactionUseCase.SendBill(ctx, req.Email, req.Fullname, len(transactionItems), paymentMethod, event, transaction, ticketCategory, venueSector)
	if err != nil {
		log.Warn().Err(err).Msg("error send bill to email")
		return
	}

	accessToken, err := helper.GenerateAccessToken(s.Env, transaction.ID)
	if err != nil {
		log.Error().Err(err).Msg("failed to generate access token")
		return
	}
	// set via cookie
	// helper.SetAccessToken(ctx, accessToken)
	// TODO ADD JWT
	res = dto.EventTransactionResponse{
		OrderNumber:        orderNumber,
		PaymentMethod:      req.PaymentMethod,
		TotalPrice:         transaction.TotalPrice,
		TaxPercentage:      transaction.TaxPercentage,
		TotalTax:           transaction.TotalTax,
		AdminFeePercentage: transaction.AdminFeePercentage,
		TotalAdminFee:      transaction.TotalAdminFee,
		GrandTotal:         transaction.GrandTotal,
		ExpiredAt:          transaction.PaymentExpiredAt,
		CreatedAt:          transaction.CreatedAt,
		AccessToken:        accessToken,
		TransactionID:      transaction.ID,
	}

	log.Info().Msg("success create transaction")

	return
}

// static Eventtransaction without any business logic
func (s *EventTransactionServiceImpl) paylabsVASnap(ctx *gin.Context, transaction model.EventTransaction) (vaNo string, err error) {
	//  VA SNAP Init
	expiredDate := transaction.PaymentExpiredAt.Format("2006-01-02T15:04:05+07:00")
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
		VirtualAccountName:  transaction.Fullname,             // Payer name
		VirtualAccountEmail: transaction.Email,                // Payer email
		// VirtualAccountPhone: req.Items[0].PhoneNumber,  // Mobile phone number in Indonesian format
		TrxID: transaction.OrderNumber, // Merchant transaction number
		TotalAmount: dto.Amount{
			Value:    totalPriceStr, // Amount with 2 decimal
			Currency: "IDR",         // Fixed currency
		},
		AdditionalInfo: dto.AdditionalInfo{
			PaymentType: transaction.PaymentMethod, // Payment type
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
	return paylabsVaNumber, nil
}
func (s *EventTransactionServiceImpl) paylabsQris(ctx *gin.Context, transaction model.EventTransaction, productName string) (barcode string, err error) {
	// Define the request data
	currentTime := time.Now().Local() // UTC +07:00
	date := currentTime.Format("2006-01-02T15:04:05.999+07:00")
	merchantId := s.Env.Paylabs.AccountID[len(s.Env.Paylabs.AccountID)-6:] // 6 characters
	requestID := transaction.OrderNumber                                   // 20 characters

	path := "/qris/create"
	privateKeyPem := s.Env.Paylabs.PrivateKey // Private key in PEM format
	// VA
	var jsonBody = dto.PaylabsQRISRequest{
		MerchantID:      merchantId,                                   // 6 characters
		MerchantTradeNo: requestID,                                    // 8 characters
		RequestID:       requestID,                                    // 20 characters //for lookup purposes
		PaymentType:     "QRIS",                                       // Payment type
		Amount:          strconv.Itoa(transaction.GrandTotal) + ".00", // Amount with 2 decimal
		ProductName:     productName,
		Expire:          int(s.Env.Transaction.ExpirationDuration.Seconds()),      // ISO-8601 formatted expiration
		NotifyURL:       s.Env.Api.Url + "/api/v1/external/paylabs/qris/callback", // Callback URL
	}

	log.Info().Msgf("Creating event transaction with ID: %s", requestID)
	// Encode JSON body
	jsonData, err := json.Marshal(jsonBody)
	if err != nil {
		log.Error().Err(err).Msg("Failed to encode JSON body for Paylabs QRIS")
		err = &lib.ErrorTransactionPaylabs
		return
	}

	// Hash the JSON body
	shaJson := sha256.Sum256(jsonData)
	signature := helper.GenerateQRISSignature(shaJson, date, privateKeyPem)

	// Create HTTP headers
	headers := map[string]string{
		"X-TIMESTAMP":  date,
		"X-SIGNATURE":  signature,
		"X-PARTNER-ID": merchantId,
		"X-REQUEST-ID": requestID,
		"Content-Type": "application/json;charset=utf-8",
	}

	// Send HTTP request
	url := s.Env.Paylabs.BaseUrl + "/payment/v2" + path
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		log.Error().Err(err).Msg("Failed to create new request for Paylabs QRIS")
		err = &lib.ErrorTransactionPaylabs
		return

	}
	for key, value := range headers {
		req.Header.Set(key, value)
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Error().Err(err).Msg("Failed to send request to Paylabs")
		err = &lib.ErrorTransactionPaylabs
		return
	}
	defer resp.Body.Close()
	log.Info().Interface("response", resp)
	// Decode response
	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		log.Error().Err(err).Msg("Failed to decode response from Paylabs")
		err = &lib.ErrorTransactionPaylabs
		return
	}
	log.Info().Interface("Response", response).Msg("Response from Paylabs QRIS")
	barcode, ok := response["qrCode"].(string)
	if !ok {
		err = &lib.ErrorTransactionPaylabs
		log.Error().Err(err).Msg("Failed to get QR code from response")
		err = &lib.ErrorTransactionPaylabs
		return
	}

	// Print response

	return barcode, nil
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
	transactionData, err := s.EventTransactionRepo.FindByOrderNumber(ctx, tx, *req.TrxId)
	if err != nil {
		log.Error().Err(err).Msg("Failed to find transaction by order number")
		return
	}
	if transactionData.ID == "" {
		log.Error().Msg("Transaction not found")
		return &lib.ErrorOrderNotFound
	}

	transactionDetail, err := s.EventTransactionRepo.FindTransactionDetailByTransactionId(ctx, tx, transactionData.ID)
	if err != nil {
		return
	}

	transactionItems, err := s.EventTransactionItemRepo.GetTransactionItemsByTransactionId(ctx, tx, transactionData.ID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to find transaction by order number")
		return
	}

	markResult, err := s.EventTransactionRepo.MarkTransactionAsSuccess(ctx, tx, transactionData.ID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to mark transaction as success")
		return
	}

	err = tx.Commit(ctx)
	if err != nil {
		return
	}

	// sent invoice email to users with goroutine
	go func() {
		err = s.TransactionUseCase.SendInvoice(
			ctx,
			transactionDetail.Email,
			transactionDetail.Fullname,
			len(transactionItems),
			transactionDetail,
		)
		if err != nil {
			log.Warn().Str("email", transactionData.Email).Err(err).Msg("failed to send job invoice")
		}
	}()

	// Request generate eticket with goroutine
	go func() {
		tx, err := s.DB.Postgres.Begin(ctx)
		if err != nil {
			return
		}
		defer tx.Rollback(ctx)

		var eventTickets []model.EventTicket

		for _, val := range transactionItems {
			if val.Email.Valid && val.Fullname.Valid {
				ticketNumber := helper.GenerateTicketNumber(helper.PREFIX_TICKET_NUMBER)

				eventTicket := model.EventTicket{
					EventID:          transactionDetail.Event.ID,
					TicketCategoryID: transactionDetail.TicketCategory.ID,
					TransactionID:    transactionDetail.ID,

					TicketOwnerEmail:       val.Email.String,
					TicketOwnerFullname:    val.Fullname.String,
					TicketOwnerPhoneNumber: val.PhoneNumber,
					TicketOwnerGarudaId:    val.GarudaID,
					TicketNumber:           ticketNumber,
					TicketCode:             "sekarang masih kosong",

					EventTime:    transactionDetail.Event.EventTime,
					EventVenue:   transactionDetail.VenueSector.Venue.Name,
					EventCity:    transactionDetail.VenueSector.Venue.City,
					EventCountry: transactionDetail.VenueSector.Venue.Country,
					SectorName:   transactionDetail.VenueSector.Name,
					AreaCode:     transactionDetail.VenueSector.AreaCode.String,
					Entrance:     transactionDetail.TicketCategory.Entrance,
					SeatRow:      1,
					SeatColumn:   1,
					SeatLabel:    "125",
					IsCompliment: false,
				}
				ticketId, err := s.EventTicketRepo.Create(ctx, tx, eventTicket)
				if err != nil {
					log.Error().Err(err).Msg("failed to create data eticket")
					return
				}
				eventTicket.ID = ticketId
				eventTickets = append(eventTickets, eventTicket)
			}
		}

		err = tx.Commit(ctx)
		if err != nil {
			log.Warn().Err(err).Msg("failed to create eticket")
		}

		for _, val := range eventTickets {
			err = s.TransactionUseCase.SendETicket(
				ctx,
				val.TicketOwnerEmail,
				val.TicketOwnerFullname,
				val,
				transactionDetail,
			)
			if err != nil {
				log.Warn().Str("email", transactionData.Email).Err(err).Msg("failed to send job invoice")
			}
		}
	}()

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

func (s *EventTransactionServiceImpl) ValidateEmailIsAlreadyBook(ctx *gin.Context, eventId, email string) (err error) {
	tx, err := s.DB.Postgres.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

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

	err = s.EventOrderInformationBookRepo.ValidateOrderInformationByEmailEventId(ctx, tx, eventId, email)
	if err != nil {
		return err
	}

	err = tx.Commit(ctx)
	if err != nil {
		return
	}

	return nil
}

func (s *EventTransactionServiceImpl) GetAvailablePaymentMethods(ctx *gin.Context, eventId string) (res []dto.EventGrouppedPaymentMethodsResponse, err error) {
	grouppedPayments, err := s.PaymentMethodRepo.GetGrouppedActivePaymentMethod(ctx, nil)
	if err != nil {
		return res, err
	}

	for key, paymentMethods := range grouppedPayments {
		var payments []dto.EventPaymentMethodResponse = make([]dto.EventPaymentMethodResponse, 0)

		for _, payment := range paymentMethods {
			payments = append(payments, dto.EventPaymentMethodResponse{
				Name:         payment.Name,
				Logo:         payment.Logo,
				IsPaused:     payment.IsPaused,
				PauseMessage: payment.PauseMessage,
				PausedAt:     helper.ConvertNullTimeToPointer(payment.PausedAt),
				PaymentType:  payment.PaymentType,
				PaymentCode:  payment.PaymentCode,
			})
		}

		res = append(res, dto.EventGrouppedPaymentMethodsResponse{
			PaymentGroup: key,
			Payments:     payments,
		})
	}

	return res, nil
}

func (s *EventTransactionServiceImpl) FindById(ctx context.Context, transactionID string) (res dto.OrderDetails, err error) {
	tx, err := s.DB.Postgres.Begin(ctx)
	if err != nil {
		log.Error().Err(err).Msg("failed to begin transaction")
		return
	}
	defer tx.Rollback(ctx)
	log.Info().Str("transactionID", transactionID).Msg("find event transaction by id")
	res, err = s.EventTransactionRepo.FindById(ctx, tx, transactionID)
	if err != nil {
		log.Error().Err(err).Msg("failed to find event transaction by id")
		return
	}

	err = tx.Commit(ctx)
	if err != nil {
		log.Error().Err(err).Msg("failed to commit transaction")
		return
	}
	log.Info().Interface("OrderDetails", res).Msg("found event transaction by id")
	return
}
