package router

import (
	"github.com/ansrivas/fiberprometheus/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"time"
)

type ServerConfig struct {
	Port         string        `koanf:"port"`
	Debug        bool          `koanf:"debug"`
	NameSpace    string        `koanf:"name_space"`
	ReadTimeout  time.Duration `koanf:"read_timeout"`
	WriteTimeout time.Duration `koanf:"write_timeout"`
	IdleTimeout  time.Duration `koanf:"idle_timeout"`
}

func New(cfg ServerConfig) *fiber.App {
	app := fiber.New(fiber.Config{
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		IdleTimeout:  cfg.IdleTimeout,
	})

	app.Use(cors.New())

	if cfg.Debug {
		app.Use(logger.New())
	}

	prometheus := fiberprometheus.New(cfg.NameSpace)
	prometheus.RegisterAt(app, "/metrics")
	app.Use(prometheus.Middleware)

	return app
}
