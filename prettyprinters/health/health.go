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

package health

import (
	"github.com/asteris-llc/converge/prettyprinters/human"
	"github.com/asteris-llc/converge/resource"
)

// Printer for health checks
type Printer struct {
	*human.Printer
}

// New returns a new Printer with an embedded human printer that hides
// non-healthcheck nodes
func New() *Printer {
	humanPrinter := human.NewFiltered(func(id string, value human.Printable) bool {
		_, ok := value.(resource.Check)
		return ok
	})
	return &Printer{humanPrinter}
}
