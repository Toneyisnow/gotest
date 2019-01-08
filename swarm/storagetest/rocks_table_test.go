package storagetest


import (
	"../storage"
	"github.com/smartswarm/go/log"
	"testing"
)

// This is a simple test case that shows how to use LevelQueue
// The output of the Pop() should be sorted first by the level, and then by the ID, so the output
// of the following input should be: (1, 111) (1, 333) (2, 222)
func TestTable_Success(t *testing.T) {

	rstorage := storage.ComposeRocksDBInstance("storage_test")

	table := storage.NewRocksTable(rstorage, "sampleTable")
	table.InsertOrUpdate([]byte("key1"), []byte("111"))

	val := table.Get([]byte("key1"))

	log.I("Table got value: ", val)
}

