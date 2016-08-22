package vg

import (
	"fmt"
	"os"
)

func resolveSymlink(dev string) (string, error) {
	for i := 0; i < 10; i++ {
		fi, err := os.Lstat(dev)
		if err != nil {
			return "", err
		}

		if fi.Mode() != os.ModeSymlink {
			return dev, nil
		}

		dev, err = os.Readlink(dev)
		if err != nil {
			return "", err
		}
	}
	return "", fmt.Errorf("too many symlinks for %s", dev)
}
