package sharepoint

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"elastic-feeder-script/internal/config"
	"elastic-feeder-script/internal/db"
)

type Client struct {
	http          *http.Client
	baseURL       string
	usuarioAccion string
}

type Result struct {
	DocumentoId string
	OrdenEcho   string
	Success     bool
	StatusCode  int
	Response    string
	Error       string
}

type metadataPayload struct {
	Metadata      map[string]string `json:"metadata"`
	UsuarioAccion string            `json:"usuarioAccion"`
}

func NewClient(cfg config.Config) *Client {
	return &Client{
		http:          &http.Client{Timeout: 20 * time.Second},
		baseURL:       cfg.APIBaseURL,
		usuarioAccion: cfg.UsuarioAccion,
	}
}

func (c *Client) UpdateMetadata(doc db.Documento) Result {
	result := Result{
		DocumentoId: doc.DocumentoId,
		OrdenEcho:   doc.OrdenEcho,
	}

	body, err := json.Marshal(metadataPayload{
		Metadata: map[string]string{
			"ordenEcho":             doc.OrdenEcho,
			"ordenDobra":            doc.OrdenDobra,
			"nombreArchivoOriginal": doc.NombreOriginal,
			"tipoDocumento":         doc.TipoDocumentoNombre,
			"anio":                  doc.AnoTramite,
			"numItems":              doc.NumeroItems,
		},
		UsuarioAccion: c.usuarioAccion,
	})
	if err != nil {
		result.Error = fmt.Sprintf("serializando payload: %v", err)
		return result
	}

	url := fmt.Sprintf("%s/%s/metadata", c.baseURL, doc.DocumentoId)
	req, err := http.NewRequest(http.MethodPatch, url, bytes.NewReader(body))
	if err != nil {
		result.Error = fmt.Sprintf("creando request: %v", err)
		return result
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "text/plain")

	resp, err := c.http.Do(req)
	if err != nil {
		result.Error = fmt.Sprintf("ejecutando HTTP request: %v", err)
		return result
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	result.StatusCode = resp.StatusCode
	result.Response = string(respBody)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		result.Error = fmt.Sprintf("HTTP %d → %s", resp.StatusCode, string(respBody))
		return result
	}

	result.Success = true
	return result
}
