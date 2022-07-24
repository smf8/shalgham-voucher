package main

import (
	"arvan-challenge/config"
	"arvan-challenge/database"
	"arvan-challenge/handler"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/sirupsen/logrus"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	cfg := config.New()

	app := fiber.New()

	app.Use(cors.New())

	if cfg.Debug {
		app.Use(logger.New())
	}

	api := app.Group("/api")
	api.Get("/healthz", handler.CheckHealth)

	go func() {
		if err := app.Listen(cfg.Port); err != nil {
			logrus.Fatalf("http server failed: %s", err.Error())
		}
	}()

	_, err := database.NewConnection(cfg.Database)
	if err != nil {
		logrus.Fatalf("database failed: %s", err.Error())
	}

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)

	s := <-sig
	logrus.Infof("signal %s received\n", s)

	if err = app.Shutdown(); err != nil {
		logrus.Errorf("failed to shutdown server: %s", err.Error())
	}
}
