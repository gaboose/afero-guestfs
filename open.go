package aferoguestfs

import (
	"fmt"

	"github.com/gaboose/afero-guestfs/libguestfs.org/guestfs"
)

// PartitionFs is a utility type for opening a partition in a disk image.
type PartitionFs struct {
	*Fs
	inner *guestfs.Guestfs
}

// OpenPartitionFs opens a new partition.
// It takes a path to an image file and a partition device.
func OpenPartitionFs(image string, partition string) (*PartitionFs, error) {
	g, err := guestfs.Create()
	if err != nil {
		return nil, fmt.Errorf("create failed: %w", err)
	}

	if err := g.Add_drive(image, nil); err != nil {
		g.Close()
		return nil, fmt.Errorf("add drive failed: %w", err)
	}

	if err := g.Launch(); err != nil {
		g.Close()
		return nil, fmt.Errorf("launch failed: %w", err)
	}

	if err := g.Mount(partition, "/"); err != nil {
		g.Close()
		return nil, fmt.Errorf("failed to mount partition %s: %w", partition, err)
	}

	return &PartitionFs{Fs: New(g)}, nil
}

func (p *PartitionFs) Close() error {
	if err := p.inner.Umount_all(); err != nil {
		return fmt.Errorf("umount all failed: %w", err)
	}
	if err := p.inner.Close(); err != nil {
		return fmt.Errorf("guestfs close failed: %w", err)
	}
	return nil
}
