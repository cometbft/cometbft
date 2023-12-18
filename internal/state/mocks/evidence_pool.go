// Code generated by mockery. DO NOT EDIT.

package mocks

import (
	state "github.com/cometbft/cometbft/internal/state"
	mock "github.com/stretchr/testify/mock"

	types "github.com/cometbft/cometbft/types"
)

// EvidencePool is an autogenerated mock type for the EvidencePool type
type EvidencePool struct {
	mock.Mock
}

// AddEvidence provides a mock function with given fields: ev
func (_m *EvidencePool) AddEvidence(ev types.Evidence) error {
	ret := _m.Called(ev)

	if len(ret) == 0 {
		panic("no return value specified for AddEvidence")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(types.Evidence) error); ok {
		r0 = rf(ev)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// CheckEvidence provides a mock function with given fields: evList
func (_m *EvidencePool) CheckEvidence(evList types.EvidenceList) error {
	ret := _m.Called(evList)

	if len(ret) == 0 {
		panic("no return value specified for CheckEvidence")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(types.EvidenceList) error); ok {
		r0 = rf(evList)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// PendingEvidence provides a mock function with given fields: maxBytes
func (_m *EvidencePool) PendingEvidence(maxBytes int64) ([]types.Evidence, int64) {
	ret := _m.Called(maxBytes)

	if len(ret) == 0 {
		panic("no return value specified for PendingEvidence")
	}

	var r0 []types.Evidence
	var r1 int64
	if rf, ok := ret.Get(0).(func(int64) ([]types.Evidence, int64)); ok {
		return rf(maxBytes)
	}
	if rf, ok := ret.Get(0).(func(int64) []types.Evidence); ok {
		r0 = rf(maxBytes)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]types.Evidence)
		}
	}

	if rf, ok := ret.Get(1).(func(int64) int64); ok {
		r1 = rf(maxBytes)
	} else {
		r1 = ret.Get(1).(int64)
	}

	return r0, r1
}

// Update provides a mock function with given fields: _a0, evList
func (_m *EvidencePool) Update(_a0 state.State, evList types.EvidenceList) {
	_m.Called(_a0, evList)
}

// NewEvidencePool creates a new instance of EvidencePool. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewEvidencePool(t interface {
	mock.TestingT
	Cleanup(func())
},
) *EvidencePool {
	mock := &EvidencePool{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
