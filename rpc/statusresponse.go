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

package rpc

import (
	"github.com/asteris-llc/converge/graph/node"
	"github.com/asteris-llc/converge/prettyprinters/human"
	"github.com/asteris-llc/converge/rpc/pb"
)

func statusResponseFromPrintable(meta *node.Node, p human.Printable, stage pb.StatusResponse_Stage, run pb.StatusResponse_Run) *pb.StatusResponse {
	resp := &pb.StatusResponse{
		Id:    meta.ID, // TODO: deprecated, remove in 0.4.0
		Stage: stage,
		Run:   run,
		Meta:  pb.MetaFromNode(meta),

		Details: &pb.StatusResponse_Details{
			Messages:   p.Messages(),
			Changes:    map[string]*pb.DiffResponse{},
			HasChanges: p.HasChanges(),
			Warning:    p.Warning(),
		},
	}

	if err := p.Error(); err != nil {
		resp.Details.Error = err.Error()
	}

	for key, diff := range p.Changes() {
		resp.Details.Changes[key] = &pb.DiffResponse{
			Original: diff.Original(),
			Current:  diff.Current(),
			Changes:  diff.Changes(),
		}
	}

	return resp
}
