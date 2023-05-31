module errors {
    type Error = str
    val noError: Error = "none"

    val InvalidTxError: Error = "err:invalid-tx"
    val MempoolIsFullError: Error = "err:mempool-full"
    val TxInCacheError: Error = "err:tx-in-cache"
    val InvalidTxWarning: Error = "warn:invalid-tx"
}
