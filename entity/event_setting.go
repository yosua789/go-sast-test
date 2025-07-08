package entity

import "time"

type EventSetting struct {
	ID           string
	Setting      Setting
	SettingValue string

	CreatedAt time.Time
	UpdatedAt *time.Time
}
