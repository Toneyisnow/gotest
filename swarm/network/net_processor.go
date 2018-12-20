package network

import (
	"../common"
	"github.com/smartswarm/go/log"
	"golang.org/x/crypto/openpgp/errors"
	"sync"
)

type NetManagerStatus int
const (
	NetManagerStatus_Idle = 0
	NetManagerStatus_Running = 1
)

type NetProcessor struct {

	_option *NetOption
	_eventHandler EventHandler
	_contextManager *NetContextManager
	_topology *NetTopology

	_mu_status sync.Mutex

	_status NetManagerStatus
}

func CreateProcessor(topo *NetTopology, handler EventHandler) *NetProcessor {

	if (topo == nil || handler == nil) {
		// throw exception here
		return nil
	}

	processor := new(NetProcessor)
	processor._option = DefaultOption()
	processor._topology = topo
	processor._eventHandler = handler

	processor.Initialize()
	processor._status = NetManagerStatus_Idle
	return processor
}

func (this *NetProcessor) GetOption() *NetOption {
	return this._option
}

func (this *NetProcessor) GetEventHandler() EventHandler {
	return this._eventHandler
}

func (this *NetProcessor) Initialize() {

	this._contextManager = CreateContextManager(this)
	this._status = NetManagerStatus_Idle
}

func (this *NetProcessor) Start() {

	this._mu_status.Lock()
	defer this._mu_status.Unlock()

	if (this._status == NetManagerStatus_Running) {
		return
	}

	this._status = NetManagerStatus_Running

	go StartWebSocketListen(this._topology.Self().Port, this.HandleIncomingConnection)

	// Start to connect to peers

	selfDevice := this._topology._self
	for _, peer := range this._topology._peers {
		if (peer.IPAddress != selfDevice.IPAddress && peer.Port != selfDevice.Port) {

		}
	}

}

func (this *NetProcessor) Stop() {

	this._mu_status.Lock()
	defer this._mu_status.Unlock()

	if (this._status == NetManagerStatus_Idle) {
		return
	}

	this._status = NetManagerStatus_Idle
	this._contextManager.ClearAll()
	this._status = NetManagerStatus_Idle
}

func (this *NetProcessor) HandleIncomingConnection(socket *NetWebSocket) {

	log.I("[network] handle incoming connection start. Socket=", socket)

	//// Why closing it?
	//// defer socket.Close()

	remoteHost := socket.RemoteHostAddress().String()

	remoteHostAddress := common.ComposeHostAddress(remoteHost)

	if remoteHostAddress == nil || !remoteHostAddress.IsValid() {
		log.E("[network] invalid remote address.")
		return
	}

	// TODO: 连接数限制判断

	// TODO: 小黑屋，1小时内有invalid的节点事件

	log.I("[network] receive new connection:", remoteHost)

	device := this._topology.GetPeerDeviceByIP(remoteHostAddress.IpAddress)

	if (device == nil) {
		log.I("client address is not in white list, ignore it.")
		return
	}

	this._contextManager.CreateIncomingContext(socket, device)

	/*
	// 订阅消息
	if !sync_mode.IsLight() {
		go m.subscribe(ctx)
	}
	m.Owner().EventMgr().Emit("connected", ctx)
	*/

	log.I("[network] handle incoming connection end.")
}

// Deprecated?
func (this *NetProcessor) HandleMessage(context *NetContext, rawMessage *NetMessage) {

	log.I("[network] receive message. context=[%d]:", context._index)


}

func (this *NetProcessor) SendEventToDeviceAsync(device *NetDevice, eventData []byte) (resultChan chan *NetEventResult) {

	log.I2("[network] send event. device=[%s]:", device.GetHostUrl())

	resultChan = make(chan *NetEventResult)

	go func() {

		event := ComposeEvent(eventData)

		context, err := this._contextManager.CreateOrGetOutgoingContext(device)
		if (err != nil || context == nil) {
			resultChan <- ComposeEventResult(event, errors.InvalidArgumentError("CreateOrGetOutgoingContext failed."))
			return
		}

		message := ComposeEventMessage("0", eventData)

		context.SendMessage(message)

		// Success
		resultChan <- ComposeEventResult(event, nil)
	}()

	return resultChan
}

func (this *NetProcessor) SendEventToContextAsync(context *NetContext, eventData []byte, result chan *NetEventResult) {

	log.I("[network] send message. context=[%d]:", context._index)

	go func() {

		event := ComposeEvent(eventData)
		message := ComposeEventMessage("0", eventData)

		context.SendMessage(message)

		// Success
		result <- ComposeEventResult(event, nil)
	}()
}


