package storage

import (
	"errors"
	"github.com/golang/protobuf/proto"
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

	iterateIndex uint32
	itemCount  uint32

	capacity uint32

	queueMutex sync.Mutex
}

func NewRocksSequenceQueue(st *RocksStorage, queueId string, capacity uint32) *RocksSequenceQueue {

	queue := new(RocksSequenceQueue)

	queue.storage = st
	queue.containerId = queueId
	queue.containerType = RocksContainerType_SequenceQueue

	queue.capacity = capacity
	queue.beginIndex = queue.GetMetadataValueUint32("b")
	queue.endIndex = queue.GetMetadataValueUint32("e")
	queue.itemCount = queue.GetMetadataValueUint32("c")

	queue.iterateIndex = queue.beginIndex

	return queue
}

func (this *RocksSequenceQueue) Push(value []byte) (err error) {

	this.queueMutex.Lock()
	defer this.queueMutex.Unlock()

	if this.itemCount >= this.capacity {
		// The queue is full
		return errors.New("push to queue failed: the queue is full")
	}

	// Put the new value
	bs := ConvertUint32ToBytes(this.endIndex)
	this.storage.PutSeek(this.GenerateKey(bs), value)

	if this.endIndex < RocksQueueMaxIndex {
		this.endIndex ++
	} else {
		this.endIndex = 0
	}
	this.SetMetadataValueUint32("e", this.endIndex)

	this.itemCount ++
	this.SetMetadataValueUint32("c", this.itemCount)

	return nil
}

func (this *RocksSequenceQueue) PushProto(pb proto.Message) (err error) {

	if data, err := proto.Marshal(pb); err != nil {
		return err
	} else {
		return this.Push(data)
	}
}

func (this *RocksSequenceQueue) Pop() (result []byte) {

	this.queueMutex.Lock()
	defer this.queueMutex.Unlock()

	if this.itemCount == 0 || this.beginIndex == this.endIndex {
		// Queue is empty
		return nil
	}

	bs := ConvertUint32ToBytes(this.beginIndex)
	key, value, err := this.storage.SeekNext(this.GenerateKey(bs))

	if err != nil {
		return
	}

	if key == nil || value == nil {

		this.beginIndex = 0
		bs := ConvertUint32ToBytes(this.beginIndex)
		key, value, err = this.storage.SeekNext(this.GenerateKey(bs))

		if key == nil || value == nil {
			return nil
		}
	}

	// Delete the current value and move beginIndex
	err = this.storage.DelSeek(key)

	// Find the next beginIndex
	subKey := this.GetSubKey(key)
	this.beginIndex = ConvertBytesToUint32(subKey)
	if this.beginIndex < RocksQueueMaxIndex {
		this.beginIndex ++
	} else {
		this.beginIndex = 0
	}
	this.SetMetadataValueUint32("b", this.beginIndex)

	this.itemCount --
	this.SetMetadataValueUint32("c", this.itemCount)

	return value
}

// Iterate the values from beginIndex to endIndex
func (this *RocksSequenceQueue) StartIterate() {

	this.iterateIndex = this.beginIndex
}

func (this *RocksSequenceQueue) IterateNext() (index uint32, result []byte) {

	this.queueMutex.Lock()
	defer this.queueMutex.Unlock()

	if this.iterateIndex == this.endIndex {
		// No more value to get
		return 0, nil
	}

	bs := ConvertUint32ToBytes(this.iterateIndex)
	key, value, err := this.storage.SeekNext(this.GenerateKey(bs))

	if err != nil {
		return 0, nil
	}

	if key == nil || value == nil {
		this.iterateIndex = 0
		bs := ConvertUint32ToBytes(this.iterateIndex)
		key, value, err = this.storage.SeekNext(this.GenerateKey(bs))

		if key == nil || value == nil {
			return 0, nil
		}
	}

	// Get the iterate key
	subKey := this.GetSubKey(key)
	this.iterateIndex = ConvertBytesToUint32(subKey)

	index = this.iterateIndex
	result = value

	if this.iterateIndex < RocksQueueMaxIndex {
		this.iterateIndex ++
	} else {
		this.iterateIndex = 0
	}

	return
}

func (this *RocksSequenceQueue) Delete(index uint32) {

	bs := ConvertUint32ToBytes(index)
	key, value, err := this.storage.SeekNext(this.GenerateKey(bs))

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

func (this *RocksSequenceQueue) DataSize() uint32 {

	return this.itemCount
}