package network

import (
	"encoding/base64"
	"github.com/smartswarm/core/base"
	"github.com/smartswarm/go/log"
	"golang.org/x/crypto/openpgp/errors"
	"sync"
	"sync/atomic"
)

type NetContextManager struct {

	_processor *NetProcessor
	_option *NetOption

	_incomingContexts map[string]*NetContext
	_outgoingContexts map[string]*NetContext

	_addremove_mutex sync.RWMutex

	_seed int32 // 用于连接标号的种子
}

func CreateContextManager(processor *NetProcessor) *NetContextManager {

	contextManager := new(NetContextManager)
	contextManager._processor = processor
	contextManager._option = processor.GetOption()
	contextManager.Initialize()

	return contextManager
}

func (this *NetContextManager) Initialize() {

	this._incomingContexts = make(map[string]*NetContext)
	this._outgoingContexts = make(map[string]*NetContext)
}

func (this *NetContextManager) CreateIncomingContext(socket *NetWebSocket, device *NetDevice) (context *NetContext, err error) {

	context = CreateIncomingContext(socket, device)
	if context == nil {
		err = errors.InvalidArgumentError("Create context failed.")
		return
	}

	this.Add(context)

	// 发送版本号
	// context.SendNotify("version", m.Owner().Ver())

	context.Open()

	// challenge
	if (this._option._needChallenge) {
		log.I("[network] require challenge...")
		var challenge string = base64.StdEncoding.EncodeToString(base.RandomBytes(30))
		context.SetMetadata("challenge", challenge)

		message := ComposeChallengeMessage("", "ran_chars_here")
		context.SendMessage(message)
	}

	err = nil
	return
}

func (this *NetContextManager) CreateOrGetOutgoingContext(device *NetDevice) (context *NetContext, err error) {

	if device == nil {
		context = nil
		err = errors.InvalidArgumentError("device is null")
		return
	}

	hostUrl := device.GetHostUrl()
	context, exist := this._outgoingContexts[hostUrl]
	if exist {
		return
	}

	log.I("[network] start creating connection to: ", device.GetHostUrl())

	socket, err := StartWebSocketDial(device)

	context = CreateOutgoingContext(socket, device)
	this.Add(context)


	// 开始心跳
	context.Open()

	return context, nil

	/*
	// 订阅消息
	if !sync_mode.IsLight() {
		go m.subscribe(ctx)
	}
	*/

	/*
	// 如果定义了外部连接，扩散之
	if pub_url, ok := m.Owner().Config().GetString("pub_url"); ok && pub_url != "" {
		ctx.SendNotify("my_url", pub_url)
	}

	m.Owner().EventMgr().Emit("connected", ctx)

	go m._message_loop(ctx)

	if onOpen != nil {
		onOpen(nil, ctx)
	}
	*/

}

// 添加连接
func (this *NetContextManager) Add(context *NetContext) {

	if (context == nil) {
		return
	}

	hostAddress := context.GetAddress()
	this._addremove_mutex.Lock()
	defer this._addremove_mutex.Unlock()

	contextList := this._incomingContexts
	if context._direction == NetConextDirection_Outgoing {
		contextList = this._outgoingContexts
	}

	_, exist := contextList[hostAddress]
	if exist {
		log.W("[network]", "duplicated connection, just ignore.", hostAddress)
		return
	}

	contextList[hostAddress] = context
	context._manager = this
	context._index = atomic.AddInt32(&this._seed, 1)
}

// 移除连接
func (this *NetContextManager) Remove(context *NetContext) {

	if (context == nil) {
		return
	}

	hostAddress := context.GetAddress()
	this._addremove_mutex.Lock()
	defer this._addremove_mutex.Unlock()

	if context._direction == NetConextDirection_Incoming {
		delete(this._incomingContexts, hostAddress)
	}
	if context._direction == NetConextDirection_Outgoing {
		delete(this._outgoingContexts, hostAddress)
	}

	context._manager = nil
	// Debug("[netowrk] connection removed.", mgr)
}


// 通过url获取连出连接
func (this *NetContextManager) FindOutgoingByUrl(hostUrl string) (context *NetContext, exist bool) {

	this._addremove_mutex.Lock()
	defer this._addremove_mutex.Unlock()

	for _, v := range this._outgoingContexts {

		if v.GetAddress() == hostUrl {
			context = v
			exist = true
			return
		}
	}

	context = nil
	exist = false
	return
}

// 断开所有网络连接
func (this *NetContextManager) ClearAll() {

}