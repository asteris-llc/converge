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

import "github.com/stretchr/testify/mock"

type ExecutorMock struct {
	mock.Mock
}

func (m *ExecutorMock) ListUnits() ([]*Unit, error) {
	args := m.Called()
	return args.Get(0).([]*Unit), args.Error(1)
}

func (m *ExecutorMock) QueryUnit(unitName string, verify bool) (*Unit, error) {
	args := m.Called(unitName, verify)
	return args.Get(0).(*Unit), args.Error(1)
}

func (m *ExecutorMock) StartUnit(u *Unit) error {
	args := m.Called(u)
	return args.Error(0)
}

func (m *ExecutorMock) StopUnit(u *Unit) error {
	args := m.Called(u)
	return args.Error(0)
}

func (m *ExecutorMock) RestartUnit(u *Unit) error {
	args := m.Called(u)
	return args.Error(0)
}

func (m *ExecutorMock) ReloadUnit(u *Unit) error {
	args := m.Called(u)
	return args.Error(0)
}

func (m *ExecutorMock) SendSignal(u *Unit, signal Signal) {
	m.Called(u, signal)
	return
}
