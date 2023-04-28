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
		UtilsKOS.ClearCache()
		return c.SendString("Cache cleared!")
	})

	app.Get("/getCache", func(c *fiber.Ctx) error {
		return c.JSON(UtilsKOS.GetCache())
	})

	// SCHEDULERS
	ApiKOS.UpdateTotalYieldPointsScheduler().Start()
	ApiKOS.CloseSubpoolsOnStakeEndScheduler().Start()
	ApiKOS.VerifyStakerOwnershipScheduler().Start()

	app.Listen(":" + port)
}
