package responses

import "github.com/gofiber/fiber/v2"

/*
`Response` is the response struct for all API responses.
*/
type Response struct {
	Status  int        `json:"status"`
	Message string     `json:"message"`
	Data    *fiber.Map `json:"data"`
}
