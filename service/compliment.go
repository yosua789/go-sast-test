package service

import (
	"assist-tix/config"
	"assist-tix/database"
	"assist-tix/dto"
	"assist-tix/helper"
	"assist-tix/internal/job"
	"assist-tix/internal/usecase"
	"assist-tix/model"
	"assist-tix/repository"
	"mime/multipart"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"github.com/xuri/excelize/v2"
)

type ComplimentService interface {
}

type ComplimentServiceImpl struct {
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
	PaymentLogsRepo               repository.PaymentLogRepository

	CheckStatusTransactionJob job.CheckStatusTransactionJob

	TransactionUseCase usecase.TransactionUsecase
}

func NewComplimentService(
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
	paymentLogsRepo repository.PaymentLogRepository,
	transactionUseCase usecase.TransactionUsecase,
) ComplimentService {
	return &ComplimentServiceImpl{
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
		PaymentLogsRepo:               paymentLogsRepo,

		CheckStatusTransactionJob: checkStatusTransactionJob,

		TransactionUseCase: transactionUseCase,
	}
}

func (s *ComplimentServiceImpl) ImportBatchComplimentTickets(ctx *gin.Context, file multipart.File) (err error) {
	// validation needed  =  event_id, event_ticket_category_id ->check stock,garuda_id already used, garuda_id valid
	// need to insert into event_transactions, event_transaction_garuda_id_books, event_transaction_items
	tx, err := s.DB.Postgres.Begin(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to begin transaction")
		return
	}
	defer tx.Rollback(ctx)
	startRow := 2
	xlsx, err := excelize.OpenReader(file)
	if err != nil {
		return
	}
	defer xlsx.Close() // add new style to make the cell red

	// A = ticketNo , B = BarcodeString, C= seatno.Section , D= ticketType.name,E= fullname, F= Fans ID , G = ticket_id_bms, H=Match name
	// sheetname=match_id
	sheets := xlsx.GetSheetList()

	var (
		payload []dto.ComplimentTableRequest
		fansIDs []string
	)

	//A Email
	//B Name
	//C MatchID
	//D TicketCategoryID
	//E GarudaID
	for _, sheet := range sheets {
		rows, errSheet := xlsx.GetRows(sheet)
		if errSheet != nil {
			log.Error().Err(errSheet).Msgf("Failed to get rows for sheet: %s", sheet)
			return errSheet
		}
		log.Info().Msgf("Processing sheet: %s\n", sheet)

		for rowIndex := startRow - 1; rowIndex < len(rows); rowIndex++ {
			row := rows[rowIndex]
			email := strings.TrimSpace(row[0])                 // A
			name := strings.TrimSpace(row[1])                  // B
			eventID := strings.TrimSpace(row[2])               // C
			eventTicketCategoryID := strings.TrimSpace(row[3]) // D
			garudaID := strings.TrimSpace(row[4])              // E
			email = strings.ToLower(email)
			garudaID = strings.ToUpper(garudaID)
			if !helper.IsValidEmail(email) {
				log.Error().Msgf("Invalid email format: %s", email)
				return
			}
			if len(row) < 5 { // Ensure there are enough columns
				continue
			}
			payload = append(payload, dto.ComplimentTableRequest{
				Email:                 email,
				Name:                  name,
				EventID:               eventID,
				EventTicketCategoryID: eventTicketCategoryID,
				GarudaID:              garudaID,
			})

			fansIDs = append(fansIDs, garudaID)
		}

	}

	return
}

func (s *ComplimentServiceImpl) CreateComplimentTickets(ctx *gin.Context, payload dto.ComplimentApiRequest) (message string, err error) {

	tx, err := s.DB.Postgres.Begin(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to begin transaction")
		return
	}
	defer tx.Rollback(ctx)
	for i, garudaID := range payload.GarudaID {
		payload.GarudaID[i] = strings.ToUpper(garudaID)
	}
	eventData, err := s.EventRepo.FindById(ctx, tx, payload.EventID)
	if err != nil {
		log.Error().Err(err).Msgf("Failed to find event with ID: %s", payload.EventID)
		return
	}
	eventCategory, err := s.EventTicketCategoryRepo.FindByIdAndEventId(ctx, tx, eventData.ID, payload.EventTicketCategoryID)
	if err != nil {
		log.Error().Err(err).Msgf("Failed to find event category with ID: %s and Event ID: %s", payload.EventTicketCategoryID, payload.EventID)
		return
	}
	if eventCategory.ComplimentStock < len(payload.GarudaID) {
		log.Error().Msgf("Compliment stock for event category %s is not enough, available: %d, requested: %d",
			payload.EventTicketCategoryID, eventCategory.ComplimentStock, len(payload.GarudaID))
		return
	}
	eventTransaction := model.EventTransaction{}

	transactionRes, err := s.EventTransactionRepo.CreateTransaction(ctx, tx, payload.EventID, payload.EventTicketCategoryID, eventTransaction)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create event transaction")
		return
	}
	eventTransaction.ID = transactionRes.ID
	eventTransaction.CreatedAt = transactionRes.CreatedAt

	return
}
