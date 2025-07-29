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
	"github.com/rs/zerolog/log"
)

type SectorHandler interface {
	GetByVenueId(ctx *gin.Context)
}

type SectorHandlerImpl struct {
	Env          *config.EnvironmentVariable
	VenueService service.VenueService
	Validator    *validator.Validate
}

func NewSectorHandler(
	env *config.EnvironmentVariable,
	venueService service.VenueService,
	validator *validator.Validate,
) SectorHandler {
	return &SectorHandlerImpl{
		Env:          env,
		VenueService: venueService,
		Validator:    validator,
	}
}

// @Summary Get venue sectors
// @Description venue sectors
// @Tags venue
// @Produce json
// @Accept json
// @Param venueId path string false "Venue ID"
// @Success 200 {object} lib.APIResponse{data=dto.VenueSectorResponse} "Get venue sectors by venue id"
// @Failure 400 {object} lib.HTTPError "Invalid request body"
// @Failure 404 {object} lib.HTTPError "Not Found"
// @Failure 500 {object} lib.HTTPError "Internal server error"
// @Router /venues/{venueId}/sectors [get]
func (h *SectorHandlerImpl) GetByVenueId(ctx *gin.Context) {
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

	res, err := h.VenueService.GetSectorsByVenueId(ctx, uriParams.VenueID)
	if err != nil {
		log.Error().Err(err).Msg("error get sectors by venue")
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

	lib.RespondSuccess(ctx, http.StatusOK, "success", res)
}
