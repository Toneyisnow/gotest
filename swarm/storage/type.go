package storage

import (
)

// storage.Transaction
type Transaction interface {
	Destroy()
	Get(key []byte) ([]byte, error)
	Put(key []byte, value []byte) error
	Del(key []byte) error
	Seek(prefix []byte, f func(value []byte)) error
	RollBack() error
	Commit() error
	Close() error
}

