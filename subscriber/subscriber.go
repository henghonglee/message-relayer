package subscriber

import (
	"fmt"

	"example.com/messagerelay/message"
	"example.com/messagerelay/relayer"
)

type MessageSubscriber interface {
	Subscribe(r relayer.MessageRelayer, t message.MessageType, done chan bool)
}

type subscriberImpl struct {
	inbound chan message.Message
}

func NewMessageSubscriber() *subscriberImpl {
	return &subscriberImpl{
		inbound: make(chan message.Message, 10),
	}
}

func (s *subscriberImpl) Subscribe(r relayer.MessageRelayer, t message.MessageType, done chan bool) {
	r.SubscribeToMessages(t, s.inbound)
	go func() {
		for {
			select {
			case msg := <-s.inbound:
				fmt.Printf("Subscriber[%p] Recv %v\n", s, msg)
			case <-done:
				fmt.Println("Subscriber Stopped")
				return
			}

		}
	}()
}
