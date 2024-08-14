package mempool

import (
	"slices"

	"github.com/cometbft/cometbft/types"
)

type LanesInfo struct {
	lanes       []types.Lane
	defaultLane types.Lane
}

// BuildLanesInfo builds the information required to initialize
// lanes given the data queried from the app.
func BuildLanesInfo(laneList []uint32, defLane types.Lane) (*LanesInfo, error) {
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

	// Default lane is set but empty lane list
	if len(info.lanes) == 0 && info.defaultLane != 0 {
		return ErrEmptyLanesDefaultLaneSet{
			Info: *info,
		}
	}

	// Lane 0 is reserved for when there are no lanes or for invalid txs; it should not be used for the default lane.
	if info.defaultLane == 0 && len(info.lanes) != 0 {
		return ErrBadDefaultLaneNonEmptyLaneList{
			Info: *info,
		}
	}

	// The default lane is not contained in the list of lanes
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
