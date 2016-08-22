package lowlevel

import (
	"os/exec"
)

func VGCreate(vg string, devs []string) error {
	args := []string{vg}
	args = append(args, devs...)
	return exec.Command("vgcreate", args...).Wait()
}

func VGExtend(vg string, dev string) error {
	return exec.Command("vgextend", vg, dev).Wait()
}

func VGReduce(vg string, dev string) error {
	return exec.Command("vgreduce", vg, dev).Wait()
}

func PVCreate(dev string) error {
	return exec.Command("pvcreate", dev).Wait()
}
