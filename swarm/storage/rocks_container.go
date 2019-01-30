package storage

import "strconv"

const (
	RocksContainerType_Table = 1
	RocksContainerType_SequenceQueue = 2
	RocksContainerType_LevelQueue = 3
)

type RocksContainer struct {

	storage *RocksStorage
	containerType int
	containerId string

	containerKey []byte
	metadataKey []byte
}

func (this *RocksContainer) GetContainerKey() []byte {

	if this.containerKey == nil {
		this.containerKey = []byte(strconv.Itoa(this.containerType) +  ":" + this.containerId + ":")
	}

	return this.containerKey
}

func (this *RocksContainer) GetContainerKeyLength() int {

	return len(this.GetContainerKey())
}

func (this *RocksContainer) GetMetadataKey() []byte {

	if this.metadataKey == nil {
		this.metadataKey = []byte("meta:" + strconv.Itoa(this.containerType) +  ":" + this.containerId + ":")
	}

	return this.metadataKey
}

func (this *RocksContainer) GenerateKey(subKey []byte) []byte {

	return append(this.GetContainerKey(), subKey...)
}

func (this *RocksContainer) GetSubKey(key []byte) []byte {

	containerKey := this.GetContainerKey()
	if len(key) <= len(containerKey) {
		return nil
	}

	return key[len(containerKey):]
}

func (this *RocksContainer) SetMetadataValueUint32(key string, value uint32) {

	bs := ConvertUint32ToBytes(value)
	this.storage.Put(append(this.GetMetadataKey(), []byte(key)...), bs)
}

func (this *RocksContainer) GetMetadataValueUint32(key string) uint32 {

	val, _ := this.storage.Get(append(this.GetMetadataKey(), []byte(key)...))
	if val == nil || len(val) == 0 {
		return 0
	}

	return ConvertBytesToUint32(val)
}

func (this *RocksContainer) SetMetadataValueUint64(key string, value uint64) {

	bs := ConvertUint64ToBytes(value)
	this.storage.Put(append(this.GetMetadataKey(), []byte(key)...), bs)
}

func (this *RocksContainer) GetMetadataValueUint64(key string) uint64 {

	val, _ := this.storage.Get(append(this.GetMetadataKey(), []byte(key)...))
	if val == nil || len(val) == 0 {
		return 0
	}

	return ConvertBytesToUint64(val)
}

func (this *RocksContainer) SetMetadataValueBytes(key string, value []byte) {

	this.storage.Put(append(this.GetMetadataKey(), []byte(key)...), value)
}

func (this *RocksContainer) GetMetadataValueBytes(key string) []byte {

	val, _ := this.storage.Get(append(this.GetMetadataKey(), []byte(key)...))
	if val == nil || len(val) == 0 {
		return nil
	}

	return val
}
