package api

import (
	"math/big"
	UtilsNFT "nbc-backend-api-v2/utils/nfts"
)

/*
`KeychainOwnerIDs` returns the owned token IDs for the given `address` for the Keychain collection.

	`address` the EVM address of the owner
*/
func KeychainOwnerIDs(address string) ([]*big.Int, error) {
	// calls `GetOwnerIds` for the Keychain contract with the given address
	ownerIds, err := UtilsNFT.GetOwnerIDs(
		"ALCHEMY_ETH_API_KEY",
		true,
		"https://eth-mainnet.g.alchemy.com/v2/",
		"abi/Keychain.json",
		"KEYCHAIN_ADDRESS",
		address,
		nil,
		nil,
		nil,
	)
	if err != nil {
		return nil, err
	}

	return ownerIds.TokenIDs, nil
}
