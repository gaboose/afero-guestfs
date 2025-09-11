package aferoguestfs_test

import (
	"bytes"
	_ "embed"
	"fmt"
	"io"
	"os"
	"testing"
	"time"

	aferoguestfs "github.com/gaboose/afero-guestfs"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	guestfs "libguestfs.org/guestfs"
)

//go:embed testdata/test1.img
var test1Img []byte

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

func setup(t *testing.T) (*aferoguestfs.Fs, func()) {
	f, err := os.CreateTemp("", "afero-guestfs-test*.img")
	assert.Nil(t, err)

	_, err = io.Copy(f, bytes.NewBuffer(test1Img))
	assert.Nil(t, err)

	err = f.Close()
	assert.Nil(t, err)

	gfs, gfsClose, err := newGuestFs(f.Name(), "/dev/sda2")
	assert.Nil(t, err)
	return aferoguestfs.New(gfs), func() {
		if err := gfsClose(); err != nil {
			fmt.Fprintf(os.Stderr, "warning: failed to close guestfs: %+v", err)
		}
		if err := os.Remove(f.Name()); err != nil {
			fmt.Fprintf(os.Stderr, "warning: failed to remove %s: %+v", f.Name(), err)
		}
	}
}

func TestChmod(t *testing.T) {
	gfs, setupClose := setup(t)
	defer setupClose()

	err := afero.WriteFile(gfs, "/test.txt", []byte("some text"), os.ModePerm)
	assert.Nil(t, err)

	stat, err := gfs.Stat("/test.txt")
	assert.Nil(t, err)

	assert.Equal(t, stat.Mode(), os.FileMode(0644))

	err = gfs.Chmod("/test.txt", os.FileMode(0777))
	assert.Nil(t, err)

	stat, err = gfs.Stat("/test.txt")
	assert.Nil(t, err)

	assert.Equal(t, stat.Mode(), os.FileMode(0777))
}

func TestChown(t *testing.T) {
	gfs, setupClose := setup(t)
	defer setupClose()

	err := afero.WriteFile(gfs, "/test.txt", []byte("some text"), os.ModePerm)
	assert.Nil(t, err)

	stat, err := gfs.Stat("/test.txt")
	assert.Nil(t, err)
	statns := stat.Sys().(*guestfs.StatNS)

	assert.Equal(t, statns.St_uid, int64(0))
	assert.Equal(t, statns.St_gid, int64(0))

	err = gfs.Chown("/test.txt", 1000, 1000)
	assert.Nil(t, err)

	stat, err = gfs.Stat("/test.txt")
	assert.Nil(t, err)
	statns = stat.Sys().(*guestfs.StatNS)

	assert.Equal(t, statns.St_uid, int64(1000))
	assert.Equal(t, statns.St_gid, int64(1000))
}

func TestChtimes(t *testing.T) {
	gfs, setupClose := setup(t)
	defer setupClose()

	err := afero.WriteFile(gfs, "/test.txt", []byte("some text"), os.ModePerm)
	assert.Nil(t, err)

	expected := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)

	err = gfs.Chtimes("/test.txt", expected, expected)
	assert.Nil(t, err)

	stat, err := gfs.Stat("/test.txt")
	assert.Nil(t, err)

	assert.Equal(t, stat.ModTime(), expected)
}

func TestCreate(t *testing.T) {
	gfs, setupClose := setup(t)
	defer setupClose()

	f, err := gfs.Create("/test.txt")
	assert.Nil(t, err)
	f.Close()

	exists, err := afero.Exists(gfs, "/test.txt")
	assert.True(t, exists)
	assert.Nil(t, err)
}

func TestMkdir(t *testing.T) {
	gfs, setupClose := setup(t)
	defer setupClose()

	err := gfs.Mkdir("/etc", os.ModePerm)
	assert.Nil(t, err)

	exists, err := afero.DirExists(gfs, "/etc")
	assert.True(t, exists)
	assert.Nil(t, err)
}

func TestMkdirAll(t *testing.T) {
	gfs, setupClose := setup(t)
	defer setupClose()

	err := gfs.MkdirAll("/home/user", os.ModePerm)
	assert.Nil(t, err)

	exists, err := afero.DirExists(gfs, "/home/user")
	assert.True(t, exists)
	assert.Nil(t, err)
}

func TestRemove(t *testing.T) {
	gfs, setupClose := setup(t)
	defer setupClose()

	err := afero.WriteFile(gfs, "/test1.txt", []byte("some text"), os.ModePerm)
	assert.Nil(t, err)

	err = gfs.Remove("/test1.txt")
	assert.Nil(t, err)

	exists, err := afero.Exists(gfs, "/test1.txt")
	assert.False(t, exists)
	assert.Nil(t, err)
}

func TestRemoveAll(t *testing.T) {
	gfs, setupClose := setup(t)
	defer setupClose()

	err := gfs.Mkdir("/etc", os.ModePerm)
	assert.Nil(t, err)

	err = afero.WriteFile(gfs, "/etc/test1.txt", []byte("some text"), os.ModePerm)
	assert.Nil(t, err)

	err = afero.WriteFile(gfs, "/etc/test2.txt", []byte("some more text"), os.ModePerm)
	assert.Nil(t, err)

	err = gfs.RemoveAll("/etc")
	assert.Nil(t, err)

	exists, err := afero.DirExists(gfs, "/etc")
	assert.False(t, exists)
	assert.Nil(t, err)
}

func TestRename(t *testing.T) {
	gfs, setupClose := setup(t)
	defer setupClose()

	err := afero.WriteFile(gfs, "/test1.txt", []byte("some text"), os.ModePerm)
	assert.Nil(t, err)

	err = gfs.Rename("/test1.txt", "/test2.txt")
	assert.Nil(t, err)

	exists, err := afero.Exists(gfs, "/test1.txt")
	assert.False(t, exists)
	assert.Nil(t, err)

	bts, err := afero.ReadFile(gfs, "/test2.txt")
	assert.Nil(t, err)
	assert.Equal(t, []byte("some text"), bts)
}

func TestOpen(t *testing.T) {
	gfs, setupClose := setup(t)
	defer setupClose()

	err := afero.WriteFile(gfs, "/test1.txt", []byte("some text"), os.ModePerm)
	assert.Nil(t, err)

	f, err := gfs.Open("/test1.txt")
	assert.Nil(t, err)
	defer f.Close()

	bts, err := afero.ReadAll(f)
	assert.Nil(t, err)
	assert.Equal(t, []byte("some text"), bts)
}

func TestOpenFile(t *testing.T) {
	gfs, setupClose := setup(t)
	defer setupClose()

	err := afero.WriteFile(gfs, "/test1.txt", []byte("some text"), os.ModePerm)
	assert.Nil(t, err)

	f, err := gfs.OpenFile("/test1.txt", 0, os.ModePerm)
	assert.Nil(t, err)
	defer f.Close()

	bts, err := afero.ReadAll(f)
	assert.Nil(t, err)
	assert.Equal(t, []byte("some text"), bts)
}
