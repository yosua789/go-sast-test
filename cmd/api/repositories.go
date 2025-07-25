package api

import (
	"assist-tix/config"
	"assist-tix/database"
	"assist-tix/repository"

	"cloud.google.com/go/storage"
)

type Repository struct {
	OrganizerRepo                repository.OrganizerRepository
	VenueRepo                    repository.VenueRepository
	VenueSectorRepo              repository.VenueSectorRepository
	EventRepo                    repository.EventRepository
	EventSettingRepo             repository.EventSettingsRepository
	EventTicketCategoryRepo      repository.EventTicketCategoryRepository
	EventTransactionRepo         repository.EventTransactionRepository
	EventTransactionItemRepo     repository.EventTransactionItemRepository
	EventSeatmapBookRepo         repository.EventSeatmapBookRepository
	EventTransactionGarudaIDRepo repository.EventTransactionGarudaIDRepository

	// Storage Section
	GcsStorageRepository repository.GCSStorageRepository
}

func Newrepository(
	wrapDB *database.WrapDB,
	env *config.EnvironmentVariable,
	gcsClient *storage.Client,
) Repository {
	return Repository{
		OrganizerRepo:                repository.NewOrganizerRepository(wrapDB, env),
		VenueRepo:                    repository.NewVenueRepository(wrapDB, env),
		VenueSectorRepo:              repository.NewVenueSectorRepository(wrapDB, env),
		EventRepo:                    repository.NewEventRepository(wrapDB, env),
		EventSettingRepo:             repository.NewEventSettingsRepository(wrapDB, env),
		EventTicketCategoryRepo:      repository.NewEventTicketCategoryRepository(wrapDB, env),
		EventTransactionRepo:         repository.NewEventTransactionRepository(wrapDB, env),
		EventTransactionItemRepo:     repository.NewEventTransactionItemRepository(wrapDB, env),
		EventSeatmapBookRepo:         repository.NewEventSeatmapBookRepository(wrapDB, env),
		EventTransactionGarudaIDRepo: repository.NewEventTransactionGarudaIDRepository(wrapDB, env),
		GcsStorageRepository:         repository.NewGCSFileRepositoryImpl(gcsClient, env),
	}
}
