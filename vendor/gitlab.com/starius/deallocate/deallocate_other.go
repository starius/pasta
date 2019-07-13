// +build !linux,!windows

package deallocate

import (
	"errors"
	"os"
)

var errNotImplemented = errors.New("deallocate not implemented")

// MakeSparse is a no-op on non-windows operating systems.
func MakeSparse(file *os.File) error {
	return nil
}

// PunchHole deallocates file segment from filesystem.
// If the function returns nil, the range is zeroed.
func PunchHole(file *os.File, off, length int64) error {
	return errNotImplemented
}
