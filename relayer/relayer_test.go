package relayer

import (
	"testing"
	"time"

	"example.com/messagerelay/message"
	"example.com/messagerelay/mocks"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestSubscribeToMessages(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSocket := mocks.NewMockNetworkSocket(ctrl)
	relayer := NewMessageRelayer(mockSocket)

	testChan := make(chan message.Message)
	msgType := message.StartNewRound

	// Subscribe to a message type
	relayer.SubscribeToMessages(msgType, testChan)

	// Use assert from testify to check if the channel is correctly added to subscribers
	subs, exists := relayer.subscribers[msgType]
	assert.True(t, exists, "Subscription should exist")
	assert.Equal(t, 1, len(subs), "There should be one subscriber")
	assert.Equal(t, testChan, subs[0], "Subscriber channel should match")
}

func TestStart(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockSocket := mocks.NewMockNetworkSocket(ctrl)

	// Set up the expected behavior for the mock
	mockMsg := message.Message{Type: message.StartNewRound, Data: []byte{123}}
	mockSocket.EXPECT().Read().Return(mockMsg, nil).AnyTimes()

	relayer := NewMessageRelayer(mockSocket)

	// Create a subscriber channel
	subscriberChan := make(chan message.Message, 5)
	relayer.SubscribeToMessages(message.StartNewRound, subscriberChan)

	relayer.Start()

	// Wait for a message to be broadcast to the subscriber
	select {
	case receivedMsg := <-subscriberChan:
		// Assert that the received message matches the expected message using testify
		assert.Equal(t, mockMsg.Type, receivedMsg.Type, "Received message type does not match expected type")
		assert.Equal(t, mockMsg.Data, receivedMsg.Data, "Received message content does not match expected content")
	case <-time.After(time.Second * 2):
		// Fail the test if no message is received within the timeout
		assert.Fail(t, "No message was received from the relayer within the expected time")
	}
}
