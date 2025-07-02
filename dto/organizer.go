package dto

import (
	"mime/multipart"
	"time"
)

type OrganizerResponse struct {
	ID        string     `json:"id"`
	Name      string     `json:"name"`
	Slug      string     `json:"slug"`
	Logo      string     `json:"logo"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt *time.Time `json:"updated_at"`
}

type CreateOrganizerRequest struct {
	Name string                `form:"name" binding:"required"`
	Slug string                `form:"slug" binding:"required"`
	Logo *multipart.FileHeader `form:"logo" binding:"required"`
}

type GetOrganizerByIdParams struct {
	OrganizerId string `uri:"organizerId" binding:"required,min=1,uuid"`
}
