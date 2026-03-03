# Manual de Ejecución — Equipo de Infraestructura

Este documento explica paso a paso cómo configurar y ejecutar el script
de sincronización de metadata de documentos en SharePoint.

---

## Lo que recibirás

El desarrollador te entregará los siguientes archivos:

```
sync-linux          ← ejecutable para Linux
sync-windows.exe    ← ejecutable para Windows
.env.example        ← plantilla de configuración
```

> No necesitas instalar ningún programa adicional. Solo configurar el `.env`
> y ejecutar el binario correspondiente a tu sistema operativo.

---

## Paso 1 — Crea tu archivo de configuración

Copia la plantilla:

```bash
# Linux
cp .env.example .env

# Windows (cmd)
copy .env.example .env
```

Abre el archivo `.env` con cualquier editor de texto y completa los valores:

```env
DB_CONNECTION=sqlsrv
DB_HOST=           ← IP o hostname del servidor de base de datos
DB_PORT=           ← dejar vacío para usar el puerto por defecto (1433)
DB_DATABASE=       ← nombre de la base de datos
DB_USERNAME=       ← usuario de base de datos
DB_PASSWORD=       ← contraseña de base de datos

API_BASE_URL=      ← URL base de la API de SharePoint
USUARIO_ACCION=    ← usuario que aparecerá en el registro de auditoría

LIMIT=100          ← documentos por lote
WORKERS=3          ← lotes corriendo en paralelo al mismo tiempo
SLEEP=10           ← segundos de descanso entre rondas
CHECKPOINT_FILE=checkpoint.json
LOG_DIR=logs
```

### ¿Qué significan LIMIT, WORKERS y SLEEP?

El script procesa los documentos en **rondas**. En cada ronda corren varios
lotes al mismo tiempo (workers). Al terminar la ronda descansa antes de
iniciar la siguiente.

```
Ejemplo: 900 docs | LIMIT=100 | WORKERS=3 | SLEEP=10

→ 9 lotes de 100 docs
→ 3 rondas de 3 workers cada una

Ronda 1: [Lote 1 - 100 docs] [Lote 2 - 100 docs] [Lote 3 - 100 docs]  ← corren juntos
         ↓ descanso 10s
Ronda 2: [Lote 4 - 100 docs] [Lote 5 - 100 docs] [Lote 6 - 100 docs]
         ↓ descanso 10s
Ronda 3: [Lote 7 - 100 docs] [Lote 8 - 100 docs] [Lote 9 - 100 docs]
         ↓ resumen final
```

| Parámetro | Qué controla | Recomendación |
|---|---|---|
| `LIMIT` | Cuántos documentos procesa cada lote | Entre 50 y 200 |
| `WORKERS` | Cuántos lotes corren al mismo tiempo | Entre 2 y 5 |
| `SLEEP` | Segundos de descanso entre rondas | Entre 5 y 15 |

> **Importante:** el archivo `.env` contiene credenciales sensibles.
> No lo compartas ni lo subas a ningún repositorio.

---

## Paso 2 — Ejecuta el script

```bash
# Linux
./sync-linux

# Windows (cmd o PowerShell)
sync-windows.exe
```

Al iniciar verás algo como esto:

```
2026-02-27 14:19:53 ═══════════════════════════════════════════
2026-02-27 14:19:53    SharePoint Metadata Sync
2026-02-27 14:19:53    Docs por lote          : 100
2026-02-27 14:19:53    Workers concurrentes   : 3
2026-02-27 14:19:53    Descanso entre rondas  : 10s
2026-02-27 14:19:53    BD                     : 192.168.1.10:1433/OperativoTyT
2026-02-27 14:19:53    API                    : https://sharepointapi.ejemplo.com/api/files
2026-02-27 14:19:53    Checkpoint             : checkpoint.json
2026-02-27 14:19:53 ═══════════════════════════════════════════

2026-02-27 14:19:55 [INFO] Total en BD           : 2500
2026-02-27 14:19:55 [INFO] Ya procesados         : 0
2026-02-27 14:19:55 [INFO] Pendientes            : 2500

2026-02-27 14:19:55 [INFO] 2500 docs → 25 lotes de 100 | 3 workers | 10s entre rondas

2026-02-27 14:19:55 ┌─ RONDA 1/9 — 3 lote(s) corriendo en paralelo
2026-02-27 14:19:55 [Lote 1/25] (1/100) → ed867ce5-... | OrdenEcho: TIQ-11 | ...
2026-02-27 14:19:55 [Lote 2/25] (1/100) → 86727eaa-... | OrdenEcho: TIQ-12 | ...
2026-02-27 14:19:55 [Lote 3/25] (1/100) → 69cf7971-... | OrdenEcho: TIM-9  | ...
...
2026-02-27 14:20:10 └─ RONDA 1/9 completada

2026-02-27 14:20:10 [INFO] Descansando 10s antes de la siguiente ronda...

2026-02-27 14:20:20 ┌─ RONDA 2/9 — 3 lote(s) corriendo en paralelo
...
```

---

## Paso 3 — Revisa el resumen al finalizar

Al terminar verás un resumen completo:

```
2026-02-27 14:30:00 ═══════════════════════════════════════════
2026-02-27 14:30:00   RESUMEN DE EJECUCIÓN
2026-02-27 14:30:00   Procesados este run    : 2500
2026-02-27 14:30:00     ✓ Exitosos           : 2495
2026-02-27 14:30:00     ✗ Fallidos           : 5
2026-02-27 14:30:00   ─────────────────────────────────────────
2026-02-27 14:30:00   PROGRESO ACUMULADO
2026-02-27 14:30:00     Total en BD          : 2500
2026-02-27 14:30:00     ✓ Completados        : 2495
2026-02-27 14:30:00     ✗ Con error          : 5
2026-02-27 14:30:00     ⏳ Pendientes         : 0
2026-02-27 14:30:00 ═══════════════════════════════════════════
```

---

## Archivos que se generan automáticamente

No tienes que crear nada — el script los genera solo:

| Archivo | Para qué sirve |
|---|---|
| `checkpoint.json` | Registra qué documentos ya fueron procesados |
| `logs/sync_YYYY-MM-DD_HH-MM-SS.log` | Log detallado de cada ejecución |

Tu carpeta irá creciendo así:

```
mi-carpeta/
├── sync-linux
├── sync-windows.exe
├── .env.example
├── .env                               ← creado por ti
├── checkpoint.json                    ← creado automáticamente
└── logs/
    ├── sync_2026-02-27_14-19-53.log   ← Ejecución 1
    ├── sync_2026-02-27_15-00-00.log   ← Ejecución 2
    └── sync_2026-02-27_16-30-00.log   ← Ejecución 3
```

---

## Ver el historial de logs

Cada ejecución genera su propio archivo de log. Puedes abrirlos con
cualquier editor de texto o desde la terminal:

```bash
# Linux — ver el log más reciente
cat logs/$(ls logs/ | tail -1)

# Linux — ver un log específico
cat logs/sync_2026-02-27_14-19-53.log

# Windows (PowerShell) — ver el log más reciente
Get-Content (Get-ChildItem logs | Sort-Object LastWriteTime | Select-Object -Last 1).FullName
```

---

## Ver el progreso acumulado

El archivo `checkpoint.json` muestra el estado de cada documento:

```bash
# Linux
cat checkpoint.json

# Windows (PowerShell)
Get-Content checkpoint.json
```

Verás algo como:

```json
{
  "created_at": "2026-02-27T14:19:53Z",
  "updated_at": "2026-02-27T14:20:15Z",
  "documents": {
    "ed867ce5-dfab-4bb6-92cd-03f7f28466c8": {
      "status": "ok",
      "orden_echo": "TIQ-11",
      "updated_at": "2026-02-27T14:19:56Z"
    },
    "86727eaa-fc11-422e-ba33-d4dc30564458": {
      "status": "error",
      "orden_echo": "TIQ-12",
      "updated_at": "2026-02-27T14:20:01Z",
      "error": "HTTP 500 → Internal Server Error"
    }
  }
}
```

---

## ¿Qué hacer si el proceso se interrumpe?

Si el script se corta por cualquier razón (error de red, reinicio, etc.),
simplemente vuelve a ejecutarlo:

```bash
# Linux
./sync-linux

# Windows
sync-windows.exe
```

Retoma automáticamente desde donde se quedó. Los documentos ya marcados
como `ok` en el checkpoint se omiten sin tocarlos de nuevo.

---

## ¿Qué hacer si necesito empezar desde cero?

Si necesitas reprocesar todos los documentos desde el principio,
elimina el checkpoint:

```bash
# Linux
rm checkpoint.json

# Windows (cmd)
del checkpoint.json
```

La próxima ejecución tratará todos los documentos como nuevos.

---

## Problemas comunes

| Problema | Causa probable | Solución |
|---|---|---|
| `Error conectando a BD` | Credenciales incorrectas o sin acceso de red | Verificar `.env` y conectividad |
| `✗ ERROR: HTTP 401` | API sin autorización | Verificar `API_BASE_URL` y `USUARIO_ACCION` en `.env` |
| `✗ ERROR: HTTP 404` | El documento no existe en SharePoint | Revisar el `DocumentoId` en el log |
| `✗ ERROR: HTTP 500` | Error interno de la API | Reintentar — si persiste contactar al desarrollador |
| El script termina pero quedan pendientes | Algunos documentos fallaron | Volver a ejecutar — reintenta solo los fallidos |
