package usecase

import (
	"assist-tix/config"
	"assist-tix/entity"
	"assist-tix/internal/domain"
	domainEvent "assist-tix/internal/domain/event"
	"assist-tix/model"
	"context"
	"encoding/json"

	"github.com/rs/zerolog/log"
)

type TransactionUsecase struct {
	Env            *config.EnvironmentVariable
	EventPublisher domain.EventPublisher
}

func NewTransactionUsecase(
	env *config.EnvironmentVariable,
	publisher domain.EventPublisher,
) TransactionUsecase {
	return TransactionUsecase{
		Env:            env,
		EventPublisher: publisher,
	}
}

func (u *TransactionUsecase) SendBill(
	ctx context.Context,
	email, name string,
	itemCount int,
	paymentMethod model.PaymentMethod,
	event model.Event,
	transaction model.EventTransaction,
	ticketCategory model.EventTicketCategory,
	venueSector entity.VenueSector,
) (err error) {
	log.Info().Msg("send email bill")
	var transactionPayload = domainEvent.TransactionBill{
		TransactionID: transaction.ID,
		OrderNumber:   transaction.OrderNumber,
		Payment: domainEvent.PaymentInformation{
			Method:      transaction.PaymentMethod,
			DisplayName: paymentMethod.Name,
			Code:        paymentMethod.PaymentCode,
			VANumber:    transaction.PaymentAdditionalInfo,
			GrandTotal:  transaction.GrandTotal,
		},
		Status: transaction.Status,
		DetailInformation: domainEvent.DetailInformationTransaction{
			BookEmail: email,
			TicketCategory: domainEvent.TicketCategoryInformation{
				Code:     ticketCategory.Code,
				Price:    ticketCategory.Price,
				Name:     ticketCategory.Name,
				Entrance: ticketCategory.Entrance,
				Sector: domainEvent.TicketSector{
					Name: venueSector.Name,
				},
			},
			Location: domainEvent.LocationInformation{
				VenueType: venueSector.Venue.VenueType,
				VenueName: venueSector.Venue.Name,
				Country:   venueSector.Venue.Country,
				City:      venueSector.Venue.City,
			},
		},
		Event: domainEvent.EventInformation{
			Name: event.Name,
			Time: event.EventTime,
		},
		ItemCount: itemCount,
		ExpiredAt: transaction.PaymentExpiredAt,
		CreatedAt: transaction.CreatedAt,
	}

	var emailPayload = domainEvent.RequestSendEmail{
		Recipient: domainEvent.Recipient{
			Email: email,
			Name:  name,
		},
		Data: transactionPayload,
	}

	log.Info().Interface("data", emailPayload).Msg("payload")

	bytes, err := json.Marshal(emailPayload)
	if err != nil {
		return
	}

	err = u.EventPublisher.Publish(ctx, u.Env.Nats.Subjects.SendBill, bytes)
	if err != nil {
		return
	}

	log.Info().Msg("success send email")

	return
}

func (u *TransactionUsecase) SendInvoice(
	ctx context.Context,
	email, name string,
	itemCount int,
	transactionDetail entity.EventTransaction,
) (err error) {
	log.Info().Msg("send email invoice")
	var transactionPayload = domainEvent.TransactionInvoice{
		TransactionID: transactionDetail.ID,
		OrderNumber:   transactionDetail.OrderNumber,
		Payment: domainEvent.PaymentInformation{
			Method:      transactionDetail.PaymentMethod.PaymentCode,
			DisplayName: transactionDetail.PaymentMethod.Name,
			Code:        transactionDetail.PaymentMethod.PaymentCode,
			VANumber:    transactionDetail.PaymentAdditionalInfo,
			GrandTotal:  transactionDetail.GrandTotal,
		},
		Status: transactionDetail.Status,
		DetailInformation: domainEvent.DetailInformationTransaction{
			BookEmail: email,
			TicketCategory: domainEvent.TicketCategoryInformation{
				Code:     transactionDetail.TicketCategory.Code,
				Price:    transactionDetail.TicketCategory.Price,
				Name:     transactionDetail.TicketCategory.Name,
				Entrance: transactionDetail.TicketCategory.Entrance,
				Sector: domainEvent.TicketSector{
					Name: transactionDetail.VenueSector.Name,
				},
			},
			Location: domainEvent.LocationInformation{
				VenueType: transactionDetail.Event.Venue.VenueType,
				VenueName: transactionDetail.Event.Venue.Name,
				Country:   transactionDetail.Event.Venue.Country,
				City:      transactionDetail.Event.Venue.City,
			},
		},
		Event: domainEvent.EventInformation{
			Name: transactionDetail.Event.Name,
			Time: transactionDetail.Event.EventTime,
		},
		ItemCount: itemCount,
		ExpiredAt: transactionDetail.PaymentExpiredAt,
		CreatedAt: transactionDetail.CreatedAt,
	}

	var emailPayload = domainEvent.RequestSendEmail{
		Recipient: domainEvent.Recipient{
			Email: email,
			Name:  name,
		},
		Data: transactionPayload,
	}

	log.Info().Interface("data", emailPayload).Msg("payload")

	bytes, err := json.Marshal(emailPayload)
	if err != nil {
		return
	}

	err = u.EventPublisher.Publish(ctx, u.Env.Nats.Subjects.SendInvoice, bytes)
	if err != nil {
		return
	}

	log.Info().Msg("success send email")

	return
}

func (u *TransactionUsecase) SendETicket(
	ctx context.Context,
	email, name string,
	eventTicket model.EventTicket,
	transactionDetail entity.EventTransaction,
) (err error) {
	log.Info().Msg("send email eticket")
	var transactionPayload = domainEvent.TransactionETicket{
		TransactionID: transactionDetail.ID,
		TicketNumber:  eventTicket.TicketNumber,
		Payment: domainEvent.PaymentInformation{
			Method:      transactionDetail.PaymentMethod.PaymentCode,
			DisplayName: transactionDetail.PaymentMethod.Name,
			Code:        transactionDetail.PaymentMethod.PaymentCode,
			VANumber:    transactionDetail.PaymentAdditionalInfo,
			GrandTotal:  transactionDetail.GrandTotal,
		},
		DetailInformation: domainEvent.DetailInformationTransaction{
			BookEmail: email,
			TicketCategory: domainEvent.TicketCategoryInformation{
				Code:     transactionDetail.TicketCategory.Code,
				Price:    transactionDetail.TicketCategory.Price,
				Name:     transactionDetail.TicketCategory.Name,
				Entrance: transactionDetail.TicketCategory.Entrance,
				Sector: domainEvent.TicketSector{
					Name: transactionDetail.VenueSector.Name,
				},
			},
			Location: domainEvent.LocationInformation{
				VenueType: transactionDetail.Event.Venue.VenueType,
				VenueName: transactionDetail.Event.Venue.Name,
				Country:   transactionDetail.Event.Venue.Country,
				City:      transactionDetail.Event.Venue.City,
			},
		},
		Event: domainEvent.EventInformation{
			Name: transactionDetail.Event.Name,
			Time: transactionDetail.Event.EventTime,
		},
		CreatedAt: transactionDetail.CreatedAt,
	}

	var emailPayload = domainEvent.RequestSendEmail{
		Recipient: domainEvent.Recipient{
			Email: email,
			Name:  name,
		},
		Data: transactionPayload,
	}

	log.Info().Interface("data", emailPayload).Msg("payload")

	bytes, err := json.Marshal(emailPayload)
	if err != nil {
		return
	}

	err = u.EventPublisher.Publish(ctx, u.Env.Nats.Subjects.SendETicket, bytes)
	if err != nil {
		return
	}

	log.Info().Msg("success send email")

	return
}
