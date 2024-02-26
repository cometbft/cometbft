package types

import (
	"errors"
	"fmt"
	"time"

	cmtproto "github.com/cometbft/cometbft/api/cometbft/types/v1"
	"github.com/cometbft/cometbft/crypto/ed25519"
	"github.com/cometbft/cometbft/crypto/secp256k1"
	"github.com/cometbft/cometbft/crypto/tmhash"
	gogo "github.com/cosmos/gogoproto/types" //nolint:allz
)

const (
	// MaxBlockSizeBytes is the maximum permitted size of the blocks.
	MaxBlockSizeBytes = 100 * 1024 * 1024

	// BlockPartSizeBytes is the size of one block part.
	BlockPartSizeBytes uint32 = 65536 // 64kB

	// MaxBlockPartsCount is the maximum number of block parts.
	MaxBlockPartsCount = (MaxBlockSizeBytes / BlockPartSizeBytes) + 1

	ABCIPubKeyTypeEd25519   = ed25519.KeyType
	ABCIPubKeyTypeSecp256k1 = secp256k1.KeyType
)

var ABCIPubKeyTypesToNames = map[string]string{
	ABCIPubKeyTypeEd25519:   ed25519.PubKeyName,
	ABCIPubKeyTypeSecp256k1: secp256k1.PubKeyName,
}

// ConsensusParams contains consensus critical parameters that determine the
// validity of blocks.
type ConsensusParams struct {
	Block     BlockParams     `json:"block"`
	Evidence  EvidenceParams  `json:"evidence"`
	Validator ValidatorParams `json:"validator"`
	Version   VersionParams   `json:"version"`
	Synchrony SynchronyParams `json:"synchrony"`
	Feature   FeatureParams   `json:"feature"`
}

// BlockParams define limits on the block size and gas plus minimum time
// between blocks.
type BlockParams struct {
	MaxBytes int64 `json:"max_bytes"`
	MaxGas   int64 `json:"max_gas"`
}

// EvidenceParams determine how we handle evidence of malfeasance.
type EvidenceParams struct {
	MaxAgeNumBlocks int64         `json:"max_age_num_blocks"` // only accept new evidence more recent than this
	MaxAgeDuration  time.Duration `json:"max_age_duration"`
	MaxBytes        int64         `json:"max_bytes"`
}

// ValidatorParams restrict the public key types validators can use.
// NOTE: uses ABCI pubkey naming, not Amino names.
type ValidatorParams struct {
	PubKeyTypes []string `json:"pub_key_types"`
}

type VersionParams struct {
	App uint64 `json:"app"`
}

// FeatureParams configure the height from which features of CometBFT are enabled.
type FeatureParams struct {
	VoteExtensionsEnableHeight int64 `json:"vote_extensions_enable_height"`
	PbtsEnableHeight           int64 `json:"pbts_enable_height"`
}

// VoteExtensionsEnabled returns true if vote extensions are enabled at height h
// and false otherwise.
func (p FeatureParams) VoteExtensionsEnabled(h int64) bool {
	enabledHeight := p.VoteExtensionsEnableHeight

	return featureEnabled(enabledHeight, h, "Vote Extensions")
}

// PbtsEnabled returns true if PBTS is enabled at height h and false otherwise.
func (p FeatureParams) PbtsEnabled(h int64) bool {
	enabledHeight := p.PbtsEnableHeight

	return featureEnabled(enabledHeight, h, "PBTS")
}

// featureEnabled returns true if `enabledHeight` points to a height that is smaller than `currentHeight“.
func featureEnabled(enableHeight int64, currentHeight int64, f string) bool {
	if currentHeight < 1 {
		panic(fmt.Errorf("cannot check if %s is enabled for height %d (< 1)", f, currentHeight))
	}

	if enableHeight <= 0 {
		return false
	}

	return enableHeight <= currentHeight
}

// SynchronyParams influence the validity of block timestamps.
// For more information on the relationship of the synchrony parameters to
// block validity, see the Proposer-Based Timestamps specification:
// https://github.com/tendermint/spec/blob/master/spec/consensus/proposer-based-timestamp/README.md
type SynchronyParams struct {
	Precision    time.Duration `json:"precision,string"`
	MessageDelay time.Duration `json:"message_delay,string"`
}

// DefaultConsensusParams returns a default ConsensusParams.
func DefaultConsensusParams() *ConsensusParams {
	return &ConsensusParams{
		Block:     DefaultBlockParams(),
		Evidence:  DefaultEvidenceParams(),
		Validator: DefaultValidatorParams(),
		Version:   DefaultVersionParams(),
		Feature:   DefaultFeatureParams(),
		Synchrony: DefaultSynchronyParams(),
	}
}

// DefaultBlockParams returns a default BlockParams.
func DefaultBlockParams() BlockParams {
	return BlockParams{
		MaxBytes: 4194304,  // four megabytes
		MaxGas:   10000000, // ten million
	}
}

// DefaultEvidenceParams returns a default EvidenceParams.
func DefaultEvidenceParams() EvidenceParams {
	return EvidenceParams{
		MaxAgeNumBlocks: 100000, // 27.8 hrs at 1block/s
		MaxAgeDuration:  48 * time.Hour,
		MaxBytes:        1048576, // 1MB
	}
}

// DefaultValidatorParams returns a default ValidatorParams, which allows
// only ed25519 pubkeys.
func DefaultValidatorParams() ValidatorParams {
	return ValidatorParams{
		PubKeyTypes: []string{ABCIPubKeyTypeEd25519},
	}
}

func DefaultVersionParams() VersionParams {
	return VersionParams{
		App: 0,
	}
}

// Disabled by default.
func DefaultFeatureParams() FeatureParams {
	return FeatureParams{
		VoteExtensionsEnableHeight: 0,
		PbtsEnableHeight:           0,
	}
}

func DefaultSynchronyParams() SynchronyParams {
	// TODO(@wbanfield): Determine experimental values for these defaults
	// https://github.com/tendermint/tendermint/issues/7202
	return SynchronyParams{
		Precision:    500 * time.Millisecond,
		MessageDelay: 2 * time.Second,
	}
}

func IsValidPubkeyType(params ValidatorParams, pubkeyType string) bool {
	for i := 0; i < len(params.PubKeyTypes); i++ {
		if params.PubKeyTypes[i] == pubkeyType {
			return true
		}
	}
	return false
}

// ValidateBasic validates the ConsensusParams to ensure **all** values are within their
// allowed limits, and returns an error if they are not.
func (params ConsensusParams) ValidateBasic() error {
	if params.Block.MaxBytes == 0 {
		return fmt.Errorf("block.MaxBytes cannot be 0")
	}
	if params.Block.MaxBytes < -1 {
		return fmt.Errorf("block.MaxBytes must be -1 or greater than 0. Got %d",

			params.Block.MaxBytes)
	}
	if params.Block.MaxBytes > MaxBlockSizeBytes {
		return fmt.Errorf("block.MaxBytes is too big. %d > %d",
			params.Block.MaxBytes, MaxBlockSizeBytes)
	}

	if params.Block.MaxGas < -1 {
		return fmt.Errorf("block.MaxGas must be greater or equal to -1. Got %d",
			params.Block.MaxGas)
	}

	if params.Evidence.MaxAgeNumBlocks <= 0 {
		return fmt.Errorf("evidence.MaxAgeNumBlocks must be greater than 0. Got %d",
			params.Evidence.MaxAgeNumBlocks)
	}

	if params.Evidence.MaxAgeDuration <= 0 {
		return fmt.Errorf("evidence.MaxAgeDuration must be greater than 0 if provided, Got %v",
			params.Evidence.MaxAgeDuration)
	}

	maxBytes := params.Block.MaxBytes
	if maxBytes == -1 {
		maxBytes = int64(MaxBlockSizeBytes)
	}
	if params.Evidence.MaxBytes > maxBytes {
		return fmt.Errorf("evidence.MaxBytesEvidence is greater than upper bound, %d > %d",
			params.Evidence.MaxBytes, params.Block.MaxBytes)
	}

	if params.Evidence.MaxBytes < 0 {
		return fmt.Errorf("evidence.MaxBytes must be non negative. Got: %d",
			params.Evidence.MaxBytes)
	}
	if params.Feature.VoteExtensionsEnableHeight < 0 {
		return fmt.Errorf("Feature.VoteExtensionsEnabledHeight cannot be negative. Got: %d", params.Feature.VoteExtensionsEnableHeight)
	}

	if params.Feature.PbtsEnableHeight < 0 {
		return fmt.Errorf("Feature.PbtsEnableHeight cannot be negative. Got: %d", params.Feature.PbtsEnableHeight)
	}

	if params.Synchrony.MessageDelay <= 0 {
		return fmt.Errorf("synchrony.MessageDelay must be greater than 0. Got: %d",
			params.Synchrony.MessageDelay)
	}

	if params.Synchrony.Precision <= 0 {
		return fmt.Errorf("synchrony.Precision must be greater than 0. Got: %d",
			params.Synchrony.Precision)
	}

	if len(params.Validator.PubKeyTypes) == 0 {
		return errors.New("len(Validator.PubKeyTypes) must be greater than 0")
	}

	// Check if keyType is a known ABCIPubKeyType
	for i := 0; i < len(params.Validator.PubKeyTypes); i++ {
		keyType := params.Validator.PubKeyTypes[i]
		if _, ok := ABCIPubKeyTypesToNames[keyType]; !ok {
			return fmt.Errorf("params.Validator.PubKeyTypes[%d], %s, is an unknown pubkey type",
				i, keyType)
		}
	}

	return nil
}

// ValidateUpdate validates the updated Consensus Params
// if updated == nil, then pass.
func (params ConsensusParams) ValidateUpdate(updated *cmtproto.ConsensusParams, h int64) error {
	if updated == nil {
		return nil
	}

	var err error
	// Validate feature update parameters.
	if updated.Feature != nil {
		err = validateUpdateFeatures(params.Feature, *updated.Feature, h)
	}
	return err
}

// validateUpdateFeatures validates the updated PBTSEnableHeight.
// | r | params...EnableHeight | updated...EnableHeight | result (nil == pass)
// |  2 | *                    | < 0                    | EnableHeight must be positive
// |  3 | <=0                  | 0                      | nil
// |  4 | X                    | X (>=0)                | nil
// |  5 | > 0; <=height        | 0                      | Feature cannot be disabled once enabled
// |  6 | > 0; > height        | 0                      | nil (disable a previous proposal)
// |  7 | *                    | <=height               | Feature cannot be updated to a past height
// |  8 | <=0                  | > height (*)           | nil
// |  9 | (> 0) <=height       | > height (*)           | Feature cannot be modified once enabled
// | 10 | (> 0) > height       | > height (*)           | nil
// The table above reflects all cases covered.
func validateUpdateFeatures(params FeatureParams, updated cmtproto.FeatureParams, h int64) error {
	if updated.VoteExtensionsEnableHeight != nil {
		err := validateUpdateFeatureEnableHeight(params.VoteExtensionsEnableHeight, updated.VoteExtensionsEnableHeight.Value, h, "Vote Extensions")
		if err != nil {
			return err
		}
	}

	if updated.PbtsEnableHeight != nil {
		err := validateUpdateFeatureEnableHeight(params.PbtsEnableHeight, updated.PbtsEnableHeight.Value, h, "PBTS")
		if err != nil {
			return err
		}
	}
	return nil
}

func validateUpdateFeatureEnableHeight(param int64, updated int64, h int64, featureName string) error {
	// 2
	if updated < 0 {
		return fmt.Errorf("%s EnableHeight must be positive", featureName)
	}
	// 3
	if param <= 0 && updated == 0 {
		return nil
	}
	// 4 (implicit: updated >= 0)
	if param == updated {
		return nil
	}
	// 5 & 6
	if param > 0 && updated == 0 {
		// 5
		if param <= h {
			return fmt.Errorf("%s cannot be disabled once enabled"+
				"enabled height: %d, current height: %d",
				featureName, param, h)
		}
		// 6
		return nil
	}
	// 7 (implicit: updated > 0)
	if updated <= h {
		return fmt.Errorf("%s cannot be updated to a past or current height, "+
			"enabled height: %d, enable height: %d, current height %d",
			featureName, param, updated, h)
	}
	// 8 (implicit: updated > h)
	if param <= 0 {
		return nil
	}
	// 9 (implicit: param > 0 && updated > h)
	if param <= h {
		return fmt.Errorf("%s cannot be modified once enabled"+
			"enabled height: %d, current height: %d",
			featureName, param, h)
	}
	// 10 (implicit: param > h && updated > h)
	return nil
}

// Hash returns a hash of a subset of the parameters to store in the block header.
// Only the Block.MaxBytes and Block.MaxGas are included in the hash.
// This allows the ConsensusParams to evolve more without breaking the block
// protocol. No need for a Merkle tree here, just a small struct to hash.
func (params ConsensusParams) Hash() []byte {
	hasher := tmhash.New()

	hp := cmtproto.HashedParams{
		BlockMaxBytes: params.Block.MaxBytes,
		BlockMaxGas:   params.Block.MaxGas,
	}

	bz, err := hp.Marshal()
	if err != nil {
		panic(err)
	}

	_, err = hasher.Write(bz)
	if err != nil {
		panic(err)
	}
	return hasher.Sum(nil)
}

// Update returns a copy of the params with updates from the non-zero fields of p2.
// NOTE: note: must not modify the original.
func (params ConsensusParams) Update(params2 *cmtproto.ConsensusParams) ConsensusParams {
	res := params // explicit copy

	if params2 == nil {
		return res
	}

	// we must defensively consider any structs may be nil
	if params2.Block != nil {
		res.Block.MaxBytes = params2.Block.MaxBytes
		res.Block.MaxGas = params2.Block.MaxGas
	}
	if params2.Evidence != nil {
		res.Evidence.MaxAgeNumBlocks = params2.Evidence.MaxAgeNumBlocks
		res.Evidence.MaxAgeDuration = params2.Evidence.MaxAgeDuration
		res.Evidence.MaxBytes = params2.Evidence.MaxBytes
	}
	if params2.Validator != nil {
		// Copy params2.Validator.PubkeyTypes, and set result's value to the copy.
		// This avoids having to initialize the slice to 0 values, and then write to it again.
		res.Validator.PubKeyTypes = append([]string{}, params2.Validator.PubKeyTypes...)
	}
	if params2.Version != nil {
		res.Version.App = params2.Version.App
	}
	if params2.Feature != nil {
		if params2.Feature.VoteExtensionsEnableHeight != nil {
			res.Feature.VoteExtensionsEnableHeight = params2.Feature.GetVoteExtensionsEnableHeight().Value
		}

		if params2.Feature.PbtsEnableHeight != nil {
			res.Feature.PbtsEnableHeight = params2.Feature.GetPbtsEnableHeight().Value
		}
	}
	if params2.Synchrony != nil {
		if params2.Synchrony.MessageDelay != nil {
			res.Synchrony.MessageDelay = *params2.Synchrony.GetMessageDelay()
		}
		if params2.Synchrony.Precision != nil {
			res.Synchrony.Precision = *params2.Synchrony.GetPrecision()
		}
	}

	return res
}

func (params *ConsensusParams) ToProto() cmtproto.ConsensusParams {
	return cmtproto.ConsensusParams{
		Block: &cmtproto.BlockParams{
			MaxBytes: params.Block.MaxBytes,
			MaxGas:   params.Block.MaxGas,
		},
		Evidence: &cmtproto.EvidenceParams{
			MaxAgeNumBlocks: params.Evidence.MaxAgeNumBlocks,
			MaxAgeDuration:  params.Evidence.MaxAgeDuration,
			MaxBytes:        params.Evidence.MaxBytes,
		},
		Validator: &cmtproto.ValidatorParams{
			PubKeyTypes: params.Validator.PubKeyTypes,
		},
		Version: &cmtproto.VersionParams{
			App: params.Version.App,
		},
		Feature: &cmtproto.FeatureParams{
			PbtsEnableHeight:           &gogo.Int64Value{Value: params.Feature.PbtsEnableHeight},
			VoteExtensionsEnableHeight: &gogo.Int64Value{Value: params.Feature.VoteExtensionsEnableHeight},
		},
		Synchrony: &cmtproto.SynchronyParams{
			MessageDelay: &params.Synchrony.MessageDelay,
			Precision:    &params.Synchrony.Precision,
		},
	}
}

func ConsensusParamsFromProto(pbParams cmtproto.ConsensusParams) ConsensusParams {
	c := ConsensusParams{ // TODO: Should we use DefaultConsensusParams here?
		Block: BlockParams{
			MaxBytes: pbParams.Block.MaxBytes,
			MaxGas:   pbParams.Block.MaxGas,
		},
		Evidence: EvidenceParams{
			MaxAgeNumBlocks: pbParams.Evidence.MaxAgeNumBlocks,
			MaxAgeDuration:  pbParams.Evidence.MaxAgeDuration,
			MaxBytes:        pbParams.Evidence.MaxBytes,
		},
		Validator: ValidatorParams{
			PubKeyTypes: pbParams.Validator.PubKeyTypes,
		},
		Version: VersionParams{
			App: pbParams.Version.App,
		},
		Feature: FeatureParams{
			VoteExtensionsEnableHeight: pbParams.GetFeature().GetVoteExtensionsEnableHeight().GetValue(),
			PbtsEnableHeight:           pbParams.GetFeature().GetPbtsEnableHeight().GetValue(),
		},
		Synchrony: SynchronyParams{
			MessageDelay: *pbParams.GetSynchrony().GetMessageDelay(),
			Precision:    *pbParams.GetSynchrony().GetPrecision(),
		},
	}
	if pbParams.GetAbci().GetVoteExtensionsEnableHeight() > 0 {
		// Value set before the upgrade to V1.
		// ABCIParams are only set in <V1 and FeatureParams is only present in >=V1, hence
		// we can safely overwrite here.
		c.Feature.VoteExtensionsEnableHeight = pbParams.Abci.VoteExtensionsEnableHeight
	}
	return c
}
