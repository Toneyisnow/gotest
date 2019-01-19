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

	netProcessor *NetProcessor
	netOption    *NetOption

	incomingContexts map[string]*NetContext
	outgoingContexts map[string]*NetContext

	addremoveMutex sync.RWMutex

	seed int32 // 用于连接标号的种子
}

func NewContextManager(processor *NetProcessor) *NetContextManager {

	contextManager := new(NetContextManager)
	contextManager.netProcessor = processor
	contextManager.netOption = processor.GetOption()
	contextManager.Initialize()

	return contextManager
}

func (this *NetContextManager) Initialize() {

	this.addremoveMutex = sync.RWMutex{}
	this.incomingContexts = make(map[string]*NetContext)
	this.outgoingContexts = make(map[string]*NetContext)
}

func (this *NetContextManager) CreateIncomingContext(socket *NetWebSocket, device *NetDevice) (context *NetContext, err error) {

	context = NewIncomingContext(socket, device)
	if context == nil {
		err = errors.InvalidArgumentError("Create context failed.")
		return
	}

	this.Add(context)

	// 发送版本号
	// context.SendNotify("version", m.Owner().Ver())

	context.Open()

	// challenge
	if (this.netOption._needChallenge) {
		log.I("[network] require challenge...")
		var challenge string = base64.StdEncoding.EncodeToString(base.RandomBytes(30))
		context.SetMetadata("challenge", challenge)

		message := ComposeChallengeMessage("", challenge)
		context.SendMessage(message)
	}

	err = nil
	return
}

func (this *NetContextManager) CreateOrGetOutgoingContext(device *NetDevice) (context *NetContext, err error) {

	if device == nil {
		context = nil
		err = errors.InvalidArgumentError("device is null")
		return nil, nil
	}

	hostUrl := device.GetHostUrl()
	context, exist := this.outgoingContexts[hostUrl]
	if exist {
		return context, nil
	}

	log.I("[network] start creating connection to: ", device.GetHostUrl())

	clientServerHostUrl := this.netProcessor.topology.Self().GetHostUrl()
	socket, err := StartWebSocketDial(device, clientServerHostUrl)
	if err != nil || socket == nil {
		return nil, nil
	}

	context = NewOutgoingContext(socket, device)
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

func (this *NetContextManager) GetIncomingContext(device *NetDevice) (context *NetContext, err error) {

	if device == nil {
		context = nil
		err = errors.InvalidArgumentError("[network] device is null")
		return nil, nil
	}

	hostUrl := device.GetHostUrl()
	context, exist := this.incomingContexts[hostUrl]
	if exist {
		return context, nil
	}

	return nil, nil
}

// 添加连接
func (this *NetContextManager) Add(context *NetContext) {

	if (context == nil) {
		return
	}

	hostAddress := context.GetAddress()
	this.addremoveMutex.Lock()
	defer this.addremoveMutex.Unlock()

	contextList := this.incomingContexts
	if context.direction == NetConextDirection_Outgoing {
		contextList = this.outgoingContexts
	}

	_, exist := contextList[hostAddress]
	if exist {
		log.W("[network]", "duplicated connection, just ignore.", hostAddress)
		return
	}

	contextList[hostAddress] = context
	context.contextManager = this
	context.index = atomic.AddInt32(&this.seed, 1)
}

// 移除连接
func (this *NetContextManager) Remove(context *NetContext) {

	if (context == nil) {
		return
	}

	hostAddress := context.GetAddress()
	this.addremoveMutex.Lock()
	defer this.addremoveMutex.Unlock()

	if context.direction == NetConextDirection_Incoming {
		delete(this.incomingContexts, hostAddress)
	}
	if context.direction == NetConextDirection_Outgoing {
		delete(this.outgoingContexts, hostAddress)
	}

	context.contextManager = nil
	log.I("[network] connection removed. context index=", context.index)
}

// 通过url获取连出连接
func (this *NetContextManager) FindOutgoingByUrl(hostUrl string) (context *NetContext, exist bool) {

	this.addremoveMutex.Lock()
	defer this.addremoveMutex.Unlock()

	for _, v := range this.outgoingContexts {

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