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

// Event status
const (
	EventStatusUpComing  = "UPCOMING"
	EventStatusCanceled  = "CANCELED"
	EventStatusPostponed = "POSTPONED"
	EventStatusFinished  = "FINISHED"
	EventStatusOnGoing   = "ON_GOING"
)
