package mempool

import (
	v1 "github.com/cometbft/cometbft/api/cometbft/abci/v1"
	"github.com/cometbft/cometbft/types"
)

type LaneData struct {
	lanes       map[string]uint32
	defaultLane v1.Lane
}

// BuildLanesInfo builds the information required to initialize
// lanes given the data queried from the app.
func BuildLanesInfo(laneList map[string]uint32, defLane v1.Lane) (*LaneData, error) {
	// lanes := make([]types.LaneID, len(laneList))
	// for i, l := range laneList {
	// 	lanes[i] = types.LaneID(l)
	// }
	info := LaneData{lanes: laneList, defaultLane: defLane}
	if err := validate(info); err != nil {
		return nil, err
	}

	return &info, nil
}

func validate(info LaneData) error {
	// If no lanes are provided the default priority is 0
	if len(info.lanes) == 0 && info.defaultLane.Prio == 0 && info.defaultLane.Id == "" {
		info.defaultLane.Id = "default"
		return nil
	}

	// Default lane is set but empty lane list
	if len(info.lanes) == 0 && info.defaultLane.Prio != 0 {
		return ErrEmptyLanesDefaultLaneSet{
			Info: info,
		}
	}

	// Lane 0 is reserved for when there are no lanes or for invalid txs; it should not be used for the default lane.
	if info.defaultLane.Prio == 0 && len(info.lanes) != 0 {
		return ErrBadDefaultLaneNonEmptyLaneList{
			Info: info,
		}
	}

	found := false
	for laneID, lanePrio := range info.lanes {
		if laneID == info.defaultLane.Id && lanePrio == info.defaultLane.Prio {
			found = true
			break
		}
	}

	// The default lane is not contained in the list of lanes
	if !found { ///slices.Contains(info.lanes, &info.defaultLane) {
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
