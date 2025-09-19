package aferoguestfs_test

import (
	"io"
	"os"
	"sort"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

func TestRead(t *testing.T) {
	defer clear(t, gfs)

	err := afero.WriteFile(gfs, "/test1.txt", []byte("some text"), os.ModePerm)
	assert.Nil(t, err)

	f, err := gfs.Open("/test1.txt")
	assert.Nil(t, err)
	defer f.Close()

	var bts [4]byte
	n, err := f.Read(bts[:])
	assert.Nil(t, err)
	assert.Equal(t, 4, n)

	assert.Equal(t, []byte("some"), bts[:])
}

func TestReadAt(t *testing.T) {
	defer clear(t, gfs)

	err := afero.WriteFile(gfs, "/test1.txt", []byte("some text"), os.ModePerm)
	assert.Nil(t, err)

	f, err := gfs.Open("/test1.txt")
	assert.Nil(t, err)
	defer f.Close()

	var bts [4]byte
	n, err := f.ReadAt(bts[:], 5)
	assert.Nil(t, err)
	assert.Equal(t, 4, n)

	assert.Equal(t, []byte("text"), bts[:])
}

func TestReaddirnames(t *testing.T) {
	defer clear(t, gfs)

	err := gfs.Mkdir("/etc", os.ModePerm)
	assert.Nil(t, err)

	err = afero.WriteFile(gfs, "/etc/test1.txt", []byte("some text"), os.ModePerm)
	assert.Nil(t, err)

	err = afero.WriteFile(gfs, "/etc/test2.txt", []byte("some more text"), os.ModePerm)
	assert.Nil(t, err)

	f, err := gfs.Open("/etc")
	assert.Nil(t, err)
	defer f.Close()

	dirnames, err := f.Readdirnames(-1)
	assert.Nil(t, err)

	assert.Equal(t, []string{"test1.txt", "test2.txt"}, dirnames)
}

func TestReaddir(t *testing.T) {
	defer clear(t, gfs)

	err := gfs.Mkdir("/etc", os.ModePerm)
	assert.Nil(t, err)

	err = afero.WriteFile(gfs, "/etc/test1.txt", []byte("some text"), os.ModePerm)
	assert.Nil(t, err)

	err = afero.WriteFile(gfs, "/etc/test2.txt", []byte("some more text"), os.ModePerm)
	assert.Nil(t, err)

	f, err := gfs.Open("/etc")
	assert.Nil(t, err)
	defer f.Close()

	fileInfos, err := f.Readdir(-1)
	assert.Nil(t, err)
	assert.Equal(t, len(fileInfos), 2)

	sort.Slice(fileInfos, func(i int, j int) bool {
		return fileInfos[i].Name() < fileInfos[j].Name()
	})

	assert.Equal(t, "test1.txt", fileInfos[0].Name())
	assert.Equal(t, false, fileInfos[0].IsDir())
	assert.Equal(t, os.FileMode(0777), fileInfos[0].Mode())
	assert.Equal(t, int64(9), fileInfos[0].Size())

	assert.Equal(t, "test2.txt", fileInfos[1].Name())
	assert.Equal(t, false, fileInfos[1].IsDir())
	assert.Equal(t, os.FileMode(0777), fileInfos[1].Mode())
	assert.Equal(t, int64(14), fileInfos[1].Size())
}

func TestSeek(t *testing.T) {
	defer clear(t, gfs)

	err := afero.WriteFile(gfs, "/test1.txt", []byte("some more text"), os.ModePerm)
	assert.Nil(t, err)

	f, err := gfs.Open("/test1.txt")
	assert.Nil(t, err)
	defer f.Close()

	nseek, err := f.Seek(5, io.SeekStart)
	assert.Nil(t, err)
	assert.Equal(t, int64(5), nseek)

	var bts [4]byte
	n, err := f.Read(bts[:])
	assert.Nil(t, err)
	assert.Equal(t, 4, n)

	assert.Equal(t, []byte("more"), bts[:])
}

func TestStat(t *testing.T) {
	defer clear(t, gfs)

	err := afero.WriteFile(gfs, "/test1.txt", []byte("some more text"), os.ModePerm)
	assert.Nil(t, err)

	f, err := gfs.Open("/test1.txt")
	assert.Nil(t, err)
	defer f.Close()

	stat, err := f.Stat()
	assert.Nil(t, err)

	assert.Equal(t, false, stat.IsDir())
	assert.Equal(t, os.FileMode(0777), stat.Mode())
	assert.Equal(t, "test1.txt", stat.Name())
	assert.Equal(t, int64(14), stat.Size())
}

func TestWrite(t *testing.T) {
	defer clear(t, gfs)

	f, err := gfs.Create("/test1.txt")
	assert.Nil(t, err)

	n, err := f.Write([]byte("some text"))
	assert.Nil(t, err)
	assert.Equal(t, 9, n)

	err = f.Close()
	assert.Nil(t, err)

	bts, err := afero.ReadFile(gfs, "/test1.txt")
	assert.Nil(t, err)

	assert.Equal(t, []byte("some text"), bts)
}

func TestWriteAt(t *testing.T) {
	defer clear(t, gfs)

	f, err := gfs.Create("/test1.txt")
	assert.Nil(t, err)

	n, err := f.WriteAt([]byte("some text"), 4)
	assert.Nil(t, err)
	assert.Equal(t, 9, n)

	err = f.Close()
	assert.Nil(t, err)

	bts, err := afero.ReadFile(gfs, "/test1.txt")
	assert.Nil(t, err)

	assert.Equal(t, append([]byte{0, 0, 0, 0}, []byte("some text")...), bts)
}

func TestWriteString(t *testing.T) {
	defer clear(t, gfs)

	f, err := gfs.Create("/test1.txt")
	assert.Nil(t, err)

	n, err := f.WriteString("some text")
	assert.Nil(t, err)
	assert.Equal(t, 9, n)

	err = f.Close()
	assert.Nil(t, err)

	bts, err := afero.ReadFile(gfs, "/test1.txt")
	assert.Nil(t, err)

	assert.Equal(t, []byte("some text"), bts)
}

func TestTruncate(t *testing.T) {
	defer clear(t, gfs)

	f, err := gfs.Create("/test1.txt")
	assert.Nil(t, err)

	err = f.Truncate(4)
	assert.Nil(t, err)

	err = f.Close()
	assert.Nil(t, err)

	bts, err := afero.ReadFile(gfs, "/test1.txt")
	assert.Nil(t, err)

	assert.Equal(t, []byte{0, 0, 0, 0}, bts)
}
