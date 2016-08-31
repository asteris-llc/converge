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
	"io/ioutil"
	"path"
	"strings"

	"github.com/asteris-llc/converge/rpc/pb"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/kardianos/osext"
	"github.com/pkg/errors"
	"golang.org/x/net/context"
)

type resourceHost struct {
	auth *authorizer

	root                 string
	enableBinaryDownload bool
}

func (rh *resourceHost) GetBinary(ctx context.Context, _ *empty.Empty) (*pb.ContentResponse, error) {
	logger := getLogger(ctx).WithField("function", "resourceHost.GetBinary")

	if err := rh.auth.authorize(ctx); err != nil {
		logger.WithError(err).Info("failed authorization")
		return nil, err
	}

	if !rh.enableBinaryDownload {
		logger.Debug("got request for binary, but binary download not enabled")
		return nil, errors.New("binary download not enabled")
	}

	name, err := osext.Executable()
	if err != nil {
		logger.WithError(err).Error("could not determine binary location")
		return nil, errors.Wrap(err, "could not determine binary location")
	}

	content, err := ioutil.ReadFile(name)
	if err != nil {
		logger.WithError(err).Error("could not read binary")
		return nil, errors.Wrap(err, "could not read binary")
	}

	return &pb.ContentResponse{Content: string(content)}, nil
}

func (rh *resourceHost) GetModule(ctx context.Context, loc *pb.LocationRequest) (*pb.ContentResponse, error) {
	logger := getLogger(ctx).WithField("function", "resourceHost.GetModule").WithField("location", loc.Location)

	if err := rh.auth.authorize(ctx); err != nil {
		return nil, err
	}

	if rh.root == "" {
		logger.Debug("got request for module, but module download not enabled")
		return nil, errors.New("module download not enabled")
	}

	if strings.Contains(loc.Location, "..") {
		logger.Warning("got request for relative path")
		return nil, errors.New("cannot use relative paths")
	}

	content, err := ioutil.ReadFile(path.Join(rh.root, loc.Location))
	if err != nil {
		logger.WithError(err).Error("could not read location")
		return nil, errors.Wrap(err, "could not read location")
	}

	return &pb.ContentResponse{Content: string(content)}, nil
}
