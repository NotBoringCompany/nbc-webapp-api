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

		res := ApiKOS.FetchMetadata(tokenIdInt)

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

	// app.Get("/kos/calculate-subpool-points", func(c *fiber.Ctx) error {
	// 	// get the keyIds param from the request query params
	// 	keyIdsParam := c.Query("keyIds")

	// 	// convert the keyIds param to an array of ints
	// 	keyIdsStr := strings.Split(keyIdsParam, ",")
	// 	keyIds := make([]int, len(keyIdsStr))
	// 	for i, id := range keyIdsStr {
	// 		idInt, err := strconv.Atoi(id)
	// 		if err != nil {
	// 			return c.JSON(&responses.Response{
	// 				Status:  fiber.StatusBadRequest,
	// 				Message: fmt.Sprintf("unable to successfully convert given keyId to int: %v", err),
	// 				Data:    nil,
	// 			})
	// 		}

	// 		keyIds[i] = idInt
	// 	}

	// 	// get the keychainId and superiorKeychainId params from the request query params
	// 	keychainIdParam := c.Query("keychainId")
	// 	superiorKeychainIdParam := c.Query("superiorKeychainId")

	// 	// convert the keychainId and superiorKeychainId params to ints
	// 	keychainId, err := strconv.Atoi(keychainIdParam)
	// 	if err != nil {
	// 		return c.JSON(&responses.Response{
	// 			Status:  fiber.StatusBadRequest,
	// 			Message: fmt.Sprintf("unable to successfully convert given keychainId to int: %v", err),
	// 			Data:    nil,
	// 		})
	// 	}

	// 	superiorKeychainId, err := strconv.Atoi(superiorKeychainIdParam)
	// 	if err != nil {
	// 		return c.JSON(&responses.Response{
	// 			Status:  fiber.StatusBadRequest,
	// 			Message: fmt.Sprintf("unable to successfully convert given superiorKeychainId to int: %v", err),
	// 			Data:    nil,
	// 		})
	// 	}

	// 	// call the CalculateSubpoolPoints function
	// 	points := ApiKOS.CalculateSubpoolPoints(keyIds, keychainId, superiorKeychainId)

	// 	return c.JSON(&responses.Response{
	// 		Status:  fiber.StatusOK,
	// 		Message: "successfully calculated subpool points.",
	// 		Data:    &fiber.Map{"points": points},
	// 	})
	// })

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

	app.Get("/check-if-staker-banned/:address", func(c *fiber.Ctx) error {
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

	app.Get("/check-pool-time-allowance-exceeded/:stakingPoolId", func(c *fiber.Ctx) error {
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

	app.Get("/check-if-keys-staked", func(c *fiber.Ctx) error {
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

	app.Post("/add-subpool", func(c *fiber.Ctx) error {
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
}