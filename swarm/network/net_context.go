package network

import (
	"../common/log"
)

type NetConextDirection int
type NetConextStatus int

const (
	NetConextDirection_Incoming = 0
	NetConextDirection_Outgoing = 1
)
const (
	NetConextStatus_Closed = 0
	NetConextStatus_Ready = 1
)

type NetContext struct {

	contextManager *NetContextManager
	index          int32 // 连接编号，用于简单排序
	device         *NetDevice
	direction      NetConextDirection

	webSocket      *NetWebSocket

	metadata       map[string]string
}

func NewIncomingContext(socket *NetWebSocket, device *NetDevice) *NetContext {

	log.I("[net] new incomingContext for device=", device.Id)
	context := new(NetContext)

	context.webSocket = socket
	context.device = device
	context.direction = NetConextDirection_Incoming
	context.metadata = make(map[string]string)

	return context
}

func NewOutgoingContext(socket *NetWebSocket, device *NetDevice) *NetContext {

	context := new(NetContext)

	context.webSocket = socket
	context.device = device
	context.direction = NetConextDirection_Outgoing
	context.metadata = make(map[string]string)

	return context
}

func (this *NetContext) GetAddress() string {

	// For testing purpose, using IP + Port, so that can be tested on one machine
	return this.device.GetHostUrl()

	// return this.device.IPAddress
}

func (this *NetContext) Open() {

	log.I("[network] Start NetContext.Open()")
	// If incoming, message loop; if outgoing, start heartbeating
	if (this.direction == NetConextDirection_Incoming) {

		go this.MessageLoop()
	} else if  (this.direction == NetConextDirection_Outgoing) {

		go this.HeartBeat()
	}

	log.I("[network] End NetContext.Open()")
}

func (this *NetContext) Close() {

	log.I("[network] Start NetContext.Close()")

	this.contextManager.Remove(this)
	this.webSocket.Close()


	log.I("[network] End NetContext.Close()")
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
		HandleMessage(this, message)

		/// time.Sleep(time.Second)
	}

	defer this.Close()

	log.I("[network] finish connection: ", this.index)

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
	this.metadata[key] = value
	}

func (this *NetContext) GetMetadata(key string) string {
	val, err := this.metadata[key]
	if (err) {
		return ""
	}

	return val
}