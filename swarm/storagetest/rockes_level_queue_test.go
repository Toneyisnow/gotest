package storagetest

import (
	"../storage"
	"github.com/smartswarm/go/log"
	"testing"
)

// This is a simple test case that shows how to use LevelQueue
// The output of the Pop() should be sorted first by the level, and then by the ID, so the output
// of the following input should be: (1, 111) (1, 333) (2, 222)
func TestLevelQueue_Success(t *testing.T) {

	rstorage := storage.ComposeRocksDBInstance("storage_test")

	levelQueue := storage.NewRocksLevelQueue(rstorage, "incomingVertexQueue")
	levelQueue.Push(1, []byte("333"))
	levelQueue.Push(2, []byte("222"))
	levelQueue.Push(1, []byte("111"))

	result := levelQueue.Pop()
	log.I("LevelQueue: Pop result: ", result)
	result = levelQueue.Pop()
	log.I("LevelQueue: Pop result: ", result)
	result = levelQueue.Pop()
	log.I("LevelQueue: Pop result: ", result)

}
