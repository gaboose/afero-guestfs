package aferoguestfs

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"syscall"
	"time"

	"github.com/gaboose/afero-guestfs/libguestfs.org/guestfs"
	"github.com/spf13/afero"
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
	name = normalizePath(name)
	return wrapErr(fs.guestfs.Chmod(int(posixMode(mode)), name), name)
}

// Chown implements afero.Fs.
func (fs *Fs) Chown(name string, uid int, gid int) error {
	name = normalizePath(name)
	return wrapErr(fs.guestfs.Chown(uid, gid, name), name)
}

// Chtimes implements afero.Fs.
func (fs *Fs) Chtimes(name string, atime time.Time, mtime time.Time) error {
	name = normalizePath(name)
	return wrapErr(fs.guestfs.Utimens(name, atime.Unix(), int64(atime.Nanosecond()), mtime.Unix(), int64(mtime.Nanosecond())), name)
}

// Create implements afero.Fs.
func (fs *Fs) Create(name string) (afero.File, error) {
	return fs.OpenFile(name, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
}

// Mkdir implements afero.Fs.
func (fs *Fs) Mkdir(name string, perm os.FileMode) error {
	name = normalizePath(name)
	return wrapErr(fs.guestfs.Mkdir_mode(name, int(posixMode(perm))), name)
}

// MkdirAll implements afero.Fs.
func (fs *Fs) MkdirAll(path string, perm os.FileMode) error {
	path = normalizePath(path)
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
	name = normalizePath(name)
	f, err := newFile(fs, name, flag, perm)
	return f, wrapErr(err, name)
}

// Remove implements afero.Fs.
func (fs *Fs) Remove(name string) error {
	name = normalizePath(name)
	return wrapErr(fs.guestfs.Rm(name), name)
}

// RemoveAll implements afero.Fs.
func (fs *Fs) RemoveAll(path string) error {
	path = normalizePath(path)
	return wrapErr(fs.guestfs.Rm_rf(path), path)
}

// Rename implements afero.Fs.
func (fs *Fs) Rename(oldname string, newname string) error {
	oldname = normalizePath(oldname)
	newname = normalizePath(newname)
	return wrapErr(fs.guestfs.Rename(oldname, newname), oldname)
}

// Stat implements afero.Fs.
func (fs *Fs) Stat(name string) (os.FileInfo, error) {
	name = normalizePath(name)

	// calling Exists before Statns prevents a "No such file or directory"
	// error from being printed by libguestfs
	if err := fs.exists(name); err != nil {
		return nil, err
	}

	s, err := fs.guestfs.Statns(name)
	if err != nil {
		return nil, wrapErr(err, name)
	}

	return newFileInfo(name, s), nil
}

// Lstat is the analogue of os.Lstat.
func (fs *Fs) Lstat(name string) (os.FileInfo, error) {
	name = normalizePath(name)

	// calling Exists before Lstatns prevents a "No such file or directory"
	// error from being printed by libguestfs
	if err := fs.exists(name); err != nil {
		return nil, err
	}

	s, err := fs.guestfs.Lstatns(name)
	if err != nil {
		return nil, wrapErr(err, name)
	}

	return newFileInfo(name, s), nil
}

// LstatIfPossible implements afero.Symlinker.
func (fs *Fs) LstatIfPossible(name string) (os.FileInfo, bool, error) {
	fi, err := fs.Lstat(name)
	return fi, true, err
}

// Readlink is the analogue of os.Readlink.
func (fs *Fs) Readlink(name string) (string, error) {
	name = normalizePath(name)
	target, err := fs.guestfs.Readlink(name)
	return target, wrapErr(err, name)
}

// ReadlinkIfPossible implements afero.Symlinker.
func (fs *Fs) ReadlinkIfPossible(name string) (string, error) {
	return fs.Readlink(name)
}

// Symlink is analogous to os.Symlink.
func (fs *Fs) Symlink(oldname string, newname string) error {
	newname = normalizePath(newname)
	return wrapErr(fs.guestfs.Ln_s(oldname, newname), newname)
}

// SymlinkIfPossible implements afero.Symlinker.
func (fs *Fs) SymlinkIfPossible(oldname string, newname string) error {
	return fs.Symlink(oldname, newname)
}

// Link is analogous to os.Link.
func (fs *Fs) Link(oldname string, newname string) error {
	oldname = normalizePath(oldname)
	newname = normalizePath(newname)
	return wrapErr(fs.guestfs.Ln(oldname, newname), newname)
}

// Lchown implements aferosync.Lchowner.
func (fs *Fs) Lchown(name string, uid, gid int) error {
	name = normalizePath(name)
	return wrapErr(fs.guestfs.Lchown(uid, gid, name), name)
}

// TarOut implements aferosync.TarOuter.
func (fs *Fs) TarOut(dir string, w io.Writer) error {
	dir = normalizePath(dir)

	f, err := os.CreateTemp("", "afero-guestfs-tarout-*.tar")
	if err != nil {
		return fmt.Errorf("failed to create tmp tar: %w", err)
	}
	f.Close()
	defer os.Remove(f.Name())

	if err = fs.guestfs.Tar_out(dir, f.Name(), nil); err != nil {
		return fmt.Errorf("failed to write tar: %w", err)
	}

	f, err = os.Open(f.Name())
	if err != nil {
		return fmt.Errorf("failed to open tar: %w", err)
	}

	if _, err := io.Copy(w, f); err != nil {
		return fmt.Errorf("failed to copy: %w", err)
	}

	return nil
}

func (fs *Fs) exists(name string) error {
	exists, err := fs.guestfs.Exists(name)
	if err != nil {
		return wrapErr(err, name)
	}

	if !exists {
		return &os.PathError{
			Op:   "exists",
			Path: name,
			Err:  syscall.ENOENT,
		}
	}

	return nil
}

func normalizePath(path string) string {
	path = filepath.Clean(path)
	if path == string('.') {
		path = string(filepath.Separator)
	} else if path[0] != filepath.Separator {
		path = string(append([]rune{filepath.Separator}, []rune(path)...))
	}
	return path
}

// reference: /usr/local/go/src/os/file_posix.go
func posixMode(i os.FileMode) (o uint32) {
	o |= uint32(i.Perm())
	if i&os.ModeSetuid != 0 {
		o |= syscall.S_ISUID
	}
	if i&os.ModeSetgid != 0 {
		o |= syscall.S_ISGID
	}
	if i&os.ModeSticky != 0 {
		o |= syscall.S_ISVTX
	}
	return
}
