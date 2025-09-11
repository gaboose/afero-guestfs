package aferoguestfs

import (
	"errors"
	"io"
	"os"
	"path/filepath"
)

type file struct {
	fs   *Fs
	name string
	flag int
	perm os.FileMode

	stat os.FileInfo

	buf      []byte
	pos      int64
	modified bool
}

func newFile(fs *Fs, name string, flag int, perm os.FileMode) (*file, error) {
	ret := &file{
		fs:   fs,
		name: name,
		flag: flag,
		perm: perm,
	}

	fileMustNotExist := flag&os.O_CREATE != 0 && flag&os.O_EXCL != 0
	fileMustExist := flag&os.O_CREATE == 0

	fileExists, err := fs.guestfs.Exists(name)
	if err != nil {
		return nil, wrapErr(err, name)
	}

	if fileMustExist && !fileExists {
		return nil, os.ErrNotExist
	} else if fileMustNotExist && fileExists {
		return nil, os.ErrExist
	}

	if !fileExists {
		if flag&os.O_CREATE != 0 {
			// trigger file creation on close even if nothing is written
			ret.modified = true
		}

		return ret, nil
	}

	s, err := fs.guestfs.Statns(name)
	if err != nil {
		return nil, wrapErr(err, name)
	}

	ret.stat = newFileInfo(name, s)

	if ret.stat.IsDir() && ret.writeAllowed() {
		return nil, errors.New("is a directory")
	}

	if ret.stat.Mode().IsRegular() && flag&os.O_TRUNC == 0 {
		ret.buf, err = fs.guestfs.Read_file(name)
		if err != nil {
			return nil, wrapErr(err, name)
		}
	}

	if flag&os.O_APPEND != 0 {
		ret.pos = int64(len(ret.buf))
	}

	return ret, nil
}

func (f *file) writeAllowed() bool {
	return f.flag&(os.O_RDWR|os.O_WRONLY) != 0
}

func (f *file) readAllowed() bool {
	return f.flag&os.O_WRONLY == 0
}

// Close implements afero.File.
func (f *file) Close() error {
	if f.modified {
		if err := f.fs.guestfs.Write(f.name, f.buf); err != nil {
			return wrapErr(err, f.name)
		}
	}
	return nil
}

// Name implements afero.File.
func (f *file) Name() string {
	return f.name
}

// Read implements afero.File.
func (f *file) Read(p []byte) (n int, err error) {
	n, err = f.ReadAt(p, f.pos)
	f.pos += int64(n)
	return
}

// ReadAt implements afero.File.
func (f *file) ReadAt(p []byte, off int64) (n int, err error) {
	if !f.readAllowed() {
		return 0, os.ErrInvalid
	}
	if off >= int64(len(f.buf)) {
		return 0, io.EOF
	}
	n = copy(p, f.buf[off:])
	return
}

// Readdir implements afero.File.
func (f *file) Readdir(count int) ([]os.FileInfo, error) {
	names, err := f.Readdirnames(count)
	if err != nil {
		return nil, err
	}

	ret := make([]os.FileInfo, 0, len(names))
	for _, n := range names {
		fi, err := f.fs.Stat(filepath.Join(f.name, n))
		if err != nil {
			return nil, err
		}

		ret = append(ret, fi)
	}

	return ret, nil
}

// Readdirnames implements afero.File.
func (f *file) Readdirnames(n int) ([]string, error) {
	names, err := f.fs.guestfs.Ls(f.name)
	if err != nil {
		return nil, wrapErr(err, f.name)
	}
	if n > 0 && len(names) > n {
		names = names[:n]
	}
	return names, nil
}

// Seek implements afero.File.
func (f *file) Seek(offset int64, whence int) (int64, error) {
	switch whence {
	case io.SeekStart:
		f.pos = offset
	case io.SeekCurrent:
		f.pos += offset
	case io.SeekEnd:
		f.pos = int64(len(f.buf)) + offset
	default:
		return 0, os.ErrInvalid
	}
	if f.pos < 0 {
		return 0, os.ErrInvalid
	}
	return f.pos, nil
}

// Stat implements afero.File.
func (f *file) Stat() (os.FileInfo, error) {
	return f.stat, nil
}

// Sync implements afero.File.
func (f *file) Sync() error {
	return nil
}

// Truncate implements afero.File.
func (f *file) Truncate(size int64) error {
	if !f.writeAllowed() {
		return os.ErrInvalid
	}
	f.modified = true

	if size > int64(len(f.buf)) {
		pad := make([]byte, int(size)-len(f.buf))
		f.buf = append(f.buf, pad...)
	} else if size < int64(len(f.buf)) {
		f.buf = f.buf[:size]
	}
	if f.pos > size {
		f.pos = size
	}
	return nil
}

// Write implements afero.File.
func (f *file) Write(p []byte) (n int, err error) {
	n, err = f.WriteAt(p, f.pos)
	f.pos += int64(n)
	return
}

// WriteAt implements afero.File.
func (f *file) WriteAt(p []byte, off int64) (n int, err error) {
	if !f.writeAllowed() {
		return 0, os.ErrInvalid
	}
	f.modified = true

	if int(off) > len(f.buf) {
		pad := make([]byte, int(off)-len(f.buf))
		f.buf = append(f.buf, pad...)
	}

	n = copy(f.buf[off:], p)
	f.buf = append(f.buf, p[n:]...)
	return len(p), nil
}

// WriteString implements afero.File.
func (f *file) WriteString(s string) (ret int, err error) {
	return f.Write([]byte(s))
}
