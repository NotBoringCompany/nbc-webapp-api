package utils_kos

import (
	"fmt"
	"math/big"
	"nbc-backend-api-v2/models"
	UtilsNFT "nbc-backend-api-v2/utils/nfts"
)

/*
Calls `GetExplicitOwnerships` for the Key Of Salvation contract.
*/
func kosExplicitOwnership() ([]models.ExplicitOwnership, error) {
	// calls `GetExplicitOwnerships` for the Key Of Salvation with the given address
	ownerships, err := UtilsNFT.GetExplicitOwnerships(
		"ALCHEMY_ETH_API_KEY",
		true,
		"https://eth-mainnet.g.alchemy.com/v2/",
		"abi/KeyOfSalvation.json",
		"KOS_ADDRESS",
		5000,
		nil,
		nil,
		nil,
	)
	if err != nil {
		return nil, err
	}

	return ownerships, nil
}

/*
`OwnerIDs` returns the owned token IDs for the given `address` for the KOS collection.

	`address` the EVM address of the owner
*/
func OwnerIDs(address string) ([]*big.Int, error) {
	// calls `GetOwnerIds` for the Key Of Salvation contract with the given address
	ownerIds, err := UtilsNFT.GetOwnerIDs(
		"ALCHEMY_ETH_API_KEY",
		true,
		"https://eth-mainnet.g.alchemy.com/v2/",
		"abi/KeyOfSalvation.json",
		"KOS_ADDRESS",
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
`VerifyOwnership` checks if `address` still owns ANY of the mentioned `ids` for the KOS collection.

If even just one of the ids are no longer owned by `address`, this function returns false.

Called for staking purposes. `ids` should be the ids of the NFTs the user has staked in a PARTICULAR pool at a single time.

For multiple pools, this function should be called multiple times, each for each pool and with different IDs.

	`address` the EVM address of the owner
	`ids` the token IDs to check
*/
func VerifyOwnership(address string, ids []int) (bool, error) {
	currentOwnedIds, err := OwnerIDs(address)
	fmt.Println("Current owned ids: ", currentOwnedIds)
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
			fmt.Println("Id not found: ", id)
			return false, nil
		}
	}

	fmt.Println("Ownership verified: true")
	return true, nil
}
