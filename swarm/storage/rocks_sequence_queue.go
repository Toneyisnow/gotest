package storage

import (
	"errors"
	"sync"
	"github.com/golang/protobuf/proto"
)

const (
	RocksQueueMaxIndex = ^uint32(0)
)

type RocksSequenceQueue struct {

	RocksContainer

	queueId    string
	beginIndex uint32
	endIndex   uint32

	iterateIndex uint32
	//// itemCount  uint32

	queueMutex sync.Mutex
}

func NewRocksSequenceQueue(st *RocksStorage, queueId string) *RocksSequenceQueue {

	queue := new(RocksSequenceQueue)

	queue.storage = st
	queue.containerId = queueId
	queue.containerType = RocksContainerType_SequenceQueue

	queue.beginIndex = queue.GetMetadataValueUint32("b")
	queue.endIndex = queue.GetMetadataValueUint32("e")

	queue.iterateIndex = queue.beginIndex

	return queue
}

func (this *RocksSequenceQueue) Push(value []byte) (err error) {

	this.queueMutex.Lock()
	defer this.queueMutex.Unlock()

	curIndex := this.endIndex
	if this.endIndex < RocksQueueMaxIndex {

		if this.beginIndex == 0 || this.endIndex != this.beginIndex- 1 {
			this.endIndex ++
		} else {
			// The queue is full since the End has caught up with Begin
			return errors.New("Push to queue failed: the queue is full.")
		}

	} else {
		this.endIndex = 0
	}
	this.SetMetadataValueUint32("e", this.endIndex)

	bs := ConvertUintToBytes(curIndex)
	this.storage.PutSeek(this.GenerateKey(bs), value)

	return nil
}

func (this *RocksSequenceQueue) PushProto(pb proto.Message) (err error) {

	if data, err := proto.Marshal(pb); err != nil {
		return err
	} else {
		return this.Push(data)
	}
}

// Pop the value, and try to handle it with callback, if callback return false, then don't delete it from database
func (this *RocksSequenceQueue) Pop() (result []byte) {

	this.queueMutex.Lock()
	defer this.queueMutex.Unlock()

	if this.beginIndex == this.endIndex {
		// Queue is empty
		return
	}

	bs := ConvertUintToBytes(this.beginIndex)
	key, value, err := this.storage.SeekNext(this.GenerateKey(bs))

	if err != nil {
		return
	}

	if key == nil || value == nil {
		return nil
	}

	// Delete the current value and move beginIndex
	this.storage.DelSeek(key)

	if this.beginIndex < RocksQueueMaxIndex {
		this.beginIndex ++
	} else {
		this.beginIndex = 0
	}
	this.SetMetadataValueUint32("b", this.beginIndex)

	return value
}

// Iterate the values from beginIndex to endIndex
func (this *RocksSequenceQueue) StartIterate() {

	this.iterateIndex = this.beginIndex
}

func (this *RocksSequenceQueue) IterateNext() (result []byte) {

	this.queueMutex.Lock()
	defer this.queueMutex.Unlock()

	if this.iterateIndex == this.endIndex {
		// No more value to get
		return
	}

	bs := ConvertUintToBytes(this.iterateIndex)
	key, value, err := this.storage.SeekNext(this.GenerateKey(bs))

	if err != nil {
		return
	}

	if key == nil || value == nil {
		return nil
	}

	if this.iterateIndex < RocksQueueMaxIndex {
		this.iterateIndex ++
	} else {
		this.iterateIndex = 0
	}

	return value
}


func (this *RocksSequenceQueue) DataSize() uint32 {

	if this.beginIndex > this.endIndex {
		return this.beginIndex - this.endIndex
	} else {
		return RocksQueueMaxIndex - (this.endIndex - this.beginIndex)
	}
}