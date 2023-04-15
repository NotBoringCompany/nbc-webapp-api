package kos

import (
	"encoding/json"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

// an explicit ownership struct to read the data from `explicitOwnershipsOf`
type ExplicitOwnership struct {
	Addr           common.Address `json:"addr"`
	StartTimestamp uint64         `json:"startTimestamp"`
	Burned         bool           `json:"burned"`
	ExtraData      *big.Int       `json:"extraData"`
}

func GetHolderData() ([]ExplicitOwnership, error) {
	kos, err := loadKOS(nil, nil, nil)

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
	var results []ExplicitOwnership
	data, err := json.Marshal(rawResult[0])
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(data, &results)
	if err != nil {
		return nil, err
	}

	// return all the results formatted as an array of ExplicitOwnership structs
	var formattedResult []ExplicitOwnership

	for _, ownership := range results {
		formattedResult = append(formattedResult, ExplicitOwnership{ownership.Addr, ownership.StartTimestamp, ownership.Burned, ownership.ExtraData})
	}

	return formattedResult, nil
}
