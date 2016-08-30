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

package pb

import (
	"errors"

	"github.com/asteris-llc/converge/prettyprinters/human"
	"github.com/asteris-llc/converge/resource"
)

// This file implements the Printable interface for pretty-printing graphs that
// have been rehydrated from information from the RPC

// ToPrintable returns a view that can be used in a human printer
func (sr *StatusResponse_Details) ToPrintable() human.Printable {
	psr := &printableStatusResponse{
		changes:    map[string]resource.Diff{},
		messages:   sr.Messages,
		hasChanges: sr.HasChanges,
		error:      nil,
	}

	// set up changes
	for k, v := range sr.GetChanges() {
		psr.changes[k] = v.ToPrintable()
	}

	// set up error
	if sr.Error != "" {
		psr.error = errors.New(sr.Error)
	}

	return psr
}

type printableStatusResponse struct {
	changes    map[string]resource.Diff
	messages   []string
	hasChanges bool
	error      error
}

func (psr *printableStatusResponse) Changes() map[string]resource.Diff { return psr.changes }
func (psr *printableStatusResponse) Messages() []string                { return psr.messages }
func (psr *printableStatusResponse) HasChanges() bool                  { return psr.hasChanges }
func (psr *printableStatusResponse) Error() error                      { return psr.error }

// ToPrintable returns a view that can be used in a human printer
func (d *DiffResponse) ToPrintable() resource.Diff {
	return &printableDiff{
		original: d.Original,
		current:  d.Current,
		changes:  d.Changes,
	}
}

type printableDiff struct {
	original string
	current  string
	changes  bool
}

func (pd *printableDiff) Original() string { return pd.original }
func (pd *printableDiff) Current() string  { return pd.current }
func (pd *printableDiff) Changes() bool    { return pd.changes }
