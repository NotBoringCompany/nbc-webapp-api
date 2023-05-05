package utils_kos

import (
	"context"
	"errors"
	"fmt"
	"log"
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
		Wallet: strings.ToLower(wallet),
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

	// checks if reward is claimable.
	if !subpool.RewardClaimable {
		return errors.New("rewards for subpool is not claimable")
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
Updates a specific subpool with ID `subpoolId` in Staking Pool `stakingPoolId`'s `RewardClaimed` field to true.
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

	log.Printf("Updated subpool ID %d's (from staking pool ID %d) rewardClaimed field to true", subpoolId, stakingPoolId)
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
			subpool.RewardClaimable = true
			stakingPool.ClosedSubpools = append(stakingPool.ClosedSubpools, subpool)

			log.Printf("moved subpool %v from staking pool %v to closed subpools \n", subpool.SubpoolID, stakingPool.StakingPoolID)
		}

		// clear `ActiveSubpools` and update the document.
		stakingPool.ActiveSubpools = nil

		// update the StakingPool document.
		result, err := collection.ReplaceOne(context.Background(), bson.M{"stakingPoolID": stakingPool.StakingPoolID}, stakingPool)
		if err != nil {
			return err
		}

		log.Printf("Updated staking pool %v and shifted all its active subpools to closed subpools", result.UpsertedID)
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

	log.Printf("reward to give to staker: %v \n", reward)

	if !exists {
		log.Printf("staker with wallet %v does not exist. creating new staker... \n", wallet)

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
		log.Printf("staker with wallet %v already exists. updating staker... \n", wallet)

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
			log.Printf("staker with wallet %v already has an existing `earnedRewards` field. checking if reward with name %v already exists... \n", wallet, rewardName)

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

	log.Printf("Successfully added %f tokens to %s's wallet.\n", tokensToGive, wallet)
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

	filter := bson.M{"wallet": strings.ToLower(wallet)}

	var staker models.Staker
	err := collection.FindOne(context.Background(), filter).Decode(&staker)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil // return nil if no documents are found
		}
		return nil, err
	}

	return &staker.ID, nil
}

/*
If a subpool gets banned, then the staker will be imposed a ban penalty.
In this case, it checks if the staker already has a BannedData instance. if not, it creates one.
*/
func UpdateStakerBannedData(collection *mongo.Collection, stakerId *primitive.ObjectID) error {
	if collection.Name() != "RHStakerData" {
		return errors.New("collection must be RHStakerData")
	}

	// check if the staker already has a BannedData instance
	filter := bson.M{"_id": stakerId}
	var staker models.Staker

	err := collection.FindOne(context.Background(), filter).Decode(&staker)
	if err != nil {
		return err
	}

	// if staker already has a BannedData instance, we first update the LastBanTime to now.
	if staker.BannedData != nil {
		// update the LastBanTime to now
		staker.BannedData.LastBanTime = time.Now()

		// update current unban time based on bannedcount
		switch staker.BannedData.BannedCount {
		case 1:
			staker.BannedData.CurrentUnbanTime = time.Now().AddDate(0, 0, 7) // 7 day ban if banned once prior to this
		case 2, 3:
			staker.BannedData.CurrentUnbanTime = time.Now().AddDate(0, 0, 14) // 14 day ban if banned 2 or 3 times
		default:
			staker.BannedData.CurrentUnbanTime = time.Now().AddDate(0, 0, 30) // 30 day ban if banned 4 or more times
		}

		// update banned count by 1
		staker.BannedData.BannedCount++

		// update the staker in the database
		_, err := collection.UpdateOne(context.Background(), bson.M{"_id": stakerId}, bson.M{"$set": bson.M{"bannedData": staker.BannedData}})
		if err != nil {
			return err
		}

		log.Printf("banned staker %s. they have been banned %d times thus far", stakerId.Hex(), staker.BannedData.BannedCount)
	} else { // if staker does not have a BannedData instance, we create one and set the LastBanTime to now.
		// create a new BannedData instance
		bannedData := &models.BannedData{
			BannedCount:      1,
			LastBanTime:      time.Now(),
			CurrentUnbanTime: time.Now().AddDate(0, 0, 7), // 7 day ban
		}

		// update the staker in the database
		_, err = collection.UpdateOne(context.Background(), bson.M{"_id": stakerId}, bson.M{"$set": bson.M{"bannedData": bannedData}})
		if err != nil {
			return fmt.Errorf("failed to create staker banned data: %s", err)
		}

		log.Printf("created ban instance and banned staker %s. they have been banned %d times thus far", stakerId.Hex(), bannedData.BannedCount)

		return nil
	}

	return nil
}

/*
Allows a staker to unstake from a subpool with `subpoolId` from a staking pool with `stakingPoolId`.
This assumes that the subpool is part of the `ActiveSubpools`. Closed subpools cannot be `unstaked` from.

Unstaking only is allowed if the time now has NOT passed the `startTime` of the staking pool yet.
*/
func UnstakeFromSubpool(collection *mongo.Collection, stakingPoolId, subpoolId int) error {
	if collection.Name() != "RHStakingPool" {
		return errors.New("collection must be RHStakingPool")
	}

	// check if now is past the startTime of the staking pool
	startTime, err := GetStartTimeOfStakingPool(collection, stakingPoolId)
	if err != nil {
		return err
	}
	if time.Now().After(startTime) {
		return errors.New("cannot unstake from a subpool after the staking pool has started")
	}

	filter := bson.M{"stakingPoolID": stakingPoolId}
	update := bson.M{
		"$pull": bson.M{
			"activeSubpools": bson.M{"subpoolID": subpoolId},
		},
	}

	result, err := collection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		return err
	}

	if result.ModifiedCount == 0 {
		return fmt.Errorf("no subpool with ID %d exists in the ActiveSubpools of staking pool %d", subpoolId, stakingPoolId)
	}

	log.Printf("unstaked subpool %d from staking pool %d", subpoolId, stakingPoolId)
	return nil
}

/*
Allows a staker to unstake from all subpools from a staking pool with `stakingPoolId`.
This assumes that the subpools are part of the `ActiveSubpools`. Closed subpools cannot be `unstaked` from.

Unstaking only is allowed if the time now has NOT passed the `startTime` of the staking pool yet.
*/
func UnstakeFromStakingPool(collection *mongo.Collection, stakingPoolId int, stakerWallet string) error {
	if collection.Name() != "RHStakingPool" {
		return errors.New("collection must be RHStakingPool")
	}

	// check if now is past the startTime of the staking pool
	startTime, err := GetStartTimeOfStakingPool(collection, stakingPoolId)
	if err != nil {
		return err
	}
	if time.Now().After(startTime) {
		return errors.New("cannot unstake from a subpool after the staking pool has started")
	}

	// get the object ID based on the wallet
	stakerObjId, err := GetStakerInstance(configs.GetCollections(configs.DB, "RHStakerData"), stakerWallet)
	if err != nil {
		return err
	}

	filter := bson.M{"stakingPoolID": stakingPoolId}
	update := bson.M{
		"$pull": bson.M{
			"activeSubpools": bson.M{"staker": stakerObjId},
		},
	}

	result, err := collection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		return err
	}

	if result.ModifiedCount == 0 {
		return fmt.Errorf("no subpool exists for staker with wallet %s in staking pool %d", stakerWallet, stakingPoolId)
	}

	log.Printf("unstaked all subpools for staker %s from staking pool %d", stakerWallet, stakingPoolId)
	return nil
}

/*
Gets the staking pool data for a staking pool with `stakingPoolId`.
*/
func GetStakingPoolData(collection *mongo.Collection, stakingPoolId int) (*models.StakingPool, error) {
	if collection.Name() != "RHStakingPool" {
		return nil, errors.New("collection must be RHStakingPool")
	}

	var stakingPool models.StakingPool
	filter := bson.M{"stakingPoolID": stakingPoolId}
	err := collection.FindOne(context.Background(), filter).Decode(&stakingPool)
	if err != nil {
		return nil, err
	}

	return &stakingPool, nil
}

/*
Gets the subpool data for a subpool with `subpoolId` from a staking pool with `stakingPoolId`.
*/
func GetSubpoolData(collection *mongo.Collection, stakingPoolId, subpoolId int) (*models.StakingSubpool, error) {
	if collection.Name() != "RHStakingPool" {
		return nil, errors.New("collection must be RHStakingPool")
	}

	var stakingPool models.StakingPool
	filter := bson.M{"stakingPoolID": stakingPoolId}
	err := collection.FindOne(context.Background(), filter).Decode(&stakingPool)
	if err != nil {
		return nil, err
	}

	for _, sp := range stakingPool.ActiveSubpools {
		if sp.SubpoolID == subpoolId {
			return sp, nil
		}
	}
	for _, sp := range stakingPool.ClosedSubpools {
		if sp.SubpoolID == subpoolId {
			return sp, nil
		}
	}

	return nil, fmt.Errorf("no subpool with ID %d exists in staking pool %d", subpoolId, stakingPoolId)
}

/*
`GetSubpoolData` but API-friendly.
*/
func GetSubpoolDataAPI(collection *mongo.Collection, stakingPoolId, subpoolId int) (*models.StakingSubpoolAlt, error) {
	subpoolData, err := GetSubpoolData(collection, stakingPoolId, subpoolId)
	if err != nil {
		return nil, err
	}

	var nftData []*models.NFTData
	for _, nft := range subpoolData.StakedKeys {
		metadata := map[string]interface{}{
			"tokenID":        nft.TokenID,
			"animationUrl":   nft.AnimationUrl,
			"houseTrait":     nft.HouseTrait,
			"typeTrait":      nft.TypeTrait,
			"luckTrait":      nft.LuckTrait,
			"luckBoostTrait": nft.LuckBoostTrait,
		}

		modified := &models.NFTData{
			Name:     fmt.Sprint("Key Of Salvation #", nft.TokenID),
			ImageUrl: nft.AnimationUrl,
			TokenID:  nft.TokenID,
			Metadata: metadata,
		}

		nftData = append(nftData, modified)
	}

	var keychainData []*models.NFTData
	for _, keychainId := range subpoolData.StakedKeychainIDs {
		data := &models.NFTData{
			Name:     fmt.Sprint("Keychain #", keychainId),
			ImageUrl: "https://realmhunter-kos.fra1.digitaloceanspaces.com/keychains/keychain.mp4",
			TokenID:  keychainId,
		}

		keychainData = append(keychainData, data)
	}

	superiorKeychainData := &models.NFTData{
		Name:     fmt.Sprint("Superior Keychain #", subpoolData.StakedSuperiorKeychainID),
		ImageUrl: "https://realmhunter-kos.fra1.digitaloceanspaces.com/keychains/superiorKeychain.mp4",
		TokenID:  subpoolData.StakedSuperiorKeychainID,
	}

	return &models.StakingSubpoolAlt{
		SubpoolID:              subpoolData.SubpoolID,
		Staker:                 subpoolData.Staker,
		EnterTime:              subpoolData.EnterTime.Unix(),
		ExitTime:               subpoolData.ExitTime.Unix(),
		StakedKeys:             nftData,
		StakedKeychains:        keychainData,
		StakedSuperiorKeychain: superiorKeychainData,
		SubpoolPoints:          subpoolData.SubpoolPoints,
		RewardClaimable:        subpoolData.RewardClaimable,
		RewardClaimed:          subpoolData.RewardClaimed,
		Banned:                 subpoolData.Banned,
	}, nil
}

/*
Gets the start time of a staking pool.
*/
func GetStartTimeOfStakingPool(collection *mongo.Collection, stakingPoolId int) (time.Time, error) {
	if collection.Name() != "RHStakingPool" {
		return time.Time{}, errors.New("collection must be RHStakingPool")
	}

	var stakingPool models.StakingPool
	filter := bson.M{"stakingPoolID": stakingPoolId}
	err := collection.FindOne(context.Background(), filter).Decode(&stakingPool)
	if err != nil {
		return time.Time{}, err
	}

	return stakingPool.StartTime, nil
}

/*
Gets all currently stakeable staking pools (where entry allowance has already started, but start time has not yet passed)
*/
func GetAllStakeableStakingPools(collection *mongo.Collection) ([]*models.StakingPool, error) {
	if collection.Name() != "RHStakingPool" {
		return nil, errors.New("collection must be RHStakingPool")
	}
	currentTime := time.Now()
	// filters through all staking pools that allow entry now and have not ended yet
	filter := bson.M{
		"$and": []bson.M{
			{"entryAllowance": bson.M{"$lte": currentTime}},
			{"startTime": bson.M{"$gt": currentTime}},
			{"endTime": bson.M{"$gt": currentTime}},
		},
	}

	cursor, err := collection.Find(context.Background(), filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	var stakingPools []*models.StakingPool
	if err = cursor.All(context.Background(), &stakingPools); err != nil {
		return nil, err
	}

	return stakingPools, nil
}

/*
Gets all ongoing staking pools. This means that entry and start time has already passed, but end time has not yet passed.
*/
func GetAllOngoingStakingPools(collection *mongo.Collection) ([]*models.StakingPool, error) {
	if collection.Name() != "RHStakingPool" {
		return nil, errors.New("collection must be RHStakingPool")
	}
	currentTime := time.Now()
	// filters through all staking pools that have passed entry and start time but have not ended yet
	filter := bson.M{
		"$and": []bson.M{
			{"entryAllowance": bson.M{"$lte": currentTime}},
			{"startTime": bson.M{"$lte": currentTime}},
			{"endTime": bson.M{"$gt": currentTime}},
		},
	}

	cursor, err := collection.Find(context.Background(), filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	var stakingPools []*models.StakingPool
	if err = cursor.All(context.Background(), &stakingPools); err != nil {
		return nil, err
	}

	return stakingPools, nil
}

/*
Gets all currently closed staking pools. this is for all staking pools whose end time has already passed.
*/
func GetAllClosedStakingPools(collection *mongo.Collection) ([]*models.StakingPool, error) {
	if collection.Name() != "RHStakingPool" {
		return nil, errors.New("collection must be RHStakingPool")
	}
	currentTime := time.Now()

	filter := bson.M{
		"endTime": bson.M{
			"$lte": currentTime,
		},
	}

	cursor, err := collection.Find(context.Background(), filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	var stakingPools []*models.StakingPool
	if err = cursor.All(context.Background(), &stakingPools); err != nil {
		return nil, err
	}

	return stakingPools, nil
}

/*
Bans a subpool from being able to claim rewards, removes it from `ActiveSubpools` and moves it to `ClosedSubpools`.
if the subpool is already in `ClosedSubpools` (i.e. not in ActiveSubpools), this function will return an error.
*/
func BanSubpool(collection *mongo.Collection, stakingPoolId, subpoolId int) error {
	if collection.Name() != "RHStakingPool" {
		return errors.New("collection must be RHStakingPool")
	}

	filter := bson.M{"stakingPoolID": stakingPoolId}

	var stakingPool models.StakingPool
	err := collection.FindOne(context.Background(), filter).Decode(&stakingPool)
	if err != nil {
		return err
	}

	for i, subpool := range stakingPool.ActiveSubpools {
		if subpool.SubpoolID == subpoolId {
			// change the subpool's Banned to true
			subpool.Banned = true

			// remove the subpool from the `ActiveSubpools` slice
			stakingPool.ActiveSubpools = append(stakingPool.ActiveSubpools[:i], stakingPool.ActiveSubpools[i+1:]...)

			// add the subpool to the `ClosedSubpools` slice
			stakingPool.ClosedSubpools = append(stakingPool.ClosedSubpools, subpool)

			// update the stakingpool in the database
			_, err := collection.ReplaceOne(context.Background(), bson.M{"_id": stakingPool.ID}, stakingPool)
			if err != nil {
				return err
			}

			log.Printf("subpool %d has been banned from staking pool %d", subpoolId, stakingPoolId)
			return nil
		}
	}

	return errors.New("subpool not found in active subpools")
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
Gets all staking pools from `RHStakingPool` and returns them as a slice of `StakingPool` instances.
*/
func GetAllStakingPools(collection *mongo.Collection) ([]*models.StakingPool, error) {
	if collection.Name() != "RHStakingPool" {
		return nil, errors.New("collection must be RHStakingPool") // defaults to false if an error occurs
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
Removes expired unclaimable subpools (48 hours after the staking pool of the subpool's end time).
*/
func RemoveExpiredUnclaimableSubpools(collection *mongo.Collection) error {
	if collection.Name() != "RHStakingPool" {
		return errors.New("collection must be RHStakingPool")
	}

	currentTime := time.Now()
	// find all staking pools that are 2 days or older (for the end time)
	filter := bson.M{"endTime": bson.M{"$lte": currentTime.Add(-2 * 24 * time.Hour)}}
	cursor, err := collection.Find(context.Background(), filter)
	if err != nil {
		return err
	}
	defer cursor.Close(context.Background())

	for cursor.Next(context.Background()) {
		var stakingPool models.StakingPool
		if err := cursor.Decode(&stakingPool); err != nil {
			return err
		}

		// loop through all closedSubpools
		for _, subpool := range stakingPool.ClosedSubpools {
			// change rewardclaimable to false for each subpool
			if subpool.RewardClaimable {
				_, err := collection.UpdateOne(
					context.Background(),
					bson.M{"_id": stakingPool.ID, "closedSubpools.subpoolID": subpool.SubpoolID},
					bson.M{"$set": bson.M{"closedSubpools.$.rewardClaimable": false}},
				)
				if err != nil {
					return err
				}

				log.Printf("Removed claimable for subpool %d in staking pool %d\n", subpool.SubpoolID, stakingPool.StakingPoolID)
			}
		}
	}
	if err := cursor.Err(); err != nil {
		return err
	}

	return nil
}

/*
Updates the `TotalYieldPoints` field across all staking pools.
*/
func UpdateTotalYieldPoints(collection *mongo.Collection) error {
	// gets all staking pools
	if collection.Name() != "RHStakingPool" {
		return errors.New("collection must be RHStakingPool")
	}

	var stakingPools []*models.StakingPool
	cursor, err := collection.Find(context.Background(), bson.D{})
	if err != nil {
		return err
	}
	defer cursor.Close(context.Background())
	for cursor.Next(context.Background()) {
		var stakingPool models.StakingPool
		if err := cursor.Decode(&stakingPool); err != nil {
			return err
		}
		stakingPools = append(stakingPools, &stakingPool)
	}
	if err := cursor.Err(); err != nil {
		return err
	}

	for _, stakingPool := range stakingPools {
		var totalYieldPoints float64
		for _, subpool := range stakingPool.ActiveSubpools {
			totalYieldPoints += subpool.SubpoolPoints
		}
		for _, subpool := range stakingPool.ClosedSubpools {
			totalYieldPoints += subpool.SubpoolPoints
		}

		// update the total yield points for this staking pool
		_, err := collection.UpdateOne(context.Background(), bson.D{{"stakingPoolID", stakingPool.StakingPoolID}}, bson.D{{"$set", bson.D{{"totalYieldPoints", math.Round(totalYieldPoints*100) / 100}}}})
		if err != nil {
			return err
		}
	}

	log.Printf("Updated total yield points for %d staking pools\n", len(stakingPools))
	return nil
}

/*
Gets all active subpools from each staking pool in `RHStakingPool` and returns them as a slice of `StakingSubpool` instances.
*/
func GetAllActiveSubpools(collection *mongo.Collection) ([]*models.StakingSubpoolWithID, error) {
	if collection.Name() != "RHStakingPool" {
		return nil, errors.New("collection must be RHStakingPool")
	}

	// retrieve all staking pools
	cursor, err := collection.Find(context.Background(), bson.D{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	// get all active subpools from each staking pool.
	// we don't mind if a staking pool has no active subpools or that this `activeSubpools` has multiple subpools with the same ID (as its from different staking pools)
	var activeSubpools []*models.StakingSubpoolWithID
	for cursor.Next(context.Background()) {
		var stakingPool models.StakingPool
		if err := cursor.Decode(&stakingPool); err != nil {
			return nil, err
		}

		for _, subpool := range stakingPool.ActiveSubpools {
			activeSubpools = append(activeSubpools, &models.StakingSubpoolWithID{
				StakingPoolID:  stakingPool.StakingPoolID,
				StakingSubpool: subpool,
			})
		}
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return activeSubpools, nil
}

/*
Gets all subpools from a staker with wallet `stakerWallet`.
*/
func GetStakerSubpools(collection *mongo.Collection, stakerWallet string) ([]*models.StakingSubpoolWithID, error) {
	if collection.Name() != "RHStakingPool" {
		return nil, errors.New("collection must be RHStakingPool")
	}
	// get the staker obj ID from the wallet address
	stakerObjId, err := GetStakerInstance(configs.GetCollections(configs.DB, "RHStakerData"), stakerWallet)
	if err != nil {
		return nil, err
	}

	// retrieve all staking pools
	cursor, err := collection.Find(context.Background(), bson.D{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	// get all active and closed subpools from each staking pool.
	var stakerSubpools []*models.StakingSubpoolWithID
	for cursor.Next(context.Background()) {
		var stakingPool models.StakingPool
		if err := cursor.Decode(&stakingPool); err != nil {
			return nil, err
		}

		for _, subpool := range stakingPool.ActiveSubpools {
			if subpool.Staker.Hex() == stakerObjId.Hex() {
				stakerSubpools = append(stakerSubpools, &models.StakingSubpoolWithID{
					StakingPoolID:  stakingPool.StakingPoolID,
					StakingSubpool: subpool,
				})
			}
		}

		for _, subpool := range stakingPool.ClosedSubpools {
			if subpool.Staker.Hex() == stakerObjId.Hex() {
				stakerSubpools = append(stakerSubpools, &models.StakingSubpoolWithID{
					StakingPoolID:  stakingPool.StakingPoolID,
					StakingSubpool: subpool,
				})
			}
		}
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return stakerSubpools, nil
}

/*
Gets Staker Data from `RHStakerData` collection using the staker's object ID.
*/
func GetStakerFromObjID(collection *mongo.Collection, stakerObjId *primitive.ObjectID) (*models.Staker, error) {
	if collection.Name() != "RHStakerData" {
		return nil, errors.New("Collection must be RHStakerData")
	}

	var staker models.Staker
	err := collection.FindOne(context.Background(), bson.M{"_id": stakerObjId}).Decode(&staker)
	if err != nil {
		return nil, err
	}

	return &staker, nil
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
Gets all the key IDs that have been staked in a specific staking pool.
*/
func GetAllStakedKeyIDs(collection *mongo.Collection, stakingPoolId int) ([]int, error) {
	if collection.Name() != "RHStakingPool" {
		return nil, errors.New("invalid collection name")
	}

	filter := bson.M{"stakingPoolID": stakingPoolId}
	var stakingPool *models.StakingPool
	if err := collection.FindOne(context.Background(), filter).Decode(&stakingPool); err != nil {
		return nil, err
	}

	var stakedKeyIDs []int

	for _, subpool := range stakingPool.ActiveSubpools {
		for _, stakedKey := range subpool.StakedKeys {
			stakedKeyIDs = append(stakedKeyIDs, stakedKey.TokenID)
		}
	}

	return stakedKeyIDs, nil
}

/*
Gets all the keychain IDs that have been staked in a specific staking pool.
*/
func GetAllStakedKeychainIDs(collection *mongo.Collection, stakingPoolId int) ([]int, error) {
	if collection.Name() != "RHStakingPool" {
		return nil, errors.New("invalid collection name")
	}

	filter := bson.M{"stakingPoolID": stakingPoolId}
	var stakingPool *models.StakingPool
	if err := collection.FindOne(context.Background(), filter).Decode(&stakingPool); err != nil {
		return nil, err
	}

	var stakedKeychainIDs []int

	for _, subpool := range stakingPool.ActiveSubpools {
		for _, stakedKeychainId := range subpool.StakedKeychainIDs {
			stakedKeychainIDs = append(stakedKeychainIDs, stakedKeychainId)
		}
	}

	fmt.Println("staked keychain IDs: ", stakedKeychainIDs)
	return stakedKeychainIDs, nil
}

/*
Gets all the superior keychain IDs that have been staked in a specific staking pool.
*/
func GetAllStakedSuperiorKeychainIDs(collection *mongo.Collection, stakingPoolId int) ([]int, error) {
	if collection.Name() != "RHStakingPool" {
		return nil, errors.New("invalid collection name")
	}

	pipeline := mongo.Pipeline{
		bson.D{{"$match", bson.D{{"stakingPoolID", stakingPoolId}}}},                                 // match the staking pool ID with `1`
		bson.D{{"$unwind", "$activeSubpools"}},                                                       // unwinds the activeSubpools array to get separate document for each `Subpool` in the array
		bson.D{{"$match", bson.D{{"activeSubpools.stakedSuperiorKeychainId", bson.D{{"$ne", -1}}}}}}, // filter out documents where `stakedSuperiorKeychainId` is `-1`
		bson.D{{"$project", bson.D{{"_id", 0}, {"stakedSuperiorKeychainId", "$activeSubpools.stakedSuperiorKeychainId"}}}},
	}

	cursor, err := collection.Aggregate(context.Background(), pipeline)
	if err != nil {
		return nil, err
	}

	defer cursor.Close(context.Background())

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
	`keychainIds` the keychain IDs staked (if applicable, otherwise nil)`
	`superiorKeychainId` the superior keychain ID staked
*/
func AddSubpool(
	collection *mongo.Collection,
	stakingPoolId int,
	stakerWallet string,
	keys []*models.KOSSimplifiedMetadata,
	keychainIds []int,
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

	// check if the staker is banned.
	banned, err := CheckIfStakerBanned(configs.GetCollections(configs.DB, "RHStakerData"), stakerWallet)
	if err != nil {
		return err
	}
	if banned {
		return errors.New("staker is temporarily banned from staking")
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
	err = CheckKeysToStakeEligibility(keys, keychainIds, superiorKeychainId)
	if err != nil {
		return err
	}

	// calls `CheckSubpoolComboEligibility` to check how many times a user has staked X amount of keys
	subpoolComboEligiblity, err := CheckSubpoolComboEligibility(collection, stakingPoolId, stakerWallet, keys)
	if err != nil {
		return err
	}
	if !subpoolComboEligiblity {
		return errors.New("you have already staked this combination of keys more times than allowed for this staking pool")
	}

	// checks if keychains are already staked in this staking pool (assuming id is not -1 or 0)
	if len(keychainIds) > 0 {
		for _, keychainId := range keychainIds {
			if keychainId != 1 && keychainId != 0 {
				staked, err := CheckIfKeychainStaked(collection, stakingPoolId, keychainId)
				if err != nil {
					return err
				}
				if staked {
					return errors.New("keychain has already been staked in another subpool for this staking pool")
				}
			}
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
		log.Printf("staker with address %v does not exist. creating a new staker instance...", stakerWallet)
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
	subpoolPoints := CalculateSubpoolPoints(keys, keychainIds, superiorKeychainId)

	subpool := &models.StakingSubpool{
		SubpoolID:                nextSubpoolId,
		Staker:                   stakerObjId,
		EnterTime:                time.Now(),
		StakedKeys:               keys,
		StakedKeychainIDs:        keychainIds,
		StakedSuperiorKeychainID: superiorKeychainId,
		SubpoolPoints:            math.Round(subpoolPoints*100) / 100, // 2 decimal places
		RewardClaimable:          false,
	}

	updatePool := bson.M{"$push": bson.M{"activeSubpools": subpool}}
	update, err := collection.UpdateOne(context.Background(), filter, updatePool)
	if err != nil {
		return err
	}

	log.Printf("Added Subpool ID %d to Staking Pool ID %d. Updated %d document(s)", nextSubpoolId, stakingPoolId, update.ModifiedCount)

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

	// searches for the highest staking pool id from both active and closed subpools.
	pipeline := mongo.Pipeline{
		bson.D{{Key: "$match", Value: bson.M{"stakingPoolID": stakingPoolId}}},
		bson.D{{Key: "$project", Value: bson.M{
			"activeSubpools": bson.M{"$ifNull": []interface{}{"$activeSubpools", bson.A{}}},
			"closedSubpools": bson.M{"$ifNull": []interface{}{"$closedSubpools", bson.A{}}},
		}}},
		bson.D{{Key: "$project", Value: bson.M{
			"allSubpools": bson.M{"$concatArrays": []interface{}{"$activeSubpools", "$closedSubpools"}},
		}}},
		bson.D{{Key: "$unwind", Value: "$allSubpools"}},
		bson.D{{Key: "$group", Value: bson.M{
			"_id":        "$_id",
			"maxSubpool": bson.M{"$max": "$allSubpools.subpoolID"},
		}}},
	}

	var result struct{ MaxSubpool int } // returns the highest subpool ID from the staking pool here.
	cursor, err := collection.Aggregate(context.Background(), pipeline)
	if err != nil {
		return 0, err
	}
	defer cursor.Close(context.Background())

	if cursor.Next(context.Background()) {
		err = cursor.Decode(&result)
		if err != nil {
			return 0, err
		}
	}

	return result.MaxSubpool + 1, nil
}
