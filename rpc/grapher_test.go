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
	"bytes"
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/Sirupsen/logrus"
	"github.com/asteris-llc/converge/helpers/logging"
	"github.com/asteris-llc/converge/rpc/pb"
	"github.com/asteris-llc/converge/rpc/pb/mocks"
	"github.com/fgrid/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGrapherGraph(t *testing.T) {
	g := grapher{auth: new(authorizer)}

	logger := logrus.New()
	logger.Out = new(bytes.Buffer)
	ctx := logging.WithLogger(context.Background(), logger)

	t.Run("good request", func(t *testing.T) {
		stream := new(mocks.Grapher_GraphServer)
		stream.On("Context").Return(ctx)
		stream.On("Send", mock.Anything).Return(nil)

		err := g.Graph(&pb.LoadRequest{Location: "../samples/basic.hcl"}, stream)
		assert.NoError(t, err)

		stream.AssertNumberOfCalls(t, "Send", 9)
	})

	t.Run("bad file", func(t *testing.T) {
		stream := new(mocks.Grapher_GraphServer)
		stream.On("Context").Return(ctx)
		stream.On("Send", mock.Anything).Return(nil)

		filename := uuid.NewV4().String()
		err := g.Graph(&pb.LoadRequest{Location: filename}, stream)
		assert.EqualError(
			t,
			err,
			fmt.Sprintf(
				"loading failed: loading %s: loading failed: file://%s: open %s: no such file or directory",
				filename,
				filename,
				filename,
			),
		)
	})

	t.Run("stream error", func(t *testing.T) {
		stream := new(mocks.Grapher_GraphServer)
		stream.On("Context").Return(ctx)
		stream.On("Send", mock.Anything).Return(errors.New("fake error"))

		err := g.Graph(&pb.LoadRequest{Location: "../samples/basic.hcl"}, stream)
		assert.Error(t, err)
	})
}
