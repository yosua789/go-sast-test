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

type EventTicketCategoryHandler interface {
	Create(ctx *gin.Context)
	GetByEventId(ctx *gin.Context)
	GetById(ctx *gin.Context)
	GetSeatmap(ctx *gin.Context)
}

type EventTicketCategoryHandlerImpl struct {
	Env                        *config.EnvironmentVariable
	EventTicketCategoryService service.EventTicketCategoryService
	Validator                  *validator.Validate
}

func NewEventTicketCategoryHandler(
	env *config.EnvironmentVariable,
	eventTicketCategoryService service.EventTicketCategoryService,
	validator *validator.Validate,
) EventTicketCategoryHandler {
	return &EventTicketCategoryHandlerImpl{
		Env:                        env,
		EventTicketCategoryService: eventTicketCategoryService,
		Validator:                  validator,
	}
}

// @Summary Create event ticket category
// @Description Create event ticket category
// @Tags events
// @Produce json
// @Accept json
// @Param eventId path string false "Event ID"
// @Param request body dto.CreateEventTicketCategoryRequest true "Create event ticket category"
// @Success 200 {object} lib.APIResponse{data=nil} "Success create event ticket category"
// @Failure 400 {object} lib.HTTPError "Invalid request body"
// @Failure 404 {object} lib.HTTPError "Not Found"
// @Failure 500 {object} lib.HTTPError "Internal server error"
// @Router /events/{eventId}/ticket-categories [post]
func (h *EventTicketCategoryHandlerImpl) Create(ctx *gin.Context) {
	var uriParams dto.GetEventTicketCategoryByIdParams

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

	var request dto.CreateEventTicketCategoryRequest

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

	err := h.EventTicketCategoryService.Create(ctx, uriParams.EventID, request)
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

	lib.RespondSuccess(ctx, http.StatusOK, "success", nil)
}

// @Summary Get public ticket categories by event ID
// @Description Get event By ID
// @Tags events
// @Produce json
// @Param eventId path string false "Event ID"
// @Success 200 {object} lib.APIResponse{data=dto.VenueEventTicketCategoryResponse} "Get venue by id"
// @Failure 400 {object} lib.HTTPError "Invalid request body"
// @Failure 404 {object} lib.HTTPError "Not Found"
// @Failure 500 {object} lib.HTTPError "Internal server error"
// @Router /events/{eventId}/ticket-categories [get]
func (h *EventTicketCategoryHandlerImpl) GetByEventId(ctx *gin.Context) {
	var uriParams dto.GetEventTicketCategoryByIdParams

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

	res, err := h.EventTicketCategoryService.GetVenueTicketsByEventId(ctx, uriParams.EventID)
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

	lib.RespondSuccess(ctx, http.StatusOK, "success", res)
}

// @Summary Get detail event ticket category by id
// @Description Get event ticket category by id
// @Tags events
// @Produce json
// @Accept json
// @Param eventId path string false "Event ID"
// @Param ticketCategoryId path string false "Ticket Category ID"
// @Success 200 {object} lib.APIResponse{data=dto.DetailEventResponse} "Get venue by id"
// @Failure 400 {object} lib.HTTPError "Invalid request body"
// @Failure 404 {object} lib.HTTPError "Not Found"
// @Failure 500 {object} lib.HTTPError "Internal server error"
// @Router /events/{eventId}/ticket-categories/{ticketCategoryId} [get]
func (h *EventTicketCategoryHandlerImpl) GetById(ctx *gin.Context) {
	var uriParams dto.GetDetailEventTicketCategoryByIdParams

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

	res, err := h.EventTicketCategoryService.GetById(ctx, uriParams.EventID, uriParams.TicketCategoryId)
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

	lib.RespondSuccess(ctx, http.StatusOK, "success", res)
}

// @Summary Get seatmap by event and ticket category id
// @Description Get seatmap by event and ticket category id
// @Tags events
// @Produce json
// @Accept json
// @Param eventId path string false "Event ID"
// @Param ticketCategoryId path string false "Ticket Category ID"
// @Success 200 {object} lib.APIResponse{data=dto.EventSectorSeatmapResponse} "Success get seatmap"
// @Failure 400 {object} lib.HTTPError "Invalid request body"
// @Failure 404 {object} lib.HTTPError "Not Found"
// @Failure 500 {object} lib.HTTPError "Internal server error"
// @Router /events/{eventId}/ticket-categories/{ticketCategoryId}/seatmap [get]
func (h *EventTicketCategoryHandlerImpl) GetSeatmap(ctx *gin.Context) {
	var uriParams dto.GetDetailEventTicketCategoryByIdParams

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

	res, err := h.EventTicketCategoryService.GetSeatmapByTicketCategoryId(ctx, uriParams.EventID, uriParams.TicketCategoryId)
	if err != nil {
		var tixErr *lib.TIXError
		if errors.As(err, &tixErr) {
			switch *tixErr {
			case lib.ErrorTicketCategoryNotFound, lib.ErrorEventNotFound:
				lib.RespondError(ctx, http.StatusNotFound, "error", err, tixErr.Code, h.Env.App.Debug)
			case lib.ErrorVenueSectorDoesntHaveSeatmap:
				lib.RespondError(ctx, http.StatusNotFound, "error", err, lib.ErrorVenueSectorDoesntHaveSeatmap.Code, h.Env.App.Debug)
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
