package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"elastic-feeder-script/internal/config"

	_ "github.com/microsoft/go-mssqldb"
)

type Documento struct {
	DocumentoId         string
	OrdenEcho           string
	AnoTramite          string
	OrdenDobra          string
	TipoDocumentoNombre string
	NombreOriginal      string
	NumeroItems         string
}

func Connect(cfg config.Config) (*sql.DB, error) {
	connStr := fmt.Sprintf(
		"sqlserver://%s:%s@%s:%s?database=%s&connection+timeout=30",
		cfg.DBUser, cfg.DBPass, cfg.DBHost, cfg.DBPort, cfg.DBName,
	)

	database, err := sql.Open("sqlserver", connStr)
	if err != nil {
		return nil, fmt.Errorf("abriendo conexión: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := database.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("ping a BD (%s:%s/%s): %w", cfg.DBHost, cfg.DBPort, cfg.DBName, err)
	}

	return database, nil
}

func FetchDocuments(ctx context.Context, database *sql.DB, cfg config.Config) ([]Documento, error) {
	query := `
SELECT
    d.path                                      AS DocumentoId,
    ISNULL(t.orden, '')                         AS ordenEcho,
    ISNULL(CAST(t.ano AS NVARCHAR(10)), '')     AS AnoTramite,
    ISNULL(tt.Orden, '')                        AS ordenDobra,
    ISNULL(td.nombre, '')                       AS TipoDocumentoNombre,
    ISNULL(d.nombreArchivoOriginal, '')         AS NombreOriginal,
    ISNULL(d.numItems, '')                      AS NumeroItems
FROM documentos d WITH(NOLOCK)
    INNER JOIN tramites t WITH(NOLOCK)
        ON t.id = d.documentoable_id
        AND d.documentoable_type = 'App\Models\Transaccionales\Tramite'
    INNER JOIN tyt..TRM_TRAMITES_VARIOS ttv WITH(NOLOCK)
        ON ttv.idTramiteOperativoTyT = t.id
    INNER JOIN tyt..trm_Tramites tt WITH(NOLOCK)
        ON tt.ID = ttv.TRAMITEID
    INNER JOIN tipos_documentos td WITH(NOLOCK)
        ON td.id = d.idTipoDocumento
WHERE d.created_at >= @fromDate AND d.created_at <= @toDate
ORDER BY t.created_at DESC`

	rows, err := database.QueryContext(ctx, query,
		sql.Named("fromDate", cfg.FROM_DATE),
		sql.Named("toDate", cfg.TO_DATE),
	)
	if err != nil {
		return nil, fmt.Errorf("ejecutando query: %w", err)
	}
	defer rows.Close()

	var docs []Documento
	for rows.Next() {
		var d Documento
		if err := rows.Scan(
			&d.DocumentoId,
			&d.OrdenEcho,
			&d.AnoTramite,
			&d.OrdenDobra,
			&d.TipoDocumentoNombre,
			&d.NombreOriginal,
			&d.NumeroItems,
		); err != nil {
			return nil, fmt.Errorf("escaneando fila: %w", err)
		}
		docs = append(docs, d)
	}

	return docs, rows.Err()
}
