// Code generated by mockery. DO NOT EDIT.

package mocks

import (
	context "context"

	mock "github.com/stretchr/testify/mock"

	types "github.com/cometbft/cometbft/abci/types"
)

// AppConnQuery is an autogenerated mock type for the AppConnQuery type
type AppConnQuery struct {
	mock.Mock
}

// Echo provides a mock function with given fields: _a0, _a1
func (_m *AppConnQuery) Echo(_a0 context.Context, _a1 string) (*types.ResponseEcho, error) {
	ret := _m.Called(_a0, _a1)

	if len(ret) == 0 {
		panic("no return value specified for EchoSync")
	}

	var r0 *types.ResponseEcho
	var r1 error
<<<<<<< HEAD
	if rf, ok := ret.Get(0).(func(context.Context, string) (*types.ResponseEcho, error)); ok {
		return rf(_a0, _a1)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) *types.ResponseEcho); ok {
		r0 = rf(_a0, _a1)
=======
	if rf, ok := ret.Get(0).(func(string) (*types.ResponseEcho, error)); ok {
		return rf(_a0)
	}
	if rf, ok := ret.Get(0).(func(string) *types.ResponseEcho); ok {
		r0 = rf(_a0)
>>>>>>> 3215ee16a (build(deps): Bump Go to 1.22 (backport #4059) (#4072))
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*types.ResponseEcho)
		}
	}

<<<<<<< HEAD
	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(_a0, _a1)
=======
	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(_a0)
>>>>>>> 3215ee16a (build(deps): Bump Go to 1.22 (backport #4059) (#4072))
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Error provides a mock function with given fields:
func (_m *AppConnQuery) Error() error {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for Error")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Info provides a mock function with given fields: _a0, _a1
func (_m *AppConnQuery) Info(_a0 context.Context, _a1 *types.RequestInfo) (*types.ResponseInfo, error) {
	ret := _m.Called(_a0, _a1)

	if len(ret) == 0 {
		panic("no return value specified for InfoSync")
	}

	var r0 *types.ResponseInfo
	var r1 error
<<<<<<< HEAD
	if rf, ok := ret.Get(0).(func(context.Context, *types.RequestInfo) (*types.ResponseInfo, error)); ok {
		return rf(_a0, _a1)
	}
	if rf, ok := ret.Get(0).(func(context.Context, *types.RequestInfo) *types.ResponseInfo); ok {
		r0 = rf(_a0, _a1)
=======
	if rf, ok := ret.Get(0).(func(types.RequestInfo) (*types.ResponseInfo, error)); ok {
		return rf(_a0)
	}
	if rf, ok := ret.Get(0).(func(types.RequestInfo) *types.ResponseInfo); ok {
		r0 = rf(_a0)
>>>>>>> 3215ee16a (build(deps): Bump Go to 1.22 (backport #4059) (#4072))
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*types.ResponseInfo)
		}
	}

<<<<<<< HEAD
	if rf, ok := ret.Get(1).(func(context.Context, *types.RequestInfo) error); ok {
		r1 = rf(_a0, _a1)
=======
	if rf, ok := ret.Get(1).(func(types.RequestInfo) error); ok {
		r1 = rf(_a0)
>>>>>>> 3215ee16a (build(deps): Bump Go to 1.22 (backport #4059) (#4072))
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Query provides a mock function with given fields: _a0, _a1
func (_m *AppConnQuery) Query(_a0 context.Context, _a1 *types.RequestQuery) (*types.ResponseQuery, error) {
	ret := _m.Called(_a0, _a1)

	if len(ret) == 0 {
		panic("no return value specified for QuerySync")
	}

	var r0 *types.ResponseQuery
	var r1 error
<<<<<<< HEAD
	if rf, ok := ret.Get(0).(func(context.Context, *types.RequestQuery) (*types.ResponseQuery, error)); ok {
		return rf(_a0, _a1)
	}
	if rf, ok := ret.Get(0).(func(context.Context, *types.RequestQuery) *types.ResponseQuery); ok {
		r0 = rf(_a0, _a1)
=======
	if rf, ok := ret.Get(0).(func(types.RequestQuery) (*types.ResponseQuery, error)); ok {
		return rf(_a0)
	}
	if rf, ok := ret.Get(0).(func(types.RequestQuery) *types.ResponseQuery); ok {
		r0 = rf(_a0)
>>>>>>> 3215ee16a (build(deps): Bump Go to 1.22 (backport #4059) (#4072))
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*types.ResponseQuery)
		}
	}

<<<<<<< HEAD
	if rf, ok := ret.Get(1).(func(context.Context, *types.RequestQuery) error); ok {
		r1 = rf(_a0, _a1)
=======
	if rf, ok := ret.Get(1).(func(types.RequestQuery) error); ok {
		r1 = rf(_a0)
>>>>>>> 3215ee16a (build(deps): Bump Go to 1.22 (backport #4059) (#4072))
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// NewAppConnQuery creates a new instance of AppConnQuery. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewAppConnQuery(t interface {
	mock.TestingT
	Cleanup(func())
}) *AppConnQuery {
	mock := &AppConnQuery{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
