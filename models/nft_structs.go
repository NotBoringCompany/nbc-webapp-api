package models

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

// an explicit ownership struct to read the data from the contract's `explicitOwnershipsOf` method
type ExplicitOwnership struct {
	Addr           common.Address `json:"addr"`
	StartTimestamp uint64         `json:"startTimestamp"`
	Burned         bool           `json:"burned"`
	ExtraData      *big.Int       `json:"extraData"`
}

// gets ALL token IDs owned by `Addr` from a specific NFT collection.
type OwnershipData struct {
	Addr     common.Address `json:"addr"`
	TokenIDs []*big.Int     `json:"tokenIDs"`
}
