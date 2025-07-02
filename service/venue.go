package service

import (
	"assist-tix/config"
	"assist-tix/database"
	"assist-tix/dto"
	"assist-tix/helper"
	"assist-tix/model"
	"assist-tix/repository"
	"context"
)

type VenueService interface {
	CreateVenue(ctx context.Context, req dto.CreateVenueRequest) (res dto.VenueResponse, err error)
	GetAllVenue(ctx context.Context) (res []dto.VenueResponse, err error)
	GetVenueById(ctx context.Context, venueId string) (res dto.VenueResponse, err error)
	Update(ctx context.Context, venueId string, req dto.UpdateVenueRequest) (err error)
	Delete(ctx context.Context, venueId string) (err error)
}

type VenueServiceImpl struct {
	DB        *database.WrapDB
	Env       *config.EnvironmentVariable
	VenueRepo repository.VenueRepository
}

func NewVenueService(
	db *database.WrapDB,
	env *config.EnvironmentVariable,
	venueRepo repository.VenueRepository,
) VenueService {
	return &VenueServiceImpl{
		DB:        db,
		Env:       env,
		VenueRepo: venueRepo,
	}
}

func (s *VenueServiceImpl) GetAllVenue(ctx context.Context) (res []dto.VenueResponse, err error) {
	venues, err := s.VenueRepo.FindAll(ctx, nil)
	if err != nil {
		return
	}

	res = make([]dto.VenueResponse, 0)

	for _, val := range venues {
		res = append(res, dto.VenueResponse{
			ID:        val.ID,
			Type:      val.VenueType,
			Name:      val.Name,
			Country:   val.Country,
			City:      val.City,
			Capacity:  val.Capacity,
			CreatedAt: val.CreatedAt,
			UpdatedAt: helper.FromNilTime(val.UpdatedAt),
		})
	}

	return
}

func (s *VenueServiceImpl) CreateVenue(ctx context.Context, req dto.CreateVenueRequest) (res dto.VenueResponse, err error) {
	data := model.Venue{
		Name:      req.Name,
		VenueType: req.Type,
		Country:   req.Country,
		City:      req.City,
		Status:    req.Status,
		Capacity:  req.Capacity,
	}
	id, err := s.VenueRepo.Create(ctx, nil, data)
	if err != nil {
		return
	}

	data.ID = id
	res = dto.VenueResponse{
		ID:       id,
		Type:     data.VenueType,
		Name:     data.Name,
		Country:  data.Country,
		City:     data.City,
		Capacity: data.Capacity,
	}

	return
}

func (s *VenueServiceImpl) GetVenueById(ctx context.Context, venueId string) (res dto.VenueResponse, err error) {
	venue, err := s.VenueRepo.FindById(ctx, nil, venueId)
	if err != nil {
		return
	}

	res = dto.VenueResponse{
		ID:        venue.ID,
		Type:      venue.VenueType,
		Name:      venue.Name,
		Country:   venue.Country,
		City:      venue.City,
		Capacity:  venue.Capacity,
		CreatedAt: venue.CreatedAt,
		UpdatedAt: helper.FromNilTime(venue.UpdatedAt),
	}

	return
}

func (s *VenueServiceImpl) Update(ctx context.Context, venueId string, req dto.UpdateVenueRequest) (err error) {
	venue, err := s.VenueRepo.FindById(ctx, nil, venueId)
	if err != nil {
		return
	}

	venue.VenueType = req.Type
	venue.Name = req.Name
	venue.Country = req.Country
	venue.City = req.City
	venue.Capacity = req.Capacity
	venue.Status = req.Status

	err = s.VenueRepo.Update(ctx, nil, venue)
	if err != nil {
		return
	}

	return
}

func (s *VenueServiceImpl) Delete(ctx context.Context, venueId string) (err error) {
	_, err = s.VenueRepo.FindById(ctx, nil, venueId)
	if err != nil {
		return
	}

	err = s.VenueRepo.SoftDelete(ctx, nil, venueId)
	if err != nil {
		return
	}

	return
}
