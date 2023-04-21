package utils_kos

import (
	"context"
	"errors"
	"nbc-backend-api-v2/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

/*
Gets all the key IDs that have been staked in a specific staking pool.
*/
func GetAllStakedKeyIDs(collection *mongo.Collection, stakingPoolId int) ([]int, error) {
	if collection.Name() != "RHStakingPool" {
		return nil, errors.New("invalid collection name")
	}

	pipeline := mongo.Pipeline{
		bson.D{{"$match", bson.D{{"stakingPoolID", stakingPoolId}}}},              // match the staking pool ID with `stakingPoolId`
		bson.D{{"$unwind", "$activeSubpools"}},                                    // unwinds the activeSubpools array to get separate document for each `Subpool` in the array
		bson.D{{"$unwind", "$activeSubpools.stakedKeys"}},                         // unwinds the stakedKeys array to get separate document for each `StakedKey` in the array
		bson.D{{"$group", bson.D{{"_id", "$activeSubpools.stakedKeys.tokenID"}}}}, // groups the documents by the TokenID field
		bson.D{{"$project", bson.D{{"_id", 0}, {"TokenID", "$_id"}}}},             // only project the TokenID field
	}

	cursor, err := collection.Aggregate(context.Background(), pipeline)
	if err != nil {
		return nil, err
	}

	var tokenIDs []int
	for cursor.Next(context.Background()) {
		var result struct {
			TokenID int `bson:"TokenID"`
		}
		if err := cursor.Decode(&result); err != nil {
			return nil, err
		}
		tokenIDs = append(tokenIDs, result.TokenID)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	defer cursor.Close(context.Background())

	return tokenIDs, nil
}

/*
Checks if ANY of the keys that a user wants to add to a subpool has already been staked in that particular staking pool.
If even just one key has already been staked, then the entire batch of keys will be rejected.
*/
func CheckIfKeysStaked(collection *mongo.Collection, stakingPoolId int, keys []*models.KOSSimplifiedMetadata) (bool, error) {
	if collection.Name() != "RHStakingPool" {
		return true, errors.New("invalid collection name") // defaults to true if an error occurs
	}

	// for each key in `keys`, check if it has already been staked in the staking pool by calling `CheckIfKeyStaked`
	for _, key := range keys {
		isStaked, err := CheckIfKeyStaked(collection, stakingPoolId, key)
		if err != nil {
			return true, err
		}

		if isStaked {
			return true, nil
		}
	}

	return false, nil
}

/*
Checks if a key that a user wants to add to a subpool has already been staked in that particular staking pool.
*/
func CheckIfKeyStaked(collection *mongo.Collection, stakingPoolId int, key *models.KOSSimplifiedMetadata) (bool, error) {
	// call `GetAllStakedKeyIDs` to get all the key IDs that have been staked in the staking pool
	stakedKeyIDs, err := GetAllStakedKeyIDs(collection, stakingPoolId)
	if err != nil {
		return true, err
	}

	// check if the key ID is in the list of staked key IDs
	for _, stakedKeyID := range stakedKeyIDs {
		if stakedKeyID == key.TokenID {
			return true, nil
		}
	}

	return false, nil
}
