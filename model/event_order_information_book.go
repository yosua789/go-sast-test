package model

import "time"

type EventOrderInformationBook struct {
	ID        int
	EventID   string
	Email     string
	FullName  string
	CreatedAt time.Time
}
