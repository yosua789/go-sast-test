package service

import (
	"assist-tix/config"
	"assist-tix/database"
	"assist-tix/internal/usecase"
	"assist-tix/lib"
	"assist-tix/repository"
	"context"

	"github.com/rs/zerolog/log"
)

type RetryEmailService interface {
	RetryInvoiceEmail(ctx context.Context) (err error)
}

type RetryEmailServiceImpl struct {
	DB                       *database.WrapDB
	Env                      *config.EnvironmentVariable
	EventSettingRepo         repository.EventSettingsRepository
	EventTransactionRepo     repository.EventTransactionRepository
	EventTransactionItemRepo repository.EventTransactionItemRepository
	TransactionUseCase       usecase.TransactionUsecase
}

func NewRetryEmailServiceImpl(
	db *database.WrapDB,
	env *config.EnvironmentVariable,
	eventSettingRepo repository.EventSettingsRepository,
	eventTransactionRepo repository.EventTransactionRepository,
	eventTransactionItemRepo repository.EventTransactionItemRepository,
	transactionUseCase usecase.TransactionUsecase,
) RetryEmailService {
	return &RetryEmailServiceImpl{
		DB:                       db,
		Env:                      env,
		EventSettingRepo:         eventSettingRepo,
		EventTransactionRepo:     eventTransactionRepo,
		EventTransactionItemRepo: eventTransactionItemRepo,
		TransactionUseCase:       transactionUseCase,
	}
}

func (s *RetryEmailServiceImpl) RetryInvoiceEmail(ctx context.Context) (erro error) {

	tx, err := s.DB.Postgres.Begin(ctx)
	if err != nil {
		return
	}
	defer tx.Rollback(ctx)

	log.Info().Msg("Get all transactions success")
	transactions, err := s.EventTransactionRepo.GetAllSuccessPayTransaction(ctx, tx)
	if err != nil {
		return
	}
	log.Info().Int("Count", len(transactions)).Msg("transactions success found")

	if len(transactions) == 0 {
		log.Error().Msg("transactions empty")
		return
	}

	// var garudaActive map[string]bool = make(map[string]bool)

	// rawEventSettings, err := s.EventSettingRepo.FindByEventId(ctx, tx, trx.EventID)
	// if err != nil {
	// 	return err
	// }

	// eventSettings := lib.MapEventSettings(rawEventSettings)

	var eventId string
	if s.Env.App.Mode == lib.ModeProd {
		// Event "Tajikistan vs Mali - Uzbekistan vs Indonesia"
		eventId = "91f8394d-cafc-41fb-9936-719f315b3df3"
	} else {
		// Event staging "Panama vs South Africa - Indonesia vs Tajikistan"
		eventId = "4ecca486-1ec2-4af6-b3c9-bceca149d7d8"
	}
	additionalFees, err := s.EventSettingRepo.FindAdditionalFee(ctx, nil, eventId)
	if err != nil {
		log.Error().Err(err).Msg("failed get event settings for invoice")
		return err
	}

	for _, trx := range transactions {
		log.Info().Str("TransactionID", trx.ID).Msg("Retry transaction")

		log.Info().Msg("get trx items")
		trxItems, err := s.EventTransactionItemRepo.GetTransactionItemsByTransactionId(ctx, tx, trx.ID)
		if err != nil {
			return err
		}

		log.Info().Msg("get trx detail")
		transactionDetail, err := s.EventTransactionRepo.FindTransactionDetailByTransactionId(ctx, tx, trx.ID)
		if err != nil {
			log.Error().Err(err).Msg("Failed to find transaction detail by transaction id")
			return err
		}

		err = s.TransactionUseCase.SendInvoice(
			ctx,
			trx.Email,
			trx.Fullname,
			true,
			len(trxItems),

			additionalFees,
			transactionDetail,

			*trx.PaidAt,
		)

		if err != nil {
			log.Error().Err(err).Msg("failed to send invoice job")
			return err
		}
	}

	log.Info().Msg("Retry all invoice email")
	err = tx.Commit(ctx)
	if err != nil {
		return err
	}

	log.Info().Msg("Success")

	return
}
