// Code generated by mockery v2.43.2. DO NOT EDIT.

package mocks

import (
	context "context"

	domain "github.com/flapenna/go-ddd-crud/internal/domain/user"
	mock "github.com/stretchr/testify/mock"
)

// MockUserWatcher is an autogenerated mock type for the UserWatcher type
type MockUserWatcher struct {
	mock.Mock
}

type MockUserWatcher_Expecter struct {
	mock *mock.Mock
}

func (_m *MockUserWatcher) EXPECT() *MockUserWatcher_Expecter {
	return &MockUserWatcher_Expecter{mock: &_m.Mock}
}

// WatchUsers provides a mock function with given fields: ctx
func (_m *MockUserWatcher) WatchUsers(ctx context.Context) <-chan *domain.UserEvent {
	ret := _m.Called(ctx)

	if len(ret) == 0 {
		panic("no return value specified for WatchUsers")
	}

	var r0 <-chan *domain.UserEvent
	if rf, ok := ret.Get(0).(func(context.Context) <-chan *domain.UserEvent); ok {
		r0 = rf(ctx)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(<-chan *domain.UserEvent)
		}
	}

	return r0
}

// MockUserWatcher_WatchUsers_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'WatchUsers'
type MockUserWatcher_WatchUsers_Call struct {
	*mock.Call
}

// WatchUsers is a helper method to define mock.On call
//   - ctx context.Context
func (_e *MockUserWatcher_Expecter) WatchUsers(ctx interface{}) *MockUserWatcher_WatchUsers_Call {
	return &MockUserWatcher_WatchUsers_Call{Call: _e.mock.On("WatchUsers", ctx)}
}

func (_c *MockUserWatcher_WatchUsers_Call) Run(run func(ctx context.Context)) *MockUserWatcher_WatchUsers_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context))
	})
	return _c
}

func (_c *MockUserWatcher_WatchUsers_Call) Return(_a0 <-chan *domain.UserEvent) *MockUserWatcher_WatchUsers_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockUserWatcher_WatchUsers_Call) RunAndReturn(run func(context.Context) <-chan *domain.UserEvent) *MockUserWatcher_WatchUsers_Call {
	_c.Call.Return(run)
	return _c
}

// NewMockUserWatcher creates a new instance of MockUserWatcher. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockUserWatcher(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockUserWatcher {
	mock := &MockUserWatcher{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
