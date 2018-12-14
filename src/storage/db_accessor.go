package storage

import (
	"github.com/czsilence/gorocksdb"
	"storage/pb"
)

type DBAccessor struct {

	_database *gorocksdb.DB

	_writOptions *gorocksdb.WriteOptions
	_readOptions *gorocksdb.ReadOptions
}

func (this *DBAccessor) Open() {

	dir := "/var/folders/gorocsdb-Test"

	opts := gorocksdb.NewDefaultOptions()
	// test the ratelimiter
	rateLimiter := gorocksdb.NewRateLimiter(1024, 100*1000, 10)
	opts.SetRateLimiter(rateLimiter)
	opts.SetCreateIfMissing(true)

	db, _ := gorocksdb.OpenDb(opts, dir)
	// ensure.Nil(t, err)

	this._database = db

	this._writOptions = gorocksdb.NewDefaultWriteOptions()
	this._readOptions = gorocksdb.NewDefaultReadOptions()
}

func (this *DBAccessor) Close() {

	this._database.Close()
}

func (this *DBAccessor) SaveRawEvent(raw *pb.DBEventRaw) {

	// create
	this._database.Put(this._writOptions, givenKey, givenVal1)
}

