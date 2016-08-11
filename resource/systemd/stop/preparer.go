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

package stop

import (
	"errors"
	"fmt"
	"time"

	"github.com/asteris-llc/converge/resource"
	"github.com/asteris-llc/converge/resource/systemd/common"
)

// Preparer for Content
type Preparer struct {
	Unit    string `hcl:"unit"`
	Mode    string `hcl:"mode"`
	Timeout string `hcl:"timeout"`
}

// Prepare a new task
func (p *Preparer) Prepare(render resource.Renderer) (resource.Task, error) {
	unit, err := render.Render("unit", p.Unit)
	if err != nil {
		return nil, err
	}

	timeout, err := render.Render("timeout", p.Timeout)
	if err != nil {
		return nil, err
	}
	var t time.Duration
	if timeout == "" {
		t = common.DefaultTimeout
	} else {
		t, err = time.ParseDuration(timeout)
		if err != nil {
			return nil, err
		}
	}

	mode, err := render.Render("mode", p.Mode)
	if err != nil {
		return nil, err
	}
	var m common.StartMode
	if mode == "" {
		m = common.SMReplace
	} else {
		m = common.StartMode(mode)
	}

	stopModule := &Stop{Unit: unit, Mode: m, Timeout: t}
	return stopModule, ValidateTask(stopModule)
}

func ValidateTask(stopModule *Stop) error {
	if stopModule.Unit == "" {
		return errors.New("resource requires a `unit` parameter")
	}
	switch stopModule.Mode {
	case common.SMReplace:
		fallthrough
	case common.SMFail:
		fallthrough
	case common.SMIsolate:
		fallthrough
	case common.SMIgnoreDependencies:
		fallthrough
	case common.SMIgnoreRequirements:
		break
	default:
		return fmt.Errorf("resource's parameter 'mode' is not one of %s, found %s", common.ValidStartModes, stopModule.Mode)
	}
	return nil
}
