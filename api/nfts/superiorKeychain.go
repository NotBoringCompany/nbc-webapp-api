package api_nfts

import (
	"nbc-backend-api-v2/configs"
	UtilsKOS "nbc-backend-api-v2/utils/nfts/kos"
)

func CheckIfSuperiorKeychainStaked(stakingPoolId, superiorKeychainId int) (bool, error) {
	return UtilsKOS.CheckIfSuperiorKeychainStaked(configs.GetCollections(configs.DB, "RHStakingPool"), stakingPoolId, superiorKeychainId)
}
