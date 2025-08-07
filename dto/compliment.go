package dto

type ComplimentTableRequest struct {
	Name                  string `json:"name" binding:"required"` // name for ticket distribution
	Email                 string `json:"email" binding:"required"`
	EventID               string `json:"event_id" binding:"required"`
	EventTicketCategoryID string `json:"event_ticket_category_id" binding:"required"`
	GarudaID              string `json:"garuda_id" binding:"required"` // garuda_id is the fans_id
}

type ComplimentApiRequest struct {
	Name                  string   `json:"name" binding:"required"` // name for ticket distribution
	Email                 string   `json:"email" binding:"required"`
	EventID               string   `json:"event_id" binding:"required"`
	EventTicketCategoryID string   `json:"event_ticket_category_id" binding:"required"`
	GarudaID              []string `json:"garuda_id" binding:"required"` // garuda_id is the fans_id
}
