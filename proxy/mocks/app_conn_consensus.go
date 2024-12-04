// Code generated by mockery. DO NOT EDIT.

package mocks

import (
	mock "github.com/stretchr/testify/mock"
	abcicli "github.com/tendermint/tendermint/abci/client"

	types "github.com/tendermint/tendermint/abci/types"
)

// AppConnConsensus is an autogenerated mock type for the AppConnConsensus type
type AppConnConsensus struct {
	mock.Mock
}

// BeginBlockSync provides a mock function with given fields: _a0
func (_m *AppConnConsensus) BeginBlockSync(_a0 types.RequestBeginBlock) (*types.ResponseBeginBlock, error) {
	ret := _m.Called(_a0)

	if len(ret) == 0 {
		panic("no return value specified for BeginBlockSync")
	}

	var r0 *types.ResponseBeginBlock
	var r1 error
	if rf, ok := ret.Get(0).(func(types.RequestBeginBlock) (*types.ResponseBeginBlock, error)); ok {
		return rf(_a0)
	}
	if rf, ok := ret.Get(0).(func(types.RequestBeginBlock) *types.ResponseBeginBlock); ok {
		r0 = rf(_a0)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*types.ResponseBeginBlock)
		}
	}

	if rf, ok := ret.Get(1).(func(types.RequestBeginBlock) error); ok {
		r1 = rf(_a0)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// CommitSync provides a mock function with no fields
func (_m *AppConnConsensus) CommitSync() (*types.ResponseCommit, error) {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for CommitSync")
	}

	var r0 *types.ResponseCommit
	var r1 error
	if rf, ok := ret.Get(0).(func() (*types.ResponseCommit, error)); ok {
		return rf()
	}
	if rf, ok := ret.Get(0).(func() *types.ResponseCommit); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*types.ResponseCommit)
		}
	}

	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// DeliverTxAsync provides a mock function with given fields: _a0
func (_m *AppConnConsensus) DeliverTxAsync(_a0 types.RequestDeliverTx) *abcicli.ReqRes {
	ret := _m.Called(_a0)

	if len(ret) == 0 {
		panic("no return value specified for DeliverTxAsync")
	}

	var r0 *abcicli.ReqRes
	if rf, ok := ret.Get(0).(func(types.RequestDeliverTx) *abcicli.ReqRes); ok {
		r0 = rf(_a0)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*abcicli.ReqRes)
		}
	}

	return r0
}

// EndBlockSync provides a mock function with given fields: _a0
func (_m *AppConnConsensus) EndBlockSync(_a0 types.RequestEndBlock) (*types.ResponseEndBlock, error) {
	ret := _m.Called(_a0)

	if len(ret) == 0 {
		panic("no return value specified for EndBlockSync")
	}

	var r0 *types.ResponseEndBlock
	var r1 error
	if rf, ok := ret.Get(0).(func(types.RequestEndBlock) (*types.ResponseEndBlock, error)); ok {
		return rf(_a0)
	}
	if rf, ok := ret.Get(0).(func(types.RequestEndBlock) *types.ResponseEndBlock); ok {
		r0 = rf(_a0)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*types.ResponseEndBlock)
		}
	}

	if rf, ok := ret.Get(1).(func(types.RequestEndBlock) error); ok {
		r1 = rf(_a0)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Error provides a mock function with no fields
func (_m *AppConnConsensus) Error() error {
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

// InitChainSync provides a mock function with given fields: _a0
func (_m *AppConnConsensus) InitChainSync(_a0 types.RequestInitChain) (*types.ResponseInitChain, error) {
	ret := _m.Called(_a0)

	if len(ret) == 0 {
		panic("no return value specified for InitChainSync")
	}

	var r0 *types.ResponseInitChain
	var r1 error
	if rf, ok := ret.Get(0).(func(types.RequestInitChain) (*types.ResponseInitChain, error)); ok {
		return rf(_a0)
	}
	if rf, ok := ret.Get(0).(func(types.RequestInitChain) *types.ResponseInitChain); ok {
		r0 = rf(_a0)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*types.ResponseInitChain)
		}
	}

	if rf, ok := ret.Get(1).(func(types.RequestInitChain) error); ok {
		r1 = rf(_a0)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// SetResponseCallback provides a mock function with given fields: _a0
func (_m *AppConnConsensus) SetResponseCallback(_a0 abcicli.Callback) {
	_m.Called(_a0)
}

// NewAppConnConsensus creates a new instance of AppConnConsensus. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewAppConnConsensus(t interface {
	mock.TestingT
	Cleanup(func())
}) *AppConnConsensus {
	mock := &AppConnConsensus{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
