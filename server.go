package main

import (
	"nbc-backend-api-v2/configs"
	RoutesNFTs "nbc-backend-api-v2/routes/nfts"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

func main() {
	// err := configs.LoadEnv()
	// if err != nil {
	// 	log.Fatal(err)
	// }

	app := fiber.New()

	app.Use(cors.New())

	// runs the ConnectMongo function
	configs.ConnectMongo()

	RoutesNFTs.KOSRoutes(app)

	app.Listen(":8080")
}
