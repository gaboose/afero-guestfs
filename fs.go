package aferoguestfs

import (
	"os"
	"syscall"
	"time"

	"github.com/spf13/afero"
	"libguestfs.org/guestfs"
)

type Fs struct {
	guestfs *guestfs.Guestfs
}

func New(g *guestfs.Guestfs) *Fs {
	return &Fs{
		guestfs: g,
	}
}

// Chmod implements afero.Fs.
func (fs *Fs) Chmod(name string, mode os.FileMode) error {
	return wrapErr(fs.guestfs.Chmod(int(mode), name), name)
}

// Chown implements afero.Fs.
func (fs *Fs) Chown(name string, uid int, gid int) error {
	return wrapErr(fs.guestfs.Chown(uid, gid, name), name)
}

// Chtimes implements afero.Fs.
func (fs *Fs) Chtimes(name string, atime time.Time, mtime time.Time) error {
	return wrapErr(fs.guestfs.Utimens(name, atime.Unix(), int64(atime.Nanosecond()), mtime.Unix(), int64(mtime.Nanosecond())), name)
}

// Create implements afero.Fs.
func (fs *Fs) Create(name string) (afero.File, error) {
	return fs.OpenFile(name, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
}

// Mkdir implements afero.Fs.
func (fs *Fs) Mkdir(name string, perm os.FileMode) error {
	return wrapErr(fs.guestfs.Mkdir_mode(name, int(perm)), name)
}

// MkdirAll implements afero.Fs.
func (fs *Fs) MkdirAll(path string, perm os.FileMode) error {
	return wrapErr(fs.guestfs.Mkdir_p(path), path)
}

// Name implements afero.Fs.
func (fs *Fs) Name() string {
	return "guestfs"
}

// Open implements afero.Fs.
func (fs *Fs) Open(name string) (afero.File, error) {
	return fs.OpenFile(name, os.O_RDONLY, 0)
}

// OpenFile implements afero.Fs.
func (fs *Fs) OpenFile(name string, flag int, perm os.FileMode) (afero.File, error) {
	f, err := newFile(fs, name, flag, perm)
	return f, wrapErr(err, name)
}

// Remove implements afero.Fs.
func (fs *Fs) Remove(name string) error {
	return wrapErr(fs.guestfs.Rm(name), name)
}

// RemoveAll implements afero.Fs.
func (fs *Fs) RemoveAll(path string) error {
	return wrapErr(fs.guestfs.Rm_rf(path), path)
}

// Rename implements afero.Fs.
func (fs *Fs) Rename(oldname string, newname string) error {
	return wrapErr(fs.guestfs.Rename(oldname, newname), oldname)
}

// Stat implements afero.Fs.
func (fs *Fs) Stat(name string) (os.FileInfo, error) {
	// calling Exists before Statns prevents a "No such file or directory"
	// error from being printed by libguestfs
	exists, err := fs.guestfs.Exists(name)
	if err != nil {
		return nil, wrapErr(err, name)
	}

	if !exists {
		return nil, &os.PathError{
			Op:   "exists",
			Path: name,
			Err:  syscall.ENOENT,
		}
	}

	s, err := fs.guestfs.Statns(name)
	if err != nil {
		return nil, wrapErr(err, name)
	}

	return newFileInfo(name, s), nil
}
