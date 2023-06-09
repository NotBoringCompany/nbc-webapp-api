package api_kos

import (
	"fmt"
	"log"
	"math/big"
	"nbc-backend-api-v2/configs"
	"nbc-backend-api-v2/models"
	UtilsKeychain "nbc-backend-api-v2/utils/nfts/keychain"
	UtilsKOS "nbc-backend-api-v2/utils/nfts/kos"
	UtilsSuperiorKeychain "nbc-backend-api-v2/utils/nfts/superior_keychain"
	"sync"

	"github.com/robfig/cron/v3"
)

/********************
CHAIN FUNCTIONS (FOR SIMPLIFICATION IN SINGLE API CALLS)
********************/

/*
Returns the following:

1. the wallet's owned key, keychain and superior keychain IDs
2. the metadata for the wallet's owned key ID
3. checks if any of the keys, keychains and/or superior keychains are staked (in the staking pool with the `stakingPoolId`)
*/
func StakerInventory(wallet string, stakingPoolId int) (*models.KOSStakerInventory, error) {
	ownedKeyIds, err := UtilsKOS.OwnerIDs(wallet)
	if err != nil {
		return nil, err
	}
	ownedKeychainIds, err := UtilsKeychain.OwnerIDs(wallet)
	if err != nil {
		return nil, err
	}
	ownedSuperiorKeychainIds, err := UtilsSuperiorKeychain.OwnerIDs(wallet)
	if err != nil {
		return nil, err
	}

	keyIds := make([]int, len(ownedKeyIds))
	for i, id := range ownedKeyIds {
		keyIds[i] = int(id.Int64())
	}

	keyData, err := UtilsKOS.FetchSimplifiedMetadataConcurrent(keyIds)
	if err != nil {
		return nil, err
	}

	var keyMetadataAPI, keychainData, superiorKeychainData []*models.NFTData
	var wg sync.WaitGroup

	for _, md := range keyData {
		wg.Add(1)
		go func(md *models.KOSSimplifiedMetadata) {
			defer wg.Done()
			isStaked, err := UtilsKOS.CheckIfKeyStaked(configs.GetCollections(configs.DB, "RHStakingPool"), stakingPoolId, md)
			if err != nil {
				log.Printf("Error checking if key is staked for token ID %d: %v\n", md.TokenID, err)
				return
			}
			keyMetadataAPI = append(keyMetadataAPI, &models.NFTData{
				Name:      fmt.Sprintf("Key Of Salvation #%d", md.TokenID),
				ImageUrl:  md.AnimationUrl,
				TokenID:   md.TokenID,
				Metadata:  md,
				Stakeable: !isStaked,
			})
		}(md)
	}

	for _, id := range ownedKeychainIds {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			isStaked, err := UtilsKOS.CheckIfKeychainStaked(configs.GetCollections(configs.DB, "RHStakingPool"), stakingPoolId, id)
			if err != nil {
				log.Printf("Error checking if keychain is staked for token ID %d: %v\n", id, err)
				return
			}
			keychainData = append(keychainData, &models.NFTData{
				Name:      fmt.Sprintf("Keychain #%d", id),
				ImageUrl:  "https://realmhunter-kos.fra1.digitaloceanspaces.com/keychains/keychain.mp4",
				TokenID:   id,
				Stakeable: !isStaked,
			})
		}(int(id.Int64()))
	}

	for _, id := range ownedSuperiorKeychainIds {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			isStaked, err := UtilsKOS.CheckIfSuperiorKeychainStaked(configs.GetCollections(configs.DB, "RHStakingPool"), stakingPoolId, id)
			if err != nil {
				log.Printf("Error checking if superior keychain is staked for token ID %d: %v\n", id, err)
				return
			}
			superiorKeychainData = append(superiorKeychainData, &models.NFTData{
				Name:      fmt.Sprintf("Superior Keychain #%d", id),
				ImageUrl:  "https://realmhunter-kos.fra1.digitaloceanspaces.com/keychains/superiorKeychain.mp4",
				TokenID:   id,
				Stakeable: !isStaked,
			})
		}(int(id.Int64()))
	}

	wg.Wait()

	return &models.KOSStakerInventory{
		KeyData:              keyMetadataAPI,
		KeychainData:         keychainData,
		SuperiorKeychainData: superiorKeychainData,
	}, nil
	//////////////////// START OF CHANGE //////////////////////////////

	// ownedKeyIds, err := UtilsKOS.OwnerIDs(wallet)
	// if err != nil {
	// 	return nil, err
	// }
	// ownedKeychainIds, err := UtilsKeychain.OwnerIDs(wallet)
	// if err != nil {
	// 	return nil, err
	// }
	// ownedSuperiorKeychainIds, err := UtilsSuperiorKeychain.OwnerIDs(wallet)
	// if err != nil {
	// 	return nil, err
	// }

	// var keyIds []int
	// // convert big.Int to int
	// for _, id := range ownedKeyIds {
	// 	keyIds = append(keyIds, int(id.Int64()))
	// }

	// // fetch the metadata for each owned key, keychain and superior keychain
	// keyData, err := UtilsKOS.FetchSimplifiedMetadataConcurrent(keyIds)
	// if err != nil {
	// 	return nil, err
	// }
	// var keychainData []*models.NFTData
	// var superiorKeychainData []*models.NFTData
	// // we need to convert the key data above into NFTData to be read by the API.
	// var keyMetadataAPI []*models.NFTData

	// for _, md := range keyData {
	// 	isStaked, err := UtilsKOS.CheckIfKeyStaked(configs.GetCollections(configs.DB, "RHStakingPool"), stakingPoolId, md)
	// 	if err != nil {
	// 		log.Printf("Error checking if key is staked for token ID %d: %v\n", md.TokenID, err)
	// 	}
	// 	keyMetadataAPI = append(keyMetadataAPI, &models.NFTData{
	// Name:      fmt.Sprintf("Key Of Salvation #%d", md.TokenID),
	// ImageUrl:  md.AnimationUrl,
	// TokenID:   md.TokenID,
	// Metadata:  md,
	// Stakeable: !isStaked,
	// 	})
	// }

	// for _, id := range ownedKeychainIds {
	// 	var idInt int = int(id.Int64())
	// 	isStaked, err := UtilsKOS.CheckIfKeychainStaked(configs.GetCollections(configs.DB, "RHStakingPool"), stakingPoolId, idInt)
	// 	if err != nil {
	// 		log.Printf("Error checking if keychain is staked for token ID %d: %v\n", id, err)
	// 	}
	// 	keychainData = append(keychainData, &models.NFTData{
	// 		Name:      fmt.Sprintf("Keychain #%d", idInt),
	// ImageUrl:  "https://realmhunter-kos.fra1.digitaloceanspaces.com/keychains/keychain.mp4",
	// TokenID:   idInt,
	// Stakeable: !isStaked,
	// 	})
	// }

	// for _, id := range ownedSuperiorKeychainIds {
	// 	var idInt int = int(id.Int64())
	// 	isStaked, err := UtilsKOS.CheckIfSuperiorKeychainStaked(configs.GetCollections(configs.DB, "RHStakingPool"), stakingPoolId, idInt)
	// 	if err != nil {
	// 		log.Printf("Error checking if superior keychain is staked for token ID %d: %v\n", id, err)
	// 	}
	// 	superiorKeychainData = append(superiorKeychainData, &models.NFTData{
	// Name:      fmt.Sprintf("Superior Keychain #%d", idInt),
	// ImageUrl:  "https://realmhunter-kos.fra1.digitaloceanspaces.com/keychains/superiorKeychain.mp4",
	// TokenID:   idInt,
	// Stakeable: !isStaked,
	// 	})
	// }

	// return &models.KOSStakerInventory{
	// 	KeyData:              keyMetadataAPI,
	// 	KeychainData:         keychainData,
	// 	SuperiorKeychainData: superiorKeychainData,
	// }, nil

	////////////////////// END OF CHANGE /////////////////////////////

	// totalIDs := len(ownedKeyIds) + len(ownedKeychainIds) + len(ownedSuperiorKeychainIds)
	// allIds := make([]int, 0, totalIDs)

	// for _, id := range ownedKeyIds {
	// 	allIds = append(allIds, int(id.Int64()))
	// }
	// for _, id := range ownedKeychainIds {
	// 	allIds = append(allIds, int(id.Int64()))
	// }
	// for _, id := range ownedSuperiorKeychainIds {
	// 	allIds = append(allIds, int(id.Int64()))
	// }

	// // Fetch metadata concurrently
	// allMetadata, err := UtilsKOS.FetchSimplifiedMetadataConcurrent(allIds)
	// if err != nil {
	// 	return nil, err
	// }

	// // collect metadata results
	// var keyMetadata []*models.KOSSimplifiedMetadata
	// // both keychain and superior keychain metadata doesn't technically use the `KOSSimplifiedMetadata` struct, but for the sake of simplicity, we'll use it here
	// var keychainMetadata []*models.KOSSimplifiedMetadata
	// var superiorKeychainMetadata []*models.KOSSimplifiedMetadata

	// containsBigInt := func(s []*big.Int, val *big.Int) bool {
	// 	for _, v := range s {
	// 		if v.Cmp(val) == 0 {
	// 			return true
	// 		}
	// 	}
	// 	return false
	// }

	// for _, metadata := range allMetadata {
	// 	tokenID := metadata.TokenID
	// 	switch {
	// 	case containsBigInt(ownedKeyIds, big.NewInt(int64(tokenID))):
	// 		keyMetadata = append(keyMetadata, metadata)
	// 	case containsBigInt(ownedKeychainIds, big.NewInt(int64(tokenID))):
	// 		keychainMetadata = append(keychainMetadata, metadata)
	// 	case containsBigInt(ownedSuperiorKeychainIds, big.NewInt(int64(tokenID))):
	// 		superiorKeychainMetadata = append(superiorKeychainMetadata, metadata)
	// 	}
	// }

	// check if any of the keys, keychains and/or superior keychains are staked in the specified staking pool
	// Parallelize staking checks
	// keyDataCh := make(chan *models.NFTData, len(keyMetadata))
	// keychainDataCh := make(chan *models.NFTData, len(keychainMetadata))
	// superiorKeychainDataCh := make(chan *models.NFTData, len(superiorKeychainMetadata))

	// for _, metadata := range keyMetadata {
	// 	go func(md *models.KOSSimplifiedMetadata) {
	// 		isStaked, err := UtilsKOS.CheckIfKeyStaked(configs.GetCollections(configs.DB, "RHStakingPool"), stakingPoolId, md)
	// 		if err != nil {
	// 			log.Printf("Error checking if key is staked for token ID %d: %v\n", md.TokenID, err)
	// 		} else {
	// 			keyDataCh <- &models.NFTData{
	// 				Name:      fmt.Sprintf("Key Of Salvation #%d", md.TokenID),
	// 				ImageUrl:  md.AnimationUrl,
	// 				TokenID:   md.TokenID,
	// 				Metadata:  md,
	// 				Stakeable: !isStaked,
	// 			}
	// 		}
	// 	}(metadata)
	// }

	// for _, metadata := range keychainMetadata {
	// 	go func(md *models.KOSSimplifiedMetadata) {
	// 		isStaked, err := UtilsKOS.CheckIfKeychainStaked(configs.GetCollections(configs.DB, "RHStakingPool"), stakingPoolId, md.TokenID)
	// 		if err != nil {
	// 			log.Printf("Error checking if keychain is staked for token ID %d: %v\n", md.TokenID, err)
	// 		} else {
	// 			keychainDataCh <- &models.NFTData{
	// 				Name:      fmt.Sprintf("Keychain #%d", md.TokenID),
	// 				ImageUrl:  "https://realmhunter-kos.fra1.digitaloceanspaces.com/keychains/keychain.mp4",
	// 				TokenID:   md.TokenID,
	// 				Stakeable: !isStaked,
	// 			}
	// 		}
	// 	}(metadata)
	// }

	// for _, metadata := range superiorKeychainMetadata {
	// 	go func(md *models.KOSSimplifiedMetadata) {
	// 		isStaked, err := UtilsKOS.CheckIfSuperiorKeychainStaked(configs.GetCollections(configs.DB, "RHStakingPool"), stakingPoolId, md.TokenID)
	// 		if err != nil {
	// 			log.Printf("Error checking if superior keychain is staked for token ID %d: %v\n", md.TokenID, err)
	// 		} else {
	// 			superiorKeychainDataCh <- &models.NFTData{
	// 				Name:      fmt.Sprintf("Superior Keychain #%d", md.TokenID),
	// 				ImageUrl:  "https://realmhunter-kos.fra1.digitaloceanspaces.com/keychains/superiorKeychain.mp4",
	// 				TokenID:   md.TokenID,
	// 				Stakeable: !isStaked,
	// 			}
	// 		}
	// 	}(metadata)
	// }

	// // Collect staking check results
	// keyData := make([]*models.NFTData, 0, len(keyMetadata))
	// keychainData := make([]*models.NFTData, 0, len(keychainMetadata))
	// superiorKeychainData := make([]*models.NFTData, 0, len(superiorKeychainMetadata))

	// for i := 0; i < len(keyMetadata); i++ {
	// 	keyData = append(keyData, <-keyDataCh)
	// }

	// for i := 0; i < len(keychainMetadata); i++ {
	// 	keychainData = append(keychainData, <-keychainDataCh)
	// }

	// for i := 0; i < len(superiorKeychainMetadata); i++ {
	// 	superiorKeychainData = append(superiorKeychainData, <-superiorKeychainDataCh)
	// }
}

/*
Returns all active and closed staking pools, each with their respective staking pool data
*/
func FetchStakingPoolData() (*models.AllStakingPools, error) {
	// we first fetch all stakeable staking pools.
	stakeablePools, err := UtilsKOS.GetAllStakeableStakingPools(configs.GetCollections(configs.DB, "RHStakingPool"))
	if err != nil {
		return nil, err
	}

	// we then fetch all ongoing staking pools
	ongoingPools, err := UtilsKOS.GetAllOngoingStakingPools(configs.GetCollections(configs.DB, "RHStakingPool"))
	if err != nil {
		return nil, err
	}

	// we then fetch all closed staking pools.
	closedPools, err := UtilsKOS.GetAllClosedStakingPools(configs.GetCollections(configs.DB, "RHStakingPool"))
	if err != nil {
		return nil, err
	}

	return &models.AllStakingPools{
		StakeablePools: stakeablePools,
		OngoingPools:   ongoingPools,
		ClosedPools:    closedPools,
	}, nil
}

func FetchTokenPreAddSubpoolData(
	stakingPoolId int,
	keyIds,
	keychainIds []int,
	superiorKeychainId int,
) (*models.DetailedTokenSubpoolPreAddCalc, error) {
	return UtilsKOS.GetTokenPreAddSubpoolData(configs.GetCollections(configs.DB, "RHStakingPool"), stakingPoolId, keyIds, keychainIds, superiorKeychainId)
}

/*
Fetches the subpool data but with an API request format for StakingSubpool data.
*/
func FetchSubpoolData(stakingPoolId, subpoolId int) (*models.StakingSubpoolAlt, error) {
	// fetch the subpool data
	subpool, err := UtilsKOS.GetSubpoolDataAPI(configs.GetCollections(configs.DB, "RHStakingPool"), stakingPoolId, subpoolId)
	if err != nil {
		return nil, err
	}

	// get the staker wallet from the obj ID
	staker, err := UtilsKOS.GetStakerFromObjID(configs.GetCollections(configs.DB, "RHStakerData"), subpool.Staker)
	if err != nil {
		return nil, err
	}

	// return the subpool data in the format we want
	return &models.StakingSubpoolAlt{
		SubpoolID:              subpool.SubpoolID,
		Staker:                 subpool.Staker,
		StakerWallet:           staker.Wallet,
		EnterTime:              subpool.EnterTime,
		ExitTime:               subpool.ExitTime,
		StakedKeys:             subpool.StakedKeys,
		StakedKeychains:        subpool.StakedKeychains,
		StakedSuperiorKeychain: subpool.StakedSuperiorKeychain,
		SubpoolPoints:          subpool.SubpoolPoints,
		RewardClaimable:        subpool.RewardClaimable,
		RewardClaimed:          subpool.RewardClaimed,
		Banned:                 subpool.Banned,
	}, nil
}

/********************
END OF CHAIN FUNCTIONS
********************/

func FetchMetadata(tokenId int) (*models.KOSMetadata, error) {
	return UtilsKOS.FetchMetadata(tokenId)
}

func GetStakerRECBalance(wallet string) (float64, error) {
	return UtilsKOS.GetStakerRECBalance(configs.GetCollections(configs.DB, "RHStakerData"), wallet)
}

func CalculateStakerTotalSubpoolPoints(stakingPoolId int, stakerWallet string) (float64, error) {
	return UtilsKOS.CalculateStakerTotalSubpoolPoints(configs.GetCollections(configs.DB, "RHStakingPool"), stakingPoolId, stakerWallet)
}

func CalcTotalTokenShare(stakingPoolId int, stakerWallet string) (float64, error) {
	return UtilsKOS.CalcTotalTokenShare(configs.GetCollections(configs.DB, "RHStakingPool"), stakingPoolId, stakerWallet)
}

func FetchSimplifiedMetadata(tokenId int) (*models.KOSSimplifiedMetadata, error) {
	return UtilsKOS.FetchSimplifiedMetadata(tokenId)
}

func OwnerIDs(address string) ([]*big.Int, error) {
	return UtilsKOS.OwnerIDs(address)
}

func GetTotalTokenReward(stakingPoolId int) (float64, error) {
	return UtilsKOS.GetTotalTokenReward(configs.GetCollections(configs.DB, "RHStakingPool"), stakingPoolId)
}

func GetStakingPoolData(stakingPoolId int) (*models.StakingPool, error) {
	return UtilsKOS.GetStakingPoolData(configs.GetCollections(configs.DB, "RHStakingPool"), stakingPoolId)
}

func CheckSubpoolComboEligibility(stakingPoolId, keyCount int, stakerWallet string) (bool, error) {
	return UtilsKOS.CheckSubpoolComboEligibilityAlt(configs.GetCollections(configs.DB, "RHStakingPool"), stakingPoolId, stakerWallet, keyCount)
}

func CalculateSubpoolPoints(keyIds, keychainIds []int, superiorKeychainId int) float64 {
	metadatas := UtilsKOS.GetMetadataFromIDs(keyIds)

	return UtilsKOS.CalculateSubpoolPoints(metadatas, keychainIds, superiorKeychainId)
}

func BacktrackSubpoolPoints(stakingPoolId, subpoolId int) (*struct {
	LuckAndLuckBoostSum float64 `json:"luckAndLuckBoostSum"`
	AngelMultiplier     float64 `json:"angelMultiplier"`
	KeyCombo            float64 `json:"keyCombo"`
	KeychainCombo       float64 `json:"keychainCombo"`
	TotalSubpoolPoints  float64 `json:"totalSubpoolPoints"`
}, error) {
	return UtilsKOS.BacktrackSubpoolPoints(configs.GetCollections(configs.DB, "RHStakingPool"), stakingPoolId, subpoolId)
}
func CalculateSubpoolTokenShare(stakingPoolId, subpoolId int) (float64, error) {
	return UtilsKOS.CalcSubpoolTokenShare(configs.GetCollections(configs.DB, "RHStakingPool"), stakingPoolId, subpoolId)
}

func CheckIfStakerBanned(wallet string) (bool, error) {
	return UtilsKOS.CheckIfStakerBanned(configs.GetCollections(configs.DB, "RHStakerData"), wallet)
}

func GetStakerSubpools(stakerWallet string) ([]*models.StakingSubpoolWithID, error) {
	return UtilsKOS.GetStakerSubpools(configs.GetCollections(configs.DB, "RHStakingPool"), stakerWallet)
}

func CheckPoolTimeAllowanceExceeded(stakingPoolId int) (bool, error) {
	return UtilsKOS.CheckPoolTimeAllowanceExceeded(configs.GetCollections(configs.DB, "RHStakingPool"), stakingPoolId)
}

func CheckIfKeysStaked(stakingPoolId int, keyIds []int) (bool, error) {
	metadatas := UtilsKOS.GetMetadataFromIDs(keyIds)
	return UtilsKOS.CheckIfKeysStaked(configs.GetCollections(configs.DB, "RHStakingPool"), stakingPoolId, metadatas)
}

func AddSubpool(keyIds []int, sessionToken, stakerWallet string, stakingPoolId int, keychainIds []int, superiorKeychainId int) error {
	metadatas := UtilsKOS.GetMetadataFromIDs(keyIds)

	return UtilsKOS.AddSubpool(configs.GetCollections(configs.DB, "RHStakingPool"), sessionToken, stakingPoolId, stakerWallet, metadatas, keychainIds, superiorKeychainId)
}

func AddStakingPool(rewardName string, rewardAmount float64) error {
	return UtilsKOS.AddStakingPool(configs.GetCollections(configs.DB, "RHStakingPool"), rewardName, rewardAmount)
}

func ClaimReward(sessionToken, stakerWallet string, stakingPoolId, subpoolId int) error {
	return UtilsKOS.ClaimReward(configs.GetCollections(configs.DB, "RHStakingPool"), sessionToken, stakerWallet, stakingPoolId, subpoolId)
}

func UnstakeFromSubpool(sessionToken, wallet string, stakingPoolId, subpoolId int) error {
	return UtilsKOS.UnstakeFromSubpool(configs.GetCollections(configs.DB, "RHStakingPool"), sessionToken, wallet, stakingPoolId, subpoolId)
}

func UnstakeFromStakingPool(stakingPoolId int, stakerWallet string) error {
	return UtilsKOS.UnstakeFromStakingPool(configs.GetCollections(configs.DB, "RHStakingPool"), stakingPoolId, stakerWallet)
}

func GetTotalSubpoolPoints(stakingPoolId int) (float64, error) {
	return UtilsKOS.GetTotalSubpoolPoints(configs.GetCollections(configs.DB, "RHStakingPool"), stakingPoolId)
}

func GetAllStakingPools() ([]*models.StakingPool, error) {
	return UtilsKOS.GetAllStakingPools(configs.GetCollections(configs.DB, "RHStakingPool"))
}

func GetAllActiveSubpools() ([]*models.StakingSubpoolWithID, error) {
	return UtilsKOS.GetAllActiveSubpools(configs.GetCollections(configs.DB, "RHStakingPool"))
}

func GetAllStakedKeyIDs(stakingPoolId int) ([]int, error) {
	return UtilsKOS.GetAllStakedKeyIDs(configs.GetCollections(configs.DB, "RHStakingPool"), stakingPoolId)
}

func GetAllStakedKeychainIDs(stakingPoolId int) ([]int, error) {
	return UtilsKOS.GetAllStakedKeychainIDs(configs.GetCollections(configs.DB, "RHStakingPool"), stakingPoolId)
}

func GetAllStakedSuperiorKeychainIDs(stakingPoolId int) ([]int, error) {
	return UtilsKOS.GetAllStakedSuperiorKeychainIDs(configs.GetCollections(configs.DB, "RHStakingPool"), stakingPoolId)
}

/*
Gets the detailed subpool points (how it was calculated)
*/
func DetailedSubpoolPoints(keyIds, keychainIds []int, superiorKeychainId int) *models.DetailedSubpoolPoints {
	metadatas := UtilsKOS.GetMetadataFromIDs(keyIds)

	var luckAndLuckBoostSum float64
	for _, metadata := range metadatas {
		luckAndLuckBoostSum += (metadata.LuckTrait * metadata.LuckBoostTrait)
	}

	keyCombo := UtilsKOS.CalculateKeyCombo(metadatas)
	keychainCombo := UtilsKOS.CalculateKeychainCombo(keychainIds, superiorKeychainId)

	return &models.DetailedSubpoolPoints{
		LuckAndLuckBoostSum: luckAndLuckBoostSum,
		KeyCombo:            keyCombo,
		KeychainCombo:       keychainCombo,
		ComboSum:            CalculateSubpoolPoints(keyIds, keychainIds, superiorKeychainId),
	}
}

/*********************

CRON SCHEDULER FUNCTIONS

**********************/

/*
Updates the total yield points of all active staking pools every 1 minute.
*/
func UpdateTotalYieldPointsScheduler() *cron.Cron {
	scheduler := cron.New()

	// runs every 1 minute
	scheduler.AddFunc("*/1 * * * *", func() {
		err := UtilsKOS.UpdateTotalYieldPoints(configs.GetCollections(configs.DB, "RHStakingPool"))
		if err != nil {
			panic(err)
		}
	})

	return scheduler
}

/*
Adds a scheduler to `CloseSubpoolsOnStakeEnd` to run it every 5 mins.
*/
func CloseSubpoolsOnStakeEndScheduler() *cron.Cron {
	scheduler := cron.New()

	// runs every 5 mins
	scheduler.AddFunc("*/5 * * * *", func() {
		err := UtilsKOS.CloseSubpoolsOnStakeEnd(configs.GetCollections(configs.DB, "RHStakingPool"))
		if err != nil {
			panic(err)
		}
	})

	return scheduler
}

/*
Adds a scheduler to `VerifyStakerOwnership` to run it every 5 seconds.
*/
func VerifyStakerOwnershipScheduler() *cron.Cron {
	scheduler := cron.New()

	// runs every 5 seconds
	scheduler.AddFunc("*/5 * * * * *", func() {
		err := UtilsKOS.VerifyStakerOwnership(configs.GetCollections(configs.DB, "RHStakingPool"))
		if err != nil {
			panic(err)
		}
	})

	return scheduler
}

/*
Adds a scheduler to `VerifyStakingPoolStakerCount` to run it every minute.
*/
func VerifyStakingPoolStakerCountScheduler() *cron.Cron {
	scheduler := cron.New()

	// runs every minute
	scheduler.AddFunc("*/1 * * * *", func() {
		err := UtilsKOS.CheckStakingPoolStakerCount(configs.GetCollections(configs.DB, "RHStakingPool"))
		if err != nil {
			panic(err)
		}
	})

	return scheduler
}

/*
Adds a scheduler to `RemoveExpiredUnclaimableSubpools` to run it every minute.
*/
func RemoveExpiredUnclaimableSubpoolsScheduler() *cron.Cron {
	scheduler := cron.New()

	// run every minute
	scheduler.AddFunc("*/1 * * * *", func() {
		err := UtilsKOS.RemoveExpiredUnclaimableSubpools(configs.GetCollections(configs.DB, "RHStakingPool"))
		if err != nil {
			panic(err)
		}
	})

	return scheduler
}

/*********************

END OF CRON SCHEDULER FUNCTIONS

**********************/
