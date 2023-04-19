package utils

import (
	"encoding/json"
	"math/big"
	"nbc-backend-api-v2/models"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
)

/*
`GetExplicitOwnerships` calls the `explicitOwnershipsOf` method of a specific NFT contract, returning an array of `ExplicitOwnership` structs.
*/
func GetExplicitOwnerships(
	apiKey string,
	concat bool,
	rawClientUrl string,
	abiPath string,
	contractAddress string,
	collectionSize int,
	caller bind.ContractCaller,
	transactor bind.ContractTransactor,
	filterer bind.ContractFilterer,
) ([]models.ExplicitOwnership, error) {
	// loads the contract
	contract, err := LoadContract(
		apiKey,
		concat,
		rawClientUrl,
		abiPath,
		contractAddress,
		caller,
		transactor,
		filterer,
	)
	if err != nil {
		return nil, err
	}

	// create an array of all ids of the NFT collection (the collection size)
	allIds := make([]*big.Int, collectionSize)
	for i := 1; i <= collectionSize; i++ {
		allIds[i-1] = big.NewInt(int64(i))
	}

	var rawResult []interface{}

	// calls the `explicitOwnershipsOf` method of the contract
	err = contract.Call(nil, &rawResult, "explicitOwnershipsOf", allIds)
	if err != nil {
		return nil, err
	}

	// the result returned in `rawResult` is an array of unknown interfaces.
	// upon further inspection, each interface of the 0th index (only result) is rather an ExplicitOwnership struct, so we need to convert it in order to be able to loop through each element.
	var result []models.ExplicitOwnership
	data, err := json.Marshal(rawResult[0])
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(data, &result)
	if err != nil {
		return nil, err
	}

	// return all the results formatted as an array of ExplicitOwnership structs
	formattedResults := []models.ExplicitOwnership{}

	for _, ownership := range result {
		formattedResults = append(
			formattedResults,
			models.ExplicitOwnership{
				Addr:           ownership.Addr,
				StartTimestamp: ownership.StartTimestamp,
				Burned:         ownership.Burned,
				ExtraData:      ownership.ExtraData,
			},
		)
	}

	return formattedResults, nil
}

/*
`GetOwnerIDs` gets ALL token IDs owned by `ownerAddress` from a specific NFT collection.
*/
func GetOwnerIDs(
	apiKey string,
	concat bool,
	rawClientUrl string,
	abiPath string,
	contractAddress string,
	ownerAddress string,
	caller bind.ContractCaller,
	transactor bind.ContractTransactor,
	filterer bind.ContractFilterer,
) (*models.OwnershipData, error) {
	// loads the contract
	contract, err := LoadContract(
		apiKey,
		concat,
		rawClientUrl,
		abiPath,
		contractAddress,
		caller,
		transactor,
		filterer,
	)
	if err != nil {
		return nil, err
	}

	var rawResult []interface{}

	// calls the `tokensOfOwner` method of the contract
	err = contract.Call(nil, &rawResult, "tokensOfOwner", common.HexToAddress(ownerAddress))
	if err != nil {
		return nil, err
	}

	// the result returned in `rawResult` is an array of unknown interfaces.
	// upon further inspection, each interface of the 0th index (only result) is rather a big integer, so we will store them in `result`.
	var result []*big.Int
	data, err := json.Marshal(rawResult[0])
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(data, &result)
	if err != nil {
		return nil, err
	}

	ownershipData := models.OwnershipData{
		Addr:     common.HexToAddress(ownerAddress),
		TokenIDs: result,
	}

	return &ownershipData, nil
}
