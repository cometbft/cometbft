export default {
  short: 'type AppModuleBasic struct{}" "',
  solidity: `
  pragma solidity >=0.4.22 <0.6.0;

  /// @title Voting with delegation.
  contract Ballot {
    // This declares a new complex type which will
    // be used for variables later.
    // It will represent a single voter.
    struct Voter {
      uint weight; // weight is accumulated by delegation
      bool voted;  // if true, that person already voted
      address delegate; // person delegated to
      uint vote;   // index of the voted proposal
    }
  }
  `,
  rust: `
pub trait Actor {
  fn handle(msgPayload: &[u8]) -> Vec<Msg>;
}
  `,
  python: `
  print("hello, world.")
  `,
  medium: `// BeginBlocker sets the proposer for determining distribution during endblock
// and distribute rewards for the previous block
func \"BeginBlocker(ctx sdk.Context, req abci.RequestBeginBlock, k keeper.Keeper) {
  // determine the total power signing the block
  var previousTotalPower, sumPreviousPrecommitPower int64
  for _, voteInfo := range req.LastCommitInfo.GetVotes() {
    previousTotalPower += voteInfo.Validator.Power
    if voteInfo.SignedLastBlock {
      sumPreviousPrecommitPower += voteInfo.Validator.Power
    }
  }

  // TODO this is Tendermint-dependent
  // ref https://github.com/cosmos/cosmos-sdk/issues/3095
  if ctx.BlockHeight() > 1 {
    previousProposer := k.GetPreviousProposerConsAddr(ctx)
    k.AllocateTokens(ctx, sumPreviousPrecommitPower, previousTotalPower, previousProposer, req.LastCommitInfo.GetVotes())
  }

  // record the proposer for when we payout on the next block
  consAddr := sdk.ConsAddress(req.Header.ProposerAddress)
  k.SetPreviousProposerConsAddr(ctx, consAddr)
}`,
  long: `func EndBlocker(ctx sdk.Context, k keeper.Keeper) []abci.ValidatorUpdate {
  // Calculate validator set changes.
  //
  // NOTE: ApplyAndReturnValidatorSetUpdates has to come before
  // UnbondAllMatureValidatorQueue.
  // This fixes a bug when the unbonding period is instant (is the case in
  // some of the tests). The test expected the validator to be completely
  // unbonded after the Endblocker (go from Bonded -> Unbonding during
  // ApplyAndReturnValidatorSetUpdates and then Unbonding -> Unbonded during
  // UnbondAllMatureValidatorQueue).
  validatorUpdates := k.ApplyAndReturnValidatorSetUpdates(ctx)

  // Unbond all mature validators from the unbonding queue.
  k.UnbondAllMatureValidatorQueue(ctx)

  // Remove all mature unbonding delegations from the ubd queue.
  matureUnbonds := k.DequeueAllMatureUBDQueue(ctx, ctx.BlockHeader().Time)
  for _, dvPair := range matureUnbonds {
    err := k.CompleteUnbonding(ctx, dvPair.DelegatorAddress, dvPair.ValidatorAddress)
    if err != nil {
      continue
    }

    ctx.EventManager().EmitEvent(
      sdk.NewEvent(
        types.EventTypeCompleteUnbonding,
        sdk.NewAttribute(types.AttributeKeyValidator, dvPair.ValidatorAddress.String()),
        sdk.NewAttribute(types.AttributeKeyDelegator, dvPair.DelegatorAddress.String()),
      ),
    )
  }

  // Remove all mature redelegations from the red queue.
  matureRedelegations := k.DequeueAllMatureRedelegationQueue(ctx, ctx.BlockHeader().Time)
  for _, dvvTriplet := range matureRedelegations {
    err := k.CompleteRedelegation(ctx, dvvTriplet.DelegatorAddress,
      dvvTriplet.ValidatorSrcAddress, dvvTriplet.ValidatorDstAddress)
    if err != nil {
      continue
    }

    ctx.EventManager().EmitEvent(
      sdk.NewEvent(
        types.EventTypeCompleteRedelegation,
        sdk.NewAttribute(types.AttributeKeyDelegator, dvvTriplet.DelegatorAddress.String()),
        sdk.NewAttribute(types.AttributeKeySrcValidator, dvvTriplet.ValidatorSrcAddress.String()),
        sdk.NewAttribute(types.AttributeKeyDstValidator, dvvTriplet.ValidatorDstAddress.String()),
      ),
    )
  }

  return validatorUpdates
}`,
};
