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

package systemd

import "github.com/asteris-llc/converge/resource"

func AppendStatus(a, b *resource.Status) *resource.Status {
	if a == nil && b == nil {
		return nil
	}
	if a == nil {
		return b
	}
	if b == nil {
		return a
	}

	for key, value := range b.Differences {
		a.Differences[key] = value
	}
	a.Output = append(a.Output, b.Output...)
	a.WillChange = a.WillChange || b.WillChange
	if b.WarningLevel > a.WarningLevel {
		a.WarningLevel = b.WarningLevel
		a.Status = b.Status
	}
	a.FailingDeps = append(a.FailingDeps, b.FailingDeps...)

	return a
}
