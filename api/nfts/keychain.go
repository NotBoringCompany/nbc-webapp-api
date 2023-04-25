package api_nfts

import (
	"nbc-backend-api-v2/configs"
	UtilsKOS "nbc-backend-api-v2/utils/nfts/kos"
)

func CheckIfKeychainStaked(stakingPoolId, keychainId int) (bool, error) {
	return UtilsKOS.CheckIfKeychainStaked(configs.GetCollections(configs.DB, "RHStakingPool"), stakingPoolId, keychainId)
}
