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

// +build !solaris

package providers

import (
	"encoding/json"
	"fmt"

	pp "github.com/asteris-llc/converge/prettyprinters"
	"github.com/asteris-llc/converge/resource/docker/container"
	"github.com/asteris-llc/converge/resource/docker/image"
	"github.com/asteris-llc/converge/rpc/pb"
	"github.com/pkg/errors"
)

func dockerContainerLabel(val *pb.GraphComponent_Vertex) (pp.VisibleRenderable, error) {
	var dest = new(container.Container)
	if err := json.Unmarshal(val.Details, dest); err != nil {
		return nil, errors.Wrap(err, "could not unmarshal docker container")
	}

	return pp.VisibleString(fmt.Sprintf("Docker Container: %s", dest.Name)), nil
}

func dockerImageLabel(val *pb.GraphComponent_Vertex) (pp.VisibleRenderable, error) {
	var dest = new(image.Image)
	if err := json.Unmarshal(val.Details, dest); err != nil {
		return nil, errors.Wrap(err, "could not unmarshal docker image")
	}

	return pp.VisibleString(fmt.Sprintf("Docker Image: %s:%s", dest.Name, dest.Tag)), nil
}
