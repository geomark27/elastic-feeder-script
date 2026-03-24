package processor

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"sync"
	"time"

	"elastic-feeder-script/internal/checkpoint"
	"elastic-feeder-script/internal/config"
	"elastic-feeder-script/internal/db"
	"elastic-feeder-script/internal/sharepoint"
)

func Run(database *sql.DB, client *sharepoint.Client, cfg config.Config, cp *checkpoint.Checkpoint) error {
	printHeader(cfg)

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	docs, err := db.FetchDocuments(ctx, database, cfg)
	if err != nil {
		return fmt.Errorf("obteniendo documentos: %w", err)
	}

	if len(docs) == 0 {
		log.Println("[INFO] No se encontraron documentos con los criterios especificados.")
		return nil
	}

	pending, skipped := filterPending(docs, cp)
	okPrev, _ := cp.Stats()

	log.Printf("[INFO] Total en BD           : %d", len(docs))
	log.Printf("[INFO] Ya procesados         : %d", okPrev)
	log.Printf("[INFO] Pendientes            : %d", len(pending))
	fmt.Println()

	if len(pending) == 0 {
		log.Println("[INFO] No hay documentos pendientes. Todos ya fueron procesados.")
		return nil
	}

	results := runConcurrent(client, pending, cp, cfg)

	printSummary(results, skipped, len(docs), cp)
	return nil
}

// filterPending separa los documentos pendientes de los ya procesados.
func filterPending(docs []db.Documento, cp *checkpoint.Checkpoint) ([]db.Documento, int) {
	pending := make([]db.Documento, 0, len(docs))
	skipped := 0
	for _, doc := range docs {
		if cp.IsDone(doc.DocumentoId) {
			skipped++
		} else {
			pending = append(pending, doc)
		}
	}
	return pending, skipped
}

// runConcurrent orquesta el procesamiento por rondas de workers concurrentes.
func runConcurrent(client *sharepoint.Client, pending []db.Documento, cp *checkpoint.Checkpoint, cfg config.Config) []sharepoint.Result {
	batches := splitBatches(pending, cfg.Limit)
	totalBatches := len(batches)
	totalRounds := (totalBatches + cfg.Workers - 1) / cfg.Workers

	log.Printf("[INFO] %d docs → %d lotes de %d | %d workers | %ds entre rondas",
		len(pending), totalBatches, cfg.Limit, cfg.Workers, cfg.Sleep)
	fmt.Println()

	var allResults []sharepoint.Result
	var mu sync.Mutex

	for roundIdx := 0; roundIdx < totalBatches; roundIdx += cfg.Workers {
		end := min(roundIdx+cfg.Workers, totalBatches)
		round := batches[roundIdx:end]
		roundNum := roundIdx/cfg.Workers + 1

		log.Printf("┌─ RONDA %d/%d — %d lote(s) corriendo en paralelo", roundNum, totalRounds, len(round))

		var wg sync.WaitGroup
		for i, batch := range round {
			wg.Add(1)
			batchNum := roundIdx + i + 1
			go func(num int, b []db.Documento) {
				defer wg.Done()
				results := processBatch(client, b, cp, num, totalBatches)
				mu.Lock()
				allResults = append(allResults, results...)
				mu.Unlock()
			}(batchNum, batch)
		}
		wg.Wait()

		log.Printf("└─ RONDA %d/%d completada", roundNum, totalRounds)
		fmt.Println()

		// Descanso entre rondas (no después de la última)
		if roundIdx+cfg.Workers < totalBatches {
			log.Printf("[INFO] Descansando %ds antes de la siguiente ronda...", cfg.Sleep)
			time.Sleep(time.Duration(cfg.Sleep) * time.Second)
			fmt.Println()
		}
	}

	return allResults
}

// processBatch procesa un lote de documentos de forma secuencial dentro del worker.
func processBatch(client *sharepoint.Client, docs []db.Documento, cp *checkpoint.Checkpoint, batchNum, totalBatches int) []sharepoint.Result {
	results := make([]sharepoint.Result, 0, len(docs))
	prefix := fmt.Sprintf("[Lote %d/%d]", batchNum, totalBatches)

	for i, doc := range docs {
		log.Printf("%s (%d/%d) → %s | OrdenEcho: %s | OrdenDobra: %s | Año: %s",
			prefix, i+1, len(docs), doc.DocumentoId, doc.OrdenEcho, doc.OrdenDobra, doc.AnoTramite)

		r := client.UpdateMetadata(doc)

		if r.Success {
			if err := cp.MarkOK(doc.DocumentoId, doc.OrdenEcho); err != nil {
				log.Printf("%s [WARN] No se pudo guardar checkpoint: %v", prefix, err)
			}
			log.Printf("%s (%d/%d) ✓ OK (HTTP %d)", prefix, i+1, len(docs), r.StatusCode)
		} else {
			if err := cp.MarkError(doc.DocumentoId, doc.OrdenEcho, r.Error); err != nil {
				log.Printf("%s [WARN] No se pudo guardar checkpoint: %v", prefix, err)
			}
			log.Printf("%s (%d/%d) ✗ ERROR: %s", prefix, i+1, len(docs), r.Error)
		}

		results = append(results, r)
	}

	return results
}

// splitBatches divide un slice de documentos en lotes de tamaño size.
func splitBatches(docs []db.Documento, size int) [][]db.Documento {
	batches := make([][]db.Documento, 0, (len(docs)+size-1)/size)
	for i := 0; i < len(docs); i += size {
		batches = append(batches, docs[i:min(i+size, len(docs))])
	}
	return batches
}

func printHeader(cfg config.Config) {
	log.Println("═══════════════════════════════════════════════")
	log.Println("   SharePoint Metadata Sync")
	log.Printf("   Docs por lote          : %d", cfg.Limit)
	log.Printf("   Workers concurrentes   : %d", cfg.Workers)
	log.Printf("   Descanso entre rondas  : %ds", cfg.Sleep)
	log.Printf("   BD                     : %s:%s/%s", cfg.DBHost, cfg.DBPort, cfg.DBName)
	log.Printf("   API                    : %s", cfg.APIBaseURL)
	log.Printf("   Checkpoint             : %s", cfg.CheckpointFile)
	log.Println("═══════════════════════════════════════════════")
	fmt.Println()
}

func printSummary(results []sharepoint.Result, skipped, total int, cp *checkpoint.Checkpoint) {
	ok, failed := 0, 0
	for _, r := range results {
		if r.Success {
			ok++
		} else {
			failed++
		}
	}

	totalOK, totalErr := cp.Stats()
	pendientes := total - totalOK - totalErr

	log.Println("═══════════════════════════════════════════════")
	log.Println("  RESUMEN DE EJECUCIÓN")
	log.Printf("  Procesados este run    : %d", len(results))
	log.Printf("    ✓ Exitosos           : %d", ok)
	log.Printf("    ✗ Fallidos           : %d", failed)
	log.Println("  ─────────────────────────────────────────────")
	log.Println("  PROGRESO ACUMULADO")
	log.Printf("    Total en BD          : %d", total)
	log.Printf("    ✓ Completados        : %d", totalOK)
	log.Printf("    ✗ Con error          : %d", totalErr)
	log.Printf("    ⏳ Pendientes         : %d", pendientes)
	log.Println("═══════════════════════════════════════════════")

	if failed > 0 {
		log.Println("  Documentos fallidos este run:")
		for _, r := range results {
			if !r.Success {
				log.Printf("    - %s (%s): %s", r.DocumentoId, r.OrdenEcho, r.Error)
			}
		}
	}
}
