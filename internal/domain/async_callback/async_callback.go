package async_callback

import "time"

// AsyncCallback
type AsyncCallback struct {
	TransactionId string    `json:"transaction_id"`
	CallbackTime  time.Time `json:"callback_time"`
}
