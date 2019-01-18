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

	option         *NetOption
	eventHandler   EventHandler
	contextManager *NetContextManager
	topology       *NetTopology

	mu_status sync.Mutex

	status NetManagerStatus
}

func CreateProcessor(topo *NetTopology, handler EventHandler) *NetProcessor {

	if topo == nil || handler == nil {
		// throw exception here
		return nil
	}

	processor := new(NetProcessor)
	processor.option = DefaultOption()
	processor.topology = topo
	processor.eventHandler = handler

	processor.Initialize()
	processor.status = NetManagerStatus_Idle
	return processor
}

func (this *NetProcessor) GetOption() *NetOption {
	return this.option
}

func (this *NetProcessor) GetEventHandler() EventHandler {
	return this.eventHandler
}

func (this *NetProcessor) Initialize() {

	this.contextManager = NewContextManager(this)
	this.status = NetManagerStatus_Idle
}

func (this *NetProcessor) StartServer() {

	this.mu_status.Lock()
	defer this.mu_status.Unlock()

	if this.status == NetManagerStatus_Running {
		return
	}

	this.status = NetManagerStatus_Running

	go StartWebSocketListen(this.topology.Self().Port, this.HandleIncomingConnection)
}

func (this *NetProcessor) Stop() {

	this.mu_status.Lock()
	defer this.mu_status.Unlock()

	if (this.status == NetManagerStatus_Idle) {
		return
	}

	this.status = NetManagerStatus_Idle
	this.contextManager.ClearAll()
	this.status = NetManagerStatus_Idle
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

	device := this.topology.GetPeerDeviceByIP(remoteHostAddress.IpAddress)

	if (device == nil) {
		log.I("client address is not in white list, ignore it.")
		return
	}

	this.contextManager.CreateIncomingContext(socket, device)

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

	log.I("[network] receive message. context=", context.index)


}

func (this *NetProcessor) ConnectToDeviceAsync(device *NetDevice) (resultChan chan bool) {

	log.I("[network] ConnectToDevice. device=", device.GetHostUrl())

	resultChan = make(chan bool)

	go func() {
		context, err := this.contextManager.GetIncomingContext(device)
		if err != nil || context == nil {
			context, err = this.contextManager.CreateOrGetOutgoingContext(device)
			if err != nil || context == nil {
				resultChan <- false
				return
			}
		}


		resultChan <- true
	}()

	return resultChan
}

func (this *NetProcessor) SendEventToDeviceAsync(device *NetDevice, eventData []byte) (resultChan chan *NetEventResult) {

	log.I2("[network] send event. device=[%s]:", device.GetHostUrl())

	resultChan = make(chan *NetEventResult)

	go func() {

		event := ComposeEvent(eventData)

		context, err := this.contextManager.CreateOrGetOutgoingContext(device)
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

	log.I("[network] send message. context=[%d]:", context.index)

	go func() {

		event := ComposeEvent(eventData)
		message := ComposeEventMessage("0", eventData)

		context.SendMessage(message)

		// Success
		result <- ComposeEventResult(event, nil)
	}()
}


