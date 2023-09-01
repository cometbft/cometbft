---
order: 2
---

# Creating a Data Companion for CometBFT

## Fetching data

The first step in a Data Companion workflow is to retrieve the data that you want to offload from the node

### Block

gRPC Block Service

### Block Results

gRPC Block Results Service

## Store data

The second step in a Data Companion workflow is to store the retrieved data on an external storage (e.g. database). Once
the data is stored externally, the data companion can let the node "know" that the data is node needed anymore on the node.
This can be influenced by the concept of a "retain_height"

## Prune data

The third step in a Data Companion workflow is to invoke the new gRPC APIs that can control the retain height value for the
data companion.
