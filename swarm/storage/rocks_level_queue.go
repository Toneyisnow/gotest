package storage

import (
	"encoding/hex"
	"errors"
	"github.com/golang/protobuf/proto"
	"github.com/smartswarm/go/log"
	"sync"
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

	seedIndex uint32
	beginIndex uint64
	iterateIndex uint64

	itemCount uint32

	capacity uint32
	queueMutex sync.Mutex
}

func NewRocksLevelQueue(storage *RocksStorage, queueId string, capacity uint32) *RocksLevelQueue {

	queue := new(RocksLevelQueue)

	queue.storage = storage
	queue.containerId = queueId
	queue.containerType = RocksContainerType_LevelQueue

	queue.seedIndex = queue.GetMetadataValueUint32("s")
	queue.beginIndex = queue.GetMetadataValueUint64("b")
	queue.itemCount = queue.GetMetadataValueUint32("c")
	queue.iterateIndex = queue.beginIndex

	queue.capacity = capacity

	return queue
}

// Push data with Level
func (this *RocksLevelQueue) Push(level uint32, value []byte) (err error) {

	this.queueMutex.Lock()
	defer this.queueMutex.Unlock()

	if this.itemCount >= this.capacity {
		// Queue is full
		return errors.New("Push to queue failed: the queue is full.")
	}

	// Put the value
	levelByte := ConvertUint32ToBytes(level)
	bs := ConvertUint32ToBytes(this.seedIndex)
	key := append(levelByte, bs...)
	log.D("[storage][level queue] pushing level=", level, "value=", hex.EncodeToString(value), "subkey=", key, "key=", this.GenerateKey(key))

	err = this.storage.PutSeek(this.GenerateKey(key), value)
	if err != nil {
		return
	}

	this.itemCount ++
	this.SetMetadataValueUint32("c", this.itemCount)
	log.D("[storage][level queue] push succeeded, item count=", this.itemCount)

	// Update next beginIndex
	key64 := ConvertBytesToUint64(key)
	if key64 < this.beginIndex {
		this.beginIndex = key64
		this.SetMetadataValueUint64("b", this.beginIndex)
	}

	// Get next seedIndex
	if this.seedIndex < RocksQueueMaxIndex {
		this.seedIndex ++
	} else {
		this.seedIndex = 0
	}
	this.SetMetadataValueUint32("s", this.seedIndex)

	return nil
}

func (this *RocksLevelQueue) PushProto(level uint32, pb proto.Message) (err error) {

	if data, err := proto.Marshal(pb); err != nil {
		return err
	} else {
		return this.Push(level, data)
	}
}

// Pop the value
func (this *RocksLevelQueue) Pop() (result []byte) {

	this.queueMutex.Lock()
	defer this.queueMutex.Unlock()

	if this.itemCount <= 0 {
		// Queue is empty
		log.W("[storage][level queue] queue is empty, nothing to pop")
		return
	}

	bs := ConvertUint64ToBytes(this.beginIndex)
	log.D("[storage][level queue] pop item, seeking subkey=", bs, "key=", this.GenerateKey(bs))
	key, value, err := this.storage.SeekNext(this.GenerateKey(bs), this.GetContainerKeyLength())

	if err != nil {
		log.D("[storage][level queue] error while pop item:", err.Error())
		return
	}

	if key == nil || value == nil {
		log.D("[storage][level queue] pop item: key or value is nil")
		return nil
	}

	result = value
	log.I("[storage][level queue] pop succeeded")

	// Delete the current value and move lastKey
	this.storage.DelSeek(key)

	// Found value, update index
	this.itemCount = this.itemCount - 1
	this.SetMetadataValueUint32("c", this.itemCount)

	log.D("[storage][level queue] key:", key)
	subKey := this.GetSubKey(key)
	log.D("[storage][level queue] sub key:", subKey)
	this.beginIndex = ConvertBytesToUint64(subKey) + 1
	this.SetMetadataValueUint64("b", this.beginIndex)
	log.D("[storage][level queue] update begin index to:", this.beginIndex)

	return
}

func (this *RocksLevelQueue) StartIterate() {

	this.iterateIndex = this.beginIndex
}

func (this *RocksLevelQueue) IterateNext() (resultIndex uint64, result []byte) {

	this.queueMutex.Lock()
	defer this.queueMutex.Unlock()

	if this.itemCount <= 0 {
		// Queue is empty
		log.D("[storage][level queue] iterate: queue is empty")
		return
	}

	bs := ConvertUint64ToBytes(this.iterateIndex)
	key, value, err := this.storage.SeekNext(this.GenerateKey(bs), this.GetContainerKeyLength())

	if err != nil {
		return
	}

	if key == nil || value == nil {
		return 0, nil
	}

	log.D("[storage][level queue] iterate: value found")

	// Found value, update index
	subKey := this.GetSubKey(key)

	resultIndex = ConvertBytesToUint64(subKey)
	result = value

	this.iterateIndex = resultIndex + 1

	return
}


func (this *RocksLevelQueue) Delete(index uint64) {

	bs := ConvertUint64ToBytes(index)
	key, value, err := this.storage.SeekNext(this.GenerateKey(bs), this.GetContainerKeyLength())

	if err != nil {
		return
	}

	if key == nil || value == nil {
		return
	}

	// Delete the current value and move beginIndex
	err = this.storage.DelSeek(key)

	this.itemCount --
	this.SetMetadataValueUint32("c", this.itemCount)
}

func (this *RocksLevelQueue) DataSize() uint32 {

	return this.itemCount
}

func (this *RocksLevelQueue) IsFull() bool {

	return this.itemCount >= this.capacity
}