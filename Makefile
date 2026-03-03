.PHONY: run build clean reset tidy vet

BINARY   := bin/sync
CMD_PATH := ./cmd/sync

run:
	go run $(CMD_PATH)

build:
	@mkdir -p bin
	go build -o $(BINARY) $(CMD_PATH)
	@echo "Binario generado: $(BINARY)"

build-all:
	@mkdir -p bin
	GOOS=linux   GOARCH=amd64 go build -o bin/sync-linux   $(CMD_PATH)
	GOOS=windows GOARCH=amd64 go build -o bin/sync-windows.exe $(CMD_PATH)
	@echo "Binarios generados:"
	@echo "  → bin/sync-linux"
	@echo "  → bin/sync-windows.exe"

clean:
	@rm -rf bin/
	@echo "Binarios eliminados."

reset:
	@rm -f checkpoint.json
	@echo "Checkpoint eliminado. El próximo run comenzará desde cero."

tidy:
	go mod tidy

vet:
	go vet ./...
