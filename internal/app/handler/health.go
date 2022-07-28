package handler

import (
	"net/http"

	"github.com/gofiber/fiber/v2"
)

func CheckHealth(c *fiber.Ctx) error {
	return c.SendStatus(http.StatusNoContent)
}
