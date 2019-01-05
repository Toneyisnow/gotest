package storage

import (
	"errors"
	"sync"
)

const (
	RocksQueueMaxIndex = ^uint32(0) - 1
)

type RocksQueue struct {

	RocksContainer

	QueueId string
	BeginIndex uint32
	EndIndex uint32

	queueMutex sync.Mutex
}

func ComposeRocksQueue(st *RocksStorage, queueId string) *RocksQueue {

	queue := new(RocksQueue)

	queue.storage = st
	queue.containerId = queueId
	queue.containerType = RocksContainerType_SequenceQueue

	queue.BeginIndex = queue.GetMetadataValueUint32("b")
	queue.EndIndex = queue.GetMetadataValueUint32("e")

	return queue
}

func (this *RocksQueue) Push(value []byte) (err error) {

	this.queueMutex.Lock()
	defer this.queueMutex.Unlock()

	curIndex := this.EndIndex
	if this.EndIndex < RocksQueueMaxIndex {

		if this.BeginIndex == 0 || this.EndIndex != this.BeginIndex - 1 {
			this.EndIndex ++
		} else {
			// The queue is full since the End has caught up with Begin
			return errors.New("Push to queue failed: the queue is full.")
		}

	} else {
		this.EndIndex = 0
	}
	this.SetMetadataValueUint32("e", this.EndIndex)

	bs := ConvertUintToBytes(curIndex)
	this.storage.PutSeek(this.GenerateKey(bs), value)

	return nil
}

func (this *RocksQueue) Pop() []byte {

	this.queueMutex.Lock()
	defer this.queueMutex.Unlock()

	if this.BeginIndex == this.EndIndex {
		// Queue is empty
		return nil
	}

	bs := ConvertUintToBytes(this.BeginIndex)
	key, value, err := this.storage.SeekNext(this.GenerateKey(bs))

	if err != nil {
		return nil
	}

	// Delete the current value and move BeginIndex
	this.storage.DelSeek(key)

	if this.BeginIndex < RocksQueueMaxIndex {
		this.BeginIndex ++
	} else {
		this.BeginIndex = 0
	}
	this.SetMetadataValueUint32("b", this.BeginIndex)

	return value
}
