package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

/*
Returns all active and closed staking pools, each with their respective staking pool data
*/
type AllStakingPools struct {
	ActivePools []*StakingPool `json:"activePools,omitempty"` // all active staking pools
	ClosedPools []*StakingPool `json:"closedPools,omitempty"` // all closed staking pools
}

/*
Defines the `StakingPool` collection which is used to store all staking pool data.
*/
type StakingPool struct {
	ID               primitive.ObjectID `bson:"_id,omitempty"`              // the object ID of the staking pool
	StakingPoolID    int                `bson:"stakingPoolID,omitempty"`    // unique ID for each staking pool (starts at 1 for the first staking pool, increments everytime)
	Reward           Reward             `bson:"reward,omitempty"`           // the reward for staking in this pool
	TotalYieldPoints float64            `bson:"totalYieldPoints,omitempty"` // the total yield points generated across ALL stakers' subpools (calculated from `StakingSubpool`)
	EntryAllowance   time.Time          `bson:"entryAllowance,omitempty"`   // the time when stakers are allowed to enter the pool (also when the staking pool is created)
	StartTime        time.Time          `bson:"startTime,omitempty"`        // the start time of the staking pool (where entry is no longer allowed and staking has started)
	EndTime          time.Time          `bson:"endTime,omitempty"`          // when the staking pool ends (when the staking pool is closed)
	ActiveSubpools   []*StakingSubpool  `bson:"activeSubpools,omitempty"`   // the active subpools for this staking pool (points to a subpool instance from the `StakingSubpool` collection)
	ClosedSubpools   []*StakingSubpool  `bson:"closedSubpools,omitempty"`   // the closed subpools for this staking pool (either by unstaking, bans or after the pool ends. points to a subpool instance from the `StakingSubpool` collection)
}

/*
Defines the `StakingSubpool` collection which is used to store all staking subpool data for each staking pool.
*/
type StakingSubpool struct {
	ID                       primitive.ObjectID       `bson:"_id,omitempty"`                      // the object ID of the staking subpool
	SubpoolID                int                      `bson:"subpoolID,omitempty"`                // unique ID for each staking subpool (starts at 1 for the first staking subpool IN EACH POOL, increments everytime)
	Staker                   *primitive.ObjectID      `bson:"staker,omitempty"`                   // the staker that owns this subpool (points to a staker instance from the `Staker` collection)
	EnterTime                time.Time                `bson:"enterTime,omitempty"`                // the time when the staker enters the pool with this subpool
	ExitTime                 time.Time                `bson:"exitTime,omitempty"`                 // the time when the staker exits the pool with this subpool
	StakedKeys               []*KOSSimplifiedMetadata `bson:"stakedKeys,omitempty"`               // the keys of salvation staked in this subpool
	StakedKeychainID         int                      `bson:"stakedKeychainId,omitempty"`         // the keychain staked in this subpool
	StakedSuperiorKeychainID int                      `bson:"stakedSuperiorKeychainId,omitempty"` // the superior keychain staked in this subpool
	SubpoolPoints            float64                  `bson:"subpoolPoints,omitempty"`            // the subpool's yield points generated by this subpool based on the NFTs staked
	RewardClaimed            bool                     `bson:"rewardClaimed,omitempty"`            // whether the reward has been claimed or not
	Banned                   bool                     `bson:"banned,omitempty"`                   // whether the staker is banned for this particular subpool. if yes, they cannot claim the reward, even if `RewardClaimed` is false.
}

/*
Defines a StakingSubpool struct but also takes into account the `StakingPoolID` of the subpool.
*/
type StakingSubpoolWithID struct {
	StakingPoolID int `bson:"stakingPoolID,omitempty"`
	*StakingSubpool
}

/*
Represents a staker which is used to store all staker data.
*/
type Staker struct {
	ID                 primitive.ObjectID `bson:"_id,omitempty"`             // the object ID of the staker
	Wallet             string             `bson:"wallet,omitempty"`          // the wallet address of the staker
	EarnedRewards      []*Reward          `bson:"earnedRewards,omitempty"`   // the rewards earned by the staker (if they've already earned Reward A from previous pools and earn again, it will be incremented by that amount)
	TotalSubpoolPoints int                `bson:"totalPoolPoints,omitempty"` // the total pool points generated by the staker across ALL staking pools
	BannedData         *BannedData        `bson:"bannedData,omitempty"`      // the banned data of the staker. nil if the user has not been banned.
}

/*
Represents the banned data of a staker.
*/
type BannedData struct {
	ID               primitive.ObjectID `bson:"_id,omitempty"`    // the object ID of the banned data
	BannedCount      int                `bson:"bannedCount"`      // the number of times the staker has been banned
	LastBanTime      time.Time          `bson:"lastBanTime"`      // the last time the staker was banned
	CurrentUnbanTime time.Time          `bson:"currentUnbanTime"` // the time when the staker will be unbanned. if now > unban time, then the staker is considered 'unbanned'. they will be allowed staking.
}

/*
Represents a Reward for a staking pool.
*/
type Reward struct {
	Name   string  `bson:"name"`   // example: "REC", "Limited Edition Collection X", etc.
	Amount float64 `bson:"amount"` // the amount of the rewards. example: if Name is "REC" and Amount is 100, then the reward is 100 REC in total.
}

/*
Represents a KeyCombo struct, used to determine the key combo multiplier when staking.
*/
type KeyCombo struct {
	KeyCount int      `bson:"keyCount"` // the number of keys in the combo (only 1, 2, 3, 5 and 15 accepted)
	Houses   []string `bson:"houses"`   // the house of each key
	Types    []string `bson:"types"`    // the type of each key
}

/*
Represents a detailed way of calculating the subpool points, breaking down how the points are calculated.
*/
type DetailedSubpoolPoints struct {
	LuckAndLuckBoostSum float64 `bson:"luckAndLuckBoostSum"` // the sum of the luck and luck boost of all keys
	KeyCombo            float64 `bson:"keyCombo"`            // the key combo multiplier
	KeychainCombo       float64 `bson:"keychainCombo"`       // the keychain combo multiplier
	Total               float64 `bson:"total"`               // the total subpool points (calculated by the formula)
}

/*
A staker's inventory. Used for API calls to display all the key, keychain and superior keychain IDs of a staker.
*/
type KOSStakerInventory struct {
	Wallet               string          `json:"wallet"`
	KeyData              []*KeyData      `json:"keyData"`
	KeychainData         []*KeychainData `json:"keychainData"`
	SuperiorKeychainData []*KeychainData `json:"superiorKeychainData"`
}

/*
Additional data for a Key Of Salvation. Checks (on top of having the metadata) if the key is stakeable.
*/
type KeyData struct {
	KeyMetadata *KOSSimplifiedMetadata `json:"keyMetadata"`
	Stakeable   bool                   `json:"stakeable"`
}

/*
Used for both keychain and superior keychain. Checks if the keychain(s) is/are stakeable.
*/
type KeychainData struct {
	KeychainID int  `json:"keychainID"`
	Stakeable  bool `json:"stakeable"`
}
