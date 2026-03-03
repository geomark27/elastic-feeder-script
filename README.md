# SharePoint Metadata Sync

Script para actualizar masivamente la metadata de documentos en SharePoint,
leyendo la información desde la base de datos operativa.

---

## ¿Necesito tener Go instalado?

**Depende de cómo recibas el proyecto:**

| Escenario | ¿Necesita Go? |
|---|---|
| Recibes el proyecto fuente (zip) y quieres usar `make run` | **Sí** |
| Recibes el proyecto fuente y quieres compilar el binario | **Sí** |
| Recibes directamente el binario `sync` compilado | **No** |

---

## Para equipos de infraestructura (sin Go)

Esta es la forma más sencilla de ejecutar el script. Solo necesitas
**2 archivos** que te entregará el desarrollador:

```
cualquier-carpeta/
├── sync-linux          ← si usas Linux
├── sync-windows.exe    ← si usas Windows
└── .env.example        ← plantilla de configuración (solo uno de los binarios)
```

### Paso 1 — Crear el archivo de configuración

```bash
cp .env.example .env
```

Edita `.env` con las credenciales del ambiente correspondiente:

```env
DB_HOST=<ip-servidor-bd>
DB_PORT=                    # dejar vacío para usar el default 1433
DB_DATABASE=<nombre-bd>
DB_USERNAME=<usuario>
DB_PASSWORD=<contraseña>

API_BASE_URL=https://sharepointapi.ejemplo.com/api/files
USUARIO_ACCION=proceso-sync-metadata

LIMIT=50                    # documentos a procesar por ejecución
CHECKPOINT_FILE=checkpoint.json
LOG_DIR=logs
```

> Las credenciales **nunca están dentro del binario** — el binario las lee
> del `.env` en el momento de ejecutarse. Cada ambiente tiene su propio `.env`.

### Paso 2 — Ejecutar

```bash
# Linux
./sync-linux

# Windows (desde cmd o PowerShell)
sync-windows.exe
```

### Paso 3 — Lo que se genera automáticamente

Después del primer run la carpeta queda así:

```
cualquier-carpeta/
├── sync
├── .env.example
├── .env                               ← creado por ti
├── checkpoint.json                    ← creado automáticamente
└── logs/
    ├── sync_2026-02-27_14-19-53.log
    ├── sync_2026-02-27_14-29-27.log
    └── sync_2026-02-27_16-26-27.log
```

### Operaciones disponibles

Sin Go ni Makefile, todo se maneja con comandos simples:

| Qué quieres hacer | Comando |
|---|---|
| Correr el siguiente batch | `./sync-linux` o `sync-windows.exe` |
| Ver el progreso acumulado | `cat checkpoint.json` |
| Empezar desde cero | `rm checkpoint.json` |
| Ver un log específico | `cat logs/sync_2026-02-27_14-19-53.log` |
| Ver el log más reciente | `cat logs/$(ls logs/ | tail -1)` |

### Historial de ejecuciones (logs)

Cada ejecución genera su propio archivo de log con timestamp:

```
logs/
├── sync_2026-02-27_14-19-53.log   ← Run 1
├── sync_2026-02-27_14-29-27.log   ← Run 2
├── sync_2026-02-27_16-26-27.log   ← Run 3
└── sync_2026-02-27_16-29-53.log   ← Run 4
```

Cada archivo contiene el detalle completo: qué se procesó, qué falló
y el resumen final. Se abren con cualquier editor de texto.

---

## Para desarrolladores (con Go instalado)

### Requisitos
- [Go 1.23+](https://go.dev/dl/)
- `make` (en Linux/Mac viene por defecto; en Windows usar Git Bash o WSL)
- Acceso de red a la BD y a la API de SharePoint

### Configuración

```bash
cp .env.example .env
# editar .env con las credenciales del ambiente
```

### Comandos disponibles

```bash
make run      # ejecutar el script directamente
make build    # compilar el binario en bin/sync
make reset    # borrar el checkpoint (empezar desde cero)
make clean    # eliminar el binario compilado
make tidy     # actualizar dependencias
make vet      # verificar el código
```

### Compilar para otro sistema operativo

El binario es específico del SO donde se compiló. Para generar
el ejecutable destinado a otro ambiente:

```bash
# Para Linux desde cualquier OS
GOOS=linux GOARCH=amd64 go build -o bin/sync ./cmd/sync

# Para Windows desde cualquier OS
GOOS=windows GOARCH=amd64 go build -o bin/sync.exe ./cmd/sync
```

---

## ¿Cómo funciona el proceso?

```
1. Lee la BD y obtiene todos los documentos desde la fecha configurada.
2. Por cada documento revisa si ya fue procesado (checkpoint).
   → Si ya fue procesado: lo omite silenciosamente.
   → Si es nuevo: llama al endpoint PATCH de SharePoint con la metadata.
3. Guarda el resultado en checkpoint.json después de cada documento.
4. Al finalizar muestra un resumen con completados, fallidos y pendientes.
```

### Ejemplo de flujo con 125 documentos y LIMIT=50

```
Run 1 → procesa   1– 50  | Completados:  50 | Pendientes: 75
Run 2 → procesa  51–100  | Completados: 100 | Pendientes: 25
Run 3 → procesa 101–125  | Completados: 125 | Pendientes:  0
```

### Recuperación ante fallos

Si el proceso se interrumpe (error de red, corte de luz, etc.),
al volver a ejecutar el script retoma automáticamente desde el último
documento pendiente. Los ya marcados como `ok` se omiten.

Para empezar desde cero:

```bash
rm checkpoint.json          # sin Go
make reset                  # con Go
```

---

## Estructura del proyecto

```
elastic-feeder-script/
├── cmd/sync/main.go          ← entrada del programa
├── internal/
│   ├── config/config.go      ← carga de variables de entorno
│   ├── db/db.go              ← conexión a BD y query
│   ├── sharepoint/client.go  ← cliente HTTP para la API
│   ├── checkpoint/           ← control de progreso
│   └── processor/            ← orquestación del proceso
├── .env.example              ← plantilla de configuración
├── Makefile
└── README.md
```
