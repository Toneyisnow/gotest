package network

import (
	"github.com/smartswarm/go/log"
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

	_manager *NetContextManager
	_index int32		// 连接编号，用于简单排序
	_device *NetDevice
	_direction NetConextDirection

	_webSocket *NetWebSocket

	_metadata map[string]string
}

func CreateIncomingContext(socket *NetWebSocket, device *NetDevice) *NetContext {

	context := new(NetContext)

	context._webSocket = socket
	context._device = device
	context._direction = NetConextDirection_Incoming
	context._metadata = make(map[string]string)

	return context
}

func CreateOutgoingContext(socket *NetWebSocket, device *NetDevice) *NetContext {

	context := new(NetContext)

	context._webSocket = socket
	context._device = device
	context._direction = NetConextDirection_Outgoing
	context._metadata = make(map[string]string)

	return context
}

func (this *NetContext) GetAddress() string {

	// For testing purpose, using IP + Port, so that can be tested on one machine
	return this._device.HostUrl

	// return this._device.IPAddress
}

func (this *NetContext) Open() {

	// If incoming, message loop; if outgoing, start heartbeating
	if (this._direction == NetConextDirection_Incoming) {

		go this.MessageLoop()
	} else if  (this._direction == NetConextDirection_Outgoing) {

		go this.HeartBeat()
	}

}

func (this *NetContext) Close() {

}

func (this *NetContext) SendMessage(message *NetMessage) {


}

/// This is the main loop function for context, listen to message and deliver to upper
func (this *NetContext) MessageLoop() {

	defer this.Close()

	//conn.SetReadDeadline(time.Now().Add(300 * time.Second))
	//conn.SetWriteDeadline(time.Now().Add(300 * time.Second))
	for {
		message, err := this._webSocket.ReadMessage()
		if err != nil {
			log.W("[network] got err from conn:", err, this._device.HostUrl)
			break
		}

		processor := this._manager._processor
		processor.HandleMessage(this, message)
	}
	log.I("[network] finish connection:", this._index)
}

func (this *NetContext) HeartBeat() {

}

func (this *NetContext) SetMetadata(key string, value string) {
	this._metadata[key] = value
	}

func (this *NetContext) GetMetadata(key string) string {
	val, err := this._metadata[key]
	if (err) {
		return ""
	}

	return val
}