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

type EventHandler interface {
	GetAll(ctx *gin.Context)
	GetAllPaginated(ctx *gin.Context)
	GetById(ctx *gin.Context)
	Delete(ctx *gin.Context)
}

type EventHandlerImpl struct {
	Env          *config.EnvironmentVariable
	EventService service.EventService
	Validator    *validator.Validate
}

func NewEventHandler(
	env *config.EnvironmentVariable,
	eventService service.EventService,
	validator *validator.Validate,
) EventHandler {
	return &EventHandlerImpl{
		Env:          env,
		EventService: eventService,
		Validator:    validator,
	}
}

// @Summary Get all event
// @Description Get all event
// @Tags events
// @Produce json
// @Deprecated
// @Success 200 {object} lib.APIResponse{data=dto.EventResponse} "List events"
// @Failure 400 {object} lib.HTTPError "Invalid request body"
// @Failure 404 {object} lib.HTTPError "Not Found"
// @Failure 500 {object} lib.HTTPError "Internal server error"
// @Router /events [get]
func (h *EventHandlerImpl) GetAll(ctx *gin.Context) {
	res, err := h.EventService.GetAllEvent(ctx)
	if err != nil {
		var tixErr *lib.TIXError
		if errors.As(err, &tixErr) {
			switch *tixErr {
			default:
				lib.RespondError(ctx, http.StatusInternalServerError, "error", err, lib.ErrorInternalServer.Code, h.Env.App.Debug)
			}
		} else {
			lib.RespondError(ctx, http.StatusInternalServerError, "error", err, lib.ErrorInternalServer.Code, h.Env.App.Debug)
		}
		return
	}

	lib.RespondSuccess(ctx, http.StatusOK, "success", res)
}

// @Summary Get all paginated event
// @Description Get all paginated event
// @Tags events
// @Produce json
// @Param search query string false "Search event"
// @Param status query string false "Status event" Enums(UPCOMING, CANCELED, POSTPONED, FINISHED, ON_GOING)
// @Param page query string false "page event"
// @Success 200 {object} lib.APIResponse{data=dto.PaginatedEvents} "Paginated events"
// @Failure 400 {object} lib.HTTPError "Invalid request body"
// @Failure 404 {object} lib.HTTPError "Not Found"
// @Failure 500 {object} lib.HTTPError "Internal server error"
// @Router /events [get]
func (h *EventHandlerImpl) GetAllPaginated(ctx *gin.Context) {
	var paginationParam dto.PaginationParam
	if err := ctx.ShouldBindQuery(&paginationParam); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			for _, fieldErr := range validationErrors {
				lib.RespondError(ctx, http.StatusBadRequest, fieldErr.Field()+" is invalid", fieldErr, lib.ErrorBadRequest.Code, h.Env.App.Debug)
				return
			}
		}
		lib.RespondError(ctx, http.StatusBadRequest, "bad request. check your payload", nil, lib.ErrorBadRequest.Code, h.Env.App.Debug)
		return
	}

	if err := h.Validator.Struct(paginationParam); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			for _, fieldErr := range validationErrors {
				lib.RespondError(ctx, http.StatusBadRequest, fieldErr.Field()+" is invalid", fieldErr, lib.ErrorBadRequest.Code, h.Env.App.Debug)
				return
			}
		}
		lib.RespondError(ctx, http.StatusBadRequest, "bad request. check your payload", nil, lib.ErrorBadRequest.Code, h.Env.App.Debug)
		return
	}

	var filterEventParam dto.FilterEventRequest
	if err := ctx.ShouldBindQuery(&filterEventParam); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			for _, fieldErr := range validationErrors {
				lib.RespondError(ctx, http.StatusBadRequest, fieldErr.Field()+" is invalid", fieldErr, lib.ErrorBadRequest.Code, h.Env.App.Debug)
				return
			}
		}
		lib.RespondError(ctx, http.StatusBadRequest, "bad request. check your payload", nil, lib.ErrorBadRequest.Code, h.Env.App.Debug)
		return
	}

	if err := h.Validator.Struct(filterEventParam); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			for _, fieldErr := range validationErrors {
				lib.RespondError(ctx, http.StatusBadRequest, fieldErr.Field()+" is invalid", fieldErr, lib.ErrorBadRequest.Code, h.Env.App.Debug)
				return
			}
		}
		lib.RespondError(ctx, http.StatusBadRequest, "bad request. check your payload", nil, lib.ErrorBadRequest.Code, h.Env.App.Debug)
		return
	}

	res, err := h.EventService.GetAllEventPaginated(ctx, filterEventParam, paginationParam)
	if err != nil {
		var tixErr *lib.TIXError
		if errors.As(err, &tixErr) {
			switch *tixErr {
			case lib.ErrorPaginationPageIsInvalid:
				lib.RespondError(ctx, http.StatusBadRequest, "error", tixErr, lib.ErrorPaginationPageIsInvalid.Code, h.Env.App.Debug)
			case lib.ErrorPaginationReachMaxPage:
				lib.RespondError(ctx, http.StatusBadRequest, "error", tixErr, lib.ErrorPaginationReachMaxPage.Code, h.Env.App.Debug)
			default:
				lib.RespondError(ctx, http.StatusInternalServerError, "error", err, lib.ErrorInternalServer.Code, h.Env.App.Debug)
			}
		} else {
			lib.RespondError(ctx, http.StatusInternalServerError, "error", err, lib.ErrorInternalServer.Code, h.Env.App.Debug)
		}
		return
	}

	lib.RespondSuccess(ctx, http.StatusOK, "success", res)
}

// @Summary Get event By ID
// @Description Get event By ID
// @Tags events
// @Produce json
// @Param eventId path string false "Event ID"
// @Success 200 {object} lib.APIResponse{data=dto.DetailEventResponse} "Get venue by id"
// @Failure 400 {object} lib.HTTPError "Invalid request body"
// @Failure 404 {object} lib.HTTPError "Not Found"
// @Failure 500 {object} lib.HTTPError "Internal server error"
// @Router /events/{eventId} [get]
func (h *EventHandlerImpl) GetById(ctx *gin.Context) {
	var uriParams dto.GetEventByIdParams

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

	res, err := h.EventService.GetEventById(ctx, uriParams.EventID)
	if err != nil {
		if err != nil {
			var tixErr *lib.TIXError
			if errors.As(err, &tixErr) {
				switch *tixErr {
				case lib.ErrorEventNotFound:
					lib.RespondError(ctx, http.StatusNotFound, "error", err, lib.ErrorEventNotFound.Code, h.Env.App.Debug)
				case lib.ErrorEventIdInvalid:
					lib.RespondError(ctx, http.StatusBadRequest, "error", err, lib.ErrorEventIdInvalid.Code, h.Env.App.Debug)
				default:
					lib.RespondError(ctx, http.StatusInternalServerError, "error", err, lib.ErrorInternalServer.Code, h.Env.App.Debug)
				}
			} else {
				lib.RespondError(ctx, http.StatusInternalServerError, "error", err, lib.ErrorInternalServer.Code, h.Env.App.Debug)
			}
			return
		}
	}

	lib.RespondSuccess(ctx, http.StatusOK, "success", res)
}

// @Summary Delete event
// @Description Delete event
// @Tags events
// @Produce json
// @Param eventId path string false "Event ID"
// @Success 200 {object} lib.APIResponse{data=nil} "Delete successfully"
// @Failure 400 {object} lib.HTTPError "Invalid request body"
// @Failure 404 {object} lib.HTTPError "Not Found"
// @Failure 500 {object} lib.HTTPError "Internal server error"
// @Router /events/{eventId} [delete]
func (h *EventHandlerImpl) Delete(ctx *gin.Context) {
	var uriParams dto.GetEventByIdParams

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

	err := h.EventService.Delete(ctx, uriParams.EventID)
	if err != nil {
		var tixErr *lib.TIXError
		if errors.As(err, &tixErr) {
			switch *tixErr {
			case lib.ErrorEventNotFound:
				lib.RespondError(ctx, http.StatusNotFound, "error", err, lib.ErrorEventNotFound.Code, h.Env.App.Debug)
			case lib.ErrorEventIdInvalid:
				lib.RespondError(ctx, http.StatusBadRequest, "error", err, lib.ErrorEventIdInvalid.Code, h.Env.App.Debug)
			default:
				lib.RespondError(ctx, http.StatusInternalServerError, "error", err, lib.ErrorInternalServer.Code, h.Env.App.Debug)
			}
		} else {
			lib.RespondError(ctx, http.StatusInternalServerError, "error", err, lib.ErrorInternalServer.Code, h.Env.App.Debug)
		}
		return
	}

	lib.RespondSuccess(ctx, http.StatusOK, "success", nil)
}
