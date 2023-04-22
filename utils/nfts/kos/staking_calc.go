package utils_kos

import (
	"math"
	"nbc-backend-api-v2/models"
)

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
