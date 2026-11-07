package database

import (
	"bytes"
	"crypto/aes"
	"crypto/rand"
	"encoding/binary"
	"errors"
	"io"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func testDatabase(t *testing.T, cache bool) *Database {
	t.Helper()
	dir := t.TempDir()
	index, err := os.Create(dir + "/index")
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.Create(dir + "/data")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { index.Close(); data.Close() })
	block, err := aes.NewCipher(make([]byte, 16))
	if err != nil {
		t.Fatal(err)
	}
	cacheRecords := 0
	if cache {
		cacheRecords = 10
	}
	db, err := NewDatabase(index, data, block, block, 2<<20, cacheRecords, 2<<20, false)
	if err != nil {
		t.Fatal(err)
	}
	return db
}

func TestConcurrentSelfBurningLookup(t *testing.T) {
	db := testDatabase(t, true)
	content := make([]byte, 64<<10)
	rand.Read(content)
	key, err := db.Add(&Record{Content: content, SelfBurning: true})
	if err != nil {
		t.Fatal(err)
	}
	var successes atomic.Int32
	var wg sync.WaitGroup
	start := make(chan struct{})
	for range 32 {
		wg.Go(func() {
			<-start
			if rec, err := db.Lookup(key, "curl", false); err == nil {
				if !bytes.Equal(rec.Content, content) {
					t.Error("lookup returned corrupted content")
				}
				successes.Add(1)
			}
		})
	}
	close(start)
	wg.Wait()
	if got := successes.Load(); got != 1 {
		t.Fatalf("self-burning record was returned %d times, want 1", got)
	}
}

func TestValidationBeforeLookupEffects(t *testing.T) {
	for _, burn := range []bool{false, true} {
		t.Run(map[bool]string{false: "cached", true: "self-burning"}[burn], func(t *testing.T) {
			db := testDatabase(t, true)
			key, err := db.Add(&Record{Content: []byte("secret"), LongId: true, SelfBurning: burn})
			if err != nil {
				t.Fatal(err)
			}
			denied := errors.New("wrong link type")
			_, err = db.LookupValidated(key, "curl", false, func(r *Record) error {
				if r.LongId {
					return denied
				}
				return nil
			})
			if !errors.Is(err, denied) {
				t.Fatalf("validation error = %v", err)
			}
			rec, err := db.Lookup(key, "curl", false)
			if err != nil || string(rec.Content) != "secret" {
				t.Fatalf("rejected lookup consumed record: %v", err)
			}
		})
	}
}

type pausedReadFile struct {
	File
	once    sync.Once
	read    chan struct{}
	proceed chan struct{}
}

func (f *pausedReadFile) ReadAt(p []byte, off int64) (int, error) {
	n, err := f.File.ReadAt(p, off)
	f.once.Do(func() { close(f.read); <-f.proceed })
	return n, err
}

func TestDeleteDoesNotResurrectCachedRecord(t *testing.T) {
	db := testDatabase(t, true)
	key, err := db.Add(&Record{Content: []byte("secret")})
	if err != nil {
		t.Fatal(err)
	}
	db.lru.DeleteIfExists(key)
	paused := &pausedReadFile{File: db.data, read: make(chan struct{}), proceed: make(chan struct{})}
	db.data = paused
	lookedUp := make(chan error, 1)
	go func() { _, err := db.Lookup(key, "curl", false); lookedUp <- err }()
	<-paused.read
	deleted := make(chan error, 1)
	go func() { deleted <- db.Delete(key) }()
	select {
	case err := <-deleted:
		close(paused.proceed)
		<-lookedUp
		t.Fatalf("delete completed during an in-flight lookup: %v", err)
	case <-time.After(20 * time.Millisecond):
	}
	close(paused.proceed)
	if err := <-lookedUp; err != nil {
		t.Fatal(err)
	}
	if err := <-deleted; err != nil {
		t.Fatal(err)
	}
	if _, err := db.Lookup(key, "curl", false); err == nil {
		t.Fatal("deleted record remained cached")
	}
}

type shortWriteFile struct{ File }

func (f shortWriteFile) WriteAt(p []byte, off int64) (int, error) {
	return f.File.WriteAt(p[:len(p)/2], off)
}

func TestFailedAddDoesNotPublishOrCorruptRecords(t *testing.T) {
	for _, target := range []string{"index", "data"} {
		t.Run(target, func(t *testing.T) {
			db := testDatabase(t, true)
			if _, err := db.Add(&Record{Content: []byte("existing")}); err != nil {
				t.Fatal(err)
			}
			indexLen, dataLen := db.indexLen, db.dataLen
			if target == "index" {
				db.index = shortWriteFile{db.index}
			} else {
				db.data = shortWriteFile{db.data}
			}
			if _, err := db.Add(&Record{Content: []byte("failed")}); !errors.Is(err, io.ErrShortWrite) {
				t.Fatalf("short write error = %v", err)
			}
			if db.indexLen != indexLen || db.dataLen != dataLen || db.RecordsCount() != 1 {
				t.Fatal("failed write advanced database boundaries")
			}
			if _, err := db.Lookup(1, "curl", false); err == nil {
				t.Fatal("failed write was cached")
			}
			if _, err := db.Add(&Record{Content: []byte("retry")}); err == nil {
				t.Fatal("write failure did not disable subsequent writes")
			}
			for file, want := range map[*os.File]uint64{db.rawIndex: indexLen, db.rawData: dataLen} {
				stat, err := file.Stat()
				if err != nil || uint64(stat.Size()) != want {
					t.Fatalf("rollback left wrong file length: %v", err)
				}
			}
			block, _ := aes.NewCipher(make([]byte, 16))
			reopened, err := NewDatabase(db.rawIndex, db.rawData, block, block, 2<<20, 0, 0, false)
			if err != nil {
				t.Fatal(err)
			}
			rec, err := reopened.Lookup(0, "curl", false)
			if err != nil || string(rec.Content) != "existing" {
				t.Fatalf("rollback corrupted previous record: %v", err)
			}
		})
	}
}

func TestRejectOversizedDecodedRecord(t *testing.T) {
	db := testDatabase(t, false)
	db.maxSize = 1024
	if _, err := db.index.WriteAt(make([]byte, 8), 0); err != nil {
		t.Fatal(err)
	}
	db.indexLen = 8
	// Use only a Snappy length header: reject before Decode allocates the buffer.
	var header [10]byte
	n := binary.PutUvarint(header[:], db.maxSize+recordMetadataAllowance+1)
	if _, err := db.data.WriteAt(header[:n], 0); err != nil {
		t.Fatal(err)
	}
	db.dataLen = uint64(n)
	if _, err := db.Lookup(0, "curl", false); err == nil || !strings.Contains(err.Error(), "decoded record is too long") {
		t.Fatalf("decoded size error = %v", err)
	}
}

func TestRejectInvalidDatabaseConfiguration(t *testing.T) {
	db := testDatabase(t, false)
	block, _ := aes.NewCipher(make([]byte, 16))
	for _, config := range [][3]int{{0, 0, 0}, {-1, 0, 0}, {1024, -1, 0}, {1024, 0, -1}} {
		if _, err := NewDatabase(db.rawIndex, db.rawData, block, block, config[0], config[1], config[2], false); err == nil {
			t.Fatalf("invalid size limits accepted: %v", config)
		}
	}
	if err := db.rawIndex.Truncate(1); err != nil {
		t.Fatal(err)
	}
	if _, err := NewDatabase(db.rawIndex, db.rawData, block, block, 1024, 0, 0, false); err == nil {
		t.Fatal("partial index entry accepted")
	}
}

func TestOpeningLegacyDatabaseWithUnindexedPrefix(t *testing.T) {
	db := testDatabase(t, false)
	block, _ := aes.NewCipher(make([]byte, 16))
	// The old writer could write data successfully and fail to write its index
	// entry. After restart it appended the first indexed record after that data.
	prefix := []byte("unindexed data left by a failed legacy append")
	if _, err := db.rawData.WriteAt(prefix, 0); err != nil {
		t.Fatal(err)
	}
	reopen := func() *Database {
		t.Helper()
		opened, err := NewDatabase(db.rawIndex, db.rawData, block, block, 2<<20, 0, 0, false)
		if err != nil {
			t.Fatalf("legacy database rejected: %v", err)
		}
		return opened
	}
	db = reopen()
	key, err := db.Add(&Record{Content: []byte("existing paste")})
	if err != nil || key != 0 {
		t.Fatalf("first indexed append: key=%d, err=%v", key, err)
	}
	var firstOffset [8]byte
	if _, err := db.index.ReadAt(firstOffset[:], 0); err != nil {
		t.Fatal(err)
	}
	if got := binary.LittleEndian.Uint64(firstOffset[:]); got != uint64(len(prefix)) {
		t.Fatalf("first record offset = %d, want %d", got, len(prefix))
	}
	beforeIndex, _ := os.ReadFile(db.rawIndex.Name())
	beforeData, _ := os.ReadFile(db.rawData.Name())
	db = reopen()
	for path, before := range map[string][]byte{db.rawIndex.Name(): beforeIndex, db.rawData.Name(): beforeData} {
		after, err := os.ReadFile(path)
		if err != nil || !bytes.Equal(after, before) {
			t.Fatalf("opening legacy database changed %s: %v", path, err)
		}
	}
	record, err := db.Lookup(key, "curl", false)
	if err != nil || string(record.Content) != "existing paste" {
		t.Fatalf("legacy record unreadable: %v", err)
	}
	next, err := db.Add(&Record{Content: []byte("new paste")})
	if err != nil || next != 1 {
		t.Fatalf("next append: key=%d, err=%v", next, err)
	}
	db = reopen()
	for id, want := range map[uint64]string{key: "existing paste", next: "new paste"} {
		record, err := db.Lookup(id, "curl", false)
		if err != nil || string(record.Content) != want {
			t.Fatalf("record %d unreadable after restart: %v", id, err)
		}
	}
	if err := db.Delete(key); err != nil {
		t.Fatal(err)
	}
	db = reopen()
	if _, err := db.Lookup(key, "curl", false); err == nil {
		t.Fatal("deleted first record was readable")
	}
	if record, err := db.Lookup(next, "curl", false); err != nil || string(record.Content) != "new paste" {
		t.Fatalf("later record lost after deleting first: %v", err)
	}
}

func TestOpeningDatabaseWithDamagedFirstEntry(t *testing.T) {
	db := testDatabase(t, false)
	for _, content := range []string{"old paste", "live paste"} {
		if _, err := db.Add(&Record{Content: []byte(content)}); err != nil {
			t.Fatal(err)
		}
	}
	var invalid [8]byte
	binary.LittleEndian.PutUint64(invalid[:], db.dataLen+1)
	if _, err := db.index.WriteAt(invalid[:], 0); err != nil {
		t.Fatal(err)
	}
	block, _ := aes.NewCipher(make([]byte, 16))
	opened, err := NewDatabase(db.rawIndex, db.rawData, block, block, 2<<20, 0, 0, false)
	if err != nil {
		t.Fatalf("damaged first entry blocked access to live records: %v", err)
	}
	if _, err := opened.Lookup(0, "curl", false); err == nil {
		t.Fatal("damaged record was accepted")
	}
	if record, err := opened.Lookup(1, "curl", false); err != nil || string(record.Content) != "live paste" {
		t.Fatalf("live record unreadable: %v", err)
	}
}

func TestRejectIndexWithoutData(t *testing.T) {
	db := testDatabase(t, false)
	if err := db.rawIndex.Truncate(8); err != nil {
		t.Fatal(err)
	}
	block, _ := aes.NewCipher(make([]byte, 16))
	if _, err := NewDatabase(db.rawIndex, db.rawData, block, block, 2<<20, 0, 0, false); err == nil {
		t.Fatal("index without data was accepted")
	}
}

func TestRejectCorruptOffsets(t *testing.T) {
	db := testDatabase(t, false)
	if _, err := db.Add(&Record{Content: []byte("test")}); err != nil {
		t.Fatal(err)
	}
	var offset [8]byte
	binary.LittleEndian.PutUint64(offset[:], db.dataLen+1)
	if _, err := db.index.WriteAt(offset[:], 0); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Lookup(0, "curl", false); err == nil {
		t.Fatal("out-of-bounds offset accepted")
	}
	if err := db.Delete(0); err == nil {
		t.Fatal("out-of-bounds wipe accepted")
	}
}
