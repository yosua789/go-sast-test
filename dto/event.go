package dto

import (
	"time"
)

type EventResponse struct {
	ID          string                  `json:"id"`
	Organizer   SimpleOrganizerResponse `json:"organizer"`
	Name        string                  `json:"name"`
	Description string                  `json:"description"`
	Banner      string                  `json:"banner"`
	EventTime   time.Time               `json:"event_time"`
	Status      string                  `json:"status"`
	Venue       SimpleVenueResponse     `json:"venue"`

	StartSaleAt *time.Time `json:"start_sale_at"`
	EndSaleAt   *time.Time `json:"end_sale_at"`

	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt *time.Time `json:"updated_at"`
}

type DetailEventResponse struct {
	ID          string                  `json:"id"`
	Organizer   SimpleOrganizerResponse `json:"organizer"`
	Name        string                  `json:"name"`
	Description string                  `json:"description"`
	Banner      string                  `json:"banner"`
	EventTime   time.Time               `json:"event_time"`
	Status      string                  `json:"status"`
	Venue       SimpleVenueResponse     `json:"venue"`

	AdditionalInformation string `json:"additional_information"`

	ActiveSettings EventSettings `json:"active_settings"`

	StartSaleAt *time.Time `json:"start_sale_at"`
	EndSaleAt   *time.Time `json:"end_sale_at"`

	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt *time.Time `json:"updated_at"`
}

type EventSettings struct {
	GarudaIdVerification         bool `json:"garuda_id_verification,omitempty"`
	MaxAdultTicketPerTransaction int  `json:"max_adult_ticket_per_transaction,omitempty"`
}

type PaginatedEvents struct {
	Events     []EventResponse `json:"events"`
	Pagination Pagination      `json:"pagination"`
}

type CreateEventRequest struct {
	Type        string    `json:"type" validate:"required,oneof=STADIUM VENUE HALL OTHER"`
	OrganizerID string    `json:"organizer_id" validate:"required,uuid"`
	Name        string    `json:"name" validate:"required,max=500"`
	Description string    `json:"description" validate:"required"`
	Banner      string    `json:"banner" validate:"required"`
	EventTime   time.Time `json:"event_time" validate:"required"`
	Status      string    `json:"status" validate:"required,oneof=UPCOMING CANCELED POSTPONED FINISHED ON_GOING"`
	VenueID     string    `json:"venue_id" validate:"required,uuid"`

	IsActive bool `json:"is_active" validate:"required"`

	StartSaleAt *time.Time `json:"start_sale_at"`
	EndSaleAt   *time.Time `json:"end_sale_at"`
}

type EditEventRequest struct {
	Type        string    `json:"type" validate:"required,oneof=STADIUM VENUE HALL OTHER"`
	OrganizerID string    `json:"organizer_id" validate:"required,uuid"`
	Name        string    `json:"name" validate:"required,max=500"`
	Description string    `json:"description" validate:"required"`
	Banner      string    `json:"banner" validate:"required"`
	EventTime   time.Time `json:"event_time" validate:"required"`
	Status      string    `json:"status" validate:"required,oneof=UPCOMING CANCELED POSTPONED FINISHED ON_GOING"`
	VenueID     string    `json:"venue_id" validate:"required,uuid"`

	IsActive bool `json:"is_active" validate:"required"`

	StartSaleAt *time.Time `json:"start_sale_at"`
	EndSaleAt   *time.Time `json:"end_sale_at"`
}

type GetEventByIdParams struct {
	EventID string `uri:"eventId" binding:"required,min=1,uuid"`
}

type FilterEventRequest struct {
	Search string `form:"search" validate:"omitempty,min=3"`
	Status string `form:"status" validate:"omitempty,oneof=UPCOMING CANCELED POSTPONED FINISHED ON_GOING"`
}
