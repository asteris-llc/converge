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

package unit

import (
	"testing"
	"time"

	"github.com/pkg/errors"

	"github.com/asteris-llc/converge/helpers/fakerenderer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/context"
)

func TestCheck(t *testing.T) {
	fs := &mockFsExecutor{}
	fs.On("Walk", any, any).Return(nil)

	t.Parallel()
	t.Run("send-signal", func(t *testing.T) {
		r := &Resource{
			State:        "running",
			SignalName:   "SIGKILL",
			SignalNumber: 9,
			sendSignal:   true,
			fs:           fs,
		}
		e := &ExecutorMock{}
		r.systemdExecutor = e
		e.On("QueryUnit", any, any).Return(&Unit{ActiveState: "running"}, nil)
		status, err := r.Check(context.Background(), fakerenderer.New())
		assert.NoError(t, err)
		assert.True(t, status.HasChanges())
		assert.True(t, includesString(status.Messages(), "Sending signal `SIGKILL` to unit"))
	})
	t.Run("reload", func(t *testing.T) {
		r := &Resource{
			State:  "running",
			Reload: true,
			fs:     fs,
		}
		e := &ExecutorMock{}
		r.systemdExecutor = e
		e.On("QueryUnit", any, any).Return(&Unit{ActiveState: "running"}, nil)
		status, err := r.Check(context.Background(), fakerenderer.New())
		assert.NoError(t, err)
		assert.True(t, status.HasChanges())
		assert.True(t, includesString(status.Messages(), "Reloading unit configuration"))
		_, ok := status.Diffs()["state"]
		assert.True(t, ok)
	})
	t.Run("running", func(t *testing.T) {
		r := &Resource{
			Name:  "resource1",
			State: "running",
			fs:    fs,
		}
		t.Run("query-unit-returns-error", func(t *testing.T) {
			expected := errors.New("error1")
			e := &ExecutorMock{}
			e.On("QueryUnit", any, any).Return((*Unit)(nil), expected)
			r.systemdExecutor = e
			_, actual := r.Check(context.Background(), fakerenderer.New())
			assert.Equal(t, expected, actual)
		})
		t.Run("calls-query-unit", func(t *testing.T) {
			e := &ExecutorMock{}
			r.systemdExecutor = e
			e.On("QueryUnit", any, any).Return(&Unit{}, nil)
			r.Check(context.Background(), fakerenderer.New())
			e.AssertCalled(t, "QueryUnit", r.Name, false)
		})
		t.Run("when-status-active", func(t *testing.T) {
			unit := &Unit{ActiveState: "active"}
			e := &ExecutorMock{}
			r.systemdExecutor = e
			e.On("QueryUnit", any, any).Return(unit, nil)
			status, _ := r.Check(context.Background(), fakerenderer.New())
			assert.False(t, status.HasChanges())
			assert.True(t, includesString(status.Messages(), "already running"))
		})
		t.Run("when-status-reloading", func(t *testing.T) {
			unit := &Unit{ActiveState: "reloading"}
			e := &ExecutorMock{}
			r.systemdExecutor = e
			e.On("QueryUnit", any, any).Return(unit, nil)
			status, _ := r.Check(context.Background(), fakerenderer.New())
			assert.True(t, status.HasChanges())
			assert.True(t, includesString(status.Messages(), "unit is reloading, will re-check status during apply"))
		})
		t.Run("when-status-inactive", func(t *testing.T) {
			unit := &Unit{ActiveState: "inactive"}
			e := &ExecutorMock{}
			r.systemdExecutor = e
			e.On("QueryUnit", any, any).Return(unit, nil)
			status, _ := r.Check(context.Background(), fakerenderer.New())
			assert.True(t, status.HasChanges())
			_, ok := status.Diffs()["state"]
			assert.True(t, ok)
		})
		t.Run("when-status-failed", func(t *testing.T) {
			unit := &Unit{
				ActiveState:       "failed",
				Type:              UnitTypeService,
				ServiceProperties: &ServiceTypeProperties{Result: "success"},
			}
			e := &ExecutorMock{}
			r.systemdExecutor = e
			e.On("QueryUnit", any, any).Return(unit, nil)
			status, _ := r.Check(context.Background(), fakerenderer.New())
			assert.True(t, status.HasChanges())
			_, ok := status.Diffs()["state"]
			assert.True(t, ok)
			msg := `unit previously failed, the message was: the unit was activated successfully`
			assert.True(t, includesString(status.Messages(), msg))
		})
		t.Run("when-status-activating", func(t *testing.T) {
			unit := &Unit{ActiveState: "activating"}
			e := &ExecutorMock{}
			r.systemdExecutor = e
			e.On("QueryUnit", any, any).Return(unit, nil)
			status, _ := r.Check(context.Background(), fakerenderer.New())
			assert.True(t, status.HasChanges())
			assert.True(t, includesString(status.Messages(), "unit is alread activating, will re-check status during apply"))
		})
		t.Run("when-status-deactivating", func(t *testing.T) {
			unit := &Unit{ActiveState: "deactivating"}
			e := &ExecutorMock{}
			r.systemdExecutor = e
			e.On("QueryUnit", any, any).Return(unit, nil)
			status, _ := r.Check(context.Background(), fakerenderer.New())
			assert.True(t, status.HasChanges())
			_, ok := status.Diffs()["state"]
			assert.True(t, ok)
		})
	})
	t.Run("stopped", func(t *testing.T) {
		r := &Resource{
			Name:  "resource1",
			State: "stopped",
			fs:    fs,
		}
		t.Run("when-status-active", func(t *testing.T) {
			unit := &Unit{ActiveState: "active"}
			e := &ExecutorMock{}
			r.systemdExecutor = e
			e.On("QueryUnit", any, any).Return(unit, nil)
			status, _ := r.Check(context.Background(), fakerenderer.New())
			assert.True(t, status.HasChanges())
			_, ok := status.Diffs()["state"]
			assert.True(t, ok)
		})
		t.Run("when-status-reloading", func(t *testing.T) {
			unit := &Unit{ActiveState: "reloading"}
			e := &ExecutorMock{}
			r.systemdExecutor = e
			e.On("QueryUnit", any, any).Return(unit, nil)
			status, _ := r.Check(context.Background(), fakerenderer.New())
			assert.True(t, status.HasChanges())
		})
		t.Run("when-status-inactive", func(t *testing.T) {
			unit := &Unit{ActiveState: "inactive"}
			e := &ExecutorMock{}
			r.systemdExecutor = e
			e.On("QueryUnit", any, any).Return(unit, nil)
			status, _ := r.Check(context.Background(), fakerenderer.New())
			assert.False(t, status.HasChanges())
		})
		t.Run("when-status-failed", func(t *testing.T) {
			unit := &Unit{
				ActiveState:       "failed",
				Type:              UnitTypeService,
				ServiceProperties: &ServiceTypeProperties{Result: "success"},
			}
			e := &ExecutorMock{}
			r.systemdExecutor = e
			e.On("QueryUnit", any, any).Return(unit, nil)
			status, _ := r.Check(context.Background(), fakerenderer.New())
			assert.False(t, status.HasChanges())
			msg := `unit previously failed, the message was: the unit was activated successfully`
			assert.True(t, includesString(status.Messages(), msg))
		})
		t.Run("when-status-activating", func(t *testing.T) {
			unit := &Unit{ActiveState: "activating"}
			e := &ExecutorMock{}
			r.systemdExecutor = e
			e.On("QueryUnit", any, any).Return(unit, nil)
			status, _ := r.Check(context.Background(), fakerenderer.New())
			assert.True(t, status.HasChanges())
			_, ok := status.Diffs()["state"]
			assert.True(t, ok)
		})
		t.Run("when-status-deactivating", func(t *testing.T) {
			unit := &Unit{ActiveState: "deactivating"}
			e := &ExecutorMock{}
			r.systemdExecutor = e
			e.On("QueryUnit", any, any).Return(unit, nil)
			status, _ := r.Check(context.Background(), fakerenderer.New())
			assert.True(t, status.HasChanges())
		})
	})
	t.Run("restarted", func(t *testing.T) {
		r := &Resource{
			Name:  "resource1",
			State: "restarted",
			fs:    fs,
		}
		e := &ExecutorMock{}
		r.systemdExecutor = e
		states := []string{
			"active",
			"reloading",
			"inactive",
			"failed",
			"activating",
			"deactivating",
		}
		for _, st := range states {
			u := &Unit{ActiveState: st}
			e.On("QueryUnit", any, any).Return(u, nil)
			status, err := r.Check(context.Background(), fakerenderer.New())
			require.NoError(t, err)
			assert.True(t, status.HasChanges())
		}
	})
}

func TestGetFailedReason(t *testing.T) {
	t.Parallel()

	t.Run("returns-error-when-no-properties", func(t *testing.T) {
		t.Parallel()
		supportedTypes := []UnitType{
			UnitTypeService,
			UnitTypeSocket,
			UnitTypeMount,
			UnitTypeAutoMount,
			UnitTypeSwap,
			UnitTypeTimer,
		}
		for _, typ := range supportedTypes {
			_, err := getFailedReason(&Unit{Type: typ})
			assert.EqualError(t, err, "unable to determine cause of failure: no properties available")
		}
	})
	t.Run("looks-at-correct-field-for-type", func(t *testing.T) {
		t.Parallel()
		t.Run("service", func(t *testing.T) {
			t.Parallel()
			u := &Unit{Type: UnitTypeService, ServiceProperties: &ServiceTypeProperties{Result: "success"}}
			actual, err := getFailedReason(u)
			require.NoError(t, err)
			assert.Equal(t, "the unit was activated successfully", actual)
		})
		t.Run("socket", func(t *testing.T) {
			t.Parallel()
			u := &Unit{Type: UnitTypeSocket, SocketProperties: &SocketTypeProperties{Result: "success"}}
			actual, err := getFailedReason(u)
			require.NoError(t, err)
			assert.Equal(t, "the unit was activated successfully", actual)
		})
		t.Run("mount", func(t *testing.T) {
			t.Parallel()
			u := &Unit{Type: UnitTypeMount, MountProperties: &MountTypeProperties{Result: "success"}}
			actual, err := getFailedReason(u)
			require.NoError(t, err)
			assert.Equal(t, "the unit was activated successfully", actual)
		})
		t.Run("automount", func(t *testing.T) {
			t.Parallel()
			u := &Unit{Type: UnitTypeAutoMount, AutomountProperties: &AutomountTypeProperties{Result: "success"}}
			actual, err := getFailedReason(u)
			require.NoError(t, err)
			assert.Equal(t, "the unit was activated successfully", actual)
		})
		t.Run("swap", func(t *testing.T) {
			t.Parallel()
			u := &Unit{Type: UnitTypeSwap, SwapProperties: &SwapTypeProperties{Result: "success"}}
			actual, err := getFailedReason(u)
			require.NoError(t, err)
			assert.Equal(t, "the unit was activated successfully", actual)
		})
		t.Run("timer", func(t *testing.T) {
			t.Parallel()
			u := &Unit{Type: UnitTypeTimer, TimerProperties: &TimerTypeProperties{Result: "success"}}
			actual, err := getFailedReason(u)
			require.NoError(t, err)
			assert.Equal(t, "the unit was activated successfully", actual)
		})
	})
	t.Run("returns-correct-reason", func(t *testing.T) {
		t.Parallel()
		reasons := map[string]string{
			"success":                  "the unit was activated successfully",
			"resources":                "not enough resources available to create the process",
			"timeout":                  "a timeout occurred while starting the unit",
			"exit-code":                "unit exited with a non-zero exit code",
			"signal":                   "unit exited due to an unhandled signal",
			"core-dump":                "unit exited and dumped core",
			"watchdog":                 "watchdog terminated the service due to slow or missing responses",
			"start-limit":              "process has been restarted too many times",
			"service-failed-permanent": "continual failure of this socket",
		}
		for reason, explanation := range reasons {
			u := &Unit{Type: UnitTypeService, ServiceProperties: &ServiceTypeProperties{Result: reason}}
			actual, err := getFailedReason(u)
			require.NoError(t, err)
			assert.Equal(t, explanation, actual)
		}
	})
	t.Run("returns-unkown-for-types-without-result", func(t *testing.T) {
		t.Parallel()
		unknownTypes := []UnitType{
			UnitTypeUnknown,
			UnitTypeDevice,
			UnitTypeTarget,
			UnitTypePath,
			UnitTypeSnapshot,
			UnitTypeSlice,
			UnitTypeScope,
		}
		for _, typ := range unknownTypes {
			actual, err := getFailedReason(&Unit{Type: typ})
			require.NoError(t, err)
			assert.Equal(t, "unknown reason", actual)
		}
	})
}

// TestApply runs a test
func TestApply(t *testing.T) {
	fs := &mockFsExecutor{}
	fs.On("Walk", any, any).Return(nil)

	t.Parallel()
	t.Run("query-unit-returns-error", func(t *testing.T) {
		t.Parallel()
		expected := errors.New("error1")
		r := &Resource{fs: fs}
		e := &ExecutorMock{}
		e.On("QueryUnit", any, any).Return((*Unit)(nil), expected)
		r.systemdExecutor = e
		_, err := r.Apply(context.Background())
		assert.Equal(t, expected, err)
	})

	t.Run("when-send-signal", func(t *testing.T) {
		t.Parallel()
		u := &Unit{ActiveState: "active"}
		r := &Resource{
			ActiveState:  "running",
			SignalName:   "SIGKILL",
			SignalNumber: 9,
			sendSignal:   true,
			fs:           fs,
		}
		e := &ExecutorMock{}
		e.On("QueryUnit", any, any).Return(u, nil)
		e.On("SendSignal", any, any).Return()
		r.systemdExecutor = e
		status, err := r.Apply(context.Background())
		assert.NoError(t, err)
		e.AssertCalled(t, "SendSignal", any, any)
		assert.True(t, includesString(status.Messages(), "Sending signal `SIGKILL` to unit"))
	})

	t.Run("when-reload", func(t *testing.T) {
		t.Parallel()

		t.Run("when-no-error", func(t *testing.T) {
			t.Parallel()
			r := &Resource{
				State:  "running",
				Reload: true,
				fs:     fs,
			}
			e := &ExecutorMock{}
			u := &Unit{ActiveState: "active"}
			r.systemdExecutor = e
			e.On("QueryUnit", any, any).Return(u, nil)
			e.On("ReloadUnit", any, any).Return(nil)
			_, err := r.Apply(context.Background())
			assert.NoError(t, err)
			e.AssertCalled(t, "ReloadUnit", u)
		})

		t.Run("when-error", func(t *testing.T) {
			t.Parallel()
			r := &Resource{
				State:  "running",
				Reload: true,
				fs:     fs,
			}
			e := &ExecutorMock{}
			u := &Unit{ActiveState: "active"}
			r.systemdExecutor = e
			expected := errors.New("error1")
			e.On("QueryUnit", any, any).Return(u, nil)
			e.On("ReloadUnit", any, any).Return(expected)
			_, err := r.Apply(context.Background())
			assert.Equal(t, expected, err)
		})
	})

	t.Run("when-want-running", func(t *testing.T) {
		t.Parallel()
		t.Run("start-returns-error", func(t *testing.T) {
			r := &Resource{State: "running", fs: fs}
			u := &Unit{ActiveState: "inactive"}
			e := &ExecutorMock{}
			expected := errors.New("error1")
			e.On("QueryUnit", any, any).Return(u, nil)
			e.On("StartUnit", any).Return(expected)
			e.On("StopUnit", any).Return(nil)
			e.On("RestartUnit", any).Return(nil)
			r.systemdExecutor = e
			_, err := r.Apply(context.Background())
			assert.Equal(t, expected, err)
		})
		t.Run("status-is-active", func(t *testing.T) {
			r := &Resource{State: "running", fs: fs}
			u := &Unit{ActiveState: "active"}
			e := &ExecutorMock{}
			e.On("QueryUnit", any, any).Return(u, nil)
			e.On("StartUnit", any).Return(nil)
			e.On("StopUnit", any).Return(nil)
			e.On("RestartUnit", any).Return(nil)
			r.systemdExecutor = e
			_, err := r.Apply(context.Background())
			require.NoError(t, err)
			e.AssertNotCalled(t, "StartUnit", u)
		})
		t.Run("status-is-reloading", func(t *testing.T) {
			r := &Resource{State: "running", fs: fs}
			u := &Unit{ActiveState: "reloading"}
			e := &ExecutorMock{}
			e.On("QueryUnit", any, any).Return(u, nil)
			e.On("StartUnit", any).Return(nil)
			e.On("StopUnit", any).Return(nil)
			e.On("RestartUnit", any).Return(nil)
			r.systemdExecutor = e
			_, err := r.Apply(context.Background())
			require.NoError(t, err)
			e.AssertCalled(t, "StartUnit", u)
		})
		t.Run("status-is-inactive", func(t *testing.T) {
			r := &Resource{State: "running", fs: fs}
			u := &Unit{ActiveState: "inactive"}
			e := &ExecutorMock{}
			e.On("QueryUnit", any, any).Return(u, nil)
			e.On("StartUnit", any).Return(nil)
			e.On("StopUnit", any).Return(nil)
			e.On("RestartUnit", any).Return(nil)
			r.systemdExecutor = e
			_, err := r.Apply(context.Background())
			require.NoError(t, err)
			e.AssertCalled(t, "StartUnit", u)
		})
		t.Run("status-is-failed", func(t *testing.T) {
			r := &Resource{State: "running", fs: fs}
			u := &Unit{ActiveState: "failed"}
			e := &ExecutorMock{}
			e.On("QueryUnit", any, any).Return(u, nil)
			e.On("StartUnit", any).Return(nil)
			e.On("StopUnit", any).Return(nil)
			e.On("RestartUnit", any).Return(nil)
			r.systemdExecutor = e
			_, err := r.Apply(context.Background())
			require.NoError(t, err)
			e.AssertCalled(t, "StartUnit", u)
		})
		t.Run("status-is-activating", func(t *testing.T) {
			r := &Resource{State: "running", fs: fs}
			u := &Unit{ActiveState: "activating"}
			e := &ExecutorMock{}
			e.On("QueryUnit", any, any).Return(u, nil)
			e.On("StartUnit", any).Return(nil)
			e.On("StopUnit", any).Return(nil)
			e.On("RestartUnit", any).Return(nil)
			r.systemdExecutor = e
			_, err := r.Apply(context.Background())
			require.NoError(t, err)
			e.AssertCalled(t, "StartUnit", u)
		})
		t.Run("status-is-deactivating", func(t *testing.T) {
			r := &Resource{State: "running", fs: fs}
			u := &Unit{ActiveState: "deactivating"}
			e := &ExecutorMock{}
			e.On("QueryUnit", any, any).Return(u, nil)
			e.On("StartUnit", any).Return(nil)
			e.On("StopUnit", any).Return(nil)
			e.On("RestartUnit", any).Return(nil)
			r.systemdExecutor = e
			_, err := r.Apply(context.Background())
			require.NoError(t, err)
			e.AssertCalled(t, "StartUnit", u)
		})
	})

	t.Run("when-want-stopped", func(t *testing.T) {
		t.Parallel()
		t.Run("stop-returns-error", func(t *testing.T) {
			r := &Resource{State: "stopped", fs: fs}
			u := &Unit{ActiveState: "active"}
			e := &ExecutorMock{}
			expected := errors.New("error1")
			e.On("QueryUnit", any, any).Return(u, nil)
			e.On("StartUnit", any).Return(nil)
			e.On("StopUnit", any).Return(expected)
			e.On("RestartUnit", any).Return(nil)
			r.systemdExecutor = e
			_, err := r.Apply(context.Background())
			assert.Equal(t, expected, err)
		})
		t.Run("status-is-active", func(t *testing.T) {
			r := &Resource{State: "stopped", fs: fs}
			u := &Unit{ActiveState: "active"}
			e := &ExecutorMock{}
			e.On("QueryUnit", any, any).Return(u, nil)
			e.On("StartUnit", any).Return(nil)
			e.On("StopUnit", any).Return(nil)
			e.On("RestartUnit", any).Return(nil)
			r.systemdExecutor = e
			_, err := r.Apply(context.Background())
			require.NoError(t, err)
			e.AssertCalled(t, "StopUnit", u)
		})
		t.Run("status-is-reloading", func(t *testing.T) {
			r := &Resource{State: "stopped", fs: fs}
			u := &Unit{ActiveState: "reloading"}
			e := &ExecutorMock{}
			e.On("QueryUnit", any, any).Return(u, nil)
			e.On("StartUnit", any).Return(nil)
			e.On("StopUnit", any).Return(nil)
			e.On("RestartUnit", any).Return(nil)
			r.systemdExecutor = e
			_, err := r.Apply(context.Background())
			require.NoError(t, err)
			e.AssertCalled(t, "StopUnit", u)
		})
		t.Run("status-is-inactive", func(t *testing.T) {
			r := &Resource{State: "stopped", fs: fs}
			u := &Unit{ActiveState: "inactive"}
			e := &ExecutorMock{}
			e.On("QueryUnit", any, any).Return(u, nil)
			e.On("StartUnit", any).Return(nil)
			e.On("StopUnit", any).Return(nil)
			e.On("RestartUnit", any).Return(nil)
			r.systemdExecutor = e
			_, err := r.Apply(context.Background())
			require.NoError(t, err)
			e.AssertNotCalled(t, "StopUnit", u)
		})
		t.Run("status-is-failed", func(t *testing.T) {
			r := &Resource{State: "stopped", fs: fs}
			u := &Unit{ActiveState: "failed"}
			e := &ExecutorMock{}
			e.On("QueryUnit", any, any).Return(u, nil)
			e.On("StartUnit", any).Return(nil)
			e.On("StopUnit", any).Return(nil)
			e.On("RestartUnit", any).Return(nil)
			r.systemdExecutor = e
			_, err := r.Apply(context.Background())
			require.NoError(t, err)
			e.AssertNotCalled(t, "StopUnit", u)
		})
		t.Run("status-is-activating", func(t *testing.T) {
			r := &Resource{State: "stopped", fs: fs}
			u := &Unit{ActiveState: "activating"}
			e := &ExecutorMock{}
			e.On("QueryUnit", any, any).Return(u, nil)
			e.On("StartUnit", any).Return(nil)
			e.On("StopUnit", any).Return(nil)
			e.On("RestartUnit", any).Return(nil)
			r.systemdExecutor = e
			_, err := r.Apply(context.Background())
			require.NoError(t, err)
			e.AssertCalled(t, "StopUnit", u)
		})
		t.Run("status-is-deactivating", func(t *testing.T) {
			r := &Resource{State: "stopped", fs: fs}
			u := &Unit{ActiveState: "deactivating"}
			e := &ExecutorMock{}
			e.On("QueryUnit", any, any).Return(u, nil)
			e.On("StartUnit", any).Return(nil)
			e.On("StopUnit", any).Return(nil)
			e.On("RestartUnit", any).Return(nil)
			r.systemdExecutor = e
			_, err := r.Apply(context.Background())
			require.NoError(t, err)
			e.AssertCalled(t, "StopUnit", u)
		})
	})
	t.Run("when-want-restarted", func(t *testing.T) {
		t.Parallel()
		t.Run("when-restart-returns-error", func(t *testing.T) {
			t.Parallel()
			r := &Resource{State: "restarted", fs: fs}
			u := &Unit{ActiveState: "active"}
			e := &ExecutorMock{}
			r.systemdExecutor = e
			expected := errors.New("error1")
			e.On("QueryUnit", any, any).Return(u, nil)
			e.On("RestartUnit", any).Return(expected)
			_, err := r.Apply(context.Background())
			assert.Equal(t, expected, err)
		})
		t.Run("calls-restart", func(t *testing.T) {
			t.Parallel()
			states := []string{"active", "inactive", "activating", "deactivating", "reloading", "failed"}
			for _, st := range states {
				t.Run(st, func(t *testing.T) {
					u := &Unit{ActiveState: st}
					r := &Resource{State: "restarted", fs: fs}
					e := &ExecutorMock{}
					e.On("RestartUnit", any).Return(nil)
					e.On("QueryUnit", any, any).Return(u, nil)
					r.systemdExecutor = e
					_, err := r.Apply(context.Background())
					require.NoError(t, err)
					e.AssertCalled(t, "RestartUnit", u)
				})
			}
		})
	})
	t.Run("tries-to-enable-unit", func(t *testing.T) {
		True := true
		fs := newMockWithPaths()
		fs.On("Walk", any, any).Return(nil)
		fs.On("EvalSymlinks", any).Return("", nil)
		u := &Unit{
			Name: "name1.service",
			Path: "/lib/systemd/system/name1.service",
		}
		e := &ExecutorMock{}
		e.On("QueryUnit", any, any).Return(u, nil)
		e.On("EnableUnit", any, any, any).Return(false, []*unitFileChange{}, nil)
		r := &Resource{
			Name:            "name1.service",
			systemdExecutor: e,
			enableChange:    &True,
			fs:              fs,
		}
		_, err := r.Apply(context.Background())
		e.AssertCalled(t, "EnableUnit", any, any, any)
		require.NoError(t, err)
	})
	t.Run("tries-to-disable-unit", func(t *testing.T) {
		fs := newMockWithPaths("/etc/systemd/system/name1.service")
		fs.On("Walk", any, any).Return(nil)
		fs.On("EvalSymlinks", any).Return("/etc/systemd/system/name1.service", nil)
		u := &Unit{
			Name: "name1.service",
		}
		e := &ExecutorMock{}
		e.On("QueryUnit", any, any).Return(u, nil)
		e.On("EnableUnit", any, any, any).Return(false, []*unitFileChange{}, nil)
		e.On("DisableUnit", any, any).Return([]*unitFileChange{}, nil)
		r := &Resource{
			Name:            "name1.service",
			systemdExecutor: e,
			enableChange:    new(bool),
			fs:              fs,
		}
		_, err := r.Apply(context.Background())
		e.AssertCalled(t, "DisableUnit", any, any)
		require.NoError(t, err)
	})
	t.Run("updates-enablement", func(t *testing.T) {
		t.Parallel()
		fs := newMockWithPaths(
			"/run/systemd/system/name1.service",
		)
		fs.On("Walk", any, any).Return(nil)
		fs.On("EvalSymlinks", any).Return("/lib/systemd/system/name1.service", nil)
		u := &Unit{
			Name: "name1.service",
			Path: "/lib/systemd/system/name1.service",
		}
		e := &ExecutorMock{}
		e.On("QueryUnit", any, any).Return(u, nil)
		r := &Resource{
			Name:            "name1.service",
			systemdExecutor: e,
			fs:              fs,
		}
		_, err := r.Apply(context.Background())
		require.NoError(t, err)
		assert.True(t, r.EnabledRuntime)
	})
}

// TestCheckAfterApply runs a test
func TestCheckAfterApply(t *testing.T) {
	t.Parallel()
	fs := &mockFsExecutor{}
	fs.On("Walk", any, any).Return(nil)

	t.Run("when-send-signal", func(t *testing.T) {
		t.Parallel()
		r := &Resource{
			State:        "running",
			SignalName:   "SIGKILL",
			SignalNumber: 9,
			sendSignal:   true,
			fs:           fs,
		}
		u := &Unit{ActiveState: "active"}
		e := &ExecutorMock{}
		r.systemdExecutor = e
		e.On("QueryUnit", any, any).Return(u, nil)
		e.On("SendSignal", any, any).Return()
		status, err := r.Check(context.Background(), fakerenderer.New())
		status, err = r.Apply(context.Background())
		status, err = r.Check(context.Background(), fakerenderer.New())
		assert.NoError(t, err)
		assert.False(t, status.HasChanges())
	})

	t.Run("when-reload", func(t *testing.T) {
		t.Parallel()
		r := &Resource{
			State:  "running",
			Reload: true,
			fs:     fs,
		}
		u := &Unit{ActiveState: "active"}
		e := &ExecutorMock{}
		r.systemdExecutor = e
		e.On("QueryUnit", any, any).Return(u, nil)
		e.On("ReloadUnit", any).Return(nil)
		status, err := r.Check(context.Background(), fakerenderer.New())
		status, err = r.Apply(context.Background())
		status, err = r.Check(context.Background(), fakerenderer.New())
		assert.NoError(t, err)
		assert.False(t, status.HasChanges())
	})
}

// TestHandlesContext runs a test
func TestHandlesContext(t *testing.T) {
	fs := &mockFsExecutor{}
	fs.On("Walk", any, any).Return(nil)

	t.Parallel()

	t.Run("Check", func(t *testing.T) {
		t.Parallel()
		t.Run("when-timeout", func(t *testing.T) {
			t.Parallel()
			expected := "context was cancelled"
			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
			time.Sleep(2 * time.Millisecond)
			r := &Resource{fs: fs}
			e := &ExecutorMock{DoSleep: true, SleepFor: (10 * time.Millisecond)}
			e.On("QueryUnit", any, any).Return(&Unit{}, nil)
			r.systemdExecutor = e
			_, err := r.Check(ctx, fakerenderer.New())
			assert.EqualError(t, err, expected)
			cancel()
		})

		t.Run("when-canceled", func(t *testing.T) {
			t.Parallel()
			ctx, cancel := context.WithCancel(context.Background())
			r := &Resource{fs: fs}
			e := &ExecutorMock{DoSleep: true, SleepFor: (10 * time.Millisecond)}
			e.On("QueryUnit", any, any).Return(&Unit{}, nil)
			r.systemdExecutor = e
			cancel()
			_, err := r.Check(ctx, fakerenderer.New())
			assert.Error(t, err)
		})
	})

	t.Run("Apply", func(t *testing.T) {
		t.Run("when-timeout", func(t *testing.T) {
			t.Parallel()
			expected := "context was cancelled"
			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
			time.Sleep(2 * time.Millisecond)
			r := &Resource{fs: fs}
			e := &ExecutorMock{DoSleep: true, SleepFor: (10 * time.Millisecond)}
			e.On("QueryUnit", any, any).Return(&Unit{}, nil)
			r.systemdExecutor = e
			_, err := r.Apply(ctx)
			assert.EqualError(t, err, expected)
			cancel()
		})
		t.Run("when-canceled", func(t *testing.T) {
			t.Parallel()
			ctx, cancel := context.WithCancel(context.Background())
			r := &Resource{fs: fs}
			e := &ExecutorMock{DoSleep: true, SleepFor: (10 * time.Millisecond)}
			e.On("QueryUnit", any, any).Return(&Unit{}, nil)
			r.systemdExecutor = e
			cancel()
			_, err := r.Apply(ctx)
			assert.Error(t, err)
		})

	})
}

// TestIsEnabled tests validation of whether or not a unit is enabled
func TestIsEnabled(t *testing.T) {
	t.Parallel()
	t.Run("check-sets-fields", func(t *testing.T) {
		t.Parallel()
		t.Run("when-no-path-set", func(t *testing.T) {
			t.Run("when-not-enabled", func(t *testing.T) {
				t.Parallel()
				fs := &mockFsExecutor{}
				fs.On("Walk", any, any).Return(nil)
				r := &Resource{fs: fs}
				u := &Unit{Name: "name1.service", Path: ""}
				runtime, persistent, err := r.isEnabled(u)
				assert.NoError(t, err)
				assert.False(t, runtime)
				assert.False(t, persistent)
			})
			t.Run("when-enabled", func(t *testing.T) {
				t.Parallel()
				fs := newMockWithPaths("/etc/systemd/system/name1.service")
				fs.On("Walk", any, any).Return(nil)
				r := &Resource{fs: fs}
				u := &Unit{Name: "name1.service", Path: ""}
				runtime, persistent, err := r.isEnabled(u)
				assert.NoError(t, err)
				assert.False(t, runtime)
				assert.True(t, persistent)
			})
			t.Run("when-runtime-enabled", func(t *testing.T) {
				t.Parallel()
				fs := newMockWithPaths("/run/systemd/system/name1.service")
				fs.On("Walk", any, any).Return(nil)
				r := &Resource{fs: fs}
				u := &Unit{Name: "name1.service", Path: ""}
				runtime, persistent, err := r.isEnabled(u)
				assert.NoError(t, err)
				assert.True(t, runtime)
				assert.False(t, persistent)
			})
			t.Run("when-enabled-and-runtime-enabled", func(t *testing.T) {
				t.Parallel()
				fs := newMockWithPaths(
					"/run/systemd/system/name1.service",
					"/etc/systemd/system/name1.service",
				)
				fs.On("Walk", any, any).Return(nil)
				r := &Resource{fs: fs}
				u := &Unit{Name: "name1.service", Path: ""}
				runtime, persistent, err := r.isEnabled(u)
				assert.NoError(t, err)
				assert.True(t, runtime)
				assert.True(t, persistent)
			})
		})
		t.Run("when-path-set", func(t *testing.T) {
			t.Parallel()
			t.Run("when-disabled", func(t *testing.T) {
				t.Parallel()
				fs := newMockWithPaths()
				fs.On("Walk", any, any).Return(nil)
				fs.On("EvalSymlinks", any).Return("", nil)
				r := &Resource{fs: fs}
				u := &Unit{Name: "name1.service", Path: "/lib/systemd/system/name1.service"}
				runtime, persistent, err := r.isEnabled(u)
				assert.NoError(t, err)
				assert.False(t, runtime)
				assert.False(t, persistent)
			})
			t.Run("when-enabled", func(t *testing.T) {
				t.Parallel()
				fs := newMockWithPaths("/run/systemd/system/name1.service")
				fs.On("Walk", any, any).Return(nil)
				fs.On("EvalSymlinks", any).Return("/path1", nil)
				r := &Resource{fs: fs}
				u := &Unit{Name: "name1.service", Path: "/path1"}
				runtime, persistent, err := r.isEnabled(u)
				assert.NoError(t, err)
				assert.True(t, runtime)
				assert.False(t, persistent)
			})
			t.Run("when-enabled-with-different-symlink", func(t *testing.T) {
				t.Parallel()
				fs := newMockWithPaths("/run/systemd/system/name1.service")
				fs.On("Walk", any, any).Return(nil)
				fs.On("EvalSymlinks", any).Return("/path2", nil)
				r := &Resource{fs: fs}
				u := &Unit{Name: "name1.service", Path: "/path1"}
				runtime, persistent, err := r.isEnabled(u)
				assert.NoError(t, err)
				assert.False(t, runtime)
				assert.False(t, persistent)
			})
		})
	})
}

func TestExistsInTree(t *testing.T) {
	t.Parallel()
	t.Run("when-path-and-symlink-target-match", func(t *testing.T) {
		fs := newMockWithPaths("/run/systemd/system/name1.service")
		fs.On("Walk", any, any).Return(nil)
		fs.On("EvalSymlinks", any).Return("/lib/systemd/system/name-full.service", nil)
		r := &Resource{fs: fs}
		u := &Unit{Name: "name1.service", Path: "/lib/systemd/system/name-full.service"}
		inTree, err := r.existsInTree("/run", u)
		require.NoError(t, err)
		assert.True(t, inTree)
	})
	t.Run("when-path-and-symlink-target-mismatch", func(t *testing.T) {
		fs := newMockWithPaths("/run/systemd/system/name1.service")
		fs.On("Walk", any, any).Return(nil)
		fs.On("EvalSymlinks", any).Return("/lib/systemd/system/name-full.service", nil)
		r := &Resource{fs: fs}
		u := &Unit{Name: "name1.service", Path: "/etc/init.d/name-full"}
		inTree, err := r.existsInTree("/run", u)
		require.NoError(t, err)
		assert.False(t, inTree)
	})
	t.Run("when-symlink-error", func(t *testing.T) {
		expected := errors.New("error1")
		fs := newMockWithPaths("/run/systemd/system/name1.service")
		fs.On("Walk", any, any).Return(expected)
		fs.On("EvalSymlinks", any).Return("", nil)
		r := &Resource{fs: fs}
		u := &Unit{Name: "name1.service", Path: "/etc/init.d/name-full"}
		_, err := r.existsInTree("/run", u)
		assert.Equal(t, expected, err)
	})
}
