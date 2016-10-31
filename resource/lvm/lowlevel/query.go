// Copyright © 2016 Asteris, LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package lowlevel

import (
	"fmt"
	"github.com/mitchellh/mapstructure"
	"strings"
)

func (lvm *realLVM) Blkid(dev string) (string, error) {
	if ok, err := lvm.backend.Exists(dev); err != nil || !ok {
		return "", err
	}

	blkid, rc, err := lvm.backend.ReadWithExitCode("blkid", []string{"-c", "/dev/null", "-o", "value", "-s", "TYPE", dev})
	if err != nil {
		return "", err
	}
	// excerpt from `man blkid`:
	// RETURN CODE
	// If  the  specified  device  or  device addressed by specified token (option -t) was found and it's possible to gather any information about the device, an exit code 0 is returned.
	// Note the option -s filters output tags, but it does not affect return code.
	// If the specified token was not found, or no (specified) devices could be identified, an exit code of 2 is returned.
	// For usage or other errors, an exit code of 4 is returned.
	//  If an ambivalent low-level probing result was detected, an exit code of 8 is returned.
	if rc != 2 && rc != 0 {
		return blkid, fmt.Errorf("blkid terminated with rc == %d, and output `%s`", rc, blkid)
	}
	return blkid, nil
}

func (lvm *realLVM) QueryDeviceMapperName(dmName string) (string, error) {
	out, err := lvm.backend.Read("dmsetup", []string{"info", "-C", "--noheadings", "-o", "name", dmName})
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("/dev/mapper/%s", out), nil
}

func (lvm *realLVM) Query(prog string, out string, extras []string) ([]map[string]interface{}, error) {
	result := []map[string]interface{}{}
	args := []string{"--nameprefix", "--noheadings", "--unquoted", "--units", "b", "-o", out, "--separator", ";"}
	args = append(args, extras...)
	output, err := lvm.backend.Read(prog, args)
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

func (lvm *realLVM) parse(values interface{}, dest interface{}) error {
	return mapstructure.Decode(values, dest)
}
