package main

import (
	"fmt"
	"nbc-backend-api-v2/api/kos"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

func main() {
	app := fiber.New()

	app.Use(cors.New())

	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello, Worlddd!")
	})

	app.Get("/test", func(c *fiber.Ctx) error {
		holder, err := kos.GetHolders()

		if err != nil {
			return c.SendString(err.Error())
		}
		return c.SendString(fmt.Sprintf("%d", *holder))
	})

	app.Listen("localhost:3000")
}
