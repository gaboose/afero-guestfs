# Libguestfs Backend for Afero

# How to use

```go
package main

import (
	"fmt"
	"io/fs"

	aferoguestfs "github.com/gaboose/afero-guestfs"
	"github.com/spf13/afero"
)

func main() {
	gfs, err := aferoguestfs.OpenPartitionFs("disk.img", "/dev/sda2")
	if err != nil {
		panic(err)
	}
	defer gfs.Close()

	afero.Walk(gfs, "/home", func(path string, info fs.FileInfo, err error) error {
		fmt.Println(path)
		return nil
	})
}
```
