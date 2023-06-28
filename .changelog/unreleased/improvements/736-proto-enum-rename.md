- `[proto]` Add the `abci/v4` versioned proto package with naming changes
  for enum types to suit the
  [buf guidelines](https://buf.build/docs/best-practices/style-guide/#enums):
  ([\#736](https://github.com/cometbft/cometbft/issues/736)):
  * `CheckTxType` values renamed with the `CHECK_TX_TYPE_` prefix.
  * `MisbehaviorType` values renamed with the `MISBEHAVIOR_TYPE_` prefix.
  * `Result` enum in `ResponseOfferSnapshot` renamed to package level
    `OfferSnapshotResult`, its values named with the
    `OFFER_SNAPSHOT_RESULT_` prefix.
  * `Result` enum in `ResponseApplyShapshotChunk` renamed to package level
    `ApplySnapshotChunkResult`, its values named with the
    `APPLY_SNAPSHOT_CHUNK_RESULT_` prefix.
  * `Status` enum in `ResponseProcessProposal` renamed to package level
    `ProcessProposalStatus`, its values named with the
    `PROCESS_PROPOSAL_STATUS_` prefix.
  * `Status` enum in `ResponseVerifyVoteExtension` renamed to package level
    `VerifyVoteExtensionStatus`, its values named with the
    `VERIFY_VOTE_EXTENSION_STATUS_` prefix.
  * Message types using the enumeration types listed above get new definitions:
    `Request`,
    `RequestCheckTx`,
    `RequestPrepareProposal`,
    `RequestProcessProposal`,
    `RequestFinalizeBlock`,
    `Response`,
    `ResponseOfferSnapshot`,
    `ResponseApplySnapshotChunk`,
    `ResponseProcessProposal`,
    `ResponseVerifyVoteExtension`,
    `Misbehavior`.
  * New version of the `ABCI` service defined using the types listed above.
