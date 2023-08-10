// Code generated by mockery. DO NOT EDIT.

package mocks

import (
	abcitypes "github.com/cometbft/cometbft/abci/types"
	mock "github.com/stretchr/testify/mock"

	state "github.com/cometbft/cometbft/state"

	types "github.com/cometbft/cometbft/types"
)

// Store is an autogenerated mock type for the Store type
type Store struct {
	mock.Mock
}

// Bootstrap provides a mock function with given fields: _a0
func (_m *Store) Bootstrap(_a0 state.State) error {
	ret := _m.Called(_a0)

	var r0 error
	if rf, ok := ret.Get(0).(func(state.State) error); ok {
		r0 = rf(_a0)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Close provides a mock function with given fields:
func (_m *Store) Close() error {
	ret := _m.Called()

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// GetABCIResRetainHeight provides a mock function with given fields:
func (_m *Store) GetABCIResRetainHeight() (int64, error) {
	ret := _m.Called()

	var r0 int64
	var r1 error
	if rf, ok := ret.Get(0).(func() (int64, error)); ok {
		return rf()
	}
	if rf, ok := ret.Get(0).(func() int64); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(int64)
	}

	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetApplicationRetainHeight provides a mock function with given fields:
func (_m *Store) GetApplicationRetainHeight() (int64, error) {
	ret := _m.Called()

	var r0 int64
	var r1 error
	if rf, ok := ret.Get(0).(func() (int64, error)); ok {
		return rf()
	}
	if rf, ok := ret.Get(0).(func() int64); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(int64)
	}

	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetCompanionBlockRetainHeight provides a mock function with given fields:
func (_m *Store) GetCompanionBlockRetainHeight() (int64, error) {
	ret := _m.Called()

	var r0 int64
	var r1 error
	if rf, ok := ret.Get(0).(func() (int64, error)); ok {
		return rf()
	}
	if rf, ok := ret.Get(0).(func() int64); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(int64)
	}

	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Load provides a mock function with given fields:
func (_m *Store) Load() (state.State, error) {
	ret := _m.Called()

	var r0 state.State
	var r1 error
	if rf, ok := ret.Get(0).(func() (state.State, error)); ok {
		return rf()
	}
	if rf, ok := ret.Get(0).(func() state.State); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(state.State)
	}

	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// LoadConsensusParams provides a mock function with given fields: height
func (_m *Store) LoadConsensusParams(height int64) (types.ConsensusParams, error) {
	ret := _m.Called(height)

	var r0 types.ConsensusParams
	var r1 error
	if rf, ok := ret.Get(0).(func(int64) (types.ConsensusParams, error)); ok {
		return rf(height)
	}
	if rf, ok := ret.Get(0).(func(int64) types.ConsensusParams); ok {
		r0 = rf(height)
	} else {
		r0 = ret.Get(0).(types.ConsensusParams)
	}

	if rf, ok := ret.Get(1).(func(int64) error); ok {
		r1 = rf(height)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// LoadFinalizeBlockResponse provides a mock function with given fields: height
func (_m *Store) LoadFinalizeBlockResponse(height int64) (*abcitypes.ResponseFinalizeBlock, error) {
	ret := _m.Called(height)

	var r0 *abcitypes.ResponseFinalizeBlock
	var r1 error
	if rf, ok := ret.Get(0).(func(int64) (*abcitypes.ResponseFinalizeBlock, error)); ok {
		return rf(height)
	}
	if rf, ok := ret.Get(0).(func(int64) *abcitypes.ResponseFinalizeBlock); ok {
		r0 = rf(height)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*abcitypes.ResponseFinalizeBlock)
		}
	}

	if rf, ok := ret.Get(1).(func(int64) error); ok {
		r1 = rf(height)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// LoadFromDBOrGenesisDoc provides a mock function with given fields: doc
func (_m *Store) LoadFromDBOrGenesisDoc(doc *types.GenesisDoc) (state.State, error) {
	ret := _m.Called(doc)

	var r0 state.State
	var r1 error
	if rf, ok := ret.Get(0).(func(*types.GenesisDoc) (state.State, error)); ok {
		return rf(doc)
	}
	if rf, ok := ret.Get(0).(func(*types.GenesisDoc) state.State); ok {
		r0 = rf(doc)
	} else {
		r0 = ret.Get(0).(state.State)
	}

	if rf, ok := ret.Get(1).(func(*types.GenesisDoc) error); ok {
		r1 = rf(doc)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// LoadFromDBOrGenesisFile provides a mock function with given fields: filename
func (_m *Store) LoadFromDBOrGenesisFile(filename string) (state.State, error) {
	ret := _m.Called(filename)

	var r0 state.State
	var r1 error
	if rf, ok := ret.Get(0).(func(string) (state.State, error)); ok {
		return rf(filename)
	}
	if rf, ok := ret.Get(0).(func(string) state.State); ok {
		r0 = rf(filename)
	} else {
		r0 = ret.Get(0).(state.State)
	}

	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(filename)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// LoadLastFinalizeBlockResponse provides a mock function with given fields: height
func (_m *Store) LoadLastFinalizeBlockResponse(height int64) (*abcitypes.ResponseFinalizeBlock, error) {
	ret := _m.Called(height)

	var r0 *abcitypes.ResponseFinalizeBlock
	var r1 error
	if rf, ok := ret.Get(0).(func(int64) (*abcitypes.ResponseFinalizeBlock, error)); ok {
		return rf(height)
	}
	if rf, ok := ret.Get(0).(func(int64) *abcitypes.ResponseFinalizeBlock); ok {
		r0 = rf(height)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*abcitypes.ResponseFinalizeBlock)
		}
	}

	if rf, ok := ret.Get(1).(func(int64) error); ok {
		r1 = rf(height)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// LoadValidators provides a mock function with given fields: height
func (_m *Store) LoadValidators(height int64) (*types.ValidatorSet, error) {
	ret := _m.Called(height)

	var r0 *types.ValidatorSet
	var r1 error
	if rf, ok := ret.Get(0).(func(int64) (*types.ValidatorSet, error)); ok {
		return rf(height)
	}
	if rf, ok := ret.Get(0).(func(int64) *types.ValidatorSet); ok {
		r0 = rf(height)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*types.ValidatorSet)
		}
	}

	if rf, ok := ret.Get(1).(func(int64) error); ok {
		r1 = rf(height)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// PruneABCIResponses provides a mock function with given fields: targetRetainHeight
func (_m *Store) PruneABCIResponses(targetRetainHeight int64) (int64, int64, error) {
	ret := _m.Called(targetRetainHeight)

	var r0 int64
	var r1 int64
	var r2 error
	if rf, ok := ret.Get(0).(func(int64) (int64, int64, error)); ok {
		return rf(targetRetainHeight)
	}
	if rf, ok := ret.Get(0).(func(int64) int64); ok {
		r0 = rf(targetRetainHeight)
	} else {
		r0 = ret.Get(0).(int64)
	}

	if rf, ok := ret.Get(1).(func(int64) int64); ok {
		r1 = rf(targetRetainHeight)
	} else {
		r1 = ret.Get(1).(int64)
	}

	if rf, ok := ret.Get(2).(func(int64) error); ok {
		r2 = rf(targetRetainHeight)
	} else {
		r2 = ret.Error(2)
	}

	return r0, r1, r2
}

// PruneStates provides a mock function with given fields: fromHeight, toHeight, evidenceThresholdHeight
func (_m *Store) PruneStates(fromHeight int64, toHeight int64, evidenceThresholdHeight int64) error {
	ret := _m.Called(fromHeight, toHeight, evidenceThresholdHeight)

	var r0 error
	if rf, ok := ret.Get(0).(func(int64, int64, int64) error); ok {
		r0 = rf(fromHeight, toHeight, evidenceThresholdHeight)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Save provides a mock function with given fields: _a0
func (_m *Store) Save(_a0 state.State) error {
	ret := _m.Called(_a0)

	var r0 error
	if rf, ok := ret.Get(0).(func(state.State) error); ok {
		r0 = rf(_a0)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// SaveABCIResRetainHeight provides a mock function with given fields: height
func (_m *Store) SaveABCIResRetainHeight(height int64) error {
	ret := _m.Called(height)

	var r0 error
	if rf, ok := ret.Get(0).(func(int64) error); ok {
		r0 = rf(height)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// SaveApplicationRetainHeight provides a mock function with given fields: height
func (_m *Store) SaveApplicationRetainHeight(height int64) error {
	ret := _m.Called(height)

	var r0 error
	if rf, ok := ret.Get(0).(func(int64) error); ok {
		r0 = rf(height)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// SaveCompanionBlockRetainHeight provides a mock function with given fields: height
func (_m *Store) SaveCompanionBlockRetainHeight(height int64) error {
	ret := _m.Called(height)

	var r0 error
	if rf, ok := ret.Get(0).(func(int64) error); ok {
		r0 = rf(height)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// SaveFinalizeBlockResponse provides a mock function with given fields: height, res
func (_m *Store) SaveFinalizeBlockResponse(height int64, res *abcitypes.ResponseFinalizeBlock) error {
	ret := _m.Called(height, res)

	var r0 error
	if rf, ok := ret.Get(0).(func(int64, *abcitypes.ResponseFinalizeBlock) error); ok {
		r0 = rf(height, res)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// NewStore creates a new instance of Store. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewStore(t interface {
	mock.TestingT
	Cleanup(func())
}) *Store {
	mock := &Store{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
