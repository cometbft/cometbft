# Mempool Lanes

This guide provides a set of best practices and rules of thumb to help application developers set up
and use the Lanes feature in the mempool, which allows to classify and prioritise transactions.

While this is not an exhaustive list, it captures key insights from the design, implementation, and
testing phases of the feature.

## Transaction classification and ordering

- **Independent transactions**: Transactions can only be classified into different lanes if they are
  independent of each other. If there is a relationship or dependency between transactions (e.g
  transaction A must be executed before transaction B), both must be placed in the same lane.
  Failing to do so may result in an incorrect ordering, where B could be processed and executed
  before A.
- **Ordering across lanes**: Transactions in separate lanes are not guaranteed to maintain the order
  in which they are processed, disseminated to other nodes. Developers should be aware that
  classification in lanes can result in transaction being committed to different blocks and executed
  in different order.
- **Immutable lane assignment**: Once a transaction is assigned to a lane upon entering the mempool,
  its lane cannot be changed, even during rechecking.
- **Execution timing**: The time gap between the execution of two transactions is unpredictable,
  especially if they are in lanes with significantly different priority levels.

## Number of lanes

- **One lane minimum**: Setting up one lane replicates the behavior of the mempool before lanes were
  introduced. The same behaviour is obtained when the application does not set up lanes: the mempool
  will assign all transactions to the single, default lane. The latter is transparent to users.
- **Start small**: We recommend starting with fewer than 5 or 10 lanes and test them thoroughly on a
  testnet. You can gradually introduce more lanes as necessary once performance and behavior are
  validated.
- **Constraints**: Lanes are identified by strings. In theory, there is no limit to the number of
  lanes that can be defined. However, keep in mind that both memory and CPU usage will increase in
  proportion to the number of lanes.

## Lane priorities

- **Priority values**: Lane priorities are values of type `uint32`. Valid priorities range from 1 to
  `math.MaxUint32`. Priority 0 is reserved for cases where there are no lanes to assign, such as
  invalid transactions or applications that do not utilize lanes. However, if the application
  returns an empty `lane_id` on `CheckTx`, the mempool will assign the default lane as specified in
  `InfoResponse`.
- **Fair scheduling**: Lanes implement a scheduling algorithm for picking transactions
  for dissemination to peers and for creating blocks. The algorithm is designed to be
  _starvation-free_, ensuring that even transactions from lower-priority lanes will eventually be
  disseminated and included in blocks. It also _fair_, because it picks transactions across all
  lanes by interleaving them when possible.
- **Equal priorities**: Multiple lanes are allowed to have the same priority. This could help
  prevent one class of transaction monopolizing the entire mempool. When lanes share the same
  priority, the order in which they are processed is undefined.

## Lane capacity

- **Capacity distribution**: The mempool's capacity is divided evenly among the lanes, with each
  lane's capacity being constrained by both the number of transactions and the total transaction
  size in bytes. Once either limit is reached, no further transactions will be accepted into that
  lane.
- **Preventing spam**: Lane capacity helps mitigate the risk of large transactions flooding the
  network. For optimal performance, large transactions should be assigned to lower-priority lanes
  whenever possible.
- **Adjusting capacities**: If you find that the capacity of a lane is insufficient, you still have
  the option of increasing the total mempool size, which will proportionally increase the capacity
  of all lanes. In future iterations, we may introduce more granular control over lane capacities if
  needed.

## Network setup

- **Limited resources**: Lanes are especially useful in networks with constrained resources, such as
  block size, mempool capacity, or network throughput. In such environments, lanes ensure
  higher-priority transactions will be prioritized during dissemination and block inclusion. In
  networks without these limitations, lanes will not significantly affect the behavior compared to
  nodes that do not implement lanes.
- **Consistent setup**: To fully benefit from lanes, all nodes in the network should implement the
  same lanes configuration. If some nodes do not support lanes, the benefits of lane prioritization
  will not be observed, because transaction ordering during dissemination and processing will be
  inconsistent across nodes. While mixing nodes with and without lanes does not affect network
  correctness, consistent lane configuration is strongly recommended for improved performance and
  consistent behavior.
