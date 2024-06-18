- `[p2p]` Remove `p2p_peer_send_bytes_total` metric as it is costly to track,
  and not that informative in debugging.
  ([\#3184](https://github.com/cometbft/cometbft/issues/3184))
- `[p2p]` Remove `p2p_peer_receive_bytes_total`,
  `p2p_message_receive_bytes_total` and `p2p_message_send_bytes_total` metrics
  ([\#3184](https://github.com/cometbft/cometbft/issues/3184))
