// SPDX-License-Identifier: Apache-2.0
pragma solidity =0.6.12;
pragma experimental ABIEncoderV2;

import "./vendor/chengwenxi-ed25519/Ed25519.sol";

contract Ed25519VerifierBench {
    function verify(bytes32 publicKey, bytes memory signature, bytes memory message) public pure returns (bool) {
        require(signature.length == 64, "bad signature length");

        bytes32 r;
        bytes32 s;
        assembly {
            r := mload(add(signature, 32))
            s := mload(add(signature, 64))
        }
        return Ed25519.verify(publicKey, r, s, message);
    }
}
