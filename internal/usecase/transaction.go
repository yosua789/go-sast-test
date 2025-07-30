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
			DisplayName: "Mandiri Virtual Account",
			Code:        "CODE",
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
		EventTime: event.EventTime,
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
