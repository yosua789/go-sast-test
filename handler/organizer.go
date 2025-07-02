package handler

import (
	"assist-tix/config"
	"assist-tix/dto"
	"assist-tix/lib"
	"assist-tix/service"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type OrganizerHandler interface {
	Create(ctx *gin.Context)
	GetByID(ctx *gin.Context)
	GetAll(ctx *gin.Context)
	Update(ctx *gin.Context)
	Delete(ctx *gin.Context)
}

type OrganizerHandlerImpl struct {
	Env              *config.EnvironmentVariable
	OrganizerService service.OrganizerService
	Validator        *validator.Validate
}

func NewOrganizerHandler(
	env *config.EnvironmentVariable,
	organizerService service.OrganizerService,
	validator *validator.Validate,
) OrganizerHandler {
	return &OrganizerHandlerImpl{
		Env:              env,
		OrganizerService: organizerService,
		Validator:        validator,
	}
}

// @Summary Create organizer
// @Description Create organizer
// @Tags organizer
// @Produce json
// @Param name formData string false "Name"
// @Param slug formData string false "Slug"
// @Param logo formData file true "Logo"
// @Success 200 {object} lib.APIResponse{data=nil} "Organizer created successfully"
// @Failure 400 {object} lib.HTTPError "Invalid request body"
// @Failure 404 {object} lib.HTTPError "Not Found"
// @Failure 500 {object} lib.HTTPError "Internal server error"
// @Router /organizers [post]
func (h *OrganizerHandlerImpl) Create(ctx *gin.Context) {
	var request dto.CreateOrganizerRequest

	if err := ctx.ShouldBind(&request); err != nil {
		lib.RespondError(ctx, http.StatusBadRequest, err.Error(), err, lib.ErrorBadRequest.Code, h.Env.App.Debug)
		return
	}

	if err := h.Validator.Struct(request); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			for _, fieldErr := range validationErrors {
				lib.RespondError(ctx, http.StatusBadRequest, fieldErr.Field()+" is invalid", fieldErr, lib.ErrorBadRequest.Code, h.Env.App.Debug)
				return
			}
		}
		lib.RespondError(ctx, http.StatusBadRequest, "bad request. check your payload", nil, lib.ErrorBadRequest.Code, h.Env.App.Debug)
		return
	}

	// int64(h.Env.FileUpload.MaxSize)<<20 -> Calculate as MegaBytes
	if request.Logo.Size > int64(h.Env.FileUpload.MaxSize)<<20 {
		lib.RespondError(ctx, http.StatusRequestEntityTooLarge, lib.ErrorOrganizerPosterSizeExceeds.Error(), lib.ErrorOrganizerPosterSizeExceeds.Err, lib.ErrorOrganizerPosterSizeExceeds.Code, h.Env.App.Debug)
		return
	}

	// Validate upload image
	// * Memastikan field adalah *multipart.FileHeader
	// * Mengambil 512 byte awal untuk mengecek MIME type
	// * Gunakan http.DetectContentType	Dapatkan tipe file berdasarkan isi, bukan ekstensi
	// * Membandingkan dengan daftar MIME type yang diizinkan
	// * Return true jika cocok, false jika tidak

	file, _, err := ctx.Request.FormFile("logo")
	if err != nil {
		lib.RespondError(ctx, http.StatusBadRequest, "file is required", err, lib.ErrorBadRequest.Code, h.Env.App.Debug)
		return
	}
	defer file.Close()

	_, err = h.OrganizerService.CreateOrganizer(ctx, request, file)
	if err != nil {
		var tixErr *lib.TIXError
		if errors.As(err, &tixErr) {
			switch *tixErr {
			default:
				lib.RespondError(ctx, http.StatusInternalServerError, "failed to create organizer", err, lib.ErrorInternalServer.Code, h.Env.App.Debug)
			}
		} else {
			lib.RespondError(ctx, http.StatusInternalServerError, "failed to create organizer", err, lib.ErrorInternalServer.Code, h.Env.App.Debug)
		}
		return
	}

	lib.RespondSuccess(ctx, http.StatusCreated, "success", nil)
}

// @Summary Get organizer by ID
// @Description Get organizer by ID
// @Tags organizer
// @Produce json
// @Param organizerId path string true "Organizer ID"
// @Success 200 {object} lib.APIResponse{data=dto.OrganizerResponse} "Organizer response"
// @Failure 400 {object} lib.HTTPError "Invalid request body"
// @Failure 404 {object} lib.HTTPError "Not Found"
// @Failure 500 {object} lib.HTTPError "Internal server error"
// @Router /organizers/{organizerId} [get]
func (h *OrganizerHandlerImpl) GetByID(ctx *gin.Context) {
	var uriParams dto.GetOrganizerByIdParams

	if err := ctx.ShouldBindUri(&uriParams); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			for _, fieldErr := range validationErrors {
				lib.RespondError(ctx, http.StatusBadRequest, fieldErr.Field()+" is invalid", fieldErr, lib.ErrorBadRequest.Code, h.Env.App.Debug)
				return
			}
		}
		lib.RespondError(ctx, http.StatusBadRequest, "bad request. check your payload", nil, lib.ErrorBadRequest.Code, h.Env.App.Debug)
		return
	}

	res, err := h.OrganizerService.GetOrganizerById(ctx, uriParams.OrganizerId)
	if err != nil {
		var tixErr *lib.TIXError
		if errors.As(err, &tixErr) {
			switch *tixErr {
			case lib.ErrorOrganizerNotFound:
				lib.RespondError(ctx, http.StatusNotFound, "organizer not found", err, lib.ErrorOrganizerNotFound.Code, h.Env.App.Debug)
			case lib.ErrorOrganizerIdInvalid:
				lib.RespondError(ctx, http.StatusBadRequest, "failed to find organizer", err, lib.ErrorInternalServer.Code, h.Env.App.Debug)
			default:
				lib.RespondError(ctx, http.StatusInternalServerError, "failed to find organizer", err, lib.ErrorInternalServer.Code, h.Env.App.Debug)
			}
		} else {
			lib.RespondError(ctx, http.StatusInternalServerError, "failed to find organizer", err, lib.ErrorInternalServer.Code, h.Env.App.Debug)
		}
		return
	}

	lib.RespondSuccess(ctx, http.StatusOK, "success", res)
}

// @Summary Get all organizer
// @Description Get all organizer
// @Tags organizer
// @Produce json
// @Success 200 {object} lib.APIResponse{data=[]dto.OrganizerResponse} "Organizers response"
// @Failure 400 {object} lib.HTTPError "Invalid request body"
// @Failure 404 {object} lib.HTTPError "Not Found"
// @Failure 500 {object} lib.HTTPError "Internal server error"
// @Router /organizers [get]
func (h *OrganizerHandlerImpl) GetAll(ctx *gin.Context) {
	res, err := h.OrganizerService.GetAllOrganizer(ctx)
	if err != nil {
		var tixErr *lib.TIXError
		if errors.As(err, &tixErr) {
			switch *tixErr {
			default:
				lib.RespondError(ctx, http.StatusInternalServerError, "failed to create organizer", err, lib.ErrorInternalServer.Code, h.Env.App.Debug)
			}
		} else {
			lib.RespondError(ctx, http.StatusInternalServerError, "failed to create organizer", err, lib.ErrorInternalServer.Code, h.Env.App.Debug)
		}
		return
	}

	lib.RespondSuccess(ctx, http.StatusOK, "success", res)
}

// @Summary Update organizer
// @Description Update organizer
// @Tags organizer
// @Produce json
// @Accept json
// @Param organizerId path string false "Organizer ID"
// @Param request body dto.UpdateOrganizerRequest true "Create venue request"
// @Success 200 {object} lib.APIResponse{data=nil} "Venue created successfully"
// @Failure 400 {object} lib.HTTPError "Invalid request body"
// @Failure 404 {object} lib.HTTPError "Not Found"
// @Failure 500 {object} lib.HTTPError "Internal server error"
// @Router /organizers/{organizerId} [put]
func (h *OrganizerHandlerImpl) Update(ctx *gin.Context) {
	var uriParams dto.GetOrganizerByIdParams

	if err := ctx.ShouldBindUri(&uriParams); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			for _, fieldErr := range validationErrors {
				lib.RespondError(ctx, http.StatusBadRequest, fieldErr.Field()+" is invalid", fieldErr, lib.ErrorBadRequest.Code, h.Env.App.Debug)
				return
			}
		}
		lib.RespondError(ctx, http.StatusBadRequest, "bad request. check your payload", nil, lib.ErrorBadRequest.Code, h.Env.App.Debug)
		return
	}

	var request dto.UpdateOrganizerRequest

	if err := ctx.ShouldBind(&request); err != nil {
		lib.RespondError(ctx, http.StatusBadRequest, err.Error(), err, lib.ErrorBadRequest.Code, h.Env.App.Debug)
		return
	}

	if err := h.Validator.Struct(request); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			for _, fieldErr := range validationErrors {
				lib.RespondError(ctx, http.StatusBadRequest, fieldErr.Field()+" is invalid", fieldErr, lib.ErrorBadRequest.Code, h.Env.App.Debug)
				return
			}
		}
		lib.RespondError(ctx, http.StatusBadRequest, "bad request. check your payload", nil, lib.ErrorBadRequest.Code, h.Env.App.Debug)
		return
	}

	err := h.OrganizerService.Update(ctx, uriParams.OrganizerId, request)
	if err != nil {
		var tixErr *lib.TIXError
		if errors.As(err, &tixErr) {
			switch *tixErr {
			case lib.ErrorOrganizerNotFound:
				lib.RespondError(ctx, http.StatusNotFound, "not found", err, lib.ErrorOrganizerNotFound.Code, h.Env.App.Debug)
			case lib.ErrorOrganizerIdInvalid:
				lib.RespondError(ctx, http.StatusBadRequest, "invalid", err, lib.ErrorOrganizerIdInvalid.Code, h.Env.App.Debug)
			default:
				lib.RespondError(ctx, http.StatusInternalServerError, "failed to create venue", err, lib.ErrorInternalServer.Code, h.Env.App.Debug)
			}
		} else {
			lib.RespondError(ctx, http.StatusInternalServerError, "failed to create venue", err, lib.ErrorInternalServer.Code, h.Env.App.Debug)
		}
		return
	}

	lib.RespondSuccess(ctx, http.StatusOK, "success", nil)
}

// @Summary Delete organizer
// @Description Delete organizer
// @Tags organizer
// @Produce json
// @Accept json
// @Param organizerId path string false "Organizer ID"
// @Success 200 {object} lib.APIResponse{data=nil} "Venue created successfully"
// @Failure 400 {object} lib.HTTPError "Invalid request body"
// @Failure 404 {object} lib.HTTPError "Not Found"
// @Failure 500 {object} lib.HTTPError "Internal server error"
// @Router /organizers/{organizerId} [delete]
func (h *OrganizerHandlerImpl) Delete(ctx *gin.Context) {
	var uriParams dto.GetOrganizerByIdParams

	if err := ctx.ShouldBindUri(&uriParams); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			for _, fieldErr := range validationErrors {
				lib.RespondError(ctx, http.StatusBadRequest, fieldErr.Field()+" is invalid", fieldErr, lib.ErrorBadRequest.Code, h.Env.App.Debug)
				return
			}
		}
		lib.RespondError(ctx, http.StatusBadRequest, "bad request. check your payload", nil, lib.ErrorBadRequest.Code, h.Env.App.Debug)
		return
	}

	err := h.OrganizerService.Delete(ctx, uriParams.OrganizerId)
	if err != nil {
		var tixErr *lib.TIXError
		if errors.As(err, &tixErr) {
			switch *tixErr {
			case lib.ErrorOrganizerNotFound:
				lib.RespondError(ctx, http.StatusNotFound, "not found", err, lib.ErrorOrganizerNotFound.Code, h.Env.App.Debug)
			case lib.ErrorOrganizerIdInvalid:
				lib.RespondError(ctx, http.StatusBadRequest, "invalid", err, lib.ErrorOrganizerIdInvalid.Code, h.Env.App.Debug)
			default:
				lib.RespondError(ctx, http.StatusInternalServerError, "failed to create venue", err, lib.ErrorInternalServer.Code, h.Env.App.Debug)
			}
		} else {
			lib.RespondError(ctx, http.StatusInternalServerError, "failed to create venue", err, lib.ErrorInternalServer.Code, h.Env.App.Debug)
		}
		return
	}

	lib.RespondSuccess(ctx, http.StatusOK, "success", nil)
}
