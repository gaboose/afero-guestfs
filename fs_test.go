package aferoguestfs_test

import (
	"bytes"
	_ "embed"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"testing"
	"time"

	aferoguestfs "github.com/gaboose/afero-guestfs"
	"github.com/gaboose/afero-guestfs/libguestfs.org/guestfs"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

//go:embed testdata/test1.img
var test1Img []byte

var gfs *aferoguestfs.Fs

func newGuestFs(diskPath, partDev string) (g *guestfs.Guestfs, close func() error, err error) {
	g, err = guestfs.Create()
	if err != nil {
		return nil, nil, fmt.Errorf("create failed: %w", err)
	}
	defer func() {
		if err != nil {
			g.Close()
		}
	}()

	if err := g.Add_drive(diskPath, nil); err != nil {
		return nil, nil, fmt.Errorf("add drive failed: %w", err)
	}

	if err := g.Launch(); err != nil {
		return nil, nil, fmt.Errorf("launch failed: %w", err)
	}

	if err := g.Mount(partDev, "/"); err != nil {
		return nil, nil, fmt.Errorf("failed to mount partition %s: %w", partDev, err)
	}

	return g, func() error {
		if err := g.Umount_all(); err != nil {
			return fmt.Errorf("umount all failed: %w", err)
		}
		if err := g.Close(); err != nil {
			return fmt.Errorf("guestfs close failed: %w", err)
		}
		return nil
	}, nil
}

func setup() (*aferoguestfs.Fs, func(), error) {
	f, err := os.CreateTemp("", "afero-guestfs-test*.img")
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create temp file: %w", err)
	}

	_, err = io.Copy(f, bytes.NewBuffer(test1Img))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to write to temp file: %w", err)
	}

	err = f.Close()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to close temp file: %w", err)
	}

	gfs, gfsClose, err := newGuestFs(f.Name(), "/dev/sda2")
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create guestfs: %w", err)
	}
	return aferoguestfs.New(gfs), func() {
		if err := gfsClose(); err != nil {
			fmt.Fprintf(os.Stderr, "warning: failed to close guestfs: %+v", err)
		}
		if err := os.Remove(f.Name()); err != nil {
			fmt.Fprintf(os.Stderr, "warning: failed to remove %s: %+v", f.Name(), err)
		}
	}, nil
}

func TestMain(m *testing.M) {
	fs, gfsClose, err := setup()
	if err != nil {
		panic(err)
	}

	gfs = fs

	code := m.Run()

	gfsClose()

	os.Exit(code)
}

func TestChmod(t *testing.T) {
	clear(t, gfs)

	err := afero.WriteFile(gfs, "test.txt", []byte("some text"), 0644)
	assert.Nil(t, err)

	stat, err := gfs.Stat("test.txt")
	assert.Nil(t, err)

	assert.Equal(t, stat.Mode(), os.FileMode(0644))

	err = gfs.Chmod("test.txt", os.FileMode(0777))
	assert.Nil(t, err)

	stat, err = gfs.Stat("test.txt")
	assert.Nil(t, err)

	assert.Equal(t, stat.Mode(), os.FileMode(0777))
}

func TestChown(t *testing.T) {
	clear(t, gfs)

	err := afero.WriteFile(gfs, "test.txt", []byte("some text"), os.ModePerm)
	assert.Nil(t, err)

	stat, err := gfs.Stat("test.txt")
	assert.Nil(t, err)
	statns := stat.Sys().(*guestfs.StatNS)

	assert.Equal(t, statns.St_uid, int64(0))
	assert.Equal(t, statns.St_gid, int64(0))

	err = gfs.Chown("test.txt", 1000, 1000)
	assert.Nil(t, err)

	stat, err = gfs.Stat("test.txt")
	assert.Nil(t, err)
	statns = stat.Sys().(*guestfs.StatNS)

	assert.Equal(t, statns.St_uid, int64(1000))
	assert.Equal(t, statns.St_gid, int64(1000))
}

func TestChtimes(t *testing.T) {
	clear(t, gfs)

	err := afero.WriteFile(gfs, "test.txt", []byte("some text"), os.ModePerm)
	assert.Nil(t, err)

	expected := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)

	err = gfs.Chtimes("test.txt", expected, expected)
	assert.Nil(t, err)

	stat, err := gfs.Stat("test.txt")
	assert.Nil(t, err)

	assert.Equal(t, expected, stat.ModTime().UTC())
}

func TestCreate(t *testing.T) {
	clear(t, gfs)

	f, err := gfs.Create("test.txt")
	assert.Nil(t, err)
	f.Close()

	exists, err := afero.Exists(gfs, "test.txt")
	assert.True(t, exists)
	assert.Nil(t, err)
}

func TestMkdir(t *testing.T) {
	clear(t, gfs)

	err := gfs.Mkdir("etc", os.ModePerm)
	assert.Nil(t, err)

	exists, err := afero.DirExists(gfs, "etc")
	assert.True(t, exists)
	assert.Nil(t, err)
}

func TestMkdirAll(t *testing.T) {
	clear(t, gfs)

	err := gfs.MkdirAll("home/user", os.ModePerm)
	assert.Nil(t, err)

	exists, err := afero.DirExists(gfs, "home/user")
	assert.True(t, exists)
	assert.Nil(t, err)
}

func TestRemove(t *testing.T) {
	clear(t, gfs)

	err := afero.WriteFile(gfs, "test1.txt", []byte("some text"), os.ModePerm)
	assert.Nil(t, err)

	err = gfs.Remove("test1.txt")
	assert.Nil(t, err)

	exists, err := afero.Exists(gfs, "test1.txt")
	assert.False(t, exists)
	assert.Nil(t, err)
}

func TestRemoveAll(t *testing.T) {
	clear(t, gfs)

	err := gfs.Mkdir("etc", os.ModePerm)
	assert.Nil(t, err)

	err = afero.WriteFile(gfs, "etc/test1.txt", []byte("some text"), os.ModePerm)
	assert.Nil(t, err)

	err = afero.WriteFile(gfs, "etc/test2.txt", []byte("some more text"), os.ModePerm)
	assert.Nil(t, err)

	err = gfs.RemoveAll("etc")
	assert.Nil(t, err)

	exists, err := afero.DirExists(gfs, "etc")
	assert.False(t, exists)
	assert.Nil(t, err)
}

func TestRename(t *testing.T) {
	clear(t, gfs)

	err := afero.WriteFile(gfs, "test1.txt", []byte("some text"), os.ModePerm)
	assert.Nil(t, err)

	err = gfs.Rename("test1.txt", "test2.txt")
	assert.Nil(t, err)

	exists, err := afero.Exists(gfs, "test1.txt")
	assert.False(t, exists)
	assert.Nil(t, err)

	bts, err := afero.ReadFile(gfs, "test2.txt")
	assert.Nil(t, err)
	assert.Equal(t, []byte("some text"), bts)
}

func TestOpen(t *testing.T) {
	clear(t, gfs)

	err := afero.WriteFile(gfs, "test1.txt", []byte("some text"), os.ModePerm)
	assert.Nil(t, err)

	f, err := gfs.Open("test1.txt")
	assert.Nil(t, err)
	defer f.Close()

	bts, err := afero.ReadAll(f)
	assert.Nil(t, err)
	assert.Equal(t, []byte("some text"), bts)
}

func TestOpenFile(t *testing.T) {
	t.Run("CanRead", func(t *testing.T) {
		clear(t, gfs)

		err := afero.WriteFile(gfs, "test1.txt", []byte("some text"), os.ModePerm)
		assert.Nil(t, err)

		f, err := gfs.OpenFile("test1.txt", 0, os.ModePerm)
		assert.Nil(t, err)
		defer f.Close()

		bts, err := afero.ReadAll(f)
		assert.Nil(t, err)
		assert.Equal(t, []byte("some text"), bts)
	})

	t.Run("SetsPerm", func(t *testing.T) {
		clear(t, gfs)

		f, err := gfs.OpenFile("test1.txt", os.O_CREATE|os.O_TRUNC, os.ModePerm)
		assert.Nil(t, err)
		f.Close()

		fi, err := gfs.Lstat("test1.txt")
		assert.Nil(t, err)

		assert.Equal(t, fi.Mode().Perm(), fs.ModePerm)
	})
}

func TestLstatIfPossible(t *testing.T) {
	clear(t, gfs)

	err := afero.WriteFile(gfs, "test1.txt", []byte("some text"), os.ModePerm)
	assert.Nil(t, err)

	err = gfs.SymlinkIfPossible("test1.txt", "test2.txt")
	assert.Nil(t, err)

	fi, ok, err := gfs.LstatIfPossible("test2.txt")
	assert.Nil(t, err)
	assert.True(t, ok)

	assert.True(t, fi.Mode()&os.ModeSymlink != 0)
}

func TestReadlinkIfPossible(t *testing.T) {
	clear(t, gfs)

	err := afero.WriteFile(gfs, "test1.txt", []byte("some text"), os.ModePerm)
	assert.Nil(t, err)

	err = gfs.SymlinkIfPossible("test1.txt", "test2.txt")
	assert.Nil(t, err)

	target, err := gfs.ReadlinkIfPossible("test2.txt")
	assert.Nil(t, err)
	assert.Equal(t, "test1.txt", target)
}

func TestSymlinkIfPossible(t *testing.T) {
	clear(t, gfs)

	err := afero.WriteFile(gfs, "test1.txt", []byte("some text"), os.ModePerm)
	assert.Nil(t, err)

	err = gfs.SymlinkIfPossible("test1.txt", "test2.txt")
	assert.Nil(t, err)

	err = gfs.SymlinkIfPossible("/test1.txt", "test3.txt")
	assert.Nil(t, err)

	target, err := gfs.Readlink("test2.txt")
	assert.Nil(t, err)
	assert.Equal(t, "test1.txt", target)

	target, err = gfs.Readlink("test3.txt")
	assert.Nil(t, err)
	assert.Equal(t, "/test1.txt", target)
}

func TestLink(t *testing.T) {
	clear(t, gfs)

	err := afero.WriteFile(gfs, "test1.txt", []byte("some text"), os.ModePerm)
	assert.Nil(t, err)

	err = gfs.Link("test1.txt", "test2.txt")
	assert.Nil(t, err)

	fi1, err := gfs.Stat("test1.txt")
	assert.Nil(t, err)

	fi2, err := gfs.Stat("test2.txt")
	assert.Nil(t, err)

	assert.Equal(t, fi1.Sys().(*guestfs.StatNS).St_ino, fi2.Sys().(*guestfs.StatNS).St_ino)
}

func clear(t *testing.T, gfs *aferoguestfs.Fs) {
	root, err := gfs.Open("/")
	assert.Nil(t, err)

	dirnames, err := root.Readdirnames(-1)
	assert.Nil(t, err)

	for _, dirname := range dirnames {
		err = gfs.RemoveAll(filepath.Join("/", dirname))
		assert.Nil(t, err)
	}
}
