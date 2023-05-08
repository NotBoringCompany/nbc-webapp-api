package configs

import "github.com/gofiber/fiber/v2/middleware/cors"

/*
Returns a cors.Config instance to allow requests from webapp.nbcompany.io
*/
func CorsConfig() cors.Config {
	return cors.Config{
		AllowOrigins: "https://webapp.nbcompany.io",
		AllowMethods: "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders: "Origin,X-Requested-With,Content-Type,Accept,session-token",
	}
}
