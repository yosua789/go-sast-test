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
)

var (
	ErrorForbidden = TIXError{
		Code: 40302,
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
		Code: 40006,
		Err:  errors.New("max reach page"),
	}
)

var (
	ErrorNotImplemented = TIXError{
		Code: 50099,
		Err:  errors.New("internal server error"),
	}
)
