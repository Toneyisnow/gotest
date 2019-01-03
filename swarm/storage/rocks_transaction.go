package storage

import (
	"github.com/czsilence/gorocksdb"
)

type RocksDBTransactionProxy struct {
	transaction *gorocksdb.Transaction
}

func (transProxy *RocksDBTransactionProxy) Destroy() {
	transProxy.transaction.Destroy()
}

func (transProxy *RocksDBTransactionProxy) Get(key []byte) ([]byte, error) {
	value, err := transProxy.transaction.Get(gorocksdb.NewDefaultReadOptions(), key)
	if err != nil {
		return nil, err
	}
	return value.Data(), err
}

func (transProxy *RocksDBTransactionProxy) Put(key []byte, value []byte) error {
	//log.D2("rocks put2, key:%s, value:%+v", string(key), value)
	return transProxy.transaction.Put(key, value)
}

func (transProxy *RocksDBTransactionProxy) Del(key []byte) error {
	return transProxy.transaction.Delete(key)
}
func (transProxy *RocksDBTransactionProxy) RollBack() error {
	return transProxy.transaction.Rollback()
}

func (transProxy *RocksDBTransactionProxy) Commit() error {
	return transProxy.transaction.Commit()
}

// Seek prefix query
func (transProxy *RocksDBTransactionProxy) Seek(prefix []byte, f func(v []byte)) error {
	panic("not supported in rocksdb transaction")
	return nil
}

func (transProxy *RocksDBTransactionProxy) Close() error {
	return nil
}
func (storage *RocksStorage) DoTransaction(f func(transction Transaction) error) error {
	txn := storage.db.TransactionBegin(gorocksdb.NewDefaultWriteOptions(), gorocksdb.NewDefaultTransactionOptions(), nil)
	defer txn.Destroy()
	transProxy := &RocksDBTransactionProxy{transaction: txn}
	err := f(transProxy)
	if err == nil {
		err = txn.Commit()
	} else {
		err = txn.Rollback()
	}
	return err
}
