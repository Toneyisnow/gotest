package services

import (
	"flag"
	"github.com/golang/protobuf/proto"
	"github.com/gorilla/websocket"
	"log"
	"net/url"
	"objectmodels/network"
	"strconv"
)

type NodeClient struct {

	config *NodeConfig
	connection *websocket.Conn
	isConnected bool
}

func (this *NodeClient) Initialize() {

	this.config = LoadConfigFromFile()
	this.isConnected = false
}

func (this *NodeClient) Connect(toServer *NodeInfo) {

	var peeraddress = "localhost:" + strconv.Itoa(toServer.ServerPort)
	var peeraddr = flag.String("peeraddr", peeraddress, "http service peeraddress")

	u := url.URL{Scheme: "ws", Host: *peeraddr, Path: "/events"}
	log.Printf("connecting to %s", u.String())

	this.connection, _, _ = websocket.DefaultDialer.Dial(u.String(), nil)
	this.isConnected = true

}

func (this *NodeClient) SendMessage(message string) {

	if (!this.isConnected) {
		return
	}

	mess := new (network.BaseMessage)
	mess.OwnerId = "123456"
	mess.Hash = "xxxxxx"
	mess.Type = network.BaseMessage_SendEvents
	mess.SendEventMessage = new(network.SendEventMessage)
	mess.SendEventMessage.EventId = "4321"

	messageBuffer, _ := proto.Marshal(mess)
	err := this.connection.WriteMessage(websocket.TextMessage, messageBuffer)
	if err != nil {
		log.Println("write:", err)
		return
	}
}

func (this *NodeClient) Close() {

	this.connection.Close()
	this.isConnected = false

}