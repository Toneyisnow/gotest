package storage

import (
	"../common/log"
	"bytes"
	"encoding/binary"
	"errors"
	"github.com/smartswarm/go/app"

	"github.com/czsilence/gorocksdb"
	"github.com/golang/protobuf/proto"
	//// "github.com/syndtr/goleveldb/leveldb/opt"
)

// The interface for data callback function, will deal with bytes and return true if everything is fine
type dataCallbackFunc func([]byte) bool

type batchOpt struct {
	key     []byte
	value   []byte
	deleted bool
}

// RocksStorage the nodes in trie.
type RocksStorage struct {

	db     		*gorocksdb.TransactionDB
	seekableDb  *gorocksdb.DB

	//enableBatch bool
	//mutex       sync.Mutex
	//batchOpts   map[string]*batchOpt
	//
	ro *gorocksdb.ReadOptions
	wo *gorocksdb.WriteOptions

	cache *gorocksdb.Cache
}

// Deprecated
func NewRocksDB(name string, applyOpts func(opts *gorocksdb.Options)) *gorocksdb.DB {
	//dir, _ := ioutil.TempDir("", "gorocksdb-"+name)


	dir := "/var/folders/gorocsdb-Test"
	log.I2("tempDir: %s", dir)
	// ensure.Nil(t, err)

	opts := gorocksdb.NewDefaultOptions()
	// test the ratelimiter
	rateLimiter := gorocksdb.NewRateLimiter(1024, 100*1000, 10)
	opts.SetRateLimiter(rateLimiter)
	opts.SetCreateIfMissing(true)
	if applyOpts != nil {
		applyOpts(opts)
	}
	db, _ := gorocksdb.OpenDb(opts, dir)
	// ensure.Nil(t, err)

	return db
}

//GetRocksDBInstance singleton
func GetRocksDBInstance(a app.AppDelegate) (rocksStorage *RocksStorage) {
	var err error
	if rocksStorage, err = newRocksTransactionStorage(a.GetDataPath("rdbtest")); err != nil {
		log.E("[db] init storage failed!", err)
	}
	return
}

//GetRocksDBInstance singleton
func ComposeRocksDBInstance(databaseFullPath string) (rocksStorage *RocksStorage) {
	var err error
	if rocksStorage, err = newRocksTransactionStorage(databaseFullPath); err != nil {
		log.E("[db] init storage failed!", err)
	}
	return
}

// NewRocksStorage init a storage
func newRocksTransactionStorage(databaseFullPath string) (*RocksStorage, error) {

	filter := gorocksdb.NewBloomFilter(10)
	bbto := gorocksdb.NewDefaultBlockBasedTableOptions()
	bbto.SetFilterPolicy(filter)

	cache := gorocksdb.NewLRUCache(512 << 20)
	bbto.SetBlockCache(cache)
	opts := gorocksdb.NewDefaultOptions()
	opts.SetBlockBasedTableFactory(bbto)
	opts.SetCreateIfMissing(true)
	opts.SetMaxOpenFiles(500)
	opts.SetWriteBufferSize(32 * 1024 * 1024) //Default: 4MB
	opts.IncreaseParallelism(4)           //flush and compaction thread
	transactionDBOpts := gorocksdb.NewDefaultTransactionDBOptions()

	db, err := gorocksdb.OpenTransactionDb(opts, transactionDBOpts, databaseFullPath + "_trans")
	if err != nil {
		return nil, err
	}

	seekable, err := gorocksdb.OpenDb(opts, databaseFullPath)

	rOptions := gorocksdb.NewDefaultReadOptions()
	rOptions.SetIterateUpperBound(ConvertUint32ToBytes(^uint32(0)))

	storage := &RocksStorage{
		db:     db,
		seekableDb: seekable,
		//cache:       cache,
		//enableBatch: false,
		//batchOpts:   make(map[string]*batchOpt),
		ro: gorocksdb.NewDefaultReadOptions(),
		wo: gorocksdb.NewDefaultWriteOptions(),
	}

	//go RecordMetrics(storage)

	return storage, nil
}
func (storage *RocksStorage) SeekAll(prefix []byte, f func(value []byte), prefixLength int) error {

	iter := storage.seekableDb.NewIterator(storage.ro)
	defer iter.Close()

	for iter.Seek(prefix); iter.Valid(); iter.Next() {

		key := iter.Key().Data()
		if key != nil && len(key) >= len(prefix) && bytes.Equal(key[:prefixLength], prefix[:prefixLength]) {
			f(iter.Value().Data())
		}
	}
	return nil
}

func (storage *RocksStorage) SeekNext(prefix []byte, prefixLength int) (keyResult []byte, valueResult []byte, err error) {

	iter := storage.seekableDb.NewIterator(storage.ro)
	defer iter.Close()

	for iter.Seek(prefix); iter.Valid(); iter.Next() {

		key := iter.Key().Data()
		if key != nil && len(key) >= len(prefix) && bytes.Equal(key[:prefixLength], prefix[:prefixLength]) {

			value := iter.Value().Data()

			keyResult = make([]byte, len(key))
			copy(keyResult, key)

			valueResult = make([]byte, len(value))
			copy(valueResult, value)

			return keyResult, valueResult, nil
		}
	}

	return nil, nil, nil
}

// Get return value to the key in Storage
func (storage *RocksStorage) Get(key []byte) ([]byte, error) {

	value, err := storage.db.Get(storage.ro, key)
	defer value.Free()
	if err != nil {
		return nil, err
	}
	if len(value.Data()) == 0 {
		return nil, nil
	}
	dst := make([]byte, len(value.Data()))
	copy(dst, value.Data())
	//fmt.Println(string(key), value.Data())
	//fmt.Println(string(dst), dst)
	return dst, err
}

// Put put the key-value entry to Storage
func (storage *RocksStorage) Put(key []byte, value []byte) error {

	return storage.db.Put(storage.wo, key, value)
}

// Put put the key-value entry to Storage
func (storage *RocksStorage) PutSeek(key []byte, value []byte) error {

	return storage.seekableDb.Put(storage.wo, key, value)

}

// check if entry exists
func (storage *RocksStorage) Has(key []byte) bool {
	val, err := storage.Get(key)
	return err == nil && val != nil
}

// Del delete the key in Storage.
func (storage *RocksStorage) Del(key []byte) error {

	return storage.db.Delete(storage.wo, key)
}

// Del delete the key in Storage.
func (storage *RocksStorage) DelSeek(key []byte) error {

	return storage.seekableDb.Delete(storage.wo, key)
}

func (storage *RocksStorage) Close() error {
	storage.db.Close()
	return nil
}



// 保存proto序列化数据
func (storage *RocksStorage) SaveProto(key string, pb proto.Message) error {
	if data, err := proto.Marshal(pb); err != nil {
		return err
	} else {
		storage.Put([]byte(key), data)
		return nil
	}
}

// 读取并解析proto序列化数据
func (storage *RocksStorage) LoadProto(key string, pb proto.Message) error {
	if data, err := storage.Get([]byte(key)); err != nil {
		return err
	} else if data == nil {
		// no data found
		return errors.New("no data found")
	} else if err := proto.Unmarshal(data, pb); err != nil {
		return err
	}
	return nil
}

// This should be moved to common utils method
func ConvertUint32ToBytes(value uint32) []byte {

	bs := make([]byte, 4)
	binary.BigEndian.PutUint32(bs, value)

	return bs
}

func ConvertBytesToUint32(value []byte) uint32 {
	return binary.BigEndian.Uint32(value)
}

func ConvertUint64ToBytes(value uint64) []byte {

	bs := make([]byte, 8)
	binary.BigEndian.PutUint64(bs, value)

	return bs
}

func ConvertBytesToUint64(value []byte) uint64 {
	return binary.BigEndian.Uint64(value)
}

// // EnableBatch enable batch write.
// func (storage *RocksStorage) EnableBatch() {
// 	storage.enableBatch = true
// }

// // Flush write and flush pending batch write.
// func (storage *RocksStorage) Flush() error {
// 	storage.mutex.Lock()
// 	defer storage.mutex.Unlock()

// 	if !storage.enableBatch {
// 		return nil
// 	}

// 	startAt := time.Now().UnixNano()

// 	wb := gorocksdb.NewWriteBatch()
// 	defer wb.Destroy()

// 	bl := len(storage.batchOpts)

// 	for _, opt := range storage.batchOpts {
// 		if opt.deleted {
// 			wb.Delete(opt.key)
// 		} else {
// 			wb.Put(opt.key, opt.value)
// 		}
// 	}
// 	storage.batchOpts = make(map[string]*batchOpt)

// 	err := storage.db.write(storage.wo, wb)

// 	endAt := time.Now().UnixNano()
// 	fmt.Printf("batch flush ops %d cost %d", bl, endAt-startAt)
// 	// metricsRocksdbFlushTime.Update(endAt - startAt)
// 	// metricsRocksdbFlushLen.Update(int64(bl))

// 	return err
// }

// // DisableBatch disable batch write.
// func (storage *RocksStorage) DisableBatch() {
// 	storage.mutex.Lock()
// 	defer storage.mutex.Unlock()
// 	storage.batchOpts = make(map[string]*batchOpt)

// 	storage.enableBatch = false
// }
