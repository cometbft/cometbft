// -*- mode: Bluespec; -*-
/* Messages as defined in the protobuf files */
module abciMessages {
    import base.* from "./base"
    import errors.* from "./errors"

    // [tendermint.abci.CheckTxType]: https://github.com/CometBFT/cometbft/blob/4790ea3e46475064d5475c787427ae926c5a9e94/proto/tendermint/abci/types.proto#L94
    val CheckTxTypes = Set("New", "Recheck")

    // [tendermint.abci.RequestCheckTx]: https://github.com/CometBFT/cometbft/blob/4790ea3e46475064d5475c787427ae926c5a9e94/proto/tendermint/abci/types.proto#L99
    type Request = {
        tag: str, // The only value that we use is "CheckTx".
        tx: Tx, 
        checkTxType: str,
    }

    // [tendermint.abci.ResponseCheckTx]: https://github.com/CometBFT/cometbft/blob/4790ea3e46475064d5475c787427ae926c5a9e94/proto/tendermint/abci/types.proto#L254
    type Response = {
        tag: str, 
        err: Error, // called `code` in protobuf
    }
}