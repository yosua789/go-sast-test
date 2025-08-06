package usecase

import (
	"assist-tix/config"
	"assist-tix/entity"
	"assist-tix/internal/domain"
	"assist-tix/internal/domain/async_callback"
	"assist-tix/internal/domain/async_order"
	domainEvent "assist-tix/internal/domain/event"
	"assist-tix/model"
	"context"
	"encoding/json"
	"time"

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

func (u *TransactionUsecase) SendAsyncOrder(
	ctx context.Context,
	useGarudaId bool, // dkirim
	itemCount int, // dikirim
	trxAccessToken string, // dikirim
	paymentMethod model.PaymentMethod, //
	event model.Event, // dikirim
	transaction model.EventTransaction, // dikirim
	ticketCategory model.EventTicketCategory,
	venueSector entity.VenueSector,
	eventTransactionItems []model.EventTransactionItem,
	clientIP string,
	orderInformationBookID int,
) (err error) {
	jsonData := async_order.AsyncOrder{
		UseGarudaId:            useGarudaId,
		ItemCount:              itemCount,
		TransactionAccessToken: trxAccessToken,
		PaymentMethod:          paymentMethod,
		Event:                  event,
		Transaction:            transaction,
		TicketCategory:         ticketCategory,
		VenueSector:            venueSector,
		EventTransactionItem:   eventTransactionItems,
		ClientIP:               clientIP,
		OrderInformationBookID: orderInformationBookID,
	}
	log.Info().Interface("data", jsonData).Msg("payload")

	bytes, err := json.Marshal(jsonData)
	if err != nil {
		return
	}

	err = u.EventPublisher.Publish(ctx, u.Env.Nats.Subjects.AsyncOrder, bytes)
	if err != nil {
		return
	}

	log.Info().Msg("success send email")

	return
}

func (u *TransactionUsecase) SendAsyncCallback(
	ctx context.Context,
	transactionId string,
) (err error) {
	jsonData := async_callback.AsyncCallback{
		TransactionId: transactionId,
		CallbackTime:  time.Now(),
	}
	log.Info().Interface("data", jsonData).Msg("payload")

	bytes, err := json.Marshal(jsonData)
	if err != nil {
		return
	}

	err = u.EventPublisher.Publish(ctx, u.Env.Nats.Subjects.AsyncOrder, bytes)
	if err != nil {
		return
	}

	log.Info().Msg("success send email")

	return
}

func (u *TransactionUsecase) SendBill(
	ctx context.Context,
	email, name string,
	useGarudaId bool,
	itemCount int,
	trxAccessToken string,
	paymentMethod model.PaymentMethod,
	event model.Event,
	transaction model.EventTransaction,
	ticketCategory model.EventTicketCategory,
	venueSector entity.VenueSector,
) (err error) {
	log.Info().Msg("send email bill")
	var transactionPayload = domainEvent.TransactionBill{
		TransactionID:          transaction.ID,
		OrderNumber:            transaction.OrderNumber,
		TransactionAccessToken: trxAccessToken,
		Payment: domainEvent.PaymentInformation{
			Type:                         paymentMethod.PaymentType,
			Group:                        paymentMethod.PaymentGroup,
			Channel:                      paymentMethod.PaymentChannel,
			Code:                         paymentMethod.PaymentCode,
			PaymentAdditionalInformation: transaction.PaymentAdditionalInfo,
			DisplayName:                  paymentMethod.Name,
			GrandTotal:                   transaction.GrandTotal,
		},
		Status: transaction.Status,
		DetailInformation: domainEvent.DetailInformationTransaction{
			BookEmail: email,
			BookName:  name,

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
			ID:             event.ID,
			BannerFilename: event.Banner,
			Name:           event.Name,
			Time:           event.EventTime,
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
	useGarudaId bool,
	itemCount int,
	additionalFees []entity.AdditionalFee,
	transactionDetail entity.EventTransaction,
	paidAt time.Time, // this is the time when the transaction was paid
) (err error) {
	log.Info().Msg("send email invoice")
	log.Info().Msgf("%v ini paid at", transactionDetail.PaidAt)
	invoiceAdditionalFees := make([]domainEvent.AdditionalFee, 0)
	for _, val := range additionalFees {
		invoiceAdditionalFees = append(invoiceAdditionalFees, domainEvent.AdditionalFee{
			Name:         val.Name,
			IsPercentage: val.IsPercentage,
			IsTax:        val.IsTax,
			Value:        val.Value,
		})
	}

	var transactionPayload = domainEvent.TransactionInvoice{
		TransactionID:  transactionDetail.ID,
		OrderNumber:    transactionDetail.OrderNumber,
		AdditionalFees: invoiceAdditionalFees,
		Payment: domainEvent.PaymentInformation{
			DisplayName:                  transactionDetail.PaymentMethod.Name,
			Type:                         transactionDetail.PaymentMethod.PaymentType,
			Group:                        transactionDetail.PaymentMethod.PaymentGroup,
			Channel:                      transactionDetail.PaymentMethod.PaymentChannel,
			Code:                         transactionDetail.PaymentMethod.PaymentCode,
			PaymentAdditionalInformation: transactionDetail.PaymentAdditionalInfo,
			GrandTotal:                   transactionDetail.GrandTotal,
		},
		PaidAt: paidAt,
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
			ID:             transactionDetail.Event.ID,
			BannerFilename: transactionDetail.Event.Banner,
			Name:           transactionDetail.Event.Name,
			Time:           transactionDetail.Event.EventTime,
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
	useGarudaId bool,
	eventTicket model.EventTicket,
	transactionDetail entity.EventTransaction,
) (err error) {
	log.Info().Msg("send email eticket")
	var transactionPayload = domainEvent.TransactionETicket{
		TicketID:      eventTicket.ID,
		TransactionID: transactionDetail.ID,
		TicketNumber:  eventTicket.TicketNumber,
		TicketCode:    eventTicket.TicketCode,

		TicketSeatLabel:  eventTicket.SeatLabel.String,
		TicketSeatRow:    eventTicket.SeatRow,
		TicketSeatColumn: eventTicket.SeatColumn,
		Payment: domainEvent.PaymentInformation{
			// Method:                       transactionDetail.PaymentMethod.PaymentCode,
			DisplayName:                  transactionDetail.PaymentMethod.Name,
			Code:                         transactionDetail.PaymentMethod.PaymentCode,
			PaymentAdditionalInformation: transactionDetail.PaymentAdditionalInfo,
			GrandTotal:                   transactionDetail.GrandTotal,
		},
		DetailInformation: domainEvent.DetailInformationTransaction{
			BookEmail:       eventTicket.TicketOwnerEmail,
			BookName:        eventTicket.TicketOwnerFullname,
			BookPhoneNumber: eventTicket.TicketOwnerPhoneNumber.String,
			BookGarudaID:    eventTicket.TicketOwnerGarudaId.String,
			UseGarudaId:     useGarudaId,
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
			ID:             transactionDetail.Event.ID,
			Name:           transactionDetail.Event.Name,
			BannerFilename: transactionDetail.Event.Banner,
			Time:           transactionDetail.Event.EventTime,
		},
		CreatedAt: transactionDetail.CreatedAt,
	}

	var emailPayload = domainEvent.RequestSendEmail{
		Recipient: domainEvent.Recipient{
			Email: eventTicket.TicketOwnerEmail,
			Name:  eventTicket.TicketOwnerFullname,
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
