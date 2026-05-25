# PRODE — Predicciones Mundial 2026

## ¿Qué es PRODE?

PRODE permite a los usuarios registrados predecir el resultado exacto de los partidos de la Selección Argentina en el Mundial 2026. Si aciertan el resultado exacto (goles de Argentina y del rival en 90 minutos), ganan un batido/voucher directo.

> **Importante**: PRODE está diseñado exclusivamente para los partidos de Argentina. No es un prode general de todo el mundial.

---

## Endpoints

### Base URL

Todas las rutas bajo `/api/v1/prode/`.

### Autenticación

- Endpoints de usuario: requieren `Authorization: Bearer {{JWT}}` (mismo JWT que el login del usuario).
- Endpoints admin: requieren header `X-Prode-Admin-Key: {{ADMIN_KEY}}`.

---

### Endpoints de Usuario

#### `GET /api/v1/prode/matches`

Lista los partidos visibles de Argentina con la predicción del usuario autenticado (si existe).

**Response**:
```json
{
  "success": true,
  "data": [
    {
      "id": "uuid",
      "stage": "GRUPO_1",
      "opponent": "Brasil",
      "kickoff_at": "2026-06-14T19:00:00Z",
      "cutoff_at": "2026-06-14T18:00:00-03:00",
      "status": "SCHEDULED",
      "is_open": true,
      "argentina_goals": null,
      "opponent_goals": null,
      "my_prediction": null
    }
  ]
}
```

| Campo | Tipo | Descripción |
|---|---|---|
| `id` | string | UUID del partido |
| `stage` | string | Etapa: `GRUPO_1`, `GRUPO_2`, `GRUPO_3`, `OCTAVOS`, `CUARTOS`, `SEMIS`, `FINAL` |
| `opponent` | string | Rival de Argentina |
| `kickoff_at` | datetime (ISO 8601) | Fecha/hora del partido |
| `cutoff_at` | datetime (ISO 8601) | Fecha/hora límite para predecir (kickoff - 1h, hora Argentina) |
| `status` | string | Estado del partido |
| `is_open` | boolean | `true` si sigue aceptando predicciones |
| `argentina_goals` | int or null | Goles de Argentina (solo si ya se cargó resultado) |
| `opponent_goals` | int or null | Goles del rival (solo si ya se cargó resultado) |
| `my_prediction` | object or null | La predicción del usuario si existe |

#### `GET /api/v1/prode/matches/{matchID}`

Detalle de un partido con la predicción del usuario.

#### `PUT /api/v1/prode/matches/{matchID}/prediction`

Crear o editar la predicción del usuario para un partido.

**Request**:
```json
{
  "argentina_goals": 2,
  "opponent_goals": 1
}
```

**Reglas**:
- El usuario puede crear/editar su predicción **todas las veces que quiera** hasta 1 hora antes del partido.
- El corte se calcula **server-side** en zona horaria `America/Argentina/Buenos_Aires`.
- Si el cutoff ya pasó, responde `409 Conflict` con mensaje "La hora límite para predecir este partido ya pasó".
- Solo una predicción por usuario por partido (la edición reemplaza la anterior).
- Los goles deben ser no negativos y máximo 50.

**Response**:
```json
{
  "success": true,
  "data": {
    "id": "uuid",
    "match_id": "uuid",
    "argentina_goals": 2,
    "opponent_goals": 1,
    "status": "PENDING",
    "created_at": "2026-05-24T13:00:00Z",
    "updated_at": "2026-05-24T13:00:00Z"
  }
}
```

#### `GET /api/v1/prode/predictions/me`

Lista todas las predicciones del usuario autenticado, ordenadas por fecha de creación descendente.

---

### Endpoints Admin

Requieren header `X-Prode-Admin-Key: {{ADMIN_KEY}}`.

#### `POST /api/v1/prode/admin/matches`

Crear un partido.

**Request**:
```json
{
  "stage": "GRUPO_1",
  "opponent": "Brasil",
  "kickoff_at": "2026-06-14T19:00:00Z",
  "is_visible": true
}
```

| Campo | Tipo | Obligatorio | Descripción |
|---|---|---|---|
| `stage` | string | sí | Etapa del torneo |
| `opponent` | string | sí | Rival |
| `kickoff_at` | datetime | sí | Fecha/hora del partido (ISO 8601) |
| `is_visible` | boolean | sí | `true` = visible para usuarios, `false` = borrador |

Si `is_visible=true` y `kickoff_at` tiene fecha, el partido arranca como `SCHEDULED`. Si no, arranca como `DRAFT`.

#### `PATCH /api/v1/prode/admin/matches/{matchID}`

Actualizar campos de un partido.

**Request** (todos opcionales):
```json
{
  "stage": "OCTAVOS",
  "opponent": "Francia",
  "kickoff_at": "2026-06-30T17:00:00Z",
  "is_visible": true,
  "status": "CANCELLED"
}
```

Campos editables:
- `stage` — etapa
- `opponent` — rival
- `kickoff_at` — horario
- `is_visible` — visibilidad
- `status` — estado: `DRAFT`, `SCHEDULED`, `OPEN`, `CLOSED`, `CANCELLED`

#### `PUT /api/v1/prode/admin/matches/{matchID}/result`

Cargar el resultado de 90 minutos del partido.

**Request**:
```json
{
  "argentina_goals": 2,
  "opponent_goals": 0
}
```

El partido pasa a estado `RESULT_RECORDED`.

#### `POST /api/v1/prode/admin/matches/{matchID}/settle`

Ejecutar el settlement: evaluar todas las predicciones contra el resultado cargado.

- Predicciones con acierto exacto → se asigna un voucher del inventario + email al usuario.
- Si no hay vouchers disponibles → el premio queda `PENDING_INVENTORY`.
- Predicciones incorrectas → se marcan `INCORRECT`.
- El partido pasa a `EVALUATED`.
- **Idempotente**: si se ejecuta dos veces, no duplica premios.

**Response**:
```json
{
  "success": true,
  "data": {
    "match_id": "uuid",
    "status": "EVALUATED",
    "total_predictions": 15,
    "correct": 3
  }
}
```

#### `POST /api/v1/prode/admin/rewards/retry`

Reintentar asignar voucher a los premios que quedaron `PENDING_INVENTORY` por falta de stock. Útil después de cargar más vouchers al sistema.

**Response**:
```json
{
  "success": true,
  "data": {
    "processed": 3,
    "assigned": 2,
    "failed": 1,
    "pending": 1
  }
}
```

---

## Estados de un partido

```
DRAFT ──→ SCHEDULED ──→ OPEN ──→ CLOSED ──→ RESULT_RECORDED ──→ EVALUATED
  │                                                              │
  └──→ CANCELLED                                                │
                                                                │
                                                  PENDING_INVENTORY (cuando no hay vouchers)
                                                  se resuelve con retry después
```

| Estado | Significado |
|---|---|
| `DRAFT` | Borrador, no visible aún |
| `SCHEDULED` | Programado, acepta predicciones |
| `OPEN` | Abierto para predicciones (equivalente a SCHEDULED) |
| `CLOSED` | Pasó el cutoff, ya no se aceptan predicciones |
| `RESULT_RECORDED` | Admin cargó el resultado |
| `EVALUATED` | Settlement ejecutado |
| `CANCELLED` | Partido cancelado |

---

## Estados de una predicción

| Estado | Significado |
|---|---|
| `PENDING` | El usuario predijo, aún no se evaluó |
| `CORRECT` | Acertó el resultado exacto |
| `INCORRECT` | No acertó |

---

## Estados de un premio (reward)

| Estado | Significado |
|---|---|
| `PENDING` | Recién creado |
| `FULFILLED` | Voucher asignado exitosamente |
| `PENDING_INVENTORY` | No había vouchers disponibles, pendiente de retry |
| `FAILED` | Error al asignar |
| `SKIPPED` | Se saltó (ej: la predicción ya no es correcta) |

---

## Cutoff (hora límite para predecir)

Se calcula server-side:

```go
cutoff = kickoff.In("America/Argentina/Buenos_Aires").Add(-1 * time.Hour)
```

- El frontend NO debe confiar en el reloj del cliente para habilitar/deshabilitar el botón de predicción.
- El campo `cutoff_at` en la respuesta de `GET /matches` ya viene calculado.
- El campo `is_open` indica si el match está aceptando predicciones en este momento.
- Si el usuario intenta predecir después del cutoff, la API responde `409 Conflict`.

---

## Feature Flag

Toda la funcionalidad PRODE se controla con la variable de entorno `PRODE_ENABLED`:

- `PRODE_ENABLED=true` → las rutas `/api/v1/prode/*` están registradas y funcionales.
- `PRODE_ENABLED=false` → las rutas no existen (404), las tablas en DB siguen intactas.
- Esto permite **rollback sin pérdida de datos**: solo cambiar la flag y reiniciar.

Variables de entorno adicionales:

| Variable | Descripción |
|---|---|
| `PRODE_ENABLED` | true/false, activa toda la feature |
| `PRODE_MAINTENANCE_ENABLED` | true/false, activa protección de endpoints admin |
| `PRODE_ADMIN_API_KEY` | Clave para header `X-Prode-Admin-Key` (requerida si maintenance activo) |
| `PRODE_ADMIN_EMAILS` | Emails separados por coma para notificaciones de stock agotado |

---

## Notas para el frontend

- **Timezones**: todos los `datetime` se devuelven en ISO 8601. El `cutoff_at` ya está en hora Argentina.
- **Loading/Empty states**: `GET /matches` puede devolver array vacío si no hay partidos visibles.
- **Edición de predicción**: el usuario puede editar su predicción múltiples veces hasta el cutoff. El frontend debería permitir re-abrir el selector de score si `is_open=true`.
- **Resultado del partido**: `argentina_goals` y `opponent_goals` vienen como `null` hasta que el admin cargue el resultado.
- **Mi predicción**: viene dentro de cada match en `my_prediction`, o en el listado de `GET /predictions/me`.
- **Premio**: después del settlement, si el usuario acertó, se asigna un voucher. El reward no tiene un endpoint específico todavía, pero el estado se puede inferir: si la predicción está `CORRECT`, el usuario debería tener un voucher asignado (visible en `/api/v1/voucher/me`).
- **Admin**: necesita una interfaz para crear partidos, cargar resultados, ejecutar settlement y reintentar premios. No hay un frontend admin todavía.
