package message

type MessageType int

const (
	StartNewRound MessageType = 1 << iota
	ReceivedAnswer
)

type Message struct {
	Type MessageType
	Data []byte
}
