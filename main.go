package main

import (
	"fmt"
	"time"

	"example.com/messagerelay/message"
	"example.com/messagerelay/network"
	"example.com/messagerelay/subscriber"

	"example.com/messagerelay/relayer"
)

// What running main does:
// 1/ Starts the relayer and its socket and starts reading from it. for the first 2 seconds, there are no subscribers
// 2/ Subscribers 1 & 2 are added
// 3/ Subscribers receive messages, for 5000ms
// 4/ Subscribers are stopped
// 5/ Runs for 5 more seconds with stopped subscribers
// 6/ End
func main() {
	fmt.Println("Starting..")
	s := network.NewSocket()
	relay := relayer.NewMessageRelayer(s)
	relay.Start()
	time.Sleep(time.Millisecond * 2000)

	sub := subscriber.NewMessageSubscriber()
	sub_done := make(chan (bool))
	sub.Subscribe(relay, message.StartNewRound, sub_done)

	sub2 := subscriber.NewMessageSubscriber()
	sub_done2 := make(chan (bool))
	sub2.Subscribe(relay, message.ReceivedAnswer, sub_done2)

	fmt.Println("Running Subscribers for both message types for 5000 ms")
	time.Sleep(time.Millisecond * 5000)
	fmt.Println("Stopping Subscribers")
	close(sub_done)
	close(sub_done2)
	time.Sleep(time.Millisecond * 5000)

	fmt.Println("End")
}
