package database

import (
	"crypto/cipher"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/golang/protobuf/proto"
	"github.com/golang/snappy"
	"gitlab.com/starius/deallocate"
)

//go:generate protoc --proto_path=. --go_out=. record.proto

//go:generate go get gitlab.com/starius/lru-gen
//go:generate lru-gen -package database -key "uint64" -value "*Record" -output lru.go

type File interface {
	io.ReaderAt
	io.WriterAt
}

type Database struct {
	index, data       File
	rawData           *os.File
	indexLen, dataLen uint64
	mu                sync.Mutex
	maxSize           uint64
	lru               *LRU
}

func NewDatabase(index, data *os.File, indexBlock, dataBlock cipher.Block, maxSize, cacheRecords, cacheBytes int) (*Database, error) {
	indexStat, err := index.Stat()
	if err != nil {
		return nil, err
	}
	dataStat, err := data.Stat()
	if err != nil {
		return nil, err
	}
	if err := deallocate.MakeSparse(data); err != nil {
		return nil, fmt.Errorf("failed to mark file sparse: %v", err)
	}
	db := &Database{
		index:    WrapInCTR(indexBlock, index),
		data:     WrapInCTR(dataBlock, data),
		rawData:  data,
		indexLen: uint64(indexStat.Size()),
		dataLen:  uint64(dataStat.Size()),
		maxSize:  uint64(maxSize),
	}
	if cacheRecords != 0 && cacheBytes != 0 {
		lru, err := NewLRU(uint64(cacheRecords), uint64(cacheBytes))
		if err != nil {
			return nil, err
		}
		db.lru = lru
	}
	return db, nil
}

func recordSize(record *Record) uint64 {
	return uint64(len(record.Filename) + len(record.Content))
}

func (d *Database) getOffsets(key, indexLen, dataLen uint64) (dataBegin, dataEnd uint64, err error) {
	if key >= indexLen/8 {
		return 0, 0, fmt.Errorf("the record does not exist")
	}
	indexBuffer := make([]byte, 8+8)
	indexBuffer2 := indexBuffer
	if key*8 == indexLen-8 {
		// Last element.
		indexBuffer2 = indexBuffer2[0:8]
	}
	if _, err := d.index.ReadAt(indexBuffer2, int64(key*8)); err != nil {
		return 0, 0, err
	}
	dataBegin = binary.LittleEndian.Uint64(indexBuffer[0:8])
	dataEnd = binary.LittleEndian.Uint64(indexBuffer[8:16])
	if key*8 == indexLen-8 {
		// Last element.
		dataEnd = dataLen
	}
	if dataBegin == 0 && key != 0 {
		return 0, 0, fmt.Errorf("the record does not exist in the middle")
	}
	size := dataEnd - dataBegin
	if size > d.maxSize {
		return 0, 0, fmt.Errorf("the record seems to be too long")
	}
	return dataBegin, dataEnd, nil
}

func (d *Database) wipeData(dataBegin, dataEnd uint64) error {
	return deallocate.PunchHoleWithFallback(d.rawData, int64(dataBegin), int64(dataEnd-dataBegin))
}

func (d *Database) Lookup(key uint64) (*Record, error) {
	if d.lru != nil {
		d.mu.Lock()
		r, has := d.lru.Get(key)
		d.mu.Unlock()
		if has {
			return r, nil
		}
	}
	d.mu.Lock()
	indexLen := d.indexLen
	dataLen := d.dataLen
	d.mu.Unlock()
	dataBegin, dataEnd, err := d.getOffsets(key, indexLen, dataLen)
	if err != nil {
		return nil, err
	}
	dataBuffer := make([]byte, dataEnd-dataBegin)
	if _, err := d.data.ReadAt(dataBuffer, int64(dataBegin)); err != nil {
		return nil, err
	}
	dataBuffer, err = snappy.Decode(nil, dataBuffer)
	if err != nil {
		return nil, err
	}
	var record Record
	if err := proto.Unmarshal(dataBuffer, &record); err != nil {
		return nil, err
	}
	if record.SelfBurning {
		if err := d.wipeData(dataBegin, dataEnd); err != nil {
			return nil, err
		}
	} else if d.lru != nil {
		d.mu.Lock()
		d.lru.Set(key, &record, recordSize(&record))
		d.mu.Unlock()
	}
	return &record, nil
}

func (d *Database) Delete(key uint64) error {
	d.mu.Lock()
	if d.lru != nil {
		d.lru.DeleteIfExists(key)
	}
	indexLen := d.indexLen
	dataLen := d.dataLen
	d.mu.Unlock()
	dataBegin, dataEnd, err := d.getOffsets(key, indexLen, dataLen)
	if err != nil {
		return err
	}
	if err := d.wipeData(dataBegin, dataEnd); err != nil {
		return err
	}
	return nil
}

func (d *Database) Add(record *Record) (uint64, error) {
	data, err := proto.Marshal(record)
	if err != nil {
		return 0, err
	}
	data = snappy.Encode(nil, data)
	if uint64(len(data)) > d.maxSize {
		return 0, fmt.Errorf("the record seems to be too long")
	}
	d.mu.Lock()
	indexLen := d.indexLen
	d.indexLen += 8
	key := indexLen / 8
	dataLen := d.dataLen
	d.dataLen += uint64(len(data))
	if !record.SelfBurning && d.lru != nil {
		d.lru.Set(key, record, recordSize(record))
	}
	d.mu.Unlock()
	if _, err := d.data.WriteAt(data, int64(dataLen)); err != nil {
		return 0, err
	}
	indexBuffer := make([]byte, 8)
	binary.LittleEndian.PutUint64(indexBuffer, dataLen)
	if _, err := d.index.WriteAt(indexBuffer, int64(indexLen)); err != nil {
		return 0, err
	}
	return key, nil
}

func (d *Database) RecordsCount() int64 {
	d.mu.Lock()
	indexLen := d.indexLen
	d.mu.Unlock()
	return int64(indexLen) / 8
}
