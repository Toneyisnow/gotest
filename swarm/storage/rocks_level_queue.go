package storage

import (
	"errors"
	"sync"
	"github.com/golang/protobuf/proto"
)

const (
	RocksLevelQueueMaxIndex = ^uint32(0) - 1
)

// Rocks Queue with level information, that means the data in the queue will be sorted by the level fist and then by the index
// Sample:
// Push: (1, "aa"), (2, "bb"), (1, "cc")
// Pop Result: (0x01+0x01, "aa"), (0x01+0x03, "cc"), (0x02+0x02, "bb")

type RocksLevelQueue struct {

	RocksContainer

	lastKey []byte
	nowIndex uint32

	iterateKey []byte

	itemCount uint32
	queueMutex sync.Mutex
}

func NewRocksLevelQueue(storage *RocksStorage, queueId string) *RocksLevelQueue {

	queue := new(RocksLevelQueue)

	queue.storage = storage
	queue.containerId = queueId
	queue.containerType = RocksContainerType_LevelQueue

	queue.nowIndex = queue.GetMetadataValueUint32("n")
	queue.lastKey = queue.GetMetadataValueBytes("l")
	queue.itemCount = queue.GetMetadataValueUint32("c")
	queue.iterateKey = queue.lastKey

	return queue
}

// Push data with Level
func (this *RocksLevelQueue) Push(level uint32, value []byte) (err error) {

	this.queueMutex.Lock()
	defer this.queueMutex.Unlock()

	if this.itemCount >= RocksQueueMaxIndex {
		// Queue is full
		return errors.New("Push to queue failed: the queue is full.")
	}

	curIndex := this.nowIndex
	if this.nowIndex < RocksQueueMaxIndex {
		this.nowIndex = this.nowIndex + 1
	} else {
		this.nowIndex = 0
	}
	this.SetMetadataValueUint32("n", this.nowIndex)

	levelByte := ConvertUintToBytes(level)
	bs := ConvertUintToBytes(curIndex)
	key := append(levelByte, bs...)

	this.storage.PutSeek(this.GenerateKey(key), value)

	this.itemCount = this.itemCount + 1
	this.SetMetadataValueUint32("c", this.itemCount)

	return nil
}

func (this *RocksLevelQueue) PushProto(level uint32, pb proto.Message) (err error) {

	if data, err := proto.Marshal(pb); err != nil {
		return err
	} else {
		return this.Push(level, data)
	}
}

// Pop the value, and try to handle it with callback, if callback return false, then don't delete it from database
func (this *RocksLevelQueue) Pop() (result []byte) {

	this.queueMutex.Lock()
	defer this.queueMutex.Unlock()

	if this.itemCount <= 0 {
		// Queue is empty
		return
	}

	queueKey := this.GetContainerKey()
	lastKey := queueKey
	if this.lastKey != nil {
		lastKey = append(lastKey, this.lastKey...)
	}

	key, value, err := this.storage.SeekNext(lastKey)
	if err != nil {
		return
	}

	// If there is no begin key, should seek from the start
	if value == nil && this.itemCount > 0 {
		key, value, err = this.storage.SeekNext(queueKey)
		if err != nil {
			return
		}
	}

	if key == nil || value == nil {
		return
	}

	// Delete the current value and move lastKey
	this.storage.DelSeek(key)

	this.lastKey = key
	this.SetMetadataValueBytes("l", this.lastKey)
	this.itemCount = this.itemCount - 1
	this.SetMetadataValueUint32("c", this.itemCount)

	return value
}

func (this *RocksLevelQueue) StartIterate() {

	this.iterateKey = this.lastKey
}

func (this *RocksLevelQueue) IterateNext() (result []byte) {

	this.queueMutex.Lock()
	defer this.queueMutex.Unlock()

	if this.itemCount <= 0 {
		// Queue is empty
		return
	}

	queueKey := this.GetContainerKey()
	lastKey := queueKey
	if this.iterateKey != nil {
		lastKey = append(lastKey, this.iterateKey...)
	}

	key, value, err := this.storage.SeekNext(lastKey)
	if err != nil {
		return
	}

	// If there is no begin key, should seek from the start
	if value == nil && this.itemCount > 0 {
		key, value, err = this.storage.SeekNext(queueKey)
		if err != nil {
			return
		}
	}

	if key == nil || value == nil {
		return
	}

	return value
}

func (this *RocksLevelQueue) DataSize() uint32 {

	return this.itemCount
}