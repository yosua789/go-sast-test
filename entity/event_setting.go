package entity

import "time"

type EventSetting struct {
	ID           int
	Setting      Setting
	SettingValue string

	CreatedAt time.Time
	UpdatedAt *time.Time
}
