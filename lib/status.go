package lib

// Sector status
const (
	SectorStatusActive   = "ACTIVE"
	SectorStatusInactive = "INACTIVE"
	SectorStatusDisabled = "DISABLE"
)

// Venue seatmap status
const (
	SeatmapStatusAvailable   = "AVAILABLE"
	SeatmapStatusUnavailable = "UNAVAILABLE"
	SeatmapStatusDisable     = "DISABLE"
)

// Event venue seatmap status
const (
	EventVenueSeatmapStatusAvailable   = "AVAILABLE"
	EventVenueSeatmapStatusUnavailable = "UNAVAILABLE"
	EventVenueSeatmapStatusPreBooked   = "PREBOOKED"
	EventVenueSeatmapStatusCompliment  = "COMPLIMENT"
	EventVenueSeatmapStatusDisable     = "DISABLE"
)

// Event publish status
const (
	EventPublishStatusDraft     = "DRAFT"
	EventPublishStatusPublished = "PUBLISHED"
	EventPublishStatusPaused    = "PAUSED"
)

// Event status
const (
	EventStatusUpComing  = "UPCOMING"
	EventStatusCanceled  = "CANCELED"
	EventStatusPostponed = "POSTPONED"
	EventStatusFinished  = "FINISHED"
	EventStatusOnGoing   = "ON_GOING"
	EventStatusAll       = "ALL"
)

// Event Transaction status
const (
	EventTransactionStatusPending          = "PENDING"
	EventTransactionStatusSuccess          = "SUCCESS"
	EventTransactionStatusProcessingTicket = "PROCESSING_TICKET"
	EventTransactionStatusExpired          = "EXPIRED"
	EventTransactionStatusFailed           = "FAILED"
)

type EventTicketStatus string

const (
	EventTicketStatusInProgress EventTicketStatus = "IN PROGRESS"
	EventTicketStatusFailed     EventTicketStatus = "FAILED"
	EventTicketStatusSuccess    EventTicketStatus = "SUCCESS"
)
