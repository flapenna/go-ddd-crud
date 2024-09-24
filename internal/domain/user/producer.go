package domain

type UserProducer interface {
	SendMessage(event *UserEvent) error
}
