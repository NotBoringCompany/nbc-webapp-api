package main

import (
	"fmt"
	"log"
	SuperiorKeychain "nbc-backend-api-v2/api/nfts/superior-keychain"
	"nbc-backend-api-v2/configs"

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
		data, err := SuperiorKeychain.OwnerIDs("0x8FbFE537A211d81F90774EE7002ff784E352024a")
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

	app.Listen("localhost:3000")
}
