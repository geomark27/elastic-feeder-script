package main

import (
	"log"

	"elastic-feeder-script/internal/checkpoint"
	"elastic-feeder-script/internal/config"
	"elastic-feeder-script/internal/db"
	"elastic-feeder-script/internal/logger"
	"elastic-feeder-script/internal/processor"
	"elastic-feeder-script/internal/sharepoint"
)

func main() {
	cfg := config.Load()

	logFile, err := logger.Setup(cfg.LogDir)
	if err != nil {
		log.Fatalf("[FATAL] %v", err)
	}
	defer logFile.Close()

	cp, err := checkpoint.Load(cfg.CheckpointFile)
	if err != nil {
		log.Fatalf("[FATAL] Cargando checkpoint: %v", err)
	}

	database, err := db.Connect(cfg)
	if err != nil {
		log.Fatalf("[FATAL] %v", err)
	}
	defer database.Close()

	client := sharepoint.NewClient(cfg)

	if err := processor.Run(database, client, cfg, cp); err != nil {
		log.Fatalf("[FATAL] %v", err)
	}
}
