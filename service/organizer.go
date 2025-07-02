package service

import (
	"assist-tix/config"
	"assist-tix/database"
	"assist-tix/dto"
	"assist-tix/helper"
	"assist-tix/model"
	"assist-tix/repository"
	"context"
	"fmt"
	"mime/multipart"
	"time"

	"github.com/rs/zerolog/log"
)

type OrganizerService interface {
	CreateOrganizer(ctx context.Context, req dto.CreateOrganizerRequest, logoFile multipart.File) (res dto.OrganizerResponse, err error)
	UploadLogo(ctx context.Context, organizerId string, fileExtension string, fileHeader multipart.File) (err error)
	GetAllOrganizer(ctx context.Context) (res []dto.OrganizerResponse, err error)
	GetOrganizerById(ctx context.Context, organizerId string) (res dto.OrganizerResponse, err error)
}

type OrganizerServiceImpl struct {
	DB            *database.WrapDB
	Env           *config.EnvironmentVariable
	OrganizerRepo repository.OrganizerRepository
}

func NewOrganizerService(
	db *database.WrapDB,
	env *config.EnvironmentVariable,
	organizerRepo repository.OrganizerRepository,
) OrganizerService {
	return &OrganizerServiceImpl{
		DB:            db,
		Env:           env,
		OrganizerRepo: organizerRepo,
	}
}

func (s *OrganizerServiceImpl) CreateOrganizer(ctx context.Context, req dto.CreateOrganizerRequest, logoFile multipart.File) (res dto.OrganizerResponse, err error) {
	fileExtension := helper.GetFileExtension(req.Logo.Filename)

	_, filepath, err := singleUploadLogo(req.Name, logoFile, fileExtension)
	if err != nil {
		return
	}

	organizer := model.Organizer{
		Name: req.Name,
		Slug: req.Slug,
		Logo: filepath,
	}

	id, err := s.OrganizerRepo.Create(ctx, nil, organizer)
	if err != nil {
		return
	}

	organizer.ID = id

	res = dto.OrganizerResponse{
		ID:   organizer.ID,
		Name: organizer.Name,
		Slug: organizer.Slug,
		Logo: organizer.Logo,
	}

	return
}

func singleUploadLogo(organizerName string, file multipart.File, fileExtension string) (filename string, filepath string, err error) {
	defer file.Close()

	log.Info().Str("OrganizerName", organizerName).Msg("Start write logo")

	filename = helper.Hash256Key(fmt.Sprintf("%s-logo", organizerName)) + "." + fileExtension

	fileBuffer, err := helper.CopyFileToBuffer(file)
	if err != nil {
		log.Error().Err(err).Msg("Error copying file to buffer")
		return "", "", err
	}

	filename = helper.Hash256Key(fmt.Sprintf("%s-logo", organizerName)) + "." + fileExtension

	filepath, err = helper.SaveImage(helper.LogoDir, filename, *fileBuffer)
	if err != nil {
		return "", "", err
	}

	log.Info().Str("Filename", filename).Msg("Success write file")

	return filename, filepath, nil
}

func (s *OrganizerServiceImpl) GetAllOrganizer(ctx context.Context) (res []dto.OrganizerResponse, err error) {
	organizers, err := s.OrganizerRepo.FindAll(ctx, nil)
	if err != nil {
		return
	}

	for _, organizer := range organizers {
		var updatedAt *time.Time
		if organizer.UpdatedAt.Valid {
			updatedAt = &organizer.UpdatedAt.Time
		}
		res = append(res, dto.OrganizerResponse{
			ID:        organizer.ID,
			Name:      organizer.Name,
			Slug:      organizer.Slug,
			Logo:      organizer.Logo,
			CreatedAt: organizer.CreatedAt,
			UpdatedAt: updatedAt,
		})
	}

	return
}

func (c *OrganizerServiceImpl) UploadLogo(ctx context.Context, organizerId string, fileExtension string, file multipart.File) (err error) {
	organizer, err := c.OrganizerRepo.FindById(ctx, nil, organizerId)
	if err != nil {
		return err
	}

	_, filepath, err := singleUploadLogo(organizer.Name, file, fileExtension)
	if err != nil {
		return
	}

	oldLogo := organizer.Logo

	organizer.Logo = filepath

	err = c.OrganizerRepo.Update(ctx, nil, organizer)
	if err != nil {
		log.Error().Err(err).Msg("failed to update organizer")
		return err
	}

	profileDeleted := helper.DeleteUploadFile(oldLogo)
	log.Info().Bool("IsDeleted", profileDeleted).Msg("Delete profile")

	return nil
}

func (c *OrganizerServiceImpl) GetOrganizerById(ctx context.Context, organizerId string) (res dto.OrganizerResponse, err error) {
	organizer, err := c.OrganizerRepo.FindById(ctx, nil, organizerId)
	if err != nil {
		return res, err
	}

	var updatedAt *time.Time
	if organizer.UpdatedAt.Valid {
		updatedAt = &organizer.UpdatedAt.Time
	}

	res = dto.OrganizerResponse{
		ID:        organizer.ID,
		Name:      organizer.Name,
		Slug:      organizer.Slug,
		Logo:      organizer.Logo,
		CreatedAt: organizer.CreatedAt,
		UpdatedAt: updatedAt,
	}

	return
}
