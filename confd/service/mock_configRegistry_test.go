// Code generated by mockery v2.42.1. DO NOT EDIT.

package service

import (
	mock "github.com/stretchr/testify/mock"
	client "go.etcd.io/etcd/client/v2"
)

// mockConfigRegistry is an autogenerated mock type for the configRegistry type
type mockConfigRegistry struct {
	mock.Mock
}

type mockConfigRegistry_Expecter struct {
	mock *mock.Mock
}

func (_m *mockConfigRegistry) EXPECT() *mockConfigRegistry_Expecter {
	return &mockConfigRegistry_Expecter{mock: &_m.Mock}
}

// Get provides a mock function with given fields: key
func (_m *mockConfigRegistry) Get(key string) (*client.Response, error) {
	ret := _m.Called(key)

	if len(ret) == 0 {
		panic("no return value specified for Get")
	}

	var r0 *client.Response
	var r1 error
	if rf, ok := ret.Get(0).(func(string) (*client.Response, error)); ok {
		return rf(key)
	}
	if rf, ok := ret.Get(0).(func(string) *client.Response); ok {
		r0 = rf(key)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*client.Response)
		}
	}

	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(key)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// mockConfigRegistry_Get_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Get'
type mockConfigRegistry_Get_Call struct {
	*mock.Call
}

// Get is a helper method to define mock.On call
//   - key string
func (_e *mockConfigRegistry_Expecter) Get(key interface{}) *mockConfigRegistry_Get_Call {
	return &mockConfigRegistry_Get_Call{Call: _e.mock.On("Get", key)}
}

func (_c *mockConfigRegistry_Get_Call) Run(run func(key string)) *mockConfigRegistry_Get_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string))
	})
	return _c
}

func (_c *mockConfigRegistry_Get_Call) Return(_a0 *client.Response, _a1 error) *mockConfigRegistry_Get_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *mockConfigRegistry_Get_Call) RunAndReturn(run func(string) (*client.Response, error)) *mockConfigRegistry_Get_Call {
	_c.Call.Return(run)
	return _c
}

// Watch provides a mock function with given fields: key, recursive, eventChannel
func (_m *mockConfigRegistry) Watch(key string, recursive bool, eventChannel chan *client.Response) {
	_m.Called(key, recursive, eventChannel)
}

// mockConfigRegistry_Watch_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Watch'
type mockConfigRegistry_Watch_Call struct {
	*mock.Call
}

// Watch is a helper method to define mock.On call
//   - key string
//   - recursive bool
//   - eventChannel chan *client.Response
func (_e *mockConfigRegistry_Expecter) Watch(key interface{}, recursive interface{}, eventChannel interface{}) *mockConfigRegistry_Watch_Call {
	return &mockConfigRegistry_Watch_Call{Call: _e.mock.On("Watch", key, recursive, eventChannel)}
}

func (_c *mockConfigRegistry_Watch_Call) Run(run func(key string, recursive bool, eventChannel chan *client.Response)) *mockConfigRegistry_Watch_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string), args[1].(bool), args[2].(chan *client.Response))
	})
	return _c
}

func (_c *mockConfigRegistry_Watch_Call) Return() *mockConfigRegistry_Watch_Call {
	_c.Call.Return()
	return _c
}

func (_c *mockConfigRegistry_Watch_Call) RunAndReturn(run func(string, bool, chan *client.Response)) *mockConfigRegistry_Watch_Call {
	_c.Call.Return(run)
	return _c
}

// newMockConfigRegistry creates a new instance of mockConfigRegistry. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func newMockConfigRegistry(t interface {
	mock.TestingT
	Cleanup(func())
}) *mockConfigRegistry {
	mock := &mockConfigRegistry{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}