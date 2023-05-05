----------------------------- MODULE ABCIMessages ------------------------------
EXTENDS Base

\* @type: ERROR;
InvalidTxError == "invalid-tx"
\* @type: Set(ERROR);
Errors == {NoError, InvalidTxError}

--------------------------------------------------------------------------------
\* https://github.com/CometBFT/cometbft/blob/4790ea3e46475064d5475c787427ae926c5a9e94/proto/tendermint/abci/types.proto#L94
CheckTxTypes == {"New", "Recheck"}

\* https://github.com/CometBFT/cometbft/blob/4790ea3e46475064d5475c787427ae926c5a9e94/proto/tendermint/abci/types.proto#L99
\* @typeAlias: REQUEST = [tag: Str, tx: TX, checkTxType: Str];
\* @type: Set(REQUEST);
RequestCheckTx == [
    tag: {"CheckTx"}, 
    tx: Txs, 
    checkTxType: CheckTxTypes
]

\* https://github.com/CometBFT/cometbft/blob/4790ea3e46475064d5475c787427ae926c5a9e94/proto/tendermint/abci/types.proto#L254
\* @typeAlias: RESPONSE = [tag: Str, error: ERROR];
\* @type: Set(RESPONSE);
ResponseCheckTx == [
    tag: {"CheckTx"}, 
    error: Errors \* called `code` in protobuf
]

================================================================================
Created by Hernán Vanzetto on 1 May 2023
