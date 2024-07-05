- `[consensus]` Make broadcasting HasVote and HasBlockpart control messages
    use TrySend instead of Send. This saves notable amounts of performance,
    as TrySend just drops sending when the channel is full, instead of creating
    a timer / new channel.
  ([\#3342](https://github.com/cometbft/cometbft/issues/3342))
