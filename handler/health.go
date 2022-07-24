package handler

import (
	"github.com/gofiber/fiber/v2"
	"net/http"
)

func CheckHealth(c *fiber.Ctx) error {
	return c.SendStatus(http.StatusNoContent)
}
