package service

import (
	"assist-tix/domain"
	"assist-tix/dto"
	"assist-tix/entity"
	"assist-tix/helper"
	"assist-tix/lib"
	"assist-tix/model"
	"bytes"
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"sync"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

// return transactionID and jwt. throw process on the nats consumer
func (s *EventTransactionServiceImpl) CreateEventTransactionV2(ctx *gin.Context, eventId, ticketCategoryId string, req dto.CreateEventTransaction) (res dto.EventTransactionResponse, err error) {
	log.Info().Interface("Payload", req).Str("eventId", eventId).Str("ticketCategoryId", ticketCategoryId).Str("paymentMethod", req.PaymentMethod).Msg("create event transaction")

	log.Info().Msg("validate event by id")
	event, err := s.EventRepo.FindById(ctx, nil, eventId)
	if err != nil {
		sentry.CaptureException(err)
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

	paymentMethod, err := s.PaymentMethodRepo.ValidatePaymentCodeIsActive(ctx, nil, req.PaymentMethod)
	if err != nil {
		sentry.CaptureException(err)
		return
	}

	log.Info().Msg("find event settings by event id")
	settings, err := s.EventSettingRepo.FindByEventId(ctx, nil, eventId)
	if err != nil {
		sentry.CaptureException(err)
		return
	}

	log.Info().Interface("SettingsRaw", settings).Msg("mapping event settings")
	eventSettings := lib.MapEventSettings(settings)

	log.Info().Interface("Settings", eventSettings).Msg("Event settings")

	buyCount := len(req.Items)
	log.Info().Int("count", buyCount).Int("MaxAdultTicketPerTransaction", eventSettings.MaxAdultTicketPerTransaction).Msg("buy items")
	if buyCount > eventSettings.MaxAdultTicketPerTransaction {
		err = &lib.ErrorPurchaseQuantityExceedTheLimit
		return
	}

	usedGarudaID := make(map[string]interface{})
	detailGarudaID := make(map[string]GarudaIdDetail)

	if eventSettings.GarudaIdVerification {

		// Validating garuda id to external
		var wg sync.WaitGroup
		responses := make(chan domain.PararelGarudaIDResponse, len(req.Items))

		var hasAdult bool = false
		for _, val := range req.Items {
			if val.GarudaID == "" {
				log.Error().Msg("GarudaID is required")
				return res, &lib.ErrorBadRequest
			}
			if _, ok := usedGarudaID[val.GarudaID]; ok {
				log.Warn().Str("GarudaID", val.GarudaID).Msg("Duplicate GarudaID on payload")
				return res, &lib.ErrorDuplicateGarudaIDPayload
			}

			wg.Add(1)

			usedGarudaID[val.GarudaID] = struct{}{}

			go func(wg *sync.WaitGroup, response chan<- domain.PararelGarudaIDResponse) {
				defer wg.Done()

				// Verify garuda id Validity  by external service
				externalResp, errExternal := helper.VerifyUserGarudaIDByID(s.Env.GarudaID.BaseUrl, val.GarudaID, s.Env.GarudaID.ApiKey)
				if errExternal != nil {
					log.Error().Err(errExternal).Msg("failed to verify garuda id")
					err = &lib.ErrorGetGarudaID
				} else {
					if externalResp != nil && !externalResp.Success {
						switch externalResp.ErrorCode {
						case 40401:
							err = &lib.ErrorGarudaIDNotFound
						case 42205:
							err = &lib.ErrorGarudaIDBlacklisted
						case 40909:
							err = &lib.ErrorGarudaIDInvalid
						case 40910:
							err = &lib.ErrorGarudaIDRejected
						case 50001:
							err = &lib.ErrorGetGarudaID
						}
					}
				}

				response <- domain.PararelGarudaIDResponse{
					Response: externalResp,
					Error:    err,
				}
			}(&wg, responses)

			log.Info().Str("garudaId", val.GarudaID).Msg("Calling")
		}

		// Waiting check to external API. Blocking!!!
		wg.Wait()

		// Close responses
		close(responses)

		for resp := range responses {
			if resp.Error != nil {
				err = resp.Error
				return
			}

			detailGarudaID[resp.Response.Data.FansID] = GarudaIdDetail{
				GarudaID:    resp.Response.Data.FansID,
				Name:        resp.Response.Data.Name,
				PhoneNumber: resp.Response.Data.PhoneNumber,
				Email:       resp.Response.Data.Email,
			}

			if resp.Response.Data.Age > s.Env.GarudaID.MinimumAge {
				hasAdult = true
			}

			if !hasAdult {
				err = &lib.TransactionWithoutAdultError
				log.Error().Err(err).Msg("transaction must contain at least one adult ticket")
				return res, err
			}
		}
	}

	// Start flow trx !!!
	tx, err := s.DB.Postgres.Begin(ctx)
	if err != nil {
		sentry.CaptureException(err)
		return res, err
	}
	defer tx.Rollback(ctx)

	log.Info().Str("eventId", eventId).Msg("validate email is booked in the event")
	orderInformationBookId, err := s.EventOrderInformationBookRepo.CreateOrderInformation(ctx, tx, eventId, req.Email, req.Fullname)
	if err != nil {
		sentry.CaptureException(err)
		return
	}

	log.Info().Str("eventId", eventId).Str("ticketCategoryId", ticketCategoryId).Msg("find ticket category by id and event id")
	ticketCategory, err := s.EventTicketCategoryRepo.FindByIdAndEventId(ctx, tx, eventId, ticketCategoryId)
	if err != nil {
		sentry.CaptureException(err)
		return
	}

	log.Info().Str("venueSectorId", ticketCategory.VenueSectorId).Msg("find venue by venue sector id")
	// use redis
	venueSector, err := s.VenueSectorRepo.FindVenueSectorById(ctx, tx, ticketCategory.VenueSectorId)
	if err != nil {
		sentry.CaptureException(err)
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
		sentry.CaptureException(err)
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

	// If venue doesn't have seatmap it will always empty
	var selectedSectorSeatmap map[string]entity.EventVenueSector
	if venueSector.HasSeatmap {
		log.Info().Msg("venueSector in ticket category has seatmap")
		// var seatParams []domain.SeatmapParam
		// for _, val := range req.Items {
		// 	seatParams = append(seatParams, domain.SeatmapParam{
		// 		SeatRow:    val.SeatRow,
		// 		SeatColumn: val.SeatColumn,
		// 	})
		// }

		// if s.Env.App.AutoAssignSeat {
		// 	// TODO: Add assign seat
		// } else {
		// 	// Checking choosen seat is in available status
		// 	log.Info().Msg("checking choosen seat is in available status")
		// 	sectorSeatmap, sectorSeatmapErr := s.EventTicketCategoryRepo.FindSeatmapStatusByEventSectorId(ctx, tx, eventId, ticketCategory.VenueSectorId, seatParams)
		// 	if sectorSeatmapErr != nil {
		// 		return
		// 	}
		// 	selectedSectorSeatmap = sectorSeatmap

		// 	for _, val := range req.Items {
		// 		seat, ok := sectorSeatmap[helper.ConvertRowColumnKey(val.SeatRow, val.SeatColumn)]
		// 		if !ok {
		// 			err = &lib.ErrorBookedSeatNotFound
		// 			return
		// 		} else {
		// 			switch seat.Status {
		// 			case lib.SeatmapStatusUnavailable:
		// 				err = &lib.ErrorSeatIsAlreadyBooked
		// 				return
		// 			case lib.SeatmapStatusDisable:
		// 				// err = &lib.ErrorFailedToBookSeat
		// 				err = &lib.ErrorBookedSeatNotFound
		// 				return
		// 			}
		// 		}
		// 	}

		// 	// Checking seat is already booked by try to insert
		// 	log.Info().Msg("checking seat is already booked by try to insert")
		// 	err = s.EventSeatmapBookRepo.CreateSeatBook(ctx, tx, eventId, ticketCategory.VenueSectorId, seatParams)
		// 	if err != nil {
		// 		sentry.CaptureException(err)
		// 		return
		// 	}
		// }
	}

	// TODO: Checking bulk garuda id
	if eventSettings.GarudaIdVerification {
		// var hasAdult bool
		// Verify garuda id
		// for i, item := range req.Items {

		// check internal database whether garuda id is hold
		// _, garudaIdErr := s.EventTransactionGarudaIDRepo.GetEventGarudaID(ctx, tx, eventId, item.GarudaID)
		// if garudaIdErr == nil {
		// 	return res, &lib.ErrorGarudaIDAlreadyUsed
		// }

		var garudaIds []string

		for _, val := range req.Items {
			garudaIds = append(garudaIds, val.GarudaID)
		}

		err = s.EventTransactionGarudaIDRepo.CreateGarudaIdBooks(ctx, tx, eventId, garudaIds...)
		if err != nil {
			return
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
		sentry.CaptureException(err)
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
	pgAdditionalFee := 0
	transaction.GrandTotal = transaction.TotalPrice + transaction.TotalTax + transaction.TotalAdminFee
	if paymentMethod.IsPercentage {
		log.Info().Msg("payment method is percentage, calculating additional fee")
		pgAdditionalFee = int(float64(transaction.GrandTotal) * paymentMethod.AdditionalFee / 100)

	} else {
		pgAdditionalFee = int(paymentMethod.AdditionalFee)
		log.Info().Int("pgAdditionalFee", pgAdditionalFee).Msg("payment method is fixed additional fee")
	}
	transaction.PGAdditionalFee = pgAdditionalFee
	log.Info().Int("PGAdditionalFee", pgAdditionalFee).Msg("payment method additional fee")
	transaction.GrandTotal += pgAdditionalFee

	log.Info().Int("GrandTotal", transaction.GrandTotal).Msg("got grand total price")

	log.Info().Msg("create transaction to database")
	transactionRes, err := s.EventTransactionRepo.CreateTransaction(ctx, tx, eventId, ticketCategoryId, transaction)
	if err != nil {
		sentry.CaptureException(err)
		return
	}

	//  sisanya proses by consumer
	// job async consumer
	//  create order to paylabs -> update paymentadditionalinformation -> qr string / va number
	// =  s.EventOrderInformationBookRepo.UpdateTransactionIdByID(ctx, tx, orderInformationBookId, transaction.ID)
	//  data yang dikirim = model.Transaction
	//  orderInformationBookId
	transaction.ID = transactionRes.ID
	transaction.CreatedAt = transactionRes.CreatedAt

	var transactionItems []model.EventTransactionItem
	for _, item := range req.Items {
		var garudaId sql.NullString = helper.ToSQLString(item.GarudaID)

		var fullName sql.NullString
		var email sql.NullString
		var phoneNumber sql.NullString

		if eventSettings.GarudaIdVerification {
			fullName = helper.ToSQLString(detailGarudaID[item.GarudaID].Name)
			email = helper.ToSQLString(detailGarudaID[item.GarudaID].Email)
			phoneNumber = helper.ToSQLString(detailGarudaID[item.GarudaID].PhoneNumber)
		} else {
			fullName = helper.ToSQLString(item.FullName)
			email = helper.ToSQLString(item.Email)
			phoneNumber = helper.ToSQLString(item.PhoneNumber)
		}

		var seatLabel sql.NullString

		seat, ok := selectedSectorSeatmap[helper.ConvertRowColumnKey(item.SeatRow, item.SeatColumn)]
		if ok {
			seatLabel = sql.NullString{String: seat.Label, Valid: true}
		}

		transactionItem := model.EventTransactionItem{
			TransactionID: transaction.ID,
			// TicketCategoryID:      ticketCategoryId,

			Quantity: 1,

			SeatRow:    item.SeatRow,
			SeatColumn: item.SeatColumn,
			SeatLabel:  seatLabel,

			GarudaID:    garudaId,
			Fullname:    fullName,
			Email:       email,
			PhoneNumber: phoneNumber,

			AdditionalInformation: sql.NullString{String: item.AdditionalInformation},
			TotalPrice:            ticketCategory.Price,

			CreatedAt: transaction.CreatedAt,
		}

		transactionItems = append(transactionItems, transactionItem)
	}
	log.Info().Str("transactionId", transaction.ID).Int("count", len(transactionItems)).Msg("create transaction item")

	err = tx.Commit(ctx)
	if err != nil {
		sentry.CaptureException(err)
		return
	}
	accessToken, err := helper.GenerateAccessToken(s.Env, transaction.ID)
	if err != nil {
		sentry.CaptureException(err)
		log.Error().Err(err).Msg("failed to generate access token")
		return
	}

	err = s.TransactionUseCase.SendAsyncOrder(ctx,
		eventSettings.GarudaIdVerification,
		len(transactionItems),
		accessToken,
		paymentMethod,
		event,
		transaction,
		ticketCategory,
		venueSector,
		transactionItems,
		ctx.ClientIP(),
		orderInformationBookId)
	if err != nil {
		sentry.CaptureException(err)
		log.Warn().Err(err).Msg("error send async order to nats")
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
		PgAdditionalFee:    transaction.PGAdditionalFee,
	}
	//
	log.Info().Msg("success create transaction")

	return
}

func (s *EventTransactionServiceImpl) CallbackVASnapV2(ctx *gin.Context, req dto.SnapCallbackPaymentRequest) (res dto.CallbackSnapResponse, err error) {
	log.Info().Msg("Processing Paylabs VA snap callback")
	header := map[string]interface{}{}
	tx, err := s.DB.Postgres.Begin(ctx)
	if err != nil {
		sentry.CaptureException(err)
		log.Error().Err(err).Msg("Failed to begin transaction")
		return
	}
	defer tx.Rollback(ctx)

	for key, value := range ctx.Request.Header {
		header[key] = value
	}

	headerString, err := json.Marshal(header)
	if err != nil {
		sentry.CaptureException(err)
		log.Error().Err(err).Msg("Failed to marshal headers")
		return
	}
	ctx.Set("headers", string(headerString))

	log.Info().Msgf("Headers: %v", header)

	rawPayload := ctx.GetString("rawPayload")
	var buf bytes.Buffer
	json.Compact(&buf, []byte(rawPayload))
	log.Info().Msgf("Raw Payload: %s", buf.String())
	log.Info().Msgf("Request URL: %v", req)
	isValid := helper.IsValidPaylabsRequest(ctx, "/transfer-va/payment", buf.String(), s.Env.Paylabs.PublicKey)
	if !isValid {
		return dto.CallbackSnapResponse{}, &lib.ErrorCallbackSignatureInvalid
	}
	//  actual callback processing
	transactionData, err := s.EventTransactionRepo.FindByOrderNumber(ctx, tx, *req.TrxId)
	if err != nil {
		sentry.CaptureException(err)
		log.Error().Err(err).Msg("Failed to find transaction by order number")
		return
	}
	if transactionData.ID == "" {
		log.Error().Msg("Transaction not found")
		return res, &lib.ErrorOrderNotFound
	}

	// MOVE TO ASYNC
	// transactionDetail, err := s.EventTransactionRepo.FindTransactionDetailByTransactionId(ctx, tx, transactionData.ID)
	// if err != nil {
	// 	sentry.CaptureException(err)
	// 	return
	// }
	// rawEventSettings, err := s.EventSettingRepo.FindByEventId(ctx, tx, transactionData.EventID)
	// if err != nil {
	// 	sentry.CaptureException(err)
	// 	return
	// }

	// eventSettings := lib.MapEventSettings(rawEventSettings)

	// transactionItems, err := s.EventTransactionItemRepo.GetTransactionItemsByTransactionId(ctx, tx, transactionData.ID)
	// if err != nil {
	// 	sentry.CaptureException(err)
	// 	log.Error().Err(err).Msg("Failed to find transaction by order number")
	// 	return
	// }

	transactionTime, err := time.Parse(time.RFC3339, *req.TrxDateTime)
	if err != nil {
		sentry.CaptureException(err)
		log.Error().Err(err).Msg("Failed to parse transaction time")
		return
	}
	log.Info().Msgf("Transaction time: %v", transactionTime)

	markResult, err := s.EventTransactionRepo.MarkTransactionStatus(ctx, tx, transactionData.ID, lib.EventTransactionStatusProcessingTicket, transactionTime, req.PaymentRequestId)
	if err != nil {
		sentry.CaptureException(err)
		log.Error().Err(err).Msg("Failed to mark transaction as success")
		return
	}

	err = s.TransactionUseCase.SendAsyncCallback(ctx,
		transactionData.ID,
	)
	if err != nil {
		sentry.CaptureException(err)
		log.Warn().Err(err).Msg("error send async callback to nats")
		return
	}

	err = tx.Commit(ctx)
	if err != nil {
		sentry.CaptureException(err)
		return
	}

	// MOVE TO ASYNC ORDER
	// sent invoice email to users with goroutine
	// go func() {
	// 	additionalFees, err := s.EventSettingRepo.FindAdditionalFee(ctx, nil, transactionDetail.Event.ID)
	// 	if err != nil {
	// 		log.Error().Err(err).Msg("failed get event settings for invoice")
	// 		return
	// 	}

	// 	if transactionData.PGAdditionalFee > 0 {
	// 		// TODO: refactor to dynamic NOT HARD CODED
	// 		additionalFees = append(additionalFees, entity.AdditionalFee{
	// 			ID:           "payment_fee",
	// 			Name:         "Payment Fee",
	// 			IsPercentage: false,
	// 			Value:        float64(transactionData.PGAdditionalFee),
	// 			IsTax:        false,
	// 			CreatedAt:    time.Now(),
	// 			UpdatedAt:    time.Now(),
	// 		})
	// 	}

	// 	err = s.TransactionUseCase.SendInvoice(
	// 		ctx,
	// 		transactionDetail.Email,
	// 		transactionDetail.Fullname,
	// 		eventSettings.GarudaIdVerification,
	// 		len(transactionItems),
	// 		additionalFees,
	// 		transactionDetail,
	// 		transactionTime,
	// 	)
	// 	if err != nil {
	// 		sentry.CaptureException(err)
	// 		log.Warn().Str("email", transactionData.Email).Err(err).Msg("failed to send job invoice")
	// 	}
	// }()

	// Request generate eticket with goroutine
	// go func() {
	// 	tx, err := s.DB.Postgres.Begin(ctx)
	// 	if err != nil {
	// 		sentry.CaptureException(err)
	// 		return
	// 	}
	// 	defer tx.Rollback(ctx)

	// 	log.Info().Str("SectorID", transactionDetail.VenueSector.ID).Str("EventID", transactionDetail.Event.ID).Msg("find last order seatmap")
	// 	res, err := s.EventSeatmapBookRepo.GetLastSeatOrderBySectorRowColumnId(ctx, tx, transactionDetail.Event.ID, transactionDetail.VenueSector.ID)
	// 	if err != nil {
	// 		log.Error().Err(err).Str("EventID", transactionDetail.Event.ID).Str("SectorID", transactionDetail.VenueSector.ID).Msg("failed to get last seat last order in event and sector")
	// 		return
	// 	}

	// 	log.Info().Str("SectorID", transactionDetail.VenueSector.ID).Str("EventID", transactionDetail.Event.ID).Int("Num", len(transactionItems)).Int("LastRow", res.SeatRow).Int("LastColumn", res.SeatColumn).Msg("find available seats for auto assign")
	// 	availableSeats, err := s.EventTicketCategoryRepo.FindNAvailableSeatAfterSectorRowColumn(ctx, tx, transactionDetail.Event.ID, transactionDetail.VenueSector.ID, len(transactionItems), res.SeatRow, res.SeatColumn)
	// 	if err != nil {
	// 		var tixErr *lib.TIXError
	// 		if errors.As(err, &tixErr) {
	// 			if tixErr.Code == lib.ErrorSeatAvailableSeatNotMatcheWithRequestSeats.Code {
	// 				log.Error().Err(err).Msg("available seats not match with requested seats")
	// 				return
	// 			}
	// 		}
	// 		log.Error().Err(err).Msg("failed to find available seats in sector by row and column")
	// 		return
	// 	}

	// 	var eventTickets []model.EventTicket

	// 	for i, val := range transactionItems {
	// 		if val.Email.Valid && val.Fullname.Valid {
	// 			ticketNumber := helper.GenerateTicketNumber(helper.PREFIX_TICKET_NUMBER)
	// 			ticketCode, err := helper.GenerateTicketCode()
	// 			if err != nil {
	// 				sentry.CaptureException(err)
	// 				log.Warn().Err(err).Msg("failed to generate ticket code")
	// 				return
	// 			}

	// 			eventTicket := model.EventTicket{
	// 				EventID:          transactionDetail.Event.ID,
	// 				TicketCategoryID: transactionDetail.TicketCategory.ID,
	// 				TransactionID:    transactionDetail.ID,

	// 				TicketOwnerEmail:       val.Email.String,
	// 				TicketOwnerFullname:    val.Fullname.String,
	// 				TicketOwnerPhoneNumber: val.PhoneNumber,
	// 				TicketOwnerGarudaId:    val.GarudaID,
	// 				TicketNumber:           ticketNumber,
	// 				TicketCode:             ticketCode,

	// 				EventTime:    transactionDetail.Event.EventTime,
	// 				EventVenue:   transactionDetail.Event.Venue.Name,
	// 				EventCity:    transactionDetail.Event.Venue.City,
	// 				EventCountry: transactionDetail.Event.Venue.Country,

	// 				SectorName: transactionDetail.VenueSector.Name,
	// 				AreaCode:   transactionDetail.VenueSector.AreaCode.String,
	// 				Entrance:   transactionDetail.TicketCategory.Entrance,

	// 				SeatRow:      availableSeats[i].SeatRow,
	// 				SeatColumn:   availableSeats[i].SeatColumn,
	// 				SeatRowLabel: helper.ToSQLInt16(int16(availableSeats[i].SeatRowLabel)),
	// 				SeatLabel:    helper.ToSQLString(availableSeats[i].Label),

	// 				IsCompliment: false,
	// 			}
	// 			ticketId, err := s.EventTicketRepo.Create(ctx, tx, eventTicket)
	// 			if err != nil {
	// 				sentry.CaptureException(err)
	// 				log.Error().Err(err).Msg("failed to create data eticket")
	// 				return
	// 			}
	// 			eventTicket.ID = ticketId
	// 			eventTickets = append(eventTickets, eventTicket)
	// 		}
	// 	}

	// 	err = tx.Commit(ctx)
	// 	if err != nil {
	// 		sentry.CaptureException(err)
	// 		log.Warn().Err(err).Msg("failed to create eticket")
	// 	}

	// 	for _, val := range eventTickets {
	// 		err = s.TransactionUseCase.SendETicket(
	// 			ctx,
	// 			eventSettings.GarudaIdVerification,
	// 			val,
	// 			transactionDetail,
	// 		)
	// 		if err != nil {
	// 			sentry.CaptureException(err)
	// 			log.Warn().Str("email", transactionData.Email).Err(err).Msg("failed to send job invoice")
	// 		}
	// 	}
	// }()

	serviceCode := "25"
	caseCode := "00"
	res.ResponseCode = "200" + serviceCode + caseCode
	res.ResponseMessage = "Transaction marked as success"
	res.VirtualAccountData.CustomerNo = transactionData.ID[:20] // 20 characters
	res.VirtualAccountData.VirtualAccountNo = req.VirtualAccountNo
	res.VirtualAccountData.VirtualAccountName = transactionData.Fullname
	res.VirtualAccountData.VirtualAccountEmail = &transactionData.Email
	res.VirtualAccountData.PaymentRequestId = req.PaymentRequestId
	// res.VirtualAccountData.PaidAmount.Value = strconv.Itoa(transactionData.GrandTotal) + ".00" // Amount with 2 decimal
	// res.VirtualAccountData.PaidAmount.Currency = "IDR" // Fixed currency

	ctx.Header("Content-Type", "application/json")
	ctx.Header("X-TIMESTAMP", time.Now().Format("2006-01-02T15:04:05.999+07:00"))
	log.Info().Msgf("Transaction marked as success: %v", markResult)

	return
}

func (s *EventTransactionServiceImpl) CallbackQRISPaylabsV2(ctx *gin.Context, req dto.QRISCallbackRequest) (res dto.QRISCallbackResponse, err error) {

	log.Info().Msg("Processing Paylabs VA snap callback")
	header := map[string]interface{}{}

	tx, err := s.DB.Postgres.Begin(ctx)
	if err != nil {
		sentry.CaptureException(err)
		log.Error().Err(err).Msg("Failed to begin transaction")
		return
	}
	defer tx.Rollback(ctx)

	for key, value := range ctx.Request.Header {
		header[key] = value
	}

	headerString, err := json.Marshal(header)
	if err != nil {
		sentry.CaptureException(err)
		log.Error().Err(err).Msg("Failed to marshal headers")
		return
	}
	ctx.Set("headers", string(headerString))
	log.Info().Msgf("Headers: %v", header)

	rawPayload := ctx.GetString("rawPayload")
	var buf bytes.Buffer
	json.Compact(&buf, []byte(rawPayload))
	log.Info().Msgf("Raw Payload: %s", buf.String())
	log.Info().Msgf("Request URL: %v", req)
	isValid := helper.IsValidPaylabsRequest(ctx, ctx.FullPath(), buf.String(), s.Env.Paylabs.PublicKey)
	if !isValid {
		return res, &lib.ErrorCallbackSignatureInvalid
	}
	var isSuccess bool
	if req.Status == "02" {
		isSuccess = true
	} else {
		isSuccess = false
	}
	//  actual callback processing
	transactionData, err := s.EventTransactionRepo.FindByOrderNumber(ctx, tx, req.MerchantTradeNo)
	if err != nil {
		sentry.CaptureException(err)
		log.Error().Err(err).Msg("Failed to find transaction by order number")
		return
	}
	if transactionData.ID == "" {
		log.Error().Msg("Transaction not found")
		return res, &lib.ErrorOrderNotFound
	}

	layout := "20060102150405"
	loc, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		sentry.CaptureException(err)
		log.Error().Err(err).Msg("Failed to load location")
		return
	}

	// Parse the time string in Asia/Jakarta location
	var markResult model.EventTransaction
	if isSuccess {
		paidAt, errConvertTime := time.ParseInLocation(layout, req.SuccessTime, loc)
		if errConvertTime != nil {
			log.Error().Err(errConvertTime).Msg("Failed to parse transaction time")
			return
		}
		transactionData.PaidAt = &paidAt
		log.Info().Msgf("Transaction time: %v", paidAt)

		// markResult, err = s.EventTransactionRepo.MarkTransactionAsSuccess(ctx, tx, transactionData.ID, t, req.PaymentMethodInfo.RRN)
		markResult, err = s.EventTransactionRepo.MarkTransactionStatus(ctx, tx, transactionData.ID, lib.EventTransactionStatusProcessingTicket, paidAt, req.PaymentMethodInfo.RRN)
		if err != nil {
			sentry.CaptureException(err)
			log.Error().Err(err).Msg("Failed to mark transaction as success")
			return
		}
		isSuccess = true
	} else {
		markResult, err = s.EventTransactionRepo.MarkTransactionAsFailed(ctx, tx, transactionData.ID, req.PaymentMethodInfo.RRN)
		if err != nil {
			sentry.CaptureException(err)
			log.Error().Err(err).Msg("Failed to mark transaction as failed")
			return
		}
	}
	// transactionDetail, err := s.EventTransactionRepo.FindTransactionDetailByTransactionId(ctx, tx, transactionData.ID)
	// if err != nil {
	// 	sentry.CaptureException(err)
	// 	log.Error().Err(err).Msg("Failed to find transaction detail by transaction id")
	// 	return
	// }

	// transactionItems, err := s.EventTransactionItemRepo.GetTransactionItemsByTransactionId(ctx, tx, transactionData.ID)
	// if err != nil {
	// 	sentry.CaptureException(err)
	// 	log.Error().Err(err).Msg("Failed to find transaction by order number")
	// 	return
	// }

	err = s.TransactionUseCase.SendAsyncCallback(ctx,
		transactionData.ID,
	)
	if err != nil {
		sentry.CaptureException(err)
		log.Warn().Err(err).Msg("error send async callback to nats")
		return
	}

	err = tx.Commit(ctx)
	if err != nil {
		sentry.CaptureException(err)
		log.Error().Err(err).Msg("Failed to commit transaction")
		return
	}

	// sent email to users
	// -----------------signature recipe----------
	date := time.Now().Format("2006-01-02T15:04:05.999+07:00")
	path := ctx.FullPath()
	partnerID := s.Env.Paylabs.AccountID
	privateKey := s.Env.Paylabs.PrivateKey
	requestID := helper.GenerateRequestID()
	// -----------------signature recipe----------
	log.Info().Msgf("Transaction marked as success: %v", markResult)
	res.MerchantID = s.Env.Paylabs.AccountID
	res.ErrCode = "0"
	res.RequestID = requestID
	jsonData, err := json.Marshal(res)
	if err != nil {
		sentry.CaptureException(err)
		log.Error().Err(err)
		return
	}
	log.Info().Msgf("JSON Payload: %s", jsonData)

	// rawEventSettings, err := s.EventSettingRepo.FindByEventId(ctx, nil, transactionData.ID)
	// if err != nil {
	// 	log.Error().Err(err).Msg("Failed to find event settings by event id")
	// 	sentry.CaptureException(err)
	// 	return
	// }

	// eventSettings := lib.MapEventSettings(rawEventSettings)

	// sent invoice email to users with goroutine
	if isSuccess {
		log.Info().Msg("Transaction is success, sending invoice and eticket")
		// go func() {
		// 	additionalFees, err := s.EventSettingRepo.FindAdditionalFee(ctx, nil, transactionDetail.Event.ID)
		// 	if err != nil {
		// 		log.Error().Err(err).Msg("failed get event settings for invoice")
		// 		return
		// 	}
		// 	if transactionDetail.PgAdditionalFee.Int32 > 0 {
		// 		// TODO: refactor to dynamic NOT HARD CODED
		// 		additionalFees = append(additionalFees, entity.AdditionalFee{
		// 			ID:           "payment_fee",
		// 			Name:         "Payment Fee",
		// 			IsPercentage: false,
		// 			Value:        float64(transactionDetail.PgAdditionalFee.Int32),
		// 			IsTax:        false,
		// 			CreatedAt:    time.Now(),
		// 			UpdatedAt:    time.Now(),
		// 		})
		// 	}
		// 	err = s.TransactionUseCase.SendInvoice(
		// 		ctx,
		// 		transactionDetail.Email,
		// 		transactionDetail.Fullname,
		// 		eventSettings.GarudaIdVerification,
		// 		len(transactionItems),
		// 		additionalFees,
		// 		transactionDetail,
		// 		paidAt,
		// 	)
		// 	if err != nil {
		// 		sentry.CaptureException(err)
		// 		log.Warn().Str("email", transactionData.Email).Err(err).Msg("failed to send job invoice")
		// 	}
		// 	log.Info().Msg("success send invoice email")
		// }()

		// // Request generate eticket with goroutine
		// go func() {
		// 	tx, err := s.DB.Postgres.Begin(ctx)
		// 	if err != nil {
		// 		sentry.CaptureException(err)
		// 		log.Error().Err(err).Msg("failed to begin transaction for eticket creation")
		// 		return
		// 	}
		// 	defer tx.Rollback(ctx)

		// 	var eventTickets []model.EventTicket

		// 	for _, val := range transactionItems {
		// 		if val.Email.Valid && val.Fullname.Valid {
		// 			ticketNumber := helper.GenerateTicketNumber(helper.PREFIX_TICKET_NUMBER)
		// 			ticketCode, err := helper.GenerateTicketCode()
		// 			if err != nil {
		// 				sentry.CaptureException(err)
		// 				log.Warn().Err(err).Msg("failed to generate ticket code")
		// 				return
		// 			}

		// 			eventTicket := model.EventTicket{
		// 				EventID:          transactionDetail.Event.ID,
		// 				TicketCategoryID: transactionDetail.TicketCategory.ID,
		// 				TransactionID:    transactionDetail.ID,

		// 				TicketOwnerEmail:       val.Email.String,
		// 				TicketOwnerFullname:    val.Fullname.String,
		// 				TicketOwnerPhoneNumber: val.PhoneNumber,
		// 				TicketOwnerGarudaId:    val.GarudaID,
		// 				TicketNumber:           ticketNumber,
		// 				TicketCode:             ticketCode,

		// 				EventTime:    transactionDetail.Event.EventTime,
		// 				EventVenue:   transactionDetail.Event.Venue.Name,
		// 				EventCity:    transactionDetail.Event.Venue.City,
		// 				EventCountry: transactionDetail.Event.Venue.Country,

		// 				SectorName:   transactionDetail.VenueSector.Name,
		// 				AreaCode:     transactionDetail.VenueSector.AreaCode.String,
		// 				Entrance:     transactionDetail.TicketCategory.Entrance,
		// 				SeatRow:      val.SeatRow,
		// 				SeatColumn:   val.SeatColumn,
		// 				SeatLabel:    val.SeatLabel,
		// 				IsCompliment: false,
		// 			}
		// 			ticketId, err := s.EventTicketRepo.Create(ctx, tx, eventTicket)
		// 			if err != nil {
		// 				sentry.CaptureException(err)
		// 				log.Error().Err(err).Msg("failed to create data eticket")
		// 				return
		// 			}
		// 			eventTicket.ID = ticketId
		// 			eventTickets = append(eventTickets, eventTicket)
		// 		}
		// 		log.Info().Msgf("created eticket for %s", val.Email.String)
		// 	}

		// 	err = tx.Commit(ctx)
		// 	if err != nil {
		// 		sentry.CaptureException(err)
		// 		log.Warn().Err(err).Msg("failed to create eticket")
		// 	}

		// 	for _, val := range eventTickets {
		// 		err = s.TransactionUseCase.SendETicket(
		// 			ctx,
		// 			eventSettings.GarudaIdVerification,
		// 			val,
		// 			transactionDetail,
		// 		)
		// 		if err != nil {
		// 			sentry.CaptureException(err)
		// 			log.Warn().Str("email", transactionData.Email).Err(err).Msg("failed to send job invoice")
		// 		}
		// 		log.Info().Msgf("success send eticket to %s", val.TicketOwnerEmail)
		// 	}
		// }()

	}
	// Hash the JSON body
	shaJson := sha256.Sum256(jsonData)
	signature := helper.GenerateSignature(shaJson, path, date, privateKey)
	ctx.Header("X-PARTNER-ID", partnerID)
	ctx.Header("X-REQUEST-ID", requestID)
	ctx.Header("X-TIMESTAMP", date)
	ctx.Header("X-SIGNATURE", signature)
	ctx.Header("Content-Type", "application/json;charset=utf-8")
	log.Info().Msg("success process callback qris paylabs")
	return

}
