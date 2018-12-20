package network

import (
	"encoding/base64"
	"github.com/smartswarm/core/base"
	"github.com/smartswarm/core/network/connection"
	"github.com/smartswarm/core/sync_mode"
	"github.com/smartswarm/go/log"
	"github.com/smartswarm/go/timer"
	"sync"
)

type NetManagerStatus int
const (
	NetManagerStatus_Idle = 0
	NetManagerStatus_Running = 1
)

type NetProcessor struct {

	_option *NetOption
	_eventHandler *EventHandler
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
	processor._eventHandler = &handler
	processor._status = NetManagerStatus_Idle
	return processor
}

func (this *NetProcessor) GetOption() *NetOption {
	return this._option
}

func (this *NetProcessor) GetEventHandler() *EventHandler {
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

	go StartWebSocketListen(this._topology.Self().Port, this.HandleIncoming)

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

func (this *NetProcessor) CreateOrGetOutgoingContext(device *NetDevice) {


}


func (m *Manager) connectToPeer(url string, onOpen OpenCallback) (ctx *WSContext, err error) {
	log.I("[network] will connect to:", url)

	var conn connection.Conn
	if conn, err = connection.Dial(url); err != nil {
		log.W("[network] connect failed!", err)
		// retry connect
		timer.SetTimeout(func() {
			m.connectToPeer(url, onOpen)
		}, 5000)
		return
	} else {
		log.I("[network] connected to:", conn.RemoteAddr())
	}

	ctx = new(WSContext)
	ctx.NewOutgoing(conn, url)
	m.ctx_mgr.add(ctx)

	// 订阅消息
	if !sync_mode.IsLight() {
		go m.subscribe(ctx)
	}

	// 如果定义了外部连接，扩散之
	if pub_url, ok := m.Owner().Config().GetString("pub_url"); ok && pub_url != "" {
		ctx.SendNotify("my_url", pub_url)
	}

	m.Owner().EventMgr().Emit("connected", ctx)

	go m._message_loop(ctx)

	if onOpen != nil {
		onOpen(nil, ctx)
	}

	return ctx, nil
}

func (this *NetProcessor) HandleIncoming(socket *NetWebSocket) {

	defer socket.Close()

	ip := socket.RemoteHostAddress().String()
	if len(ip) <= 0 {
		log.E("[network] invalid remote ip.")
		return
	}

	// TODO: 连接数限制判断

	// TODO: 小黑屋，1小时内有invalid的节点事件

	log.I("[network] receive new connection:", ip)

	device := this._topology.GetPeerDeviceByAddress(ip)

	if (device == nil) {
		log.I("client address is not in white list, ignore it.")
		return
	}

	context := CreateIncomingContext(socket, device)
	this._contextManager.Add(context)

	// 发送版本号
	// context.SendNotify("version", m.Owner().Ver())

	// challenge
	if (this._option._needChallenge) {
		log.I("[network] require challenge...")
		var challenge string = base64.StdEncoding.EncodeToString(base.RandomBytes(30))
		context.SetMetadata("challenge", challenge)

		message := ComposeChallengeMessage("", "ran_chars_here")
		context.SendMessage(message)
	}

	context.Open()

	/*
	// 订阅消息
	if !sync_mode.IsLight() {
		go m.subscribe(ctx)
	}
	m.Owner().EventMgr().Emit("connected", ctx)
	*/

}

func (this *NetProcessor) HandleMessage(context *NetContext, rawMessage *NetMessage) {

	log.I("[network] receive message. context=[]:", context._index)

}