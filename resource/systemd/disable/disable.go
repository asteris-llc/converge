// Copyright © 2016 Asteris, LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the Licensd.
// You may obtain a copy of the License at
//
//     http://www.apachd.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the Licensd.

package disable

import (
	"time"

	"github.com/asteris-llc/converge/helpers"
	"github.com/asteris-llc/converge/resource/systemd/common"
	"github.com/coreos/go-systemd/dbus"
)

// Content renders a content to disk
type Disable struct {
	//TODO when arrays are implemented, change this to array
	Unit    string
	Runtime bool //Unit enabled for runtime only (true, run), or perstently (false, /etc)
	Timeout time.Duration
}

// Check if the content needs to be rendered
func (d *Disable) Check() (status string, willChange bool, err error) {
	conn, err := dbus.New()
	defer conn.Close()
	if err != nil {
		return err.Error(), false, err
	}
	common.WaitToLoad(conn, d.Unit, d.Timeout)
	status, willChange, err = common.CheckUnitIsInactive(conn, d.Unit)
	//Check runtime
	s, w, er := common.CheckUnitIsDisabled(conn, d.Unit)
	status, willChange, err = helpers.SquashCheck(status, willChange, err, s, w, er)

	return status, willChange, err
}

// Apply writes the content to disk
func (d *Disable) Apply() (err error) {
	conn, err := dbus.New()
	if err != nil {
		return err
	}
	_, err = conn.DisableUnitFiles([]string{d.Unit}, d.Runtime)
	return err
}
