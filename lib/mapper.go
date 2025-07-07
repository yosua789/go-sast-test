package lib

import (
	"assist-tix/dto"
	"assist-tix/entity"
	"assist-tix/helper"
	"assist-tix/model"
)

func MapOrganizerModelToSimpleResponse(
	organizer model.Organizer,
) dto.SimpleOrganizerResponse {
	return dto.SimpleOrganizerResponse{
		ID:   organizer.ID,
		Name: organizer.Name,
		Slug: organizer.Slug,
		Logo: organizer.Logo,
	}
}

func MapOrganizerEntityToSimpleResponse(
	organizer entity.Organizer,
) dto.SimpleOrganizerResponse {
	return dto.SimpleOrganizerResponse{
		ID:   organizer.ID,
		Name: organizer.Name,
		Slug: organizer.Slug,
		Logo: organizer.Logo,
	}
}

func MapVenueModelToSimpleResponse(
	venue model.Venue,
) dto.SimpleVenueResponse {
	return dto.SimpleVenueResponse{
		ID:        venue.ID,
		VenueType: venue.VenueType,
		Name:      venue.Name,
		Country:   venue.Country,
		City:      venue.City,
	}
}

func MapVenueEntityToSimpleResponse(
	venue entity.Venue,
) dto.SimpleVenueResponse {
	return dto.SimpleVenueResponse{
		ID:        venue.ID,
		VenueType: venue.VenueType,
		Name:      venue.Name,
		Country:   venue.Country,
		City:      venue.City,
	}
}

func MapEventEntityToEventResponse(
	event entity.Event,
) dto.EventResponse {
	return dto.EventResponse{
		ID:          event.ID,
		Organizer:   MapOrganizerEntityToSimpleResponse(event.Organizer),
		Venue:       MapVenueEntityToSimpleResponse(event.Venue),
		Name:        event.Name,
		Description: event.Description,
		Banner:      event.Banner,
		EventTime:   event.EventTime,
		Status:      event.Status,
		StartSaleAt: helper.ConvertNullTimeToPointer(event.StartSaleAt),
		EndSaleAt:   helper.ConvertNullTimeToPointer(event.EndSaleAt),
		CreatedAt:   event.CreatedAt,
		UpdatedAt:   helper.ConvertNullTimeToPointer(event.UpdatedAt),
	}
}
