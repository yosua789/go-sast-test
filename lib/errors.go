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
	ErrorOrderInformationIsAlreadyBook = TIXError{
		Code: 40908,
		Err:  errors.New("email is already booked for this event"),
	}
	ErrorTransactionPaylabs = TIXError{
		Code: 50004,
		Err:  errors.New("failed to create transaction on paylabs, please try again"),
	}
	TransactionWithoutAdultError = TIXError{
		Code: 40914,
		Err:  errors.New("transaction must contain at least one adult ticket"),
	}
)

var (
	ErrorPaymentMethodInvalid = TIXError{
		Code: 40012,
		Err:  errors.New("payment method is invalid"),
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

	// garuda id under review
	ErrorGarudaIDInvalid = TIXError{
		Code: 40909,
		Err:  errors.New("garuda id under review"),
	}
	// garuda id is rejected
	ErrorGarudaIDRejected = TIXError{
		Code: 40910,
		Err:  errors.New("garuda id is rejected"),
	}
	// garuda id is blacklisted
	ErrorGarudaIDBlacklisted = TIXError{
		Code: 40911,
		Err:  errors.New("garuda id is blacklisted"),
	}
	// Garuda ID Already used
	ErrorGarudaIDAlreadyUsed = TIXError{
		Code: 40912,
		Err:  errors.New("garuda id already used on this event"),
	}
	ErrorGetGarudaID = TIXError{
		Code: 50003,
		Err:  errors.New("failed to get garuda id, please try again"),
	}
	ErrorEventNonGarudaID = TIXError{
		Code: 40305,
		Err:  errors.New("event does not require garuda id verification"),
	}
	ErrorDuplicateGarudaIDPayload = TIXError{
		Code: 40913,
		Err:  errors.New("duplicate garuda id found in payload"),
	}
)

// callback
var (
	ErrorCallbackSignatureInvalid = TIXError{
		Code: 40008,
		Err:  errors.New("callback signature is invalid"),
	}
	ErrorOrderNotFound = TIXError{
		Code: 40412,
		Err:  errors.New("order not found"),
	}
	ErrorFailedToMarkTransactionAsSuccess = TIXError{
		Code: 50005,
		Err:  errors.New("failed to mark transaction as success, please try again"),
	}
	ErrorFailedToUpdateVANo = TIXError{
		Code: 50006,
		Err:  errors.New("failed to update va no, please try again"),
	}
)

// transaction details
var (
	ErrorTransactionDetailsNotFound = TIXError{
		Code: 40413,
		Err:  errors.New("transaction details not found"),
	}
	InvalidJWTError = TIXError{
		Code: 40101,
		Err:  errors.New("invalid JWT token"),
	}
	MissmatchTxIDParameterBearerError = TIXError{
		Code: 40302,
		Err:  errors.New("transaction ID in parameter does not match with the one in bearer token"),
	}
)

// Event tickets
var (
	FailedToCreateEventTIcket = TIXError{
		Code: 50007,
		Err:  errors.New("failed to create event ticket"),
	}
	EventTicketNotFound = TIXError{
		Code: 40414,
		Err:  errors.New("event ticket not found"),
	}
)

var (
	ErrorCreatePaymentLog = TIXError{
		Code: 50008,
		Err:  errors.New("failed to create payment log"),
	}
)

var (
	ErrorSeatmapBookEmpty = TIXError{
		Code: 50009,
		Err:  errors.New("seatmaps is empty"),
	}
	ErrorSeatAvailableSeatNotMatcheWithRequestSeats = TIXError{
		Code: 50010,
		Err:  errors.New("count available seat in the sector not match with requested seat"),
	}
)

var (
	ErrorNilTransactionItem = TIXError{
		Code: 40015,
		Err:  errors.New("transaction item is nil"),
	}
)
