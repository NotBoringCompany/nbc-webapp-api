package utils_keychain

import (
	"fmt"
	"math/big"
	UtilsNFT "nbc-backend-api-v2/utils/nfts"
)

/*
`OwnerIDs` returns the owned token IDs for the given `address` for the Keychain collection.

	`address` the EVM address of the owner
*/
func OwnerIDs(address string) ([]*big.Int, error) {
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

/*
Verifies that `address` owns ALL of the mentioned `ids` for the Keychain collection.

	`address` the EVM address of the owner
	`ids` the token IDs to verify
*/
func VerifyOwnership(address string, ids []int) (bool, error) {
	currentOwnedIds, err := OwnerIDs(address)
	if err != nil {
		return false, err
	}

	for _, id := range ids {
		found := false
		// check if `id` exists in `currentOwnedIds`. the moment one id is not owned, return false
		for _, currentOwnedId := range currentOwnedIds {
			if currentOwnedId.Cmp(big.NewInt(int64(id))) == 0 {
				found = true
				break
			}
		}

		if !found {
			return false, nil
		}
	}

	fmt.Printf("All keychain `ids` are owned by `address` %s", address)
	return true, nil
}