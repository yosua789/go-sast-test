package model

import "time"

type EventSetting struct {
	ID           string
	SettingID    string
	EventID      string
	SettingValue string
}

type AdditionalFee struct {
	ID           string
	EventID      string
	Name         string
	IsPercentage bool
	IsTax        bool    // if true, this fee is tax and if false then it's an admin
	Value        float64 // if IsPercentage is true, this is a percentage value, otherwise it's a fixed value
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
