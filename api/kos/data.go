package kos

import (
	"encoding/json"
	"math/big"
	"nbc-backend-api-v2/models"
	"nbc-backend-api-v2/utils"
)

func GetExplicitOwnerships() ([]models.ExplicitOwnership, error) {
	kos, err := utils.LoadContract(
		"ALCHEMY_ETH_API_KEY",
		true,
		"https://eth-mainnet.g.alchemy.com/v2/",
		"abi/KeyOfSalvation.json",
		"0x34BFF2Dbf20cF39dB042cb68D42D6d06fdbd85D3",
		nil,
		nil,
		nil,
	)

	if err != nil {
		return nil, err
	}

	//create an array of 1 to 5000 (all ids for the NFTs)
	allIds := make([]*big.Int, 5000)
	for i := 1; i <= 5000; i++ {
		allIds[i-1] = big.NewInt(int64(i))
	}

	var rawResult []interface{}

	ownerships := kos.Call(nil, &rawResult, "explicitOwnershipsOf", allIds)
	if ownerships != nil {
		return nil, err
	}

	// the result returned in `rawResult` is an array of unknown interfaces.
	// upon further inspection, each interface is rather an ExplicitOwnership struct, so we need to convert it in order to be able to loop through each element.
	var results []models.ExplicitOwnership
	data, err := json.Marshal(rawResult[0])
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(data, &results)
	if err != nil {
		return nil, err
	}

	// return all the results formatted as an array of ExplicitOwnership structs
	formattedResults := []models.ExplicitOwnership{}

	for _, ownership := range results {
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
