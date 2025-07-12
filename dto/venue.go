package dto

import "time"

type VenueResponse struct {
	ID        string     `json:"id"`
	VenueType string     `json:"venue_type"`
	Name      string     `json:"name"`
	Country   string     `json:"country"`
	City      string     `json:"city"`
	IsActive  bool       `json:"is_active"`
	Capacity  int        `json:"capacity"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt *time.Time `json:"updated_at"`
}

type SimpleVenueResponse struct {
	ID        string `json:"id"`
	VenueType string `json:"venue_type"`
	Name      string `json:"name"`
	Country   string `json:"country"`
	City      string `json:"city"`
}

type CreateVenueRequest struct {
	VenueType string `json:"venue_type" validate:"required,oneof=STADIUM VENUE HALL OTHER"`
	Name      string `json:"name" validate:"required,max=255"`
	Country   string `json:"country" validate:"required,max=255"`
	City      string `json:"city" validate:"required,max=255"`
	Capacity  int    `json:"capacity" validate:"min=0"`
}

type UpdateVenueRequest struct {
	VenueType string `json:"venue_type" validate:"required,not_blank,oneof=STADIUM VENUE HALL OTHER"`
	Name      string `json:"name" validate:"required,not_blank,max=255"`
	Country   string `json:"country" validate:"required,not_blank,max=255"`
	City      string `json:"city" validate:"required,not_blank,max=255"`
	IsActive  bool   `json:"is_active" validate:"required"`
	Capacity  int    `json:"capacity" validate:"min=0"`
}

type GetVenueByIdParams struct {
	VenueID string `uri:"venueId" binding:"required,min=1,uuid"`
}

type VenueEventTicketCategoryResponse struct {
	Venue            VenueResponse                             `json:"venue"`
	TicketCategories []DetailEventPublicTicketCategoryResponse `json:"ticket_categories"`
}
