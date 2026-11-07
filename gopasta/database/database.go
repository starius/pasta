package database

import (
	"crypto/cipher"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/golang/snappy"
	"github.com/starius/pasta/gopasta/blockedbots"
	"gitlab.com/starius/deallocate"
	"google.golang.org/protobuf/proto"
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
	rawIndex          *os.File
	rawData           *os.File
	indexLen, dataLen uint64
	mu                sync.Mutex
	maxSize           uint64
	lru               *LRU
	allBurn           bool
	writeErr          error
}

// Uploaded content is bounded separately from its serialized metadata. Allow
// room for protobuf fields and MIME types when reading existing records.
const recordMetadataAllowance = 64 << 10

func NewDatabase(index, data *os.File, indexBlock, dataBlock cipher.Block, maxSize, cacheRecords, cacheBytes int, allBurn bool) (*Database, error) {
	if maxSize <= 0 || cacheRecords < 0 || cacheBytes < 0 {
		return nil, fmt.Errorf("invalid database size limits")
	}
	indexStat, err := index.Stat()
	if err != nil {
		return nil, err
	}
	if indexStat.Size()%8 != 0 {
		return nil, fmt.Errorf("invalid index file length")
	}
	dataStat, err := data.Stat()
	if err != nil {
		return nil, err
	}
	if indexStat.Size() != 0 && dataStat.Size() == 0 {
		return nil, fmt.Errorf("index entries without data")
	}
	// Legacy failed appends can leave unindexed data before the first record.
	// The first offset is not a key verifier; validate each requested record's
	// offsets in getOffsets instead of rejecting the entire database here.
	if err := deallocate.MakeSparse(data); err != nil {
		return nil, fmt.Errorf("failed to mark file sparse: %v", err)
	}
	db := &Database{
		index:    WrapInCTR(indexBlock, index),
		data:     WrapInCTR(dataBlock, data),
		rawIndex: index,
		rawData:  data,
		indexLen: uint64(indexStat.Size()),
		dataLen:  uint64(dataStat.Size()),
		maxSize:  uint64(maxSize),
		allBurn:  allBurn,
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
	if dataBegin >= dataEnd || dataEnd > dataLen {
		return 0, 0, fmt.Errorf("invalid record offsets")
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

func (d *Database) Lookup(key uint64, userAgent string, probeRequest bool) (rec *Record, err error) {
	return d.LookupValidated(key, userAgent, probeRequest, nil)
}

// LookupValidated checks request-specific authorization before returning or
// consuming a record. The validator must not call back into the database.
func (d *Database) LookupValidated(key uint64, userAgent string, probeRequest bool, validate func(*Record) error) (rec *Record, err error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.lru != nil {
		r, has := d.lru.Get(key)
		if has {
			if validate != nil {
				if err := validate(r); err != nil {
					return nil, err
				}
			}
			return r, nil
		}
	}
	dataBegin, dataEnd, err := d.getOffsets(key, d.indexLen, d.dataLen)
	if err != nil {
		return nil, err
	}
	dataBuffer := make([]byte, dataEnd-dataBegin)
	if n, err := d.data.ReadAt(dataBuffer, int64(dataBegin)); err != nil {
		return nil, err
	} else if n != len(dataBuffer) {
		return nil, io.ErrUnexpectedEOF
	}
	decodedLen, err := snappy.DecodedLen(dataBuffer)
	if err != nil {
		return nil, err
	}
	if uint64(decodedLen) > d.maxSize+recordMetadataAllowance {
		return nil, fmt.Errorf("decoded record is too long")
	}
	dataBuffer, err = snappy.Decode(nil, dataBuffer)
	if err != nil {
		return nil, err
	}
	var record Record
	if err := proto.Unmarshal(dataBuffer, &record); err != nil {
		return nil, err
	}
	if validate != nil {
		if err := validate(&record); err != nil {
			return nil, err
		}
	}
	if record.SelfBurning || d.allBurn {
		if blockedbots.IsBlockedCrawler(userAgent) {
			return nil, fmt.Errorf("preventing link burning by a crawler")
		}
		if probeRequest {
			// For DuckDuckGo probe request, return content of the
			// same size filled with zero bytes.
			// See https://github.com/duckduckgo/Android/issues/5164
			record.Content = make([]byte, len(record.Content))
			return &record, nil
		}
		if err := d.wipeData(dataBegin, dataEnd); err != nil {
			return nil, err
		}
	} else if d.lru != nil {
		d.lru.Set(key, &record, recordSize(&record))
	}
	return &record, nil
}

func (d *Database) Delete(key uint64) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	dataBegin, dataEnd, err := d.getOffsets(key, d.indexLen, d.dataLen)
	if err != nil {
		return err
	}
	if d.lru != nil {
		d.lru.DeleteIfExists(key)
	}
	if err := d.wipeData(dataBegin, dataEnd); err != nil {
		return err
	}
	return nil
}

func (d *Database) Add(record *Record) (uint64, error) {
	if uint64(proto.Size(record)) > d.maxSize+recordMetadataAllowance {
		return 0, fmt.Errorf("decoded record is too long")
	}
	data, err := proto.Marshal(record)
	if err != nil {
		return 0, err
	}
	data = snappy.Encode(nil, data)
	if uint64(len(data)) > d.maxSize {
		return 0, fmt.Errorf("the record seems to be too long")
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.writeErr != nil {
		return 0, d.writeErr
	}
	indexLen := d.indexLen
	key := indexLen / 8
	dataLen := d.dataLen
	const maxFileSize = uint64(1<<63 - 1)
	if indexLen > maxFileSize-8 || uint64(len(data)) > maxFileSize-dataLen {
		return 0, fmt.Errorf("database is too large")
	}
	rollback := func(writeErr error) (uint64, error) {
		// Never reuse an offset after a failed encrypted write in this process.
		// Keep existing records readable, but require operator intervention
		// before accepting further writes.
		d.writeErr = fmt.Errorf("database append failed; writes disabled: %w", writeErr)
		// Restore the previous last record's boundary on disk as well as in
		// memory. A partial append must not become visible after a restart.
		rollbackErr := errors.Join(d.rawData.Truncate(int64(dataLen)), d.rawIndex.Truncate(int64(indexLen)))
		if rollbackErr != nil {
			d.writeErr = errors.Join(d.writeErr, fmt.Errorf("database rollback failed: %w", rollbackErr))
		}
		return 0, d.writeErr
	}
	if n, err := d.data.WriteAt(data, int64(dataLen)); err != nil {
		return rollback(err)
	} else if n != len(data) {
		return rollback(io.ErrShortWrite)
	}
	indexBuffer := make([]byte, 8)
	binary.LittleEndian.PutUint64(indexBuffer, dataLen)
	if n, err := d.index.WriteAt(indexBuffer, int64(indexLen)); err != nil {
		return rollback(err)
	} else if n != len(indexBuffer) {
		return rollback(io.ErrShortWrite)
	}
	d.indexLen += 8
	d.dataLen += uint64(len(data))
	if !(record.SelfBurning || d.allBurn) && d.lru != nil {
		d.lru.Set(key, record, recordSize(record))
	}
	return key, nil
}

func (d *Database) RecordsCount() int64 {
	d.mu.Lock()
	indexLen := d.indexLen
	d.mu.Unlock()
	return int64(indexLen) / 8
}
