package dto

// response from external API
type GarudaIDApiResponse struct {
	Success   bool                    `json:"success"`
	ErrorCode string                  `json:"error_code,omitempty"`
	Error     interface{}             `json:"error,omitempty"`
	Data      DataGarudaIDAPIResponse `json:"data,omitempty"`
	Message   string                  `json:"message,omitempty"`
}

type DataGarudaIDAPIResponse struct {
	GarudaID    string `json:"garuda_id"`
	Name        string `json:"name"`
	IsAvailable bool   `json:"is_available"`
}

type GetGarudaIDByIdParams struct {
	GarudaID string `uri:"garudaID" binding:"required,min=1,uuid"`
	EventID  string `uri:"eventID" binding:"required,min=1,uuid"`
}
