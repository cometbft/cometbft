## Forum Application

The **ABCI 2.0 Forum Application** is a demo application where users can come and post messages in a forum running on a
blockchain powered by [CometBFT](https://github.com/cometbft/cometbft) state machine replication engine.

- **Users**

   - Can post messages (by submitting transactions)
   - Can view all the message history (querying the blockchain)
   - Are banned if they post messages that contain curse words
   - Message rate is dynamically adjusted based on network performance metrics

## ABCI 2.0

**This application demonstrates the use of various [ABCI 2.0](https://docs.cometbft.com/v1.0/spec/abci/) methods such as:**

- PrepareProposal
- ProcessProposal
- FinalizeBlock
- ExtendVote
- VerifyVoteExtension
- Commit
- CheckTx
- Query

### Vote Extensions Usage

This application demonstrates a practical use case for Vote Extensions in ABCI 2.0:

- Validators collect and report performance metrics (CPU usage, memory usage, etc.) via vote extensions
- The application aggregates these metrics to determine network health
- Based on the aggregated metrics, the application dynamically adjusts forum parameters, such as:
  - Message rate limits for users 
  - Resource allocation for processing transactions
  - Moderation thresholds

This approach showcases how vote extensions can be used for real-time network governance and adaptive parameter adjustment, which is essential for blockchain applications that need to respond to changing network conditions.

To follow this tutorial, please check the [Introduction to ABCI 2.0](../../../docs/tutorials/forum-application/1.abci-intro.md) document.

> Many thanks to the original team for brainstorming and bringing forth this idea. Their original repo can be found [here](https://github.com/interchainio/forum)

