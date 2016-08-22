package lowlevel

import (
	"fmt"
	"github.com/mitchellh/mapstructure"
	"os/exec"
	"strings"
)

func queryBlkid(dev string) (string, error) {
	out, err := exec.Command("blkid", "-c", "/dev/null", "-o", "value", "-s", "TYPE", dev).Output()
	if err != nil {
		return "", err
	}
	return string(out), err
}

func queryDeviceMapperName(dmName string) (string, error) {
	out, err := exec.Command("dmsetup", "info", "-C", "--noheadings", "-o", "name", dmName).Output()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("/dev/mapper/%s", string(out)), nil
}

func queryLVM(prog string, out string, extras []string) ([]map[string]interface{}, error) {
	result := []map[string]interface{}{}
	args := []string{"--nameprefix", "--noheadings", "--unquoted", "-o", out, "--separator", ";"}
	args = append(args, extras...)
	cmd := exec.Command(prog, args...)
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	for _, line := range strings.Split(string(output), "\n") {
		values := map[string]interface{}{}
		for _, field := range strings.Split(line, ";") {
			parts := strings.Split(field, "=")
			values[parts[0]] = parts[1]
		}
		result = append(result, values)
	}
	return result, nil
}

func parseLVM(values interface{}, dest interface{}) error {
	return mapstructure.Decode(values, dest)
}
