# Orderbook

`Orderbook` is a basic decentralised exchange enabling trading of currencies.

It is built ontop of Tendermint, outlining the key methods of ABCI++ to give users an example of how it is used.

The orderbook demonstrates the abci methods such as CheckTX, ValidateTX, PrepareProposal, ProcessProposal, DeliverTX and Commit.


# Guidelines
Each currency can be added only as a pair of currencies in which one can be traded for the other. Each account that is added can either create an `ask`, which sets an amount that the user would like to sell their commodity at or they can create a `bid`. A `bid` proposes a price that an account would like to purchase a commodity, a bid and ask are then matched based on the MaxPrice is greater than the AskPrice and the MaxQuantity is greater than the quantity.

# Explaination of each 

// CheckTx indicates which transactions should be accepted in the mempool. It is
// not a perfect validity check because we're unsure of the state that the transaction
// will be executed against. We should treat this as a gatekeeper to the mempool.
// Apart from adding transactions to the app-side mempool, this check is stateless.

// ValidateTx validates the transactions against state.

// PrepareProposal is called whenever the validator is the proposer for that round. First, it adds the non order
// transactions provided by tendermint. The orderbook then loops through each market and tries to match as many
// transactions as possible. For each new transaction it checks that the max bytes has not been exceeded.

// Process Proposal either rejects or accepts transactions
//
// It uses the same validity function for prepare proposal. This ensures the coherence property
// is adhered to i.e. all honest validators must accept a proposal by an honest proposer

// DeliverTx is called for each tx in a block once it has been finalized. This is where the
// execution code lives. Most importantly it's where we update the user accounts following
// a successful order.

// Commit is called to tell the app it is safe to persist state to disk.
// We now take the in-memory representation and update the parts that have
// changed on to disk.

The actions that are accessible through the CLI are creating a new account, creating a new pair of commodities, create a bid, create an ask, querying the account, listing all the pairs registered and listing all the current orders for a pair.