package utils_kos

import (
	"context"
	"errors"
	"log"
	"math"
	"nbc-backend-api-v2/configs"
	"nbc-backend-api-v2/models"
	"strings"
	"sync"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

/*
Gets the reward details for the staker before adding a subpool to show what they will earn.
Only for token based reward pools.
*/
func GetTokenPreAddSubpoolData(
	collection *mongo.Collection,
	stakingPoolId int,
	keyIds,
	keychainIds []int,
	superiorKeychainId int,
) (*models.DetailedTokenSubpoolPreAddCalc, error) {
	if collection.Name() != "RHStakingPool" {
		return nil, errors.New("collection must be RHStakingPool")
	}

	// fetch the key data of each key concurrently
	keyMetadata, err := FetchSimplifiedMetadataConcurrent(keyIds)
	if err != nil {
		return nil, err
	}

	// create buffered channels for each variable
	luckSumCh := make(chan float64, len(keyMetadata))
	keyComboCh := make(chan float64, 1)
	keychainComboCh := make(chan float64, 1)
	subpoolPointsCh := make(chan float64, 1)
	stakingPoolDataCh := make(chan *models.StakingPool, 1)

	// launch goroutines to calculate each variable concurrently
	var wg sync.WaitGroup
	wg.Add(len(keyMetadata))

	for _, metadata := range keyMetadata {
		go func(md *models.KOSSimplifiedMetadata) {
			defer wg.Done()
			luckSumCh <- (md.LuckTrait * md.LuckBoostTrait)
		}(metadata)
	}

	go func() {
		keyCombo := CalculateKeyCombo(keyMetadata)
		keyComboCh <- keyCombo
	}()

	go func() {
		keychainCombo := CalculateKeychainCombo(keychainIds, superiorKeychainId)
		keychainComboCh <- keychainCombo
	}()

	go func() {
		subpoolPoints := CalculateSubpoolPoints(keyMetadata, keychainIds, superiorKeychainId)
		subpoolPointsCh <- subpoolPoints
	}()

	go func() {
		subpoolData, err := GetStakingPoolData(collection, stakingPoolId)
		if err != nil {
			log.Println(err)
			return
		}
		stakingPoolDataCh <- subpoolData
	}()

	// wait for all goroutines to complete
	wg.Wait()

	// collect results from channels
	luckSum := 0.0
	for i := 0; i < len(keyMetadata); i++ {
		luckSum += <-luckSumCh
	}

	keyCombo := <-keyComboCh
	keychainCombo := <-keychainComboCh
	subpoolPoints := <-subpoolPointsCh
	stakingPoolData := <-stakingPoolDataCh

	// calculate the new points
	accSubpoolPoints := stakingPoolData.TotalYieldPoints

	// add the `subpoolPoints` and `accSubpoolPoints`
	newPoints := math.Round((subpoolPoints+accSubpoolPoints)*100) / 100

	// calculate the token share manually
	rewardAmt := stakingPoolData.Reward.Amount
	rewardName := stakingPoolData.Reward.Name
	tokenShare := math.Round(subpoolPoints/newPoints*stakingPoolData.Reward.Amount*100) / 100

	return &models.DetailedTokenSubpoolPreAddCalc{
		TokenShare:         tokenShare,
		PoolTotalReward:    rewardAmt,
		PoolRewardName:     rewardName,
		NewTotalPoolPoints: newPoints,
		DetailedSubpoolPoints: &models.DetailedSubpoolPoints{
			LuckAndLuckBoostSum: luckSum,
			KeyCombo:            keyCombo,
			KeychainCombo:       keychainCombo,
			ComboSum:            subpoolPoints,
		},
	}, nil
}

/*
Gets the detailed calculation for how the subpool's points were calculated.
*/
func BacktrackSubpoolPoints(collection *mongo.Collection, stakingPoolId, subpoolId int) (*struct {
	LuckAndLuckBoostSum float64 `json:"luckAndLuckBoostSum"`
	AngelMultiplier     float64 `json:"angelMultiplier"`
	KeyCombo            float64 `json:"keyCombo"`
	KeychainCombo       float64 `json:"keychainCombo"`
	TotalSubpoolPoints  float64 `json:"totalSubpoolPoints"`
}, error) {
	if collection.Name() != "RHStakingPool" {
		return nil, errors.New("collection must be RHStakingPool")
	}

	// get the subpool data
	subpoolData, err := GetSubpoolData(collection, stakingPoolId, subpoolId)
	if err != nil {
		return nil, err
	}

	// get the luck and luck boost sum
	luckAndLuckBoostSum := 0.0
	for _, key := range subpoolData.StakedKeys {
		luckAndLuckBoostSum += (key.LuckTrait * key.LuckBoostTrait)
	}

	// get the angel multiplier
	angelMultiplier := CalculateAngelMultiplier(subpoolData.StakedKeys)

	// get the keycombo
	keyCombo := CalculateKeyCombo(subpoolData.StakedKeys)
	// get the keychain combo
	keychainCombo := CalculateKeychainCombo(subpoolData.StakedKeychainIDs, subpoolData.StakedSuperiorKeychainID)
	// get the total subpool points
	subpoolPoints := CalculateSubpoolPoints(subpoolData.StakedKeys, subpoolData.StakedKeychainIDs, subpoolData.StakedSuperiorKeychainID)

	// check if subpool points matches the one from `subpoolData`
	if subpoolPoints != subpoolData.SubpoolPoints {
		return nil, errors.New("subpool points do not match")
	}

	data := &struct {
		LuckAndLuckBoostSum float64 `json:"luckAndLuckBoostSum"`
		AngelMultiplier     float64 `json:"angelMultiplier"`
		KeyCombo            float64 `json:"keyCombo"`
		KeychainCombo       float64 `json:"keychainCombo"`
		TotalSubpoolPoints  float64 `json:"totalSubpoolPoints"`
	}{
		LuckAndLuckBoostSum: math.Round(luckAndLuckBoostSum*100) / 100,
		AngelMultiplier:     angelMultiplier,
		KeyCombo:            keyCombo,
		KeychainCombo:       keychainCombo,
		TotalSubpoolPoints:  subpoolPoints,
	}

	return data, nil
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

	log.Printf("Reward share for subpool %d of staking pool %d: %f\n", subpoolId, stakingPoolId, rewardShare)

	return rewardShare, nil
}

/*
ONLY FOR TOKEN REWARDS: Calculates the total token share of a staker for a specific staking pool with ID `stakingPoolId`.
*/
func CalcTotalTokenShare(collection *mongo.Collection, stakingPoolId int, stakerWallet string) (float64, error) {
	if collection.Name() != "RHStakingPool" {
		return 0, errors.New("collection must be RHStakingPool")
	}

	filter := bson.M{"stakingPoolID": stakingPoolId}
	var stakingPool *models.StakingPool
	if err := collection.FindOne(context.Background(), filter).Decode(&stakingPool); err != nil {
		return 0, err
	}

	// get the staker's object ID
	stakerObjectId, err := GetStakerInstance(configs.GetCollections(configs.DB, "RHStakerData"), stakerWallet)
	if err != nil {
		return 0, err
	}

	// find all active subpools belonging to the staker
	var subpools []*models.StakingSubpool
	for _, subpool := range stakingPool.ActiveSubpools {
		if subpool.Staker != nil && subpool.Staker.Hex() == stakerObjectId.Hex() {
			subpools = append(subpools, subpool)
		}
	}

	for _, subpool := range stakingPool.ClosedSubpools {
		if subpool.Staker != nil && subpool.Staker.Hex() == stakerObjectId.Hex() {
			subpools = append(subpools, subpool)
		}
	}

	// since closed subpools don't count into the staker's total reward share, we only get from the active subpools.
	var totalTokenShare float64
	for _, subpool := range subpools {
		rewardShare, err := CalcSubpoolTokenShare(collection, stakingPoolId, subpool.SubpoolID)
		if err != nil {
			return 0, err
		}
		totalTokenShare += rewardShare
	}

	return math.Round(totalTokenShare*100) / 100, nil
}

/*
Calculates the subpool points generated for the user's subpool based on the keys and keychain/superior keychain staked.

	`keys` are the keys staked
	`keychain` is the keychain staked
	`superiorKeychain` is the superior keychain staked
*/
func CalculateSubpoolPoints(keys []*models.KOSSimplifiedMetadata, keychainIds []int, superiorKeychainId int) float64 {
	// for each key, calculate the sum of (luck * luckBoost)
	luckAndLuckBoostSum := 0.0
	for _, key := range keys {
		luckAndLuckBoostSum += (key.LuckTrait * key.LuckBoostTrait)
	}

	// call `CalculateKeyCombo`
	keyCombo := CalculateKeyCombo(keys)

	// call `CalculateAngelMultiplier`
	angelMultiplier := CalculateAngelMultiplier(keys)

	// call `CalculateKeychainCombo`
	keychainCombo := CalculateKeychainCombo(keychainIds, superiorKeychainId)

	// call `BaseSubpoolPoints`
	return BaseSubpoolPoints(luckAndLuckBoostSum, angelMultiplier, keyCombo, keychainCombo)
}

/*
Calculates the total subpool points that a staker has accumulated for a specific staking pool.

	`stakingPoolId` is the ID of the staking pool
	`stakerWallet` is the wallet of the staker
*/
func CalculateStakerTotalSubpoolPoints(collection *mongo.Collection, stakingPoolId int, stakerWallet string) (float64, error) {
	if collection.Name() != "RHStakingPool" {
		return 0, errors.New("collection must be RHStakingPool")
	}
	filter := bson.M{"stakingPoolID": stakingPoolId}
	var stakingPool *models.StakingPool
	if err := collection.FindOne(context.Background(), filter).Decode(&stakingPool); err != nil {
		return 0, err
	}

	// get the staker's object ID
	stakerObjectId, err := GetStakerInstance(configs.GetCollections(configs.DB, "RHStakerData"), stakerWallet)
	if err != nil {
		return 0, err
	}

	// find all subpools belonging to the staker
	var subpools []*models.StakingSubpool
	for _, subpool := range stakingPool.ActiveSubpools {
		if subpool.Staker != nil && subpool.Staker.Hex() == stakerObjectId.Hex() {
			subpools = append(subpools, subpool)
		}
	}

	for _, subpool := range stakingPool.ClosedSubpools {
		if subpool.Staker != nil && subpool.Staker.Hex() == stakerObjectId.Hex() {
			subpools = append(subpools, subpool)
		}
	}

	// add up the total subpool points
	totalSubpoolPoints := 0.0
	for _, subpool := range subpools {
		totalSubpoolPoints += subpool.SubpoolPoints
	}

	return math.Round(totalSubpoolPoints*100) / 100, nil
}

/*
Base subpool points generated formula for the user's subpool based on the given parameters.

	`luckAndLuckBoostSum` is the sum of the luck and luck boost of all keys in the subpool
	`keyCombo` the key combo bonus
	`keychainBonus` the keychain bonus of the key
*/
func BaseSubpoolPoints(luckAndLuckBoostSum, angelMultiplier, keyCombo, keychainBonus float64) float64 {
	return math.Round((100+math.Pow(luckAndLuckBoostSum, angelMultiplier)+keyCombo)*keychainBonus*100) / 100
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
Gets the amount of angel keys present in `keys` and calculates the multiplier.
*/
func CalculateAngelMultiplier(keys []*models.KOSSimplifiedMetadata) float64 {
	angelCount := 0
	for _, key := range keys {
		if key.LuckTrait == 100 {
			angelCount++
		}
	}

	var angelMultiplier float64

	if angelCount == 0 {
		angelMultiplier = 0.85
	} else {
		angelMultiplier = 0.85 + (0.07 * float64(angelCount))
	}

	return math.Round(angelMultiplier*100) / 100
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
			return 140
		} else if !sameHouse && sameType {
			return 110
		} else if sameHouse && !sameType {
			return 95
		} else {
			return 80
		}
	} else if keyCount == 3 {
		if sameHouse && sameType {
			return 300
		} else if !sameHouse && sameType {
			return 240
		} else if sameHouse && !sameType {
			return 200
		} else {
			return 175
		}
	} else if keyCount == 5 {
		if sameHouse && sameType {
			return 600
		} else if !sameHouse && sameType {
			return 485
		} else if sameHouse && !sameType {
			return 410
		} else {
			return 360
		}
	} else if keyCount == 15 {
		if sameHouse && sameType {
			return 3500
		} else if !sameHouse && sameType {
			return 2000
		} else if sameHouse && !sameType {
			return 1500
		} else {
			return 1250
		}
	} else {
		return 0 // if the key count is neither 1, 2, 3, 5 nor 15, then it is invalid. however, since this error is already being acknowledged in the main function, we just return 0.
	}
}

/*
Calculates the keychain bonus for a subpool.
*/
func CalculateKeychainCombo(keychainIds []int, superiorKeychainId int) float64 {
	var keychainBonus float64 = 1
	// if there is only 1 `keychainId` and superiorKeychainId == -1, check if the `keychainId` is -1
	if len(keychainIds) >= 1 && superiorKeychainId == -1 {
		if keychainIds[0] == -1 {
			return keychainBonus // return 1. -1 means no keychain is being staked.
		}

		return 1.1 // otherwise, return 1.1 if 1 or more keychains is/are not -1.
	}

	// if there is only 1 `keychainId` and superiorKeychainId != -1
	if len(keychainIds) == 1 && superiorKeychainId != -1 {
		if keychainIds[0] == -1 {
			return 1.5
		}
	}

	// if keychainIds is empty and superiorKeychainId == -1
	if len(keychainIds) == 0 && superiorKeychainId == -1 {
		return keychainBonus
	}

	// in case `keychainIds` is empty, check if `superiorKeychainId` is -1
	if len(keychainIds) == 0 || keychainIds == nil && superiorKeychainId != -1 {
		return 1.5
	}

	return keychainBonus
}
