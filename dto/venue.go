package dto

import "time"

type VenueResponse struct {
	ID        string     `json:"id"`
	VenueType string     `json:"venue_type"`
	Name      string     `json:"name"`
	Country   string     `json:"country"`
	City      string     `json:"city"`
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
	Status    string `json:"status" validate:"required,oneof=ACTIVE INACTIVE DISABLE"`
	Capacity  int    `json:"capacity"`
}

type UpdateVenueRequest struct {
	VenueType string `json:"venue_type" validate:"required,oneof=STADIUM VENUE HALL OTHER"`
	Name      string `json:"name" validate:"required,max=255"`
	Country   string `json:"country" validate:"required,max=255"`
	City      string `json:"city" validate:"required,max=255"`
	Status    string `json:"status" validate:"required,oneof=ACTIVE INACTIVE DISABLE"`
	Capacity  int    `json:"capacity"`
}

type GetVenueByIdParams struct {
	VenueID string `uri:"venueId" binding:"required,min=1,uuid"`
}

type VenueEventTicketCategoryResponse struct {
	Venue            VenueResponse                             `json:"venue"`
	TicketCategories []DetailEventPublicTicketCategoryResponse `json:"ticket_categories"`
}
