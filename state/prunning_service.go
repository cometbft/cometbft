package state

import (
	"errors"
	"fmt"
	"time"

	"github.com/cometbft/cometbft/libs/log"
	"github.com/cometbft/cometbft/libs/service"
)

type PruningRequester int64

const (
	AppRequester                 PruningRequester = 0
	DataCompanionRequester       PruningRequester = 1
	ABCIResRetainHeightRequester PruningRequester = 2
)

var (
	AppRetainHeightKey           = []byte("AppRetainHeightKey")
	DataCompanionRetainHeightKey = []byte("DCRetainHeightKey")
	ABCIResultsRetainHeightKey   = []byte("ABCIResRetainHeightKey")
)

type Pruner struct {
	service.BaseService
	logger log.Logger
	// DB to which we save the retain heights
	bs BlockStore
	// State store to prune state from
	stateStore Store
}

type RetainHeightInfo struct {
	Height    int64
	Requester PruningRequester
}

func NewPruner(stateStore Store, bs BlockStore, logger log.Logger) *Pruner {
	p := &Pruner{
		bs:         bs,
		stateStore: stateStore,
		logger:     logger,
	}
	p.BaseService = *service.NewBaseService(logger, "Prunning service", p)
	return p
}

// If OnStart() returns an error, it's returned by Start()
func (p *Pruner) OnStart() error {
	if err := p.BaseService.OnStart(); err != nil {
		return err
	}
	go p.pruningRoutine()
	return nil
}

// Stop the service.
// If it's already stopped, will return an error.
// OnStop must never error.

func (p *Pruner) OnStop() {
	p.BaseService.OnStop()

}

// SetPruningHeight is called by either the application or the data companion to set the
// retain height for blocks or ABCI block results.
// It returns the height which will be used as the retain height and an error message in case
// the retain height cannot be set to the desired value (if the data was already pruned, if a lower
// retain height was already set)
func (p *Pruner) SetPruningHeight(retainHeightInfo RetainHeightInfo) (retainHeight int64, err error) {

	if retainHeightInfo.Height <= 0 || retainHeightInfo.Height <= p.bs.Base() {
		return 0, ErrInvalidHeightValue
	}

	var currentRetainHeight int64
	switch retainHeightInfo.Requester {
	case AppRequester:
		currentRetainHeight, err = p.stateStore.GetApplicationRetainHeight()
		if err != nil {
			if err == ErrKeyNotFound {
				err = p.stateStore.SaveApplicationRetainHeight(retainHeightInfo.Height)
				return retainHeightInfo.Height, err
			}
			return 0, err
		}
		if currentRetainHeight > retainHeightInfo.Height {
			return currentRetainHeight, errors.New("cannot set a height lower than previously requested - blocks might have already been pruned")
		}
		err = p.stateStore.SaveApplicationRetainHeight(retainHeightInfo.Height)
		return retainHeightInfo.Height, err
	case DataCompanionRequester:
		currentRetainHeight, err = p.stateStore.GetDataCompanionRetainHeight()
		if err != nil {
			if err == ErrKeyNotFound {
				err = p.stateStore.SaveDataCompanionRetainHeight(retainHeightInfo.Height)
				return retainHeightInfo.Height, err
			}
			return 0, err
		}
		if currentRetainHeight > retainHeightInfo.Height {
			return currentRetainHeight, errors.New("cannot set a height lower than previously requested - blocks might have already been pruned")
		}
		err = p.stateStore.SaveDataCompanionRetainHeight(retainHeightInfo.Height)
		return retainHeightInfo.Height, err
	case ABCIResRetainHeightRequester:
		currentRetainHeight, err = p.stateStore.GetABCIResRetainHeight()
		if err != nil {
			if err == ErrKeyNotFound {
				err = p.stateStore.SaveABCIResRetainHeight(retainHeightInfo.Height)
				return retainHeightInfo.Height, err
			}
			return 0, err
		}
		if currentRetainHeight > retainHeightInfo.Height {
			return currentRetainHeight, errors.New("cannot set a height lower than previously requested - blocks might have already been pruned")
		}
		err = p.stateStore.SaveABCIResRetainHeight(retainHeightInfo.Height)
		return retainHeightInfo.Height, err
	}

	return 0, nil
}
func (p *Pruner) pruningRoutine() {
	lastHeightPruned := int64(0)
	lastABCIResPrunedHeight := int64(0)
	for {
		retainHeight := p.findMinRetainHeight()
		fmt.Println("Min height = ", retainHeight)
		if retainHeight != lastHeightPruned {
			_, _, _ = p.pruneBlocks(retainHeight)
			lastHeightPruned = retainHeight
		}

		ABCIResRetainHeight, err := p.stateStore.GetABCIResRetainHeight()
		if err != nil {
			time.Sleep(time.Second * 10)
		} else {
			if lastABCIResPrunedHeight != ABCIResRetainHeight {
				_, _ = p.stateStore.PruneABCIResponses(ABCIResRetainHeight)
			}
			time.Sleep(time.Second * 10)
		}
	}
}

// If no retain height has been set by the application or the data companion
// the database will not have values for the corresponding keys.
// If both retain heights were set, we pick the smaller one
// If only one is set we return that one
func (p *Pruner) findMinRetainHeight() int64 {
	var noAppRetainHeightSet, noDCRetainHeight bool
	appRetainHeight, err := p.stateStore.GetApplicationRetainHeight()

	if err != nil {
		if err == ErrKeyNotFound {
			noAppRetainHeightSet = true
		} else {
			return 0
		}
	}
	dcRetainHeight, err := p.stateStore.GetDataCompanionRetainHeight()
	if err != nil {
		if err == ErrKeyNotFound {
			noDCRetainHeight = true
		} else {
			return 0
		}
	}
	if !noAppRetainHeightSet && !noDCRetainHeight {
		if appRetainHeight < dcRetainHeight {
			return appRetainHeight
		}
		return dcRetainHeight
	}

	if !noAppRetainHeightSet {
		return appRetainHeight
	}
	if !noAppRetainHeightSet {
		return dcRetainHeight
	}

	return 0
}

func (p *Pruner) pruneBlocks(height int64) (pruned uint64, evRetainHeight int64, err error) {
	if height <= 0 {
		return 0, 0, errors.New("retain height cannot be less or equal than 0")
	}

	base := p.bs.Base()

	if base >= height {
		p.logger.Error(fmt.Sprintf("failed to prune blocks: retain height %d is below or equal to base height %d ", height, base))
	}
	var state State
	state, err = p.stateStore.Load()
	if err != nil {
		p.logger.Error("failed to load state, cannot prune")
		return
	}
	p.logger.Info(fmt.Sprintf("Received pruning request for %d. Accepted retain height is : %d", height, height))
	pruned, evRetainHeight, err = p.bs.PruneBlocks(height, state)
	if err != nil {
		p.logger.Error(fmt.Sprintf("failed to prune blocks for retain height %d with error: %s", height, err))
	} else {
		p.logger.Info("pruned blocks", "pruned", pruned, "retain_height", height)

	}
	err = p.stateStore.PruneStates(base, height, evRetainHeight)
	if err != nil {
		p.logger.Error(fmt.Sprintf("failed to prune the state store with error: %s ", err))
	}
	return
}

func (p *Pruner) PruneABCIResponses(height int64) (pruned uint64, err error) {
	pruned, err = p.stateStore.PruneABCIResponses(height)
	return
}
