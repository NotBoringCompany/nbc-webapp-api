package utils_kos

import (
	"context"
	"errors"
	"fmt"
	"log"
	"nbc-backend-api-v2/configs"
	"nbc-backend-api-v2/models"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	UtilsKeychain "nbc-backend-api-v2/utils/nfts/keychain"
	UtilsSuperiorKeychain "nbc-backend-api-v2/utils/nfts/superior_keychain"
)

/*
For ALL staking pools, this function will get the staker from each active subpool and check whether the keys, keychain and/or superior keychain that they staked in each subpool are still owned by them.
if not, the subpool will automatically be removed from `ActiveSubpools` and moved to `ClosedSubpools`, change Banned to true and impose a BannedData instance on the staker.
*/
func VerifyStakerOwnership(collection *mongo.Collection) error {
	if collection.Name() != "RHStakingPool" {
		return errors.New("collection must be RHStakingPool")
	}

	// we get the list of all subpools.
	subpools, err := GetAllActiveSubpools(collection)
	if err != nil {
		return err
	}

	// loop through each subpool, get the staked keys, keychain and/or superior keychain and get the staker.
	// then, check whether the staker still owns the keys, keychain and/or superior keychain.
	for _, subpool := range subpools {
		stakerData, err := GetStakerFromObjID(configs.GetCollections(configs.DB, "RHStakerData"), subpool.Staker)
		if err != nil {
			return err
		}

		// check whether the staker still owns the keys, keychain and/or superior keychain.
		// if not, remove the subpool from `ActiveSubpools` and move it to `ClosedSubpools`, change Banned to true and impose a BannedData instance on the staker.
		// if yes, do nothing.
		if subpool.StakedKeys != nil {
			var stakedKeyIds []int
			// get the token IDs of the staked keys
			for _, key := range subpool.StakedKeys {
				stakedKeyIds = append(stakedKeyIds, key.TokenID)
			}

			// check whether the staker still owns the staked keys
			stillOwned, err := VerifyOwnership(stakerData.Wallet, stakedKeyIds)
			if err != nil {
				return err
			}
			if !stillOwned {
				// first, impose a BannedData instance on the staker.
				err = UpdateStakerBannedData(configs.GetCollections(configs.DB, "RHStakerData"), subpool.Staker)
				if err != nil {
					return err
				}
				// then, ban the subpool.
				err := BanSubpool(collection, subpool.StakingPoolID, subpool.SubpoolID)
				if err != nil {
					return err
				}

				log.Printf("verifying complete. staker does NOT own at least one of the staked keys anymore. ban imposed.")
				return nil
			}
		}

		// check whether the staker still owns the keychain
		// if a keychain is staked, the keychain id is not -1 or 0.
		if subpool.StakedKeychainID != -1 && subpool.StakedKeychainID != 0 {
			stillOwned, err := UtilsKeychain.VerifyOwnership(stakerData.Wallet, []int{subpool.StakedKeychainID})
			if err != nil {
				return err
			}

			// if the staker does not own the keychain anymore, ban the subpool and impose a BannedData instance on the staker.
			if !stillOwned {
				// first, impose a BannedData instance on the staker.
				err = UpdateStakerBannedData(configs.GetCollections(configs.DB, "RHStakerData"), subpool.Staker)
				if err != nil {
					return err
				}
				// then, ban the subpool.
				err := BanSubpool(collection, subpool.StakingPoolID, subpool.SubpoolID)
				if err != nil {
					return err
				}

				log.Printf("verifying complete. staker does NOT own at least one of the staked keys anymore. ban imposed.")
				return nil
			}
		}

		// check whether the staker still owns the superior keychain
		// if a superior keychain is staked, the superior keychain id is not -1 or 0.
		if subpool.StakedSuperiorKeychainID != -1 && subpool.StakedSuperiorKeychainID != 0 {
			stillOwned, err := UtilsSuperiorKeychain.VerifyOwnership(stakerData.Wallet, []int{subpool.StakedSuperiorKeychainID})
			if err != nil {
				return err
			}

			// if the staker does not own the superior keychain anymore, ban the subpool and impose a BannedData instance on the staker.
			if !stillOwned {
				// first, impose a BannedData instance on the staker.
				err = UpdateStakerBannedData(configs.GetCollections(configs.DB, "RHStakerData"), subpool.Staker)
				if err != nil {
					return err
				}
				// then, ban the subpool.
				err := BanSubpool(collection, subpool.StakingPoolID, subpool.SubpoolID)
				if err != nil {
					return err
				}

				log.Printf("verifying complete. staker does NOT own at least one of the staked keys anymore. ban imposed.")
				return nil
			}
		}

		log.Printf("verifying complete. staker still owns all staked items for subpool %d of staking pool %d", subpool.SubpoolID, subpool.StakingPoolID)
	}

	return nil
}

/*
Checks all relevant staking pools that are about to start. if they have less than 20 stakers, they will be cancelled.
*/
func CheckStakingPoolStakerCount(collection *mongo.Collection) error {
	if collection.Name() != "RHStakingPool" {
		return errors.New("collection must be RHStakingPool")
	}

	cursor, err := collection.Find(context.Background(), bson.M{})
	if err != nil {
		return err
	}
	defer cursor.Close(context.Background())

	for cursor.Next(context.Background()) {
		var stakingPool models.StakingPool
		err := cursor.Decode(&stakingPool)
		if err != nil {
			return err
		}

		// check if the staking pool has started
		if time.Now().After(stakingPool.StartTime) && time.Now().Before(stakingPool.EndTime) {
			// collect all unique stakers from the subpools
			uniqueStakers := make(map[primitive.ObjectID]bool)
			for _, subpool := range stakingPool.ActiveSubpools {
				uniqueStakers[*subpool.Staker] = true
			}

			if len(uniqueStakers) < 20 {
				// cancel the staking pool
				_, err := collection.DeleteOne(context.Background(), bson.M{"_id": stakingPool.ID})
				if err != nil {
					return err
				}

				log.Printf("staking pool %d has been cancelled due to insufficient stakers\n", stakingPool.ID)
			}
		}
	}
	if err := cursor.Err(); err != nil {
		return err
	}

	return nil
}

/*
Checks if a staker with the given wallet is banned.
*/
func CheckIfStakerBanned(collection *mongo.Collection, wallet string) (bool, error) {
	if collection.Name() != "RHStakerData" {
		return true, errors.New("collection must be RHStakerData") // defaults to true if an error occurs
	}

	filter := bson.M{"wallet": wallet}
	var staker models.Staker
	err := collection.FindOne(context.Background(), filter).Decode(&staker)

	if err != nil {
		return true, err //defaults to true if an error occurs
	}

	if staker.BannedData == nil {
		return false, nil
	}

	// checks if the CurrentUnbanTime is greater than the current time. if so, then the staker is banned and the fn returns true.
	return time.Now().Before(staker.BannedData.CurrentUnbanTime), nil
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
func CheckKeysToStakeEligibility(keys []*models.KOSSimplifiedMetadata, keychainId, superiorKeychainId int) error {
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
A user is allowed to (at the moment) create a limited amount of subpools per staking pool.
For Flush combos (15 keys), they are allowed to create an unlimited amount of subpools.
For Pentuple (5), they are only allowed to create 5.
For Single, Duo and Trio combos (1, 2, 3), they are only allowed to create 3 each.
*/
func CheckSubpoolComboEligibility(collection *mongo.Collection, stakingPoolId int, stakerWallet string, keys []*models.KOSSimplifiedMetadata) (bool, error) {
	if collection.Name() != "RHStakingPool" {
		return false, errors.New("invalid collection name") // defaults to false if an error occurs
	}

	fmt.Println(collection.Name())
	fmt.Println(stakingPoolId)
	fmt.Println(stakerWallet)
	fmt.Println(keys)

	// fetch the staker's object ID
	stakerObjId, err := GetStakerInstance(configs.GetCollections(configs.DB, "RHStakerData"), stakerWallet)
	if err != nil {
		return false, err
	}
	log.Printf("staker object ID: %v", stakerObjId)
	// if staker object doesn't exist, we create a new staker instance.
	if stakerObjId == nil {
		newStaker := &models.Staker{
			Wallet: stakerWallet,
		}

		addStaker, err := configs.GetCollections(configs.DB, "RHStakerData").InsertOne(context.Background(), newStaker)
		if err != nil {
			return false, err
		}

		log.Printf("staker not found while checking combo. created new staker instance: %v", newStaker)
		stakerObjId = addStaker.InsertedID.(*primitive.ObjectID)
	}

	filter := bson.M{"stakingPoolID": stakingPoolId}
	var stakingPool models.StakingPool
	err = collection.FindOne(context.Background(), filter).Decode(&stakingPool)
	if err != nil {
		return false, err
	}

	var stakersSubpools []*models.StakingSubpool

	// fetch the active subpools only (we don't check for closed subpools here since subpool creations are automatically only allowed during the `EntryAllowance` period)
	// in this case, any closed subpools are treated as if they don't exist at the first place.
	for _, subpool := range stakingPool.ActiveSubpools {
		// find all subpools that the staker has created
		log.Printf("staker objId when checking subpool: %v", stakerObjId.Hex())
		if subpool.Staker.Hex() == stakerObjId.Hex() {
			stakersSubpools = append(stakersSubpools, subpool)
		}
	}

	// the amount of each of these subpools that the staker has created for `stakingPoolId`.
	// note that flush doesn't count here since it is unlimited.
	var singleCombo, duoCombo, trioCombo, pentupleCombo int

	for _, subpool := range stakersSubpools {
		switch len(subpool.StakedKeys) {
		case 1:
			singleCombo++
		case 2:
			duoCombo++
		case 3:
			trioCombo++
		case 5:
			pentupleCombo++
		}
	}

	// get the length of the `keys`
	keysLength := len(keys)

	if keysLength == 15 {
		return true, nil // return true
	} else if keysLength == 5 {
		if pentupleCombo >= 2 {
			return false, nil // return false
		} else {
			return true, nil // return true
		}
	} else if keysLength == 3 {
		if trioCombo >= 2 {
			return false, nil // return false
		} else {
			return true, nil // return true
		}
	} else if keysLength == 2 {
		if duoCombo >= 2 {
			return false, nil // return false
		} else {
			return true, nil // return true
		}
	} else if keysLength == 1 {
		if singleCombo >= 2 {
			return false, nil // return false
		} else {
			return true, nil // return true
		}
	} else {
		return false, errors.New("invalid keys length")
	}
}

/*
Same as `CheckSubpoolComboEligiblityAlt`, but uses `keyCount `instead of `keys`.
Used mainly for API calls.
*/
func CheckSubpoolComboEligibilityAlt(collection *mongo.Collection, stakingPoolId int, stakerWallet string, keyCount int) (bool, error) {
	if collection.Name() != "RHStakingPool" {
		return false, errors.New("invalid collection name") // defaults to false if an error occurs
	}

	// fetch the staker's object ID
	stakerObjId, err := GetStakerInstance(configs.GetCollections(configs.DB, "RHStakerData"), stakerWallet)
	log.Printf("staker object ID: %v", stakerObjId)
	if err != nil {
		return false, err
	}
	log.Printf("staker object ID is nil: %v", stakerObjId == nil)
	// if staker object doesn't exist, we create a new staker instance.
	if stakerObjId == nil {
		log.Printf("staker not found while checking combo. creating new staker instance...")
		newStaker := &models.Staker{
			Wallet: stakerWallet,
		}

		addStaker, err := configs.GetCollections(configs.DB, "RHStakerData").InsertOne(context.Background(), newStaker)
		if err != nil {
			return false, err
		}

		log.Printf("staker not found while checking combo. created new staker instance: %v", newStaker)
		stakerObjId = addStaker.InsertedID.(*primitive.ObjectID)
	}

	filter := bson.M{"stakingPoolID": stakingPoolId}
	var stakingPool models.StakingPool
	log.Printf("staking pool object ID: %v", stakingPool.ID.Hex())
	err = collection.FindOne(context.Background(), filter).Decode(&stakingPool)
	if err != nil {
		return false, err
	}

	var stakersSubpools []*models.StakingSubpool

	// fetch the active subpools only (we don't check for closed subpools here since subpool creations are automatically only allowed during the `EntryAllowance` period)
	// in this case, any closed subpools are treated as if they don't exist at the first place.
	for _, subpool := range stakingPool.ActiveSubpools {
		log.Printf("staker objId when checking subpool: %v", stakerObjId.Hex())
		// find all subpools that the staker has created
		if subpool.Staker.Hex() == stakerObjId.Hex() {
			stakersSubpools = append(stakersSubpools, subpool)
		}
	}

	// the amount of each of these subpools that the staker has created for `stakingPoolId`.
	// note that flush doesn't count here since it is unlimited.
	var singleCombo, duoCombo, trioCombo, pentupleCombo int

	for _, subpool := range stakersSubpools {
		switch len(subpool.StakedKeys) {
		case 1:
			singleCombo++
		case 2:
			duoCombo++
		case 3:
			trioCombo++
		case 5:
			pentupleCombo++
		}
	}

	log.Printf("singleCombo: %v, dualCombo: %v, trioCombo: %v, pentupleCombo: %v", singleCombo, duoCombo, trioCombo, pentupleCombo)

	if keyCount == 15 {
		return true, nil // return true
	} else if keyCount == 5 {
		if pentupleCombo >= 2 {
			return false, nil // return false
		} else {
			return true, nil // return true
		}
	} else if keyCount == 3 {
		if trioCombo >= 2 {
			return false, nil // return false
		} else {
			return true, nil // return true
		}
	} else if keyCount == 2 {
		if duoCombo >= 2 {
			return false, nil // return false
		} else {
			return true, nil // return true
		}
	} else if keyCount == 1 {
		if singleCombo >= 2 {
			return false, nil // return false
		} else {
			return true, nil // return true
		}
	} else {
		return false, errors.New("invalid keys length")
	}
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
