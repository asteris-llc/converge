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

package exec

import (
	"sync"

	"golang.org/x/net/context"

	"github.com/asteris-llc/converge/load"
	"github.com/asteris-llc/converge/resource"
)

// PlanWithStatus plans the operations to be performed and outputs status
func PlanWithStatus(ctx context.Context, graph *load.Graph, status chan<- *StatusMessage) (results []*PlanResult, err error) {
	lock := new(sync.Mutex)

	err = graph.Walk(func(path string, res resource.Resource) error {
		select {
		case <-ctx.Done():
			return nil
		default:
		}

		monitor, ok := res.(resource.Monitor)
		if !ok {
			return nil
		}

		var (
			err    error
			result = &PlanResult{Path: path}
		)

		select {
		case <-ctx.Done():
			return nil
		case status <- &StatusMessage{Path: path, Status: "checking status"}:
		}

		result.CurrentStatus, result.WillChange, err = monitor.Check()
		if err != nil {
			return err
		}

		lock.Lock()
		defer lock.Unlock()
		results = append(results, result)

		return nil
	})

	return results, err
}

// Plan does the same thing as PlanWithStatus, but drops all status messages
func Plan(ctx context.Context, graph *load.Graph) (results []*PlanResult, err error) {
	status := make(chan *StatusMessage, 1)
	defer close(status)
	go func() {
		for range status {
			continue
		}
	}()

	return PlanWithStatus(ctx, graph, status)
}
