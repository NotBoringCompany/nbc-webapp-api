package main

import (
	ApiKOS "nbc-backend-api-v2/api/nfts/kos"
	"nbc-backend-api-v2/configs"
	RoutesNFTs "nbc-backend-api-v2/routes/nfts"
	UtilsKOS "nbc-backend-api-v2/utils/nfts/kos"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

func main() {
	// err := configs.LoadEnv()
	// if err != nil {
	// 	log.Fatal(err)
	// }
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	app := fiber.New()

	app.Use(cors.New())

	// runs the ConnectMongo function
	configs.ConnectMongo()

	RoutesNFTs.KOSRoutes(app)

	app.Get("/clearCache", func(c *fiber.Ctx) error {
		err := UtilsKOS.ClearCache()
		if err != nil {
			return c.SendString(err.Error())
		}

		return c.SendString("Success")
	})

	// app.Get("/addStakingPool", func(c *fiber.Ctx) error {
	// 	err := UtilsKOS.AddStakingPool(configs.GetCollections(configs.DB, "RHStakingPool"), "REC", 500000)
	// 	if err != nil {
	// 		return c.SendString(err.Error())
	// 	}

	// 	return c.SendString("Success")
	// })

	// SCHEDULERS
	ApiKOS.UpdateTotalYieldPointsScheduler().Start()
	ApiKOS.CloseSubpoolsOnStakeEndScheduler().Start()
	ApiKOS.VerifyStakerOwnershipScheduler().Start()

	app.Listen(":" + port)
}
