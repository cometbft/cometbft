package mempool

import (
	"context"
	"errors"
	"fmt"
	"slices"

	"github.com/cometbft/cometbft/proxy"
	"github.com/cometbft/cometbft/types"
)

type LanesInfo struct {
	lanes       []types.Lane
	defaultLane types.Lane
}

// Query app info to return the required information to initialize lanes.
func FetchLanesInfo(proxyApp proxy.AppConnQuery) (*LanesInfo, error) {
	res, err := proxyApp.Info(context.TODO(), proxy.InfoRequest)
	if err != nil {
		return nil, fmt.Errorf("error calling Info: %v", err)
	}

	lanes := make([]types.Lane, len(res.LanePriorities))
	for i, l := range res.LanePriorities {
		lanes[i] = types.Lane(l)
	}
	info := LanesInfo{lanes: lanes, defaultLane: types.Lane(res.DefaultLanePriority)}
	if err = info.validate(); err != nil {
		return nil, fmt.Errorf("invalid lane info: %v, info: %v", err, info)
	}

	return &info, nil
}

func (info *LanesInfo) validate() error {
	// Lane 0 is reserved for when there are no lanes or for invalid txs; it should not be used for the default lane.
	if len(info.lanes) == 0 && info.defaultLane != 0 {
		return errors.New("if list of lanes is empty, then defaultLane should be 0")
	}
	if info.defaultLane == 0 && len(info.lanes) == 0 {
		return errors.New("default lane cannot be 0 if list of lanes is non empty")
	}
	if !slices.Contains(info.lanes, info.defaultLane) {
		return errors.New("list of lanes does not contain default lane")
	}
	lanesSet := make(map[types.Lane]struct{})
	for _, lane := range info.lanes {
		lanesSet[lane] = struct{}{}
	}
	if len(info.lanes) != len(lanesSet) {
		return errors.New("list of lanes cannot have repeated values")
	}
	return nil
}
