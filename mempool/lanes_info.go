package mempool

import (
	"fmt"
	"slices"

	"github.com/cometbft/cometbft/types"
)

type LanesInfo struct {
	lanes       []types.Lane
	defaultLane types.Lane
}

type ErrEmptyLaneDefaultPrioSet struct {
	Info LanesInfo
}

func (e ErrEmptyLaneDefaultPrioSet) Error() string {
	return fmt.Sprintf("invalid lane info:if list of lanes is empty, then defaultLane should be 0, but %v given; info %v", e.Info.defaultLane, e.Info)
}

type ErrBadDefaultPrioNonEmptyLaneList struct {
	Info LanesInfo
}

func (e ErrBadDefaultPrioNonEmptyLaneList) Error() string {
	return fmt.Sprintf("invalid lane info:default lane cannot be 0 if list of lanes is non empty; info: %v", e.Info)
}

type ErrDefaultLaneNotInList struct {
	Info LanesInfo
}

func (e ErrDefaultLaneNotInList) Error() string {
	return fmt.Sprintf("invalid lane info:list of lanes does not contain default lane; info %v", e.Info)
}

type ErrRepeatedPriorities struct {
	Info LanesInfo
}

func (e ErrRepeatedPriorities) Error() string {
	return fmt.Sprintf("list of lanes cannot have repeated values; info %v", e.Info)
}

// Query app info to return the required information to initialize lanes.
func FetchLanesInfo(lanePriorities []uint32, defLane types.Lane) (*LanesInfo, error) {
	lanes := make([]types.Lane, len(lanePriorities))
	for i, l := range lanePriorities {
		lanes[i] = types.Lane(l)
	}
	info := LanesInfo{lanes: lanes, defaultLane: defLane}
	if err := info.validate(); err != nil {
		return nil, err
	}

	return &info, nil
}

func (info *LanesInfo) validate() error {
	// If no lanes are provided the default priority is 0
	if len(info.lanes) == 0 && info.defaultLane == 0 {
		return nil
	}
	// Lane 0 is reserved for when there are no lanes or for invalid txs; it should not be used for the default lane.
	if len(info.lanes) == 0 && info.defaultLane != 0 {
		return ErrEmptyLaneDefaultPrioSet{
			Info: *info,
		}
	}
	if info.defaultLane == 0 && len(info.lanes) != 0 {
		return ErrBadDefaultPrioNonEmptyLaneList{
			Info: *info,
		}
	}
	if !slices.Contains(info.lanes, info.defaultLane) {
		return ErrDefaultLaneNotInList{
			Info: *info,
		}
	}
	lanesSet := make(map[types.Lane]struct{})
	for _, lane := range info.lanes {
		lanesSet[lane] = struct{}{}
	}
	if len(info.lanes) != len(lanesSet) {
		return ErrRepeatedPriorities{
			Info: *info,
		}
	}
	return nil
}
