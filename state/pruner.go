package state

import (
	"errors"
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
	AppRetainHeightKey            = []byte("AppRetainHeightKey")
	CompanionBlockRetainHeightKey = []byte("DCBlockRetainHeightKey")
	ABCIResultsRetainHeightKey    = []byte("ABCIResRetainHeightKey")
)

// Pruner is a service that reads the retain heights for blocks, state and ABCI results from the database
// and prunes the corresponding data based on the minimum retain height set.
// The service runs periodically (prunerSleepTime - 10s by default) and re-evaluates the retain height.
type Pruner struct {
	service.BaseService
	logger log.Logger
	// DB to which we save the retain heights
	bs BlockStore
	// State store to prune state from
	stateStore      Store
	prunerSleepTime time.Duration
}
type PrunerOption func(*Pruner)

func PrunerSleepTime(t time.Duration) PrunerOption {
	return func(p *Pruner) { p.prunerSleepTime = t }
}

type RetainHeightInfo struct {
	Height    int64
	Requester PruningRequester
}

func NewPruner(stateStore Store, bs BlockStore, logger log.Logger, options ...PrunerOption) *Pruner {
	p := &Pruner{
		bs:              bs,
		stateStore:      stateStore,
		logger:          logger,
		prunerSleepTime: time.Second * 10,
	}
	p.BaseService = *service.NewBaseService(logger, "Pruner", p)
	for _, option := range options {
		option(p)
	}

	return p
}

func (p *Pruner) OnStart() error {
	if err := p.BaseService.OnStart(); err != nil {
		return err
	}
	go p.pruningRoutine()
	return nil
}

func (p *Pruner) OnStop() {
	p.BaseService.OnStop()

}

// SetPruningHeight is called by either the application or the data companion to set the
// retain height for blocks or ABCI block results.
// It returns the height which will be used as the retain height and an error message in case
// the retain height cannot be set to the desired value (if the data was already pruned, if a lower
// retain height was already set)
func (p *Pruner) SetPruningHeight(retainHeightInfo RetainHeightInfo) (retainHeight int64, err error) {

	// TODO we need to lock retrieval of base and height
	if retainHeightInfo.Height <= 0 || retainHeightInfo.Height < p.bs.Base() || retainHeightInfo.Height > p.bs.Height() {
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
		currentRetainHeight, err = p.stateStore.GetCompanionBlockRetainHeight()
		if err != nil {
			if err == ErrKeyNotFound {
				err = p.stateStore.SaveCompanionBlockRetainHeight(retainHeightInfo.Height)
				return retainHeightInfo.Height, err
			}
			return 0, err
		}
		if currentRetainHeight > retainHeightInfo.Height {
			return currentRetainHeight, errors.New("cannot set a height lower than previously requested - blocks might have already been pruned")
		}
		err = p.stateStore.SaveCompanionBlockRetainHeight(retainHeightInfo.Height)
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
		select {
		case <-p.Quit():
			return
		default:
			retainHeight := p.findMinRetainHeight()
			if retainHeight != lastHeightPruned {
				pruned, evRetainHeight, err := p.pruneBlocks(retainHeight)
				if err != nil {
					p.logger.Error("Failed to prune blocks", "err", err)
				} else {
					p.logger.Debug("Pruned block(s)", "height", pruned, "evidenceRetainHeight", evRetainHeight)
				}
				lastHeightPruned = retainHeight
			}

			ABCIResRetainHeight, err := p.stateStore.GetABCIResRetainHeight()
			if err == nil {
				if lastABCIResPrunedHeight != ABCIResRetainHeight {
					pruned, _ := p.stateStore.PruneABCIResponses(ABCIResRetainHeight)
					p.logger.Debug("Number of ABCI responses pruned: ", "pruned", pruned)
				}
			}
			time.Sleep(p.prunerSleepTime)
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
	dcRetainHeight, err := p.stateStore.GetCompanionBlockRetainHeight()
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

	var state State
	state, err = p.stateStore.Load()
	if err != nil {
		p.logger.Error("Failed to load state, cannot prune")
		return
	}

	pruned, evRetainHeight, err = p.bs.PruneBlocks(height, state)
	if err != nil {
		p.logger.Error("Failed to prune blocks at height", "height", height, "err", err)
	} else {
		p.logger.Debug("Pruned blocks", "pruned", pruned, "retain_height", height)

	}
	err = p.stateStore.PruneStates(base, height, evRetainHeight)
	if err != nil {
		p.logger.Error("Failed to prune the state store", "err", err)
	}
	return pruned, evRetainHeight, err
}
