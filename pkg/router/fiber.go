package router

import (
	"github.com/ansrivas/fiberprometheus/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

type ServerConfig struct {
	Port      string `koanf:"port"`
	Debug     bool   `koanf:"debug"`
	NameSpace string `koanf:"name_space"`
}

func New(cfg ServerConfig) fiber.Router {
	app := fiber.New()

	app.Use(cors.New())

	if cfg.Debug {
		app.Use(logger.New())
	}

	api := app.Group("/api")

	prometheus := fiberprometheus.New(cfg.NameSpace)
	prometheus.RegisterAt(app, "/metrics")
	app.Use(prometheus.Middleware)

	return api
}
