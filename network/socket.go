package network

import (
	"math/rand"
	"time"

	"example.com/messagerelay/message"
)

type NetworkSocket interface {
	Read() (message.Message, error)
}

type socketImpl struct {
}

func NewSocket() *socketImpl {
	return &socketImpl{}
}

func (*socketImpl) Read() (message.Message, error) {
	rand.Seed(time.Now().UnixNano())

	m := message.Message{Type: message.MessageType(rand.Intn(2) + 1), Data: randomByteSlice(5)}
	return m, nil
}

func randomByteSlice(length int) []byte {
	rand.Seed(time.Now().UnixNano()) // Initialize the random number generator
	slice := make([]byte, length)
	for i := range slice {
		slice[i] = byte(rand.Intn(256)) // Random byte between 0-255
	}
	return slice
}
