package main

import (
	"fmt"
	"log"
	NftAPI "nbc-backend-api-v2/api/nfts"
	"nbc-backend-api-v2/configs"
	"nbc-backend-api-v2/models"
	UtilsKOS "nbc-backend-api-v2/utils/nfts/kos"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

func main() {
	err := configs.LoadEnv()
	if err != nil {
		log.Fatal(err)
	}

	app := fiber.New()

	app.Use(cors.New())

	// runs the ConnectMongo function
	configs.ConnectMongo()

	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello, Worlddd!")
	})

	// scheduler := gocron.NewScheduler(time.Local)
	// scheduler.Every(5).Seconds().Do(kos.VerifyOwnership, "0xb3bf8cd8Ba8BD013F4C318ED3C75C3f154a502fA", []*big.Int{big.NewInt(128), big.NewInt(2266)})
	// scheduler.StartAsync()

	app.Get("/test", func(c *fiber.Ctx) error {
		data, err := NftAPI.KOSOwnerIDs("0xb3bf8cd8Ba8BD013F4C318ED3C75C3f154a502fA")
		if err != nil {
			return c.SendString(err.Error())
		}

		return c.SendString(fmt.Sprintf("%v", data))
	})

	app.Get("/testDB", func(c *fiber.Ctx) error {
		data := configs.GetCollections(configs.DB, "RHDiscordAllianceData")
		fmt.Printf("Type of data is %T", *data)
		return c.SendString(fmt.Sprintf("%v", *data))
	})

	app.Get("/testFetch", func(c *fiber.Ctx) error {
		data := UtilsKOS.FetchSimplifiedMetadata(1)
		return c.SendString(fmt.Sprintf("%v", *data))
	})

	app.Get("/stakePoolID", func(c *fiber.Ctx) error {
		id, err := UtilsKOS.GetNextStakingPoolID(configs.GetCollections(configs.DB, "RHStakingPool"))
		if err != nil {
			return c.SendString(err.Error())
		}

		return c.JSON(id)
	})

	app.Get("/verifyOwnership", func(c *fiber.Ctx) error {
		ids := []int{27, 703, 1283}
		verified, err := api.VerifyKOSOwnership("0xb3bf8cd8Ba8BD013F4C318ED3C75C3f154a502fA", ids)
		if err != nil {
			return c.SendString(err.Error())
		}

		return c.JSON(verified)
	})

	app.Get("/testAddStakingPool", func(c *fiber.Ctx) error {
		err := UtilsKOS.AddStakingPool(configs.GetCollections(configs.DB, "RHStakingPool"), "REC", 500000)

		if err != nil {
			return c.SendString(err.Error())
		}

		return c.SendString("Success")
	})

	app.Get("/testAddSubpool", func(c *fiber.Ctx) error {
		metadata1 := &models.KOSSimplifiedMetadata{
			TokenID:        5,
			HouseTrait:     "Glory",
			TypeTrait:      "Electric",
			LuckTrait:      55,
			LuckBoostTrait: 1,
		}

		metadata2 := &models.KOSSimplifiedMetadata{
			TokenID:        6,
			HouseTrait:     "Chaos",
			TypeTrait:      "Electric",
			LuckTrait:      23,
			LuckBoostTrait: 1.2,
		}

		arr := []*models.KOSSimplifiedMetadata{metadata1, metadata2}
		err := UtilsKOS.AddSubpool(configs.GetCollections(configs.DB, "RHStakingPool"), 2, "0x8FbFE537A211d81F90774EE7002ff784E352024a", arr, -1, 5)
		if err != nil {
			return c.SendString(err.Error())
		}

		return c.SendString("Success")
	})

	app.Get("/getAllStakedKeyIDs", func(c *fiber.Ctx) error {
		keyIds, err := UtilsKOS.GetAllStakedKeyIDs(configs.GetCollections(configs.DB, "RHStakingPool"), 2)
		if err != nil {
			return c.SendString(err.Error())
		}

		return c.JSON(keyIds)
	})

	app.Get("/checkSuperiorKeychainStaked", func(c *fiber.Ctx) error {
		staked, err := UtilsKOS.CheckIfSuperiorKeychainStaked(configs.GetCollections(configs.DB, "RHStakingPool"), 2, 3)
		if err != nil {
			return c.SendString(err.Error())
		}

		return c.JSON(staked)
	})

	app.Get("/getAllStakedSuperiorKeychainIDs", func(c *fiber.Ctx) error {
		keychainIds, err := UtilsKOS.GetAllStakedSuperiorKeychainIDs(configs.GetCollections(configs.DB, "RHStakingPool"), 2)
		if err != nil {
			return c.SendString(err.Error())
		}

		return c.JSON(keychainIds)
	})

	app.Get("/getHighestSubpoolID", func(c *fiber.Ctx) error {
		id, err := UtilsKOS.GetNextSubpoolID(configs.GetCollections(configs.DB, "RHStakingPool"), 1)
		if err != nil {
			return c.SendString(err.Error())
		}

		return c.JSON(id)
	})

	app.Get("/getAllStakedKeyIDs", func(c *fiber.Ctx) error {
		keyIds, err := UtilsKOS.GetAllStakedKeyIDs(configs.GetCollections(configs.DB, "RHStakingPool"), 1)
		if err != nil {
			return c.SendString(err.Error())
		}

		return c.JSON(keyIds)
	})

	app.Get("/getAllStakedKeychainIDs", func(c *fiber.Ctx) error {
		keychainIds, err := UtilsKOS.GetAllStakedKeychainIDs(configs.GetCollections(configs.DB, "RHStakingPool"), 1)
		if err != nil {
			return c.SendString(err.Error())
		}

		return c.JSON(keychainIds)
	})

	app.Get("/getStakedSuperiorKeychainIds", func(c *fiber.Ctx) error {
		keychainIds, err := UtilsKOS.GetAllStakedSuperiorKeychainIDs(configs.GetCollections(configs.DB, "RHStakingPool"), 1)
		if err != nil {
			return c.SendString(err.Error())
		}

		return c.JSON(keychainIds)
	})

	app.Get("/timeExceeded", func(c *fiber.Ctx) error {
		exceeded, err := UtilsKOS.CheckPoolTimeAllowanceExceeded(configs.GetCollections(configs.DB, "RHStakingPool"), 1)
		if err != nil {
			return c.SendString(err.Error())
		}

		return c.JSON(exceeded)
	})

	app.Get("/testSubpoolPointsCalc", func(c *fiber.Ctx) error {
		points, err := UtilsKOS.GetAccSubpoolPoints(configs.GetCollections(configs.DB, "RHStakingPool"), 1, 1)
		if err != nil {
			return c.SendString(err.Error())
		}

		return c.JSON(points)
	})

	app.Get("/totalSubpoolPoints", func(c *fiber.Ctx) error {
		points, err := UtilsKOS.GetTotalSubpoolPoints(configs.GetCollections(configs.DB, "RHStakingPool"), 1)
		if err != nil {
			return c.SendString(err.Error())
		}

		return c.JSON(points)
	})

	app.Get("/tokenShare", func(c *fiber.Ctx) error {
		share, err := UtilsKOS.CalcSubpoolTokenShare(configs.GetCollections(configs.DB, "RHStakingPool"), 1, 1)
		if err != nil {
			return c.SendString(err.Error())
		}

		return c.JSON(share)
	})

	app.Get("/closeSubpools", func(c *fiber.Ctx) error {
		err := UtilsKOS.CloseSubpoolsOnStakeEnd(configs.GetCollections(configs.DB, "RHStakingPool"))
		if err != nil {
			return c.SendString(err.Error())
		}

		return c.SendString("Success")
	})

	app.Get("/claimReward", func(c *fiber.Ctx) error {
		err := UtilsKOS.ClaimReward(configs.GetCollections(configs.DB, "RHStakingPool"), "0x8FbFE537A211d81F90774EE7002ff784E352024a", 1, 3)
		if err != nil {
			return c.SendString(err.Error())
		}

		return c.SendString("Success")
	})

	// app.Get("/banSubpool", func(c *fiber.Ctx) error {
	// 	err := UtilsKOS.BanSubpool(configs.GetCollections(configs.DB, "RHStakingPool"), )
	// 	if err != nil {
	// 		return c.SendString(err.Error())
	// 	}

	// 	return c.SendString("Success")
	// })

	app.Get("/allActiveSubpools", func(c *fiber.Ctx) error {
		subpools, err := UtilsKOS.GetAllActiveSubpools(configs.GetCollections(configs.DB, "RHStakingPool"))
		if err != nil {
			return c.SendString(err.Error())
		}

		return c.JSON(subpools)
	})

	app.Get("/verifyStakerOwnership", func(c *fiber.Ctx) error {
		err := UtilsKOS.VerifyStakerOwnership(configs.GetCollections(configs.DB, "RHStakingPool"))
		if err != nil {
			return c.SendString(err.Error())
		}

		return c.JSON("Success")
	})

	app.Get("/checkIfBanned", func(c *fiber.Ctx) error {
		banned, err := UtilsKOS.CheckIfStakerBanned(configs.GetCollections(configs.DB, "RHStakerData"), "0x8FbFE537A211d81F90774EE7002ff784E352024a")
		if err != nil {
			return c.SendString(err.Error())
		}

		return c.JSON(banned)
	})

	app.Get("/unstakeFromPool", func(c *fiber.Ctx) error {
		err := UtilsKOS.UnstakeFromSubpool(configs.GetCollections(configs.DB, "RHStakingPool"), 2, 2)
		if err != nil {
			return c.SendString(err.Error())
		}

		return c.JSON("Success")
	})

	app.Get("/unstakeFromStakingPool", func(c *fiber.Ctx) error {
		err := UtilsKOS.UnstakeFromStakingPool(configs.GetCollections(configs.DB, "RHStakingPool"), 2, "0x8FbFE537A211d81F90774EE7002ff784E352024a")
		if err != nil {
			return c.SendString(err.Error())
		}

		return c.JSON("Success")
	})

	// app.Get("/testAddStaker", func(c *fiber.Ctx) error {
	// 	err := UtilsKOS.AddStaker(configs.GetCollections(configs.DB, "RHStakerData"), "0xbc01Db6ea15c344529159F9c9D8eAb37C130a3bE")
	// 	if err != nil {
	// 		return c.SendString(err.Error())
	// 	}

	// 	return c.SendString("Success")
	// })

	app.Listen("localhost:3000")
}
