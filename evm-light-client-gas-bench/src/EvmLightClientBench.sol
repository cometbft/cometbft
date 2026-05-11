// SPDX-License-Identifier: MIT
pragma solidity =0.8.30;

contract EvmLightClientBench {
    uint256 private constant SECP256K1_N_DIV_2 = 0x7fffffffffffffffffffffffffffffff5d576e7357a4501ddfe92f46681b20a0;

    error AlreadyInitialized();
    error ClientFrozen();
    error ConsensusStateAlreadyExists();
    error ConsensusStateNotFound();
    error ExpiredTrustedState();
    error FutureHeaderTime();
    error HeaderHashMismatch();
    error InvalidCommit();
    error InvalidHeader();
    error InvalidMisbehaviour();
    error InvalidProof();
    error InvalidSignature();
    error InvalidValidatorSet();
    error DuplicateValidator();
    error LengthMismatch();
    error TooManyValidators();
    error ValidatorSetNotFound();

    struct ClientState {
        bytes chainId;
        uint64 revisionNumber;
        uint64 latestHeight;
        uint64 trustingPeriod;
        uint64 maxClockDrift;
        uint64 trustLevelNumerator;
        uint64 trustLevelDenominator;
        bool frozen;
    }

    struct ConsensusState {
        uint64 timestamp;
        bytes32 appHash;
        bytes32 nextValidatorsHash;
        uint64 processedHeight;
        uint64 processedTime;
        bool exists;
    }

    struct Header {
        uint64 revisionNumber;
        uint64 height;
        uint64 timestamp;
        bytes32 validatorsHash;
        bytes32 nextValidatorsHash;
        bytes32 appHash;
        bytes32 headerHash;
        bytes32 commitBlockHash;
        uint32 round;
        uint32 partSetTotal;
        bytes32 partSetHash;
    }

    struct CompactCommit {
        uint64 height;
        bytes32 blockHash;
        uint32 round;
        uint32 partSetTotal;
        bytes32 partSetHash;
        uint256 signerBitmap;
        bytes[] signatures;
        uint64[] timestamps;
    }

    struct PrebuiltVoteCommit {
        uint64 height;
        bytes32 blockHash;
        uint32 round;
        uint32 partSetTotal;
        bytes32 partSetHash;
        uint256 signerBitmap;
        bytes[] signatures;
        bytes[] voteSignBytes;
    }

    struct StoredValidatorSet {
        address[] validators;
        uint64[] powers;
        bool exists;
    }

    struct StoredBlsValidatorSet {
        bytes[] pubKeys;
        uint64[] powers;
        bool exists;
    }

    struct DecodedIcs23ExistenceProof {
        bytes key;
        bytes value;
        bytes leafPrefix;
        bytes[] innerPrefixes;
        bytes[] innerSuffixes;
    }

    struct Ics23LeafDecodeState {
        uint256 hashOp;
        uint256 prehashKey;
        uint256 prehashValue;
        uint256 lengthOp;
        bool seenHash;
        bool seenPrehashKey;
        bool seenPrehashValue;
        bool seenLength;
        bool seenPrefix;
    }

    struct MisbehaviourHeaderInput {
        uint64 trustedHeight;
        address[] trustedValidators;
        uint64[] trustedPowers;
        address[] untrustedValidators;
        uint64[] untrustedPowers;
        Header header;
        CompactCommit commit;
    }

    ClientState public clientState;
    bool public initialized;

    mapping(uint64 => ConsensusState) public consensusStates;
    uint64[] public consensusHeights;

    mapping(bytes32 => StoredValidatorSet) private validatorSets;
    mapping(bytes32 => StoredBlsValidatorSet) private blsValidatorSets;

    function initializeClient(ClientState calldata client, uint64 trustedHeight, ConsensusState calldata trustedState)
        external
    {
        if (initialized) revert AlreadyInitialized();
        initialized = true;
        clientState = client;
        _storeConsensusState(trustedHeight, trustedState);
    }

    function hashValidatorSetNative(address[] calldata validators, uint64[] calldata powers)
        external
        pure
        returns (bytes32)
    {
        return _hashValidatorSetNative(validators, powers);
    }

    function hashValidatorLeaves(bytes[] calldata leaves) external pure returns (bytes32) {
        return _hashByteSlices(leaves);
    }

    function hashSimpleValidatorSetEd25519(bytes32[] calldata pubKeys, uint64[] calldata powers)
        external
        pure
        returns (bytes32)
    {
        if (pubKeys.length != powers.length) revert LengthMismatch();
        if (pubKeys.length > 256) revert TooManyValidators();
        if (pubKeys.length == 0) return sha256("");

        bytes32[] memory leafHashes = new bytes32[](pubKeys.length);
        for (uint256 i; i < pubKeys.length; ++i) {
            bytes memory leaf = _simpleValidatorEd25519(pubKeys[i], powers[i]);
            leafHashes[i] = sha256(abi.encodePacked(bytes1(0x00), leaf));
        }
        return _hashRange(leafHashes, 0, leafHashes.length);
    }

    function hashSimpleValidatorSetSecp256k1(bytes[] calldata pubKeys, uint64[] calldata powers)
        external
        pure
        returns (bytes32)
    {
        if (pubKeys.length != powers.length) revert LengthMismatch();
        if (pubKeys.length > 256) revert TooManyValidators();
        if (pubKeys.length == 0) return sha256("");

        bytes32[] memory leafHashes = new bytes32[](pubKeys.length);
        for (uint256 i; i < pubKeys.length; ++i) {
            if (pubKeys[i].length != 33) revert InvalidValidatorSet();
            bytes memory leaf = _simpleValidatorSecp256k1(pubKeys[i], powers[i]);
            leafHashes[i] = sha256(abi.encodePacked(bytes1(0x00), leaf));
        }
        return _hashRange(leafHashes, 0, leafHashes.length);
    }

    function hashSimpleValidatorSetBls12381(bytes[] calldata pubKeys, uint64[] calldata powers)
        external
        pure
        returns (bytes32)
    {
        return _hashSimpleValidatorSetBls12381(pubKeys, powers);
    }

    function hashHeader(Header calldata header) external view returns (bytes32) {
        return _hashHeader(header);
    }

    function reconstructCanonicalVoteSignBytes(
        bytes calldata chainId,
        int64 height,
        int64 round,
        bytes32 blockHash,
        uint32 partSetTotal,
        bytes32 partSetHash,
        uint64 timestampNanos
    ) external pure returns (bytes memory) {
        return _canonicalVoteSignBytes(chainId, height, round, blockHash, partSetTotal, partSetHash, timestampNanos);
    }

    function hashCanonicalVoteSignBytes(
        bytes calldata chainId,
        int64 height,
        int64 round,
        bytes32 blockHash,
        uint32 partSetTotal,
        bytes32 partSetHash,
        uint64 timestampNanos
    ) external pure returns (bytes32) {
        return keccak256(
            _canonicalVoteSignBytes(chainId, height, round, blockHash, partSetTotal, partSetHash, timestampNanos)
        );
    }

    function verifyCanonicalVoteSecp256k1(
        address validator,
        bytes calldata signature,
        bytes calldata chainId,
        int64 height,
        int64 round,
        bytes32 blockHash,
        uint32 partSetTotal,
        bytes32 partSetHash,
        uint64 timestampNanos
    ) external pure returns (bool) {
        bytes32 digest = keccak256(
            _canonicalVoteSignBytes(chainId, height, round, blockHash, partSetTotal, partSetHash, timestampNanos)
        );
        return _recover(digest, signature) == validator;
    }

    function verifyCommitCompact(
        address[] calldata validators,
        uint64[] calldata powers,
        CompactCommit calldata commit,
        uint256 requiredPower
    ) external view returns (bool) {
        return _verifyCommitCompact(validators, powers, commit, requiredPower);
    }

    function verifyCommitPrebuiltVoteBytes(
        address[] calldata validators,
        uint64[] calldata powers,
        PrebuiltVoteCommit calldata commit,
        uint256 requiredPower
    ) external pure returns (bool) {
        return _verifyCommitPrebuiltVoteBytes(validators, powers, commit, requiredPower);
    }

    function verifyAdjacentUpdateCalldata(
        uint64 trustedHeight,
        address[] calldata trustedValidators,
        uint64[] calldata trustedPowers,
        address[] calldata untrustedValidators,
        uint64[] calldata untrustedPowers,
        Header calldata header,
        CompactCommit calldata commit,
        uint64 currentTimestamp
    ) external view returns (bool) {
        _verifyAdjacentUpdate(
            trustedHeight,
            trustedValidators,
            trustedPowers,
            untrustedValidators,
            untrustedPowers,
            header,
            commit,
            currentTimestamp
        );
        return true;
    }

    function updateAdjacentCalldata(
        uint64 trustedHeight,
        address[] calldata trustedValidators,
        uint64[] calldata trustedPowers,
        address[] calldata untrustedValidators,
        uint64[] calldata untrustedPowers,
        Header calldata header,
        CompactCommit calldata commit,
        uint64 currentTimestamp,
        uint64 processedHeight
    ) external {
        _verifyAdjacentUpdate(
            trustedHeight,
            trustedValidators,
            trustedPowers,
            untrustedValidators,
            untrustedPowers,
            header,
            commit,
            currentTimestamp
        );
        _storeHeaderConsensusState(header, processedHeight, currentTimestamp);
    }

    function verifyNonAdjacentUpdateCalldata(
        uint64 trustedHeight,
        address[] calldata trustedValidators,
        uint64[] calldata trustedPowers,
        address[] calldata untrustedValidators,
        uint64[] calldata untrustedPowers,
        Header calldata header,
        CompactCommit calldata commit,
        uint64 currentTimestamp
    ) external view returns (bool) {
        _verifyNonAdjacentUpdate(
            trustedHeight,
            trustedValidators,
            trustedPowers,
            untrustedValidators,
            untrustedPowers,
            header,
            commit,
            currentTimestamp
        );
        return true;
    }

    function updateNonAdjacentCalldata(
        uint64 trustedHeight,
        address[] calldata trustedValidators,
        uint64[] calldata trustedPowers,
        address[] calldata untrustedValidators,
        uint64[] calldata untrustedPowers,
        Header calldata header,
        CompactCommit calldata commit,
        uint64 currentTimestamp,
        uint64 processedHeight
    ) external {
        _verifyNonAdjacentUpdate(
            trustedHeight,
            trustedValidators,
            trustedPowers,
            untrustedValidators,
            untrustedPowers,
            header,
            commit,
            currentTimestamp
        );
        _storeHeaderConsensusState(header, processedHeight, currentTimestamp);
    }

    function storeValidatorSet(address[] calldata validators, uint64[] calldata powers)
        external
        returns (bytes32 hash)
    {
        _ensureUniqueValidators(validators);
        hash = _hashValidatorSetNative(validators, powers);
        StoredValidatorSet storage set = validatorSets[hash];
        delete set.validators;
        delete set.powers;
        for (uint256 i; i < validators.length; ++i) {
            set.validators.push(validators[i]);
            set.powers.push(powers[i]);
        }
        set.exists = true;
    }

    function storeBlsValidatorSet(
        bytes32 validatorSetHash,
        bytes[] calldata canonicalPubKeys,
        bytes[] calldata eip2537PubKeys,
        uint64[] calldata powers
    ) external returns (bytes32 hash) {
        if (canonicalPubKeys.length != eip2537PubKeys.length) revert LengthMismatch();
        if (canonicalPubKeys.length != powers.length) revert LengthMismatch();
        if (canonicalPubKeys.length > 256) revert TooManyValidators();
        if (canonicalPubKeys.length == 0) revert InvalidValidatorSet();
        if (validatorSetHash == bytes32(0)) revert InvalidValidatorSet();
        if (_hashSimpleValidatorSetBls12381(canonicalPubKeys, powers) != validatorSetHash) {
            revert InvalidValidatorSet();
        }

        hash = validatorSetHash;
        StoredBlsValidatorSet storage set = blsValidatorSets[hash];
        delete set.pubKeys;
        delete set.powers;
        for (uint256 i; i < eip2537PubKeys.length; ++i) {
            if (eip2537PubKeys[i].length != 128) revert InvalidValidatorSet();
            set.pubKeys.push(eip2537PubKeys[i]);
            set.powers.push(powers[i]);
        }
        set.exists = true;
    }

    function verifyBlsAggregateStoredValidatorSet(
        bytes32 validatorSetHash,
        uint256 signerBitmap,
        uint256 requiredPower,
        bytes calldata hashedMessage,
        bytes calldata aggregateSig
    ) external view returns (bool) {
        StoredBlsValidatorSet storage set = blsValidatorSets[validatorSetHash];
        if (!set.exists) revert ValidatorSetNotFound();
        (bytes memory aggregatePubKey, uint256 signedPower) = _aggregateBlsPubKeysStorage(set, signerBitmap);
        if (signedPower <= requiredPower) return false;
        return _verifyBlsAggregate(aggregatePubKey, hashedMessage, aggregateSig);
    }

    function verifyBlsAggregateCalldataValidatorSet(
        bytes[] calldata pubKeys,
        uint64[] calldata powers,
        uint256 signerBitmap,
        uint256 requiredPower,
        bytes calldata hashedMessage,
        bytes calldata aggregateSig
    ) external view returns (bool) {
        if (pubKeys.length != powers.length) revert LengthMismatch();
        if (pubKeys.length > 256) revert TooManyValidators();
        (bytes memory aggregatePubKey, uint256 signedPower) =
            _aggregateBlsPubKeysCalldata(pubKeys, powers, signerBitmap);
        if (signedPower <= requiredPower) return false;
        return _verifyBlsAggregate(aggregatePubKey, hashedMessage, aggregateSig);
    }

    function updateBlsStoredValidatorSet(
        bytes[] calldata canonicalPubKeys,
        bytes[] calldata eip2537PubKeys,
        uint64[] calldata powers,
        bytes32 expectedValidatorSetHash
    ) external returns (bytes32 hash) {
        hash = this.storeBlsValidatorSet(expectedValidatorSetHash, canonicalPubKeys, eip2537PubKeys, powers);
    }

    function updateAdjacentStoredValidatorSet(
        uint64 trustedHeight,
        Header calldata header,
        CompactCommit calldata commit,
        uint64 currentTimestamp,
        uint64 processedHeight
    ) external {
        _requireLiveClient();
        ConsensusState storage trusted = _trustedConsensusState(trustedHeight);
        if (header.height != trustedHeight + 1) revert InvalidHeader();
        _verifyHeaderBasics(trustedHeight, trusted, header, currentTimestamp);
        _verifyCommitMatchesHeader(header, commit);
        if (header.validatorsHash != trusted.nextValidatorsHash) revert InvalidValidatorSet();

        StoredValidatorSet storage set = validatorSets[trusted.nextValidatorsHash];
        if (!set.exists) revert ValidatorSetNotFound();
        uint256 requiredPower = (_totalPowerStorage(set.powers) * 2) / 3;
        if (!_verifyCommitCompactStorage(set, commit, requiredPower)) revert InvalidCommit();

        _storeHeaderConsensusState(header, processedHeight, currentTimestamp);
    }

    function submitMisbehaviour(
        MisbehaviourHeaderInput calldata input1,
        MisbehaviourHeaderInput calldata input2,
        uint64 currentTimestamp
    ) external returns (bool) {
        _verifyHeaderForMisbehaviour(
            input1.trustedHeight,
            input1.trustedValidators,
            input1.trustedPowers,
            input1.untrustedValidators,
            input1.untrustedPowers,
            input1.header,
            input1.commit,
            currentTimestamp
        );
        _verifyHeaderForMisbehaviour(
            input2.trustedHeight,
            input2.trustedValidators,
            input2.trustedPowers,
            input2.untrustedValidators,
            input2.untrustedPowers,
            input2.header,
            input2.commit,
            currentTimestamp
        );

        if (input1.header.height == input2.header.height) {
            if (input1.header.headerHash == input2.header.headerHash) revert InvalidMisbehaviour();
        } else if (input1.header.height > input2.header.height) {
            if (input1.header.timestamp > input2.header.timestamp) revert InvalidMisbehaviour();
        } else {
            if (input2.header.timestamp > input1.header.timestamp) revert InvalidMisbehaviour();
        }

        clientState.frozen = true;
        return true;
    }

    function seedConsensusStates(uint64 count, uint64 firstHeight, uint64 firstTimestamp, uint64 timestampStep)
        external
    {
        for (uint64 i; i < count; ++i) {
            uint64 height = firstHeight + i;
            ConsensusState memory state = ConsensusState({
                timestamp: firstTimestamp + i * timestampStep,
                appHash: sha256(abi.encodePacked("app", height)),
                nextValidatorsHash: sha256(abi.encodePacked("nextVals", height)),
                processedHeight: height + 1000,
                processedTime: firstTimestamp + i * timestampStep + 1,
                exists: true
            });
            _storeConsensusState(height, state);
        }
    }

    function pruneOldestExpiredConsensusState(uint64 currentTimestamp) external returns (bool pruned) {
        for (uint256 i; i < consensusHeights.length; ++i) {
            uint64 height = consensusHeights[i];
            ConsensusState storage state = consensusStates[height];
            if (state.exists && state.timestamp + clientState.trustingPeriod <= currentTimestamp) {
                delete consensusStates[height];
                return true;
            }
        }
        return false;
    }

    function pruneExpiredConsensusStatesBounded(uint64 currentTimestamp, uint256 maxPruned)
        external
        returns (uint256 pruned)
    {
        for (uint256 i; i < consensusHeights.length && pruned < maxPruned; ++i) {
            uint64 height = consensusHeights[i];
            ConsensusState storage state = consensusStates[height];
            if (state.exists && state.timestamp + clientState.trustingPeriod <= currentTimestamp) {
                delete consensusStates[height];
                ++pruned;
            }
        }
    }

    function pruneConsensusState(uint64 height) external returns (bool) {
        if (!consensusStates[height].exists) return false;
        delete consensusStates[height];
        return true;
    }

    function storeConsensusStateForTest(uint64 height, ConsensusState calldata state) external {
        _storeConsensusState(height, state);
    }

    function verifyMembershipProof(
        bytes32 root,
        bytes calldata key,
        bytes calldata value,
        bytes32[] calldata siblings,
        bool[] calldata siblingOnLeft
    ) external pure returns (bool) {
        if (siblings.length != siblingOnLeft.length) revert InvalidProof();

        bytes32 running = sha256(abi.encodePacked(bytes1(0x00), key, value));
        for (uint256 i; i < siblings.length; ++i) {
            running = siblingOnLeft[i]
                ? sha256(abi.encodePacked(bytes1(0x01), siblings[i], running))
                : sha256(abi.encodePacked(bytes1(0x01), running, siblings[i]));
        }

        return running == root;
    }

    // verifyIavlExistenceProof verifies an IAVL-shaped pre-parsed existence
    // proof — i.e. the membership half of ICS23 with the protobuf decoding
    // already done off-chain.
    //
    // SCOPE: this function is intentionally NOT a full ICS23 verifier. The
    // caller is expected to have decoded the cosmos.ics23.v1.ExistenceProof
    // protobuf container and passed in:
    //   - `leafPrefix`: the byte string `0x00 || varint(0) || varint(0) || varint(version) || sha256(key)`
    //     concatenated with the length prefix of the leaf-hash field, exactly as
    //     the IAVL ProofLeafOp would assemble it. The on-chain step then computes
    //     `sha256(leafPrefix || sha256(value))` to obtain the leaf hash.
    //   - `innerPrefixes[i]` / `innerSuffixes[i]`: for each ProofInnerOp, the bytes
    //     before and after the running child hash (encoding the height, size,
    //     version, and sibling hash for that level).
    //
    // What this measures: the on-chain hashing cost of an IAVL membership proof
    // at depth N once protobuf decoding is amortized off-chain (or done in a
    // separate verifier component). Gas reported by this benchmark therefore
    // represents the lower bound — a full ICS23-on-EVM verifier would add the
    // protobuf decoding overhead for each inner op, which is not modeled here.
    function verifyIavlExistenceProof(
        bytes32 root,
        bytes calldata value,
        bytes calldata leafPrefix,
        bytes[] calldata innerPrefixes,
        bytes[] calldata innerSuffixes
    ) external pure returns (bool) {
        if (innerPrefixes.length != innerSuffixes.length) revert InvalidProof();

        bytes32 running = sha256(abi.encodePacked(leafPrefix, sha256(value)));
        for (uint256 i; i < innerPrefixes.length; ++i) {
            running = sha256(abi.encodePacked(innerPrefixes[i], running, innerSuffixes[i]));
        }
        return running == root;
    }

    function decodeIcs23IavlExistenceProof(bytes calldata proof)
        external
        pure
        returns (uint256 depth, bytes32 keyHash)
    {
        DecodedIcs23ExistenceProof memory decoded = _decodeIcs23ExistenceProof(proof);
        return (decoded.innerPrefixes.length, keccak256(decoded.key));
    }

    function verifyIcs23IavlExistenceProof(
        bytes32 root,
        bytes calldata proof,
        bytes calldata expectedKey,
        bytes calldata expectedValue
    ) external pure returns (bool) {
        DecodedIcs23ExistenceProof memory decoded = _decodeIcs23ExistenceProof(proof);
        if (keccak256(decoded.key) != keccak256(expectedKey)) revert InvalidProof();
        if (keccak256(decoded.value) != keccak256(expectedValue)) revert InvalidProof();
        return _verifyDecodedIcs23IavlExistenceProof(root, decoded);
    }

    function verifyIcs23IavlExistenceProof(bytes32 root, bytes calldata proof) external pure returns (bool) {
        DecodedIcs23ExistenceProof memory decoded = _decodeIcs23ExistenceProof(proof);
        return _verifyDecodedIcs23IavlExistenceProof(root, decoded);
    }

    function _verifyAdjacentUpdate(
        uint64 trustedHeight,
        address[] calldata trustedValidators,
        uint64[] calldata trustedPowers,
        address[] calldata untrustedValidators,
        uint64[] calldata untrustedPowers,
        Header calldata header,
        CompactCommit calldata commit,
        uint64 currentTimestamp
    ) private view {
        _requireLiveClient();
        ConsensusState storage trusted = _trustedConsensusState(trustedHeight);
        if (header.height != trustedHeight + 1) revert InvalidHeader();
        _verifyHeaderBasics(trustedHeight, trusted, header, currentTimestamp);
        _verifyCommitMatchesHeader(header, commit);
        _ensureUniqueValidators(trustedValidators);
        _ensureUniqueValidators(untrustedValidators);

        if (_hashValidatorSetNative(trustedValidators, trustedPowers) != trusted.nextValidatorsHash) {
            revert InvalidValidatorSet();
        }
        if (_hashValidatorSetNative(untrustedValidators, untrustedPowers) != header.validatorsHash) {
            revert InvalidValidatorSet();
        }
        if (header.validatorsHash != trusted.nextValidatorsHash) revert InvalidValidatorSet();

        uint256 requiredPower = (_totalPower(untrustedPowers) * 2) / 3;
        if (!_verifyCommitCompact(untrustedValidators, untrustedPowers, commit, requiredPower)) revert InvalidCommit();
    }

    function _verifyNonAdjacentUpdate(
        uint64 trustedHeight,
        address[] calldata trustedValidators,
        uint64[] calldata trustedPowers,
        address[] calldata untrustedValidators,
        uint64[] calldata untrustedPowers,
        Header calldata header,
        CompactCommit calldata commit,
        uint64 currentTimestamp
    ) private view {
        _requireLiveClient();
        ConsensusState storage trusted = _trustedConsensusState(trustedHeight);
        if (header.height == trustedHeight + 1) revert InvalidHeader();
        _verifyHeaderBasics(trustedHeight, trusted, header, currentTimestamp);
        _verifyCommitMatchesHeader(header, commit);
        _ensureUniqueValidators(trustedValidators);
        _ensureUniqueValidators(untrustedValidators);

        if (_hashValidatorSetNative(trustedValidators, trustedPowers) != trusted.nextValidatorsHash) {
            revert InvalidValidatorSet();
        }
        if (_hashValidatorSetNative(untrustedValidators, untrustedPowers) != header.validatorsHash) {
            revert InvalidValidatorSet();
        }

        uint256 commitPower = (_totalPower(untrustedPowers) * 2) / 3;
        address[] memory signers =
            _verifyCommitAndCollectSigners(untrustedValidators, untrustedPowers, commit, commitPower);

        uint256 trustedTotal = _totalPower(trustedPowers);
        uint256 trustPower = (trustedTotal * clientState.trustLevelNumerator) / clientState.trustLevelDenominator;
        if (_signedPowerByAddress(signers, trustedValidators, trustedPowers) <= trustPower) revert InvalidCommit();
    }

    function _verifyHeaderForMisbehaviour(
        uint64 trustedHeight,
        address[] calldata trustedValidators,
        uint64[] calldata trustedPowers,
        address[] calldata untrustedValidators,
        uint64[] calldata untrustedPowers,
        Header calldata header,
        CompactCommit calldata commit,
        uint64 currentTimestamp
    ) private view {
        if (header.height == trustedHeight + 1) {
            _verifyAdjacentUpdate(
                trustedHeight,
                trustedValidators,
                trustedPowers,
                untrustedValidators,
                untrustedPowers,
                header,
                commit,
                currentTimestamp
            );
        } else {
            _verifyNonAdjacentUpdate(
                trustedHeight,
                trustedValidators,
                trustedPowers,
                untrustedValidators,
                untrustedPowers,
                header,
                commit,
                currentTimestamp
            );
        }
    }

    function _verifyHeaderBasics(
        uint64 trustedHeight,
        ConsensusState storage trusted,
        Header calldata header,
        uint64 currentTimestamp
    ) private view {
        if (header.revisionNumber != clientState.revisionNumber) revert InvalidHeader();
        if (header.height <= trustedHeight) revert InvalidHeader();
        if (header.timestamp <= trusted.timestamp) revert InvalidHeader();
        if (trusted.timestamp + clientState.trustingPeriod <= currentTimestamp) revert ExpiredTrustedState();
        if (header.timestamp >= currentTimestamp + clientState.maxClockDrift) revert FutureHeaderTime();
        if (_hashHeader(header) != header.headerHash) revert HeaderHashMismatch();
        if (header.commitBlockHash != header.headerHash) revert HeaderHashMismatch();
    }

    function _verifyCommitMatchesHeader(Header calldata header, CompactCommit calldata commit) private pure {
        if (commit.height != header.height) revert InvalidCommit();
        if (commit.blockHash != header.commitBlockHash) revert InvalidCommit();
        if (commit.round != header.round) revert InvalidCommit();
        if (commit.partSetTotal != header.partSetTotal) revert InvalidCommit();
        if (commit.partSetHash != header.partSetHash) revert InvalidCommit();
    }

    function _storeHeaderConsensusState(Header calldata header, uint64 processedHeight, uint64 processedTime) private {
        ConsensusState memory state = ConsensusState({
            timestamp: header.timestamp,
            appHash: header.appHash,
            nextValidatorsHash: header.nextValidatorsHash,
            processedHeight: processedHeight,
            processedTime: processedTime,
            exists: true
        });
        _storeNewConsensusState(header.height, state);
        if (header.height > clientState.latestHeight) {
            clientState.latestHeight = header.height;
        }
    }

    function _storeNewConsensusState(uint64 height, ConsensusState memory state) private {
        if (consensusStates[height].exists) revert ConsensusStateAlreadyExists();
        consensusHeights.push(height);
        consensusStates[height] = state;
    }

    function _storeConsensusState(uint64 height, ConsensusState memory state) private {
        if (!consensusStates[height].exists) {
            consensusHeights.push(height);
        }
        consensusStates[height] = state;
    }

    function _trustedConsensusState(uint64 trustedHeight) private view returns (ConsensusState storage trusted) {
        trusted = consensusStates[trustedHeight];
        if (!trusted.exists) revert ConsensusStateNotFound();
    }

    function _requireLiveClient() private view {
        if (clientState.frozen) revert ClientFrozen();
    }

    function _verifyCommitCompact(
        address[] calldata validators,
        uint64[] calldata powers,
        CompactCommit calldata commit,
        uint256 requiredPower
    ) private view returns (bool) {
        if (validators.length != powers.length) revert LengthMismatch();
        if (validators.length > 256) revert TooManyValidators();
        if (commit.signatures.length != commit.timestamps.length) revert LengthMismatch();

        uint256 signedPower;
        uint256 sigIndex;
        for (uint256 i; i < validators.length; ++i) {
            if ((commit.signerBitmap & (uint256(1) << i)) == 0) continue;

            bytes32 digest = keccak256(_compactVoteSignBytes(commit, commit.timestamps[sigIndex]));
            if (_recover(digest, commit.signatures[sigIndex]) != validators[i]) revert InvalidSignature();

            signedPower += powers[i];
            ++sigIndex;
        }

        if (sigIndex != commit.signatures.length) revert LengthMismatch();
        return signedPower > requiredPower;
    }

    function _verifyCommitPrebuiltVoteBytes(
        address[] calldata validators,
        uint64[] calldata powers,
        PrebuiltVoteCommit calldata commit,
        uint256 requiredPower
    ) private pure returns (bool) {
        if (validators.length != powers.length) revert LengthMismatch();
        if (validators.length > 256) revert TooManyValidators();
        if (commit.signatures.length != commit.voteSignBytes.length) revert LengthMismatch();

        uint256 signedPower;
        uint256 sigIndex;
        for (uint256 i; i < validators.length; ++i) {
            if ((commit.signerBitmap & (uint256(1) << i)) == 0) continue;

            if (_recover(keccak256(commit.voteSignBytes[sigIndex]), commit.signatures[sigIndex]) != validators[i]) {
                revert InvalidSignature();
            }

            signedPower += powers[i];
            ++sigIndex;
        }

        if (sigIndex != commit.signatures.length) revert LengthMismatch();
        return signedPower > requiredPower;
    }

    function _verifyCommitCompactStorage(
        StoredValidatorSet storage set,
        CompactCommit calldata commit,
        uint256 requiredPower
    ) private view returns (bool) {
        if (set.validators.length > 256) revert TooManyValidators();
        if (commit.signatures.length != commit.timestamps.length) revert LengthMismatch();

        uint256 signedPower;
        uint256 sigIndex;
        for (uint256 i; i < set.validators.length; ++i) {
            if ((commit.signerBitmap & (uint256(1) << i)) == 0) continue;

            bytes32 digest = keccak256(_compactVoteSignBytes(commit, commit.timestamps[sigIndex]));
            if (_recover(digest, commit.signatures[sigIndex]) != set.validators[i]) revert InvalidSignature();

            signedPower += set.powers[i];
            ++sigIndex;
        }

        if (sigIndex != commit.signatures.length) revert LengthMismatch();
        return signedPower > requiredPower;
    }

    function _verifyCommitAndCollectSigners(
        address[] calldata validators,
        uint64[] calldata powers,
        CompactCommit calldata commit,
        uint256 requiredPower
    ) private view returns (address[] memory signers) {
        if (validators.length != powers.length) revert LengthMismatch();
        if (validators.length > 256) revert TooManyValidators();
        if (commit.signatures.length != commit.timestamps.length) revert LengthMismatch();

        signers = new address[](commit.signatures.length);
        uint256 signedPower;
        uint256 sigIndex;
        for (uint256 i; i < validators.length; ++i) {
            if ((commit.signerBitmap & (uint256(1) << i)) == 0) continue;

            address recovered = _recover(
                keccak256(_compactVoteSignBytes(commit, commit.timestamps[sigIndex])), commit.signatures[sigIndex]
            );
            if (recovered != validators[i]) revert InvalidSignature();

            signers[sigIndex] = recovered;
            signedPower += powers[i];
            ++sigIndex;
        }

        if (sigIndex != commit.signatures.length) revert LengthMismatch();
        if (signedPower <= requiredPower) revert InvalidCommit();
    }

    function _signedPowerByAddress(address[] memory signers, address[] calldata validators, uint64[] calldata powers)
        private
        pure
        returns (uint256 signedPower)
    {
        if (validators.length != powers.length) revert LengthMismatch();
        if (validators.length > 256) revert TooManyValidators();

        for (uint256 i; i < validators.length; ++i) {
            for (uint256 j; j < signers.length; ++j) {
                if (validators[i] == signers[j]) {
                    signedPower += powers[i];
                    break;
                }
            }
        }
    }

    function _ensureUniqueValidators(address[] calldata validators) private pure {
        if (validators.length > 256) revert TooManyValidators();
        for (uint256 i; i < validators.length; ++i) {
            for (uint256 j = i + 1; j < validators.length; ++j) {
                if (validators[i] == validators[j]) revert DuplicateValidator();
            }
        }
    }

    function _compactVoteSignBytes(CompactCommit calldata commit, uint64 timestamp)
        private
        view
        returns (bytes memory)
    {
        return abi.encodePacked(
            clientState.chainId,
            commit.height,
            commit.round,
            commit.blockHash,
            commit.partSetTotal,
            commit.partSetHash,
            timestamp
        );
    }

    function _canonicalVoteSignBytes(
        bytes calldata chainId,
        int64 height,
        int64 round,
        bytes32 blockHash,
        uint32 partSetTotal,
        bytes32 partSetHash,
        uint64 timestampNanos
    ) private pure returns (bytes memory) {
        if (height < 0 || round < 0) revert InvalidHeader();

        bytes memory partSetHeader =
            abi.encodePacked(bytes1(0x08), _varint(partSetTotal), bytes1(0x12), _varint(32), partSetHash);
        bytes memory blockId = abi.encodePacked(
            bytes1(0x0a), _varint(32), blockHash, bytes1(0x12), _varint(uint64(partSetHeader.length)), partSetHeader
        );
        uint64 secondsSinceEpoch = timestampNanos / 1_000_000_000;
        uint64 nanos = timestampNanos % 1_000_000_000;
        bytes memory timestamp = _timestampProto(secondsSinceEpoch, nanos);

        bytes memory body = abi.encodePacked(
            bytes1(0x08),
            _varint(2),
            bytes1(0x11),
            _fixed64LE(uint64(height)),
            round == 0 ? bytes("") : abi.encodePacked(bytes1(0x19), _fixed64LE(uint64(round))),
            bytes1(0x22),
            _varint(uint64(blockId.length)),
            blockId,
            bytes1(0x2a),
            _varint(uint64(timestamp.length)),
            timestamp,
            bytes1(0x32),
            _varint(uint64(chainId.length)),
            chainId
        );

        return abi.encodePacked(_varint(uint64(body.length)), body);
    }

    function _timestampProto(uint64 secondsSinceEpoch, uint64 nanos) private pure returns (bytes memory) {
        bytes memory out;
        if (secondsSinceEpoch != 0) {
            out = abi.encodePacked(bytes1(0x08), _varint(secondsSinceEpoch));
        }
        if (nanos != 0) {
            out = abi.encodePacked(out, bytes1(0x10), _varint(nanos));
        }
        return out;
    }

    function _fixed64LE(uint64 v) private pure returns (bytes memory out) {
        out = new bytes(8);
        for (uint256 i; i < 8; ++i) {
            out[i] = bytes1(uint8(v >> (8 * i)));
        }
    }

    function _recover(bytes32 digest, bytes calldata sig) private pure returns (address) {
        if (sig.length != 65) revert InvalidSignature();

        bytes32 r;
        bytes32 s;
        uint8 v;
        assembly {
            r := calldataload(sig.offset)
            s := calldataload(add(sig.offset, 32))
            v := byte(0, calldataload(add(sig.offset, 64)))
        }

        if (uint256(s) > SECP256K1_N_DIV_2) revert InvalidSignature();
        if (v < 27) v += 27;
        if (v != 27 && v != 28) revert InvalidSignature();
        address recovered = ecrecover(digest, v, r, s);
        if (recovered == address(0)) revert InvalidSignature();
        return recovered;
    }

    function _hashHeader(Header calldata header) private view returns (bytes32) {
        return sha256(
            abi.encodePacked(
                clientState.chainId,
                header.revisionNumber,
                header.height,
                header.timestamp,
                header.validatorsHash,
                header.nextValidatorsHash,
                header.appHash,
                header.round,
                header.partSetTotal,
                header.partSetHash
            )
        );
    }

    function _hashValidatorSetNative(address[] calldata validators, uint64[] calldata powers)
        private
        pure
        returns (bytes32)
    {
        if (validators.length != powers.length) revert LengthMismatch();
        if (validators.length > 256) revert TooManyValidators();
        if (validators.length == 0) return sha256("");

        bytes32[] memory leaves = new bytes32[](validators.length);
        for (uint256 i; i < validators.length; ++i) {
            leaves[i] = sha256(abi.encodePacked(bytes1(0x00), validators[i], powers[i]));
        }
        return _hashRange(leaves, 0, leaves.length);
    }

    function _hashByteSlices(bytes[] calldata leaves) private pure returns (bytes32) {
        if (leaves.length == 0) return sha256("");

        bytes32[] memory leafHashes = new bytes32[](leaves.length);
        for (uint256 i; i < leaves.length; ++i) {
            leafHashes[i] = sha256(abi.encodePacked(bytes1(0x00), leaves[i]));
        }
        return _hashRange(leafHashes, 0, leafHashes.length);
    }

    function _hashRange(bytes32[] memory hashes, uint256 start, uint256 end) private pure returns (bytes32) {
        uint256 count = end - start;
        if (count == 1) return hashes[start];

        uint256 split = _splitPoint(count);
        bytes32 left = _hashRange(hashes, start, start + split);
        bytes32 right = _hashRange(hashes, start + split, end);
        return sha256(abi.encodePacked(bytes1(0x01), left, right));
    }

    function _splitPoint(uint256 count) private pure returns (uint256) {
        uint256 split = 1;
        while (split * 2 < count) {
            split *= 2;
        }
        return split;
    }

    uint256 private constant ICS23_MAX_PROOF_BYTES = 16_384;
    uint256 private constant ICS23_MAX_DEPTH = 32;

    function _verifyDecodedIcs23IavlExistenceProof(bytes32 root, DecodedIcs23ExistenceProof memory proof)
        private
        pure
        returns (bool)
    {
        bytes32 valueHash = sha256(proof.value);
        bytes32 running = sha256(
            abi.encodePacked(proof.leafPrefix, _varint(uint64(proof.key.length)), proof.key, bytes1(0x20), valueHash)
        );
        for (uint256 i; i < proof.innerPrefixes.length; ++i) {
            running = sha256(abi.encodePacked(proof.innerPrefixes[i], running, proof.innerSuffixes[i]));
        }
        return running == root;
    }

    function _decodeIcs23ExistenceProof(bytes calldata proof)
        private
        pure
        returns (DecodedIcs23ExistenceProof memory decoded)
    {
        if (proof.length == 0 || proof.length > ICS23_MAX_PROOF_BYTES) revert InvalidProof();
        uint256 pathCount = _countIcs23PathEntries(proof);
        if (pathCount == 0 || pathCount > ICS23_MAX_DEPTH) revert InvalidProof();
        decoded.innerPrefixes = new bytes[](pathCount);
        decoded.innerSuffixes = new bytes[](pathCount);

        bool seenKey;
        bool seenValue;
        bool seenLeaf;
        uint256 pathIndex;
        uint256 offset;
        while (offset < proof.length) {
            (uint256 field, uint256 wireType, uint256 next) = _readProtoKey(proof, offset, proof.length);
            offset = next;
            if (wireType == 2) {
                (uint256 valueOffset, uint256 valueLength, uint256 afterValue) =
                    _readLengthDelimited(proof, offset, proof.length);
                if (field == 1) {
                    if (seenKey) revert InvalidProof();
                    decoded.key = proof[valueOffset:valueOffset + valueLength];
                    seenKey = true;
                } else if (field == 2) {
                    if (seenValue) revert InvalidProof();
                    decoded.value = proof[valueOffset:valueOffset + valueLength];
                    seenValue = true;
                } else if (field == 3) {
                    if (seenLeaf) revert InvalidProof();
                    decoded.leafPrefix = _decodeIcs23LeafOp(proof, valueOffset, valueOffset + valueLength);
                    seenLeaf = true;
                } else if (field == 4) {
                    (decoded.innerPrefixes[pathIndex], decoded.innerSuffixes[pathIndex]) =
                        _decodeIcs23InnerOp(proof, valueOffset, valueOffset + valueLength);
                    ++pathIndex;
                }
                offset = afterValue;
            } else {
                offset = _skipProtoValue(proof, offset, proof.length, wireType);
            }
        }
        if (!seenKey || !seenValue || !seenLeaf || pathIndex != pathCount) revert InvalidProof();
    }

    function _countIcs23PathEntries(bytes calldata proof) private pure returns (uint256 count) {
        uint256 offset;
        while (offset < proof.length) {
            (uint256 field, uint256 wireType, uint256 next) = _readProtoKey(proof, offset, proof.length);
            offset = next;
            if (wireType == 2) {
                (,, uint256 afterValue) = _readLengthDelimited(proof, offset, proof.length);
                if (field == 4) ++count;
                offset = afterValue;
            } else {
                offset = _skipProtoValue(proof, offset, proof.length, wireType);
            }
        }
    }

    function _decodeIcs23LeafOp(bytes calldata data, uint256 start, uint256 end)
        private
        pure
        returns (bytes memory prefix)
    {
        Ics23LeafDecodeState memory state;
        uint256 offset = start;
        while (offset < end) {
            (uint256 field, uint256 wireType, uint256 next) = _readProtoKey(data, offset, end);
            offset = next;
            if (field <= 4 && wireType == 0) {
                (uint256 value, uint256 afterValue) = _readVarint(data, offset, end);
                offset = afterValue;
                if (field == 1) {
                    if (state.seenHash) revert InvalidProof();
                    state.hashOp = value;
                    state.seenHash = true;
                } else if (field == 2) {
                    if (state.seenPrehashKey) revert InvalidProof();
                    state.prehashKey = value;
                    state.seenPrehashKey = true;
                } else if (field == 3) {
                    if (state.seenPrehashValue) revert InvalidProof();
                    state.prehashValue = value;
                    state.seenPrehashValue = true;
                } else if (field == 4) {
                    if (state.seenLength) revert InvalidProof();
                    state.lengthOp = value;
                    state.seenLength = true;
                }
            } else if (field == 5 && wireType == 2) {
                if (state.seenPrefix) revert InvalidProof();
                (uint256 valueOffset, uint256 valueLength, uint256 afterValue) = _readLengthDelimited(data, offset, end);
                prefix = data[valueOffset:valueOffset + valueLength];
                state.seenPrefix = true;
                offset = afterValue;
            } else {
                offset = _skipProtoValue(data, offset, end, wireType);
            }
        }
        if (!state.seenHash || !state.seenPrehashValue || !state.seenLength || !state.seenPrefix) {
            revert InvalidProof();
        }
        if (state.hashOp != 1 || state.prehashKey != 0 || state.prehashValue != 1 || state.lengthOp != 1) {
            revert InvalidProof();
        }
        if (state.seenPrehashKey && state.prehashKey != 0) revert InvalidProof();
        if (prefix.length == 0 || prefix[0] != 0x00) revert InvalidProof();
    }

    function _decodeIcs23InnerOp(bytes calldata data, uint256 start, uint256 end)
        private
        pure
        returns (bytes memory prefix, bytes memory suffix)
    {
        bool seenHash;
        bool seenPrefix;
        bool seenSuffix;
        uint256 offset = start;
        while (offset < end) {
            (uint256 field, uint256 wireType, uint256 next) = _readProtoKey(data, offset, end);
            offset = next;
            if (field == 1 && wireType == 0) {
                if (seenHash) revert InvalidProof();
                (uint256 hashOp, uint256 afterValue) = _readVarint(data, offset, end);
                if (hashOp != 1) revert InvalidProof();
                seenHash = true;
                offset = afterValue;
            } else if ((field == 2 || field == 3) && wireType == 2) {
                (uint256 valueOffset, uint256 valueLength, uint256 afterValue) = _readLengthDelimited(data, offset, end);
                if (field == 2) {
                    if (seenPrefix) revert InvalidProof();
                    prefix = data[valueOffset:valueOffset + valueLength];
                    seenPrefix = true;
                } else {
                    if (seenSuffix) revert InvalidProof();
                    suffix = data[valueOffset:valueOffset + valueLength];
                    seenSuffix = true;
                }
                offset = afterValue;
            } else {
                offset = _skipProtoValue(data, offset, end, wireType);
            }
        }
        if (!seenHash || !seenPrefix) revert InvalidProof();
        if (prefix.length < 4 || prefix.length > 45) revert InvalidProof();
        if (prefix[0] == 0x00) revert InvalidProof();
        if (suffix.length != 0 && suffix.length != 33) revert InvalidProof();
        if (suffix.length == 33 && suffix[0] != 0x20) revert InvalidProof();
    }

    function _readProtoKey(bytes calldata data, uint256 offset, uint256 end)
        private
        pure
        returns (uint256 field, uint256 wireType, uint256 next)
    {
        (uint256 key, uint256 afterKey) = _readVarint(data, offset, end);
        field = key >> 3;
        wireType = key & 7;
        if (field == 0) revert InvalidProof();
        next = afterKey;
    }

    function _readLengthDelimited(bytes calldata data, uint256 offset, uint256 end)
        private
        pure
        returns (uint256 valueOffset, uint256 valueLength, uint256 next)
    {
        (valueLength, valueOffset) = _readVarint(data, offset, end);
        next = valueOffset + valueLength;
        if (next > end) revert InvalidProof();
    }

    function _readVarint(bytes calldata data, uint256 offset, uint256 end)
        private
        pure
        returns (uint256 value, uint256 next)
    {
        next = offset;
        uint256 shift;
        while (next < end && shift < 70) {
            uint8 b = uint8(data[next]);
            value |= uint256(b & 0x7f) << shift;
            ++next;
            if (b < 0x80) return (value, next);
            shift += 7;
        }
        revert InvalidProof();
    }

    function _skipProtoValue(bytes calldata data, uint256 offset, uint256 end, uint256 wireType)
        private
        pure
        returns (uint256 next)
    {
        if (wireType == 0) {
            (, next) = _readVarint(data, offset, end);
        } else if (wireType == 1) {
            next = offset + 8;
        } else if (wireType == 2) {
            (,, next) = _readLengthDelimited(data, offset, end);
        } else if (wireType == 5) {
            next = offset + 4;
        } else {
            revert InvalidProof();
        }
        if (next > end) revert InvalidProof();
    }

    address private constant BLS12_G1ADD = address(0x0b);
    address private constant BLS12_G2ADD = address(0x0d);
    address private constant BLS12_PAIRING_CHECK = address(0x0f);
    address private constant BLS12_MAP_FP2_TO_G2 = address(0x11);

    // benchBlsG1AddSignerAggregation simulates the public-key aggregation step of
    // a BLS aggregate signature verification. It chains G1ADD calls so that the
    // running point is updated on each iteration, matching how on-chain
    // aggregation across `signerCount` validators would work in production.
    function benchBlsG1AddSignerAggregation(uint256 signerCount) external view returns (bytes memory acc) {
        bytes memory generator = _blsG1Generator();
        acc = generator;
        bytes memory input = new bytes(256);
        for (uint256 i = 1; i < signerCount; ++i) {
            assembly {
                let dst := add(input, 0x20)
                mcopy(dst, add(acc, 0x20), 128)
                mcopy(add(dst, 128), add(generator, 0x20), 128)
            }
            (bool ok, bytes memory out) = BLS12_G1ADD.staticcall(input);
            if (!ok) revert InvalidSignature();
            acc = out;
        }
    }

    // benchBlsG2AddSignatureAggregation does the same chained accumulation for G2.
    function benchBlsG2AddSignatureAggregation(uint256 signerCount) external view returns (bytes memory acc) {
        bytes memory generator = _blsG2Generator();
        acc = generator;
        bytes memory input = new bytes(512);
        for (uint256 i = 1; i < signerCount; ++i) {
            assembly {
                let dst := add(input, 0x20)
                mcopy(dst, add(acc, 0x20), 256)
                mcopy(add(dst, 256), add(generator, 0x20), 256)
            }
            (bool ok, bytes memory out) = BLS12_G2ADD.staticcall(input);
            if (!ok) revert InvalidSignature();
            acc = out;
        }
    }

    // benchBlsPairingCheck issues one pairing-check precompile call equivalent to
    // verifying an aggregate BLS signature: e(-G1, sig) * e(P_agg, H(m)) == 1.
    // The test inputs are constructed so that the pairing equation holds for any
    // gas-only measurement: e(G1, G2) * e(-G1, G2) = 1.
    function benchBlsPairingCheck() external view returns (bool ok) {
        bytes memory g1 = _blsG1Generator();
        bytes memory g1Neg = _blsG1GeneratorNeg();
        bytes memory g2 = _blsG2Generator();
        bytes memory input = abi.encodePacked(g1, g2, g1Neg, g2);
        (bool success, bytes memory out) = BLS12_PAIRING_CHECK.staticcall(input);
        if (!success) revert InvalidSignature();
        ok = out.length == 32 && out[31] == 0x01;
    }

    // benchBlsMapFp2ToG2 invokes MAP_FP2_TO_G2 once with a deterministic FP2 input.
    // This is one of the two map calls inside a full hash-to-curve operation.
    function benchBlsMapFp2ToG2(bytes32 seed) external view returns (bytes memory g2) {
        bytes memory input = new bytes(128);
        bytes32 c0 = sha256(abi.encodePacked(seed, "c0"));
        bytes32 c1 = sha256(abi.encodePacked(seed, "c1"));
        for (uint256 i; i < 32; ++i) {
            input[i + 32] = c0[i];
            input[i + 96] = c1[i];
        }
        (bool ok, bytes memory out) = BLS12_MAP_FP2_TO_G2.staticcall(input);
        if (!ok) revert InvalidSignature();
        g2 = out;
    }

    // benchBlsAggregateApprox is a SYNTHETIC PRECOMPILE-COST APPROXIMATION of
    // an EIP-2537 aggregate BLS verify. It composes the three dominant on-chain
    // costs:
    //   - aggregate pubkeys: (signerCount - 1) G1ADD calls
    //   - aggregate signatures: (signerCount - 1) G2ADD calls
    //   - pairing check: 1 call (the always-true e(G1,X)*e(-G1,X)=1 pattern,
    //     used so the gas charge of the precompile is exercised independent of
    //     whether the input maps to a meaningful key/signature pair)
    //   - hash-to-G2: 1 MAP_FP2_TO_G2 call (the second map + G2ADD + cofactor
    //     clearing of a full hash-to-curve are omitted; they are second-order
    //     for total cost)
    //
    // What it does NOT do: it does NOT verify any real signature. Inputs are
    // generator/hash-derived placeholders so the pairing-check passes by
    // construction. Use this function for "what does aggregate verify cost on
    // EIP-2537?" gas accounting; use `verifyBlsAggregate` for actual signature
    // verification against fixture-generated aggregates.
    function benchBlsAggregateApprox(uint256 signerCount) external view returns (bool) {
        if (signerCount < 2) revert LengthMismatch();

        bytes memory g1 = _blsG1Generator();
        bytes memory g1Neg = _blsG1GeneratorNeg();
        bytes memory g2 = _blsG2Generator();

        bytes memory aggG1 = g1;
        bytes memory addG1Input = new bytes(256);
        for (uint256 i = 1; i < signerCount; ++i) {
            assembly {
                let dst := add(addG1Input, 0x20)
                mcopy(dst, add(aggG1, 0x20), 128)
                mcopy(add(dst, 128), add(g1, 0x20), 128)
            }
            (bool ok1, bytes memory aggG1Out) = BLS12_G1ADD.staticcall(addG1Input);
            if (!ok1) revert InvalidSignature();
            aggG1 = aggG1Out;
        }

        bytes memory aggG2 = g2;
        bytes memory addG2Input = new bytes(512);
        for (uint256 i = 1; i < signerCount; ++i) {
            assembly {
                let dst := add(addG2Input, 0x20)
                mcopy(dst, add(aggG2, 0x20), 256)
                mcopy(add(dst, 256), add(g2, 0x20), 256)
            }
            (bool ok2, bytes memory aggG2Out) = BLS12_G2ADD.staticcall(addG2Input);
            if (!ok2) revert InvalidSignature();
            aggG2 = aggG2Out;
        }

        bytes memory mapInput = new bytes(128);
        bytes32 c0 = sha256(abi.encodePacked("h2c-c0"));
        bytes32 c1 = sha256(abi.encodePacked("h2c-c1"));
        for (uint256 i; i < 32; ++i) {
            mapInput[i + 32] = c0[i];
            mapInput[i + 96] = c1[i];
        }
        (bool ok3,) = BLS12_MAP_FP2_TO_G2.staticcall(mapInput);
        if (!ok3) revert InvalidSignature();

        // Pairing check cost is independent of input magnitudes. To exercise the
        // full precompile path including a non-trivial G2 (the aggregated one) we
        // pair `g1 || aggG2 || g1Neg || aggG2`, which always reduces to 1
        // because e(G1, X) * e(-G1, X) = e(G1 - G1, X) = 1.
        bytes memory pairingInput = abi.encodePacked(g1, aggG2, g1Neg, aggG2);
        // Suppress the unused-variable warning for aggG1: it was produced to
        // charge the G1ADD gas; consume it in a no-op assembly block.
        assembly {
            pop(mload(aggG1))
        }
        (bool ok4, bytes memory pairOut) = BLS12_PAIRING_CHECK.staticcall(pairingInput);
        if (!ok4) revert InvalidSignature();
        return pairOut.length == 32 && pairOut[31] == 0x01;
    }

    function benchBlsMultiMessagePairing(uint256 signerCount) external view returns (bool) {
        if (signerCount == 0) revert LengthMismatch();
        bytes memory g1 = _blsG1Generator();
        bytes memory g1Neg = _blsG1GeneratorNeg();
        bytes memory g2 = _blsG2Generator();
        bytes memory input = new bytes((signerCount + 1) * 384);
        uint256 dst;
        assembly {
            dst := add(input, 0x20)
        }
        for (uint256 i; i < signerCount; ++i) {
            assembly {
                mcopy(add(dst, mul(i, 384)), add(g1, 0x20), 128)
                mcopy(add(add(dst, mul(i, 384)), 128), add(g2, 0x20), 256)
            }
        }
        assembly {
            let last := add(dst, mul(signerCount, 384))
            mcopy(last, add(g1Neg, 0x20), 128)
            mcopy(add(last, 128), add(g2, 0x20), 256)
        }
        (bool ok, bytes memory out) = BLS12_PAIRING_CHECK.staticcall(input);
        if (!ok) revert InvalidSignature();
        return out.length == 32 && out[31] == 0x01;
    }

    // PAIRING_CHECK_2_PAIRS_GAS_BOUND caps the gas forwarded to the EIP-2537
    // PAIRING_CHECK precompile in verifyBlsAggregate. The nominal cost of
    // PAIRING_CHECK with 2 pairs (the shape used here) is
    //   37_700 (base) + 2 * 32_600 (per-pair) = 102_900 gas
    // plus a few thousand gas of solidity copy/abi overhead. Bounded inputs
    // therefore complete well under 150k gas.
    //
    // The bound matters for adversarial inputs: when the precompile is given a
    // syntactically well-shaped 768-byte input that does not encode valid
    // curve points (a one-bit flip in the y-coordinate, for example), some
    // implementations are observed to consume effectively all forwarded gas
    // before signaling failure. Without a bound, that would let any caller
    // burn the full transaction gas by submitting one bad PAIRING_CHECK
    // input. The 200_000 gas cap is ~94% above the nominal happy-path cost,
    // generous enough to absorb implementation drift, and small enough that
    // a malformed input cannot be amplified into a denial of useful work.
    uint256 internal constant PAIRING_CHECK_2_PAIRS_GAS_BOUND = 200_000;

    // verifyBlsAggregate runs the on-chain half of a real BLS aggregate signature
    // verification. Off-chain (in the fixture generator) we have already aggregated
    // the signers' public keys via blst.P1Aggregate, aggregated their signatures
    // via blst.P2Aggregate, and computed `H(m)` by hashing the message to G2 with
    // cometbft's DST. This function then evaluates the pairing identity
    //
    //   e(-G1, aggSig) * e(aggPubKey, H(m)) == 1
    //
    // which is the standard min-pubkey-size BLS aggregate verify equation. Inputs
    // are all in EIP-2537 byte format: 128 bytes for G1 points, 256 bytes for G2.
    // Returns true iff the pairing precompile reports equality. Forwarded gas is
    // capped at PAIRING_CHECK_2_PAIRS_GAS_BOUND so malformed input cannot drain
    // the caller's full gas budget.
    function verifyBlsAggregate(bytes calldata aggPubKey, bytes calldata hashedMessage, bytes calldata aggSig)
        external
        view
        returns (bool)
    {
        return _verifyBlsAggregate(aggPubKey, hashedMessage, aggSig);
    }

    function _verifyBlsAggregate(bytes memory aggPubKey, bytes calldata hashedMessage, bytes calldata aggSig)
        private
        view
        returns (bool)
    {
        if (aggPubKey.length != 128) revert LengthMismatch();
        if (hashedMessage.length != 256) revert LengthMismatch();
        if (aggSig.length != 256) revert LengthMismatch();

        bytes memory g1Neg = _blsG1GeneratorNeg();
        bytes memory input = abi.encodePacked(g1Neg, aggSig, aggPubKey, hashedMessage);
        (bool ok, bytes memory out) = BLS12_PAIRING_CHECK.staticcall{gas: PAIRING_CHECK_2_PAIRS_GAS_BOUND}(input);
        if (!ok) return false;
        return out.length == 32 && out[31] == 0x01;
    }

    function _aggregateBlsPubKeysCalldata(bytes[] calldata pubKeys, uint64[] calldata powers, uint256 signerBitmap)
        private
        view
        returns (bytes memory aggregatePubKey, uint256 signedPower)
    {
        _rejectTrailingBitmapBits(signerBitmap, pubKeys.length);
        uint256 selected;
        for (uint256 i; i < pubKeys.length; ++i) {
            if ((signerBitmap & (uint256(1) << i)) == 0) continue;
            if (pubKeys[i].length != 128) revert InvalidValidatorSet();
            signedPower += powers[i];
            if (selected == 0) {
                aggregatePubKey = pubKeys[i];
            } else {
                aggregatePubKey = _blsG1Add(aggregatePubKey, pubKeys[i]);
            }
            ++selected;
        }
        if (selected == 0) revert InvalidCommit();
    }

    function _aggregateBlsPubKeysStorage(StoredBlsValidatorSet storage set, uint256 signerBitmap)
        private
        view
        returns (bytes memory aggregatePubKey, uint256 signedPower)
    {
        _rejectTrailingBitmapBits(signerBitmap, set.pubKeys.length);
        uint256 selected;
        for (uint256 i; i < set.pubKeys.length; ++i) {
            if ((signerBitmap & (uint256(1) << i)) == 0) continue;
            signedPower += set.powers[i];
            if (selected == 0) {
                aggregatePubKey = set.pubKeys[i];
            } else {
                aggregatePubKey = _blsG1Add(aggregatePubKey, set.pubKeys[i]);
            }
            ++selected;
        }
        if (selected == 0) revert InvalidCommit();
    }

    function _rejectTrailingBitmapBits(uint256 signerBitmap, uint256 validatorCount) private pure {
        if (validatorCount < 256 && (signerBitmap >> validatorCount) != 0) revert InvalidCommit();
    }

    function _blsG1Add(bytes memory left, bytes memory right) private view returns (bytes memory out) {
        bytes memory input = abi.encodePacked(left, right);
        (bool ok, bytes memory result) = BLS12_G1ADD.staticcall(input);
        if (!ok) revert InvalidSignature();
        out = result;
    }

    function _blsG1Generator() private pure returns (bytes memory g1) {
        g1 = new bytes(128);
        bytes32 xHi = bytes32(uint256(0x0000000000000000000000000000000017f1d3a73197d7942695638c4fa9ac0f));
        bytes32 xLo = bytes32(uint256(0xc3688c4f9774b905a14e3a3f171bac586c55e83ff97a1aeffb3af00adb22c6bb));
        bytes32 yHi = bytes32(uint256(0x0000000000000000000000000000000008b3f481e3aaa0f1a09e30ed741d8ae4));
        bytes32 yLo = bytes32(uint256(0xfcf5e095d5d00af600db18cb2c04b3edd03cc744a2888ae40caa232946c5e7e1));
        assembly {
            mstore(add(g1, 0x20), xHi)
            mstore(add(g1, 0x40), xLo)
            mstore(add(g1, 0x60), yHi)
            mstore(add(g1, 0x80), yLo)
        }
    }

    function _blsG1GeneratorNeg() private pure returns (bytes memory g1) {
        g1 = new bytes(128);
        bytes32 xHi = bytes32(uint256(0x0000000000000000000000000000000017f1d3a73197d7942695638c4fa9ac0f));
        bytes32 xLo = bytes32(uint256(0xc3688c4f9774b905a14e3a3f171bac586c55e83ff97a1aeffb3af00adb22c6bb));
        bytes32 yNegHi = bytes32(uint256(0x00000000000000000000000000000000114d1d6855d545a8aa7d76c8cf2e21f2));
        bytes32 yNegLo = bytes32(uint256(0x67816aef1db507c96655b9d5caac42364e6f38ba0ecb751bad54dcd6b939c2ca));
        assembly {
            mstore(add(g1, 0x20), xHi)
            mstore(add(g1, 0x40), xLo)
            mstore(add(g1, 0x60), yNegHi)
            mstore(add(g1, 0x80), yNegLo)
        }
    }

    function _blsG2Generator() private pure returns (bytes memory g2) {
        g2 = new bytes(256);
        bytes32 x0Hi = bytes32(uint256(0x00000000000000000000000000000000024aa2b2f08f0a91260805272dc51051));
        bytes32 x0Lo = bytes32(uint256(0xc6e47ad4fa403b02b4510b647ae3d1770bac0326a805bbefd48056c8c121bdb8));
        bytes32 x1Hi = bytes32(uint256(0x0000000000000000000000000000000013e02b6052719f607dacd3a088274f65));
        bytes32 x1Lo = bytes32(uint256(0x596bd0d09920b61ab5da61bbdc7f5049334cf11213945d57e5ac7d055d042b7e));
        bytes32 y0Hi = bytes32(uint256(0x000000000000000000000000000000000ce5d527727d6e118cc9cdc6da2e351a));
        bytes32 y0Lo = bytes32(uint256(0xadfd9baa8cbdd3a76d429a695160d12c923ac9cc3baca289e193548608b82801));
        bytes32 y1Hi = bytes32(uint256(0x000000000000000000000000000000000606c4a02ea734cc32acd2b02bc28b99));
        bytes32 y1Lo = bytes32(uint256(0xcb3e287e85a763af267492ab572e99ab3f370d275cec1da1aaa9075ff05f79be));
        assembly {
            mstore(add(g2, 0x20), x0Hi)
            mstore(add(g2, 0x40), x0Lo)
            mstore(add(g2, 0x60), x1Hi)
            mstore(add(g2, 0x80), x1Lo)
            mstore(add(g2, 0xa0), y0Hi)
            mstore(add(g2, 0xc0), y0Lo)
            mstore(add(g2, 0xe0), y1Hi)
            mstore(add(g2, 0x100), y1Lo)
        }
    }

    function _simpleValidatorEd25519(bytes32 pubKey, uint64 power) private pure returns (bytes memory) {
        return abi.encodePacked(bytes2(0x0A22), bytes2(0x0A20), pubKey, bytes1(0x10), _varint(power));
    }

    function _simpleValidatorSecp256k1(bytes calldata pubKey, uint64 power) private pure returns (bytes memory) {
        return abi.encodePacked(bytes2(0x0A23), bytes2(0x1221), pubKey, bytes1(0x10), _varint(power));
    }

    function _hashSimpleValidatorSetBls12381(bytes[] calldata pubKeys, uint64[] calldata powers)
        private
        pure
        returns (bytes32)
    {
        if (pubKeys.length != powers.length) revert LengthMismatch();
        if (pubKeys.length > 256) revert TooManyValidators();
        if (pubKeys.length == 0) return sha256("");

        bytes32[] memory leafHashes = new bytes32[](pubKeys.length);
        for (uint256 i; i < pubKeys.length; ++i) {
            if (pubKeys[i].length != 96) revert InvalidValidatorSet();
            bytes memory leaf = _simpleValidatorBls12381(pubKeys[i], powers[i]);
            leafHashes[i] = sha256(abi.encodePacked(bytes1(0x00), leaf));
        }
        return _hashRange(leafHashes, 0, leafHashes.length);
    }

    // BLS12-381 SimpleValidator wire format. cometbft serializes the BLS pubkey
    // as the 96-byte uncompressed G1 form (blst.P1Affine.Serialize), so the
    // outer PublicKey message is 98 bytes (`1a60` || 96 pubkey bytes) and the
    // SimpleValidator wraps it with `0a62` || 98-byte inner message.
    function _simpleValidatorBls12381(bytes calldata pubKey, uint64 power) private pure returns (bytes memory) {
        return abi.encodePacked(bytes2(0x0A62), bytes2(0x1A60), pubKey, bytes1(0x10), _varint(power));
    }

    function _varint(uint64 v) private pure returns (bytes memory out) {
        if (v < 0x80) return abi.encodePacked(uint8(v));
        bytes memory buf = new bytes(10);
        uint256 n;
        while (v >= 0x80) {
            buf[n++] = bytes1(uint8(v) | 0x80);
            v >>= 7;
        }
        buf[n++] = bytes1(uint8(v));
        out = new bytes(n);
        for (uint256 i; i < n; ++i) {
            out[i] = buf[i];
        }
    }

    function _totalPower(uint64[] calldata powers) private pure returns (uint256 total) {
        for (uint256 i; i < powers.length; ++i) {
            total += powers[i];
        }
    }

    function _totalPowerStorage(uint64[] storage powers) private view returns (uint256 total) {
        for (uint256 i; i < powers.length; ++i) {
            total += powers[i];
        }
    }
}
