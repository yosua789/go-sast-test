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
	"context"
	"time"
)

type EventTransactionService interface {
	CreateEventTransaction(ctx context.Context, eventId, ticketCategoryId string, req dto.CreateEventTransaction) (res dto.EventTransactionResponse, err error)
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
	tx, err := s.DB.Postgres.Begin(ctx)
	if err != nil {
		return res, err
	}
	defer tx.Rollback(ctx)

	_, err = s.EventRepo.FindById(ctx, tx, eventId)
	if err != nil {
		return
	}

	settings, err := s.EventSettingRepo.FindByEventId(ctx, tx, eventId)
	if err != nil {
		return
	}

	eventSettings := lib.MapEventSettings(settings)

	ticketCategory, err := s.EventTicketCategoryRepo.FindByIdAndEventId(ctx, tx, eventId, ticketCategoryId)
	if err != nil {
		return
	}

	venueSector, err := s.VenueSectorRepo.FindById(ctx, tx, ticketCategory.VenueSectorId)
	if err != nil {
		return
	}

	buyCount := len(req.Items)
	if buyCount > eventSettings.MaxAdultTicketPerTransaction {
		err = &lib.ErrorPurchaseQuantityExceedTheLimit
		return
	}

	if ticketCategory.PublicStock < 0 || buyCount > ticketCategory.PublicStock {
		err = &lib.ErrorTicketIsOutOfStock
		return
	}

	err = s.EventTicketCategoryRepo.BuyPublicTicketById(ctx, tx, eventId, ticketCategoryId, buyCount)
	if err != nil {
		return
	}

	now := time.Now()
	expiryInvoice := now.Add(s.Env.Transaction.ExpirationDuration)
	invoiceNumber := helper.GenerateInvoiceNumber()

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
		var seatParams []domain.SeatmapParam
		for _, val := range req.Items {
			seatParams = append(seatParams, domain.SeatmapParam{
				SeatRow:    val.SeatRow,
				SeatColumn: val.SeatColumn,
			})
		}

		// Checking choosen seat is in available status
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
		err = s.EventSeatmapBookRepo.CreateSeatBook(ctx, tx, eventId, ticketCategory.VenueSectorId, seatParams)
		if err != nil {
			return
		}
	}

	// TODO: Checking bulk garuda id
	if eventSettings.GarudaIdVerification {
		// Verify garuda id
	}

	// Calculate price
	transaction.TotalPrice = ticketCategory.Price * len(req.Items)
	taxPerTransaction := (eventSettings.TaxPercentage / 100) * float64(transaction.TotalPrice)
	transaction.TotalTax = int(taxPerTransaction)

	var totalAdminFee int
	if eventSettings.AdminPercentage > 0 {
		totalAdminFee = int(eventSettings.AdminPercentage/100) * transaction.TotalPrice
	} else {
		totalAdminFee = eventSettings.AdminFee
	}

	transaction.AdminFeePercentage = float32(eventSettings.AdminPercentage)
	transaction.TotalAdminFee = totalAdminFee

	transaction.GrandTotal = transaction.TotalPrice + transaction.TotalTax + transaction.TotalAdminFee

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
			AdditionalInformation: item.AdditionalInformation,
			TotalPrice:            ticketCategory.Price,
		})
	}

	// TODO: Add item name, email phone number
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

	return
}
