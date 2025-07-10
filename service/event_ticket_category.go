package service

import (
	"assist-tix/config"
	"assist-tix/database"
	"assist-tix/dto"
	"assist-tix/lib"
	"assist-tix/model"
	"assist-tix/repository"
	"context"
)

type EventTicketCategoryService interface {
	Create(ctx context.Context, req dto.CreateEventTicketCategoryRequest) (err error)
	GetByEventId(ctx context.Context, eventId string) (res []dto.DetailEventTicketCategoryResponse, err error)
	GetById(ctx context.Context, eventId string, ticketCategoryId string) (res dto.DetailEventTicketCategoryResponse, err error)
	Delete(ctx context.Context, eventId, ticketCategoryId string) (err error)
}

type EventTicketCategoryServiceImpl struct {
	DB                            *database.WrapDB
	Env                           *config.EnvironmentVariable
	EventRepository               repository.EventRepository
	EventTicketCategoryRepository repository.EventTicketCategoryRepository
}

func NewEventTicketCategoryService(
	db *database.WrapDB,
	env *config.EnvironmentVariable,
	eventRepository repository.EventRepository,
	eventTicketCategoryRepository repository.EventTicketCategoryRepository,
) EventTicketCategoryService {
	return &EventTicketCategoryServiceImpl{
		DB:                            db,
		Env:                           env,
		EventRepository:               eventRepository,
		EventTicketCategoryRepository: eventTicketCategoryRepository,
	}
}

func (s *EventTicketCategoryServiceImpl) Create(ctx context.Context, req dto.CreateEventTicketCategoryRequest) (err error) {

	// Validate event id
	_, err = s.EventRepository.FindById(ctx, nil, req.EventID)
	if err != nil {
		return
	}

	createEventTicketCategory := model.EventTicketCategory{
		EventID:              req.EventID,
		Name:                 req.Name,
		Description:          req.Description,
		Price:                req.Price,
		TotalStock:           req.TotalStock,
		TotalPublicStock:     req.TotalPublicStock,
		PublicStock:          req.PublicStock,
		TotalComplimentStock: req.TotalComplimentStock,
		ComplimentStock:      req.ComplimentStock,
		Code:                 req.Code,
		Entrance:             req.Entrance,
	}

	err = s.EventTicketCategoryRepository.Create(ctx, nil, createEventTicketCategory)
	if err != nil {
		return
	}

	return
}

func (s *EventTicketCategoryServiceImpl) GetByEventId(ctx context.Context, eventId string) (res []dto.DetailEventTicketCategoryResponse, err error) {
	// Validate event id
	_, err = s.EventRepository.FindById(ctx, nil, eventId)
	if err != nil {
		return
	}

	ticketCategories, err := s.EventTicketCategoryRepository.FindByEventId(ctx, nil, eventId)
	if err != nil {
		return
	}

	res = make([]dto.DetailEventTicketCategoryResponse, 0)
	for _, val := range ticketCategories {
		res = append(res, lib.MapEventTicketCategoryModelToDetailEventTicketCategoryResponse(val))
	}

	return
}

func (s *EventTicketCategoryServiceImpl) GetById(ctx context.Context, eventId string, ticketCategoryId string) (res dto.DetailEventTicketCategoryResponse, err error) {
	ticketCategory, err := s.EventTicketCategoryRepository.FindByIdAndEventId(ctx, nil, eventId, ticketCategoryId)
	if err != nil {
		return
	}

	res = lib.MapEventTicketCategoryModelToDetailEventTicketCategoryResponse(ticketCategory)
	return
}

func (s *EventTicketCategoryServiceImpl) Delete(ctx context.Context, eventId, ticketCategoryId string) (err error) {
	_, err = s.EventTicketCategoryRepository.FindByIdAndEventId(ctx, nil, eventId, ticketCategoryId)
	if err != nil {
		return
	}

	err = s.EventTicketCategoryRepository.SoftDelete(ctx, nil, ticketCategoryId)
	if err != nil {
		return
	}

	return
}
