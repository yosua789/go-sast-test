package repository

import (
	"assist-tix/config"
	"assist-tix/database"
	"assist-tix/lib"
	"assist-tix/model"
	"context"
	"encoding/json"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/rs/zerolog/log"
)

type PaymentMethodRepository interface {
	GetGrouppedActivePaymentMethod(ctx context.Context, tx pgx.Tx) (grouppedPayments map[string][]model.PaymentMethod, err error)
	ValidatePaymentCodeIsActive(ctx context.Context, tx pgx.Tx, paymentCode string) (paymentMethod model.PaymentMethod, err error)
	FindPaymentMethodByPaymentCode(ctx context.Context, tx pgx.Tx, paymentCode string) (paymentMethod model.PaymentMethod, err error)
}

type PaymentMethodRepositoryImpl struct {
	Env             *config.EnvironmentVariable
	RedisRepository RedisRepository
	WrapDB          *database.WrapDB
}

func NewPaymentMethodRepository(
	wrapDB *database.WrapDB,
	redisRepo RedisRepository,
	env *config.EnvironmentVariable,
) PaymentMethodRepository {
	return &PaymentMethodRepositoryImpl{
		Env:             env,
		RedisRepository: redisRepo,
		WrapDB:          wrapDB,
	}
}

func (r *PaymentMethodRepositoryImpl) GetGrouppedActivePaymentMethod(ctx context.Context, tx pgx.Tx) (grouppedPayments map[string][]model.PaymentMethod, err error) {
	ctx, cancel := context.WithTimeout(ctx, r.Env.Database.Timeout.Read)
	defer cancel()

	val, err := r.RedisRepository.GetState(ctx, "grouppedPayments")
	if err == nil {

		err = json.Unmarshal([]byte(val), &grouppedPayments)
		if err != nil {
			log.Warn().Err(err).Msg("Error Marshal data grouppedpayments from redis")
		} else {
			log.Info().Msg("using data grouppedPayment from redis")
			return grouppedPayments, nil
		}

	} else {
		log.Warn().Err(err).Msg("Not Found on Redis, using postgresql instead")
	}

	grouppedPayments = make(map[string][]model.PaymentMethod, 0)

	query := `SELECT
		id, 
		
		name,
		logo,

		is_paused,
		pause_message,
		paused_at,

		payment_type,
		payment_group,
		payment_code,
		payment_channel,

		created_at,
		paused_at,
		is_percentage,
		additional_fee
	FROM payment_methods
	WHERE is_active = true
	ORDER BY created_at ASC`

	var rows pgx.Rows
	if tx != nil {
		rows, err = tx.Query(ctx, query)
	} else {
		rows, err = r.WrapDB.Postgres.Query(ctx, query)
	}
	if err != nil {
		return
	}

	for rows.Next() {
		var paymentMethod model.PaymentMethod

		rows.Scan(
			&paymentMethod.ID,
			&paymentMethod.Name,
			&paymentMethod.Logo,
			&paymentMethod.IsPaused,
			&paymentMethod.PauseMessage,
			&paymentMethod.PausedAt,
			&paymentMethod.PaymentType,
			&paymentMethod.PaymentGroup,
			&paymentMethod.PaymentCode,
			&paymentMethod.PaymentChannel,
			&paymentMethod.CreatedAt,
			&paymentMethod.UpdatedAt,
			&paymentMethod.IsPercentage,
			&paymentMethod.AdditionalFee,
		)

		_, ok := grouppedPayments[paymentMethod.PaymentGroup]
		if ok {
			grouppedPayments[paymentMethod.PaymentGroup] = append(grouppedPayments[paymentMethod.PaymentGroup], model.PaymentMethod{
				ID:   paymentMethod.ID,
				Name: paymentMethod.Name,
				Logo: paymentMethod.Logo,

				IsPaused:     paymentMethod.IsPaused,
				PauseMessage: paymentMethod.PauseMessage,
				PausedAt:     paymentMethod.PausedAt,

				PaymentType:    paymentMethod.PaymentType,
				PaymentGroup:   paymentMethod.PaymentGroup,
				PaymentCode:    paymentMethod.PaymentCode,
				PaymentChannel: paymentMethod.PaymentChannel,
				IsPercentage:   paymentMethod.IsPercentage,
				AdditionalFee:  paymentMethod.AdditionalFee,
			})
		} else {
			grouppedPayments[paymentMethod.PaymentGroup] = append(grouppedPayments[paymentMethod.PaymentGroup], model.PaymentMethod{
				ID:   paymentMethod.ID,
				Name: paymentMethod.Name,
				Logo: paymentMethod.Logo,

				IsPaused:     paymentMethod.IsPaused,
				PauseMessage: paymentMethod.PauseMessage,
				PausedAt:     paymentMethod.PausedAt,

				PaymentType:    paymentMethod.PaymentType,
				PaymentGroup:   paymentMethod.PaymentGroup,
				PaymentCode:    paymentMethod.PaymentCode,
				PaymentChannel: paymentMethod.PaymentChannel,
				IsPercentage:   paymentMethod.IsPercentage,
				AdditionalFee:  paymentMethod.AdditionalFee,
			})
		}
	}
	jsonData, err := json.Marshal(grouppedPayments)
	if err != nil {
		log.Error().Err(err).Msg("Failed to marshalling paymentMethod")
	} else {
		r.RedisRepository.SetState(ctx, "grouppedPayments", string(jsonData), 15)
	}
	return
}

func (r *PaymentMethodRepositoryImpl) ValidatePaymentCodeIsActive(ctx context.Context, tx pgx.Tx, paymentCode string) (paymentMethod model.PaymentMethod, err error) {
	ctx, cancel := context.WithTimeout(ctx, r.Env.Database.Timeout.Read)
	defer cancel()
	val, err := r.RedisRepository.GetState(ctx, "paymentMethod"+paymentCode)
	if err == nil {
		err = json.Unmarshal([]byte(val), &paymentMethod)
		if err != nil {
			log.Warn().Err(err).Msg("Error Marshal data paymentMethod from redis")
		} else {
			log.Info().Msg("using data grouppedPayment from redis")
			return paymentMethod, nil
		}
	} else {
		log.Warn().Err(err).Msg("Not Found on Redis, using postgresql instead")
	}

	query := `SELECT
		id, 
		
		name,
		logo,

		is_paused,
		pause_message,
		paused_at,

		payment_type,
		payment_group,
		payment_code,
		payment_channel,

		created_at,
		paused_at,
		is_percentage,
		additional_fee
	FROM payment_methods
	WHERE payment_code = $1 
		AND is_active = true
		AND is_paused = false
	LIMIT 1`

	if tx != nil {
		err = tx.QueryRow(ctx, query, paymentCode).Scan(
			&paymentMethod.ID,
			&paymentMethod.Name,
			&paymentMethod.Logo,
			&paymentMethod.IsPaused,
			&paymentMethod.PauseMessage,
			&paymentMethod.PausedAt,
			&paymentMethod.PaymentType,
			&paymentMethod.PaymentGroup,
			&paymentMethod.PaymentCode,
			&paymentMethod.PaymentChannel,
			&paymentMethod.CreatedAt,
			&paymentMethod.UpdatedAt,
			&paymentMethod.IsPercentage,
			&paymentMethod.AdditionalFee,
		)
	} else {
		err = r.WrapDB.Postgres.QueryRow(ctx, query, paymentCode).Scan(
			&paymentMethod.ID,
			&paymentMethod.Name,
			&paymentMethod.Logo,
			&paymentMethod.IsPaused,
			&paymentMethod.PauseMessage,
			&paymentMethod.PausedAt,
			&paymentMethod.PaymentType,
			&paymentMethod.PaymentGroup,
			&paymentMethod.PaymentCode,
			&paymentMethod.PaymentChannel,
			&paymentMethod.CreatedAt,
			&paymentMethod.UpdatedAt,
			&paymentMethod.IsPercentage,
			&paymentMethod.AdditionalFee,
		)
	}

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return paymentMethod, &lib.ErrorPaymentMethodInvalid
		}

		return paymentMethod, err
	}
	jsonData, err := json.Marshal(paymentMethod)
	if err != nil {
		log.Error().Err(err).Msg("Failed to marshalling paymentMethod")
	} else {
		r.RedisRepository.SetState(ctx, "paymentMethod"+paymentCode, string(jsonData), 15)
	}

	return paymentMethod, nil
}

func (r *PaymentMethodRepositoryImpl) FindPaymentMethodByPaymentCode(ctx context.Context, tx pgx.Tx, paymentCode string) (paymentMethod model.PaymentMethod, err error) {
	ctx, cancel := context.WithTimeout(ctx, r.Env.Database.Timeout.Read)
	defer cancel()

	query := `SELECT
		id, 
		
		name,
		logo,

		is_paused,
		pause_message,
		paused_at,

		payment_type,
		payment_group,
		payment_code,
		payment_channel,

		created_at,
		paused_at
	FROM payment_methods
	WHERE payment_code = $1 
		AND is_active = true
		AND is_paused = false
	LIMIT 1`

	if tx != nil {
		err = tx.QueryRow(ctx, query, paymentCode).Scan(
			&paymentMethod.ID,
			&paymentMethod.Name,
			&paymentMethod.Logo,
			&paymentMethod.IsPaused,
			&paymentMethod.PauseMessage,
			&paymentMethod.PausedAt,
			&paymentMethod.PaymentType,
			&paymentMethod.PaymentGroup,
			&paymentMethod.PaymentCode,
			&paymentMethod.PaymentChannel,
			&paymentMethod.CreatedAt,
			&paymentMethod.UpdatedAt,
		)
	} else {
		err = r.WrapDB.Postgres.QueryRow(ctx, query, paymentCode).Scan(
			&paymentMethod.ID,
			&paymentMethod.Name,
			&paymentMethod.Logo,
			&paymentMethod.IsPaused,
			&paymentMethod.PauseMessage,
			&paymentMethod.PausedAt,
			&paymentMethod.PaymentType,
			&paymentMethod.PaymentGroup,
			&paymentMethod.PaymentCode,
			&paymentMethod.PaymentChannel,
			&paymentMethod.CreatedAt,
			&paymentMethod.UpdatedAt,
		)
	}

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return paymentMethod, &lib.ErrorPaymentMethodInvalid
		}

		return paymentMethod, err
	}

	return paymentMethod, nil
}
