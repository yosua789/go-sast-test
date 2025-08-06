package lib

import (
	"assist-tix/dto"
	"assist-tix/entity"
	"assist-tix/helper"
	"assist-tix/model"
	"strconv"

	"github.com/go-playground/validator/v10"
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

func MapVenueModelToVenueResponse(
	venue model.Venue,
) dto.VenueResponse {
	return dto.VenueResponse{
		ID:        venue.ID,
		VenueType: venue.VenueType,
		Capacity:  venue.Capacity,
		Name:      venue.Name,
		Country:   venue.Country,
		City:      venue.City,
		CreatedAt: venue.CreatedAt,
		Image:     venue.Image,
		UpdatedAt: helper.ConvertNullTimeToPointer(venue.UpdatedAt),
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
		StartSaleAt: helper.ConvertNullTimeToPointer(event.StartSaleAt),
		EndSaleAt:   helper.ConvertNullTimeToPointer(event.EndSaleAt),
		CreatedAt:   event.CreatedAt,
		UpdatedAt:   helper.ConvertNullTimeToPointer(event.UpdatedAt),
	}
}

func MapEventSettingEntityToEventSettingResponse(
	eventSettings []entity.EventSetting,
) dto.EventSettingsResponse {
	var res dto.EventSettingsResponse
	res.AdditionalFees = make([]dto.EventAdditionalFeeResponse, 0)

	for _, setting := range eventSettings {
		if setting.Setting.Name == EventGarudaIdVerificationSettingName {
			if setting.SettingValue == SettingsValueBooleanTrue {
				res.GarudaIdVerification = true
				continue
			}
		}
		if setting.Setting.Name == EventPurchaseAdultTicketPerTransactionSettingName {
			val, err := strconv.Atoi(setting.SettingValue)
			if err != nil {
				defaultvalue, _ := strconv.Atoi(setting.Setting.DefaultValue)
				res.MaxAdultTicketPerTransaction = defaultvalue
			}
			res.MaxAdultTicketPerTransaction = val
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

func MapEntitySectorToTicketCategorySectorResponse(
	data entity.Sector,
) dto.TicketCategorySectorResponse {
	return dto.TicketCategorySectorResponse{
		ID:         data.ID,
		Name:       data.Name,
		Color:      data.Color.String,
		AreaCode:   data.AreaCode.String,
		HasSeatmap: data.HasSeatmap,
	}
}

func MapDetailEventPublicTicketCategoryModelToDetailEventPublicTicketCategoryResponse(
	data model.EventTicketCategory,
) dto.DetailEventPublicTicketCategoryResponse {
	return dto.DetailEventPublicTicketCategoryResponse{
		ID:          data.ID,
		Name:        data.Name,
		Description: data.Description,
		Price:       data.Price,
		PublicStock: data.PublicStock,
		Code:        data.Code,
		Entrance:    data.Entrance,
	}
}

func MapEntityTicketCategoryToDetailEventPublicTicketCategoryResponse(
	data entity.TicketCategory,
) dto.DetailEventPublicTicketCategoryResponse {
	return dto.DetailEventPublicTicketCategoryResponse{
		ID:          data.ID,
		Name:        data.Name,
		Sector:      MapEntitySectorToTicketCategorySectorResponse(data.Sector),
		Description: data.Description,
		Price:       data.Price,
		PublicStock: data.PublicStock,
		Code:        data.Code,
		Entrance:    data.Entrance,
	}
}

// Mapping error code
// - Use camel case
// - First character is capitalized
// - Use ID all capitalized not Id
func MapErrorGetDetailEventTicketCategoryByIdParams(fieldErr validator.FieldError) error {
	switch {
	case fieldErr.Field() == "EventID":
		return &ErrorEventIdInvalid
	case fieldErr.Field() == "TicketCategoryID":
		return &ErrorTicketCategoryInvalid
	}
	return nil
}

func MapErrorGetEventByIdParams(fieldErr validator.FieldError) error {
	switch {
	case fieldErr.Field() == "EventID":
		return &ErrorEventIdInvalid
	}

	return nil
}

func MapErrorGetGarudaIDByIdParams(fieldErr validator.FieldError) error {
	switch {
	case fieldErr.Field() == "GarudaID":
		return &ErrorGarudaIDInvalid
	case fieldErr.Field() == "EventID":
		return &ErrorEventIdInvalid
	}

	return nil
}

func MapErrorGetEventTicketCategoryByIdParams(fieldErr validator.FieldError) error {
	switch {
	case fieldErr.Field() == "EventID":
		return &ErrorEventIdInvalid
	}

	return nil
}
