package services

import (
	"common"
	"flag"
	"github.com/golang/protobuf/proto"
	"github.com/gorilla/websocket"
	"log"
	"net/url"
	"networking/pb"
	"objectmodels/network"
	"strconv"
)

type NodeClient struct {

	config *common.NodeConfig
	connection *websocket.Conn
	isConnected bool
	messageIndex int
	_messageQueue  chan *pb.NetMessage
}

func (this *NodeClient) Initialize(nodeId string, messageQueue chan *pb.NetMessage) {

	this.config = common.LoadConfigFromFile()
	this.isConnected = false
	this.messageIndex = 100

	this._messageQueue = messageQueue

	this.Connect(this.config.GetPeerById(nodeId))
}

func (this *NodeClient) Start() {

	log.Println("NodeClient started.")

	for {
		message := <- this._messageQueue

		log.Println("Got message, sending it...")
		this.SendMessage(message)
	}
}

func (this *NodeClient) Connect(toServer *common.NodeInfo) {

	if (toServer == nil) {
		log.Println("Connect failed: toServer is nil.")
		this.isConnected = false
		return
	}

	var peeraddress = "localhost:" + strconv.Itoa(toServer.ServerPort)
	var peeraddr = flag.String("peeraddr", peeraddress, "http service peeraddress")

	u := url.URL{Scheme: "ws", Host: *peeraddr, Path: "/events"}
	log.Printf("connecting to %s", u.String())

	this.connection, _, _ = websocket.DefaultDialer.Dial(u.String(), nil)
	this.isConnected = true

}


func (this *NodeClient) SendMessage(message *pb.NetMessage) {

	if (!this.isConnected) {
		return
	}

	log.Printf("SendMessage: pb.NetMessage=[%s, %s]", message.Hash, message.OwnerId)

	messageBuffer, _ := proto.Marshal(message)
	log.Printf("Message protobuf=[%s]", messageBuffer)
	err := this.connection.WriteMessage(websocket.TextMessage, messageBuffer)
	if err != nil {
		log.Println("write:", err)
		return
	}
}

func (this *NodeClient) SendMessage2(message string) {

	if (!this.isConnected) {
		return
	}

	mess := new (network.BaseMessage)
	mess.OwnerId = "123456"
	mess.Hash = strconv.Itoa(this.messageIndex)
	this.messageIndex ++

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