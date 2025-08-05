package repository

import (
	"assist-tix/config"
	"assist-tix/database"
	"assist-tix/entity"
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/rs/zerolog/log"
)

type EventSettingsRepository interface {
	FindByEventId(ctx context.Context, tx pgx.Tx, eventId string) (res []entity.EventSetting, err error)
	FindAdditionalFee(ctx context.Context, tx pgx.Tx, eventId string) (res []entity.AdditionalFee, err error)
}

type EventSettingsRepositoryImpl struct {
	WrapDB          *database.WrapDB
	RedisRepository RedisRepository
	Env             *config.EnvironmentVariable
}

func NewEventSettingsRepository(
	wrapDB *database.WrapDB,
	redisRepo RedisRepository,
	env *config.EnvironmentVariable,
) EventSettingsRepository {
	return &EventSettingsRepositoryImpl{
		WrapDB:          wrapDB,
		RedisRepository: redisRepo,
		Env:             env,
	}
}
func (r *EventSettingsRepositoryImpl) FindByEventId(ctx context.Context, tx pgx.Tx, eventId string) (res []entity.EventSetting, err error) {
	ctx, cancel := context.WithTimeout(ctx, r.Env.Database.Timeout.Read)
	defer cancel()
	val, err := r.RedisRepository.GetState(ctx, fmt.Sprintf("eventsetting-"+eventId))
	if err == nil {

		err = json.Unmarshal([]byte(val), &res)
		if err != nil {
			log.Warn().Err(err).Msg("Error Marshal data event setting from redis")
		} else {
			log.Info().Msg("using data  event setting from redis")
			return res, nil
		}

	} else {
		log.Warn().Err(err).Msg("Not Found on Redis, using postgresql instead")
	}

	query := `SELECT 
		es.id, 
		es.setting_value,

		s.id as setting_id,
		s.name as setting_name,
		s.default_value as setting_default_value,

		es.created_at, 
		es.updated_at 
	FROM event_settings es
	INNER JOIN settings s
		ON es.setting_id = s.id
	WHERE es.event_id = $1 AND es.deleted_at IS NULL`

	var rows pgx.Rows

	if tx != nil {
		rows, err = tx.Query(ctx, query, eventId)
	} else {
		rows, err = r.WrapDB.Postgres.Query(ctx, query, eventId)
	}

	if err != nil {
		return
	}

	defer rows.Close()

	for rows.Next() {
		var eventSetting entity.EventSetting
		rows.Scan(
			&eventSetting.ID,
			&eventSetting.SettingValue,
			&eventSetting.Setting.ID,
			&eventSetting.Setting.Name,
			&eventSetting.Setting.DefaultValue,
			&eventSetting.CreatedAt,
			&eventSetting.UpdatedAt,
		)
		res = append(res, eventSetting)
	}
	jsonData, err := json.Marshal(res)
	if err != nil {
		log.Error().Err(err).Msg("Failed to marshalling eventSetting")
	} else {
		r.RedisRepository.SetState(ctx, "eventsetting-"+eventId, string(jsonData), 15)
	}
	return
}

func (r *EventSettingsRepositoryImpl) FindAdditionalFee(ctx context.Context, tx pgx.Tx, eventId string) (res []entity.AdditionalFee, err error) {
	ctx, cancel := context.WithTimeout(ctx, r.Env.Database.Timeout.Read)
	defer cancel()
	val, err := r.RedisRepository.GetState(ctx, fmt.Sprintf("eventadditionalfee-"+eventId))
	if err == nil {
		err = json.Unmarshal([]byte(val), &res)
		if err != nil {
			log.Warn().Err(err).Msg("Error Marshal data event additional fee from redis")
		} else {
			log.Info().Msg("using data event additional fee from redis")
			return res, nil
		}
	} else {
		log.Warn().Err(err).Msg("Not Found on Redis, using postgresql instead")
	}
	query := `SELECT
		id,
		event_id,
		name,
		is_percentage,
		is_tax,
		value,
		created_at,
		updated_at
	FROM event_additional_fees
	WHERE event_id = $1 `
	var rows pgx.Rows
	if tx != nil {
		rows, err = tx.Query(ctx, query, eventId)
	} else {
		rows, err = r.WrapDB.Postgres.Query(ctx, query, eventId)
	}
	if err != nil {
		return
	}

	defer rows.Close()

	for rows.Next() {
		var additionalFee entity.AdditionalFee
		rows.Scan(
			&additionalFee.ID,
			&additionalFee.EventID,
			&additionalFee.Name,
			&additionalFee.IsPercentage,
			&additionalFee.IsTax,
			&additionalFee.Value,
			&additionalFee.CreatedAt,
			&additionalFee.UpdatedAt,
		)
		res = append(res, additionalFee)
	}

	if len(res) == 0 {
		log.Info().Msgf("No additional fees found for event ID: %s", eventId)
	}
	jsonData, err := json.Marshal(res)
	if err != nil {
		log.Error().Err(err).Msg("Failed to marshalling eventAdditionalFee")
	} else {
		err = r.RedisRepository.SetState(ctx, "eventadditionalfee-"+eventId, string(jsonData), 15)
		if err != nil {
			log.Error().Err(err).Msg("Failed to cache event additional fees in Redis")
			err = nil
		}
		log.Info().Msg("Cached event additional fees in Redis")
	}

	return
}
