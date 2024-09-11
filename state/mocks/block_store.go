// Code generated by mockery. DO NOT EDIT.

package mocks

import (
	state "github.com/cometbft/cometbft/state"
	mock "github.com/stretchr/testify/mock"

	types "github.com/cometbft/cometbft/types"
)

// BlockStore is an autogenerated mock type for the BlockStore type
type BlockStore struct {
	mock.Mock
}

// Base provides a mock function with given fields:
func (_m *BlockStore) Base() int64 {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for Base")
	}

	var r0 int64
	if rf, ok := ret.Get(0).(func() int64); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(int64)
	}

	return r0
}

// Close provides a mock function with given fields:
func (_m *BlockStore) Close() error {
	ret := _m.Called()

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// DeleteLatestBlock provides a mock function with given fields:
func (_m *BlockStore) DeleteLatestBlock() error {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for DeleteLatestBlock")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Height provides a mock function with given fields:
func (_m *BlockStore) Height() int64 {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for Height")
	}

	var r0 int64
	if rf, ok := ret.Get(0).(func() int64); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(int64)
	}

	return r0
}

// LoadBaseMeta provides a mock function with given fields:
func (_m *BlockStore) LoadBaseMeta() *types.BlockMeta {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for LoadBaseMeta")
	}

	var r0 *types.BlockMeta
	if rf, ok := ret.Get(0).(func() *types.BlockMeta); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*types.BlockMeta)
		}
	}

	return r0
}

// LoadBlock provides a mock function with given fields: height
func (_m *BlockStore) LoadBlock(height int64) *types.Block {
	ret := _m.Called(height)

	if len(ret) == 0 {
		panic("no return value specified for LoadBlock")
	}

	var r0 *types.Block
	if rf, ok := ret.Get(0).(func(int64) *types.Block); ok {
		r0 = rf(height)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*types.Block)
		}
	}

	return r0
}

// LoadBlockByHash provides a mock function with given fields: hash
func (_m *BlockStore) LoadBlockByHash(hash []byte) *types.Block {
	ret := _m.Called(hash)

	if len(ret) == 0 {
		panic("no return value specified for LoadBlockByHash")
	}

	var r0 *types.Block
	if rf, ok := ret.Get(0).(func([]byte) *types.Block); ok {
		r0 = rf(hash)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*types.Block)
		}
	}

	return r0
}

// LoadBlockCommit provides a mock function with given fields: height
func (_m *BlockStore) LoadBlockCommit(height int64) *types.Commit {
	ret := _m.Called(height)

	if len(ret) == 0 {
		panic("no return value specified for LoadBlockCommit")
	}

	var r0 *types.Commit
	if rf, ok := ret.Get(0).(func(int64) *types.Commit); ok {
		r0 = rf(height)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*types.Commit)
		}
	}

	return r0
}

// LoadBlockExtendedCommit provides a mock function with given fields: height
func (_m *BlockStore) LoadBlockExtendedCommit(height int64) *types.ExtendedCommit {
	ret := _m.Called(height)

	var r0 *types.ExtendedCommit
	if rf, ok := ret.Get(0).(func(int64) *types.ExtendedCommit); ok {
		r0 = rf(height)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*types.ExtendedCommit)
		}
	}

	return r0
}

// LoadBlockMeta provides a mock function with given fields: height
func (_m *BlockStore) LoadBlockMeta(height int64) *types.BlockMeta {
	ret := _m.Called(height)

	if len(ret) == 0 {
		panic("no return value specified for LoadBlockMeta")
	}

	var r0 *types.BlockMeta
	if rf, ok := ret.Get(0).(func(int64) *types.BlockMeta); ok {
		r0 = rf(height)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*types.BlockMeta)
		}
	}

	return r0
}

// LoadBlockMetaByHash provides a mock function with given fields: hash
func (_m *BlockStore) LoadBlockMetaByHash(hash []byte) *types.BlockMeta {
	ret := _m.Called(hash)

	if len(ret) == 0 {
		panic("no return value specified for LoadBlockMetaByHash")
	}

	var r0 *types.BlockMeta
	if rf, ok := ret.Get(0).(func([]byte) *types.BlockMeta); ok {
		r0 = rf(hash)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*types.BlockMeta)
		}
	}

	return r0
}

// LoadBlockPart provides a mock function with given fields: height, index
func (_m *BlockStore) LoadBlockPart(height int64, index int) *types.Part {
	ret := _m.Called(height, index)

	if len(ret) == 0 {
		panic("no return value specified for LoadBlockPart")
	}

	var r0 *types.Part
	if rf, ok := ret.Get(0).(func(int64, int) *types.Part); ok {
		r0 = rf(height, index)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*types.Part)
		}
	}

	return r0
}

// LoadSeenCommit provides a mock function with given fields: height
func (_m *BlockStore) LoadSeenCommit(height int64) *types.Commit {
	ret := _m.Called(height)

	if len(ret) == 0 {
		panic("no return value specified for LoadSeenCommit")
	}

	var r0 *types.Commit
	if rf, ok := ret.Get(0).(func(int64) *types.Commit); ok {
		r0 = rf(height)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*types.Commit)
		}
	}

	return r0
}

// PruneBlocks provides a mock function with given fields: height, _a1
func (_m *BlockStore) PruneBlocks(height int64, _a1 state.State) (uint64, int64, error) {
	ret := _m.Called(height, _a1)

	if len(ret) == 0 {
		panic("no return value specified for PruneBlocks")
	}

	var r0 uint64
<<<<<<< HEAD
	var r1 int64
	var r2 error
	if rf, ok := ret.Get(0).(func(int64, state.State) (uint64, int64, error)); ok {
		return rf(height, _a1)
	}
	if rf, ok := ret.Get(0).(func(int64, state.State) uint64); ok {
		r0 = rf(height, _a1)
=======
	var r1 error
	if rf, ok := ret.Get(0).(func(int64) (uint64, error)); ok {
		return rf(height)
	}
	if rf, ok := ret.Get(0).(func(int64) uint64); ok {
		r0 = rf(height)
>>>>>>> 3215ee16a (build(deps): Bump Go to 1.22 (backport #4059) (#4072))
	} else {
		r0 = ret.Get(0).(uint64)
	}

<<<<<<< HEAD
	if rf, ok := ret.Get(1).(func(int64, state.State) int64); ok {
		r1 = rf(height, _a1)
=======
	if rf, ok := ret.Get(1).(func(int64) error); ok {
		r1 = rf(height)
>>>>>>> 3215ee16a (build(deps): Bump Go to 1.22 (backport #4059) (#4072))
	} else {
		r1 = ret.Get(1).(int64)
	}

	if rf, ok := ret.Get(2).(func(int64, state.State) error); ok {
		r2 = rf(height, _a1)
	} else {
		r2 = ret.Error(2)
	}

	return r0, r1, r2
}

// SaveBlock provides a mock function with given fields: block, blockParts, seenCommit
func (_m *BlockStore) SaveBlock(block *types.Block, blockParts *types.PartSet, seenCommit *types.Commit) {
	_m.Called(block, blockParts, seenCommit)
}

// SaveBlockWithExtendedCommit provides a mock function with given fields: block, blockParts, seenCommit
func (_m *BlockStore) SaveBlockWithExtendedCommit(block *types.Block, blockParts *types.PartSet, seenCommit *types.ExtendedCommit) {
	_m.Called(block, blockParts, seenCommit)
}

// Size provides a mock function with given fields:
func (_m *BlockStore) Size() int64 {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for Size")
	}

	var r0 int64
	if rf, ok := ret.Get(0).(func() int64); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(int64)
	}

	return r0
}

// NewBlockStore creates a new instance of BlockStore. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewBlockStore(t interface {
	mock.TestingT
	Cleanup(func())
}) *BlockStore {
	mock := &BlockStore{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
