package state

import (
	"errors"
	"time"

	"github.com/cometbft/cometbft/libs/log"
	"github.com/cometbft/cometbft/libs/service"
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
	stateStore Store
	interval   time.Duration
}

type PrunerOption func(*Pruner)

func PrunerInterval(t time.Duration) PrunerOption {
	return func(p *Pruner) { p.interval = t }
}

func NewPruner(stateStore Store, bs BlockStore, logger log.Logger, options ...PrunerOption) *Pruner {
	p := &Pruner{
		bs:         bs,
		stateStore: stateStore,
		logger:     logger,
		interval:   time.Second * 10,
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
	p.Quit()
	p.BaseService.OnStop()
}

// SetApplicationRetainHeight sets the application retain height with some
// basic checks on the requested height.
// If a higher retain height is already set, we cannot accept the requested height
// because the blocks might have been pruned.
// If the data companion has already set a retain height to a higher value
// we also cannot accept the requested height as the blocks might have been pruned
func (p *Pruner) SetApplicationRetainHeight(height int64) error {
	if height <= 0 || height < p.bs.Base() || height > p.bs.Height() {
		return ErrInvalidHeightValue
	}
	currentAppRetainHeight, err := p.stateStore.GetApplicationRetainHeight()
	if err != nil {
		if err == ErrKeyNotFound {
			currentAppRetainHeight = height
		} else {
			return err
		}
	}
	currentCompanionRetainHeight, err := p.stateStore.GetCompanionBlockRetainHeight()
	var noCompanionRetainHeight bool
	if err != nil {
		if err == ErrKeyNotFound {
			noCompanionRetainHeight = true
		} else {
			return err
		}
	}
	if currentAppRetainHeight > height || (!noCompanionRetainHeight && currentCompanionRetainHeight > height) {
		return errors.New("cannot set a height lower than previously requested - blocks might have already been pruned")
	}
	err = p.stateStore.SaveApplicationRetainHeight(height)
	return err
}

// SetCompanionRetainHeight sets the application retain height with some
// basic checks on the requested height.
// If a higher retain height is already set, we cannot accept the requested height
// because the blocks might have been pruned.
// If the application has already set a retain height to a higher value
// we also cannot accept the requested height as the blocks might have been pruned
func (p *Pruner) SetCompanionRetainHeight(height int64) error {
	if height <= 0 || height < p.bs.Base() || height > p.bs.Height() {
		return ErrInvalidHeightValue
	}
	currentCompanionRetainHeight, err := p.stateStore.GetCompanionBlockRetainHeight()
	if err != nil {
		if err == ErrKeyNotFound {
			currentCompanionRetainHeight = height
		} else {
			return err
		}
	}
	currentAppRetainHeight, err := p.stateStore.GetApplicationRetainHeight()
	var noAppRetainHeight bool
	if err != nil {
		if err == ErrKeyNotFound {
			noAppRetainHeight = true
		} else {
			return err
		}
	}
	if currentCompanionRetainHeight > height || (!noAppRetainHeight && currentAppRetainHeight > height) {
		return errors.New("cannot set a height lower than previously requested - blocks might have already been pruned")
	}
	err = p.stateStore.SaveCompanionBlockRetainHeight(height)
	return err
}

// SetABCIResRetainHeight sets the retain height for ABCI responses
// If the application has set the DiscardABCIResponses flag to true
// Nothing will be pruned
func (p *Pruner) SetABCIResRetainHeight(height int64) error {
	if height <= 0 || height > p.bs.Height() {
		return ErrInvalidHeightValue
	}
	currentRetainHeight, err := p.stateStore.GetABCIResRetainHeight()
	if err != nil {
		if err == ErrKeyNotFound {
			err = p.stateStore.SaveABCIResRetainHeight(height)
			return err
		}
		return err
	}
	if currentRetainHeight > height {
		return errors.New("cannot set a height lower than previously requested - blocks might have already been pruned")
	}
	err = p.stateStore.SaveABCIResRetainHeight(height)
	return err
}

func (p *Pruner) pruningRoutine() {
	lastHeightPruned := int64(0)
	lastABCIResPrunedHeight := int64(0)
	for {
		select {
		case <-p.Quit():
			return
		default:
			retainHeight := p.FindMinRetainHeight()
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
			time.Sleep(p.interval)
		}
	}
}

// If no retain height has been set by the application or the data companion
// the database will not have values for the corresponding keys.
// If both retain heights were set, we pick the smaller one
// If only one is set we return that one
func (p *Pruner) FindMinRetainHeight() int64 {
	var noAppRetainHeightSet bool
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
			// The Application height was set so we can return that immediately
			if !noAppRetainHeightSet {
				return appRetainHeight
			}
		} else {
			return 0
		}
	}
	// If we are here, both heights were set so we are picking the minimum
	if appRetainHeight < dcRetainHeight {
		return appRetainHeight
	}
	return dcRetainHeight
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
