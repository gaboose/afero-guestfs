package aferoguestfs_test

import (
	"bytes"
	"io"
	"io/fs"
	"os"
	"testing"

	aferoguestfs "github.com/gaboose/afero-guestfs"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOpenPartition(t *testing.T) {
	f, err := os.CreateTemp("", "afero-guestfs-test-*.img")
	require.Nil(t, err)
	defer os.Remove(f.Name())
	defer f.Close()

	_, err = io.Copy(f, bytes.NewBuffer(test1Img))
	require.Nil(t, err)
	require.Nil(t, f.Close())

	fsys, err := aferoguestfs.OpenPartitionFs(f.Name(), "/dev/sda2")
	assert.Nil(t, err)
	defer fsys.Close()

	require.Nil(t, afero.WriteFile(fsys, "test.txt", []byte("some text"), fs.ModePerm))
	body, err := afero.ReadFile(fsys, "test.txt")
	require.Nil(t, err)

	assert.Equal(t, string(body), "some text")
}
