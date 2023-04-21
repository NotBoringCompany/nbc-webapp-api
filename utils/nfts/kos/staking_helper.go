package utils_kos

import (
	"context"
	"errors"
	"fmt"
	"math"
	"nbc-backend-api-v2/models"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

/*
Adds a new staking pool to the RHStakingPool collection.
*/
func AddStakingPool(collection *mongo.Collection, rewardName string, rewardAmount int) error {
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
		StartTime:          time.Now(),
		StakeTimeAllowance: time.Now().Add(time.Hour * 24 * 1),
		EndTime:            time.Now().Add(time.Hour * 24 * 7),
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
Gets the next staking pool ID from the RHStakingPool collection.
*/
func GetNextStakingPoolID(collection *mongo.Collection) (int, error) {
	// collection must be RHStakingPool.
	if collection.Name() != "RHStakingPool" {
		return -1, errors.New("collection must be RHStakingPool")
	}

	pipeline := mongo.Pipeline{
		bson.D{{Key: "$sort", Value: bson.D{{Key: "stakingPoolID", Value: -1}}}},
		bson.D{{Key: "$limit", Value: 1}},
		bson.D{{Key: "$project", Value: bson.D{{Key: "_id", Value: 0}, {Key: "stakingPoolID", Value: 1}}}},
	}

	cursor, err := collection.Aggregate(context.Background(), pipeline)
	if err != nil {
		return -1, err
	}

	defer cursor.Close(context.Background())

	var result struct{ StakingPoolID int }
	if cursor.Next(context.Background()) {
		err = cursor.Decode(&result)
		if err != nil {
			return -1, err
		}
	}

	fmt.Println("Highest stakingPoolID: ", result.StakingPoolID)

	return result.StakingPoolID + 1, nil
}

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

/*
Calculates the `keyCombo` bonus for a subpool.
NOTE: Calculations here may NOT be final and may be subject to change.

	`keyCount` the amount of keys to stake
	`houses` the houses of the keys to stake
	`types` the types of the keys to stake
*/
func BaseKeyCombo(keyCount int, houses, types []string) (float64, error) {
	// if there is only one key, then there is no combo bonus
	if keyCount == 1 {
		return 0, nil
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
			return 25, nil
		} else if !sameHouse && sameType {
			return 17, nil
		} else if sameHouse && !sameType {
			return 13, nil
		} else {
			return 10, nil
		}
	} else if keyCount == 3 {
		if sameHouse && sameType {
			return 55, nil
		} else if !sameHouse && sameType {
			return 40, nil
		} else if sameHouse && !sameType {
			return 30, nil
		} else {
			return 25, nil
		}
	} else if keyCount == 5 {
		if sameHouse && sameType {
			return 150, nil
		} else if !sameHouse && sameType {
			return 110, nil
		} else if sameHouse && !sameType {
			return 90, nil
		} else {
			return 75, nil
		}
	} else if keyCount == 15 {
		if sameHouse && sameType {
			return 1000, nil
		} else if !sameHouse && sameType {
			return 600, nil
		} else if sameHouse && !sameType {
			return 400, nil
		} else {
			return 300, nil
		}
	} else {
		return 0, errors.New("invalid key count for staking") // if the key count is neither 1, 2, 3, 5 nor 15, then it is invalid.
	}
}
