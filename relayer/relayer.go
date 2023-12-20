package relayer

import (
	"fmt"
	"sync"
	"time"

	"example.com/messagerelay/message"
	"example.com/messagerelay/network"
)

const (
	MaxReadQueueSize         = 10  // in order to not overflow on memory this is max read queue size
	ReadMessagesDelayMs      = 100 // for testing, time between reads, this throttles the read to make it "slower"
	MessageProcessDurationMs = 100 // after 500 ms, process messages in the queue for 100ms
	MessageProcessIntervalMs = 500 // after 500 ms, process messages in the queue for 100ms
)

type MessageRelayer interface {
	Start()
	SubscribeToMessages(msgType message.MessageType, ch chan message.Message)
}

type messageRelayerImpl struct {
	socket           network.NetworkSocket
	readMessageQueue chan message.Message
	messages         map[message.MessageType][]message.Message // where messages are trimmed and stored before they are broadcast
	subscribers      map[message.MessageType][]chan message.Message
	subscribersLock  sync.RWMutex
}

func NewMessageRelayer(socket network.NetworkSocket) *messageRelayerImpl {
	return &messageRelayerImpl{
		socket:           socket,
		readMessageQueue: make(chan message.Message, MaxReadQueueSize),
		messages:         make(map[message.MessageType][]message.Message),
		subscribers:      make(map[message.MessageType][]chan message.Message),
	}
}

func (m *messageRelayerImpl) SubscribeToMessages(msgType message.MessageType, ch chan message.Message) {
	m.subscribersLock.Lock()
	defer m.subscribersLock.Unlock()
	m.subscribers[msgType] = append(m.subscribers[msgType], ch)
}

func (m *messageRelayerImpl) Start() {
	go m.read()
	go func() {
		for {
			done := make(chan bool)
			// after some X seconds, read all the messages for Y seconds
			go m.handleReadMessage(done)
			time.Sleep(time.Millisecond * MessageProcessDurationMs)
			close(done)
			time.Sleep(time.Millisecond * MessageProcessIntervalMs)
		}
	}()

}

// keep reading the socket, if its too full just wait
func (m *messageRelayerImpl) read() {
	for {
		msg, err := m.socket.Read()
		if err != nil {
			// Handle the error appropriately. For now, let's just continue.
			return
		}
		m.readMessageQueue <- msg

		// for testing, throttle the read
		fmt.Println("Read() message successfully: ", msg)
		time.Sleep(time.Millisecond * ReadMessagesDelayMs)
	}
}

func (m *messageRelayerImpl) handleReadMessage(done chan bool) {
	var buf []message.Message
	for {
		select {
		case msg := <-m.readMessageQueue:
			buf = append(buf, msg)

		case <-done:
			m.trimBuf(buf)
			buf = nil

			m.broadcastMessages()
			return
		}
	}
}

// Trim buffer and persist that to messages map to be broadcast/processed
func (m *messageRelayerImpl) trimBuf(buf []message.Message) {
	for _, msg := range buf {
		switch msg.Type {
		case message.StartNewRound:
			m.messages[msg.Type] = append(m.messages[msg.Type], msg)
			if len(m.messages[msg.Type]) > 2 {
				// We must always ensure that we broadcast the 2 most recent StartNewRound messages
				m.messages[msg.Type] = m.messages[msg.Type][1:]
			}
		case message.ReceivedAnswer:
			// Requirement:
			//   We only need to ensure that we broadcast only the most recent ReceivedAnswer message
			m.messages[msg.Type] = []message.Message{msg}
		}
	}
}

// Broadcast messages from the messages map
func (m *messageRelayerImpl) broadcastMessages() {
	fmt.Println("Broadcasting Messages..")
	// Requirement:
	//      Any time that both a “StartNewRound” and a “ReceivedAnswer” are queued, the “StartNewRound” message should be broadcasted first
	for _, msg := range m.messages[message.StartNewRound] {
		m.broadcastMessage(msg)
	}
	for _, msg := range m.messages[message.ReceivedAnswer] {
		m.broadcastMessage(msg)
	}

	// clear messages
	m.messages = make(map[message.MessageType][]message.Message)
}

func (m *messageRelayerImpl) broadcastMessage(msg message.Message) {
	subscribers, exists := m.subscribers[msg.Type]
	if !exists {
		fmt.Println("No subscribers")
		return
	}

	for _, ch := range subscribers {
		select {
		case ch <- msg:
			fmt.Println("Message sent successfully")
		default:
			// Requirement:
			//     Any time that one of the subscribers of this system is busy and cannot receive a message immediately, we should just skip broadcasting to that subscriber
			fmt.Println("Subscriber is busy, skip")
		}
	}
}
