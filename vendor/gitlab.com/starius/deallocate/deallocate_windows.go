package deallocate

import (
	"os"
	"unsafe"

	"golang.org/x/sys/windows"
)

const (
	fsctl_set_sparse    = 0x000900c4
	fsctl_set_zero_data = 0x000980c8
)

// MakeSparse marks the file as sparse.
func MakeSparse(file *os.File) error {
	var bytesReturned uint32
	return windows.DeviceIoControl(
		windows.Handle(file.Fd()),
		fsctl_set_sparse,
		nil, 0,
		nil, 0,
		&bytesReturned, nil,
	)
}

type fileZeroDataInformation struct {
	fileOffset, beyondFinalZero int64
}

// PunchHole deallocates file segment from filesystem if possible.
func PunchHole(file *os.File, off, length int64) error {
	lpInBuffer := fileZeroDataInformation{
		fileOffset:      off,
		beyondFinalZero: off + length,
	}
	var bytesReturned uint32
	return windows.DeviceIoControl(
		windows.Handle(file.Fd()),
		fsctl_set_zero_data,
		(*byte)(unsafe.Pointer(&lpInBuffer)), 16,
		nil, 0,
		&bytesReturned, nil,
	)
}
