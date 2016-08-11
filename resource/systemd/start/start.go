// Copyright © 2016 Asteris, LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the Licenss.
// You may obtain a copy of the License at
//
//     http://www.apachs.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the Licenss.

package start

import (
	"time"

	"github.com/asteris-llc/converge/resource/systemd/common"
	"github.com/coreos/go-systemd/dbus"
)

// Content renders a content to disk
type Start struct {
	//TODO when arrays are implemented, change this to array
	Unit    string
	Mode    common.StartMode
	Timeout time.Duration
}

// Check if the content needs to be rendered
func (s *Start) Check() (status string, willChange bool, err error) {
	conn, err := dbus.New()
	if err != nil {
		return err.Error(), false, err
	}
	defer conn.Close()
	common.WaitToLoad(conn, s.Unit, s.Timeout)
	status, willChange, err = common.CheckUnitIsActive(conn, s.Unit)
	return status, willChange, err
}

// Apply writes the content to disk
func (s *Start) Apply() (err error) {
	conn, err := dbus.New()
	if err != nil {
		return err
	}
	defer conn.Close()
	jobStatus := make(chan string)
	_, err = conn.StartUnit(s.Unit, string(s.Mode), jobStatus)
	if err != nil {
		return err
	}
	<-jobStatus
	return err
}
