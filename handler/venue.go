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

type VenueHandler interface {
	Create(ctx *gin.Context)
	GetAll(ctx *gin.Context)
	GetById(ctx *gin.Context)
	Update(ctx *gin.Context)
	Delete(ctx *gin.Context)
}

type VenueHandlerImpl struct {
	Env          *config.EnvironmentVariable
	VenueService service.VenueService
	Validator    *validator.Validate
}

func NewVenueHandler(
	env *config.EnvironmentVariable,
	venueService service.VenueService,
	validator *validator.Validate,
) VenueHandler {
	return &VenueHandlerImpl{
		Env:          env,
		VenueService: venueService,
		Validator:    validator,
	}
}

// @Summary Create venue
// @Description Create venue
// @Tags venue
// @Produce json
// @Accept json
// @Param request body dto.CreateVenueRequest true "Create venue request"
// @Success 200 {object} lib.APIResponse{data=nil} "Venue created successfully"
// @Failure 400 {object} lib.HTTPError "Invalid request body"
// @Failure 404 {object} lib.HTTPError "Not Found"
// @Failure 500 {object} lib.HTTPError "Internal server error"
// @Router /venues [post]
func (h *VenueHandlerImpl) Create(ctx *gin.Context) {
	var request dto.CreateVenueRequest

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

	_, err := h.VenueService.CreateVenue(ctx, request)
	if err != nil {
		var tixErr *lib.TIXError
		if errors.As(err, &tixErr) {
			switch *tixErr {
			default:
				lib.RespondError(ctx, http.StatusInternalServerError, "failed to create venue", err, lib.ErrorInternalServer.Code, h.Env.App.Debug)
			}
		} else {
			lib.RespondError(ctx, http.StatusInternalServerError, "failed to create venue", err, lib.ErrorInternalServer.Code, h.Env.App.Debug)
		}
		return
	}

	lib.RespondSuccess(ctx, http.StatusCreated, "success", nil)
}

// @Summary Get all venue
// @Description Get all venue
// @Tags venue
// @Produce json
// @Success 200 {object} lib.APIResponse{data=dto.VenueResponse} "List venues"
// @Failure 400 {object} lib.HTTPError "Invalid request body"
// @Failure 404 {object} lib.HTTPError "Not Found"
// @Failure 500 {object} lib.HTTPError "Internal server error"
// @Router /venues [get]
func (h *VenueHandlerImpl) GetAll(ctx *gin.Context) {
	res, err := h.VenueService.GetAllVenue(ctx)
	if err != nil {
		var tixErr *lib.TIXError
		if errors.As(err, &tixErr) {
			switch *tixErr {
			default:
				lib.RespondError(ctx, http.StatusInternalServerError, "failed to get all venue", err, lib.ErrorInternalServer.Code, h.Env.App.Debug)
			}
		} else {
			lib.RespondError(ctx, http.StatusInternalServerError, "failed to get all venue", err, lib.ErrorInternalServer.Code, h.Env.App.Debug)
		}
		return
	}

	lib.RespondSuccess(ctx, http.StatusOK, "success", res)
}

// @Summary Get venue By ID
// @Description Get venue By ID
// @Tags venue
// @Produce json
// @Param venueId path string false "Venue ID"
// @Success 200 {object} lib.APIResponse{data=dto.VenueResponse} "Get venue by id"
// @Failure 400 {object} lib.HTTPError "Invalid request body"
// @Failure 404 {object} lib.HTTPError "Not Found"
// @Failure 500 {object} lib.HTTPError "Internal server error"
// @Router /venues/{venueId} [get]
func (h *VenueHandlerImpl) GetById(ctx *gin.Context) {
	var uriParams dto.GetVenueByIdParams

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

	res, err := h.VenueService.GetVenueById(ctx, uriParams.VenueID)
	if err != nil {
		var tixErr *lib.TIXError
		if errors.As(err, &tixErr) {
			switch *tixErr {
			case lib.ErrorVenueNotFound:
				lib.RespondError(ctx, http.StatusNotFound, "failed to find venue", err, lib.ErrorVenueNotFound.Code, h.Env.App.Debug)
			case lib.ErrorVenueIdInvalid:
				lib.RespondError(ctx, http.StatusBadRequest, "failed to find venue", err, lib.ErrorVenueIdInvalid.Code, h.Env.App.Debug)
			default:
				lib.RespondError(ctx, http.StatusInternalServerError, "failed to find venue", err, lib.ErrorInternalServer.Code, h.Env.App.Debug)
			}
		} else {
			lib.RespondError(ctx, http.StatusInternalServerError, "failed to find venue", err, lib.ErrorInternalServer.Code, h.Env.App.Debug)
		}
		return
	}

	lib.RespondSuccess(ctx, http.StatusOK, "success", res)
}

// @Summary Edit venue
// @Description Edit venue
// @Tags venue
// @Produce json
// @Accept json
// @Param venueId path string false "Venue ID"
// @Param request body dto.UpdateVenueRequest true "update venue request"
// @Success 200 {object} lib.APIResponse{data=nil} "Venue created successfully"
// @Failure 400 {object} lib.HTTPError "Invalid request body"
// @Failure 404 {object} lib.HTTPError "Not Found"
// @Failure 500 {object} lib.HTTPError "Internal server error"
// @Router /venues/{venueId} [put]
func (h *VenueHandlerImpl) Update(ctx *gin.Context) {
	var uriParams dto.GetVenueByIdParams

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

	var request dto.UpdateVenueRequest

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

	err := h.VenueService.Update(ctx, uriParams.VenueID, request)
	if err != nil {
		var tixErr *lib.TIXError
		if errors.As(err, &tixErr) {
			switch *tixErr {
			case lib.ErrorVenueNotFound:
				lib.RespondError(ctx, http.StatusNotFound, "not found", err, lib.ErrorVenueNotFound.Code, h.Env.App.Debug)
			case lib.ErrorVenueIdInvalid:
				lib.RespondError(ctx, http.StatusBadRequest, "invalid", err, lib.ErrorVenueIdInvalid.Code, h.Env.App.Debug)
			default:
				lib.RespondError(ctx, http.StatusInternalServerError, "failed to update venue", err, lib.ErrorInternalServer.Code, h.Env.App.Debug)
			}
		} else {
			lib.RespondError(ctx, http.StatusInternalServerError, "failed to update venue", err, lib.ErrorInternalServer.Code, h.Env.App.Debug)
		}
		return
	}

	lib.RespondSuccess(ctx, http.StatusOK, "success", nil)
}

// @Summary Delete venue
// @Description Delete venue
// @Tags venue
// @Produce json
// @Accept json
// @Param venueId path string false "Venue ID"
// @Success 200 {object} lib.APIResponse{data=nil} "Delete successfully"
// @Failure 400 {object} lib.HTTPError "Invalid request body"
// @Failure 404 {object} lib.HTTPError "Not Found"
// @Failure 500 {object} lib.HTTPError "Internal server error"
// @Router /venues/{venueId} [delete]
func (h *VenueHandlerImpl) Delete(ctx *gin.Context) {
	var uriParams dto.GetVenueByIdParams

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

	err := h.VenueService.Delete(ctx, uriParams.VenueID)
	if err != nil {
		var tixErr *lib.TIXError
		if errors.As(err, &tixErr) {
			switch *tixErr {
			case lib.ErrorVenueNotFound:
				lib.RespondError(ctx, http.StatusNotFound, "not found", err, lib.ErrorVenueNotFound.Code, h.Env.App.Debug)
			case lib.ErrorVenueIdInvalid:
				lib.RespondError(ctx, http.StatusBadRequest, "invalid", err, lib.ErrorVenueIdInvalid.Code, h.Env.App.Debug)
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
