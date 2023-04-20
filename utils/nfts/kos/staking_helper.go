package utils_kos

import "math"

/*
Calculates the yield points generated for the user's subpool based on the given parameters.

	`luck` the luck of the key
	`luckBoost` the luck boost of the key
	`keyCombo` the key combo bonus
	`keychainBonus` the keychain bonus of the key
*/
func YieldPointsCalc(luck, luckBoost, keyCombo, keychainBonus float64) float64 {
	return (100 + math.Pow(luck*luckBoost, 0.85) + keyCombo) * keychainBonus
}
