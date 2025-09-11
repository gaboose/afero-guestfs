# Libguestfs Backend for Afero

# How to use

```go
package main

import (
	"fmt"
	"io/fs"

	aferoguestfs "github.com/gaboose/afero-guestfs"
	"github.com/gaboose/afero-guestfs/libguestfs.org/guestfs"
	"github.com/spf13/afero"
)

func main() {
	g, err := guestfs.Create()
	if err != nil {
		panic(err)
	}
	defer g.Close()

	if err := g.Add_drive("diskimage.img", nil); err != nil {
		panic(err)
	}

	if err := g.Launch(); err != nil {
		panic(err)
	}

	if err := g.Mount("/dev/sda2", "/"); err != nil {
		panic(err)
	}

	gfs := aferoguestfs.New(g)

	afero.Walk(gfs, "/home", func(path string, info fs.FileInfo, err error) error {
		fmt.Println(path)
		return nil
	})
}
```
