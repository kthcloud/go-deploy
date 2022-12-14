package terraform

import (
	"io"
	"log"
	"os"
)

func Copy(srcpath, dstpath string) (err error) {
	r, err := os.Open(srcpath)
	if err != nil {
		return err
	}
	defer func(r *os.File) {
		closeErr := r.Close()
		if closeErr != nil {
			log.Println("failed to close file handle. details:", err)
		}
	}(r) // ok to ignore error: file was opened read-only.

	w, err := os.Create(dstpath)
	if err != nil {
		return err
	}

	defer func() {
		c := w.Close()
		// Report the error from Close, if any.
		// But do so only if there isn't already
		// an outgoing error.
		if c != nil && err == nil {
			err = c
		}
	}()

	_, err = io.Copy(w, r)
	return err
}
