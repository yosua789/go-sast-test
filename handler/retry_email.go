package handler

import (
	"assist-tix/config"
	"assist-tix/lib"
	"assist-tix/service"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/rs/zerolog/log"
)

type RetryEmailHandler interface {
	RetryInvoices(ctx *gin.Context)
}

type RetryEmailHandlerImpl struct {
	Env               *config.EnvironmentVariable
	Validator         *validator.Validate
	RetryEmailService service.RetryEmailService
}

func NewRetryEmailHandler(
	env *config.EnvironmentVariable,
	validator *validator.Validate,
	retryEmailService service.RetryEmailService,
) RetryEmailHandler {
	return &RetryEmailHandlerImpl{
		Env:               env,
		Validator:         validator,
		RetryEmailService: retryEmailService,
	}
}

func (h *RetryEmailHandlerImpl) RetryInvoices(ctx *gin.Context) {
	err := h.RetryEmailService.RetryInvoiceEmail(ctx)
	if err != nil {
		log.Error().Err(err).Msg("retry invoice email")
		var tixErr *lib.TIXError
		if errors.As(err, &tixErr) {
			switch *tixErr {
			case lib.ErrorVenueNotFound:
				lib.RespondError(ctx, http.StatusNotFound, "not found", err, lib.ErrorVenueNotFound.Code, h.Env.App.Debug)
			case lib.ErrorVenueIdInvalid:
				lib.RespondError(ctx, http.StatusBadRequest, "invalid", err, lib.ErrorVenueIdInvalid.Code, h.Env.App.Debug)
			default:
				lib.RespondError(ctx, http.StatusInternalServerError, "error retry email invoices", err, lib.ErrorInternalServer.Code, h.Env.App.Debug)
			}
		} else {
			lib.RespondError(ctx, http.StatusInternalServerError, "error retry email invoices", err, lib.ErrorInternalServer.Code, h.Env.App.Debug)
		}
		return
	}

	lib.RespondSuccess(ctx, http.StatusOK, "success", nil)
}
