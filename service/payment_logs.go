package service

import (
	"assist-tix/config"
	"assist-tix/database"
	"assist-tix/model"
	"assist-tix/repository"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

type PaymentLogsService interface {
	Create(ctx *gin.Context, request, header, response, errCode, errResponse string) (res model.PaymentLog, err error)
}

type PaymentLogsServiceImpl struct {
	DB              *database.WrapDB
	Env             *config.EnvironmentVariable
	PaymentLogsRepo repository.PaymentLogRepository
}

func NewPaymentLogsService(
	db *database.WrapDB,
	env *config.EnvironmentVariable,
	paymentLogRepo repository.PaymentLogRepository,
) PaymentLogsService {
	return &PaymentLogsServiceImpl{
		DB:              db,
		Env:             env,
		PaymentLogsRepo: paymentLogRepo,
	}
}

func (s *PaymentLogsServiceImpl) Create(ctx *gin.Context, request, header, response, errCode, errResponse string) (res model.PaymentLog, err error) {
	paymentLog := model.PaymentLog{
		Header:        header,
		Body:          request,
		Response:      response,
		Path:          ctx.FullPath(),
		ErrorCode:     errCode,
		ErrorResponse: errResponse,
	}

	res, err = s.PaymentLogsRepo.Create(ctx, nil, paymentLog)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create payment log")
		return
	}

	return
}
