package routes_nfts

import (
	"fmt"
	ApiKOS "nbc-backend-api-v2/api/nfts/kos"
	"strconv"
	"strings"

	"nbc-backend-api-v2/responses"

	"github.com/gofiber/fiber/v2"
)

func KOSRoutes(app *fiber.App) {
	// FetchStakerInventory route
	app.Get("/kos/fetch-staker-inventory/:wallet/:stakingPoolId", func(c *fiber.Ctx) error {
		wallet := c.Params("wallet")
		stakingPoolId := c.Params("stakingPoolId")
		stakingPoolIdInt, err := strconv.Atoi(stakingPoolId)
		if err != nil {
			return c.JSON(&responses.Response{
				Status:  fiber.StatusBadRequest,
				Message: "unable to successfully convert given stakingPoolId to int.",
				Data:    nil,
			})
		}

		res, err := ApiKOS.StakerInventory(wallet, stakingPoolIdInt)
		if err != nil {
			return c.JSON(&responses.Response{
				Status:  fiber.StatusBadRequest,
				Message: fmt.Sprintf("unable to successfully fetch staker inventory for given wallet: %v", err),
				Data:    nil,
			})
		}

		return c.JSON(&responses.Response{
			Status:  fiber.StatusOK,
			Message: "successfully fetched staker inventory for given wallet.",
			Data:    &fiber.Map{"inventory": res},
		})
	})

	// FetchTokenPreAddSubpoolData route
	app.Get("/kos/fetch-token-pre-add-subpool-data/", func(c *fiber.Ctx) error {
		// get the staking pool id, subpool id, key ids, keychain id and superior keychain id from the query params
		stakingPoolId := c.Query("stakingPoolId")
		keyIds := c.Query("keyIds")
		keychainId := c.Query("keychainId")
		superiorKeychainId := c.Query("superiorKeychainId")

		keyIdsStr := strings.Split(keyIds, ",")
		keyIdsInt := make([]int, len(keyIdsStr))
		for i, keyId := range keyIdsStr {
			keyIdInt, err := strconv.Atoi(keyId)
			if err != nil {
				return c.JSON(&responses.Response{
					Status:  fiber.StatusBadRequest,
					Message: "unable to successfully convert given keyId to int.",
					Data:    nil,
				})
			}
			keyIdsInt[i] = keyIdInt
		}

		stakingPoolIdInt, err := strconv.Atoi(stakingPoolId)
		if err != nil {
			return c.JSON(&responses.Response{
				Status:  fiber.StatusBadRequest,
				Message: "unable to successfully convert given stakingPoolId to int.",
				Data:    nil,
			})
		}
		keychainIdInt, err := strconv.Atoi(keychainId)
		if err != nil {
			return c.JSON(&responses.Response{
				Status:  fiber.StatusBadRequest,
				Message: "unable to successfully convert given keychainId to int.",
				Data:    nil,
			})
		}
		superiorKeychainIdInt, err := strconv.Atoi(superiorKeychainId)
		if err != nil {
			return c.JSON(&responses.Response{
				Status:  fiber.StatusBadRequest,
				Message: "unable to successfully convert given superiorKeychainId to int.",
				Data:    nil,
			})
		}

		res, err := ApiKOS.FetchTokenPreAddSubpoolData(stakingPoolIdInt, keyIdsInt, keychainIdInt, superiorKeychainIdInt)
		if err != nil {
			return c.JSON(&responses.Response{
				Status:  fiber.StatusBadRequest,
				Message: fmt.Sprintf("unable to successfully fetch token pre add subpool data: %v", err),
				Data:    nil,
			})
		}

		return c.JSON(&responses.Response{
			Status:  fiber.StatusOK,
			Message: "successfully fetched token pre add subpool data.",
			Data:    &fiber.Map{"tokenPreAddSubpoolData": res},
		})
	})

	// FetchSubpoolData route
	app.Get("/kos/fetch-subpool-data/:stakingPoolId/:subpoolId", func(c *fiber.Ctx) error {
		stakingPoolId := c.Params("stakingPoolId")
		subpoolId := c.Params("subpoolId")
		stakingPoolIdInt, err := strconv.Atoi(stakingPoolId)
		if err != nil {
			return c.JSON(&responses.Response{
				Status:  fiber.StatusBadRequest,
				Message: "unable to successfully convert given stakingPoolId to int.",
				Data:    nil,
			})
		}
		subpoolIdInt, err := strconv.Atoi(subpoolId)
		if err != nil {
			return c.JSON(&responses.Response{
				Status:  fiber.StatusBadRequest,
				Message: "unable to successfully convert given subpoolId to int.",
				Data:    nil,
			})
		}

		res, err := ApiKOS.FetchSubpoolData(stakingPoolIdInt, subpoolIdInt)
		if err != nil {
			return c.JSON(&responses.Response{
				Status:  fiber.StatusBadRequest,
				Message: fmt.Sprintf("unable to successfully fetch subpool data for given stakingPoolId and subpoolId: %v", err),
				Data:    nil,
			})
		}

		return c.JSON(&responses.Response{
			Status:  fiber.StatusOK,
			Message: "successfully fetched subpool data for given stakingPoolId and subpoolId.",
			Data:    &fiber.Map{"subpoolData": res},
		})
	})

	app.Get("/kos/fetch-simplified-metadata/:tokenId", func(c *fiber.Ctx) error {
		tokenId := c.Params("tokenId")
		tokenIdInt, err := strconv.Atoi(tokenId)
		if err != nil {
			return c.JSON(&responses.Response{
				Status:  fiber.StatusBadRequest,
				Message: "unable to successfully convert given tokenId to int.",
				Data:    nil,
			})
		}

		res, err := ApiKOS.FetchSimplifiedMetadata(tokenIdInt)
		if err != nil {
			return c.JSON(&responses.Response{
				Status:  fiber.StatusBadRequest,
				Message: fmt.Sprintf("unable to successfully fetch simplified metadata for given tokenId: %v", err),
				Data:    nil,
			})
		}

		return c.JSON(&responses.Response{
			Status:  fiber.StatusOK,
			Message: "successfully fetched simplified metadata for given tokenId.",
			Data:    &fiber.Map{"simplifiedMetadata": res},
		})
	})

	// CalculateStakerTotalSubpoolPoints route
	app.Get("/kos/staker-total-subpool-points/:wallet/:stakingPoolId", func(c *fiber.Ctx) error {
		wallet := c.Params("wallet")
		stakingPoolId := c.Params("stakingPoolId")
		stakingPoolIdInt, err := strconv.Atoi(stakingPoolId)
		if err != nil {
			return c.JSON(&responses.Response{
				Status:  fiber.StatusBadRequest,
				Message: "unable to successfully convert given stakingPoolId to int.",
				Data:    nil,
			})
		}

		res, err := ApiKOS.CalculateStakerTotalSubpoolPoints(stakingPoolIdInt, wallet)
		if err != nil {
			return c.JSON(&responses.Response{
				Status:  fiber.StatusBadRequest,
				Message: fmt.Sprintf("unable to successfully calculate staker total subpool points for given wallet: %v", err),
				Data:    nil,
			})
		}

		return c.JSON(&responses.Response{
			Status:  fiber.StatusOK,
			Message: "successfully calculated staker total subpool points for given wallet.",
			Data:    &fiber.Map{"totalSubpoolPoints": res},
		})
	})

	// CalculateTotalTokenShare route
	app.Get("/kos/calculate-total-token-share/:wallet/:stakingPoolId", func(c *fiber.Ctx) error {
		wallet := c.Params("wallet")
		stakingPoolId := c.Params("stakingPoolId")
		stakingPoolIdInt, err := strconv.Atoi(stakingPoolId)
		if err != nil {
			return c.JSON(&responses.Response{
				Status:  fiber.StatusBadRequest,
				Message: "unable to successfully convert given stakingPoolId to int.",
				Data:    nil,
			})
		}

		res, err := ApiKOS.CalcTotalTokenShare(stakingPoolIdInt, wallet)
		if err != nil {
			return c.JSON(&responses.Response{
				Status:  fiber.StatusBadRequest,
				Message: fmt.Sprintf("unable to successfully calculate total token share for given wallet: %v", err),
				Data:    nil,
			})
		}

		return c.JSON(&responses.Response{
			Status:  fiber.StatusOK,
			Message: "successfully calculated total token share for given wallet.",
			Data:    &fiber.Map{"totalTokenShare": res},
		})
	})

	app.Get("/kos/calculate-subpool-token-share/:stakingPoolId/:subpoolId", func(c *fiber.Ctx) error {
		stakingPoolId := c.Params("stakingPoolId")
		subpoolId := c.Params("subpoolId")
		stakingPoolIdInt, err := strconv.Atoi(stakingPoolId)
		if err != nil {
			return c.JSON(&responses.Response{
				Status:  fiber.StatusBadRequest,
				Message: "unable to successfully convert given stakingPoolId to int.",
				Data:    nil,
			})
		}
		subpoolIdInt, err := strconv.Atoi(subpoolId)
		if err != nil {
			return c.JSON(&responses.Response{
				Status:  fiber.StatusBadRequest,
				Message: "unable to successfully convert given subpoolId to int.",
				Data:    nil,
			})
		}

		res, err := ApiKOS.CalculateSubpoolTokenShare(stakingPoolIdInt, subpoolIdInt)
		if err != nil {
			return c.JSON(&responses.Response{
				Status:  fiber.StatusBadRequest,
				Message: fmt.Sprintf("unable to successfully calculate subpool token share for given stakingPoolId and subpoolId: %v", err),
				Data:    nil,
			})
		}

		return c.JSON(&responses.Response{
			Status:  fiber.StatusOK,
			Message: "successfully calculated subpool token share for given stakingPoolId and subpoolId.",
			Data:    &fiber.Map{"subpoolTokenShare": res},
		})
	})

	// GetStakingPoolData route
	app.Get("/kos/staking-pool-data/:stakingPoolId", func(c *fiber.Ctx) error {
		id := c.Params("stakingPoolId")
		idInt, err := strconv.Atoi(id)
		if err != nil {
			return c.JSON(&responses.Response{
				Status:  fiber.StatusBadRequest,
				Message: "unable to successfully convert given stakingPoolId to int.",
				Data:    nil,
			})
		}

		res, err := ApiKOS.GetStakingPoolData(idInt)
		if err != nil {
			return c.JSON(&responses.Response{
				Status:  fiber.StatusBadRequest,
				Message: fmt.Sprintf("unable to successfully fetch staking pool data for given stakingPoolId: %v", err),
				Data:    nil,
			})
		}

		return c.JSON(&responses.Response{
			Status:  fiber.StatusOK,
			Message: "successfully fetched staking pool data for given stakingPoolId.",
			Data:    &fiber.Map{"stakingPoolData": res},
		})
	})

	// FetchStakingPoolData route
	app.Get("/kos/fetch-staking-pools", func(c *fiber.Ctx) error {
		res, err := ApiKOS.FetchStakingPoolData()
		if err != nil {
			return c.JSON(&responses.Response{
				Status:  fiber.StatusBadRequest,
				Message: fmt.Sprintf("unable to successfully fetch staking pool data: %v", err),
				Data:    nil,
			})
		}

		return c.JSON(&responses.Response{
			Status:  fiber.StatusOK,
			Message: "successfully fetched staking pool data.",
			Data:    &fiber.Map{"stakingPools": res},
		})
	})

	// FetchMetadata route
	app.Get("/kos/fetch-metadata/:tokenId", func(c *fiber.Ctx) error {
		tokenId := c.Params("tokenId")
		tokenIdInt, err := strconv.Atoi(tokenId)
		if err != nil {
			return c.JSON(&responses.Response{
				Status:  fiber.StatusBadRequest,
				Message: "unable to successfully convert given tokenId to int.",
				Data:    nil,
			})
		}

		res, err := ApiKOS.FetchMetadata(tokenIdInt)
		if err != nil {
			return c.JSON(&responses.Response{
				Status:  fiber.StatusBadRequest,
				Message: fmt.Sprintf("unable to successfully fetch metadata for given tokenId: %v", err),
				Data:    nil,
			})
		}

		return c.JSON(&responses.Response{
			Status:  fiber.StatusOK,
			Message: "successfully fetched metadata for given tokenId.",
			Data:    &fiber.Map{"metadata": res},
		})
	})

	// OwnerIDs route
	app.Get("/kos/owner-ids/:address", func(c *fiber.Ctx) error {
		address := c.Params("address")
		res, err := ApiKOS.OwnerIDs(address)
		if err != nil {
			return c.JSON(&responses.Response{
				Status:  fiber.StatusBadRequest,
				Message: fmt.Sprintf("unable to successfully fetch ownerIds for given address: %v", err),
				Data:    nil,
			})
		}

		return c.JSON(&responses.Response{
			Status:  fiber.StatusOK,
			Message: "successfully fetched ownerIds for given address.",
			Data:    &fiber.Map{"ownerIds": res},
		})
	})

	app.Get("/kos/get-all-staked-key-ids/:stakingPoolId", func(c *fiber.Ctx) error {
		stakingPoolId := c.Params("stakingPoolId")
		stakingPoolIdInt, err := strconv.Atoi(stakingPoolId)
		if err != nil {
			return c.JSON(&responses.Response{
				Status:  fiber.StatusBadRequest,
				Message: "unable to successfully convert given stakingPoolId to int.",
				Data:    nil,
			})
		}

		res, err := ApiKOS.GetAllStakedKeyIDs(stakingPoolIdInt)
		if err != nil {
			return c.JSON(&responses.Response{
				Status:  fiber.StatusBadRequest,
				Message: fmt.Sprintf("unable to successfully fetch all staked key ids for given stakingPoolId: %v", err),
				Data:    nil,
			})
		}

		return c.JSON(&responses.Response{
			Status:  fiber.StatusOK,
			Message: "successfully fetched all staked key ids for given stakingPoolId.",
			Data:    &fiber.Map{"allStakedKeyIds": res},
		})
	})

	// TotalTokenReward route
	app.Get("/kos/total-token-reward/:stakingPoolId", func(c *fiber.Ctx) error {
		stakingPoolId := c.Params("stakingPoolId")
		stakingPoolIdInt, err := strconv.Atoi(stakingPoolId)
		if err != nil {
			return c.JSON(&responses.Response{
				Status:  fiber.StatusBadRequest,
				Message: "unable to successfully convert given stakingPoolId to int.",
				Data:    nil,
			})
		}
		res, err := ApiKOS.GetTotalTokenReward(stakingPoolIdInt)
		if err != nil {
			return c.JSON(&responses.Response{
				Status:  fiber.StatusBadRequest,
				Message: fmt.Sprintf("unable to successfully fetch total token reward for given stakingPoolId: %v", err),
				Data:    nil,
			})
		}

		return c.JSON(&responses.Response{
			Status:  fiber.StatusOK,
			Message: "successfully fetched total token reward for given stakingPoolId.",
			Data:    &fiber.Map{"totalTokenReward": res},
		})
	})

	// CalculateSubpoolPoints route
	app.Get("/kos/calculate-subpool-points", func(c *fiber.Ctx) error {
		// get the keyIds param from the request query params
		keyIdsParam := ("keyIds")

		// convert the keyIds param to an array of ints
		keyIdsStr := strings.Split(keyIdsParam, ",")
		keyIds := make([]int, len(keyIdsStr))
		for i, id := range keyIdsStr {
			idInt, err := strconv.Atoi(id)
			if err != nil {
				return c.JSON(&responses.Response{
					Status:  fiber.StatusBadRequest,
					Message: fmt.Sprintf("unable to successfully convert given keyId to int: %v", err),
					Data:    nil,
				})
			}

			keyIds[i] = idInt
		}

		// get the keychainId and superiorKeychainId params from the request query params
		keychainIdParam := c.Query("keychainId")
		superiorKeychainIdParam := c.Query("superiorKeychainId")

		// convert the keychainId and superiorKeychainId params to ints
		keychainId, err := strconv.Atoi(keychainIdParam)
		if err != nil {
			return c.JSON(&responses.Response{
				Status:  fiber.StatusBadRequest,
				Message: fmt.Sprintf("unable to successfully convert given keychainId to int: %v", err),
				Data:    nil,
			})
		}

		superiorKeychainId, err := strconv.Atoi(superiorKeychainIdParam)
		if err != nil {
			return c.JSON(&responses.Response{
				Status:  fiber.StatusBadRequest,
				Message: fmt.Sprintf("unable to successfully convert given superiorKeychainId to int: %v", err),
				Data:    nil,
			})
		}

		// call the CalculateSubpoolPoints function
		points := ApiKOS.CalculateSubpoolPoints(keyIds, keychainId, superiorKeychainId)

		return c.JSON(&responses.Response{
			Status:  fiber.StatusOK,
			Message: "successfully calculated subpool points.",
			Data:    &fiber.Map{"points": points},
		})
	})

	app.Get("/kos/get-staker-subpools/:wallet", func(c *fiber.Ctx) error {
		wallet := c.Params("wallet")
		res, err := ApiKOS.GetStakerSubpools(wallet)
		if err != nil {
			return c.JSON(&responses.Response{
				Status:  fiber.StatusBadRequest,
				Message: fmt.Sprintf("unable to successfully fetch staker subpools for given wallet: %v", err),
				Data:    nil,
			})
		}

		return c.JSON(&responses.Response{
			Status:  fiber.StatusOK,
			Message: "successfully fetched staker subpools for given wallet.",
			Data:    &fiber.Map{"stakerSubpools": res},
		})
	})

	app.Get("/kos/calculate-subpool-token-share/:stakingPoolId/:subpoolId", func(c *fiber.Ctx) error {
		// get the stakingPoolId and subpoolId params from the request query params
		stakingPoolIdParam := c.Params("stakingPoolId")
		subpoolIdParam := c.Params("subpoolId")

		// convert the stakingPoolId and subpoolId params to ints
		stakingPoolId, err := strconv.Atoi(stakingPoolIdParam)
		if err != nil {
			return c.JSON(&responses.Response{
				Status:  fiber.StatusBadRequest,
				Message: fmt.Sprintf("unable to successfully convert given stakingPoolId to int: %v", err),
				Data:    nil,
			})
		}

		subpoolId, err := strconv.Atoi(subpoolIdParam)
		if err != nil {
			return c.JSON(&responses.Response{
				Status:  fiber.StatusBadRequest,
				Message: fmt.Sprintf("unable to successfully convert given subpoolId to int: %v", err),
				Data:    nil,
			})
		}

		// call the CalculateSubpoolTokenShare function
		tokenShare, err := ApiKOS.CalculateSubpoolTokenShare(stakingPoolId, subpoolId)
		if err != nil {
			return c.JSON(&responses.Response{
				Status:  fiber.StatusBadRequest,
				Message: fmt.Sprintf("unable to successfully calculate subpool token share: %v", err),
				Data:    nil,
			})
		}

		return c.JSON(&responses.Response{
			Status:  fiber.StatusOK,
			Message: "successfully calculated subpool token share.",
			Data:    &fiber.Map{"tokenShare": tokenShare},
		})
	})

	app.Get("/kos/check-if-staker-banned/:address", func(c *fiber.Ctx) error {
		// get the address param from the request query params
		addressParam := c.Params("address")

		// call the CheckIfStakerBanned function
		banned, err := ApiKOS.CheckIfStakerBanned(addressParam)
		if err != nil {
			return c.JSON(&responses.Response{
				Status:  fiber.StatusBadRequest,
				Message: fmt.Sprintf("unable to successfully check if staker banned: %v", err),
				Data:    nil,
			})
		}

		return c.JSON(&responses.Response{
			Status:  fiber.StatusOK,
			Message: "successfully checked if staker banned.",
			Data:    &fiber.Map{"banned": banned},
		})
	})

	app.Get("/kos/check-pool-time-allowance-exceeded/:stakingPoolId", func(c *fiber.Ctx) error {
		// get the stakingPoolId param from the request query params
		stakingPoolIdParam := c.Params("stakingPoolId")

		// convert the stakingPoolId param to an int
		stakingPoolId, err := strconv.Atoi(stakingPoolIdParam)
		if err != nil {
			return c.JSON(&responses.Response{
				Status:  fiber.StatusBadRequest,
				Message: fmt.Sprintf("unable to successfully convert given stakingPoolId to int: %v", err),
				Data:    nil,
			})
		}

		// call the CheckPoolTimeAllowanceExceeded function
		exceeded, err := ApiKOS.CheckPoolTimeAllowanceExceeded(stakingPoolId)
		if err != nil {
			return c.JSON(&responses.Response{
				Status:  fiber.StatusBadRequest,
				Message: fmt.Sprintf("unable to successfully check if pool time allowance exceeded: %v", err),
				Data:    nil,
			})
		}

		return c.JSON(&responses.Response{
			Status:  fiber.StatusOK,
			Message: "successfully checked if pool time allowance exceeded.",
			Data:    &fiber.Map{"exceeded": exceeded},
		})
	})

	app.Get("/kos/check-if-keys-staked", func(c *fiber.Ctx) error {
		// get the keyIds param from the request query params
		keyIdsParam := c.Query("keyIds")

		// convert the keyIds param to an array of ints
		keyIdsStr := strings.Split(keyIdsParam, ",")
		keyIds := make([]int, len(keyIdsStr))
		for i, id := range keyIdsStr {
			idInt, err := strconv.Atoi(id)
			if err != nil {
				return c.JSON(&responses.Response{
					Status:  fiber.StatusBadRequest,
					Message: fmt.Sprintf("unable to successfully convert given keyId to int: %v", err),
					Data:    nil,
				})
			}

			keyIds[i] = idInt
		}

		// get the stakingPoolId param from the request query params
		stakingPoolIdParam := c.Query("stakingPoolId")

		// convert the stakingPoolId param to an int
		stakingPoolId, err := strconv.Atoi(stakingPoolIdParam)
		if err != nil {
			return c.JSON(&responses.Response{
				Status:  fiber.StatusBadRequest,
				Message: fmt.Sprintf("unable to successfully convert given stakingPoolId to int: %v", err),
				Data:    nil,
			})
		}

		// call the CheckIfKeysStaked function
		keysStaked, err := ApiKOS.CheckIfKeysStaked(stakingPoolId, keyIds)
		if err != nil {
			return c.JSON(&responses.Response{
				Status:  fiber.StatusBadRequest,
				Message: fmt.Sprintf("unable to successfully check if keys staked: %v", err),
				Data:    nil,
			})
		}

		return c.JSON(&responses.Response{
			Status:  fiber.StatusOK,
			Message: "successfully checked if keys staked.",
			Data:    &fiber.Map{"keysStaked": keysStaked},
		})
	})

	app.Post("/kos/add-subpool", func(c *fiber.Ctx) error {
		type AddSubpoolRequest struct {
			KeyIds             []int  `json:"keyIds"`
			StakerWallet       string `json:"stakerWallet"`
			StakingPoolId      int    `json:"stakingPoolId"`
			KeychainId         int    `json:"keychainId"`
			SuperiorKeychainId int    `json:"superiorKeychainId"`
		}

		// parse the req body into the AddSubpoolRequest struct
		var addSubpoolRequest AddSubpoolRequest
		err := c.BodyParser(&addSubpoolRequest)
		if err != nil {
			return c.JSON(&responses.Response{
				Status:  fiber.StatusBadRequest,
				Message: fmt.Sprintf("unable to successfully parse request body: %v", err),
				Data:    nil,
			})
		}

		fmt.Printf("addSubpoolRequest: %+v\n", addSubpoolRequest)

		// call the AddSubpool fn
		err = ApiKOS.AddSubpool(addSubpoolRequest.KeyIds, addSubpoolRequest.StakerWallet, addSubpoolRequest.StakingPoolId, addSubpoolRequest.KeychainId, addSubpoolRequest.SuperiorKeychainId)
		if err != nil {
			return c.JSON(&responses.Response{
				Status:  fiber.StatusBadRequest,
				Message: fmt.Sprintf("unable to successfully add subpool: %v", err),
				Data:    nil,
			})
		}

		return c.JSON(&responses.Response{
			Status:  fiber.StatusOK,
			Message: "successfully added subpool.",
			Data:    nil,
		})
	})

	// UnstakeFromSubpool route
	app.Post("/kos/unstake-from-subpool", func(c *fiber.Ctx) error {
		type UnstakeFromSubpoolRequest struct {
			StakingPoolID int `json:"stakingPoolId"`
			SubpoolID     int `json:"subpoolId"`
		}

		// parse the req body into the UnstakeFromSubpoolRequest struct
		var unstakeFromSubpoolRequest UnstakeFromSubpoolRequest
		err := c.BodyParser(&unstakeFromSubpoolRequest)
		if err != nil {
			return c.JSON(&responses.Response{
				Status:  fiber.StatusBadRequest,
				Message: fmt.Sprintf("unable to successfully parse request body: %v", err),
				Data:    nil,
			})
		}

		fmt.Printf("unstakeFromSubpoolRequest: %+v\n", unstakeFromSubpoolRequest)

		// call the UnstakeFromSubpool fn
		err = ApiKOS.UnstakeFromSubpool(unstakeFromSubpoolRequest.StakingPoolID, unstakeFromSubpoolRequest.SubpoolID)
		if err != nil {
			return c.JSON(&responses.Response{
				Status:  fiber.StatusBadRequest,
				Message: fmt.Sprintf("unable to successfully unstake from subpool: %v", err),
				Data:    nil,
			})
		}

		return c.JSON(&responses.Response{
			Status:  fiber.StatusOK,
			Message: fmt.Sprintf("successfully unstaked from subpool %d of staking pool id %d", unstakeFromSubpoolRequest.SubpoolID, unstakeFromSubpoolRequest.StakingPoolID),
			Data:    nil,
		})
	})
}
