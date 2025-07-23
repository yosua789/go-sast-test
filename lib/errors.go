package lib

import (
	"errors"
)

type TIXError struct {
	Code int
	Err  error
}

func (e *TIXError) Error() string {
	return e.Err.Error()
}

func HSErr(err error) *TIXError {
	var hserr *TIXError
	if errors.As(err, &hserr) {
		return hserr
	}

	return nil
}

var (
	ErrorBadRequest = TIXError{
		Code: 40001,
		Err:  errors.New("bad request"),
	}
	ErrorInternalServer = TIXError{
		Code: 50001,
		Err:  errors.New("internal server error"),
	}
)

var (
	ErrorOrganizerNotFound = TIXError{
		Code: 40401,
		Err:  errors.New("organizer not found"),
	}
	ErrorOrganizerIdInvalid = TIXError{
		Code: 40002,
		Err:  errors.New("organizer id is invalid"),
	}
	ErrorOrganizerNameConflict = TIXError{
		Code: 40901,
		Err:  errors.New("organizer name is already used"),
	}
	ErrorOrganizerPosterSizeExceeds = TIXError{
		Code: 41301,
		Err:  errors.New("organizer poster size exceeds the limit"),
	}
)

var (
	ErrorVenueNotFound = TIXError{
		Code: 40404,
		Err:  errors.New("venue not found"),
	}
	ErrorVenueIdInvalid = TIXError{
		Code: 40004,
		Err:  errors.New("organizer id is invalid"),
	}
	ErrorVenueNameConflict = TIXError{
		Code: 40903,
		Err:  errors.New("venue name is already used"),
	}
)

var (
	ErrorVenueSectorNotFound = TIXError{
		Code: 40406,
		Err:  errors.New("venue sector not found"),
	}
	ErrorVenueSectorIdInvalid = TIXError{
		Code: 40006,
		Err:  errors.New("venue sector id is invalid"),
	}
	ErrorVenueSectorDoesntHaveSeatmap = TIXError{
		Code: 40407,
		Err:  errors.New("venue sector doesn't have seatmap"),
	}
	ErrorVenueSectorNameConflict = TIXError{
		Code: 40904,
		Err:  errors.New("venue sector name is already used"),
	}
)

var (
	ErrorTicketCategoryNotFound = TIXError{
		Code: 40405,
		Err:  errors.New("ticket category not found"),
	}
	ErrorTicketCategoryInvalid = TIXError{
		Code: 40408,
		Err:  errors.New("ticket category invalid"),
	}
)

var (
	ErrorEventNotFound = TIXError{
		Code: 40402,
		Err:  errors.New("event not found"),
	}
	ErrorEventIdInvalid = TIXError{
		Code: 40003,
		Err:  errors.New("event id is invalid"),
	}
	ErrorEventNameConflict = TIXError{
		Code: 40902,
		Err:  errors.New("event name is already used"),
	}
	ErrorEventPosterSizeExceeds = TIXError{
		Code: 41302,
		Err:  errors.New("event poster size exceeds the limit"),
	}
	ErrorEventSaleIsPaused = TIXError{
		Code: 40302,
		Err:  errors.New("event ticket sale is paused"),
	}
	ErrorEventSaleIsNotStartedYet = TIXError{
		Code: 40303,
		Err:  errors.New("event ticket sale is not started yet"),
	}
	ErrorEventSaleAlreadyOver = TIXError{
		Code: 40304,
		Err:  errors.New("event ticket sale is already over"),
	}
)

var (
	ErrorEventSettingNotFound = TIXError{
		Code: 40409,
		Err:  errors.New("setting not found"),
	}
)

var (
	ErrorForbidden = TIXError{
		Code: 40301,
		Err:  errors.New("resource forbidden"),
	}
)

var (
	ErrorFileNotFound = TIXError{
		Code: 40403,
		Err:  errors.New("file not found"),
	}
)

var (
	ErrorPaginationPageIsInvalid = TIXError{
		Code: 40005,
		Err:  errors.New("page invalid"),
	}
	ErrorPaginationReachMaxPage = TIXError{
		Code: 40007,
		Err:  errors.New("max reach page"),
	}
)

var (
	ErrorNotImplemented = TIXError{
		Code: 50099,
		Err:  errors.New("internal server error"),
	}
)

var (
	ErrorFailedToCreateTransaction = TIXError{
		Code: 50002,
		Err:  errors.New("failed to create transaction, please try again"),
	}
	ErrorTicketIsOutOfStock = TIXError{
		Code: 40906,
		Err:  errors.New("ticket out of stock"),
	}
	ErrorPurchaseQuantityExceedTheLimit = TIXError{
		Code: 40907,
		Err:  errors.New("purchased items exceed the purchase limit"),
	}
	ErrorEmailIsAlreadyBooked = TIXError{
		Code: 40908,
		Err:  errors.New("email is already booked for this event"),
	}
)

var (
	ErrorBookedSeatNotFound = TIXError{
		Code: 40410,
		Err:  errors.New("booked seat not found"),
	}
	ErrorSeatIsAlreadyBooked = TIXError{
		Code: 40905,
		Err:  errors.New("seat is already booked"),
	}
	ErrorFailedToBookSeat = TIXError{
		Code: 40011,
		Err:  errors.New("failed to book seat"),
	}
)

var (
	ErrorGarudaIDNotFound = TIXError{
		Code: 40411,
		Err:  errors.New("garuda id not found"),
	}
	ErrorGarudaIDInvalid = TIXError{
		Code: 40909,
		Err:  errors.New("garuda id invalid"),
	}
)
