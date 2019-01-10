package storage

import (

	"github.com/golang/protobuf/proto"
)

type RocksChannel struct {

	capacity uint32
	sequenceQueue *RocksSequenceQueue
	channel chan bool
}

func NewRocksChannel(storage *RocksStorage, channelId string, capacity uint32) *RocksChannel {

	channel := new(RocksChannel)

	channel.capacity = capacity
	channel.sequenceQueue = NewRocksSequenceQueue(storage, channelId, capacity)
	channel.channel = make(chan bool, capacity)


	return channel
}

// Load all of the storage queue data into channel
func (this *RocksChannel) Reload() {

	this.channel = make(chan bool, this.capacity)

	for {
		_, value := this.sequenceQueue.IterateNext()
		if value == nil {
			break
		}

		this.channel <- true
	}
}

func (this *RocksChannel) PushProto(pb proto.Message) {

	this.sequenceQueue.PushProto(pb)
	this.channel <- true
}


func (this *RocksChannel) Push(value []byte) {

	this.sequenceQueue.Push(value)
	this.channel <- true
}

// Handling for the output of the channel
func (this *RocksChannel) Listen(listener func (data []byte) ) {

	for {
		<-this.channel

		this.sequenceQueue.StartIterate()
		_, value := this.sequenceQueue.IterateNext()

		if value == nil {
			continue
		}

		listener(value)

		this.sequenceQueue.Pop()
	}
}