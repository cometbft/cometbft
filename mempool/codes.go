package mempool

const (
	// Codespace common for all responses created in mempool
	Codespace = "mempool"

	CodeTypeMempoolIsFull      uint32 = 1
	CodeTypeTxAlreadyInMempool uint32 = 2
	CodeTypePostCheckFailed    uint32 = 3
)
