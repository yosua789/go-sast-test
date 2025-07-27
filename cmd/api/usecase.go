package api

import (
	"assist-tix/config"
	"assist-tix/internal/domain"
	"assist-tix/internal/usecase"
)

type UseCase struct {
	TransactionUseCase usecase.TransactionUsecase
}

func NewUseCase(
	env *config.EnvironmentVariable,
	publisher domain.EventPublisher,
) UseCase {
	return UseCase{
		TransactionUseCase: usecase.NewTransactionUsecase(env, publisher),
	}
}
