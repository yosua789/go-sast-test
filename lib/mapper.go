package lib

import (
	"assist-tix/dto"
	"assist-tix/entity"
	"assist-tix/helper"
	"assist-tix/model"
	"strconv"
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

func MapEventSettingEntityToEventSettingResponse(
	eventSettings []entity.EventSetting,
) dto.EventSettings {
	var res dto.EventSettings

	for _, setting := range eventSettings {
		if setting.Setting.Name == EventGarudaIdVerificationSettingName {
			if setting.Setting.Type == SettingsTypeBoolean && setting.SettingValue == SettingsValueBooleanTrue {
				res.GarudaIdVerification = true
				continue
			}
		}
		if setting.Setting.Name == EventPurchaseAdultTicketPerTransactionSettingName {
			defaultvalue, _ := strconv.Atoi(setting.Setting.DefaultValue)
			res.MaxAdultTicketPerTransaction = defaultvalue

			if setting.Setting.Type == SettingsTypeInteger {
				val, err := strconv.Atoi(setting.SettingValue)
				if err == nil {
					res.MaxAdultTicketPerTransaction = val
				}
			}
		}
	}

	return res
}

func MapEventTicketCategoryModelToEventTicketCategoryResponse(
	data model.EventTicketCategory,
) dto.EventTicketCategoryResponse {
	return dto.EventTicketCategoryResponse{
		ID:          data.ID,
		Name:        data.Name,
		Description: data.Description,
		Price:       data.Price,
	}
}

func MapEventTicketCategoryModelToDetailEventTicketCategoryResponse(
	data model.EventTicketCategory,
) dto.DetailEventTicketCategoryResponse {
	return dto.DetailEventTicketCategoryResponse{
		ID:          data.ID,
		Name:        data.Name,
		Description: data.Description,
		Price:       data.Price,

		TotalStock:           data.TotalStock,
		TotalPublicStock:     data.TotalPublicStock,
		PublicStock:          data.PublicStock,
		TotalComplimentStock: data.TotalComplimentStock,
		ComplimentStock:      data.ComplimentStock,
		Code:                 data.Code,
		Entrance:             data.Entrance,
		CreatedAt:            data.CreatedAt,
		UpdatedAt:            helper.ConvertNullTimeToPointer(data.UpdatedAt),
	}
}
