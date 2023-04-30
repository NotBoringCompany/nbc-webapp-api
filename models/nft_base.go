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

/*
A base struct that represents the data of any NFT.
*/
type NFTData struct {
	Name      string      `json:"name"`
	ImageUrl  string      `json:"imageUrl"`
	TokenID   int         `json:"tokenID"` // the token id just in case it doesn't exist in Metadata
	Metadata  interface{} `json:"metadata"`
	Stakeable bool        `json:"stakeable"`
}
