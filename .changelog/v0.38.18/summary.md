<!--
    Add a summary for the release here.

    If you don't change this message, or if this file is empty, the release
    will not be created. -->
This patch release adds metrics for precommit data. Specifically, cometBFT v0.38.18 will now emit metrics for the amount of time passed between a proposal and the node receiving 2/3+ precommits. Additionally, cometBFT will now emit metrics for how many precommits the node received, and what percent stake those precommits make up, within the timeout commit period. Additionally, a reindex command was added to the CLI
