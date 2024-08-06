package mempool

import (
	"slices"

	"github.com/cometbft/cometbft/types"
)

type LanesInfo struct {
	lanes       []types.Lane
	defaultLane types.Lane
}

// Query app info to return the required information to initialize lanes.
func FetchLanesInfo(laneList []uint32, defLane types.Lane) (*LanesInfo, error) {
	lanes := make([]types.Lane, len(laneList))
	for i, l := range laneList {
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
		return ErrEmptyLanesDefaultLaneSet{
			Info: *info,
		}
	}
	if info.defaultLane == 0 && len(info.lanes) != 0 {
		return ErrBadDefaultLaneNonEmptyLaneList{
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
		return ErrRepeatedLanes{
			Info: *info,
		}
	}
	return nil
}
