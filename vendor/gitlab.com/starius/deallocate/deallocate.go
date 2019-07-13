package deallocate

import "os"

func PunchHoleWithFallback(file *os.File, off, length int64) error {
	err := PunchHole(file, off, length)
	if err == nil {
		return nil
	}
	dummy := make([]byte, length)
	_, err = file.WriteAt(dummy, off)
	return err
}
