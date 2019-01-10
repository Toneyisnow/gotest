package storage

import (

	"github.com/golang/protobuf/proto"
	"github.com/smartswarm/go/log"
)

type RocksChannel struct {

	sequenceQueue *RocksSequenceQueue
	channel chan bool
}

func NewRocksChannel(storage *RocksStorage, channelId string, capacity uint32) *RocksChannel {

	channel := new(RocksChannel)

	channel.sequenceQueue = NewRocksSequenceQueue(storage, channelId, capacity)
	channel.channel = make(chan bool)


	return channel
}

// Load all of the storage queue data into channel
func (this *RocksChannel) Reload() {

	this.channel = make(chan bool)

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
func (this *RocksChannel) HandleProto(handler func (index uint32, pb proto.Message) bool) {

	for {
		signal := <-this.channel

		this.sequenceQueue.StartIterate()
		index, value := this.sequenceQueue.IterateNext()

		obj proto.Message
		err := proto.Unmarshal(value, obj)

		this.dagStorage.queueIncomingVertex.Pop()

	}

}