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

type EventTransactionHandler interface {
	CreateTransaction(ctx *gin.Context)
	PaylabsVASnap(ctx *gin.Context)
}

type EventTransactionHandlerImpl struct {
	Env                     *config.EnvironmentVariable
	EventTransactionService service.EventTransactionService
	Validator               *validator.Validate
}

func NewEventTransactionHandler(
	env *config.EnvironmentVariable,
	eventTransactionService service.EventTransactionService,
	validator *validator.Validate,
) EventTransactionHandler {
	return &EventTransactionHandlerImpl{
		Env:                     env,
		EventTransactionService: eventTransactionService,
		Validator:               validator,
	}
}

// @Summary Create event ticket transaction
// @Description Create event ticket transaction
// @Tags events
// @Produce json
// @Accept json
// @Param eventId path string true "Event ID"
// @Param ticketCategoryId path string true "Ticket Category ID"
// @Param request body dto.CreateEventTransaction true "Create event ticket transaction"
// @Success 200 {object} lib.APIResponse{data=dto.EventTransactionResponse} "Order created"
// @Failure 400 {object} lib.HTTPError "Invalid request body"
// @Failure 404 {object} lib.HTTPError "Not Found"
// @Failure 500 {object} lib.HTTPError "Internal server error"
// @Router /events/{eventId}/ticket-categories/{ticketCategoryId}/order [post]
func (h *EventTransactionHandlerImpl) CreateTransaction(ctx *gin.Context) {
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

	var request dto.CreateEventTransaction

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

	res, err := h.EventTransactionService.CreateEventTransaction(ctx, uriParams.EventID, uriParams.TicketCategoryId, request)
	if err != nil {
		log.Error().Err(err).Msg("error create event transaction")
		var tixErr *lib.TIXError
		if errors.As(err, &tixErr) {
			switch *tixErr {
			case lib.ErrorEventSaleIsPaused, lib.ErrorEventSaleIsNotStartedYet, lib.ErrorEventSaleAlreadyOver:
				lib.RespondError(ctx, http.StatusForbidden, "error", err, tixErr.Code, h.Env.App.Debug)
			case lib.ErrorSeatIsAlreadyBooked, lib.ErrorTicketIsOutOfStock, lib.ErrorPurchaseQuantityExceedTheLimit:
				lib.RespondError(ctx, http.StatusConflict, "error", err, tixErr.Code, h.Env.App.Debug)
			case lib.ErrorEventIdInvalid, lib.ErrorTicketCategoryInvalid, lib.ErrorFailedToBookSeat:
				lib.RespondError(ctx, http.StatusBadRequest, "error", err, tixErr.Code, h.Env.App.Debug)
			case lib.ErrorEventNotFound, lib.ErrorTicketCategoryNotFound:
				lib.RespondError(ctx, http.StatusNotFound, "error", err, tixErr.Code, h.Env.App.Debug)
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

// @Summary Create VA snap for event ticket transaction
// @Description Create VA snap for event ticket transaction
// @Tags events
// @Produce json
// @Accept json
// @Param eventId path string true "Event ID"
// @Success 200 {object} lib.APIResponse{data=nil} "VA snap created"
// @Failure 400 {object} lib.HTTPError "Invalid request body"
// @Failure 404 {object} lib.HTTPError "Not Found"
// @Failure 500 {object} lib.HTTPError "Internal server error"
// @Router /events/{eventId}/ticket-categories/{ticketCategoryId}/order/paylabs-vasnap [post]
func (h *EventTransactionHandlerImpl) PaylabsVASnap(ctx *gin.Context) {
	// var uriParams dto.GetDetailEventTicketCategoryByIdParams

	// if err := ctx.ShouldBindUri(&uriParams); err != nil {
	// 	if validationErrors, ok := err.(validator.ValidationErrors); ok {
	// 		for _, fieldErr := range validationErrors {
	// 			lib.RespondError(ctx, http.StatusBadRequest, fieldErr.Field()+" is invalid", fieldErr, lib.ErrorBadRequest.Code, h.Env.App.Debug)
	// 			return
	// 		}
	// 	}
	// 	lib.RespondError(ctx, http.StatusBadRequest, "bad request. check your payload", nil, lib.ErrorBadRequest.Code, h.Env.App.Debug)
	// 	return
	// }

	err := h.EventTransactionService.PaylabsVASnap(ctx)
	if err != nil {
		log.Error().Err(err).Msg("error creating VA snap")
		lib.RespondError(ctx, http.StatusInternalServerError, "error creating VA snap", err, lib.ErrorInternalServer.Code, h.Env.App.Debug)
		return
	}

	lib.RespondSuccess(ctx, http.StatusOK, "VA snap created successfully", nil)
}
