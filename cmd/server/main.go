package main

import (
	"log"
	"net/http"

	"github.com/escaleseo/buran/cmd/server/handler"
	"github.com/escaleseo/buran/internal/platform/env"
	"github.com/escaleseo/buran/internal/platform/logger"
)

func main() {
	config := env.GetConfig()
	if config.BackendURL == "" {
		log.Fatal("missing required BACKEND_URL")
	}

	log := logger.Get()
	h, err := handler.New(config, log)
	if err != nil {
		log.WithError(err).Fatal(err.Error())
	}

	log.Info("listening on port ", config.Port)
	if err := http.ListenAndServe(":"+config.Port, h); err != nil {
		log.WithError(err).Error("server quit")
	}
}
