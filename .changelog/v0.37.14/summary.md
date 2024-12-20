*December 20, 2024*

This release adjusts `reconnectBackOffBaseSeconds` to increase reconnect retries to up
1 day (~24 hours).

The `reconnectBackOffBaseSeconds` is increased by a bit over 10% (from
3.0 to 3.4 seconds) so this would not affect reconnection retries too
much.
