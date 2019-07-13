package deallocate

import (
	"os"

	"golang.org/x/sys/unix"
)

// MakeSparse is a no-op on non-windows operating systems.
func MakeSparse(file *os.File) error {
	return nil
}

// PunchHole deallocates file segment from filesystem.
// If the function returns nil, the range is zeroed.
func PunchHole(file *os.File, off, length int64) error {
	return unix.Fallocate(
		int(file.Fd()),
		unix.FALLOC_FL_PUNCH_HOLE|unix.FALLOC_FL_KEEP_SIZE,
		off, length,
	)
}
