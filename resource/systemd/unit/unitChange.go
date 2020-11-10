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

package unit

// unitFileChange mirrors github.com/coreos/go-systemd/dbus.EnableUnitFileChange
// and github.com/coreos/go-systemd/dbus.DisableUnitFileChange, and is recreated
// here to avoid including a dependency on dbus for non-linux systems.
type unitFileChange struct {
	Type        string // one of 'link' or 'unlink'
	Filename    string // filename of the symlink
	Destination string // destination of the symlink
}
