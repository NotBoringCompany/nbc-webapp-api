package utils_kos

import (
	"encoding/json"
	"io"
	"log"
	"math/big"
	"nbc-backend-api-v2/models"
	"net/http"
	"os"
)

/*
`FetchMetadata` fetches a Key Of Salvation's metadata from Pinata (IPFS) and returns it as a `KOSMetadata` struct instance.

	`tokenId` the token ID of the Key
*/
func FetchMetadata(tokenId *big.Int) *models.KOSMetadata {
	// create a new HTTP client
	client := &http.Client{}

	url := os.Getenv("KOS_URI") + tokenId.String() + ".json"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatal("Error while creating request", err)
		return nil
	}

	// send the request and get the response
	res, err := client.Do(req)
	if err != nil {
		log.Fatal("Error while sending request", err)
		return nil
	}

	defer res.Body.Close()

	// read the response body into a byte array
	body, err := io.ReadAll(res.Body)
	if err != nil {
		log.Fatal("Error while reading response body", err)
		return nil
	}

	// unmarshal the bytes array into a `KOSMetadata` struct instance.
	var metadata models.KOSMetadata
	err = json.Unmarshal(body, &metadata)
	if err != nil {
		log.Fatal("Error while unmarshalling response body", err)
		return nil
	}

	return &metadata
}

/*
`FetchSimplifiedMetadata` returns a more simplified version of a Key Of Salvation's metadata (returns a KOSSimplifiedMetadata struct).

	`tokenId` the token ID of the Key
*/
func FetchSimplifiedMetadata(tokenId *big.Int) *models.KOSSimplifiedMetadata {
	metadata := FetchMetadata(tokenId)

	simplifiedMetadata := &models.KOSSimplifiedMetadata{
		TokenID:        tokenId,
		HouseTrait:     metadata.Attributes[3].Value.(string),
		TypeTrait:      metadata.Attributes[7].Value.(string),
		LuckTrait:      metadata.Attributes[0].Value.(float64),
		LuckBoostTrait: metadata.Attributes[1].Value.(float64),
	}

	return simplifiedMetadata
}
