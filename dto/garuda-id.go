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
	GarudaID string `uri:"garudaId" binding:"required,min=1,uuid"`
	EventID  string `uri:"eventId" binding:"required,min=1,uuid"`
}
type VerifyGarudaIDResponse struct {
	GarudaID    string `json:"garuda_id"`
	IsAvailable bool   `json:"is_available"`
}

type ApiResponseGarudaIDService struct {
	StatusCode int                   `json:"status_code"`
	Success    bool                  `json:"success"`
	Message    string                `json:"message"`
	Data       RequestFansIDResponse `json:"data,omitempty"`
	ErrorCode  int                   `json:"error_code,omitempty"`
}

type RequestFansIDResponse struct {
	Name        string `json:"name"`
	Email       string `json:"email"`
	FansID      string `json:"fans_id"`
	IsAvailable bool   `json:"is_available"`
	Age         int    `json:"age"`
	PhoneNumber string `json:"phone_number"`
}
