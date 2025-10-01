package aferoguestfs

import (
	"io/fs"
	"path/filepath"
	"syscall"
	"time"

	"github.com/gaboose/afero-guestfs/libguestfs.org/guestfs"
)

type fileInfo struct {
	name string
	stat *guestfs.StatNS
}

func newFileInfo(name string, stat *guestfs.StatNS) *fileInfo {
	return &fileInfo{
		name: filepath.Base(name),
		stat: stat,
	}
}

// IsDir implements fs.FileInfo.
func (fi *fileInfo) IsDir() bool {
	return fi.stat.St_mode&syscall.S_IFMT == syscall.S_IFDIR
}

// ModTime implements fs.FileInfo.
func (fi *fileInfo) ModTime() time.Time {
	return time.Unix(fi.stat.St_mtime_sec, fi.stat.St_mtime_nsec)
}

// Mode implements fs.FileInfo.
func (fi *fileInfo) Mode() fs.FileMode {
	// Reference: https://github.com/golang/go/blob/master/src/os/stat_linux.go
	fileMode := fs.FileMode(fi.stat.St_mode & 0777)
	switch fi.stat.St_mode & syscall.S_IFMT {
	case syscall.S_IFBLK:
		fileMode |= fs.ModeDevice
	case syscall.S_IFCHR:
		fileMode |= fs.ModeDevice | fs.ModeCharDevice
	case syscall.S_IFDIR:
		fileMode |= fs.ModeDir
	case syscall.S_IFIFO:
		fileMode |= fs.ModeNamedPipe
	case syscall.S_IFLNK:
		fileMode |= fs.ModeSymlink
	case syscall.S_IFREG:
		// nothing to do
	case syscall.S_IFSOCK:
		fileMode |= fs.ModeSocket
	}
	if fi.stat.St_mode&syscall.S_ISGID != 0 {
		fileMode |= fs.ModeSetgid
	}
	if fi.stat.St_mode&syscall.S_ISUID != 0 {
		fileMode |= fs.ModeSetuid
	}
	if fi.stat.St_mode&syscall.S_ISVTX != 0 {
		fileMode |= fs.ModeSticky
	}

	return fileMode
}

// Name implements fs.FileInfo.
func (fi *fileInfo) Name() string {
	return filepath.Base(filepath.FromSlash(fi.name))
}

// Size implements fs.FileInfo.
func (fi *fileInfo) Size() int64 {
	return fi.stat.St_size
}

// Sys returns the underlying *guestfs.StatNS.
func (f *fileInfo) Sys() any {
	return f.stat
}

// UID implements aferosync.FileInfoOwner
func (f *fileInfo) Uid() int64 {
	return f.stat.St_uid
}

// GID implements aferosync.FileInfoOwner
func (f *fileInfo) Gid() int64 {
	return f.stat.St_gid
}

// Ino implements aferosync.FileInfoInoer.
func (f *fileInfo) Ino() int64 {
	return f.stat.St_ino
}
