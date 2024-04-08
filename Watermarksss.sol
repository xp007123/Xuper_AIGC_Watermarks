// SPDX-License-Identifier: MIT
pragma solidity ^0.5.8;

contract Watermarkasss {
    struct Info {
        string nonce;
        string otherInfo;
    }

    mapping(address => Info) private infoMap;

    function storeInfo(string memory nonce, string memory otherInfo) public {
        infoMap[msg.sender] = Info(nonce, otherInfo);
    }

    function retrieveNonce() public view returns (string memory) {
        return infoMap[msg.sender].nonce;
    }
}
