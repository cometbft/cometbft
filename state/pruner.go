package state

import (
	"errors"
	"sync"
	"time"

	"github.com/cometbft/cometbft/config"
	"github.com/cometbft/cometbft/libs/log"
	"github.com/cometbft/cometbft/libs/service"
)

var (
	AppRetainHeightKey            = []byte("AppRetainHeightKey")
	CompanionBlockRetainHeightKey = []byte("DCBlockRetainHeightKey")
	ABCIResultsRetainHeightKey    = []byte("ABCIResRetainHeightKey")
)

// Pruner is a service that reads the retain heights for blocks, state and ABCI
// results from the database and prunes the corresponding data based on the
// minimum retain height set. The service sleeps between each run based on the
// configured pruner interval, and re-evaluates the retain height.
type Pruner struct {
	service.BaseService
	logger log.Logger

	mtx sync.Mutex
	// DB to which we save the retain heights
	bs BlockStore
	// State store to prune state from
	stateStore Store

	interval time.Duration

	observer PrunerObserver

	metrics *Metrics
}

type prunerConfig struct {
	interval time.Duration
	observer PrunerObserver
	metrics  *Metrics
}

func defaultPrunerConfig() *prunerConfig {
	return &prunerConfig{
		interval: config.DefaultPruningInterval,
		observer: &NoopPrunerObserver{},
		metrics:  NopMetrics(),
	}
}

type PrunerOption func(*prunerConfig)

// WithPrunerInterval allows control over the interval between each run of the
// pruner.
func WithPrunerInterval(t time.Duration) PrunerOption {
	return func(p *prunerConfig) { p.interval = t }
}

func WithPrunerObserver(obs PrunerObserver) PrunerOption {
	return func(p *prunerConfig) { p.observer = obs }
}

func WithPrunerMetrics(metrics *Metrics) PrunerOption {
	return func(p *prunerConfig) {
		p.metrics = metrics
	}
}

func NewPruner(stateStore Store, bs BlockStore, logger log.Logger, options ...PrunerOption) *Pruner {
	cfg := defaultPrunerConfig()
	for _, opt := range options {
		opt(cfg)
	}
	p := &Pruner{
		bs:         bs,
		stateStore: stateStore,
		logger:     logger,
		interval:   cfg.interval,
		observer:   cfg.observer,
		metrics:    cfg.metrics,
	}
	p.BaseService = *service.NewBaseService(logger, "Pruner", p)
	return p
}

func (p *Pruner) SetObserver(obs PrunerObserver) {
	p.observer = obs
}

func (p *Pruner) OnStart() error {
	go p.pruneBlocksRoutine()
	go p.pruneABCIResRoutine()
	p.observer.PrunerStarted(p.interval)
	return nil
}

// SetApplicationRetainHeight sets the application retain height with some
// basic checks on the requested height.
//
// If a higher retain height is already set, we cannot accept the requested
// height because the blocks might have been pruned.
//
// If the data companion has already set a retain height to a higher value we
// also cannot accept the requested height as the blocks might have been
// pruned.
func (p *Pruner) SetApplicationRetainHeight(height int64) error {
	// Ensure that all requests to set retain heights via the application are
	// serialized.
	p.mtx.Lock()
	defer p.mtx.Unlock()

	if !p.checkHeightBound(height) {
		return ErrInvalidHeightValue
	}
	currentAppRetainHeight, err := p.stateStore.GetApplicationRetainHeight()
	if err != nil {
		if !errors.Is(err, ErrKeyNotFound) {
			return err
		}
		currentAppRetainHeight = height
	}
	currentCompanionRetainHeight, err := p.stateStore.GetCompanionBlockRetainHeight()
	noCompanionRetainHeight := false
	if err != nil {
		if !errors.Is(err, ErrKeyNotFound) {
			return err
		}
		noCompanionRetainHeight = true
	}
	if currentAppRetainHeight > height || (!noCompanionRetainHeight && currentCompanionRetainHeight > height) {
		return ErrPrunerCannotLowerRetainHeight
	}
	if err := p.stateStore.SaveApplicationRetainHeight(height); err != nil {
		return err
	}
	p.metrics.ApplicationBlockRetainHeight.Set(float64(height))
	return nil
}

func (p *Pruner) checkHeightBound(height int64) bool {
	if height <= 0 || height < p.bs.Base() || height > p.bs.Height() {
		return false
	}
	return true
}

// SetCompanionRetainHeight sets the application retain height with some basic
// checks on the requested height.
//
// If a higher retain height is already set, we cannot accept the requested
// height because the blocks might have been pruned.
//
// If the application has already set a retain height to a higher value we also
// cannot accept the requested height as the blocks might have been pruned.
func (p *Pruner) SetCompanionRetainHeight(height int64) error {
	// Ensure that all requests to set retain heights via the pruner are
	// serialized.
	p.mtx.Lock()
	defer p.mtx.Unlock()

	if !p.checkHeightBound(height) {
		return ErrInvalidHeightValue
	}
	currentCompanionRetainHeight, err := p.stateStore.GetCompanionBlockRetainHeight()
	if err != nil {
		if !errors.Is(err, ErrKeyNotFound) {
			return err
		}
		currentCompanionRetainHeight = height
	}
	currentAppRetainHeight, err := p.stateStore.GetApplicationRetainHeight()
	noAppRetainHeight := false
	if err != nil {
		if !errors.Is(err, ErrKeyNotFound) {
			return err
		}
		noAppRetainHeight = true
	}
	if currentCompanionRetainHeight > height || (!noAppRetainHeight && currentAppRetainHeight > height) {
		return ErrPrunerCannotLowerRetainHeight
	}
	if err := p.stateStore.SaveCompanionBlockRetainHeight(height); err != nil {
		return err
	}
	p.metrics.PruningServiceBlockRetainHeight.Set(float64(height))
	return nil
}

// SetABCIResRetainHeight sets the retain height for ABCI responses.
//
// If the application has set the DiscardABCIResponses flag to true, nothing
// will be pruned.
func (p *Pruner) SetABCIResRetainHeight(height int64) error {
	// Ensure that all requests to set retain heights via the pruner are
	// serialized.
	p.mtx.Lock()
	defer p.mtx.Unlock()

	if height <= 0 || height > p.bs.Height() {
		return ErrInvalidHeightValue
	}
	currentRetainHeight, err := p.stateStore.GetABCIResRetainHeight()
	if err != nil {
		if !errors.Is(err, ErrKeyNotFound) {
			return err
		}
		return p.stateStore.SaveABCIResRetainHeight(height)
	}
	if currentRetainHeight > height {
		return ErrPrunerCannotLowerRetainHeight
	}
	if err := p.stateStore.SaveABCIResRetainHeight(height); err != nil {
		return err
	}
	p.metrics.PruningServiceBlockResultsRetainHeight.Set(float64(height))
	return nil
}

// GetApplicationRetainHeight is a convenience method for accessing the
// GetApplicationRetainHeight method of the underlying state store.
func (p *Pruner) GetApplicationRetainHeight() (int64, error) {
	return p.stateStore.GetApplicationRetainHeight()
}

// GetCompanionBlockRetainHeight is a convenience method for accessing the
// GetCompanionBlockRetainHeight method of the underlying state store.
func (p *Pruner) GetCompanionBlockRetainHeight() (int64, error) {
	return p.stateStore.GetCompanionBlockRetainHeight()
}

// GetABCIResRetainHeight is a convenience method for accessing the
// GetABCIResRetainHeight method of the underlying state store.
func (p *Pruner) GetABCIResRetainHeight() (int64, error) {
	return p.stateStore.GetABCIResRetainHeight()
}

func (p *Pruner) pruneABCIResRoutine() {
	p.logger.Info("Started pruning ABCI results", "interval", p.interval.String())
	lastABCIResRetainHeight := int64(0)
	for {
		select {
		case <-p.Quit():
			return
		default:
			newABCIResRetainHeight := p.pruneABCIResToRetainHeight(lastABCIResRetainHeight)
			p.observer.PrunerPrunedABCIRes(&ABCIResponsesPrunedInfo{
				FromHeight: lastABCIResRetainHeight,
				ToHeight:   newABCIResRetainHeight - 1,
			})
			lastABCIResRetainHeight = newABCIResRetainHeight
			time.Sleep(p.interval)
		}
	}
}
func (p *Pruner) pruneBlocksRoutine() {
	p.logger.Info("Started pruning blocks", "interval", p.interval.String())
	lastRetainHeight := int64(0)
	for {
		select {
		case <-p.Quit():
			return
		default:
			newRetainHeight := p.pruneBlocksToRetainHeight(lastRetainHeight)
			p.observer.PrunerPrunedBlocks(&BlocksPrunedInfo{
				FromHeight: lastRetainHeight,
				ToHeight:   newRetainHeight - 1,
			})
			lastRetainHeight = newRetainHeight
			time.Sleep(p.interval)
		}
	}
}

func (p *Pruner) pruneBlocksToRetainHeight(lastRetainHeight int64) int64 {
	targetRetainHeight := p.findMinRetainHeight()
	if targetRetainHeight == lastRetainHeight {
		return lastRetainHeight
	}
	pruned, err := p.pruneBlocks(targetRetainHeight)
	// The new retain height is the current lowest point of the block store
	// indicated by Base()
	newRetainHeight := p.bs.Base()
	if err != nil {
		p.logger.Error("Failed to prune blocks", "err", err, "targetRetainHeight", targetRetainHeight, "newRetainHeight", newRetainHeight)
	} else if pruned > 0 {
		p.metrics.BlockStoreBaseHeight.Set(float64(newRetainHeight))
		p.logger.Info("Pruned blocks", "count", pruned, "newRetainHeight", newRetainHeight)
	}
	return newRetainHeight
}

func (p *Pruner) pruneABCIResToRetainHeight(lastRetainHeight int64) int64 {
	targetRetainHeight, err := p.stateStore.GetABCIResRetainHeight()
	if err != nil {
		// ABCI response retain height has not yet been set - do not log any
		// errors at this time.
		if errors.Is(err, ErrKeyNotFound) {
			return 0
		}
		p.logger.Error("Failed to get ABCI result retain height", "err", err)
		return lastRetainHeight
	}

	if lastRetainHeight == targetRetainHeight {
		return lastRetainHeight
	}

	// newRetainHeight is the height just after that which we have successfully
	// pruned. In case of an error it will be 0, but then it will also be
	// ignored.
	numPruned, newRetainHeight, err := p.stateStore.PruneABCIResponses(targetRetainHeight)
	if err != nil {
		p.logger.Error("Failed to prune ABCI responses", "err", err, "targetRetainHeight", targetRetainHeight)
		return lastRetainHeight
	}
	if numPruned > 0 {
		p.logger.Info("Pruned ABCI responses", "heights", numPruned, "newRetainHeight", newRetainHeight)
		p.metrics.ABCIResultsBaseHeight.Set(float64(newRetainHeight))
	}
	return newRetainHeight
}

// If no retain height has been set by the application or the data companion
// the database will not have values for the corresponding keys.
// If both retain heights were set, we pick the smaller one
// If only one is set we return that one
func (p *Pruner) findMinRetainHeight() int64 {
	noAppRetainHeightSet := false
	appRetainHeight, err := p.stateStore.GetApplicationRetainHeight()
	if err != nil {
		if !errors.Is(err, ErrKeyNotFound) {
			return 0
		}
		noAppRetainHeightSet = true
	}
	dcRetainHeight, err := p.stateStore.GetCompanionBlockRetainHeight()
	if err != nil {
		if !errors.Is(err, ErrKeyNotFound) {
			return 0
		}
		// The Application height was set so we can return that immediately
		if !noAppRetainHeightSet {
			return appRetainHeight
		}
	}
	// If we are here, both heights were set so we are picking the minimum
	if appRetainHeight < dcRetainHeight {
		return appRetainHeight
	}
	return dcRetainHeight
}

func (p *Pruner) pruneBlocks(height int64) (uint64, error) {
	if height <= 0 {
		return 0, ErrInvalidRetainHeight
	}

	base := p.bs.Base()

	// state, err := p.stateStore.Load()
	// if err != nil {
	// 	return 0, 0, ErrPrunerFailedToLoadState{Err: err}
	// }
	pruned, err := p.bs.PruneBlocks(height)
	if err != nil {
		return 0, ErrFailedToPruneBlocks{Height: height, Err: err}
	}
	if err := p.stateStore.PruneStates(base, height); err != nil {
		return 0, ErrFailedToPruneStates{Height: height, Err: err}
	}
	return pruned, err
}
