package checkpoint

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"
)

type status string

const (
	statusOK    status = "ok"
	statusError status = "error"
)

type Entry struct {
	Status    status `json:"status"`
	OrdenEcho string `json:"orden_echo"`
	UpdatedAt string `json:"updated_at"`
	Error     string `json:"error,omitempty"`
}

type Checkpoint struct {
	CreatedAt string            `json:"created_at"`
	UpdatedAt string            `json:"updated_at"`
	Documents map[string]*Entry `json:"documents"`

	mu   sync.Mutex
	path string
}

func Load(path string) (*Checkpoint, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &Checkpoint{
				CreatedAt: now(),
				Documents: make(map[string]*Entry),
				path:      path,
			}, nil
		}
		return nil, fmt.Errorf("leyendo checkpoint: %w", err)
	}

	var cp Checkpoint
	if err := json.Unmarshal(data, &cp); err != nil {
		return nil, fmt.Errorf("parseando checkpoint: %w", err)
	}
	cp.path = path
	return &cp, nil
}

func (cp *Checkpoint) IsDone(documentId string) bool {
	cp.mu.Lock()
	defer cp.mu.Unlock()
	e, ok := cp.Documents[documentId]
	return ok && e.Status == statusOK
}

func (cp *Checkpoint) MarkOK(documentId, ordenEcho string) error {
	cp.mu.Lock()
	defer cp.mu.Unlock()
	cp.Documents[documentId] = &Entry{
		Status:    statusOK,
		OrdenEcho: ordenEcho,
		UpdatedAt: now(),
	}
	cp.UpdatedAt = now()
	return cp.save()
}

func (cp *Checkpoint) MarkError(documentId, ordenEcho, errMsg string) error {
	cp.mu.Lock()
	defer cp.mu.Unlock()
	cp.Documents[documentId] = &Entry{
		Status:    statusError,
		OrdenEcho: ordenEcho,
		UpdatedAt: now(),
		Error:     errMsg,
	}
	cp.UpdatedAt = now()
	return cp.save()
}

func (cp *Checkpoint) Stats() (ok, errors int) {
	cp.mu.Lock()
	defer cp.mu.Unlock()
	for _, e := range cp.Documents {
		switch e.Status {
		case statusOK:
			ok++
		case statusError:
			errors++
		}
	}
	return
}

// save escribe atómicamente: primero a .tmp y luego renombra.
func (cp *Checkpoint) save() error {
	data, err := json.MarshalIndent(cp, "", "  ")
	if err != nil {
		return fmt.Errorf("serializando checkpoint: %w", err)
	}
	tmp := cp.path + ".tmp"
	if err := os.WriteFile(tmp, data, 0644); err != nil {
		return fmt.Errorf("escribiendo tmp: %w", err)
	}
	return os.Rename(tmp, cp.path)
}

func now() string {
	return time.Now().Format(time.RFC3339)
}
