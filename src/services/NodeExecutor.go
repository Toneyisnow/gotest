package services

import (
	"common"
	"components/types"
	"fmt"
	"log"
	"networking"
	"networking/pb"
	"time"
)

// Main function to start
func ExecuteNode() {


	rawMessagesQueue := make(chan *pb.NetMessage, 100)

	go StartServer(rawMessagesQueue)

	go HandleMessage(rawMessagesQueue);

	time.Sleep(2 * time.Second)

	TestingClient()

	time.Sleep(8 * time.Second)

	fmt.Println("Exit.")
}



func StartServer(messageQueue chan *pb.NetMessage) {

	server := new(NodeServer)

	server.Initialize(messageQueue)

	fmt.Println("Server Starting...")
	server.Start()
	fmt.Println("Server Started.")

	time.Sleep(10 * time.Second)

	fmt.Println("Server Stopping...")
	server.Stop()
	fmt.Println("Server Stopped.")
}

func HandleMessage(messageQueue chan *pb.NetMessage) {

	for {
		rawMessage := <- messageQueue

		log.Printf("HandleMessage: got message from rawMessageQueue")

		if (rawMessage.MessageType == pb.NetMessageType_SendEvents) {

			event1 := rawMessage.SendEventMessage.Events[0]
			fmt.Printf("HandleMessage - EventId: %s", event1.EventId)
		}
	}

}

func TestingClient() {

	clientManager := new(NodeClientManager)
	clientManager.Initialize(100)

	config := common.LoadConfigFromFile()

	// Create Fake Transaction Event
	node := types.CreateNode(config.Self)
	account1 := types.CreateAccount("toneysui", "12345678")
	account2 := types.CreateAccount("yisui", "87654321")

	event1 := types.NewEvent(node)

	tran11 := types.CreateTransaction(account1, account2, 999)
	tran12 := types.CreateTransaction(account1, account2, 111)
	tran13 := types.CreateTransaction(account1, account2, 222)
	event1.AddTransaction(tran11)
	event1.AddTransaction(tran12)
	event1.AddTransaction(tran13)

	event2 := types.NewEvent(node)

	tran21 := types.CreateTransaction(account2, account1, 999)
	tran22 := types.CreateTransaction(account2, account1, 111)
	tran23 := types.CreateTransaction(account2, account1, 222)
	event2.AddTransaction(tran21)
	event2.AddTransaction(tran22)
	event2.AddTransaction(tran23)

	events := []*types.SWEvent {event1, event2}


	netMessage := networking.ComposeNetMessage(events)

	clientManager.SendToNode("2", netMessage)

	time.Sleep(2 * time.Second)
	clientManager.SendToNode("2", netMessage)


	/*
	config := LoadConfigFromFile()
	client.Connect(&config.NetworkPeers[0])

	client.SendMessage("Good to see that")

	time.Sleep(time.Second)
	client.SendMessage("Good to see that")

	time.Sleep(time.Second)
	client.SendMessage("Good to see that")

	time.Sleep(time.Second)
	client.SendMessage("Good to see that")
	*/
}
