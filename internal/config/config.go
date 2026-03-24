package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	DBHost         string
	DBPort         string
	DBName         string
	DBUser         string
	DBPass         string
	APIBaseURL     string
	UsuarioAccion  string
	Limit          int
	Workers        int
	Sleep          int
	CheckpointFile string
	LogDir         string
	FROM_DATE      string
	TO_DATE        string
}

func Load() Config {
	if err := godotenv.Load(); err != nil {
		log.Println("[WARN] No .env encontrado, usando variables del sistema")
	}

	port := getEnv("DB_PORT", "1433")
	if port == "" {
		port = "1433"
	}

	return Config{
		DBHost:         getEnv("DB_HOST", "localhost"),
		DBPort:         port,
		DBName:         getEnv("DB_DATABASE", ""),
		DBUser:         getEnv("DB_USERNAME", ""),
		DBPass:         getEnv("DB_PASSWORD", ""),
		APIBaseURL:     getEnv("API_BASE_URL", ""),
		UsuarioAccion:  getEnv("USUARIO_ACCION", "proceso-sync-metadata"),
		Limit:          getInt("LIMIT", 100),
		Workers:        getInt("WORKERS", 3),
		Sleep:          getInt("SLEEP", 10),
		CheckpointFile: getEnv("CHECKPOINT_FILE", "checkpoint.json"),
		LogDir:         getEnv("LOG_DIR", "logs"),
		FROM_DATE:      getEnv("FROM_DATE", "2026-01-08"),
		TO_DATE:        getEnv("TO_DATE", "2026-02-27"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getInt(key string, fallback int) int {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil || n <= 0 {
		return fallback
	}
	return n
}
