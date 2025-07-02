package mailer

type Publisher interface {
	Publish(subject string, data interface{}, metadata map[string]interface{}) error
}
