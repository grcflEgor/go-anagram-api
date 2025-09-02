package mocks

import "github.com/stretchr/testify/mock"

type MockTaskStats struct {
	mock.Mock
}

func (m *MockTaskStats) IncrementTotalTasks() {
	m.Called()
}

func (m *MockTaskStats) IncrementCompletedTasks() {
	m.Called()
}

func (m *MockTaskStats) IncrementFailedTasks() {
	m.Called()
}

func (m *MockTaskStats) Get() map[string]uint64 {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(map[string]uint64)
}

var _ interface {
	IncrementTotalTasks()
	IncrementCompletedTasks()
	IncrementFailedTasks()
	Get() map[string]uint64
} = (*MockTaskStats)(nil)