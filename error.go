package aferoguestfs

import (
	"os"

	"github.com/gaboose/afero-guestfs/libguestfs.org/guestfs"
)

type guestfsError struct {
	err *guestfs.GuestfsError
}

func (e *guestfsError) Error() string {
	return e.err.Error()
}

func (e *guestfsError) String() string {
	return e.err.String()
}

func (e *guestfsError) Unwrap() []error {
	return []error{e.err, e.err.Errno}
}

// Converts *guestfs.GuestfsError into *os.PathError to make it usable with:
// - errors.Is(err, os.ErrNotExist)
// - os.IsNotExist(err).
//
// Note that we can't just wrap *guestfs.GuestfsError and syscall.Errno with
// a new error type because that breaks os.IsNotExist and afero.Exists depends
// on it.
func wrapErr(err error, path string) error {
	switch typErr := err.(type) {
	case *guestfs.GuestfsError:
		return &os.PathError{
			Op:   typErr.Op,
			Path: path,
			Err:  typErr.Errno,
		}
	}

	return err
}
