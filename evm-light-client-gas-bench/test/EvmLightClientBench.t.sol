// SPDX-License-Identifier: MIT
pragma solidity =0.8.30;

import "../src/EvmLightClientBench.sol";

interface Vm {
    function readFile(string calldata path) external view returns (string memory);
    function parseJson(string calldata json, string calldata key) external pure returns (bytes memory);
    function expectRevert(bytes4 selector) external;
    function deployCode(string calldata what) external returns (address deployed);
}

interface IEd25519VerifierBench {
    function verify(bytes32 publicKey, bytes memory signature, bytes memory message) external pure returns (bool);
}

contract EvmLightClientBenchTest {
    Vm private constant vm = Vm(address(uint160(uint256(keccak256("hevm cheat code")))));

    event CalldataGasEstimate(string name, uint256 byteLength, uint256 gasCost);
    // AnalyticalGasEstimate carries gas figures that come from analysis or from
    // public reference implementations rather than from on-chain execution in
    // this benchmark. Emitting them as events keeps them next to the measured
    // numbers in the test output so the report can quote a single source.
    event AnalyticalGasEstimate(string name, uint256 lowGas, uint256 highGas, string source);
    // CalldataCostBreakdown reports the full picture for a piece of calldata:
    //  - byteLength: raw call payload size
    //  - zeroBytes / nonzeroBytes: distribution of zero vs non-zero bytes
    //  - gasL1Standard: legacy 4/16 calldata cost (zero=4, nonzero=16)
    //  - gasL1FloorEIP7623: post-Pectra floor (zero=10, nonzero=40); the actual
    //    L1 fee for a tx is max(standard+execution, intrinsic+floor), so when
    //    most of the cost is data this floor can dominate
    //  - blobBytes: blob-equivalent footprint = ceil(byteLength/31)*32, which
    //    accounts for the 31-byte usable payload per 32-byte BLS12-381 field
    //    element used by EIP-4844 blob encoding. This is the data volume that
    //    would be charged via blob gas if the payload were posted as a blob.
    event CalldataCostBreakdown(
        string name,
        uint256 byteLength,
        uint256 zeroBytes,
        uint256 nonzeroBytes,
        uint256 gasL1Standard,
        uint256 gasL1FloorEIP7623,
        uint256 blobBytes
    );

    struct ClientFixture {
        bytes chainId;
        uint64 revisionNumber;
        uint64 latestHeight;
        uint64 trustingPeriod;
        uint64 maxClockDrift;
        uint64 trustLevelNumerator;
        uint64 trustLevelDenominator;
        uint64 currentTimestamp;
        uint64 processedHeight;
    }

    struct ValidatorFixture {
        address[] validators;
        uint64[] powers;
        bytes[] leaves;
        bytes32 hash;
    }

    struct MembershipFixture {
        bytes32 root;
        bytes key;
        bytes value;
        bytes32[] siblings;
        bool[] siblingOnLeft;
    }

    struct CanonicalEd25519Fixture {
        bytes32[] pubKeys;
        uint64[] powers;
        bytes[] leaves;
        bytes32 hash;
        bytes[] signatures;
    }

    struct CanonicalSecp256k1Fixture {
        bytes[] pubKeys;
        uint64[] powers;
        bytes[] leaves;
        bytes32 hash;
        bytes[] signatures;
    }

    struct CanonicalVoteFixture {
        bytes chainId;
        int64 height;
        int64 round;
        bytes32 blockHash;
        uint32 partSetTotal;
        bytes32 partSetHash;
        int64[] timestamps;
        bytes[] signBytes;
        bytes[] ethSignatures;
    }

    struct CanonicalBlsFixture {
        bool available;
        bytes[] pubKeys;
        bytes[] pubKeysEip2537;
        uint64[] powers;
        bytes[] leaves;
        bytes32 hash;
        bytes aggregateSig;
        bytes aggregatePubKey;
        bytes hashedMessage;
        bytes wrongMessageHashed;
        bytes missingSignerAggregateSig;
        uint256 signerBitmap;
        bytes message;
    }

    struct IavlFixture {
        bytes32 root;
        bytes key;
        bytes value;
        bytes leafPrefix;
        bytes[] innerPrefixes;
        bytes[] innerSuffixes;
        bytes existenceProof;
        bytes commitmentProof;
    }

    function testValidatorSetNativeHash50Gas() public {
        EvmLightClientBench bench = new EvmLightClientBench();
        string memory json = readFixture("update_50_equal");
        ValidatorFixture memory vals = validatorFixture(json, ".trustedValidators");
        require(bench.hashValidatorSetNative(vals.validators, vals.powers) == vals.hash, "validator hash mismatch");
    }

    function testPrebuiltValidatorLeavesHashGas() public {
        EvmLightClientBench bench = new EvmLightClientBench();
        string memory json = readFixture("update_50_equal");
        ValidatorFixture memory vals = validatorFixture(json, ".trustedValidators");
        require(bench.hashValidatorLeaves(vals.leaves) == vals.hash, "leaf hash mismatch");
    }

    function testCanonicalEd25519PrebuiltLeavesParity() public {
        EvmLightClientBench bench = new EvmLightClientBench();
        string memory json = readFixture("update_50_equal");
        CanonicalEd25519Fixture memory ed = canonicalEd25519Fixture(json);
        require(bench.hashValidatorLeaves(ed.leaves) == ed.hash, "ed25519 prebuilt leaves mismatch");
    }

    function testCanonicalSecp256k1PrebuiltLeavesParity() public {
        EvmLightClientBench bench = new EvmLightClientBench();
        string memory json = readFixture("update_50_equal");
        CanonicalSecp256k1Fixture memory sk = canonicalSecp256k1Fixture(json);
        require(bench.hashValidatorLeaves(sk.leaves) == sk.hash, "secp256k1 prebuilt leaves mismatch");
    }

    function testReconstructedEd25519SimpleValidatorHashGas() public {
        EvmLightClientBench bench = new EvmLightClientBench();
        string memory json = readFixture("update_50_equal");
        CanonicalEd25519Fixture memory ed = canonicalEd25519Fixture(json);
        bytes32 reconstructed = bench.hashSimpleValidatorSetEd25519(ed.pubKeys, ed.powers);
        require(reconstructed == ed.hash, "reconstructed ed25519 hash mismatch");
    }

    function testReconstructedSecp256k1SimpleValidatorHashGas() public {
        EvmLightClientBench bench = new EvmLightClientBench();
        string memory json = readFixture("update_50_equal");
        CanonicalSecp256k1Fixture memory sk = canonicalSecp256k1Fixture(json);
        bytes32 reconstructed = bench.hashSimpleValidatorSetSecp256k1(sk.pubKeys, sk.powers);
        require(reconstructed == sk.hash, "reconstructed secp256k1 hash mismatch");
    }

    function testCanonicalAndCompactValidatorHashesAreDistinct() public {
        EvmLightClientBench bench = new EvmLightClientBench();
        string memory json = readFixture("update_50_equal");
        ValidatorFixture memory compact = validatorFixture(json, ".trustedValidators");
        CanonicalEd25519Fixture memory ed = canonicalEd25519Fixture(json);
        require(compact.hash != ed.hash, "compact equals canonical hash unexpectedly");
        // The canonical reconstructed hash matches its own canonical hash; recompute to keep
        // this test independent of the prebuilt parity test.
        require(bench.hashSimpleValidatorSetEd25519(ed.pubKeys, ed.powers) == ed.hash, "ed25519 reconstruction broken");
    }

    function testHeaderHashGas() public {
        (EvmLightClientBench bench, string memory json) = initializedBench("update_50_equal");
        EvmLightClientBench.Header memory header = headerFixture(json, ".adjacent.header");
        require(bench.hashHeader(header) == header.headerHash, "header hash mismatch");
    }

    function testCommitCompact10Gas() public {
        assertCompactCommit("update_10_equal");
    }

    function testCommitCompact50Gas() public {
        assertCompactCommit("update_50_equal");
    }

    function testCommitCompact100Gas() public {
        assertCompactCommit("update_100_equal");
    }

    function testCommitCompact175Gas() public {
        assertCompactCommit("update_175_equal");
    }

    function testCommitPrebuiltVoteBytes50Gas() public {
        EvmLightClientBench bench = new EvmLightClientBench();
        string memory json = readFixture("update_50_equal");
        ValidatorFixture memory vals = validatorFixture(json, ".trustedValidators");
        EvmLightClientBench.PrebuiltVoteCommit memory commit = prebuiltCommitFixture(json, ".adjacent.commit");
        require(
            bench.verifyCommitPrebuiltVoteBytes(
                vals.validators, vals.powers, commit, uintFixture(json, ".adjacent.commit.requiredPower")
            ),
            "prebuilt commit failed"
        );
    }

    function testAdjacentUpdate10ValidatorsGas() public {
        (EvmLightClientBench bench, string memory json) = initializedBench("update_10_equal");
        ValidatorFixture memory vals = validatorFixture(json, ".trustedValidators");
        EvmLightClientBench.Header memory header = headerFixture(json, ".adjacent.header");
        EvmLightClientBench.CompactCommit memory commit = compactCommitFixture(json, ".adjacent.commit");
        ClientFixture memory client = clientFixture(json);

        bench.updateAdjacentCalldata(
            trustedHeight(json),
            vals.validators,
            vals.powers,
            vals.validators,
            vals.powers,
            header,
            commit,
            client.currentTimestamp,
            client.processedHeight
        );
    }

    function testAdjacentUpdateWithoutStorageGas() public {
        (EvmLightClientBench bench, string memory json) = initializedBench("update_50_equal");
        ValidatorFixture memory vals = validatorFixture(json, ".trustedValidators");
        EvmLightClientBench.Header memory header = headerFixture(json, ".adjacent.header");
        EvmLightClientBench.CompactCommit memory commit = compactCommitFixture(json, ".adjacent.commit");
        ClientFixture memory client = clientFixture(json);

        require(
            bench.verifyAdjacentUpdateCalldata(
                trustedHeight(json),
                vals.validators,
                vals.powers,
                vals.validators,
                vals.powers,
                header,
                commit,
                client.currentTimestamp
            ),
            "adjacent verify failed"
        );
    }

    function testAdjacentUpdateWithStorageGas() public {
        (EvmLightClientBench bench, string memory json) = initializedBench("update_50_equal");
        ValidatorFixture memory vals = validatorFixture(json, ".trustedValidators");
        EvmLightClientBench.Header memory header = headerFixture(json, ".adjacent.header");
        EvmLightClientBench.CompactCommit memory commit = compactCommitFixture(json, ".adjacent.commit");
        ClientFixture memory client = clientFixture(json);

        bench.updateAdjacentCalldata(
            trustedHeight(json),
            vals.validators,
            vals.powers,
            vals.validators,
            vals.powers,
            header,
            commit,
            client.currentTimestamp,
            client.processedHeight
        );

        (uint64 timestamp, bytes32 appHash, bytes32 nextValidatorsHash,,, bool exists) =
            bench.consensusStates(header.height);
        require(exists, "consensus state missing");
        require(timestamp == header.timestamp, "timestamp mismatch");
        require(appHash == header.appHash, "app hash mismatch");
        require(nextValidatorsHash == header.nextValidatorsHash, "next validators hash mismatch");
    }

    function testAdjacentMixedPowerUpdateGas() public {
        (EvmLightClientBench bench, string memory json) = initializedBench("update_50_mixed");
        ValidatorFixture memory vals = validatorFixture(json, ".trustedValidators");
        EvmLightClientBench.Header memory header = headerFixture(json, ".adjacent.header");
        EvmLightClientBench.CompactCommit memory commit = compactCommitFixture(json, ".adjacent.commit");
        ClientFixture memory client = clientFixture(json);

        bench.updateAdjacentCalldata(
            trustedHeight(json),
            vals.validators,
            vals.powers,
            vals.validators,
            vals.powers,
            header,
            commit,
            client.currentTimestamp,
            client.processedHeight
        );
    }

    function testNonAdjacentChangedValidatorSetGas() public {
        (EvmLightClientBench bench, string memory json) = initializedBench("misbehaviour_50_equal");
        ValidatorFixture memory trusted = validatorFixture(json, ".trustedValidators");
        ValidatorFixture memory untrusted = validatorFixture(json, ".untrustedValidators");
        EvmLightClientBench.Header memory header = headerFixture(json, ".nonAdjacent.header");
        EvmLightClientBench.CompactCommit memory commit = compactCommitFixture(json, ".nonAdjacent.commit");
        ClientFixture memory client = clientFixture(json);

        bench.updateNonAdjacentCalldata(
            trustedHeight(json),
            trusted.validators,
            trusted.powers,
            untrusted.validators,
            untrusted.powers,
            header,
            commit,
            client.currentTimestamp,
            client.processedHeight
        );
    }

    function testStoredValidatorSetAdjacentUpdateGas() public {
        (EvmLightClientBench bench, string memory json) = initializedBench("update_50_equal");
        ValidatorFixture memory vals = validatorFixture(json, ".trustedValidators");
        EvmLightClientBench.Header memory header = headerFixture(json, ".adjacent.header");
        EvmLightClientBench.CompactCommit memory commit = compactCommitFixture(json, ".adjacent.commit");
        ClientFixture memory client = clientFixture(json);

        bytes32 hash = bench.storeValidatorSet(vals.validators, vals.powers);
        require(hash == vals.hash, "stored set hash mismatch");
        bench.updateAdjacentStoredValidatorSet(
            trustedHeight(json), header, commit, client.currentTimestamp, client.processedHeight
        );
    }

    function testSameHeightMisbehaviour10ValidatorsGas() public {
        assertSameHeightMisbehaviour("misbehaviour_10_equal");
    }

    function testSameHeightMisbehaviourFreezesClientGas() public {
        assertSameHeightMisbehaviour("update_50_equal");
    }

    function assertSameHeightMisbehaviour(string memory fixtureName) private {
        (EvmLightClientBench bench, string memory json) = initializedBench(fixtureName);
        ValidatorFixture memory vals = validatorFixture(json, ".trustedValidators");
        ClientFixture memory client = clientFixture(json);

        bench.submitMisbehaviour(
            misbehaviourInput(json, vals, vals, ".adjacent"),
            misbehaviourInput(json, vals, vals, ".conflict"),
            client.currentTimestamp
        );
        (,,,,,,, bool frozen) = bench.clientState();
        require(frozen, "client not frozen");
    }

    function testBftTimeViolationMisbehaviour10ValidatorsGas() public {
        assertTimeViolation("misbehaviour_10_equal", false);
    }

    function testBftTimeViolationMisbehaviourGas() public {
        assertTimeViolation("update_50_equal", false);
    }

    function testBftTimeViolationMisbehaviourReversedOrderGas() public {
        assertTimeViolation("update_50_equal", true);
    }

    function testChangedValidatorSetMisbehaviour10ValidatorsGas() public {
        assertChangedValidatorSetMisbehaviour("misbehaviour_10_equal");
    }

    function testChangedValidatorSetMisbehaviourGas() public {
        assertChangedValidatorSetMisbehaviour("misbehaviour_50_equal");
    }

    function assertChangedValidatorSetMisbehaviour(string memory fixtureName) private {
        (EvmLightClientBench bench, string memory json) = initializedBench(fixtureName);
        ValidatorFixture memory trusted = validatorFixture(json, ".trustedValidators");
        ValidatorFixture memory untrusted = validatorFixture(json, ".untrustedValidators");
        ClientFixture memory client = clientFixture(json);

        bench.submitMisbehaviour(
            misbehaviourInput(json, trusted, untrusted, ".nonAdjacent"),
            misbehaviourInput(json, trusted, untrusted, ".nonAdjacentConflict"),
            client.currentTimestamp
        );
    }

    function testPruneOldestExpiredConsensusStateGas() public {
        (EvmLightClientBench bench, string memory json) = initializedBench("update_50_equal");
        ClientFixture memory client = clientFixture(json);
        bench.seedConsensusStates(16, 100, 1_700_000_000_000_000_000, 1_000_000_000);
        require(
            bench.pruneOldestExpiredConsensusState(client.currentTimestamp + client.trustingPeriod + 1), "not pruned"
        );
    }

    function testPruneBoundedExpiredConsensusStatesGas() public {
        (EvmLightClientBench bench, string memory json) = initializedBench("update_50_equal");
        ClientFixture memory client = clientFixture(json);
        bench.seedConsensusStates(16, 100, 1_700_000_000_000_000_000, 1_000_000_000);
        require(bench.pruneExpiredConsensusStatesBounded(client.currentTimestamp + client.trustingPeriod + 1, 4) == 4);
    }

    function testExplicitPruneGas() public {
        (EvmLightClientBench bench, string memory json) = initializedBench("update_50_equal");
        bench.storeConsensusStateForTest(200, trustedConsensusState(json));
        require(bench.pruneConsensusState(200), "explicit prune failed");
    }

    function testMembershipProofGas() public {
        EvmLightClientBench bench = new EvmLightClientBench();
        string memory json = readFixture("update_50_equal");
        MembershipFixture memory proof = membershipFixture(json);
        require(bench.verifyMembershipProof(proof.root, proof.key, proof.value, proof.siblings, proof.siblingOnLeft));
    }

    function testIavlExistenceProofGas() public {
        EvmLightClientBench bench = new EvmLightClientBench();
        string memory json = readFixture("update_50_equal");
        IavlFixture memory iavl = iavlFixture(json);
        require(
            bench.verifyIavlExistenceProof(
                iavl.root, iavl.value, iavl.leafPrefix, iavl.innerPrefixes, iavl.innerSuffixes
            ),
            "iavl proof failed"
        );
    }

    function testIavlExistenceProofRejectsWrongValue() public {
        EvmLightClientBench bench = new EvmLightClientBench();
        string memory json = readFixture("update_50_equal");
        IavlFixture memory iavl = iavlFixture(json);
        bytes memory wrongValue = bytes.concat(iavl.value, hex"00");
        require(
            !bench.verifyIavlExistenceProof(
                iavl.root, wrongValue, iavl.leafPrefix, iavl.innerPrefixes, iavl.innerSuffixes
            ),
            "iavl proof accepted bad value"
        );
    }

    function testIcs23IavlExistenceProofDecodeOnlyGas() public {
        EvmLightClientBench bench = new EvmLightClientBench();
        string memory json = readFixture("update_50_equal");
        IavlFixture memory iavl = iavlFixture(json);
        (uint256 depth, bytes32 keyHash) = bench.decodeIcs23IavlExistenceProof(iavl.existenceProof);
        require(depth == iavl.innerPrefixes.length, "ics23 decoded depth mismatch");
        require(keyHash == keccak256(iavl.key), "ics23 decoded key mismatch");
    }

    function testIcs23IavlExistenceProofVerifyGas() public {
        EvmLightClientBench bench = new EvmLightClientBench();
        string memory json = readFixture("update_50_equal");
        IavlFixture memory iavl = iavlFixture(json);
        require(
            bench.verifyIcs23IavlExistenceProof(iavl.root, iavl.existenceProof, iavl.key, iavl.value),
            "ics23 iavl wire proof failed"
        );
    }

    function testIcs23IavlExistenceProofRejectsWrongKey() public {
        EvmLightClientBench bench = new EvmLightClientBench();
        string memory json = readFixture("update_50_equal");
        IavlFixture memory iavl = iavlFixture(json);
        vm.expectRevert(EvmLightClientBench.InvalidProof.selector);
        bench.verifyIcs23IavlExistenceProof(iavl.root, iavl.existenceProof, bytes.concat(iavl.key, hex"00"), iavl.value);
    }

    function testIcs23IavlExistenceProofRejectsWrongValue() public {
        EvmLightClientBench bench = new EvmLightClientBench();
        string memory json = readFixture("update_50_equal");
        IavlFixture memory iavl = iavlFixture(json);
        vm.expectRevert(EvmLightClientBench.InvalidProof.selector);
        bench.verifyIcs23IavlExistenceProof(iavl.root, iavl.existenceProof, iavl.key, bytes.concat(iavl.value, hex"00"));
    }

    function testIcs23IavlExistenceProofRejectsTamperedProof() public {
        EvmLightClientBench bench = new EvmLightClientBench();
        string memory json = readFixture("update_50_equal");
        IavlFixture memory iavl = iavlFixture(json);
        bytes memory proof = iavl.existenceProof;
        proof[proof.length - 1] ^= bytes1(0x01);
        require(!bench.verifyIcs23IavlExistenceProof(iavl.root, proof), "ics23 accepted tampered proof");
    }

    function testIcs23IavlExistenceProofRejectsUnsupportedLeafPrefix() public {
        EvmLightClientBench bench = new EvmLightClientBench();
        string memory json = readFixture("update_50_equal");
        IavlFixture memory iavl = iavlFixture(json);
        bytes memory proof = iavl.existenceProof;
        bool mutated;
        for (uint256 i; i + 3 < proof.length; ++i) {
            if (proof[i] == 0x2a && proof[i + 1] == 0x03 && proof[i + 2] == 0x00) {
                proof[i + 2] = 0x01;
                mutated = true;
                break;
            }
        }
        require(mutated, "leaf prefix fixture pattern not found");
        vm.expectRevert(EvmLightClientBench.InvalidProof.selector);
        bench.decodeIcs23IavlExistenceProof(proof);
    }

    function testIcs23IavlExistenceProofRejectsUnsupportedLengthOp() public {
        EvmLightClientBench bench = new EvmLightClientBench();
        string memory json = readFixture("update_50_equal");
        IavlFixture memory iavl = iavlFixture(json);
        bytes memory proof = iavl.existenceProof;
        bool mutated;
        // LeafOp.length = VAR_PROTO is encoded as field 4, varint 1: 0x20 0x01.
        for (uint256 i; i + 1 < proof.length; ++i) {
            if (proof[i] == 0x20 && proof[i + 1] == 0x01) {
                proof[i + 1] = 0x00;
                mutated = true;
                break;
            }
        }
        require(mutated, "leaf length fixture pattern not found");
        vm.expectRevert(EvmLightClientBench.InvalidProof.selector);
        bench.decodeIcs23IavlExistenceProof(proof);
    }

    function testIcs23IavlExistenceProofRejectsUnsupportedHashOp() public {
        EvmLightClientBench bench = new EvmLightClientBench();
        string memory json = readFixture("update_50_equal");
        IavlFixture memory iavl = iavlFixture(json);
        bytes memory proof = iavl.existenceProof;
        // First LeafOp field is hash=SHA256, encoded as 0x08 0x01.
        for (uint256 i; i + 1 < proof.length; ++i) {
            if (proof[i] == 0x08 && proof[i + 1] == 0x01) {
                proof[i + 1] = 0x02;
                break;
            }
        }
        vm.expectRevert(EvmLightClientBench.InvalidProof.selector);
        bench.decodeIcs23IavlExistenceProof(proof);
    }

    function testIcs23IavlExistenceProofRejectsOversizedProofBeforeHashing() public {
        EvmLightClientBench bench = new EvmLightClientBench();
        bytes memory oversized = new bytes(16_385);
        vm.expectRevert(EvmLightClientBench.InvalidProof.selector);
        bench.decodeIcs23IavlExistenceProof(oversized);
    }

    function testBenchBlsPairingCheckGas() public {
        EvmLightClientBench bench = new EvmLightClientBench();
        require(bench.benchBlsPairingCheck(), "bls pairing identity failed");
    }

    function testBenchBlsMapFp2ToG2Gas() public {
        EvmLightClientBench bench = new EvmLightClientBench();
        bytes memory g2 = bench.benchBlsMapFp2ToG2(keccak256("bls-map-seed"));
        require(g2.length == 256, "map_fp2_to_g2 unexpected length");
    }

    function testBenchBlsG1AddSignerAggregation50Gas() public {
        EvmLightClientBench bench = new EvmLightClientBench();
        bytes memory acc = bench.benchBlsG1AddSignerAggregation(50);
        require(acc.length == 128, "g1 add 50 unexpected length");
    }

    function testBenchBlsG1AddSignerAggregation10Gas() public {
        EvmLightClientBench bench = new EvmLightClientBench();
        bytes memory acc = bench.benchBlsG1AddSignerAggregation(10);
        require(acc.length == 128, "g1 add 10 unexpected length");
    }

    function testBenchBlsG1AddSignerAggregation100Gas() public {
        EvmLightClientBench bench = new EvmLightClientBench();
        bytes memory acc = bench.benchBlsG1AddSignerAggregation(100);
        require(acc.length == 128, "g1 add 100 unexpected length");
    }

    function testBenchBlsG1AddSignerAggregation175Gas() public {
        EvmLightClientBench bench = new EvmLightClientBench();
        bytes memory acc = bench.benchBlsG1AddSignerAggregation(175);
        require(acc.length == 128, "g1 add 175 unexpected length");
    }

    function testBenchBlsG2AddSignatureAggregation50Gas() public {
        EvmLightClientBench bench = new EvmLightClientBench();
        bytes memory acc = bench.benchBlsG2AddSignatureAggregation(50);
        require(acc.length == 256, "g2 add 50 unexpected length");
    }

    function testBenchBlsG2AddSignatureAggregation10Gas() public {
        EvmLightClientBench bench = new EvmLightClientBench();
        bytes memory acc = bench.benchBlsG2AddSignatureAggregation(10);
        require(acc.length == 256, "g2 add 10 unexpected length");
    }

    function testBenchBlsG2AddSignatureAggregation100Gas() public {
        EvmLightClientBench bench = new EvmLightClientBench();
        bytes memory acc = bench.benchBlsG2AddSignatureAggregation(100);
        require(acc.length == 256, "g2 add 100 unexpected length");
    }

    function testBenchBlsG2AddSignatureAggregation175Gas() public {
        EvmLightClientBench bench = new EvmLightClientBench();
        bytes memory acc = bench.benchBlsG2AddSignatureAggregation(175);
        require(acc.length == 256, "g2 add 175 unexpected length");
    }

    function testBenchBlsAggregateApprox50Gas() public {
        EvmLightClientBench bench = new EvmLightClientBench();
        require(bench.benchBlsAggregateApprox(50), "bls approx 50 failed");
    }

    function testBenchBlsAggregateApprox10Gas() public {
        EvmLightClientBench bench = new EvmLightClientBench();
        require(bench.benchBlsAggregateApprox(10), "bls approx 10 failed");
    }

    function testBenchBlsAggregateApprox100Gas() public {
        EvmLightClientBench bench = new EvmLightClientBench();
        require(bench.benchBlsAggregateApprox(100), "bls approx 100 failed");
    }

    function testBenchBlsAggregateApprox175Gas() public {
        EvmLightClientBench bench = new EvmLightClientBench();
        require(bench.benchBlsAggregateApprox(175), "bls approx 175 failed");
    }

    function testBenchBlsMultiMessagePairing10Gas() public {
        EvmLightClientBench bench = new EvmLightClientBench();
        require(!bench.benchBlsMultiMessagePairing(10), "bls multi-message 10 unexpectedly valid");
    }

    function testBenchBlsMultiMessagePairing50Gas() public {
        EvmLightClientBench bench = new EvmLightClientBench();
        require(!bench.benchBlsMultiMessagePairing(50), "bls multi-message 50 unexpectedly valid");
    }

    function testBenchBlsMultiMessagePairing100Gas() public {
        EvmLightClientBench bench = new EvmLightClientBench();
        require(!bench.benchBlsMultiMessagePairing(100), "bls multi-message 100 unexpectedly valid");
    }

    function testBenchBlsMultiMessagePairing175Gas() public {
        EvmLightClientBench bench = new EvmLightClientBench();
        require(!bench.benchBlsMultiMessagePairing(175), "bls multi-message 175 unexpectedly valid");
    }

    // testFullBitmapWithBogusLastSignatureReverts: full bitmap, all but the last
    // signature valid. Worst case for the verifier — every preceding ecrecover
    // succeeds before the final one fails. Charges the upper bound of compact
    // commit verification gas.
    function testFullBitmapWithBogusLastSignatureReverts() public {
        (EvmLightClientBench bench, string memory json) = initializedBench("update_50_equal");
        ValidatorFixture memory vals = validatorFixture(json, ".trustedValidators");
        EvmLightClientBench.CompactCommit memory commit = compactCommitFixture(json, ".adjacent.commit");
        bytes memory bogus = commit.signatures[commit.signatures.length - 1];
        bogus[0] ^= bytes1(0x01);
        commit.signatures[commit.signatures.length - 1] = bogus;
        vm.expectRevert(EvmLightClientBench.InvalidSignature.selector);
        bench.verifyCommitCompact(
            vals.validators, vals.powers, commit, uintFixture(json, ".adjacent.commit.requiredPower")
        );
    }

    // testDuplicateValidatorAtEndReverts: place the duplicate at the highest
    // index. The duplicate-detection loop is bitmap-based so cost is the same as
    // index 1, but this confirms the check fires regardless of position.
    function testDuplicateValidatorAtEndReverts() public {
        (EvmLightClientBench bench, string memory json) = initializedBench("update_50_equal");
        ValidatorFixture memory vals = validatorFixture(json, ".trustedValidators");
        vals.validators[vals.validators.length - 1] = vals.validators[0];
        vm.expectRevert(EvmLightClientBench.DuplicateValidator.selector);
        bench.verifyAdjacentUpdateCalldata(
            trustedHeight(json),
            vals.validators,
            vals.powers,
            vals.validators,
            vals.powers,
            headerFixture(json, ".adjacent.header"),
            compactCommitFixture(json, ".adjacent.commit"),
            clientFixture(json).currentTimestamp
        );
    }

    // testIavlMaxDepthProofGas: 16-level IAVL proof — twice the typical depth.
    // Charges the upper bound of single-key membership proof gas.
    function testIavlMaxDepthProofGas() public {
        EvmLightClientBench bench = new EvmLightClientBench();
        string memory json = readFixture("update_50_equal");
        IavlFixture memory iavl = iavlMaxDepthFixture(json);
        require(iavl.innerPrefixes.length == 16, "expected depth 16");
        require(
            bench.verifyIavlExistenceProof(
                iavl.root, iavl.value, iavl.leafPrefix, iavl.innerPrefixes, iavl.innerSuffixes
            ),
            "max-depth iavl proof failed"
        );
    }

    function testIavlMaxDepthProofRejectsWrongRoot() public {
        EvmLightClientBench bench = new EvmLightClientBench();
        string memory json = readFixture("update_50_equal");
        IavlFixture memory iavl = iavlMaxDepthFixture(json);
        bytes32 tampered = iavl.root ^ bytes32(uint256(1));
        require(
            !bench.verifyIavlExistenceProof(
                tampered, iavl.value, iavl.leafPrefix, iavl.innerPrefixes, iavl.innerSuffixes
            ),
            "iavl proof accepted bad root"
        );
    }

    // testEd25519PureSolidityVerifyGas executes a real ed25519 verification
    // over CometBFT canonical vote sign-bytes using chengwenxi/Ed25519
    // (Apache-2.0), a pure-Solidity verifier with an in-Solidity SHA-512.
    function testEd25519PureSolidityVerifyGas() public {
        IEd25519VerifierBench verifier =
            IEd25519VerifierBench(vm.deployCode("Ed25519VerifierBench.sol:Ed25519VerifierBench"));
        string memory json = readFixture("update_50_equal");
        CanonicalEd25519Fixture memory ed = canonicalEd25519Fixture(json);
        CanonicalVoteFixture memory vote = canonicalVoteFixture(json);

        require(ed.pubKeys.length > 0, "missing ed25519 pubkey");
        require(ed.signatures.length > 0, "missing ed25519 signature");
        require(vote.signBytes.length > 0, "missing canonical vote signBytes");
        require(verifier.verify(ed.pubKeys[0], ed.signatures[0], vote.signBytes[0]), "ed25519 verify failed");
    }

    function testEd25519PureSolidityRejectsWrongMessage() public {
        IEd25519VerifierBench verifier =
            IEd25519VerifierBench(vm.deployCode("Ed25519VerifierBench.sol:Ed25519VerifierBench"));
        string memory json = readFixture("update_50_equal");
        CanonicalEd25519Fixture memory ed = canonicalEd25519Fixture(json);
        CanonicalVoteFixture memory vote = canonicalVoteFixture(json);

        bytes memory tampered = vote.signBytes[0];
        tampered[0] = bytes1(uint8(tampered[0]) ^ 0x01);
        require(!verifier.verify(ed.pubKeys[0], ed.signatures[0], tampered), "ed25519 accepted wrong message");
    }

    // testEd25519AnalyticalGasBaseline records derived quorum costs from the
    // measured Ed25519VerifierBench.verify gas-report row. The measured range is
    // 902,866..909,757 gas per signature for chengwenxi/Ed25519 against a real
    // CometBFT canonical ed25519 vote signature.
    //
    // No EIP/RIP for an ed25519 precompile is currently scheduled for an
    // Ethereum L1 hardfork (Pectra includes EIP-2537 BLS12-381 but not
    // ed25519). Several L2s (Optimism, Base, zkSync) have shipped or proposed
    // precompiles in the ~3k-10k gas range; we report 5k as a representative
    // L2 figure that is in the same range as ecrecover.
    function testEd25519AnalyticalGasBaseline() public {
        emit AnalyticalGasEstimate(
            "ed25519-verify-pure-solidity-measured",
            902_866,
            909_757,
            "chengwenxi/Ed25519 measured by Ed25519VerifierBench.verify (per-signature)"
        );
        emit AnalyticalGasEstimate(
            "ed25519-verify-pure-solidity-50-validator-quorum",
            30_697_444,
            30_931_738,
            "34 * measured pure-Solidity ed25519 verify"
        );
        emit AnalyticalGasEstimate(
            "ed25519-verify-pure-solidity-100-validator-quorum",
            60_492_022,
            60_953_719,
            "67 * measured pure-Solidity ed25519 verify; near/exceeds ~60M Ethereum L1 block gas limit"
        );
        emit AnalyticalGasEstimate(
            "ed25519-verify-pure-solidity-180-validator-quorum",
            108_343_920,
            109_170_840,
            "120 * measured pure-Solidity ed25519 verify; exceeds Ethereum L1 block gas limit"
        );
        emit AnalyticalGasEstimate(
            "ed25519-verify-l2-precompile",
            3_000,
            10_000,
            "Optimism/Base/zkSync ed25519 precompile proposals (per-signature)"
        );
    }

    // testCanonicalBlsAggregateVerify executes a real on-chain BLS aggregate
    // verification using the EIP-2537 PAIRING_CHECK precompile. The fixture
    // generator (built with `-tags bls12381`) aggregates the signers' public
    // keys (G1) and signatures (G2) off-chain via blst and pre-computes the
    // hash-to-curve point H(m) (G2). On-chain we evaluate
    //   e(-G1, aggSig) * e(aggPubKey, H(m)) == 1
    // which is the BLS aggregate verify equation in min-pubkey-size mode.
    //
    // Skipped (not failed) when canonical.bls.available is false — the no-bls
    // build path of the fixture generator does not populate the aggregate.
    function testCanonicalBlsAggregateVerify() public {
        assertBlsAggregateVerify("update_50_equal");
    }

    function testCanonicalBlsAggregateVerify10Gas() public {
        assertBlsAggregateVerify("update_10_equal");
    }

    function testCanonicalBlsAggregateVerify100Gas() public {
        assertBlsAggregateVerify("update_100_equal");
    }

    function testCanonicalBlsAggregateVerify175Gas() public {
        assertBlsAggregateVerify("update_175_equal");
    }

    function assertBlsAggregateVerify(string memory fixtureName) private {
        EvmLightClientBench bench = new EvmLightClientBench();
        string memory json = readFixture(fixtureName);
        CanonicalBlsFixture memory bls = canonicalBlsFixture(json);
        if (!bls.available) {
            emit AnalyticalGasEstimate(
                "bls-aggregate-verify-skipped",
                0,
                0,
                "fixture generated without -tags bls12381; rebuild fixtures with bls12381 to exercise"
            );
            return;
        }
        require(bls.aggregatePubKey.length == 128, "agg pubkey not 128 bytes (EIP-2537 G1)");
        require(bls.hashedMessage.length == 256, "hashed message not 256 bytes (EIP-2537 G2)");
        require(bls.aggregateSig.length == 256, "agg sig not 256 bytes (EIP-2537 G2)");
        require(
            bench.verifyBlsAggregate(bls.aggregatePubKey, bls.hashedMessage, bls.aggregateSig),
            "bls aggregate verify failed"
        );
    }

    function testBlsAggregateStoredValidatorSet50Gas() public {
        EvmLightClientBench bench = new EvmLightClientBench();
        string memory json = readFixture("update_50_equal");
        CanonicalBlsFixture memory bls = canonicalBlsFixture(json);
        if (!bls.available) return;
        bytes32 setHash = bench.storeBlsValidatorSet(bls.hash, bls.pubKeys, bls.pubKeysEip2537, bls.powers);
        require(setHash == bls.hash, "stored bls set hash mismatch");
        require(
            bench.verifyBlsAggregateStoredValidatorSet(
                setHash, bls.signerBitmap, (totalPower(bls.powers) * 2) / 3, bls.hashedMessage, bls.aggregateSig
            ),
            "stored bls aggregate verify failed"
        );
    }

    function testBlsAggregateCalldataValidatorSet50Gas() public {
        EvmLightClientBench bench = new EvmLightClientBench();
        string memory json = readFixture("update_50_equal");
        CanonicalBlsFixture memory bls = canonicalBlsFixture(json);
        if (!bls.available) return;
        require(
            bench.verifyBlsAggregateCalldataValidatorSet(
                bls.pubKeysEip2537,
                bls.powers,
                bls.signerBitmap,
                (totalPower(bls.powers) * 2) / 3,
                bls.hashedMessage,
                bls.aggregateSig
            ),
            "calldata bls aggregate verify failed"
        );
    }

    function testStoreBlsValidatorSet50Gas() public {
        EvmLightClientBench bench = new EvmLightClientBench();
        string memory json = readFixture("update_50_equal");
        CanonicalBlsFixture memory bls = canonicalBlsFixture(json);
        if (!bls.available) return;
        bytes32 setHash = bench.storeBlsValidatorSet(bls.hash, bls.pubKeys, bls.pubKeysEip2537, bls.powers);
        require(setHash == bls.hash, "stored bls set hash mismatch");
    }

    function testBlsAggregateStoredValidatorSetRejectsWrongBitmap() public {
        EvmLightClientBench bench = new EvmLightClientBench();
        string memory json = readFixture("update_50_equal");
        CanonicalBlsFixture memory bls = canonicalBlsFixture(json);
        if (!bls.available) return;
        bytes32 setHash = bench.storeBlsValidatorSet(bls.hash, bls.pubKeys, bls.pubKeysEip2537, bls.powers);
        uint256 wrongBitmap = (bls.signerBitmap & ~uint256(1)) | (uint256(1) << 49);
        require(
            !bench.verifyBlsAggregateStoredValidatorSet(
                setHash, wrongBitmap, (totalPower(bls.powers) * 2) / 3, bls.hashedMessage, bls.aggregateSig
            ),
            "stored bls accepted wrong bitmap"
        );
    }

    function testStoreBlsValidatorSetRejectsWrongCanonicalHash() public {
        EvmLightClientBench bench = new EvmLightClientBench();
        string memory json = readFixture("update_50_equal");
        CanonicalBlsFixture memory bls = canonicalBlsFixture(json);
        if (!bls.available) return;
        vm.expectRevert(EvmLightClientBench.InvalidValidatorSet.selector);
        bench.storeBlsValidatorSet(bytes32(uint256(bls.hash) ^ 1), bls.pubKeys, bls.pubKeysEip2537, bls.powers);
    }

    function testBlsAggregateStoredValidatorSetRejectsTrailingBitmapBits() public {
        EvmLightClientBench bench = new EvmLightClientBench();
        string memory json = readFixture("update_50_equal");
        CanonicalBlsFixture memory bls = canonicalBlsFixture(json);
        if (!bls.available) return;
        bytes32 setHash = bench.storeBlsValidatorSet(bls.hash, bls.pubKeys, bls.pubKeysEip2537, bls.powers);
        vm.expectRevert(EvmLightClientBench.InvalidCommit.selector);
        bench.verifyBlsAggregateStoredValidatorSet(
            setHash,
            bls.signerBitmap | (uint256(1) << bls.powers.length),
            (totalPower(bls.powers) * 2) / 3,
            bls.hashedMessage,
            bls.aggregateSig
        );
    }

    function testBlsAggregateCalldataValidatorSetRejectsTrailingBitmapBits() public {
        EvmLightClientBench bench = new EvmLightClientBench();
        string memory json = readFixture("update_50_equal");
        CanonicalBlsFixture memory bls = canonicalBlsFixture(json);
        if (!bls.available) return;
        vm.expectRevert(EvmLightClientBench.InvalidCommit.selector);
        bench.verifyBlsAggregateCalldataValidatorSet(
            bls.pubKeysEip2537,
            bls.powers,
            bls.signerBitmap | (uint256(1) << bls.powers.length),
            (totalPower(bls.powers) * 2) / 3,
            bls.hashedMessage,
            bls.aggregateSig
        );
    }

    function testBlsAggregateRejectsWrongMessageFixture() public {
        EvmLightClientBench bench = new EvmLightClientBench();
        string memory json = readFixture("update_50_equal");
        CanonicalBlsFixture memory bls = canonicalBlsFixture(json);
        if (!bls.available) return;
        require(
            !bench.verifyBlsAggregate(bls.aggregatePubKey, bls.wrongMessageHashed, bls.aggregateSig),
            "bls verify accepted wrong-message fixture"
        );
    }

    function testBlsAggregateRejectsMissingSignerFixture() public {
        EvmLightClientBench bench = new EvmLightClientBench();
        string memory json = readFixture("update_50_equal");
        CanonicalBlsFixture memory bls = canonicalBlsFixture(json);
        if (!bls.available) return;
        require(
            !bench.verifyBlsAggregate(bls.aggregatePubKey, bls.hashedMessage, bls.missingSignerAggregateSig),
            "bls verify accepted missing-signer aggregate signature"
        );
    }

    // testCanonicalBlsAggregateVerifyValidButWrongRejected uses VALID but
    // MISMATCHED inputs: aggregate signature from one fixture combined with
    // the hashed message from another. Both points are well-formed G1/G2
    // elements, so the precompile validates them successfully and runs the
    // full pairing computation — but the equation
    //   e(-G1, aggSig_A) * e(aggPubKey_A, H(m_B)) == 1
    // does not hold (the signature attests to message_A, not message_B), so
    // the precompile returns 0 and verifyBlsAggregate returns false.
    //
    // This is the realistic "signed-X-claim-Y" forgery test. Gas cost is the
    // ordinary 2-pair PAIRING_CHECK cost (~108k precompile + a few k solidity
    // overhead, total ~110k), the same as the happy path — the pairing check
    // charges by input length, not by whether the equation holds.
    function testCanonicalBlsAggregateVerifyValidButWrongRejected() public {
        EvmLightClientBench bench = new EvmLightClientBench();
        string memory jsonA = readFixture("update_50_equal");
        string memory jsonB = readFixture("update_100_equal");
        CanonicalBlsFixture memory blsA = canonicalBlsFixture(jsonA);
        CanonicalBlsFixture memory blsB = canonicalBlsFixture(jsonB);
        if (!blsA.available || !blsB.available) {
            return;
        }
        require(
            !bench.verifyBlsAggregate(blsA.aggregatePubKey, blsB.hashedMessage, blsA.aggregateSig),
            "bls verify accepted cross-message aggregate"
        );
    }

    // testCanonicalBlsAggregateRejectsMalformedInputBounded flips a bit in
    // the y-coordinate of the hashed-message G2 element so the resulting
    // bytes no longer encode a valid curve point. Two properties matter:
    //
    //   1. Verification returns false (the precompile rejects an invalid
    //      G2 point and verifyBlsAggregate propagates that as `false`).
    //   2. Total gas consumed is bounded. Without a gas cap on the
    //      staticcall, foundry's EIP-2537 PAIRING_CHECK has been observed to
    //      consume ~1B gas per call when given a syntactically well-shaped
    //      but invalid 768-byte input — i.e. an attacker submitting one
    //      malformed aggregate could drain the caller's full transaction
    //      gas. verifyBlsAggregate caps forwarded gas at
    //      PAIRING_CHECK_2_PAIRS_GAS_BOUND (200_000), so adversarial cost is
    //      ~that cap plus the function's own constant overhead.
    //
    // Skipped when canonical.bls.available is false (no BLS fixture).
    function testCanonicalBlsAggregateRejectsMalformedInputBounded() public {
        EvmLightClientBench bench = new EvmLightClientBench();
        string memory json = readFixture("update_50_equal");
        CanonicalBlsFixture memory bls = canonicalBlsFixture(json);
        if (!bls.available) {
            return;
        }
        bytes memory malformed = bls.hashedMessage;
        malformed[200] ^= bytes1(0x01);

        uint256 gasBefore = gasleft();
        bool result = bench.verifyBlsAggregate(bls.aggregatePubKey, malformed, bls.aggregateSig);
        uint256 gasUsed = gasBefore - gasleft();

        require(!result, "bls verify accepted malformed H(m)");
        // 250k headroom over the 200k pairing-check bound covers solidity
        // dispatch, calldata copy, return-data handling, and forge wrapper
        // overhead. Without the bound this call has been measured at ~1B gas.
        require(gasUsed < 250_000, "malformed bls verify exceeded gas bound");
    }

    // testCanonicalBlsValidatorSetHashParity recomputes the SimpleValidator
    // RFC6962 Merkle hash on-chain using the BLS12-381 wire format and
    // asserts it matches the fixture hash from cometbft's
    // ValidatorSet.Hash(). Skipped when canonical.bls.available is false.
    function testCanonicalBlsValidatorSetHashParity() public {
        EvmLightClientBench bench = new EvmLightClientBench();
        string memory json = readFixture("update_50_equal");
        CanonicalBlsFixture memory bls = canonicalBlsFixture(json);
        if (!bls.available) {
            return;
        }
        bytes32 reconstructed = bench.hashSimpleValidatorSetBls12381(bls.pubKeys, bls.powers);
        require(reconstructed == bls.hash, "bls validator-set hash mismatch");
    }

    // testCanonicalHeaderHashRecorded asserts the canonical (cometbft-style)
    // header hash is non-zero and present in the fixture. The canonical hash
    // is produced by Header.Hash() in cometbft (RFC6962 Merkle root over 14
    // proto-encoded fields) and is what a fully-canonical on-chain light
    // client must reproduce. The EVM-native hashHeader function in
    // EvmLightClientBench is intentionally a different (compact) shape, so
    // the two MUST differ — equality would indicate the canonical hash field
    // is silently shadowing the compact one in the fixture.
    function testCanonicalHeaderHashRecorded() public {
        EvmLightClientBench bench = new EvmLightClientBench();
        string memory json = readFixture("update_50_equal");
        bytes32 canonicalHash = bytes32Fixture(json, ".canonical.headerHash");
        require(canonicalHash != bytes32(0), "canonical headerHash is zero");
        EvmLightClientBench.Header memory hdr = headerFixture(json, ".adjacent.header");
        bytes32 compactHash = bench.hashHeader(hdr);
        require(canonicalHash != compactHash, "canonical and compact header hashes coincide unexpectedly");
    }

    // testCanonicalVoteSignBytesDistinctFromCompact asserts the canonical
    // CanonicalVote sign-bytes (proto-encoded with chainID, height, round,
    // BlockID, timestamp, etc.) exist for every signer and are byte-distinct
    // from the compact path the EVM commit verifier signs over. The canonical
    // bytes are what cometbft validators actually sign on-chain; the compact
    // bytes are an EVM-friendly reshape used by the secp256k1eth path.
    function testCanonicalVoteSignBytesDistinctFromCompact() public view {
        string memory json = readFixture("update_50_equal");
        CanonicalVoteFixture memory vote = canonicalVoteFixture(json);
        EvmLightClientBench.PrebuiltVoteCommit memory commit = prebuiltCommitFixture(json, ".adjacent.commit");
        require(vote.signBytes.length > 0, "canonical vote signBytes empty");
        require(vote.signBytes.length == commit.voteSignBytes.length, "signer count mismatch");
        for (uint256 i; i < vote.signBytes.length; ++i) {
            require(vote.signBytes[i].length > 0, "canonical signBytes[i] empty");
            require(commit.voteSignBytes[i].length > 0, "compact voteSignBytes[i] empty");
            require(
                keccak256(vote.signBytes[i]) != keccak256(commit.voteSignBytes[i]),
                "canonical and compact signBytes coincide unexpectedly"
            );
        }
    }

    function testCanonicalVoteSignBytesReconstructionParity() public {
        EvmLightClientBench bench = new EvmLightClientBench();
        string memory json = readFixture("update_50_equal");
        CanonicalVoteFixture memory vote = canonicalVoteFixture(json);
        for (uint256 i; i < vote.signBytes.length; ++i) {
            bytes memory reconstructed = bench.reconstructCanonicalVoteSignBytes(
                vote.chainId,
                vote.height,
                vote.round,
                vote.blockHash,
                vote.partSetTotal,
                vote.partSetHash,
                uint64Timestamp(vote.timestamps[i])
            );
            require(keccak256(reconstructed) == keccak256(vote.signBytes[i]), "canonical vote reconstruction mismatch");
        }
    }

    function testCanonicalVoteHashGas() public {
        EvmLightClientBench bench = new EvmLightClientBench();
        string memory json = readFixture("update_50_equal");
        CanonicalVoteFixture memory vote = canonicalVoteFixture(json);
        bytes32 digest = bench.hashCanonicalVoteSignBytes(
            vote.chainId,
            vote.height,
            vote.round,
            vote.blockHash,
            vote.partSetTotal,
            vote.partSetHash,
            uint64Timestamp(vote.timestamps[0])
        );
        require(digest == keccak256(vote.signBytes[0]), "canonical vote hash mismatch");
    }

    function testCanonicalVoteSecp256k1VerifyGas() public {
        EvmLightClientBench bench = new EvmLightClientBench();
        string memory json = readFixture("update_50_equal");
        ValidatorFixture memory vals = validatorFixture(json, ".trustedValidators");
        CanonicalVoteFixture memory vote = canonicalVoteFixture(json);
        require(
            bench.verifyCanonicalVoteSecp256k1(
                vals.validators[0],
                vote.ethSignatures[0],
                vote.chainId,
                vote.height,
                vote.round,
                vote.blockHash,
                vote.partSetTotal,
                vote.partSetHash,
                uint64Timestamp(vote.timestamps[0])
            ),
            "canonical vote secp256k1 verify failed"
        );
    }

    function testIavlMaxDepthProofRejectsTamperedSuffix() public {
        EvmLightClientBench bench = new EvmLightClientBench();
        string memory json = readFixture("update_50_equal");
        IavlFixture memory iavl = iavlMaxDepthFixture(json);
        for (uint256 i; i < iavl.innerSuffixes.length; ++i) {
            if (iavl.innerSuffixes[i].length > 0) {
                iavl.innerSuffixes[i][0] ^= bytes1(0x80);
                break;
            }
        }
        require(
            !bench.verifyIavlExistenceProof(
                iavl.root, iavl.value, iavl.leafPrefix, iavl.innerPrefixes, iavl.innerSuffixes
            ),
            "iavl proof accepted tampered suffix"
        );
    }

    function testMissingTrustedConsensusStateReverts() public {
        (EvmLightClientBench bench, string memory json) = initializedBench("update_50_equal");
        ValidatorFixture memory vals = validatorFixture(json, ".trustedValidators");
        ClientFixture memory client = clientFixture(json);
        vm.expectRevert(EvmLightClientBench.ConsensusStateNotFound.selector);
        bench.verifyAdjacentUpdateCalldata(
            999,
            vals.validators,
            vals.powers,
            vals.validators,
            vals.powers,
            headerFixture(json, ".adjacent.header"),
            compactCommitFixture(json, ".adjacent.commit"),
            client.currentTimestamp
        );
    }

    function testValidatorHashMismatchReverts() public {
        (EvmLightClientBench bench, string memory json) = initializedBench("update_50_equal");
        ValidatorFixture memory vals = validatorFixture(json, ".trustedValidators");
        vals.powers[0] = 99;
        ClientFixture memory client = clientFixture(json);
        vm.expectRevert(EvmLightClientBench.InvalidValidatorSet.selector);
        bench.verifyAdjacentUpdateCalldata(
            trustedHeight(json),
            vals.validators,
            vals.powers,
            vals.validators,
            vals.powers,
            headerFixture(json, ".adjacent.header"),
            compactCommitFixture(json, ".adjacent.commit"),
            client.currentTimestamp
        );
    }

    function testDuplicateValidatorReverts() public {
        (EvmLightClientBench bench, string memory json) = initializedBench("update_50_equal");
        ValidatorFixture memory vals = validatorFixture(json, ".trustedValidators");
        vals.validators[1] = vals.validators[0];
        vm.expectRevert(EvmLightClientBench.DuplicateValidator.selector);
        bench.verifyAdjacentUpdateCalldata(
            trustedHeight(json),
            vals.validators,
            vals.powers,
            vals.validators,
            vals.powers,
            headerFixture(json, ".adjacent.header"),
            compactCommitFixture(json, ".adjacent.commit"),
            clientFixture(json).currentTimestamp
        );
    }

    function testCommitTrailingGarbageReverts() public {
        (EvmLightClientBench bench, string memory json) = initializedBench("update_50_equal");
        ValidatorFixture memory vals = validatorFixture(json, ".trustedValidators");
        EvmLightClientBench.CompactCommit memory commit = compactCommitFixture(json, ".adjacent.commit");
        bytes[] memory signatures = new bytes[](commit.signatures.length + 1);
        uint64[] memory timestamps = new uint64[](commit.timestamps.length + 1);
        for (uint256 i; i < commit.signatures.length; ++i) {
            signatures[i] = commit.signatures[i];
            timestamps[i] = commit.timestamps[i];
        }
        signatures[commit.signatures.length] = commit.signatures[0];
        timestamps[commit.timestamps.length] = commit.timestamps[0];
        commit.signatures = signatures;
        commit.timestamps = timestamps;

        vm.expectRevert(EvmLightClientBench.LengthMismatch.selector);
        bench.verifyCommitCompact(
            vals.validators, vals.powers, commit, uintFixture(json, ".adjacent.commit.requiredPower")
        );
    }

    function testCommitHeightMismatchReverts() public {
        assertCommitMismatch(true);
    }

    function testCommitBlockHashMismatchReverts() public {
        assertCommitMismatch(false);
    }

    function testInsufficientCommitPowerReturnsFalse() public {
        (EvmLightClientBench bench, string memory json) = initializedBench("update_50_equal");
        ValidatorFixture memory vals = validatorFixture(json, ".trustedValidators");
        EvmLightClientBench.CompactCommit memory commit = compactCommitFixture(json, ".adjacent.commit");
        require(
            !bench.verifyCommitCompact(vals.validators, vals.powers, commit, vals.validators.length),
            "unexpected threshold"
        );
    }

    function testNonAdjacentFailsInsufficientTrustPower() public {
        assertNonAdjacentInvalidCommit(".nonAdjacent.weakCommit");
    }

    function testNonAdjacentFailsInsufficientUntrustedCommitPower() public {
        assertNonAdjacentInvalidCommit(".nonAdjacent.trustOnlyCommit");
    }

    function testDuplicateConsensusStateUpdateReverts() public {
        (EvmLightClientBench bench, string memory json) = initializedBench("update_50_equal");
        ValidatorFixture memory vals = validatorFixture(json, ".trustedValidators");
        EvmLightClientBench.Header memory header = headerFixture(json, ".adjacent.header");
        EvmLightClientBench.CompactCommit memory commit = compactCommitFixture(json, ".adjacent.commit");
        ClientFixture memory client = clientFixture(json);

        bench.updateAdjacentCalldata(
            trustedHeight(json),
            vals.validators,
            vals.powers,
            vals.validators,
            vals.powers,
            header,
            commit,
            client.currentTimestamp,
            client.processedHeight
        );
        vm.expectRevert(EvmLightClientBench.ConsensusStateAlreadyExists.selector);
        bench.updateAdjacentCalldata(
            trustedHeight(json),
            vals.validators,
            vals.powers,
            vals.validators,
            vals.powers,
            header,
            commit,
            client.currentTimestamp,
            client.processedHeight
        );
    }

    function testTooManyValidatorsReverts() public {
        EvmLightClientBench bench = new EvmLightClientBench();
        address[] memory validators = new address[](257);
        uint64[] memory powers = new uint64[](257);
        for (uint256 i; i < validators.length; ++i) {
            validators[i] = address(uint160(i + 1));
            powers[i] = 1;
        }
        vm.expectRevert(EvmLightClientBench.TooManyValidators.selector);
        bench.hashValidatorSetNative(validators, powers);
    }

    function testExpiredTrustedStateReverts() public {
        (EvmLightClientBench bench, string memory json) = initializedBench("update_50_equal");
        ValidatorFixture memory vals = validatorFixture(json, ".trustedValidators");
        ClientFixture memory client = clientFixture(json);
        vm.expectRevert(EvmLightClientBench.ExpiredTrustedState.selector);
        bench.verifyAdjacentUpdateCalldata(
            trustedHeight(json),
            vals.validators,
            vals.powers,
            vals.validators,
            vals.powers,
            headerFixture(json, ".adjacent.header"),
            compactCommitFixture(json, ".adjacent.commit"),
            client.currentTimestamp + client.trustingPeriod + 1
        );
    }

    function testIdenticalSameHeightMisbehaviourReverts() public {
        (EvmLightClientBench bench, string memory json) = initializedBench("update_50_equal");
        ValidatorFixture memory vals = validatorFixture(json, ".trustedValidators");
        ClientFixture memory client = clientFixture(json);
        EvmLightClientBench.Header memory header = headerFixture(json, ".adjacent.header");
        EvmLightClientBench.CompactCommit memory commit = compactCommitFixture(json, ".adjacent.commit");

        vm.expectRevert(EvmLightClientBench.InvalidMisbehaviour.selector);
        bench.submitMisbehaviour(
            misbehaviourInput(json, vals, vals, header, commit),
            misbehaviourInput(json, vals, vals, header, commit),
            client.currentTimestamp
        );
    }

    function testMisbehaviourWithUntrustedValidatorSetReverts() public {
        (EvmLightClientBench bench, string memory json) = initializedBench("update_50_equal");
        string memory changedJson = readFixture("misbehaviour_50_equal");
        ValidatorFixture memory vals = validatorFixture(json, ".trustedValidators");
        ValidatorFixture memory changed = validatorFixture(changedJson, ".untrustedValidators");
        ClientFixture memory client = clientFixture(json);

        vm.expectRevert(EvmLightClientBench.InvalidValidatorSet.selector);
        bench.submitMisbehaviour(
            misbehaviourInput(json, vals, vals, ".adjacent"),
            misbehaviourInput(json, vals, changed, ".conflict"),
            client.currentTimestamp
        );
    }

    function testCalldataGasEstimates() public {
        emitCommitCalldata("commit-compact-10", "update_10_equal");
        emitCommitCalldata("commit-compact-50", "update_50_equal");
        emitCommitCalldata("commit-compact-100", "update_100_equal");
        emitCommitCalldata("commit-compact-175", "update_175_equal");
        emitPrebuiltCommitCalldata();
        emitAdjacentUpdateCalldata("adjacent-update-calldata-10", "update_10_equal");
        emitAdjacentUpdateCalldata("adjacent-update-calldata-50", "update_50_equal");
        emitAdjacentUpdateCalldata("adjacent-update-mixed-power-50", "update_50_mixed");
        emitNonAdjacentUpdateCalldata();
        emitMisbehaviour10Calldata();
        emitChangedMisbehaviourCalldata("changed-valset-misbehaviour-10", "misbehaviour_10_equal");
        emitChangedMisbehaviourCalldata("changed-valset-misbehaviour-50", "misbehaviour_50_equal");
        emitStoredValidatorSetUpdateCalldata();
        emitMembershipCalldata();
        emitIavlCalldata();
        emitIavlMaxDepthCalldata();
        emitIcs23WireCalldata();
        emitCanonicalVoteCalldata();
        emitCanonicalEd25519Calldata();
        emitCanonicalSecp256k1Calldata();
        emitBlsAggregateCalldata();
        emitBlsStoredValidatorSetCalldata();
        emitBlsCalldataValidatorSetCalldata();
        emitBlsMultiMessageCalldata();
    }

    function emitIavlCalldata() private {
        string memory json = readFixture("update_50_equal");
        IavlFixture memory iavl = iavlFixture(json);
        emitCallGas(
            "iavl-existence-proof-depth8",
            abi.encodeCall(
                EvmLightClientBench.verifyIavlExistenceProof,
                (iavl.root, iavl.value, iavl.leafPrefix, iavl.innerPrefixes, iavl.innerSuffixes)
            )
        );
    }

    function emitIavlMaxDepthCalldata() private {
        string memory json = readFixture("update_50_equal");
        IavlFixture memory iavl = iavlMaxDepthFixture(json);
        emitCallGas(
            "iavl-existence-proof-depth16",
            abi.encodeCall(
                EvmLightClientBench.verifyIavlExistenceProof,
                (iavl.root, iavl.value, iavl.leafPrefix, iavl.innerPrefixes, iavl.innerSuffixes)
            )
        );
    }

    function emitIcs23WireCalldata() private {
        string memory json = readFixture("update_50_equal");
        IavlFixture memory iavl = iavlFixture(json);
        emitCallGas(
            "ics23-iavl-existence-decode-depth8",
            abi.encodeCall(EvmLightClientBench.decodeIcs23IavlExistenceProof, (iavl.existenceProof))
        );
        emitCallGas(
            "ics23-iavl-existence-verify-depth8",
            abi.encodeWithSelector(
                bytes4(keccak256("verifyIcs23IavlExistenceProof(bytes32,bytes,bytes,bytes)")),
                iavl.root,
                iavl.existenceProof,
                iavl.key,
                iavl.value
            )
        );

        IavlFixture memory maxDepth = iavlMaxDepthFixture(json);
        emitCallGas(
            "ics23-iavl-existence-verify-depth16",
            abi.encodeWithSelector(
                bytes4(keccak256("verifyIcs23IavlExistenceProof(bytes32,bytes,bytes,bytes)")),
                maxDepth.root,
                maxDepth.existenceProof,
                maxDepth.key,
                maxDepth.value
            )
        );
    }

    function emitCanonicalVoteCalldata() private {
        string memory json = readFixture("update_50_equal");
        ValidatorFixture memory vals = validatorFixture(json, ".trustedValidators");
        CanonicalVoteFixture memory vote = canonicalVoteFixture(json);
        emitCallGas(
            "canonical-vote-reconstruct",
            abi.encodeCall(
                EvmLightClientBench.reconstructCanonicalVoteSignBytes,
                (
                    vote.chainId,
                    vote.height,
                    vote.round,
                    vote.blockHash,
                    vote.partSetTotal,
                    vote.partSetHash,
                    uint64Timestamp(vote.timestamps[0])
                )
            )
        );
        emitCallGas(
            "canonical-vote-reconstruct-hash",
            abi.encodeCall(
                EvmLightClientBench.hashCanonicalVoteSignBytes,
                (
                    vote.chainId,
                    vote.height,
                    vote.round,
                    vote.blockHash,
                    vote.partSetTotal,
                    vote.partSetHash,
                    uint64Timestamp(vote.timestamps[0])
                )
            )
        );
        emitCallGas(
            "canonical-vote-reconstruct-secp256k1",
            abi.encodeCall(
                EvmLightClientBench.verifyCanonicalVoteSecp256k1,
                (
                    vals.validators[0],
                    vote.ethSignatures[0],
                    vote.chainId,
                    vote.height,
                    vote.round,
                    vote.blockHash,
                    vote.partSetTotal,
                    vote.partSetHash,
                    uint64Timestamp(vote.timestamps[0])
                )
            )
        );
    }

    function emitCanonicalEd25519Calldata() private {
        string memory json = readFixture("update_50_equal");
        CanonicalEd25519Fixture memory ed = canonicalEd25519Fixture(json);
        emitCallGas(
            "canonical-validator-set-ed25519-50",
            abi.encodeCall(EvmLightClientBench.hashSimpleValidatorSetEd25519, (ed.pubKeys, ed.powers))
        );
    }

    function emitCanonicalSecp256k1Calldata() private {
        string memory json = readFixture("update_50_equal");
        CanonicalSecp256k1Fixture memory sk = canonicalSecp256k1Fixture(json);
        emitCallGas(
            "canonical-validator-set-secp256k1-50",
            abi.encodeCall(EvmLightClientBench.hashSimpleValidatorSetSecp256k1, (sk.pubKeys, sk.powers))
        );
    }

    function emitBlsAggregateCalldata() private {
        string memory json = readFixture("update_50_equal");
        CanonicalBlsFixture memory bls = canonicalBlsFixture(json);
        if (!bls.available) return;
        emitCallGas(
            "bls-a-aggregate-verify-supplied-aggregate-pubkey",
            abi.encodeCall(
                EvmLightClientBench.verifyBlsAggregate, (bls.aggregatePubKey, bls.hashedMessage, bls.aggregateSig)
            )
        );
    }

    function emitBlsStoredValidatorSetCalldata() private {
        EvmLightClientBench bench = new EvmLightClientBench();
        string memory json = readFixture("update_50_equal");
        CanonicalBlsFixture memory bls = canonicalBlsFixture(json);
        if (!bls.available) return;
        bytes32 setHash = bench.storeBlsValidatorSet(bls.hash, bls.pubKeys, bls.pubKeysEip2537, bls.powers);
        emitCallGas(
            "bls-b-store-validator-set-50",
            abi.encodeCall(
                EvmLightClientBench.storeBlsValidatorSet, (bls.hash, bls.pubKeys, bls.pubKeysEip2537, bls.powers)
            )
        );
        emitCallGas(
            "bls-b-stored-validator-set-bitmap-50",
            abi.encodeCall(
                EvmLightClientBench.verifyBlsAggregateStoredValidatorSet,
                (setHash, bls.signerBitmap, (totalPower(bls.powers) * 2) / 3, bls.hashedMessage, bls.aggregateSig)
            )
        );
    }

    function emitBlsCalldataValidatorSetCalldata() private {
        string memory json = readFixture("update_50_equal");
        CanonicalBlsFixture memory bls = canonicalBlsFixture(json);
        if (!bls.available) return;
        emitCallGas(
            "bls-c-calldata-validator-set-50",
            abi.encodeCall(
                EvmLightClientBench.verifyBlsAggregateCalldataValidatorSet,
                (
                    bls.pubKeysEip2537,
                    bls.powers,
                    bls.signerBitmap,
                    (totalPower(bls.powers) * 2) / 3,
                    bls.hashedMessage,
                    bls.aggregateSig
                )
            )
        );
    }

    function emitBlsMultiMessageCalldata() private {
        emitCallGas(
            "bls-d-synthetic-helper-pairing-10", abi.encodeCall(EvmLightClientBench.benchBlsMultiMessagePairing, (10))
        );
        emitCallGas(
            "bls-d-synthetic-helper-pairing-50", abi.encodeCall(EvmLightClientBench.benchBlsMultiMessagePairing, (50))
        );
        emitCallGas(
            "bls-d-synthetic-helper-pairing-100", abi.encodeCall(EvmLightClientBench.benchBlsMultiMessagePairing, (100))
        );
        emitCallGas(
            "bls-d-synthetic-helper-pairing-175", abi.encodeCall(EvmLightClientBench.benchBlsMultiMessagePairing, (175))
        );
    }

    function assertCompactCommit(string memory fixtureName) private {
        (EvmLightClientBench bench, string memory json) = initializedBench(fixtureName);
        ValidatorFixture memory vals = validatorFixture(json, ".trustedValidators");
        EvmLightClientBench.CompactCommit memory commit = compactCommitFixture(json, ".adjacent.commit");
        require(
            bench.verifyCommitCompact(
                vals.validators, vals.powers, commit, uintFixture(json, ".adjacent.commit.requiredPower")
            )
        );
    }

    function assertTimeViolation(string memory fixtureName, bool reversed) private {
        (EvmLightClientBench bench, string memory json) = initializedBench(fixtureName);
        ValidatorFixture memory vals = validatorFixture(json, ".trustedValidators");
        ClientFixture memory client = clientFixture(json);
        EvmLightClientBench.Header memory higher = headerFixture(json, ".timeViolationHigher.header");
        EvmLightClientBench.CompactCommit memory higherCommit =
            compactCommitFixture(json, ".timeViolationHigher.commit");
        EvmLightClientBench.Header memory lower = headerFixture(json, ".timeViolationLower.header");
        EvmLightClientBench.CompactCommit memory lowerCommit = compactCommitFixture(json, ".timeViolationLower.commit");

        if (reversed) {
            bench.submitMisbehaviour(
                misbehaviourInput(json, vals, vals, lower, lowerCommit),
                misbehaviourInput(json, vals, vals, higher, higherCommit),
                client.currentTimestamp
            );
        } else {
            bench.submitMisbehaviour(
                misbehaviourInput(json, vals, vals, higher, higherCommit),
                misbehaviourInput(json, vals, vals, lower, lowerCommit),
                client.currentTimestamp
            );
        }
    }

    function assertCommitMismatch(bool heightMismatch) private {
        (EvmLightClientBench bench, string memory json) = initializedBench("update_50_equal");
        ValidatorFixture memory vals = validatorFixture(json, ".trustedValidators");
        EvmLightClientBench.CompactCommit memory commit = compactCommitFixture(json, ".adjacent.commit");
        if (heightMismatch) {
            commit.height += 1;
        } else {
            commit.blockHash = bytes32(uint256(1));
        }
        vm.expectRevert(EvmLightClientBench.InvalidCommit.selector);
        bench.verifyAdjacentUpdateCalldata(
            trustedHeight(json),
            vals.validators,
            vals.powers,
            vals.validators,
            vals.powers,
            headerFixture(json, ".adjacent.header"),
            commit,
            clientFixture(json).currentTimestamp
        );
    }

    function assertNonAdjacentInvalidCommit(string memory commitPath) private {
        (EvmLightClientBench bench, string memory json) = initializedBench("misbehaviour_50_equal");
        ValidatorFixture memory trusted = validatorFixture(json, ".trustedValidators");
        ValidatorFixture memory untrusted = validatorFixture(json, ".untrustedValidators");
        vm.expectRevert(EvmLightClientBench.InvalidCommit.selector);
        bench.verifyNonAdjacentUpdateCalldata(
            trustedHeight(json),
            trusted.validators,
            trusted.powers,
            untrusted.validators,
            untrusted.powers,
            headerFixture(json, ".nonAdjacent.header"),
            compactCommitFixture(json, commitPath),
            clientFixture(json).currentTimestamp
        );
    }

    function emitPrebuiltCommitCalldata() private {
        string memory json = readFixture("update_50_equal");
        ValidatorFixture memory vals = validatorFixture(json, ".trustedValidators");
        emitCallGas(
            "commit-prebuilt-vote-bytes-50",
            abi.encodeCall(
                EvmLightClientBench.verifyCommitPrebuiltVoteBytes,
                (
                    vals.validators,
                    vals.powers,
                    prebuiltCommitFixture(json, ".adjacent.commit"),
                    uintFixture(json, ".adjacent.commit.requiredPower")
                )
            )
        );
    }

    function emitAdjacentUpdateCalldata(string memory name, string memory fixtureName) private {
        string memory json = readFixture(fixtureName);
        ValidatorFixture memory vals = validatorFixture(json, ".trustedValidators");
        emitCallGas(
            name,
            abi.encodeCall(
                EvmLightClientBench.verifyAdjacentUpdateCalldata,
                (
                    trustedHeight(json),
                    vals.validators,
                    vals.powers,
                    vals.validators,
                    vals.powers,
                    headerFixture(json, ".adjacent.header"),
                    compactCommitFixture(json, ".adjacent.commit"),
                    clientFixture(json).currentTimestamp
                )
            )
        );
    }

    function emitNonAdjacentUpdateCalldata() private {
        string memory json = readFixture("misbehaviour_50_equal");
        ValidatorFixture memory trusted = validatorFixture(json, ".trustedValidators");
        ValidatorFixture memory changed = validatorFixture(json, ".untrustedValidators");
        emitCallGas(
            "non-adjacent-update-changed-valset-50",
            abi.encodeCall(
                EvmLightClientBench.verifyNonAdjacentUpdateCalldata,
                (
                    trustedHeight(json),
                    trusted.validators,
                    trusted.powers,
                    changed.validators,
                    changed.powers,
                    headerFixture(json, ".nonAdjacent.header"),
                    compactCommitFixture(json, ".nonAdjacent.commit"),
                    clientFixture(json).currentTimestamp
                )
            )
        );
    }

    function emitMisbehaviour10Calldata() private {
        string memory json = readFixture("misbehaviour_10_equal");
        ValidatorFixture memory vals = validatorFixture(json, ".trustedValidators");
        emitCallGas(
            "misbehaviour-10",
            abi.encodeCall(
                EvmLightClientBench.submitMisbehaviour,
                (
                    misbehaviourInput(json, vals, vals, ".adjacent"),
                    misbehaviourInput(json, vals, vals, ".conflict"),
                    clientFixture(json).currentTimestamp
                )
            )
        );
    }

    function emitChangedMisbehaviourCalldata(string memory name, string memory fixtureName) private {
        string memory json = readFixture(fixtureName);
        ValidatorFixture memory trusted = validatorFixture(json, ".trustedValidators");
        ValidatorFixture memory changed = validatorFixture(json, ".untrustedValidators");
        emitCallGas(
            name,
            abi.encodeCall(
                EvmLightClientBench.submitMisbehaviour,
                (
                    misbehaviourInput(json, trusted, changed, ".nonAdjacent"),
                    misbehaviourInput(json, trusted, changed, ".nonAdjacentConflict"),
                    clientFixture(json).currentTimestamp
                )
            )
        );
    }

    function emitStoredValidatorSetUpdateCalldata() private {
        string memory json = readFixture("update_50_equal");
        ClientFixture memory client = clientFixture(json);
        emitCallGas(
            "adjacent-update-stored-valset-50",
            abi.encodeCall(
                EvmLightClientBench.updateAdjacentStoredValidatorSet,
                (
                    trustedHeight(json),
                    headerFixture(json, ".adjacent.header"),
                    compactCommitFixture(json, ".adjacent.commit"),
                    client.currentTimestamp,
                    client.processedHeight
                )
            )
        );
    }

    function emitMembershipCalldata() private {
        MembershipFixture memory proof = membershipFixture(readFixture("update_50_equal"));
        emitCallGas(
            "membership-proof-baseline",
            abi.encodeCall(
                EvmLightClientBench.verifyMembershipProof,
                (proof.root, proof.key, proof.value, proof.siblings, proof.siblingOnLeft)
            )
        );
    }

    function emitCommitCalldata(string memory name, string memory fixtureName) private {
        string memory json = readFixture(fixtureName);
        ValidatorFixture memory vals = validatorFixture(json, ".trustedValidators");
        emitCallGas(
            name,
            abi.encodeCall(
                EvmLightClientBench.verifyCommitCompact,
                (
                    vals.validators,
                    vals.powers,
                    compactCommitFixture(json, ".adjacent.commit"),
                    uintFixture(json, ".adjacent.commit.requiredPower")
                )
            )
        );
    }

    function emitCallGas(string memory name, bytes memory callData) private {
        emit CalldataGasEstimate(name, callData.length, calldataGas(callData));
        (uint256 zeros, uint256 nonzeros) = countBytes(callData);
        emit CalldataCostBreakdown(
            name,
            callData.length,
            zeros,
            nonzeros,
            zeros * 4 + nonzeros * 16,
            zeros * 10 + nonzeros * 40,
            blobBytes(callData.length)
        );
    }

    function misbehaviourInput(
        string memory json,
        ValidatorFixture memory trusted,
        ValidatorFixture memory untrusted,
        string memory prefix
    ) private pure returns (EvmLightClientBench.MisbehaviourHeaderInput memory input) {
        input = misbehaviourInput(
            json,
            trusted,
            untrusted,
            headerFixture(json, string.concat(prefix, ".header")),
            compactCommitFixture(json, string.concat(prefix, ".commit"))
        );
    }

    function misbehaviourInput(
        string memory json,
        ValidatorFixture memory trusted,
        ValidatorFixture memory untrusted,
        EvmLightClientBench.Header memory header,
        EvmLightClientBench.CompactCommit memory commit
    ) private pure returns (EvmLightClientBench.MisbehaviourHeaderInput memory input) {
        input = EvmLightClientBench.MisbehaviourHeaderInput({
            trustedHeight: trustedHeight(json),
            trustedValidators: trusted.validators,
            trustedPowers: trusted.powers,
            untrustedValidators: untrusted.validators,
            untrustedPowers: untrusted.powers,
            header: header,
            commit: commit
        });
    }

    function initializedBench(string memory name) private returns (EvmLightClientBench bench, string memory json) {
        json = readFixture(name);
        ClientFixture memory client = clientFixture(json);
        bench = new EvmLightClientBench();
        bench.initializeClient(
            EvmLightClientBench.ClientState({
                chainId: client.chainId,
                revisionNumber: client.revisionNumber,
                latestHeight: client.latestHeight,
                trustingPeriod: client.trustingPeriod,
                maxClockDrift: client.maxClockDrift,
                trustLevelNumerator: client.trustLevelNumerator,
                trustLevelDenominator: client.trustLevelDenominator,
                frozen: false
            }),
            trustedHeight(json),
            trustedConsensusState(json)
        );
    }

    function trustedConsensusState(string memory json)
        private
        pure
        returns (EvmLightClientBench.ConsensusState memory state)
    {
        state = EvmLightClientBench.ConsensusState({
            timestamp: uint64(uintFixture(json, ".trusted.timestamp")),
            appHash: bytes32Fixture(json, ".trusted.appHash"),
            nextValidatorsHash: bytes32Fixture(json, ".trusted.nextValidatorsHash"),
            processedHeight: uint64(uintFixture(json, ".trusted.processedHeight")),
            processedTime: uint64(uintFixture(json, ".trusted.processedTime")),
            exists: true
        });
    }

    function clientFixture(string memory json) private pure returns (ClientFixture memory fixture) {
        fixture.chainId = bytesFixture(json, ".client.chainId");
        fixture.revisionNumber = uint64(uintFixture(json, ".client.revisionNumber"));
        fixture.latestHeight = uint64(uintFixture(json, ".client.latestHeight"));
        fixture.trustingPeriod = uint64(uintFixture(json, ".client.trustingPeriod"));
        fixture.maxClockDrift = uint64(uintFixture(json, ".client.maxClockDrift"));
        fixture.trustLevelNumerator = uint64(uintFixture(json, ".client.trustLevelNumerator"));
        fixture.trustLevelDenominator = uint64(uintFixture(json, ".client.trustLevelDenominator"));
        fixture.currentTimestamp = uint64(uintFixture(json, ".client.currentTimestamp"));
        fixture.processedHeight = uint64(uintFixture(json, ".client.processedHeight"));
    }

    function validatorFixture(string memory json, string memory prefix)
        private
        pure
        returns (ValidatorFixture memory fixture)
    {
        fixture.validators = abi.decode(vm.parseJson(json, string.concat(prefix, ".validators")), (address[]));
        fixture.powers = uint64ArrayFixture(json, string.concat(prefix, ".powers"));
        fixture.leaves = abi.decode(vm.parseJson(json, string.concat(prefix, ".leaves")), (bytes[]));
        fixture.hash = bytes32Fixture(json, string.concat(prefix, ".hash"));
    }

    function headerFixture(string memory json, string memory prefix)
        private
        pure
        returns (EvmLightClientBench.Header memory header)
    {
        header.revisionNumber = uint64(uintFixture(json, string.concat(prefix, ".revisionNumber")));
        header.height = uint64(uintFixture(json, string.concat(prefix, ".height")));
        header.timestamp = uint64(uintFixture(json, string.concat(prefix, ".timestamp")));
        header.validatorsHash = bytes32Fixture(json, string.concat(prefix, ".validatorsHash"));
        header.nextValidatorsHash = bytes32Fixture(json, string.concat(prefix, ".nextValidatorsHash"));
        header.appHash = bytes32Fixture(json, string.concat(prefix, ".appHash"));
        header.headerHash = bytes32Fixture(json, string.concat(prefix, ".headerHash"));
        header.commitBlockHash = bytes32Fixture(json, string.concat(prefix, ".commitBlockHash"));
        header.round = uint32(uintFixture(json, string.concat(prefix, ".round")));
        header.partSetTotal = uint32(uintFixture(json, string.concat(prefix, ".partSetTotal")));
        header.partSetHash = bytes32Fixture(json, string.concat(prefix, ".partSetHash"));
    }

    function compactCommitFixture(string memory json, string memory prefix)
        private
        pure
        returns (EvmLightClientBench.CompactCommit memory commit)
    {
        commit.height = uint64(uintFixture(json, string.concat(prefix, ".height")));
        commit.blockHash = bytes32Fixture(json, string.concat(prefix, ".blockHash"));
        commit.round = uint32(uintFixture(json, string.concat(prefix, ".round")));
        commit.partSetTotal = uint32(uintFixture(json, string.concat(prefix, ".partSetTotal")));
        commit.partSetHash = bytes32Fixture(json, string.concat(prefix, ".partSetHash"));
        commit.signerBitmap = uintFixture(json, string.concat(prefix, ".signerBitmap"));
        commit.signatures = abi.decode(vm.parseJson(json, string.concat(prefix, ".signatures")), (bytes[]));
        commit.timestamps = uint64ArrayFixture(json, string.concat(prefix, ".timestamps"));
    }

    function prebuiltCommitFixture(string memory json, string memory prefix)
        private
        pure
        returns (EvmLightClientBench.PrebuiltVoteCommit memory commit)
    {
        commit.height = uint64(uintFixture(json, string.concat(prefix, ".height")));
        commit.blockHash = bytes32Fixture(json, string.concat(prefix, ".blockHash"));
        commit.round = uint32(uintFixture(json, string.concat(prefix, ".round")));
        commit.partSetTotal = uint32(uintFixture(json, string.concat(prefix, ".partSetTotal")));
        commit.partSetHash = bytes32Fixture(json, string.concat(prefix, ".partSetHash"));
        commit.signerBitmap = uintFixture(json, string.concat(prefix, ".signerBitmap"));
        commit.signatures = abi.decode(vm.parseJson(json, string.concat(prefix, ".signatures")), (bytes[]));
        commit.voteSignBytes = abi.decode(vm.parseJson(json, string.concat(prefix, ".voteSignBytes")), (bytes[]));
    }

    function canonicalEd25519Fixture(string memory json) private pure returns (CanonicalEd25519Fixture memory ed) {
        ed.pubKeys = abi.decode(vm.parseJson(json, ".canonical.ed25519.pubKeys"), (bytes32[]));
        ed.powers = uint64ArrayFixture(json, ".canonical.ed25519.powers");
        ed.leaves = abi.decode(vm.parseJson(json, ".canonical.ed25519.leaves"), (bytes[]));
        ed.hash = bytes32Fixture(json, ".canonical.ed25519.hash");
        ed.signatures = abi.decode(vm.parseJson(json, ".canonical.ed25519.signatures"), (bytes[]));
    }

    function canonicalSecp256k1Fixture(string memory json) private pure returns (CanonicalSecp256k1Fixture memory sk) {
        sk.pubKeys = abi.decode(vm.parseJson(json, ".canonical.secp256k1.pubKeys"), (bytes[]));
        sk.powers = uint64ArrayFixture(json, ".canonical.secp256k1.powers");
        sk.leaves = abi.decode(vm.parseJson(json, ".canonical.secp256k1.leaves"), (bytes[]));
        sk.hash = bytes32Fixture(json, ".canonical.secp256k1.hash");
        sk.signatures = abi.decode(vm.parseJson(json, ".canonical.secp256k1.signatures"), (bytes[]));
    }

    function canonicalVoteFixture(string memory json) private pure returns (CanonicalVoteFixture memory vote) {
        vote.chainId = bytesFixture(json, ".canonical.vote.chainId");
        vote.height = int64Fixture(json, ".canonical.vote.height");
        vote.round = int64Fixture(json, ".canonical.vote.round");
        vote.blockHash = bytes32Fixture(json, ".canonical.vote.blockHash");
        vote.partSetTotal = uint32Fixture(json, ".canonical.vote.partSetTotal");
        vote.partSetHash = bytes32Fixture(json, ".canonical.vote.partSetHash");
        vote.timestamps = int64ArrayFixture(json, ".canonical.vote.timestamps");
        vote.signBytes = abi.decode(vm.parseJson(json, ".canonical.vote.signBytes"), (bytes[]));
        vote.ethSignatures = abi.decode(vm.parseJson(json, ".canonical.vote.ethSignatures"), (bytes[]));
    }

    function canonicalBlsFixture(string memory json) private pure returns (CanonicalBlsFixture memory bls) {
        bls.available = abi.decode(vm.parseJson(json, ".canonical.bls.available"), (bool));
        if (!bls.available) {
            return bls;
        }
        bls.pubKeys = abi.decode(vm.parseJson(json, ".canonical.bls.pubKeys"), (bytes[]));
        bls.pubKeysEip2537 = abi.decode(vm.parseJson(json, ".canonical.bls.pubKeysEip2537"), (bytes[]));
        bls.powers = uint64ArrayFixture(json, ".canonical.bls.powers");
        bls.leaves = abi.decode(vm.parseJson(json, ".canonical.bls.leaves"), (bytes[]));
        bls.hash = bytes32Fixture(json, ".canonical.bls.hash");
        bls.aggregateSig = bytesFixture(json, ".canonical.bls.aggregateSigEip2537");
        bls.aggregatePubKey = bytesFixture(json, ".canonical.bls.aggregatePubKeyEip2537");
        bls.hashedMessage = bytesFixture(json, ".canonical.bls.hashedMessageEip2537");
        bls.wrongMessageHashed = bytesFixture(json, ".canonical.bls.wrongMessageHashedEip2537");
        bls.missingSignerAggregateSig = bytesFixture(json, ".canonical.bls.missingSignerAggregateSigEip2537");
        bls.signerBitmap = uintFixture(json, ".canonical.bls.signerBitmap");
        bls.message = bytesFixture(json, ".canonical.bls.message");
    }

    function membershipFixture(string memory json) private pure returns (MembershipFixture memory fixture) {
        fixture.root = bytes32Fixture(json, ".membership.root");
        fixture.key = bytesFixture(json, ".membership.key");
        fixture.value = bytesFixture(json, ".membership.value");
        fixture.siblings = abi.decode(vm.parseJson(json, ".membership.siblings"), (bytes32[]));
        fixture.siblingOnLeft = abi.decode(vm.parseJson(json, ".membership.siblingOnLeft"), (bool[]));
    }

    function iavlFixture(string memory json) private pure returns (IavlFixture memory iavl) {
        iavl.root = bytes32Fixture(json, ".iavl.root");
        iavl.key = bytesFixture(json, ".iavl.key");
        iavl.value = bytesFixture(json, ".iavl.value");
        iavl.leafPrefix = bytesFixture(json, ".iavl.leafPrefix");
        iavl.innerPrefixes = abi.decode(vm.parseJson(json, ".iavl.innerPrefixes"), (bytes[]));
        iavl.innerSuffixes = abi.decode(vm.parseJson(json, ".iavl.innerSuffixes"), (bytes[]));
        iavl.existenceProof = bytesFixture(json, ".iavl.existenceProof");
        iavl.commitmentProof = bytesFixture(json, ".iavl.commitmentProof");
    }

    function iavlMaxDepthFixture(string memory json) private pure returns (IavlFixture memory iavl) {
        iavl.root = bytes32Fixture(json, ".iavlMaxDepth.root");
        iavl.key = bytesFixture(json, ".iavlMaxDepth.key");
        iavl.value = bytesFixture(json, ".iavlMaxDepth.value");
        iavl.leafPrefix = bytesFixture(json, ".iavlMaxDepth.leafPrefix");
        iavl.innerPrefixes = abi.decode(vm.parseJson(json, ".iavlMaxDepth.innerPrefixes"), (bytes[]));
        iavl.innerSuffixes = abi.decode(vm.parseJson(json, ".iavlMaxDepth.innerSuffixes"), (bytes[]));
        iavl.existenceProof = bytesFixture(json, ".iavlMaxDepth.existenceProof");
        iavl.commitmentProof = bytesFixture(json, ".iavlMaxDepth.commitmentProof");
    }

    function readFixture(string memory name) private view returns (string memory) {
        return vm.readFile(string.concat("test/fixtures/", name, ".json"));
    }

    function trustedHeight(string memory json) private pure returns (uint64) {
        return uint64(uintFixture(json, ".trusted.height"));
    }

    function uintFixture(string memory json, string memory key) private pure returns (uint256) {
        return abi.decode(vm.parseJson(json, key), (uint256));
    }

    function int64Fixture(string memory json, string memory key) private pure returns (int64) {
        return int64(int256(uintFixture(json, key)));
    }

    function uint32Fixture(string memory json, string memory key) private pure returns (uint32) {
        return uint32(uintFixture(json, key));
    }

    function bytes32Fixture(string memory json, string memory key) private pure returns (bytes32) {
        return abi.decode(vm.parseJson(json, key), (bytes32));
    }

    function bytesFixture(string memory json, string memory key) private pure returns (bytes memory) {
        return abi.decode(vm.parseJson(json, key), (bytes));
    }

    function uint64ArrayFixture(string memory json, string memory key) private pure returns (uint64[] memory out) {
        uint256[] memory raw = abi.decode(vm.parseJson(json, key), (uint256[]));
        out = new uint64[](raw.length);
        for (uint256 i; i < raw.length; ++i) {
            out[i] = uint64(raw[i]);
        }
    }

    function int64ArrayFixture(string memory json, string memory key) private pure returns (int64[] memory out) {
        uint256[] memory raw = abi.decode(vm.parseJson(json, key), (uint256[]));
        out = new int64[](raw.length);
        for (uint256 i; i < raw.length; ++i) {
            out[i] = int64(int256(raw[i]));
        }
    }

    function totalPower(uint64[] memory powers) private pure returns (uint256 total) {
        for (uint256 i; i < powers.length; ++i) {
            total += powers[i];
        }
    }

    function uint64Timestamp(int64 value) private pure returns (uint64) {
        require(value >= 0, "negative timestamp");
        return uint64(uint256(int256(value)));
    }

    function calldataGas(bytes memory data) private pure returns (uint256 gasCost) {
        for (uint256 i; i < data.length; ++i) {
            gasCost += data[i] == 0 ? 4 : 16;
        }
    }

    function countBytes(bytes memory data) private pure returns (uint256 zeros, uint256 nonzeros) {
        for (uint256 i; i < data.length; ++i) {
            if (data[i] == 0) ++zeros;
            else ++nonzeros;
        }
    }

    // blobBytes returns the blob-equivalent payload footprint. EIP-4844 blobs
    // carry 4096 BLS12-381 field elements of 32 bytes each, but only 31 bytes
    // of each element can carry arbitrary data (the high byte is reserved so
    // the value fits below the field modulus). This is the rounded-up data
    // volume that would be charged via blob gas (`tx.blobGasUsed` style).
    function blobBytes(uint256 byteLength) private pure returns (uint256) {
        if (byteLength == 0) return 0;
        return ((byteLength + 30) / 31) * 32;
    }
}
