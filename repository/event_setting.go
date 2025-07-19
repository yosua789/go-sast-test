package repository

import (
	"assist-tix/config"
	"assist-tix/database"
	"assist-tix/entity"
	"context"

	"github.com/jackc/pgx/v5"
)

type EventSettingsRepository interface {
	FindByEventId(ctx context.Context, tx pgx.Tx, eventId string) (res []entity.EventSetting, err error)
}

type EventSettingsRepositoryImpl struct {
	WrapDB *database.WrapDB
	Env    *config.EnvironmentVariable
}

func NewEventSettingsRepository(
	wrapDB *database.WrapDB,
	env *config.EnvironmentVariable,
) EventSettingsRepository {
	return &EventSettingsRepositoryImpl{
		WrapDB: wrapDB,
		Env:    env,
	}
}
func (r *EventSettingsRepositoryImpl) FindByEventId(ctx context.Context, tx pgx.Tx, eventId string) (res []entity.EventSetting, err error) {
	ctx, cancel := context.WithTimeout(ctx, r.Env.Database.Timeout.Read)
	defer cancel()

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

	return
}
