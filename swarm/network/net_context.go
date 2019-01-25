package network

import (
	"../common/log"
)

type NetContextDirection int
type NetContextStatus int

const (
	NetConextDirection_Incoming = 0
	NetConextDirection_Outgoing = 1
)
const (

	NetConextStatus_Unintialized = 0
	NetConextStatus_Closed = 1
	NetConextStatus_PendingChallenge = 2
	NetConextStatus_Ready = 3
)

type NetContext struct {

	contextManager  *NetContextManager
	index           int32 // 连接编号，用于简单排序
	device          *NetDevice
	direction       NetContextDirection
	status			NetContextStatus

	webSocket       *NetWebSocket

	metadata        map[string]string
}

func NewIncomingContext(socket *NetWebSocket, device *NetDevice) *NetContext {

	log.I("[net] creating new incoming context for device=", device.Id)
	context := new(NetContext)

	context.webSocket = socket
	context.device = device

	context.metadata = make(map[string]string)

	context.direction = NetConextDirection_Incoming
	context.status = NetConextStatus_Unintialized

	return context
}

func NewOutgoingContext(socket *NetWebSocket, device *NetDevice) *NetContext {

	context := new(NetContext)

	context.webSocket = socket
	context.device = device

	context.metadata = make(map[string]string)

	context.direction = NetConextDirection_Outgoing
	context.status = NetConextStatus_Unintialized

	return context
}

func (this *NetContext) GetAddress() string {

	// For testing purpose, using IP + Port, so that can be tested on one machine
	return this.device.GetHostUrl()

	// return this.device.IPAddress
}

func (this *NetContext) Open() {

	log.I("[network] start netContext.Open()")

	// If incoming, message loop; if outgoing, start heartbeating
	if (this.direction == NetConextDirection_Incoming) {

		go this.MessageLoop()
	} else if  (this.direction == NetConextDirection_Outgoing) {

		go this.MessageLoop()
	}
}

func (this *NetContext) Close() {

	log.I("[network] closing net context...")

	this.contextManager.Remove(this)
	err := this.webSocket.Close()
	if err != nil {
		log.W("[network] error while closing web socket:", err.Error())
	}
}

func (this *NetContext) SendMessage(message *NetMessage) (err error) {

	_, err = this.webSocket.Write(message)
	return
}

/// This is the main loop function for context, listen to message and deliver to upper
func (this *NetContext) MessageLoop() {

	for {
		message, err := this.webSocket.ReadMessage()

		if err != nil {
			log.W("[network] error while read message:", err, this.device.GetHostUrl())
			break
		}

		if message != nil {
			log.I("[network] context got message. type=", message.MessageType, " id=", message.MessageId)
			HandleMessage(this, message)
		}

		if this.status == NetConextStatus_Closed {
			break
		}
	}

	defer this.Close()

	log.I("[network] closing connection: ", this.index)

	/*
	//// defer this.Close()

	//conn.SetReadDeadline(time.Now().Add(300 * time.Second))
	//conn.SetWriteDeadline(time.Now().Add(300 * time.Second))
	for {

		//// if this.webSocket.

		message, err := this.webSocket.ReadMessage()
		if err != nil {
			log.W("[network] error while read message:", err, this.device.GetHostUrl())
			break
		}

		HandleMessage(this, message)
	}
	log.I("[network] finish connection:", this.index)
	*/
}

func (this *NetContext) HeartBeat() {

}

func (this *NetContext) SetMetadata(key string, value string) {

	log.I("[net] set metadata. context id=", this.index, "key=", key, "value=", value)
	this.metadata[key] = value
}

func (this *NetContext) GetMetadata(key string) string {

	val, _ := this.metadata[key]
	log.I("[net] get metadata. context id=", this.index, "key=", key, "value=", val)
	return val
}