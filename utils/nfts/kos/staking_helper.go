package utils_kos

import (
	"context"
	"errors"
	"fmt"
	"math"
	"nbc-backend-api-v2/configs"
	"nbc-backend-api-v2/models"
	"nbc-backend-api-v2/utils"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

/*
Adds a staker to the RHStakerData collection.
Since it's a new staker, only the wallet is needed.
*/
func AddStaker(collection *mongo.Collection, wallet string) (*primitive.ObjectID, error) {
	if collection.Name() != "RHStakerData" {
		return nil, errors.New("collection must be RHStakerData")
	}

	// checks if `wallet` exists in RHStakerData. if it exists, return an error.
	exists, err := CheckStakerExists(collection, wallet)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("staker with the given wallet already exists")
	}

	// checks whether `wallet` has a valid checksum address
	isValidChecksum := utils.ValidChecksum(wallet)
	if !isValidChecksum {
		return nil, errors.New("invalid checksum address")
	}

	// create a new staker instance and add it to `RHStakerData`
	staker := &models.Staker{
		Wallet: wallet,
	}

	result, err := collection.InsertOne(context.Background(), staker)
	if err != nil {
		return nil, err
	}

	fmt.Println("Added staker with Object ID: ", result.InsertedID)

	var stakerID primitive.ObjectID

	return &stakerID, nil
}

/*
Checks if a Staker instance with `wallet` exists in RHStakerData.
*/
func CheckStakerExists(collection *mongo.Collection, wallet string) (bool, error) {
	filter := bson.M{"wallet": wallet}

	var staker models.Staker
	err := collection.FindOne(context.Background(), filter).Decode(&staker)

	if err == mongo.ErrNoDocuments {
		return false, nil // returns false if staker with `wallet` does not exist
	} else if err != nil {
		return true, err // defaults to true if an error occurs
	} else {
		return true, nil // staker with `wallet` exists already
	}
}

/*
Returns the object ID of a Staker given a `wallet`.
*/
func GetStakerInstance(collection *mongo.Collection, wallet string) (*primitive.ObjectID, error) {
	if collection.Name() != "RHStakerData" {
		return nil, errors.New("collection must be RHStakerData")
	}

	filter := bson.M{"wallet": wallet}

	var staker models.Staker
	err := collection.FindOne(context.Background(), filter).Decode(&staker)

	if err != nil {
		return nil, err
	}

	return &staker.ID, nil
}

/*
Adds a new staking pool to the RHStakingPool collection.
*/
func AddStakingPool(collection *mongo.Collection, rewardName string, rewardAmount int) error {
	// collection must be RHStaking Pool.
	if collection.Name() != "RHStakingPool" {
		return errors.New("collection must be RHStakingPool")
	}

	// get the next staking pool id
	stakingPoolID, err := GetNextStakingPoolID(collection)
	if err != nil {
		return err
	}

	// create a new staking pool
	pool := &models.StakingPool{
		StakingPoolID: stakingPoolID,
		Reward: models.Reward{
			Name:   rewardName,
			Amount: rewardAmount,
		},
		StartTime:          time.Now(),
		StakeTimeAllowance: time.Now().Add(time.Hour * 24 * 1),
		EndTime:            time.Now().Add(time.Hour * 24 * 7),
	}

	// insert the new staking pool into the database
	result, err := collection.InsertOne(context.Background(), pool)
	if err != nil {
		return err
	}

	fmt.Println("Added staking pool with Object ID: ", result.InsertedID)

	return nil
}

/*
Adds a subpool to a staking pool. Called when a user stakes their keys (and keychains/superior keychains if applicable).
Every time a user stakes, it counts as a new subpool. If a user has 10 keys and stakes 5 and 5, then there are 2 subpools, each with 5 keys staked.

	`collection` the collection to add the subpool to (must be RHStakingPool)
	`stakingPoolId` the main staking pool ID (to add the subpool instance into)
	`stakerWallet` the staker's wallet to check against `RHStakerData`
	`keys` the key IDs staked
	`keychain` the keychain ID staked
	`superiorKeychain` the superior keychain ID staked
*/
func AddSubpool(
	collection *mongo.Collection,
	stakingPoolId int,
	stakerWallet string,
	keys []*models.KOSSimplifiedMetadata,
	keychainId int,
	superiorKeychainId int,
) error {
	// collection must be RHStakingPool.
	if collection.Name() != "RHStakingPool" {
		return errors.New("collection must be RHStakingPool")
	}

	// check if time is within stake time allowance.
	timeExceeded, err := CheckPoolTimeAllowanceExceeded(collection, stakingPoolId)
	if err != nil {
		return err
	}
	if timeExceeded {
		return errors.New("time allowance for this staking pool has passed. please wait for the next staking pool to open")
	}

	// check if any of the keys in `keys` are already staked.
	// if even just one of them are, return an error.
	checkKeysStaked, err := CheckIfKeysStaked(collection, stakingPoolId, keys)
	if err != nil {
		return err
	}
	if checkKeysStaked {
		return errors.New("1 or more keys are already staked. please stake a set of keys that are not yet staked")
	}

	// calls `CheckKeysToStakeEligibility` to check for amount of keys to stake, keychain, and superior keychain eligibility.
	// if any of the checks fail, return an error.
	err = CheckKeysToStakeEligiblity(keys, keychainId, superiorKeychainId)
	if err != nil {
		return err
	}

	var stakerObjId *primitive.ObjectID

	// after all checks, check if the staker exists in `RHStakerData`. if not, create a new staker instance.
	exists, err := CheckStakerExists(configs.GetCollections(configs.DB, "RHStakerData"), stakerWallet)
	if err != nil {
		return err
	}
	if !exists {
		fmt.Printf("staker with address %v does not exist. creating a new staker instance...", stakerWallet)
		stakerObjId, err = AddStaker(configs.GetCollections(configs.DB, "RHStakerData"), stakerWallet) // create a new staker instance and get the object ID.
		if err != nil {
			return err
		}
	} else {
		stakerObjId, err = GetStakerInstance(configs.GetCollections(configs.DB, "RHStakerData"), stakerWallet) // get the staker instance and get the object ID.
		if err != nil {
			return err
		}
	}

	// filter for the staking pool
	filter := bson.M{"stakingPoolID": stakingPoolId}

	var stakingPool models.StakingPool

	if err := collection.FindOne(context.Background(), filter).Decode(&stakingPool); err != nil {
		return err
	}

	// get the next subpool ID
	nextSubpoolId, err := GetNextSubpoolID(collection, stakingPoolId)
	if err != nil {
		return err
	}

	// call `CalculateSubpoolPoints`
	subpoolPoints := CalculateSubpoolPoints(keys, keychainId, superiorKeychainId)

	subpool := &models.StakingSubpool{
		SubpoolID:                nextSubpoolId,
		Staker:                   stakerObjId,
		EnterTime:                time.Now(),
		StakedKeys:               keys,
		StakedKeychainID:         keychainId,
		StakedSuperiorKeychainID: superiorKeychainId,
		SubpoolPoints:            subpoolPoints,
	}

	updatePool := bson.M{"$push": bson.M{"activeSubpools": subpool}}
	update, err := collection.UpdateOne(context.Background(), filter, updatePool)
	if err != nil {
		return err
	}

	fmt.Printf("Added Subpool ID %d to Staking Pool ID %d. Updated %d document(s)", nextSubpoolId, stakingPoolId, update.ModifiedCount)

	return nil
}

/*
Gets the next staking pool ID from the RHStakingPool collection.
*/
func GetNextStakingPoolID(collection *mongo.Collection) (int, error) {
	// collection must be RHStakingPool.
	if collection.Name() != "RHStakingPool" {
		return 0, errors.New("collection must be RHStakingPool")
	}

	pipeline := mongo.Pipeline{
		bson.D{{Key: "$sort", Value: bson.D{{Key: "stakingPoolID", Value: -1}}}},
		bson.D{{Key: "$limit", Value: 1}},
		bson.D{{Key: "$project", Value: bson.D{{Key: "_id", Value: 0}, {Key: "stakingPoolID", Value: 1}}}},
	}

	cursor, err := collection.Aggregate(context.Background(), pipeline)
	if err != nil {
		return 0, err
	}

	defer cursor.Close(context.Background())

	var result struct{ StakingPoolID int }
	if cursor.Next(context.Background()) {
		err = cursor.Decode(&result)
		if err != nil {
			return 0, err
		}
	}

	fmt.Println("Highest stakingPoolID: ", result.StakingPoolID)

	return result.StakingPoolID + 1, nil
}

/*
Gets the next subpool ID from a specific staking pool with `stakingPoolId`. Different staking pools will always start with ID 1.
*/
func GetNextSubpoolID(collection *mongo.Collection, stakingPoolId int) (int, error) {
	// collection must be RHStakingPool.
	if collection.Name() != "RHStakingPool" {
		return 0, errors.New("collection must be RHStakingPool")
	}

	pipeline := mongo.Pipeline{
		bson.D{{Key: "$unwind", Value: "$activeSubpools"}},
		bson.D{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: "$_id"},
			{Key: "maxSubpoolID", Value: bson.D{{Key: "$max", Value: "$activeSubpools.subpoolID"}}},
		}}},
	}

	cursor, err := collection.Aggregate(context.Background(), pipeline)
	if err != nil {
		return 0, err
	}

	defer cursor.Close(context.Background())

	var result struct{ MaxSubpoolID int }
	if cursor.Next(context.Background()) {
		err = cursor.Decode(&result)
		if err != nil {
			return 0, err
		}
	}

	fmt.Println("Highest subpoolID: ", result.MaxSubpoolID)

	return result.MaxSubpoolID + 1, nil
}

/*
Calculates the subpool points generated for the user's subpool based on the keys and keychain/superior keychain staked.

	`keys` are the keys staked
	`keychain` is the keychain staked
	`superiorKeychain` is the superior keychain staked
*/
func CalculateSubpoolPoints(keys []*models.KOSSimplifiedMetadata, keychainId, superiorKeychainId int) float64 {
	// for each key, calculate the sum of (luck * luckBoost)
	luckAndLuckBoostSum := 0.0
	for _, key := range keys {
		luckAndLuckBoostSum += (key.LuckTrait * key.LuckBoostTrait)
	}

	// call `CalculateKeyCombo`
	keyCombo := CalculateKeyCombo(keys)

	// call `CalculateKeychainCombo`
	keychainCombo := CalculateKeychainCombo(keychainId, superiorKeychainId)

	// call `BaseSubpoolPoints`
	return BaseSubpoolPoints(luckAndLuckBoostSum, keyCombo, keychainCombo)
}

/*
Base subpool points generated formula for the user's subpool based on the given parameters.

	`luckAndLuckBoostSum` is the sum of the luck and luck boost of all keys in the subpool
	`keyCombo` the key combo bonus
	`keychainBonus` the keychain bonus of the key
*/
func BaseSubpoolPoints(luckAndLuckBoostSum, keyCombo, keychainBonus float64) float64 {
	return (100 + math.Pow(luckAndLuckBoostSum, 0.85) + keyCombo) * keychainBonus
}

/*
Calculates the key combo given a list of keys.

	`keys` the keys to calculate the key combo for
*/
func CalculateKeyCombo(keys []*models.KOSSimplifiedMetadata) float64 {
	// get the houses and types of all keys
	houses := make([]string, len(keys))
	types := make([]string, len(keys))
	for i, key := range keys {
		houses[i] = key.HouseTrait
		types[i] = key.TypeTrait
	}

	// call `BaseKeyCombo` with the key count, houses and types
	return BaseKeyCombo(len(keys), houses, types)
}

/*
Base `keyCombo` bonus for a subpool.
NOTE: Calculations here may NOT be final and may be subject to change.

	`keyCount` the amount of keys to stake
	`houses` the houses of the keys to stake
	`types` the types of the keys to stake
*/
func BaseKeyCombo(keyCount int, houses, types []string) float64 {
	// if there is only one key, then there is no combo bonus
	if keyCount == 1 {
		return 0
	}

	house := houses[0]
	sameHouse := true
	for i := 1; i < len(houses); i++ {
		if houses[i] != house {
			sameHouse = false
			break
		}
	}

	typ := types[0]
	sameType := true
	for i := 1; i < len(types); i++ {
		if types[i] != typ {
			sameType = false
			break
		}
	}

	if keyCount == 2 {
		if sameHouse && sameType {
			return 25
		} else if !sameHouse && sameType {
			return 17
		} else if sameHouse && !sameType {
			return 13
		} else {
			return 10
		}
	} else if keyCount == 3 {
		if sameHouse && sameType {
			return 55
		} else if !sameHouse && sameType {
			return 40
		} else if sameHouse && !sameType {
			return 30
		} else {
			return 25
		}
	} else if keyCount == 5 {
		if sameHouse && sameType {
			return 150
		} else if !sameHouse && sameType {
			return 110
		} else if sameHouse && !sameType {
			return 90
		} else {
			return 75
		}
	} else if keyCount == 15 {
		if sameHouse && sameType {
			return 1000
		} else if !sameHouse && sameType {
			return 600
		} else if sameHouse && !sameType {
			return 400
		} else {
			return 300
		}
	} else {
		return 0 // if the key count is neither 1, 2, 3, 5 nor 15, then it is invalid. however, since this error is already being acknowledged in the main function, we just return 0.
	}
}

/*
Calculates the keychain bonus for a subpool.
*/
func CalculateKeychainCombo(keychainId, superiorKeychainId int) float64 {
	var keychainBonus float64 = 1
	if keychainId != -1 && superiorKeychainId == -1 {
		keychainBonus = 1.1 // if the user stakes a keychain, bonus is 1.1
	} else if keychainId == -1 && superiorKeychainId != -1 {
		keychainBonus = 1.5 // if the user stakes a superior keychain, bonus is 1.5
	}

	return keychainBonus
}
