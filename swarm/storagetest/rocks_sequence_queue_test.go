package storagetest

import (
	"../storage"
	"github.com/smartswarm/go/log"
	"testing"
)

func TestSequenceQueue_Success(t *testing.T) {

	rstorage := storage.ComposeRocksDBInstance("storage_test")
	sQueue := storage.ComposeRocksQueue(rstorage, "iiiQueue")
	sQueue.Push([]byte("111"))
	sQueue.Push([]byte("222"))
	sQueue.Push([]byte("333"))

	result := sQueue.Pop()
	log.I("Pop result: ", result)
	result = sQueue.Pop()
	log.I("Pop result: ", result)
	result = sQueue.Pop()
	log.I("Pop result: ", result)
	result = sQueue.Pop()
	log.I("Pop result: ", result)

}
