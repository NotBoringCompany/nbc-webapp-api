package utils_kos

import (
	"context"
	"errors"
	"fmt"
	"nbc-backend-api-v2/configs"
	"nbc-backend-api-v2/models"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

/*
For ALL staking pools, this function will get the staker from each active subpool and check whether the keys, keychain and/or superior keychain that they staked in each subpool are still owned by them.
if not, the subpool will automatically be removed from `ActiveSubpools` and moved to `ClosedSubpools`, change Banned to true and impose a BannedData instance on the staker.
*/
func VerifyOwnership(collection *mongo.Collection) (bool, error) {
	if collection.Name() != "RHStakingPool" {
		return false, errors.New("Collection must be RHStakingPool")
	}

	// we get the list of all subpools.
	subpools, err := GetAllActiveSubpools(collection)
	if err != nil {
		return true, err // defaults to true if an error occurs
	}

	/// TO DO HERE!!
}

/*
Gets all staking pools from `RHStakingPool` and returns them as a slice of `StakingPool` instances.
*/
func GetAllStakingPools(collection *mongo.Collection) ([]*models.StakingPool, error) {
	if collection.Name() != "RHStakingPool" {
		return nil, errors.New("Collection must be RHStakingPool") // defaults to false if an error occurs
	}

	// get ALL staking pools from `RHStakingPool` collection
	var stakingPools []*models.StakingPool
	cursor, err := collection.Find(context.Background(), bson.D{})
	if err != nil {
		return nil, err
	}

	defer cursor.Close(context.Background())
	for cursor.Next(context.Background()) {
		var stakingPool models.StakingPool
		if err := cursor.Decode(&stakingPool); err != nil {
			return nil, err
		}
		stakingPools = append(stakingPools, &stakingPool)
	}
	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return stakingPools, nil
}

/*
Gets all active subpools from each staking pool in `RHStakingPool` and returns them as a slice of `StakingSubpool` instances.
*/
func GetAllActiveSubpools(collection *mongo.Collection) ([]*models.StakingSubpool, error) {
	if collection.Name() != "RHStakingPool" {
		return nil, errors.New("Collection must be RHStakingPool")
	}

	// retrieve all staking pools
	cursor, err := collection.Find(context.Background(), bson.D{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	// get all active subpools from each staking pool.
	// we don't mind if a staking pool has no active subpools or that this `activeSubpools` has multiple subpools with the same ID (as its from different staking pools)
	var activeSubpools []*models.StakingSubpool
	for cursor.Next(context.Background()) {
		var stakingPool models.StakingPool
		if err := cursor.Decode(&stakingPool); err != nil {
			return nil, err
		}
		activeSubpools = append(activeSubpools, stakingPool.ActiveSubpools...)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return activeSubpools, nil
}

/*
Gets all stakers from all active subpools in `RHStakingPool` and returns them as a slice of `Staker` instances.
*/
func GetStakersFromActiveSubpools(collection *mongo.Collection) ([]*models.Staker, error) {
	// fetch all staking pools from `RHStakingPool` collection
	var stakingPools []*models.StakingPool
	cursor, err := collection.Find(context.Background(), bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	for cursor.Next(context.Background()) {
		var stakingPool models.StakingPool
		if err := cursor.Decode(&stakingPool); err != nil {
			return nil, err
		}
		stakingPools = append(stakingPools, &stakingPool)
	}
	if err := cursor.Err(); err != nil {
		return nil, err
	}

	// interate over active subpools of each StakingPool and fetch the staker
	var stakers []*models.Staker
	for _, stakingPool := range stakingPools {
		for _, subpool := range stakingPool.ActiveSubpools {
			stakerObjId := subpool.Staker
			var staker models.Staker
			// get the RHStakerData collection, find and match the staker object ID, store in `staker` variable
			err := configs.GetCollections(configs.DB, "RHStakerData").FindOne(context.Background(), bson.M{"_id": stakerObjId}).Decode(&staker)
			if err != nil {
				return nil, err
			}
			stakers = append(stakers, &staker)
		}
	}

	return stakers, nil
}

/*
Checks if the time now has passed the `StartTime` for Staking Pool ID `stakingPoolId`.
if yes, stakers are no longer allowed to add subpools into that staking pool.
It also checks if the time now is before the `EntryAllowance` for Staking Pool ID `stakingPoolId`.
If yes, stakers are not allowed to add subpools into that staking pool.
*/
func CheckPoolTimeAllowanceExceeded(collection *mongo.Collection, stakingPoolId int) (bool, error) {
	if collection.Name() != "RHStakingPool" {
		return true, errors.New("invalid collection name") // defaults to true if an error occurs
	}

	// get the staking pool document with the staking pool ID
	filter := bson.M{"stakingPoolID": stakingPoolId}
	var stakingPool models.StakingPool
	err := collection.FindOne(context.Background(), filter).Decode(&stakingPool)
	if err != nil {
		return true, err
	}

	// check if the time now has passed the `StartTime` for the staking pool
	now := time.Now()
	if now.Before(stakingPool.EntryAllowance) {
		return true, nil // entry is not allowed yet.
	} else {
		if now.After(stakingPool.StartTime) {
			return true, nil // time allowance exceeded
		} else {
			return false, nil // time allowance not exceeded
		}
	}
}

/*
Multiple checks to ensure the eligibility of a user to add a subpool with regards to the keys and keychain/superior keychain to stake.
*/
func CheckKeysToStakeEligiblity(keys []*models.KOSSimplifiedMetadata, keychainId, superiorKeychainId int) error {
	// ensures that there is at least 1 key to stake.
	if len(keys) == 0 || keys == nil {
		return errors.New("must stake at least 1 key")
	}

	// checks if there are 1, 2, 3, 5 or 15 keys to stake. no other amount is allowed.
	if len(keys) != 1 && len(keys) != 2 && len(keys) != 3 && len(keys) != 5 && len(keys) != 15 {
		return errors.New("must stake 1, 2, 3, 5 or 15 keys")
	}

	// if a user stakes 1, 2, 3 or 5 keys, they are only allowed to use EITHER 1 keychain or 1 superior keychain.
	// this means that if keychainId != -1, then superiorKeychainId must be -1 and vice versa.
	// if a user stakes 15 keys, they are ONLY allowed to use a superior keychain.
	// NOTE: each subpool is allowed to only have 1 keychain or 1 superior keychain regardless.
	if len(keys) == 15 {
		if keychainId != -1 {
			return errors.New("cannot use keychain when staking 15 keys. please use a superior keychain instead or open multiple subpools")
		}
		if keychainId == 0 {
			return errors.New("invalid keychain ID")
		}
		if superiorKeychainId == 0 {
			return errors.New("invalid superior keychain ID")
		}
	} else {
		if keychainId != -1 && superiorKeychainId != -1 {
			return errors.New("cannot stake both keychain and superior keychain in one subpool. please use either a keychain or a superior keychain")
		}
		if keychainId == 0 {
			return errors.New("invalid keychain ID")
		}
		if superiorKeychainId == 0 {
			return errors.New("invalid superior keychain ID")
		}
	}

	return nil
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
