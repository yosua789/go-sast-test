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
	Venue       SimpleVenueResponse     `json:"venue"`

	AdditionalInformation string `json:"additional_information"`

	ActiveSettings EventSettingsResponse `json:"active_settings"`

	IsSaleActive bool `json:"is_sale_active"`

	TicketCategories []EventTicketCategoryResponse `json:"ticket_categories"`

	StartSaleAt *time.Time `json:"start_sale_at"`
	EndSaleAt   *time.Time `json:"end_sale_at"`

	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt *time.Time `json:"updated_at"`
}

type EventSettingsResponse struct {
	GarudaIdVerification         bool `json:"garuda_id_verification,omitempty"`
	MaxAdultTicketPerTransaction int  `json:"max_adult_ticket_per_transaction,omitempty"`
}

type EventSettings struct {
	GarudaIdVerification         bool    `json:"garuda_id_verification,omitempty"`
	MaxAdultTicketPerTransaction int     `json:"max_adult_ticket_per_transaction,omitempty"`
	TaxPercentage                float64 `json:"tax_percentage,omitempty"`
	AdminFeePercentage           float64 `json:"admin_percentage,omitempty"`
	AdminFee                     int     `json:"admin_fee,omitempty"`
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

	IsSaleActive bool `json:"is_sale_active" validate:"required"`

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

	IsSaleActive bool `json:"is_sale_active" validate:"required"`

	StartSaleAt *time.Time `json:"start_sale_at"`
	EndSaleAt   *time.Time `json:"end_sale_at"`
}

type GetEventByIdParams struct {
	EventID string `uri:"eventId" binding:"required,min=1,uuid"`
}

type FilterEventRequest struct {
	Search string `form:"search" validate:"omitempty,min=3"`
	Status string `form:"status" validate:"omitempty,oneof=UPCOMING FINISHED"`
}

// ### Ticket category section ###
type EventTicketCategoryResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Price       int    `json:"price"`
}

type DetailEventPublicTicketCategoryResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Price       int    `json:"price"`

	Sector TicketCategorySectorResponse `json:"sector"`

	TotalPublicStock int `json:"total_public_stock"`

	Code     string `json:"code"`
	Entrance string `json:"entrance"`
}

type DetailEventTicketCategoryResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Price       int    `json:"price"`

	TotalStock           int `json:"total_stock"`
	TotalPublicStock     int `json:"total_public_stock"`
	PublicStock          int `json:"public_stock"`
	TotalComplimentStock int `json:"total_compliment_stock"`
	ComplimentStock      int `json:"compliment_stock"`

	Code     string `json:"code"`
	Entrance string `json:"entrance"`

	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt *time.Time `json:"updated_at"`
}

type CreateEventTicketCategoryRequest struct {
	SectorID             string `json:"venue_sector_id" validate:"required,uuid"`
	Name                 string `json:"name" validate:"required,min=3" example:"Ticket Reguler"`
	Description          string `json:"description" validate:"required" example:"Ticket description"`
	Price                int    `json:"price" validate:"required,min=0" example:"100000"`
	TotalStock           int    `json:"total_stock" validate:"required,min=0" example:"10"`
	TotalPublicStock     int    `json:"total_public_stock" validate:"required,min=0" example:"0"`
	PublicStock          int    `json:"public_stock" validate:"required,min=0" example:"0"`
	TotalComplimentStock int    `json:"total_compliment_stock" validate:"required,min=0" example:"0"`
	ComplimentStock      int    `json:"compliment_stock" validate:"required,min=0" example:"0"`
	Code                 string `json:"code" validate:"required,max=255"`
	Entrance             string `json:"entrance" validate:"max=255"`
}

type GetEventTicketCategoryByIdParams struct {
	EventID string `uri:"eventId" binding:"required,min=1,uuid"`
}

type GetDetailEventTicketCategoryByIdParams struct {
	EventID          string `uri:"eventId" binding:"required,min=1,uuid"`
	TicketCategoryId string `uri:"ticketCategoryId" binding:"required,min=1,uuid"`
}
