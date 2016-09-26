package lowlevel

import (
	"fmt"
	"github.com/mitchellh/mapstructure"
	"strings"
)

func (lvm *RealLVM) Blkid(dev string) (string, error) {
	return lvm.Backend.Read("blkid", []string{"-c", "/dev/null", "-o", "value", "-s", "TYPE", dev})
}

func (lvm *RealLVM) QueryDeviceMapperName(dmName string) (string, error) {
	out, err := lvm.Backend.Read("dmsetup", []string{"info", "-C", "--noheadings", "-o", "name", dmName})
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("/dev/mapper/%s", out), nil
}

func (lvm *RealLVM) Query(prog string, out string, extras []string) ([]map[string]interface{}, error) {
	result := []map[string]interface{}{}
	args := []string{"--nameprefix", "--noheadings", "--unquoted", "--units", "b", "-o", out, "--separator", ";"}
	args = append(args, extras...)
	output, err := lvm.Backend.Read(prog, args)
	if err != nil {
		return nil, err
	}
	for _, line := range strings.Split(output, "\n") {
		values := map[string]interface{}{}
		for _, field := range strings.Split(line, ";") {
			parts := strings.Split(field, "=")
			if len(parts) == 1 {
				continue
			}
			values[parts[0]] = parts[1]
		}
		if len(values) > 0 {
			result = append(result, values)
		}
	}
	return result, nil
}

func (lvm *RealLVM) parse(values interface{}, dest interface{}) error {
	return mapstructure.Decode(values, dest)
}
