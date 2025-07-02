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

type SimpleOrganizerResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Logo string `json:"logo"`
}

type CreateOrganizerRequest struct {
	Name string                `form:"name" binding:"required,max=255"`
	Slug string                `form:"slug" binding:"required,max=255"`
	Logo *multipart.FileHeader `form:"logo" binding:"required"`
}

type UpdateOrganizerRequest struct {
	Name string `form:"name" binding:"required,max=255"`
	Slug string `form:"slug" binding:"required,max=255"`
}

type GetOrganizerByIdParams struct {
	OrganizerId string `uri:"organizerId" binding:"required,min=1,uuid"`
}
