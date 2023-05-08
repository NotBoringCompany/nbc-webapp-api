package main

import (
	ApiKOS "nbc-backend-api-v2/api/nfts/kos"
	"nbc-backend-api-v2/configs"
	RoutesNFTs "nbc-backend-api-v2/routes/nfts"
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

	// Allow requests from webapp.nbcompany.io
	app.Use(cors.New(configs.CorsConfig()))

	// runs the ConnectMongo function
	configs.ConnectMongo()

	RoutesNFTs.KOSRoutes(app)

	// SCHEDULERS
	ApiKOS.UpdateTotalYieldPointsScheduler().Start()
	ApiKOS.CloseSubpoolsOnStakeEndScheduler().Start()
	ApiKOS.VerifyStakerOwnershipScheduler().Start()
	ApiKOS.VerifyStakingPoolStakerCountScheduler().Start()
	ApiKOS.RemoveExpiredUnclaimableSubpoolsScheduler().Start()

	app.Listen(":" + port)
}
