package mempool

import (
	"github.com/cometbft/cometbft/types"
)

type LanesInfo struct {
	lanes       map[string]uint32
	defaultLane string
}

// BuildLanesInfo builds the information required to initialize
// lanes given the data queried from the app.
func BuildLanesInfo(laneMap map[string]uint32, defLane string) (*LanesInfo, error) {
	info := LanesInfo{lanes: laneMap, defaultLane: defLane}
	if err := validate(info); err != nil {
		return nil, err
	}

	return &info, nil
}

func validate(info LanesInfo) error {
	// If no lanes are provided the default priority is 0
	if len(info.lanes) == 0 && info.defaultLane == "" {
		return nil
	}

	// Default lane is set but empty lane list
	if len(info.lanes) == 0 && info.defaultLane != "" {
		return ErrEmptyLanesDefaultLaneSet{
			Info: info,
		}
	}

	// Lane 0 is reserved for when there are no lanes or for invalid txs; it should not be used for the default lane.
	if info.defaultLane == "" && len(info.lanes) != 0 {
		return ErrBadDefaultLaneNonEmptyLaneList{
			Info: info,
		}
	}

	if _, ok := info.lanes[info.defaultLane]; !ok {
		return ErrDefaultLaneNotInList{
			Info: info,
		}
	}

	lanesSet := make(map[types.LaneID]struct{})
	for laneID := range info.lanes {
		lanesSet[types.LaneID(laneID)] = struct{}{}
	}
	if len(info.lanes) != len(lanesSet) {
		return ErrRepeatedLanes{
			Info: info,
		}
	}
	return nil
}
