package handler

import (
	"assist-tix/config"
	"assist-tix/dto"
	"assist-tix/lib"
	"assist-tix/service"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/getsentry/sentry-go"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/rs/zerolog/log"
)

type EventTransactionHandler interface {
	CreateTransaction(ctx *gin.Context)
	// PaylabsVASnap(ctx *gin.Context)
	CallbackVASnap(ctx *gin.Context)
	CallbackQRISPaylabs(ctx *gin.Context)
	IsEmailAlreadyBook(ctx *gin.Context)
	GetAvailablePaymentMethods(ctx *gin.Context)
	GetTransactionDetails(ctx *gin.Context)
	CreateTransactionV2(ctx *gin.Context)

	CallbackVASnapV2(ctx *gin.Context)
	CallbackQRISPaylabsV2(ctx *gin.Context)
}

type EventTransactionHandlerImpl struct {
	Env                     *config.EnvironmentVariable
	EventTransactionService service.EventTransactionService
	PaymentLogsService      service.PaymentLogsService
	Validator               *validator.Validate
}

func NewEventTransactionHandler(
	env *config.EnvironmentVariable,
	eventTransactionService service.EventTransactionService,
	paymentLogsService service.PaymentLogsService,
	validator *validator.Validate,
) EventTransactionHandler {
	return &EventTransactionHandlerImpl{
		Env:                     env,
		EventTransactionService: eventTransactionService,
		PaymentLogsService:      paymentLogsService,
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
			// Find first error
			fieldErr := validationErrors[0]

			mappedError := lib.MapErrorGetDetailEventTicketCategoryByIdParams(fieldErr)
			if mappedError != nil {
				var tixErr *lib.TIXError
				if errors.As(mappedError, &tixErr) {
					lib.RespondError(ctx, http.StatusBadRequest, tixErr.Error(), tixErr, tixErr.Code, h.Env.App.Debug)
					return
				}
			}

			lib.RespondError(ctx, http.StatusBadRequest, fieldErr.Field()+" is invalid", fieldErr, lib.ErrorBadRequest.Code, h.Env.App.Debug)
			return
		}
		lib.RespondError(ctx, http.StatusBadRequest, "bad request. check your payload", nil, lib.ErrorBadRequest.Code, h.Env.App.Debug)
		return
	}

	var request dto.CreateEventTransaction

	if err := ctx.ShouldBind(&request); err != nil {
		lib.RespondError(ctx, http.StatusBadRequest, err.Error(), err, lib.ErrorBadRequest.Code, h.Env.App.Debug)
		return
	}
	if len(request.Items) < 1 {
		lib.RespondError(ctx, http.StatusBadRequest, "transaction items cannot be empty", &lib.ErrorNilTransactionItem, lib.ErrorNilTransactionItem.Code, h.Env.App.Debug)
		return
	}

	if err := h.Validator.Struct(request); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			fieldErr := validationErrors[0]

			mappedError := lib.MapErrorGetDetailEventTicketCategoryByIdParams(fieldErr)
			if mappedError != nil {
				var tixErr *lib.TIXError
				if errors.As(mappedError, &tixErr) {
					lib.RespondError(ctx, http.StatusBadRequest, tixErr.Error(), tixErr, tixErr.Code, h.Env.App.Debug)
					return
				}
			}

			lib.RespondError(ctx, http.StatusBadRequest, fieldErr.Field()+" is invalid", fieldErr, lib.ErrorBadRequest.Code, h.Env.App.Debug)
			return
		}
		lib.RespondError(ctx, http.StatusBadRequest, "bad request. check your payload", nil, lib.ErrorBadRequest.Code, h.Env.App.Debug)
		return
	}

	res, err := h.EventTransactionService.CreateEventTransaction(ctx, uriParams.EventID, uriParams.TicketCategoryID, request)
	if err != nil {
		sentry.CaptureException(err)
		log.Error().Err(err).Msg("error create event transaction")
		var tixErr *lib.TIXError
		if errors.As(err, &tixErr) {
			switch *tixErr {
			case lib.ErrorEventSaleIsPaused, lib.ErrorEventSaleIsNotStartedYet, lib.ErrorEventSaleAlreadyOver:
				lib.RespondError(ctx, http.StatusForbidden, "error", err, tixErr.Code, h.Env.App.Debug)
			case lib.ErrorSeatIsAlreadyBooked, lib.ErrorTicketIsOutOfStock, lib.ErrorPurchaseQuantityExceedTheLimit, lib.ErrorOrderInformationIsAlreadyBook, lib.ErrorGarudaIDInvalid, lib.ErrorGarudaIDRejected, lib.ErrorGarudaIDBlacklisted, lib.ErrorGarudaIDAlreadyUsed, lib.ErrorDuplicateGarudaIDPayload, lib.TransactionWithoutAdultError:
				lib.RespondError(ctx, http.StatusConflict, "error", err, tixErr.Code, h.Env.App.Debug)
			case lib.ErrorEventIdInvalid, lib.ErrorTicketCategoryInvalid, lib.ErrorFailedToBookSeat, lib.ErrorPaymentMethodInvalid, lib.ErrorBadRequest:
				lib.RespondError(ctx, http.StatusBadRequest, "error", err, tixErr.Code, h.Env.App.Debug)
			case lib.ErrorEventNotFound, lib.ErrorTicketCategoryNotFound, lib.ErrorBookedSeatNotFound, lib.ErrorGarudaIDNotFound, lib.ErrorVenueSectorNotFound:
				lib.RespondError(ctx, http.StatusNotFound, "error", err, tixErr.Code, h.Env.App.Debug)
			case lib.ErrorGetGarudaID, lib.ErrorTransactionPaylabs:
				lib.RespondError(ctx, http.StatusInternalServerError, "error", err, tixErr.Code, h.Env.App.Debug)
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

// @Summary Create event ticket transaction V2
// @Description Create event ticket transaction V2
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
// @Router /events/{eventId}/ticket-categories/{ticketCategoryId}/order/v2 [post]
func (h *EventTransactionHandlerImpl) CreateTransactionV2(ctx *gin.Context) {
	var uriParams dto.GetDetailEventTicketCategoryByIdParams

	if err := ctx.ShouldBindUri(&uriParams); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			// Find first error
			fieldErr := validationErrors[0]

			mappedError := lib.MapErrorGetDetailEventTicketCategoryByIdParams(fieldErr)
			if mappedError != nil {
				var tixErr *lib.TIXError
				if errors.As(mappedError, &tixErr) {
					lib.RespondError(ctx, http.StatusBadRequest, tixErr.Error(), tixErr, tixErr.Code, h.Env.App.Debug)
					return
				}
			}

			lib.RespondError(ctx, http.StatusBadRequest, fieldErr.Field()+" is invalid", fieldErr, lib.ErrorBadRequest.Code, h.Env.App.Debug)
			return
		}
		lib.RespondError(ctx, http.StatusBadRequest, "bad request. check your payload", nil, lib.ErrorBadRequest.Code, h.Env.App.Debug)
		return
	}

	var request dto.CreateEventTransaction

	if err := ctx.ShouldBind(&request); err != nil {
		lib.RespondError(ctx, http.StatusBadRequest, err.Error(), err, lib.ErrorBadRequest.Code, h.Env.App.Debug)
		return
	}

	if len(request.Items) < 1 {
		lib.RespondError(ctx, http.StatusBadRequest, "transaction items cannot be empty", &lib.ErrorNilTransactionItem, lib.ErrorNilTransactionItem.Code, h.Env.App.Debug)
		return
	}
	if err := h.Validator.Struct(request); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			fieldErr := validationErrors[0]

			mappedError := lib.MapErrorGetDetailEventTicketCategoryByIdParams(fieldErr)
			if mappedError != nil {
				var tixErr *lib.TIXError
				if errors.As(mappedError, &tixErr) {
					lib.RespondError(ctx, http.StatusBadRequest, tixErr.Error(), tixErr, tixErr.Code, h.Env.App.Debug)
					return
				}
			}

			lib.RespondError(ctx, http.StatusBadRequest, fieldErr.Field()+" is invalid", fieldErr, lib.ErrorBadRequest.Code, h.Env.App.Debug)
			return
		}
		lib.RespondError(ctx, http.StatusBadRequest, "bad request. check your payload", nil, lib.ErrorBadRequest.Code, h.Env.App.Debug)
		return
	}

	res, err := h.EventTransactionService.CreateEventTransactionV2(ctx, uriParams.EventID, uriParams.TicketCategoryID, request)
	if err != nil {
		sentry.CaptureException(err)
		log.Error().Err(err).Msg("error create event transaction")
		var tixErr *lib.TIXError
		if errors.As(err, &tixErr) {
			switch *tixErr {
			case lib.ErrorEventSaleIsPaused, lib.ErrorEventSaleIsNotStartedYet, lib.ErrorEventSaleAlreadyOver:
				lib.RespondError(ctx, http.StatusForbidden, "error", err, tixErr.Code, h.Env.App.Debug)
			case lib.ErrorSeatIsAlreadyBooked, lib.ErrorTicketIsOutOfStock, lib.ErrorPurchaseQuantityExceedTheLimit, lib.ErrorOrderInformationIsAlreadyBook, lib.ErrorGarudaIDInvalid, lib.ErrorGarudaIDRejected, lib.ErrorGarudaIDBlacklisted, lib.ErrorGarudaIDAlreadyUsed, lib.ErrorDuplicateGarudaIDPayload, lib.TransactionWithoutAdultError:
				lib.RespondError(ctx, http.StatusConflict, "error", err, tixErr.Code, h.Env.App.Debug)
			case lib.ErrorEventIdInvalid, lib.ErrorTicketCategoryInvalid, lib.ErrorFailedToBookSeat, lib.ErrorPaymentMethodInvalid, lib.ErrorBadRequest:
				lib.RespondError(ctx, http.StatusBadRequest, "error", err, tixErr.Code, h.Env.App.Debug)
			case lib.ErrorEventNotFound, lib.ErrorTicketCategoryNotFound, lib.ErrorBookedSeatNotFound, lib.ErrorGarudaIDNotFound, lib.ErrorVenueSectorNotFound:
				lib.RespondError(ctx, http.StatusNotFound, "error", err, tixErr.Code, h.Env.App.Debug)
			case lib.ErrorGetGarudaID, lib.ErrorTransactionPaylabs:
				lib.RespondError(ctx, http.StatusInternalServerError, "error", err, tixErr.Code, h.Env.App.Debug)
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

// // @Summary Create VA snap for event ticket transaction
// // @Description Create VA snap for event ticket transaction
// // @Tags events
// // @Produce json
// // @Accept json
// // @Param ticketCategoryId path string true "Ticket Category ID"
// // @Param eventId path string true "Event ID"
// // @Success 200 {object} lib.APIResponse{data=nil} "VA snap created"
// // @Failure 400 {object} lib.HTTPError "Invalid request body"
// // @Failure 404 {object} lib.HTTPError "Not Found"
// // @Failure 500 {object} lib.HTTPError "Internal server error"
// // @Router /events/{eventId}/ticket-categories/{ticketCategoryId}/order/paylabs-vasnap [post]
// func (h *EventTransactionHandlerImpl) PaylabsVASnap(ctx *gin.Context) {

// 	err := h.EventTransactionService.PaylabsVASnap(ctx)
// 	if err != nil {
// 		log.Error().Err(err).Msg("error creating VA snap")
// 		lib.RespondError(ctx, http.StatusInternalServerError, "error creating VA snap", err, lib.ErrorInternalServer.Code, h.Env.App.Debug)
// 		return
// 	}

// 	lib.RespondSuccess(ctx, http.StatusOK, "VA snap created successfully", nil)
// }

func (h *EventTransactionHandlerImpl) CallbackVASnap(ctx *gin.Context) {
	// Implement the callback logic here
	var req dto.SnapCallbackPaymentRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		log.Error().Err(err).Msg("error binding JSON for callback")
		lib.RespondError(ctx, http.StatusBadRequest, "invalid request body", err, lib.ErrorBadRequest.Code, h.Env.App.Debug)
		return
	}
	log.Info().Msgf("Received callback: %+v", req)

	// This is a placeholder for the actual implementation
	res, err := h.EventTransactionService.CallbackVASnap(ctx, req)
	var tixErr *lib.TIXError
	if errors.As(err, &tixErr) {
		switch *tixErr {
		case lib.ErrorOrderNotFound:
			lib.RespondError(ctx, http.StatusNotFound, tixErr.Error(), err, tixErr.Code, h.Env.App.Debug)
			// return
		case lib.ErrorCallbackSignatureInvalid:
			lib.RespondError(ctx, http.StatusBadRequest, tixErr.Error(), err, tixErr.Code, h.Env.App.Debug)
			// return
		default:
			lib.RespondError(ctx, http.StatusInternalServerError, "error", err, lib.ErrorInternalServer.Code, h.Env.App.Debug)
			// return
		}
		headerString := ctx.GetString("headers")
		bodyString := ctx.GetString("rawPayload")
		respString, err := json.Marshal(res)
		if err != nil {
			log.Error().Err(err).Msg("Failed to marshal response")
		}
		_, _ = h.PaymentLogsService.Create(ctx, bodyString, headerString, string(respString), strconv.Itoa(tixErr.Code), tixErr.Error())
		return
	}
	headerString := ctx.GetString("headers")
	bodyString := ctx.GetString("rawPayload")
	respString, err := json.Marshal(res)
	if err != nil {
		log.Error().Err(err).Msg("Failed to marshal response")
		return
	}
	_, _ = h.PaymentLogsService.Create(ctx, bodyString, headerString, string(respString), "", "")

	ctx.JSON(http.StatusOK, res)
}

func (h *EventTransactionHandlerImpl) CallbackVASnapV2(ctx *gin.Context) {
	// Implement the callback logic here
	var req dto.SnapCallbackPaymentRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		log.Error().Err(err).Msg("error binding JSON for callback")
		lib.RespondError(ctx, http.StatusBadRequest, "invalid request body", err, lib.ErrorBadRequest.Code, h.Env.App.Debug)
		return
	}
	log.Info().Msgf("Received callback: %+v", req)

	// This is a placeholder for the actual implementation
	res, err := h.EventTransactionService.CallbackVASnapV2(ctx, req)
	var tixErr *lib.TIXError
	if errors.As(err, &tixErr) {
		switch *tixErr {
		case lib.ErrorOrderNotFound:
			lib.RespondError(ctx, http.StatusNotFound, tixErr.Error(), err, tixErr.Code, h.Env.App.Debug)
			// return
		case lib.ErrorCallbackSignatureInvalid:
			lib.RespondError(ctx, http.StatusBadRequest, tixErr.Error(), err, tixErr.Code, h.Env.App.Debug)
			// return
		default:
			lib.RespondError(ctx, http.StatusInternalServerError, "error", err, lib.ErrorInternalServer.Code, h.Env.App.Debug)
			// return
		}
		headerString := ctx.GetString("headers")
		bodyString := ctx.GetString("rawPayload")
		respString, err := json.Marshal(res)
		if err != nil {
			log.Error().Err(err).Msg("Failed to marshal response")
		}
		_, _ = h.PaymentLogsService.Create(ctx, bodyString, headerString, string(respString), strconv.Itoa(tixErr.Code), tixErr.Error())
		return
	}
	headerString := ctx.GetString("headers")
	bodyString := ctx.GetString("rawPayload")
	respString, err := json.Marshal(res)
	if err != nil {
		log.Error().Err(err).Msg("Failed to marshal response")
		return
	}
	_, _ = h.PaymentLogsService.Create(ctx, bodyString, headerString, string(respString), "", "")

	ctx.JSON(http.StatusOK, res)
}

func (h *EventTransactionHandlerImpl) CallbackQRISPaylabs(ctx *gin.Context) {
	// Implement the callback logic here

	var req dto.QRISCallbackRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		log.Error().Err(err).Msg("error binding JSON for callback")
		lib.RespondError(ctx, http.StatusBadRequest, "invalid request body", err, lib.ErrorBadRequest.Code, h.Env.App.Debug)
		return
	}

	log.Info().Msgf("Received callback: %+v", req)

	// This is a placeholder for the actual implementation
	res, err := h.EventTransactionService.CallbackQRISPaylabs(ctx, req)
	var tixErr *lib.TIXError
	if errors.As(err, &tixErr) {
		switch *tixErr {
		case lib.ErrorOrderNotFound:
			lib.RespondError(ctx, http.StatusNotFound, tixErr.Error(), err, tixErr.Code, h.Env.App.Debug)
			// return
		case lib.ErrorCallbackSignatureInvalid:
			lib.RespondError(ctx, http.StatusBadRequest, tixErr.Error(), err, tixErr.Code, h.Env.App.Debug)
			// return
		default:
			lib.RespondError(ctx, http.StatusInternalServerError, "error", err, lib.ErrorInternalServer.Code, h.Env.App.Debug)
			// return
		}
		headerString := ctx.GetString("headers")
		bodyString := ctx.GetString("rawPayload")
		respString, err := json.Marshal(res)
		if err != nil {
			log.Error().Err(err).Msg("Failed to marshal response")
		}
		_, _ = h.PaymentLogsService.Create(ctx, bodyString, headerString, string(respString), strconv.Itoa(tixErr.Code), tixErr.Error())
		return
	}
	headerString := ctx.GetString("headers")
	bodyString := ctx.GetString("rawPayload")
	respString, err := json.Marshal(res)
	if err != nil {
		log.Error().Err(err).Msg("Failed to marshal response")
	}
	_, _ = h.PaymentLogsService.Create(ctx, bodyString, headerString, string(respString), "", "")
	ctx.JSON(http.StatusOK, res)
}

func (h *EventTransactionHandlerImpl) CallbackQRISPaylabsV2(ctx *gin.Context) {
	// Implement the callback logic here

	var req dto.QRISCallbackRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		log.Error().Err(err).Msg("error binding JSON for callback")
		lib.RespondError(ctx, http.StatusBadRequest, "invalid request body", err, lib.ErrorBadRequest.Code, h.Env.App.Debug)
		return
	}

	log.Info().Msgf("Received callback: %+v", req)

	// This is a placeholder for the actual implementation
	res, err := h.EventTransactionService.CallbackQRISPaylabsV2(ctx, req)
	var tixErr *lib.TIXError
	if errors.As(err, &tixErr) {
		switch *tixErr {
		case lib.ErrorOrderNotFound:
			lib.RespondError(ctx, http.StatusNotFound, tixErr.Error(), err, tixErr.Code, h.Env.App.Debug)
			// return
		case lib.ErrorCallbackSignatureInvalid:
			lib.RespondError(ctx, http.StatusBadRequest, tixErr.Error(), err, tixErr.Code, h.Env.App.Debug)
			// return
		default:
			lib.RespondError(ctx, http.StatusInternalServerError, "error", err, lib.ErrorInternalServer.Code, h.Env.App.Debug)
			// return
		}
		headerString := ctx.GetString("headers")
		bodyString := ctx.GetString("rawPayload")
		respString, err := json.Marshal(res)
		if err != nil {
			log.Error().Err(err).Msg("Failed to marshal response")
		}
		_, _ = h.PaymentLogsService.Create(ctx, bodyString, headerString, string(respString), strconv.Itoa(tixErr.Code), tixErr.Error())
		return
	}
	headerString := ctx.GetString("headers")
	bodyString := ctx.GetString("rawPayload")
	respString, err := json.Marshal(res)
	if err != nil {
		log.Error().Err(err).Msg("Failed to marshal response")
	}
	_, _ = h.PaymentLogsService.Create(ctx, bodyString, headerString, string(respString), "", "")
	ctx.JSON(http.StatusOK, res)
}

// @Summary validate email for booking is used or not
// @Description validate email for booking is used or not
// @Tags events
// @Produce json
// @Param eventId path string false "Event ID"
// @Param email path string false "Email book"
// @Success 200 {object} lib.APIResponse{data=nil} "Success validate order information"
// @Failure 404 {object} lib.HTTPError "Not Found"
// @Failure 500 {object} lib.HTTPError "Internal server error"
// @Router /events/{eventId}/email-books/{email} [get]
func (h *EventTransactionHandlerImpl) IsEmailAlreadyBook(ctx *gin.Context) {
	var uriParams dto.GetValidateEmailForBookEventParams
	if err := ctx.ShouldBindUri(&uriParams); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			// Find first error
			log.Error().Err(err).Msg("error binding uri params")
			fieldErr := validationErrors[0]

			mappedError := lib.MapErrorGetGarudaIDByIdParams(fieldErr)
			if mappedError != nil {
				var tixErr *lib.TIXError
				if errors.As(mappedError, &tixErr) {
					lib.RespondError(ctx, http.StatusBadRequest, tixErr.Error(), tixErr, tixErr.Code, h.Env.App.Debug)
					return
				}
			}

			lib.RespondError(ctx, http.StatusBadRequest, fieldErr.Field()+" is invalid", fieldErr, lib.ErrorBadRequest.Code, h.Env.App.Debug)
			return
		}
		lib.RespondError(ctx, http.StatusBadRequest, "bad request. check your payload", nil, lib.ErrorBadRequest.Code, h.Env.App.Debug)
		return
	}

	err := h.EventTransactionService.ValidateEmailIsAlreadyBook(ctx, uriParams.EventID, uriParams.Email)
	if err != nil {
		var tixErr *lib.TIXError
		if errors.As(err, &tixErr) {
			switch *tixErr {
			case lib.ErrorEventSaleAlreadyOver, lib.ErrorEventSaleIsNotStartedYet, lib.ErrorEventSaleIsPaused:
				lib.RespondError(ctx, http.StatusForbidden, "error", tixErr, tixErr.Code, h.Env.App.Debug)
			case lib.ErrorEventNotFound:
				lib.RespondError(ctx, http.StatusNotFound, "error", tixErr, tixErr.Code, h.Env.App.Debug)
			case lib.ErrorOrderInformationIsAlreadyBook:
				lib.RespondError(ctx, http.StatusConflict, "error", tixErr, tixErr.Code, h.Env.App.Debug)
			default:
				lib.RespondError(ctx, http.StatusInternalServerError, "error", err, lib.ErrorInternalServer.Code, h.Env.App.Debug)
			}
		}

		return
	}

	lib.RespondSuccess(ctx, http.StatusOK, "success", nil)
}

// @Summary Get available payment methods
// @Description Get available payment methods
// @Tags events
// @Produce json
// @Param eventId path string false "Event ID"
// @Success 200 {object} lib.APIResponse{data=nil} "Success get available payment methods"
// @Failure 404 {object} lib.HTTPError "Not Found"
// @Failure 500 {object} lib.HTTPError "Internal server error"
// @Router /events/{eventId}/payment-methods [get]
func (h *EventTransactionHandlerImpl) GetAvailablePaymentMethods(ctx *gin.Context) {
	var uriParams dto.GetAvailablePaymentMethodParams
	if err := ctx.ShouldBindUri(&uriParams); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			// Find first error
			log.Error().Err(err).Msg("error binding uri params")
			fieldErr := validationErrors[0]

			mappedError := lib.MapErrorGetGarudaIDByIdParams(fieldErr)
			if mappedError != nil {
				var tixErr *lib.TIXError
				if errors.As(mappedError, &tixErr) {
					lib.RespondError(ctx, http.StatusBadRequest, tixErr.Error(), tixErr, tixErr.Code, h.Env.App.Debug)
					return
				}
			}

			lib.RespondError(ctx, http.StatusBadRequest, fieldErr.Field()+" is invalid", fieldErr, lib.ErrorBadRequest.Code, h.Env.App.Debug)
			return
		}
		lib.RespondError(ctx, http.StatusBadRequest, "bad request. check your payload", nil, lib.ErrorBadRequest.Code, h.Env.App.Debug)
		return
	}

	res, err := h.EventTransactionService.GetAvailablePaymentMethods(ctx, uriParams.EventID)
	if err != nil {
		lib.RespondError(ctx, http.StatusInternalServerError, "error", nil, lib.ErrorInternalServer.Code, h.Env.App.Debug)
		return
	}

	lib.RespondSuccess(ctx, http.StatusOK, "success", res)
}

// @Summary Get transaction details by ID
// @Description Get transaction details by ID
// @Tags events
// @Produce json
// @Param transactionId path string true "Transaction ID"
// @Success 200 {object} lib.APIResponse{data=entity.OrderDetails} "Success"
// @Failure 400 {object} lib.HTTPError "Invalid request body"
// @Failure 404 {object} lib.HTTPError "Transaction not found"
// @Failure 500 {object} lib.HTTPError "Internal server error"
// @Security BearerAuth
// @Router /events/transactions/{transactionId} [get]
func (h *EventTransactionHandlerImpl) GetTransactionDetails(ctx *gin.Context) {
	var uriParams dto.GetTransactionDetails
	if err := ctx.ShouldBindUri(&uriParams); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			fieldErr := validationErrors[0]
			lib.RespondError(ctx, http.StatusBadRequest, fieldErr.Field()+" is invalid", fieldErr, lib.ErrorBadRequest.Code, h.Env.App.Debug)
			return
		}
		lib.RespondError(ctx, http.StatusBadRequest, "bad request. check your payload", nil, lib.ErrorBadRequest.Code, h.Env.App.Debug)
		return
	}
	bearerTransactionID := ctx.GetString("transaction_id")
	if bearerTransactionID != uriParams.TransactionID {
		lib.RespondError(ctx, http.StatusForbidden, "you are not allowed to access this transaction", nil, lib.MissmatchTxIDParameterBearerError.Code, h.Env.App.Debug)
		return
	}
	res, err := h.EventTransactionService.FindById(ctx, uriParams.TransactionID)
	if err != nil {
		log.Error().Err(err).Msg("error finding transaction by id")
		var tixErr *lib.TIXError
		if errors.As(err, &tixErr) {
			switch *tixErr {
			case lib.ErrorTransactionDetailsNotFound:
				lib.RespondError(ctx, http.StatusNotFound, "transaction not found", tixErr, tixErr.Code, h.Env.App.Debug)
			default:
				lib.RespondError(ctx, http.StatusInternalServerError, "error", err, lib.ErrorInternalServer.Code, h.Env.App.Debug)
			}
		} else {
			lib.RespondError(ctx, http.StatusInternalServerError, "error", err, lib.ErrorInternalServer.Code, h.Env.App.Debug)
		}
		return
	}
	log.Info().Interface("transactionDetails", res).Msg("found transaction details by id")
	lib.RespondSuccess(ctx, http.StatusOK, "success", res)
}
