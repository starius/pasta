package database

import (
	"bytes"
	"testing"
)

func TestEncryptedFileRoundTripPreservesWriteBuffer(t *testing.T) {
	db := testDatabase(t, false)
	content := []byte("unaligned encrypted file contents")
	original := bytes.Clone(content)
	for _, offset := range []int64{0, 1, 15, 16, 17, 31} {
		if _, err := db.data.WriteAt(content, offset); err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(content, original) {
			t.Fatal("encrypted writer modified its caller's buffer")
		}
		got := make([]byte, len(content))
		if _, err := db.data.ReadAt(got, offset); err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(got, content) {
			t.Fatalf("round trip at offset %d returned %q", offset, got)
		}
	}
	if _, err := db.data.ReadAt(make([]byte, 1), -1); err == nil {
		t.Fatal("negative read offset accepted")
	}
	if _, err := db.data.WriteAt(content, -1); err == nil {
		t.Fatal("negative write offset accepted")
	}
}
