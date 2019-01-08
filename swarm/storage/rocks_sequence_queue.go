package storage

import (
	"errors"
	"sync"
)

const (
	RocksQueueMaxIndex = ^uint32(0)
)

type RocksSequenceQueue struct {

	RocksContainer

	queueId    string
	beginIndex uint32
	endIndex   uint32
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

// Pop the value, and try to handle it with callback, if callback return false, then don't delete it from database
func (this *RocksSequenceQueue) Pop(callback dataCallbackFunc) {

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

	// Call function
	if !callback(value) {
		return
	}

	// Delete the current value and move beginIndex
	this.storage.DelSeek(key)

	if this.beginIndex < RocksQueueMaxIndex {
		this.beginIndex ++
	} else {
		this.beginIndex = 0
	}
	this.SetMetadataValueUint32("b", this.beginIndex)

	return
}

func (this *RocksSequenceQueue) DataSize() uint32 {

	if this.beginIndex > this.endIndex {
		return this.beginIndex - this.endIndex
	} else {
		return RocksQueueMaxIndex - (this.endIndex - this.beginIndex)
	}
}