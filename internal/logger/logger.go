package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"time"
)

func Setup(logDir string) (*os.File, error) {
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, fmt.Errorf("creando directorio de logs: %w", err)
	}

	filename := fmt.Sprintf("%s/sync_%s.log", logDir, time.Now().Format("2006-01-02_15-04-05"))
	f, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("abriendo archivo de log %s: %w", filename, err)
	}

	log.SetOutput(io.MultiWriter(os.Stdout, f))
	log.SetFlags(log.Ldate | log.Ltime)
	log.Printf("[INFO] Log guardado en: %s", filename)

	return f, nil
}
