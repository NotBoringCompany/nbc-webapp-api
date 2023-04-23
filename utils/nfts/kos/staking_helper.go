package utils_kos

import (
	"context"
	"errors"
	"fmt"
	"math"
	"nbc-backend-api-v2/configs"
	"nbc-backend-api-v2/models"
	"nbc-backend-api-v2/utils"
	"strings"
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
Allows the staker to claim subpool `subpoolId`'s reward from staking pool `stakingPoolId`.
Checks if the `wallet` given matches the staker's wallet.
*/
func ClaimReward(collection *mongo.Collection, wallet string, stakingPoolId, subpoolId int) error {
	if collection.Name() != "RHStakingPool" {
		return errors.New("collection must be RHStakingPool")
	}

	filter := bson.M{"stakingPoolID": stakingPoolId}

	var stakingPool models.StakingPool
	if err := collection.FindOne(context.Background(), filter).Decode(&stakingPool); err != nil {
		return err
	}

	// reward claims only work if the subpool has been moved to `ClosedSubpools` (i.e. when the staking ends).
	// stakers CANNOT claim early.
	var subpool *models.StakingSubpool
	for _, closedSubpool := range stakingPool.ClosedSubpools {
		if closedSubpool.SubpoolID == subpoolId {
			subpool = closedSubpool
			break
		}
	}

	// if `subpool` is nil, then the subpool with ID `subpoolId` does not exist in `ClosedSubpools`.
	if subpool == nil {
		return errors.New("subpool with given ID does not exist in ClosedSubpools")
	}

	// returns the Staker's Object ID for `wallet`. if it matches with the `Staker`'s Object ID in `RHStakingPool`, then the staker is valid.
	stakerId, err := GetStakerInstance(configs.GetCollections(configs.DB, "RHStakerData"), wallet)
	if err != nil {
		return err
	}

	// convert the object IDs to hex strings since they are of type `primitive.ObjectID` struct.
	if stakerId.Hex() != subpool.Staker.Hex() {
		fmt.Println("Staker ID: ", stakerId)
		fmt.Println("Subpool Staker ID: ", subpool.Staker)
		return errors.New("staker for this subpool does not match wallet given")
	}

	// checks if Subppol is banned
	if subpool.Banned {
		return errors.New("subpool is banned from claiming rewards")
	}

	// checks if reward has been claimed.
	if subpool.RewardClaimed {
		return errors.New("reward has already been claimed")
	}

	// otherwise, we assume that any closed subpools that are NOT banned and doesn't have its reward claimed can have the rewards claimed.
	// reason: a subpool can only be `closed` if 1. the staking period has ended, or 2. the subpool has been banned.
	// we can now calculate the reward to be given to the staker.
	// we check if the reward to give is tokens.
	if strings.Contains(stakingPool.Reward.Name, "Token") {
		tokensToGive, err := CalcSubpoolTokenShare(collection, stakingPoolId, subpoolId)
		if err != nil {
			return err
		}
		// we now add the tokens to the staker's wallet.
		err = AddTokensToStaker(configs.GetCollections(configs.DB, "RHStakerData"), stakingPool.Reward.Name, wallet, tokensToGive)
		if err != nil {
			return err
		}

		// we now update the `RewardClaimed` field to true.
		err = UpdateRewardClaimedToTrue(collection, stakingPoolId, subpoolId)
		if err != nil {
			return err
		}

		return nil
	} else {
		// NOT IMPLEMENTED YET!
		// this needs to be updated once non-token rewards are out.
		return errors.New("non-token rewards are not implemented yet")
	}
}

/*
Updates a specific Subpool with ID `subpoolId` in Staking Pool `stakingPoolId`'s `RewardClaimed` field to true.
This function does NOT check if `rewardClaimed` has been set to true already. this must be checked beforehand.
*/
func UpdateRewardClaimedToTrue(collection *mongo.Collection, stakingPoolId, subpoolId int) error {
	filter := bson.M{"stakingPoolID": stakingPoolId, "closedSubpools.subpoolID": subpoolId}

	update := bson.M{"$set": bson.M{"closedSubpools.$.rewardClaimed": true}}

	// update the `RewardClaimed` field to true.
	_, err := collection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		return err
	}

	fmt.Printf("Updated subpool ID %d's (from staking pool ID %d) rewardClaimed field to true", subpoolId, stakingPoolId)
	return nil
}

/*
Shifts ALL `ActiveSubpools` to `ClosedSubpools` of ANY StakingPools when the staking period ends.
*/
func CloseSubpoolsOnStakeEnd(collection *mongo.Collection) error {
	if collection.Name() != "RHStakingPool" {
		return errors.New("collection must be RHStakingPool")
	}

	now := time.Now()

	filter := bson.M{"endTime": bson.M{"$lte": now}}

	// get all matching StakingPools
	cursor, err := collection.Find(context.Background(), filter)
	if err != nil {
		return err
	}
	defer cursor.Close(context.Background())

	// iterate over the cursor and shift all subpools to `ClosedSubpools`.
	for cursor.Next(context.Background()) {
		var stakingPool models.StakingPool
		if err := cursor.Decode(&stakingPool); err != nil {
			return err
		}

		// update all subpools in `ActiveSubpools` and move them over to `ClosedSubpools`.
		for _, subpool := range stakingPool.ActiveSubpools {
			subpool.ExitTime = now
			stakingPool.ClosedSubpools = append(stakingPool.ClosedSubpools, subpool)

			fmt.Printf("moved subpool %v from staking pool %v to closed subpools \n", subpool.SubpoolID, stakingPool.StakingPoolID)
		}

		// clear `ActiveSubpools` and update the document.
		stakingPool.ActiveSubpools = nil

		// update the StakingPool document.
		result, err := collection.ReplaceOne(context.Background(), bson.M{"stakingPoolID": stakingPool.StakingPoolID}, stakingPool)
		if err != nil {
			return err
		}

		fmt.Printf("Updated staking pool %v and shifted all its active subpools to closed subpools", result.UpsertedID)
	}

	if err := cursor.Err(); err != nil {
		return err
	}

	return nil
}

/*
AddTokensToStaker is a helper function that adds `tokensToGive` to the staker's wallet assuming all checks have passed beforehand.
*/
func AddTokensToStaker(collection *mongo.Collection, rewardName, wallet string, tokensToGive float64) error {
	if collection.Name() != "RHStakerData" {
		return errors.New("collection must be RHStakerData")
	}

	// checks if `wallet` exists in RHStakerData. if it exists, return an error.
	exists, err := CheckStakerExists(collection, wallet)
	if err != nil {
		return err
	}

	reward := &models.Reward{Name: rewardName, Amount: tokensToGive}

	fmt.Printf("reward to give to staker: %v \n", reward)

	if !exists {
		fmt.Printf("staker with wallet %v does not exist. creating new staker... \n", wallet)

		// add a new staker with the given wallet and add `tokensToGive` to the staker's wallet.
		stakerObjId, err := AddStaker(collection, wallet)
		if err != nil {
			return err
		}

		// add the tokens to the staker's wallet.
		filter := bson.M{"_id": stakerObjId}

		// since the new staker doesn't have the EarnedRewards field, we need to update it.
		update := bson.M{
			"$set": bson.M{
				"earnedRewards": []*models.Reward{reward},
			},
		}

		_, err = collection.UpdateOne(context.Background(), filter, update)
		if err != nil {
			return err
		}

		return nil
	} else { // if the staker already exists, we first do multiple checks.
		fmt.Printf("staker with wallet %v already exists. updating staker... \n", wallet)

		filter := bson.M{"wallet": wallet}

		// check if the staker already has an existing `earnedRewards` field.
		var staker models.Staker
		err := collection.FindOne(context.Background(), filter).Decode(&staker)
		if err != nil {
			return err
		}

		// if the staker already has an existing `earnedRewards` field, we just append the new reward to the existing `earnedRewards` field.
		// otherwise, we need to update the `earnedRewards` field.
		if staker.EarnedRewards == nil {
			update := bson.M{
				"$set": bson.M{
					"earnedRewards": []*models.Reward{reward},
				},
			}

			_, err = collection.UpdateOne(context.Background(), filter, update)
			if err != nil {
				return err
			}

			return nil
		} else {
			fmt.Printf("staker with wallet %v already has an existing `earnedRewards` field. checking if reward with name %v already exists... \n", wallet, rewardName)

			// if the staker already has an existing `earnedRewards` field, we need to check if the reward with the same name already exists.
			// if it does, we just add the amount to the existing reward.
			// if it doesn't, we just append the new reward to the existing `earnedRewards` field.
			for _, earnedReward := range staker.EarnedRewards {
				if earnedReward.Name == rewardName {
					earnedReward.Amount += tokensToGive

					update := bson.M{
						"$set": bson.M{
							"earnedRewards": staker.EarnedRewards,
						},
					}

					_, err = collection.UpdateOne(context.Background(), filter, update)
					if err != nil {
						return err
					}

					return nil
				} else {
					staker.EarnedRewards = append(staker.EarnedRewards, reward)

					update := bson.M{
						"$set": bson.M{
							"earnedRewards": staker.EarnedRewards,
						},
					}

					_, err = collection.UpdateOne(context.Background(), filter, update)
					if err != nil {
						return err
					}

					return nil
				}
			}
		}

	}

	fmt.Printf("Successfully added %f tokens to %s's wallet.\n", tokensToGive, wallet)
	return nil
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
Gets the subpool points accumulated for a subpool with ID `subpoolId` for a staking pool with ID `stakingPoolId`.
*/
func GetAccSubpoolPoints(collection *mongo.Collection, stakingPoolId, subpoolId int) (float64, error) {
	if collection.Name() != "RHStakingPool" {
		return 0, errors.New("collection must be RHStakingPool")
	}

	// filters through both active and closed subpools and gets the subpool with ID `subpoolId`.
	filter := bson.D{
		{"stakingPoolID", stakingPoolId},
		{"$or", bson.A{
			bson.D{{"activeSubpools.subpoolID", subpoolId}},
			bson.D{{"closedSubpools.subpoolID", subpoolId}},
		}},
	}

	var stakingPool models.StakingPool
	if err := collection.FindOne(context.Background(), filter).Decode(&stakingPool); err != nil {
		return 0, err
	}

	var subpoolPoints float64

	// Find the subpool with ID `subpoolId` and its subpool points.
	for _, subpool := range stakingPool.ActiveSubpools {
		if subpool.SubpoolID == subpoolId {
			subpoolPoints = subpool.SubpoolPoints
			break
		}
	}
	for _, subpool := range stakingPool.ClosedSubpools {
		if subpool.SubpoolID == subpoolId {
			subpoolPoints = subpool.SubpoolPoints
			break
		}
	}

	return subpoolPoints, nil
}

/*
Gets the total subpool points accumulated across ALL subpools (both active and closed) within a staking pool with ID `stakingPoolId`.
*/
func GetTotalSubpoolPoints(collection *mongo.Collection, stakingPoolId int) (float64, error) {
	if collection.Name() != "RHStakingPool" {
		return 0, errors.New("collection must be RHStakingPool")
	}

	// filters through both active and closed subpools.
	filter := bson.D{
		{"stakingPoolID", stakingPoolId},
		{"$or", bson.A{
			bson.M{"activeSubpools": bson.M{"$exists": true, "$ne": nil}},
			bson.M{"closedSubpools": bson.M{"$exists": true, "$ne": nil}},
		}},
	}

	var activeSubpools []*models.StakingSubpool
	var closedSubpools []*models.StakingSubpool

	cursor, err := collection.Find(context.Background(), filter)
	if err != nil {
		return 0, err
	}
	defer cursor.Close(context.Background())

	for cursor.Next(context.Background()) {
		var stakingPool models.StakingPool
		if err := cursor.Decode(&stakingPool); err != nil {
			return 0, err
		}
		activeSubpools = append(activeSubpools, stakingPool.ActiveSubpools...)
		closedSubpools = append(closedSubpools, stakingPool.ClosedSubpools...)
	}
	if err := cursor.Err(); err != nil {
		return 0, err
	}

	var totalSubpoolPoints float64
	for _, subpool := range append(activeSubpools, closedSubpools...) {
		totalSubpoolPoints += subpool.SubpoolPoints
	}

	return totalSubpoolPoints, nil
}

/*
ONLY FOR TOKEN REWARDS: calculate the reward share for a specific subpool of ID `subpoolId` for a staking pool with ID `stakingPoolId`.
*/
func CalcSubpoolTokenShare(collection *mongo.Collection, stakingPoolId, subpoolId int) (float64, error) {
	if collection.Name() != "RHStakingPool" {
		return 0, errors.New("collection must be RHStakingPool")
	}

	// fetch the accumulated subpool points for a specific subpool of ID `subpoolId` for a specific staking pool with ID `stakingPoolId`
	accSubpoolPoints, err := GetAccSubpoolPoints(collection, stakingPoolId, subpoolId)
	if err != nil {
		return 0, err
	}

	// fetch the total subpool points across ALL subpools for a specific staking pool with ID `stakingPoolId`
	totalSubpoolPoints, err := GetTotalSubpoolPoints(collection, stakingPoolId)
	if err != nil {
		return 0, err
	}

	// fetch the total token reward for the staking pool
	totalTokenReward, err := GetTotalTokenReward(collection, stakingPoolId)
	if err != nil {
		return 0, err
	}

	// calculate the reward share for a specific subpool of ID `subpoolId` for a specific staking pool with ID `stakingPoolId`
	rewardShare := math.Round(accSubpoolPoints/totalSubpoolPoints*totalTokenReward*100) / 100

	fmt.Printf("Reward share for subpool %d of staking pool %d: %f\n", subpoolId, stakingPoolId, rewardShare)

	return rewardShare, nil
}

/*
ONLY FOR TOKEN REWARDS: gets the total token reward for a staking pool with ID `stakingPoolId`.
*/
func GetTotalTokenReward(collection *mongo.Collection, stakingPoolId int) (float64, error) {
	if collection.Name() != "RHStakingPool" {
		return 0, errors.New("collection must be RHStakingPool")
	}

	filter := bson.M{"stakingPoolID": stakingPoolId}

	var stakingPool models.StakingPool
	if err := collection.FindOne(context.Background(), filter).Decode(&stakingPool); err != nil {
		return 0, err
	}

	reward := stakingPool.Reward
	if !strings.Contains(reward.Name, "Token") {
		return 0, errors.New("reward must be a token") // reward must be a token, or else there is no total token reward
	}

	return float64(reward.Amount), nil
}

/*
Adds a new staking pool to the RHStakingPool collection.
*/
func AddStakingPool(collection *mongo.Collection, rewardName string, rewardAmount float64) error {
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
		EntryAllowance: time.Now(),
		StartTime:      time.Now().Add(time.Hour * 24 * 1),
		EndTime:        time.Now().Add(time.Hour * 24 * 8), // 7 days after start time
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

	// checks if keychain is already staked in this staking pool (assuming id is not -1 or 0)
	if keychainId != -1 && keychainId != 0 {
		staked, err := CheckIfKeychainStaked(collection, stakingPoolId, keychainId)
		if err != nil {
			return err
		}
		if staked {
			return errors.New("keychain has already been staked in another subpool for this staking pool")
		}
	}

	// checks if superior keychain is already staked in this staking pool (assuming id is not -1 or 0)
	if superiorKeychainId != -1 && superiorKeychainId != 0 {
		staked, err := CheckIfSuperiorKeychainStaked(collection, stakingPoolId, superiorKeychainId)
		if err != nil {
			return err
		}
		if staked {
			return errors.New("superior keychain has already been staked in another subpool for this staking pool")
		}
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
		SubpoolPoints:            math.Round(subpoolPoints*100) / 100, // 2 decimal places
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

	// finds the max subpool ID from the staking pool from both active and closed subpools.
	pipeline := mongo.Pipeline{
		bson.D{{Key: "$match", Value: bson.M{"stakingPoolID": stakingPoolId}}},
		bson.D{{Key: "$project", Value: bson.M{
			"allSubpools": bson.M{"$concatArrays": []interface{}{"$activeSubpools", "$closedSubpools"}},
		}}},
		bson.D{{Key: "$unwind", Value: "$allSubpools"}},
		bson.D{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: "$_id"},
			{Key: "maxSubpoolID", Value: bson.D{{Key: "$max", Value: "$allSubpools.subpoolID"}}},
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
