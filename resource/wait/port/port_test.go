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

package port_test

import (
	"errors"
	"fmt"
	"net"
	"regexp"
	"testing"
	"time"

	"github.com/asteris-llc/converge/helpers/fakerenderer"
	"github.com/asteris-llc/converge/helpers/logging"
	"github.com/asteris-llc/converge/resource"
	"github.com/asteris-llc/converge/resource/wait"
	"github.com/asteris-llc/converge/resource/wait/port"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

var (
	fakeConnectionFailureMsg = "connection failed"
	errFakeConnectionFailure = errors.New(fakeConnectionFailureMsg)
)

// TestPortCheck tests the implementation of Port.Check
func TestPortCheck(t *testing.T) {
	t.Parallel()
	defer logging.HideLogs(t)()

	runCheck := func(retries int, err error) (*port.Port, error) {
		portNum := 80
		mock := new(mockConnector)
		mock.On("CheckConnection", "", portNum).Return(err)
		p := &port.Port{
			Port:            portNum,
			ConnectionCheck: mock,
			Retrier:         &wait.Retrier{RetryCount: retries},
		}
		r, checkErr := p.Check(fakerenderer.New())
		require.IsType(t, (*port.Port)(nil), r)
		return r.(*port.Port), checkErr
	}

	t.Run("connection down", func(t *testing.T) {
		port, err := runCheck(0, errFakeConnectionFailure)
		require.NoError(t, err)
		assert.Equal(t, resource.StatusWillChange, port.Status.Level)
		require.Equal(t, 1, len(port.Messages()))
		assert.Regexp(t, regexp.MustCompile("^Failed to connect to"), port.Messages()[0])
		assert.Regexp(t, regexp.MustCompile(fakeConnectionFailureMsg), port.Messages()[0])
	})

	t.Run("connection down after retries", func(t *testing.T) {
		port, err := runCheck(3, errFakeConnectionFailure)
		require.NoError(t, err)
		assert.Equal(t, resource.StatusWillChange, port.Status.Level)
		assert.Equal(t, 2, len(port.Messages()))
		assert.Regexp(t, regexp.MustCompile("^Failed to connect to"), port.Messages()[0])
		assert.Regexp(t, regexp.MustCompile(fakeConnectionFailureMsg), port.Messages()[0])
		assert.Regexp(t, regexp.MustCompile("^Failed after"), port.Messages()[1])
	})

	t.Run("connection alive", func(t *testing.T) {
		port, err := runCheck(0, nil)
		require.NoError(t, err)
		assert.Equal(t, resource.StatusNoChange, port.Status.Level)
	})

	t.Run("connection alive after retries", func(t *testing.T) {
		port, err := runCheck(2, nil)
		require.NoError(t, err)
		assert.Equal(t, resource.StatusNoChange, port.Status.Level)
		assert.Equal(t, 1, len(port.Messages()))
		assert.Regexp(t, regexp.MustCompile("^Passed after"), port.Messages()[0])
	})
}

// TestPortApply tests the implementation of Port.Apply
func TestPortApply(t *testing.T) {
	t.Parallel()
	defer logging.HideLogs(t)()

	runApply := func(err error) (*port.Port, error) {
		portNum := 80
		mock := new(mockConnector)
		mock.On("CheckConnection", "", portNum).Return(err)
		p := &port.Port{
			Port:            portNum,
			ConnectionCheck: mock,
			Retrier: &wait.Retrier{
				MaxRetry: 3,
				Interval: 10 * time.Millisecond,
			},
		}
		r, err := p.Apply()
		require.IsType(t, (*port.Port)(nil), r)
		return r.(*port.Port), err
	}

	t.Run("passed", func(t *testing.T) {
		port, err := runApply(nil)
		require.NoError(t, err)

		t.Run("retry count", func(t *testing.T) {
			assert.Equal(t, 1, port.RetryCount)
		})
	})

	t.Run("retried", func(t *testing.T) {
		port, err := runApply(errFakeConnectionFailure)
		assert.Error(t, err)

		t.Run("retry count", func(t *testing.T) {
			assert.Equal(t, 3, port.RetryCount)
		})
	})

	t.Run("error", func(t *testing.T) {
		_, err := runApply(errFakeConnectionFailure)
		assert.Error(t, err)
	})
}

// TestTCPConnectionCheckCheckConnection tests
// TCPConnectionCheck.CheckConnection
func TestTCPConnectionCheckCheckConnection(t *testing.T) {
	t.Parallel()
	defer logging.HideLogs(t)()

	t.Run("success", func(t *testing.T) {
		portnum := 19323
		addr := fmt.Sprintf(":%d", portnum)
		l, err := net.Listen("tcp", addr)
		require.NoError(t, err, "failed to listen on %s", addr)
		defer l.Close()

		connChk := &port.TCPConnectionCheck{}
		err = connChk.CheckConnection("", portnum)
		require.NoError(t, err)
	})

	t.Run("failed", func(t *testing.T) {
		portnum := 19324
		connChk := &port.TCPConnectionCheck{}
		err := connChk.CheckConnection("", portnum)
		assert.Error(t, err, "some process might be listening on port 19324")
	})
}

type mockConnector struct {
	mock.Mock
}

func (m *mockConnector) CheckConnection(host string, portnum int) error {
	args := m.Called(host, portnum)
	return args.Error(0)
}
