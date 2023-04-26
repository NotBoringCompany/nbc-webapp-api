package main

import (
	"log"
	"nbc-backend-api-v2/configs"
	RoutesNFTs "nbc-backend-api-v2/routes/nfts"
	"os"

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
	configs.ConnectMongo(os.Getenv("MONGODB_URI"))

	RoutesNFTs.KOSRoutes(app)

	app.Listen(":3000")
}
