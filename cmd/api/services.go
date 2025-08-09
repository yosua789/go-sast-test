package api

import (
	"assist-tix/config"
	"assist-tix/database"
	"assist-tix/service"
)

type Service struct {
	OrganizerService           service.OrganizerService
	VenueService               service.VenueService
	EventService               service.EventService
	EventTicketCategoryService service.EventTicketCategoryService
	EventTransactionService    service.EventTransactionService
	PaymentLogsService         service.PaymentLogsService
	RetryEmailService          service.RetryEmailService
}

func Newservice(
	env *config.EnvironmentVariable,
	r Repository,
	db *database.WrapDB,
	job Job,
	useCase UseCase,
) Service {
	organizerService := service.NewOrganizerService(db, env, r.OrganizerRepo)
	venueService := service.NewVenueService(db, env, r.VenueRepo, r.VenueSectorRepo)
	eventService := service.NewEventService(db, env, r.EventRepo, r.EventSettingRepo, r.EventTicketCategoryRepo, r.OrganizerRepo, r.VenueRepo, r.EventTransactionGarudaIDRepo, r.GcsStorageRepository)
	eventTicketCategoryService := service.NewEventTicketCategoryService(db, env, r.VenueRepo, r.VenueSectorRepo, r.EventRepo, r.EventTicketCategoryRepo, r.EventSeatmapBookRepo, r.GcsStorageRepository)
	paymentLogsService := service.NewPaymentLogsService(db, env, r.PaymentLogsRepository)
	eventTransactionService := service.NewEventTransactionService(
		db,
		env,
		r.EventRepo,
		r.EventSettingRepo,
		r.EventTicketCategoryRepo,
		r.EventTransactionRepo,
		r.EventTransactionItemRepo,
		r.EventSeatmapBookRepo,
		r.EventOrderInformationBookRepo,
		r.VenueSectorRepo,
		r.EventTransactionGarudaIDRepo,
		r.EventTicketRepo,
		r.PaymentMethodRepository,
		job.CheckStatusTransactionJob,
		r.PaymentLogsRepository,
		useCase.TransactionUseCase,
	)
	retryEmailService := service.NewRetryEmailServiceImpl(db, env, r.EventSettingRepo, r.EventTransactionRepo, r.EventTransactionItemRepo, useCase.TransactionUseCase)

	return Service{
		OrganizerService:           organizerService,
		VenueService:               venueService,
		EventService:               eventService,
		EventTicketCategoryService: eventTicketCategoryService,
		EventTransactionService:    eventTransactionService,
		PaymentLogsService:         paymentLogsService,
		RetryEmailService:          retryEmailService,
	}
}
