<!--
Guiding Principles:

Changelogs are for humans, not machines.
There should be an entry for every single version.
The same types of changes should be grouped.
Versions and sections should be linkable.
The latest version comes first.
The release date of each version is displayed.
Mention whether you follow Semantic Versioning.

Usage:

Change log entries are to be added to the Unreleased section under the
appropriate stanza (see below). Each entry should ideally include a tag and
the Github issue reference in the following format:

* (<tag>) \#<issue-number> message

The issue numbers will later be link-ified during the release process so you do
not have to worry about including a link manually, but you can if you wish.

Types of changes (Stanzas):

"Features" for new features.
"Improvements" for changes in existing functionality.
"Deprecated" for soon-to-be removed features.
"Bug Fixes" for any bug fixes.
"Client Breaking" for breaking CLI commands and REST routes.
"State Machine Breaking" for breaking the AppState

Ref: https://keepachangelog.com/en/1.0.0/
-->

# Changelog

## [v0.34.35-alpha.agoric.1]

* Backport [`cometbft/cometbft v0.37.15`](https://github.com/cometbft/cometbft/compare/v0.37.14..v0.37.15).

## [v0.34.30-alpha.agoric.1]

* Merge `cometbft/cometbft v0.34.30`.

## [v0.34.27-alpha.agoric.3]

* Merge `agoric-labs/tendermint` v0.34.23-alpha.agoric.4.

## [v0.34.23-alpha.agoric.4]

* Lower default `BlockParams.MaxBytes` to 5MB to mitigate asa-2023-002 

## [v0.34.27-alpha.agoric.2]

* Merge `tendermint/tendermint` v0.34.27.

## [v0.34.23-alpha.agoric.3]

* Agoric/agoric-sdk\#6945 Cherrypick fix for informalsystems/tendermint#4.

## [v0.34.23-alpha.agoric.2]

* Adapt to new callback tracking. See tendermint/tendermint#8331

## [v0.34.23-alpha.agoric.1]

* Agoric/agoric-sdk\#6305 Merge `tendermint/tendermint` v0.34.23

## [v0.34.21-alpha.agoric.1]

* Agoric/agoric-sdk\#6305 Merge `tendermint/tendermint` v0.34.21.

## [v0.34.14-alpha.agoric.1]

* Merge `tendermint/tendermint` v0.34.23.
* Add committing client for greater query concurrency.
