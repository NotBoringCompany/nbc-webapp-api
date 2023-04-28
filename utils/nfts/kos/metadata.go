package utils_kos

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"nbc-backend-api-v2/models"
	"net/http"
	"os"
	"sort"
	"sync"
)

var (
	client        = &http.Client{}
	metadataCache sync.Map
)

/*
`FetchMetadata` fetches a Key Of Salvation's metadata from Pinata (IPFS) and returns it as a `KOSMetadata` struct instance.

	`tokenId` the token ID of the Key
*/
func FetchMetadata(tokenId int) (*models.KOSMetadata, error) {
	// // check if metadata is in cache
	// if metadata, ok := metadataCache.Load(tokenId); ok {
	// 	log.Println("Here!")
	// 	return metadata.(*models.KOSMetadata), nil
	// }

	url := os.Getenv("KOS_URI") + fmt.Sprint(tokenId) + ".json"
	req, err := http.NewRequestWithContext(context.Background(), "GET", url, nil)
	if err != nil {
		return nil, err
	}

	// send the request and get the response
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	// check if response is OK
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("received non-200 status code: %d", res.StatusCode)
	}

	// unmarshal the response body into a `KOSMetadata` struct instance
	var metadata models.KOSMetadata
	err = json.NewDecoder(res.Body).Decode(&metadata)
	if err != nil {
		return nil, err
	}

	log.Println("here 2!")

	// // cache metadata
	// metadataCache.Store(tokenId, &metadata)

	return &metadata, nil
}

/*
`FetchSimplifiedMetadata` returns a more simplified version of a Key Of Salvation's metadata (returns a KOSSimplifiedMetadata struct).

	`tokenId` the token ID of the Key
*/

func FetchSimplifiedMetadata(tokenId int) (models.KOSSimplifiedMetadata, error) {
	// // Check if simplified metadata is in cache
	// if metadata, ok := metadataCache.Load(tokenId); ok {
	// 	log.Println("Here!")
	// 	return metadata.(*models.KOSSimplifiedMetadata), nil
	// }

	metadata, err := FetchMetadata(tokenId)
	if err != nil {
		return models.KOSSimplifiedMetadata{}, err
	}

	fmt.Println(metadata)
	fmt.Println("animation url", metadata.AnimationUrl)

	simplifiedMetadata := models.KOSSimplifiedMetadata{
		TokenID:        tokenId,
		AnimationUrl:   metadata.AnimationUrl,
		HouseTrait:     metadata.Attributes[3].Value.(string),
		TypeTrait:      metadata.Attributes[7].Value.(string),
		LuckTrait:      metadata.Attributes[0].Value.(float64),
		LuckBoostTrait: 1 + (metadata.Attributes[1].Value.(float64) / 100),
	}

	// // Cache simplified metadata
	// metadataCache.Store(tokenId, simplifiedMetadata)

	return simplifiedMetadata, nil
}

func FetchSimplifiedMetadataConcurrent(tokenIds []int) ([]*models.KOSSimplifiedMetadata, error) {
	type result struct {
		index    int
		metadata *models.KOSSimplifiedMetadata
		err      error
	}

	// use a buffered channel to limit the number of concurrent goroutines
	ch := make(chan result, len(tokenIds))

	// create worker goroutines
	for i, id := range tokenIds {
		go func(index, tokenId int) {
			metadata, err := FetchSimplifiedMetadata(tokenId)
			ch <- result{index, metadata, err}
		}(i, id)
	}

	// collect results
	results := make([]result, len(tokenIds))
	for range tokenIds {
		res := <-ch
		results[res.index] = res
	}

	// check for errors
	var errors []error
	for _, res := range results {
		if res.err != nil {
			errors = append(errors, res.err)
		}
	}
	if len(errors) > 0 {
		return nil, fmt.Errorf("encountered %d errors: %v", len(errors), errors)
	}

	// sort results by original order
	sort.Slice(results, func(i, j int) bool {
		return results[i].index < results[j].index
	})

	// extract simplified metadata from results
	simplifiedMetadata := make([]*models.KOSSimplifiedMetadata, len(results))
	for i, res := range results {
		simplifiedMetadata[i] =
			res.metadata
	}

	return simplifiedMetadata, nil
}

// func FetchSimplifiedMetadata(tokenId int) *models.KOSSimplifiedMetadata {
// 	metadata := FetchMetadata(tokenId)

// 	simplifiedMetadata := &models.KOSSimplifiedMetadata{
// 		TokenID:        tokenId,
// 		HouseTrait:     metadata.Attributes[3].Value.(string),
// 		TypeTrait:      metadata.Attributes[7].Value.(string),
// 		LuckTrait:      metadata.Attributes[0].Value.(float64),
// 		LuckBoostTrait: 1 + (metadata.Attributes[1].Value.(float64) / 100),
// 	}

// 	return simplifiedMetadata
// }

/*
Gets the simplified metadata struct instance for each key ID.
*/
func GetMetadataFromIDs(keyIds []int) []*models.KOSSimplifiedMetadata {
	var metadatas []*models.KOSSimplifiedMetadata
	for _, id := range keyIds {
		metadata, err := FetchSimplifiedMetadata(id)
		if err != nil {
			log.Fatal("Error while fetching metadata", err)
		}
		metadatas = append(metadatas, metadata)
	}

	return metadatas
}
