# Especificación: config

## Propósito

Define los requisitos de carga y validación de configuración de la aplicación.
Un arranque con configuración inválida o incompleta DEBE ser detectado y abortado
de forma explícita antes de que el servidor empiece a aceptar requests.

---

## Requisitos

### Requisito: Validación de variables de entorno obligatorias

La función `config.Load()` DEBE validar que todas las variables de entorno
requeridas estén presentes y no vacías.

Si alguna variable obligatoria falta, `Load()` DEBE retornar un error descriptivo
que identifique qué variable no está configurada.

Las variables obligatorias son: `HTTP_ADDR`, `DB_DRIVER`, `DSN`,
`MERCAGO_PAGO_TOKEN`, `COFFEJI_KEY`, `COFFEJI_SECRET`, `RESEND_API_KEY`,
`JWT_REFRESH_HASH`.

#### Escenario: Arranque con todas las variables presentes

- DADO que todas las variables de entorno obligatorias están seteadas
- CUANDO se llama a `config.Load()`
- ENTONCES retorna un `Config` válido y `error == nil`

#### Escenario: Arranque con variable faltante

- DADO que la variable de entorno `DSN` no está seteada
- CUANDO se llama a `config.Load()`
- ENTONCES retorna un error que menciona `"DSN"`
- Y el error NO es `nil`

#### Escenario: Variable presente pero vacía

- DADO que `DSN` está seteada como cadena vacía (`""`)
- CUANDO se llama a `config.Load()`
- ENTONCES retorna error (cadena vacía equivale a no configurada)

---

### Requisito: Fallo explícito en arranque

El punto de entrada de la aplicación (`main`) DEBE llamar a `config.Load()` y,
si retorna error, DEBE terminar el proceso con un mensaje de error claro y
código de salida distinto de cero.

La aplicación DEBE NOT continuar arrancando con una configuración inválida.

#### Escenario: main aborta si config es inválida

- DADO que `DSN` no está configurada
- CUANDO la aplicación arranca
- ENTONCES el proceso termina antes de iniciar el servidor HTTP
- Y el mensaje de error indica qué variable falta

---

### Requisito: Sin valores de relleno silenciosos

`config.Load()` MUST NOT retornar strings del tipo `"ENV_X_NOT_SET"` como
valor de ningún campo. Ese patrón enmascara errores de configuración.

#### Escenario: No hay strings de relleno

- DADO que `RESEND_API_KEY` no está seteada
- CUANDO se llama a `config.Load()`
- ENTONCES el campo `ResendKey` del `Config` retornado NO contiene `"ENV_RESEND_API_KEY_NOT_SET"`
- Y el error retornado es distinto de `nil`
