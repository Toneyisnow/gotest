package network

import (
	"github.com/smartswarm/go/log"
	"sync"
	"sync/atomic"
)

type NetContextManager struct {

	_processor *NetProcessor
	_option *NetOption

	_contexts map[string]*NetContext

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

	this._contexts = make(map[string]*NetContext)
}


// 添加连接
func (this *NetContextManager) Add(context *NetContext) {

	if (context == nil) {
		return
	}

	hostAddress := context.GetAddress()
	this._addremove_mutex.Lock()
	defer this._addremove_mutex.Unlock()

	_, exist := this._contexts[hostAddress]
	if exist {
		log.W("[network]", "duplicated connection, just ignore.", hostAddress)
		return
	}

	this._contexts[hostAddress] = context
	context._manager = this
	context._index = atomic.AddInt32(&this._seed, 1)

	// 开始心跳
	context.Open()

	// Debug("[netowrk] connection added.", mgr)
}


// 移除连接
func (this *NetContextManager) Remove(context *NetContext) {

	if (context == nil) {
		return
	}

	hostAddress := context.GetAddress()
	this._addremove_mutex.Lock()
	defer this._addremove_mutex.Unlock()

	delete(this._contexts, hostAddress)
	context._manager = nil
	// Debug("[netowrk] connection removed.", mgr)
}

// 断开所有网络连接
func (this *NetContextManager) ClearAll() {

}