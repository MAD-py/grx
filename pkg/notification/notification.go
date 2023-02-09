package notification

type Notifier struct {
	Sender   chan ServerAction
	Receiver chan ServerState
}

type Subscriber struct {
	Sender   chan ServerState
	Receiver chan ServerAction
}

func NewNotifierAndSubscriber() (*Notifier, *Subscriber) {
	action := make(chan ServerAction)
	state := make(chan ServerState)

	notifer := &Notifier{
		Sender:   action,
		Receiver: state,
	}
	subscriber := &Subscriber{
		Sender:   state,
		Receiver: action,
	}
	return notifer, subscriber
}
