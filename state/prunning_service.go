// TODO: Does the discard_abci_responses take presedence over the data companion requests? If discard_abci_responses is false,
// do we allow pruning this data based on the need of the data companion. My assumption would be no, because the oeprator should configure both
// But if discard_abci_responses is enabled, and we do have  data companion, we do not prune until we get a signal from the data companion?
// ToDo: Simultanious requests for pruning will be queued and done one after the other. If there is a pruning operation going on based on feedback from the
// data companion, and the app sets the retain height; we will take it into account after pruning is already done
// TODO: Pruning of block results entails pruning the indexer as well?

// Potential practical problems:
// 1. Block for pruning on the app  -> can slow down consensus (as it is right now)
// 2. If we spin up the pruning into a separate go routine for every request we get, too many routines; we can use buffered channels
// 3. If we want to have a return channel with the no_pruned
// TODO pass context to go routines so we can cancel them i ncase of an error

// Minor:
// TODO: Service and channels are maybe not the best solution. We can have a mutex and a function called
// The called then makes sure not to block on the mutex by issuing a call inside a go func
// TODO: Reset the heights once pruning is completed? If we prune what the data companion has requested, do we forget this?
// Probably not, we keep them unless they change to something non 0 or explicitly become 0
// TODO: What if the data companion's pruning height is discarded. Do we care about notifying it (my guess is not, except a message telling it that
// the pruned height is X)

package state

import (
	"errors"
	"fmt"

	"github.com/cometbft/cometbft/libs/log"
	cmtsync "github.com/cometbft/cometbft/libs/sync"
)

type PruningRequester int64

const (
	AppRequester           PruningRequester = 0
	DataCompanionRequester PruningRequester = 1
)

var AppRetainHeightKey = []byte("AppRetainHeightKey")
var DataCompanionRetainHeightKey = []byte("DCRetainHeightKey")

type Pruner struct {
	currentRetainHeight int64
	logger              log.Logger
	// DB to which we save the retain heights
	bs BlockStore
	// State store to prune state from
	stateStore Store

	mtx cmtsync.Mutex
}

func NewPruner(stateStore Store, bs BlockStore, logger log.Logger) *Pruner {
	p := &Pruner{
		bs:         bs,
		stateStore: stateStore,
		logger:     logger,
	}
	return p
}

func (p *Pruner) PruneBlocks(height int64, requester PruningRequester, state State) (pruned uint64, evRetainHeight int64, err error) {
	if height <= 0 {
		return 0, 0, errors.New("retain height cannot be less or equal than 0")
	}
	p.mtx.Lock()
	defer p.mtx.Unlock()
	currentRetainHeight := int64(0)
	switch requester {
	case AppRequester:
		var currentDCRetainHeight int64
		currentDCRetainHeight, err = p.stateStore.GetDataCompanionRetainHeight()
		if err != nil && err != ErrKeyNotFound {
			return 0, 0, nil
		}
		// If pruning was already triggered by the data companion, check whether the
		// retain height of the data companion is lower than the application retain height
		if currentDCRetainHeight < height && currentDCRetainHeight != 0 {
			currentRetainHeight = currentDCRetainHeight
		} else {
			currentRetainHeight = height
		}
		// Persist the application retain height
		err = p.stateStore.SaveApplicationRetainHeight(height)
		if err != nil {
			return
		}
	case DataCompanionRequester:
		var currentAppRetainHeight int64
		currentAppRetainHeight, err = p.stateStore.GetApplicationRetainHeight()
		if err != nil && err != ErrKeyNotFound {
			return
		}
		// If pruning was already triggered by the application, check whether the
		// retain height of the application is lower than the dc retain height
		if currentAppRetainHeight < height && currentAppRetainHeight != 0 {
			currentRetainHeight = currentAppRetainHeight
		} else {
			currentRetainHeight = height
		}
		// Persist the application retain height
		err = p.stateStore.SaveDataCompanionRetainHeight(height)
		if err != nil {
			return
		}
	}
	base := p.bs.Base()

	if base >= currentRetainHeight {
		p.logger.Error(fmt.Sprintf("failed to prune blocks: retain height %d is below or equal to base height %d ", currentRetainHeight, base))
	}
	p.logger.Info(fmt.Sprintf("Received pruning request for %d. Accepted retain height is : %d", height, currentRetainHeight))
	pruned, evRetainHeight, err = p.bs.PruneBlocks(currentRetainHeight, state)
	if err != nil {
		p.logger.Error(fmt.Sprintf("failed to prune blocks for retain height %d with error: %s", p.currentRetainHeight, err))
	} else {
		p.logger.Info("pruned blocks", "pruned", pruned, "retain_height", p.currentRetainHeight)

	}
	// err = p.stateStore.PruneStates(base, currentRetainHeight, evRetainHeight)
	// if err != nil {
	// 	p.logger.Error(fmt.Sprintf("failed to prune the state store with error: %s ", err))
	// }
	return
}

func (p *Pruner) PruneABCIResponses(height int64) (pruned uint64, err error) {
	pruned, err = p.stateStore.PruneABCIResponses(height)
	return
}
