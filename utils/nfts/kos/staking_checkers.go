package utils_kos

import (
	"context"
	"errors"
	"fmt"
	"nbc-backend-api-v2/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

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
Checks if a keychain with ID `keychainId` has already been staked in a specific staking pool.
*/
func CheckIfKeychainStaked(collection *mongo.Collection, stakingPoolId, keychainId int) (bool, error) {
	// call `GetAllStakedKeychainIDs` to get all the keychain IDs that have been staked in the staking pool
	stakedKeychainIDs, err := GetAllStakedKeychainIDs(collection, stakingPoolId)
	if err != nil {
		return true, err
	}

	// check if the keychain ID is in the list of staked keychain IDs
	for _, stakedKeychainID := range stakedKeychainIDs {
		if stakedKeychainID == keychainId {
			return true, nil
		}
	}

	return false, nil
}

/*
Gets all the keychain IDs that have been staked in a specific staking pool.
*/
func GetAllStakedKeychainIDs(collection *mongo.Collection, stakingPoolId int) ([]int, error) {
	if collection.Name() != "RHStakingPool" {
		return nil, errors.New("invalid collection name")
	}

	pipeline := mongo.Pipeline{
		bson.D{{"$match", bson.D{{"stakingPoolID", stakingPoolId}}}},                                       // match documents with stakingPoolID equal to 1
		bson.D{{"$unwind", "$activeSubpools"}},                                                             // unwind the activeSubpools array to get separate document for each subpool
		bson.D{{"$match", bson.D{{"activeSubpools.stakedKeychainId", bson.D{{"$ne", -1}}}}}},               // exclude subpools with stakedKeychainId equal to -1
		bson.D{{"$project", bson.D{{"_id", 0}, {"stakedKeychainId", "$activeSubpools.stakedKeychainId"}}}}, // project only the stakedKeychainId
	}

	cursor, err := collection.Aggregate(context.Background(), pipeline)
	if err != nil {
		return nil, err
	}

	defer cursor.Close(context.Background())

	var keychainIDs []int
	for cursor.Next(context.Background()) {
		var result struct {
			StakedKeychainID int `bson:"stakedKeychainId"`
		}
		if err := cursor.Decode(&result); err != nil {
			return nil, err
		}

		keychainIDs = append(keychainIDs, result.StakedKeychainID)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	defer cursor.Close(context.Background())

	return keychainIDs, nil
}

/*
Checks if a superior keychain with ID `superiorKeychainId` has already been staked in a specific staking pool.
*/
func CheckIfSuperiorKeychainStaked(collection *mongo.Collection, stakingPoolId, superiorKeychainId int) (bool, error) {
	// call `GetAllStakedSuperiorKeychainIDs` to get all the superior keychain IDs that have been staked in the staking pool
	stakedSuperiorKeychainIDs, err := GetAllStakedSuperiorKeychainIDs(collection, stakingPoolId)
	if err != nil {
		return true, err
	}

	// check if the superior keychain ID is in the list of staked superior keychain IDs
	for _, stakedSuperiorKeychainID := range stakedSuperiorKeychainIDs {
		if stakedSuperiorKeychainID == superiorKeychainId {
			return true, nil
		}
	}

	return false, nil
}

/*
Gets all the superior keychain IDs that have been staked in a specific staking pool.
*/
func GetAllStakedSuperiorKeychainIDs(collection *mongo.Collection, stakingPoolId int) ([]int, error) {
	if collection.Name() != "RHStakingPool" {
		return nil, errors.New("invalid collection name")
	}

	pipeline := mongo.Pipeline{
		bson.D{{"$match", bson.D{{"stakingPoolID", 1}}}},                                             // match the staking pool ID with `1`
		bson.D{{"$unwind", "$activeSubpools"}},                                                       // unwinds the activeSubpools array to get separate document for each `Subpool` in the array
		bson.D{{"$match", bson.D{{"activeSubpools.stakedSuperiorKeychainId", bson.D{{"$ne", -1}}}}}}, // filter out documents where `stakedSuperiorKeychainId` is `-1`
		bson.D{{"$project", bson.D{{"_id", 0}, {"stakedSuperiorKeychainId", "$activeSubpools.stakedSuperiorKeychainId"}}}},
	}

	cursor, err := collection.Aggregate(context.Background(), pipeline)
	if err != nil {
		return nil, err
	}

	defer cursor.Close(context.Background())

	fmt.Println(cursor)

	var superiorKeychainIDs []int
	for cursor.Next(context.Background()) {
		var result struct {
			StakedSuperiorKeychainID int `bson:"stakedSuperiorKeychainId"`
		}
		if err := cursor.Decode(&result); err != nil {
			return nil, err
		}

		superiorKeychainIDs = append(superiorKeychainIDs, result.StakedSuperiorKeychainID)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	defer cursor.Close(context.Background())

	return superiorKeychainIDs, nil
}
