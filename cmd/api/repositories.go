package api

import (
	"assist-tix/config"
	"assist-tix/database"
	"assist-tix/repository"

	"cloud.google.com/go/storage"
)

type Repository struct {
	OrganizerRepo                 repository.OrganizerRepository
	VenueRepo                     repository.VenueRepository
	VenueSectorRepo               repository.VenueSectorRepository
	EventRepo                     repository.EventRepository
	EventSettingRepo              repository.EventSettingsRepository
	EventTicketCategoryRepo       repository.EventTicketCategoryRepository
	EventTransactionRepo          repository.EventTransactionRepository
	EventTransactionItemRepo      repository.EventTransactionItemRepository
	EventSeatmapBookRepo          repository.EventSeatmapBookRepository
	EventTransactionGarudaIDRepo  repository.EventTransactionGarudaIDRepository
	EventOrderInformationBookRepo repository.EventOrderInformationBookRepository
	EventTicketRepo               repository.EventTicketRepository
	PaymentMethodRepository       repository.PaymentMethodRepository
	PaymentLogsRepository         repository.PaymentLogRepository
	// Storage Section
	GcsStorageRepository repository.GCSStorageRepository
}

func Newrepository(
	wrapDB *database.WrapDB,
	env *config.EnvironmentVariable,
	gcsClient *storage.Client,
	redisRepo repository.RedisRepository,
) Repository {
	return Repository{
		OrganizerRepo:                 repository.NewOrganizerRepository(wrapDB, env),
		VenueRepo:                     repository.NewVenueRepository(wrapDB, env),
		VenueSectorRepo:               repository.NewVenueSectorRepository(wrapDB, redisRepo, env),
		EventRepo:                     repository.NewEventRepository(wrapDB, redisRepo, env),
		EventSettingRepo:              repository.NewEventSettingsRepository(wrapDB, redisRepo, env),
		EventTicketCategoryRepo:       repository.NewEventTicketCategoryRepository(wrapDB, env),
		EventTransactionRepo:          repository.NewEventTransactionRepository(wrapDB, env),
		EventTransactionItemRepo:      repository.NewEventTransactionItemRepository(wrapDB, env),
		EventSeatmapBookRepo:          repository.NewEventSeatmapBookRepository(wrapDB, env),
		EventTransactionGarudaIDRepo:  repository.NewEventTransactionGarudaIDRepository(wrapDB, env),
		EventOrderInformationBookRepo: repository.NewEventOrderInformationBookRepository(wrapDB, env),
		EventTicketRepo:               repository.NewEventTicketRepository(wrapDB, env),
		PaymentMethodRepository:       repository.NewPaymentMethodRepository(wrapDB, redisRepo, env),
		GcsStorageRepository:          repository.NewGCSFileRepositoryImpl(gcsClient, env),
		PaymentLogsRepository:         repository.NewPaymentLogRepository(wrapDB, env),
	}
}
