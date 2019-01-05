package storage

import (
	"errors"
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

	LastKey []byte
	NowIndex uint32

	itemCount uint32
	queueMutex sync.Mutex
}

func ComposeRocksLevelQueue(storage *RocksStorage, queueId string) *RocksLevelQueue {

	queue := new(RocksLevelQueue)

	queue.storage = storage
	queue.containerId = queueId
	queue.containerType = RocksContainerType_LevelQueue

	queue.NowIndex = queue.GetMetadataValueUint32("n")
	queue.LastKey = queue.GetMetadataValueBytes("l")
	queue.itemCount = queue.GetMetadataValueUint32("c")

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

	curIndex := this.NowIndex
	if this.NowIndex < RocksQueueMaxIndex {
		this.NowIndex ++
	} else {
		this.NowIndex = 0
	}
	this.SetMetadataValueUint32("n", this.NowIndex)

	levelByte := ConvertUintToBytes(level)
	bs := ConvertUintToBytes(curIndex)
	key := append(levelByte, bs...)

	this.storage.PutSeek(this.GenerateKey(key), value)

	this.itemCount ++
	this.SetMetadataValueUint32("c", this.itemCount)

	return nil
}

func (this *RocksLevelQueue) Pop() []byte {

	this.queueMutex.Lock()
	defer this.queueMutex.Unlock()

	if this.itemCount <= 0 {
		// Queue is empty
		return nil
	}

	queueKey := this.GetContainerKey()
	lastKey := queueKey
	if (this.LastKey != nil) {
		lastKey = append(lastKey, this.LastKey...)
	}

	key, value, err := this.storage.SeekNext(lastKey)
	if err != nil {
		return nil
	}

	// If there is no begin key, should seek from the start
	if value == nil {
		key, value, err = this.storage.SeekNext(queueKey)
		if err != nil {
			return nil
		}
	}

	if key == nil || value == nil {
		return nil
	}

	// Delete the current value and move lastKey
	this.storage.DelSeek(key)

	this.LastKey = key
	this.SetMetadataValueBytes("l", this.LastKey)
	this.itemCount --
	this.SetMetadataValueUint32("c", this.itemCount)

	return value
}
