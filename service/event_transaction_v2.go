package service

import (
	"assist-tix/domain"
	"assist-tix/dto"
	"assist-tix/entity"
	"assist-tix/helper"
	"assist-tix/lib"
	"assist-tix/model"
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
	log.Info().Str("eventId", eventId).Str("ticketCategoryId", ticketCategoryId).Str("paymentMethod", req.PaymentMethod).Msg("create event transaction")

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
		var seatParams []domain.SeatmapParam
		for _, val := range req.Items {
			seatParams = append(seatParams, domain.SeatmapParam{
				SeatRow:    val.SeatRow,
				SeatColumn: val.SeatColumn,
			})
		}

		if s.Env.App.AutoAssignSeat {
			// TODO: Add assign seat
		} else {
			// Checking choosen seat is in available status
			log.Info().Msg("checking choosen seat is in available status")
			sectorSeatmap, sectorSeatmapErr := s.EventTicketCategoryRepo.FindSeatmapStatusByEventSectorId(ctx, tx, eventId, ticketCategory.VenueSectorId, seatParams)
			if sectorSeatmapErr != nil {
				return
			}
			selectedSectorSeatmap = sectorSeatmap

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
				sentry.CaptureException(err)
				return
			}
		}
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

	// Update order information book to set transactionId
	err = s.EventOrderInformationBookRepo.UpdateTransactionIdByID(ctx, tx, orderInformationBookId, transaction.ID)
	if err != nil {
		sentry.CaptureException(err)
		return
	}

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

	// // TODO: Add item name, email phone number ->done on validating garuda id
	// log.Info().Msg("insert transaction item")
	// err = s.EventTransactionItemRepo.CreateTransactionItems(ctx, tx, transactionItems)
	// if err != nil {
	// 	sentry.CaptureException(err)
	// 	return
	// }

	var EventTransactionGarudaID dto.BulkGarudaIDRequest
	EventTransactionGarudaID.EventID = eventId
	EventTransactionGarudaID.GarudaIDs = make([]string, 0, len(req.Items))
	for _, item := range req.Items {
		EventTransactionGarudaID.GarudaIDs = append(EventTransactionGarudaID.GarudaIDs, item.GarudaID)
	}
	err = s.EventTransactionGarudaIDRepo.CreateBatch(ctx, tx, EventTransactionGarudaID)
	if err != nil {
		sentry.CaptureException(err)
		log.Error().Err(err).Msg("Failed to create batch garuda id")
		err = &lib.ErrorInternalServer
		return
	}
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
