package api_kos

import (
	"log"
	"math/big"
	"nbc-backend-api-v2/configs"
	"nbc-backend-api-v2/models"
	UtilsKeychain "nbc-backend-api-v2/utils/nfts/keychain"
	UtilsKOS "nbc-backend-api-v2/utils/nfts/kos"
	UtilsSuperiorKeychain "nbc-backend-api-v2/utils/nfts/superior_keychain"

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

	totalIDs := len(ownedKeyIds) + len(ownedKeychainIds) + len(ownedSuperiorKeychainIds)
	allIds := make([]int, 0, totalIDs)

	for _, id := range ownedKeyIds {
		allIds = append(allIds, int(id.Int64()))
	}
	for _, id := range ownedKeychainIds {
		allIds = append(allIds, int(id.Int64()))
	}
	for _, id := range ownedSuperiorKeychainIds {
		allIds = append(allIds, int(id.Int64()))
	}

	// Fetch metadata concurrently
	allMetadata, err := UtilsKOS.FetchSimplifiedMetadataConcurrent(allIds)
	if err != nil {
		return nil, err
	}

	// collect metadata results
	var keyMetadata []*models.KOSSimplifiedMetadata
	var keychainMetadata []*models.KOSSimplifiedMetadata
	var superiorKeychainMetadata []*models.KOSSimplifiedMetadata

	containsBigInt := func(s []*big.Int, val *big.Int) bool {
		for _, v := range s {
			if v.Cmp(val) == 0 {
				return true
			}
		}
		return false
	}

	for _, metadata := range allMetadata {
		tokenID := metadata.TokenID
		switch {
		case containsBigInt(ownedKeyIds, big.NewInt(int64(tokenID))):
			keyMetadata = append(keyMetadata, metadata)
		case containsBigInt(ownedKeychainIds, big.NewInt(int64(tokenID))):
			keychainMetadata = append(keychainMetadata, metadata)
		case containsBigInt(ownedSuperiorKeychainIds, big.NewInt(int64(tokenID))):
			superiorKeychainMetadata = append(superiorKeychainMetadata, metadata)
		}
	}

	// check if any of the keys, keychains and/or superior keychains are staked in the specified staking pool
	// Parallelize staking checks
	keyDataCh := make(chan *models.KeyData, len(keyMetadata))
	keychainDataCh := make(chan *models.KeychainData, len(keychainMetadata))
	superiorKeychainDataCh := make(chan *models.KeychainData, len(superiorKeychainMetadata))

	for _, metadata := range keyMetadata {
		go func(md *models.KOSSimplifiedMetadata) {
			isStaked, err := UtilsKOS.CheckIfKeyStaked(configs.GetCollections(configs.DB, "RHStakingPool"), stakingPoolId, md)
			if err != nil {
				log.Printf("Error checking if key is staked for token ID %d: %v\n", md.TokenID, err)
			} else {
				keyDataCh <- &models.KeyData{
					KeyMetadata: md,
					Stakeable:   !isStaked,
				}
			}
		}(metadata)
	}

	for _, metadata := range keychainMetadata {
		go func(md *models.KOSSimplifiedMetadata) {
			isStaked, err := UtilsKOS.CheckIfKeychainStaked(configs.GetCollections(configs.DB, "RHStakingPool"), stakingPoolId, md.TokenID)
			if err != nil {
				log.Printf("Error checking if keychain is staked for token ID %d: %v\n", md.TokenID, err)
			} else {
				keychainDataCh <- &models.KeychainData{
					KeychainID: md.TokenID,
					Stakeable:  !isStaked,
				}
			}
		}(metadata)
	}

	for _, metadata := range superiorKeychainMetadata {
		go func(md *models.KOSSimplifiedMetadata) {
			isStaked, err := UtilsKOS.CheckIfSuperiorKeychainStaked(configs.GetCollections(configs.DB, "RHStakingPool"), stakingPoolId, md.TokenID)
			if err != nil {
				log.Printf("Error checking if superior keychain is staked for token ID %d: %v\n", md.TokenID, err)
			} else {
				superiorKeychainDataCh <- &models.KeychainData{
					KeychainID: md.TokenID,
					Stakeable:  !isStaked,
				}
			}
		}(metadata)
	}

	// Collect staking check results
	keyData := make([]*models.KeyData, 0, len(keyMetadata))
	keychainData := make([]*models.KeychainData, 0, len(keychainMetadata))
	superiorKeychainData := make([]*models.KeychainData, 0, len(superiorKeychainMetadata))

	for i := 0; i < len(keyMetadata); i++ {
		keyData = append(keyData, <-keyDataCh)
	}

	for i := 0; i < len(keychainMetadata); i++ {
		keychainData = append(keychainData, <-keychainDataCh)
	}

	for i := 0; i < len(superiorKeychainMetadata); i++ {
		superiorKeychainData = append(superiorKeychainData, <-superiorKeychainDataCh)
	}

	return &models.KOSStakerInventory{
		Wallet:               wallet,
		KeyData:              keyData,
		KeychainData:         keychainData,
		SuperiorKeychainData: superiorKeychainData,
	}, nil
}

// func StakerInventory(wallet string, stakingPoolId int) (*models.KOSStakerInventory, error) {
// 	// first, get the owned key, keychain and superior keychain IDs

// 	ownedKeyIds, err := UtilsKOS.OwnerIDs(wallet)
// 	if err != nil {
// 		return nil, err
// 	}
// 	ownedKeychainIds, err := UtilsKeychain.OwnerIDs(wallet)
// 	if err != nil {
// 		return nil, err
// 	}
// 	ownedSuperiorKeychainIds, err := UtilsSuperiorKeychain.OwnerIDs(wallet)
// 	if err != nil {
// 		return nil, err
// 	}

// 	// then, get the metadata for the owned key ID
// 	var keyMetadatas []*models.KOSSimplifiedMetadata
// 	for _, id := range ownedKeyIds {
// 		metadata := FetchSimplifiedMetadata(int(id.Int64()))
// 		keyMetadatas = append(keyMetadatas, metadata)
// 	}

// 	var keyData []*models.KeyData
// 	var keychainData []*models.KeychainData
// 	var superiorKeychainData []*models.KeychainData

// 	// IF STAKINGPOOLID ISN'T ADDED (IS -1 OR 0), THEN RETURN THE OWNED KEY, KEYCHAIN AND SUPERIOR KEYCHAIN IDS
// 	if stakingPoolId == -1 || stakingPoolId == 0 {
// 		for _, id := range ownedKeyIds {
// 			keyData = append(keyData, &models.KeyData{
// 				KeyMetadata: FetchSimplifiedMetadata(int(id.Int64())),
// 				Stakeable:   true,
// 			})
// 		}

// 		for _, id := range ownedKeychainIds {
// 			keychainData = append(keychainData, &models.KeychainData{
// 				KeychainID: int(id.Int64()),
// 				Stakeable:  true,
// 			})
// 		}

// 		for _, id := range ownedSuperiorKeychainIds {
// 			superiorKeychainData = append(superiorKeychainData, &models.KeychainData{
// 				KeychainID: int(id.Int64()),
// 				Stakeable:  true,
// 			})
// 		}
// 		// if stakingPoolId is not -1 or 0, then it is a valid staking pool ID
// 	} else {
// 		// then, check if any of the keys, keychains and/or superior keychains are staked in the specified staking pool
// 		for _, key := range keyMetadatas {
// 			isStaked, err := UtilsKOS.CheckIfKeyStaked(configs.GetCollections(configs.DB, "RHStakingPool"), stakingPoolId, key)
// 			if err != nil {
// 				return nil, err
// 			}

// 			keyData = append(keyData, &models.KeyData{
// 				KeyMetadata: key,
// 				Stakeable:   !isStaked, // if the key is staked, it is not stakeable and vice versa.
// 			})
// 		}
// 		for _, keychainId := range ownedKeychainIds {
// 			isStaked, err := UtilsKOS.CheckIfKeychainStaked(configs.GetCollections(configs.DB, "RHStakingPool"), stakingPoolId, int(keychainId.Int64()))
// 			if err != nil {
// 				return nil, err
// 			}

// 			keychainData = append(keychainData, &models.KeychainData{
// 				KeychainID: int(keychainId.Int64()),
// 				Stakeable:  !isStaked, // if the keychain is staked, it is not stakeable and vice versa.
// 			})
// 		}
// 		for _, superiorKeychainId := range ownedSuperiorKeychainIds {
// 			isStaked, err := UtilsKOS.CheckIfSuperiorKeychainStaked(configs.GetCollections(configs.DB, "RHStakingPool"), stakingPoolId, int(superiorKeychainId.Int64()))
// 			if err != nil {
// 				return nil, err
// 			}

// 			superiorKeychainData = append(superiorKeychainData, &models.KeychainData{
// 				KeychainID: int(superiorKeychainId.Int64()),
// 				Stakeable:  !isStaked, // if the superior keychain is staked, it is not stakeable and vice versa.
// 			})
// 		}
// 	}

// 	return &models.KOSStakerInventory{
// 		Wallet:               wallet,
// 		KeyData:              keyData,
// 		KeychainData:         keychainData,
// 		SuperiorKeychainData: superiorKeychainData,
// 	}, nil
// }

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

/********************
END OF CHAIN FUNCTIONS
********************/

func FetchMetadata(tokenId int) (*models.KOSMetadata, error) {
	return UtilsKOS.FetchMetadata(tokenId)
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

func CalculateSubpoolPoints(keyIds []int, keychainId, superiorKeychainId int) float64 {
	metadatas := UtilsKOS.GetMetadataFromIDs(keyIds)

	return UtilsKOS.CalculateSubpoolPoints(metadatas, keychainId, superiorKeychainId)
}

func CalculateSubpoolTokenShare(stakingPoolId, subpoolId int) (float64, error) {
	return UtilsKOS.CalcSubpoolTokenShare(configs.GetCollections(configs.DB, "RHStakingPool"), stakingPoolId, subpoolId)
}

func CheckIfStakerBanned(wallet string) (bool, error) {
	return UtilsKOS.CheckIfStakerBanned(configs.GetCollections(configs.DB, "RHStakerData"), wallet)
}

func CheckPoolTimeAllowanceExceeded(stakingPoolId int) (bool, error) {
	return UtilsKOS.CheckPoolTimeAllowanceExceeded(configs.GetCollections(configs.DB, "RHStakingPool"), stakingPoolId)
}

func CheckIfKeysStaked(stakingPoolId int, keyIds []int) (bool, error) {
	metadatas := UtilsKOS.GetMetadataFromIDs(keyIds)
	return UtilsKOS.CheckIfKeysStaked(configs.GetCollections(configs.DB, "RHStakingPool"), stakingPoolId, metadatas)
}

func AddSubpool(keyIds []int, stakerWallet string, stakingPoolId, keychainId, superiorKeychainId int) error {
	metadatas := UtilsKOS.GetMetadataFromIDs(keyIds)

	return UtilsKOS.AddSubpool(configs.GetCollections(configs.DB, "RHStakingPool"), stakingPoolId, stakerWallet, metadatas, keychainId, superiorKeychainId)
}

func ClaimReward(stakerWallet string, stakingPoolId, subpoolId int) error {
	return UtilsKOS.ClaimReward(configs.GetCollections(configs.DB, "RHStakingPool"), stakerWallet, stakingPoolId, subpoolId)
}

func UnstakeFromSubpool(stakingPoolId, subpoolId int) error {
	return UtilsKOS.UnstakeFromSubpool(configs.GetCollections(configs.DB, "RHStakingPool"), stakingPoolId, subpoolId)
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
func DetailedSubpoolPoints(keyIds []int, keychainId, superiorKeychainId int) *models.DetailedSubpoolPoints {
	metadatas := UtilsKOS.GetMetadataFromIDs(keyIds)

	var luckAndLuckBoostSum float64
	for _, metadata := range metadatas {
		luckAndLuckBoostSum += (metadata.LuckTrait * metadata.LuckBoostTrait)
	}

	keyCombo := UtilsKOS.CalculateKeyCombo(metadatas)
	keychainCombo := UtilsKOS.CalculateKeychainCombo(keychainId, superiorKeychainId)

	return &models.DetailedSubpoolPoints{
		LuckAndLuckBoostSum: luckAndLuckBoostSum,
		KeyCombo:            keyCombo,
		KeychainCombo:       keychainCombo,
		Total:               CalculateSubpoolPoints(keyIds, keychainId, superiorKeychainId),
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
Adds a scheduler to `CloseSubpoolsOnStakeEnd` to run it every hour.
*/
func CloseSubpoolsOnStakeEndScheduler() *cron.Cron {
	scheduler := cron.New()

	// runs every hour
	scheduler.AddFunc("0 0 */1 * * *", func() {
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

/*********************

END OF CRON SCHEDULER FUNCTIONS

**********************/
