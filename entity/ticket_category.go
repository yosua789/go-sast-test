package entity

type TicketCategory struct {
	ID                   string
	EventID              string
	Sector               Sector
	Name                 string
	Description          string
	Price                int
	TotalStock           int
	TotalPublicStock     int
	PublicStock          int
	TotalComplimentStock int
	ComplimentStock      int
	Code                 string
	Entrance             string
}
