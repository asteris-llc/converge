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

// +build linux

package unit

import "github.com/stretchr/testify/mock"

const (
	any = mock.Anything
)

var (
	loadStates   = []string{"loaded", "error", "masked"}
	activeStates = []string{
		"active",
		"reloading",
		"inactive",
		"failed",
		"activating",
		"deactivating",
	}
	validTypes = []string{
		"service",
		"socket",
		"device",
		"mount",
		"automount",
		"swap",
		"target",
		"path",
		"timer",
		"snapshot",
		"slice",
		"scope",
	}
	alphabet = "abcdefhjijklmnopqrstuvwxyz"
)
