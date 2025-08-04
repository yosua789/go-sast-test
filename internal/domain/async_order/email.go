package async_order

type RequestSendEmail struct {
	Recipient Recipient   `json:"recipient"`
	Data      interface{} `json:"data"`
}

type Recipient struct {
	Email string `json:"email"`
	Name  string `json:"name"`
}
