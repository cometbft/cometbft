*October 24, 2024*

This patch release addresses the issue where tx_search was not returning all results, which only arises when upgrading
to CometBFT-DB version 0.13 or later. It includes a fix in the state indexer to resolve this problem. We recommend
upgrading to this patch release if you are affected by this issue.


