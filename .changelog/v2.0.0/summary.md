_June 9, 2025_

This is a major release of CometBFT v2.0.0 that builds upon the improvements introduced in v1.0.0. This release includes several substantial changes and new features, including:

1. **Mempool Lanes**: A new feature that allows for prioritization of transactions in the mempool, enabling more efficient transaction processing. This includes configuration options, metrics, and application integration.

2. **Mempool DoG Protocol**: A new protocol for limiting the number of peers to which transactions are forwarded, optimizing gossip-related bandwidth consumption.

3. **Non-Proposer Vote Extensions**: Enhanced support for vote extensions beyond just the proposer, allowing for more flexible consensus mechanisms.

4. **Secp256k1-eth Cryptography**: Added support for Ethereum-compatible secp256k1 signatures.

5. **Improved P2P Layer**: Extracted TCP transport from P2P layer, improved connection handling, and added new APIs for channel management.

6. **Optimized Genesis Handling**: Reduced genesis chunks size, improved genesis file validation, and optimized genesis file chunking.

7. **Enhanced Logging**: Migrated to structured logging with slog, improving log readability and analysis capabilities.

8. **Security Fixes**: Several important security fixes, including protection against malicious peers causing nodes to get stuck in blocksync and improved block part validation.

9. **Performance Improvements**: Numerous optimizations including reduced marshalling overhead, improved bit array handling, and more efficient transaction handling in the mempool.

These changes aim to improve performance, security, and flexibility of CometBFT, while maintaining compatibility with existing applications where possible.

Please refer to the [upgrading guidelines](./UPGRADING.md) for more details on upgrading from the v1.x release series.

**NB: This version is still in development, which means that API-breaking changes might be introduced before the final release.** See [RELEASES.md](./RELEASES.md) for more information on the stability guarantees we provide for pre-releases.
