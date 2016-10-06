package control_test

import "github.com/stretchr/testify/mock"

type MockRenderer struct {
	mock.Mock
}

func (m *MockRenderer) Value() (string, bool) {
	args := m.Called()
	return args.String(0), args.Bool(1)
}

func (m *MockRenderer) Render(name, content string) (string, error) {
	args := m.Called(name, content)
	return args.String(0), args.Error(1)
}

func defaultMockRenderer() *MockRenderer {
	m := &MockRenderer{}
	m.On("Render", mock.Anything, mock.Anything).Return("", nil)
	m.On("Value").Return("", true)
	return m
}
