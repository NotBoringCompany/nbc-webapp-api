package main

import (
	"fmt"
	"log"
	"nbc-backend-api-v2/api/kos"
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

	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello, Worlddd!")
	})

	app.Get("/test", func(c *fiber.Ctx) error {
		holderData, err := kos.GetExplicitOwnerships()

		if err != nil {
			return c.SendString(err.Error())
		}
		return c.SendString(fmt.Sprintf("%v", holderData))
	})

	app.Listen("localhost:3000")
}
