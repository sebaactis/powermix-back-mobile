# Especificación: security (JWT y Auth)

## Propósito

Define los requisitos de seguridad para la inicialización de JWT, el manejo
de logs sensibles y la consistencia de respuestas HTTP en el módulo de autenticación.

---

## Dominio: JWT — Inicialización segura

### Requisito: JWT_SECRET obligatorio

`jwt.NewJWT()` DEBE retornar `(*JWT, error)`.

Si la variable de entorno `JWT_SECRET` no está seteada o es vacía,
`NewJWT()` MUST retornar `nil` y un error que indique que el secret es requerido.

El sistema MUST NOT usar `"dev-secret"` ni ningún valor hardcodeado como
fallback en producción.

#### Escenario: Inicialización exitosa con secrets configurados

- DADO que `JWT_SECRET` y `JWT_RECOVERY_PASS_SECRET` están seteados con valores no vacíos
- CUANDO se llama a `jwt.NewJWT()`
- ENTONCES retorna un `*JWT` válido y `error == nil`

#### Escenario: JWT_SECRET ausente

- DADO que la variable de entorno `JWT_SECRET` no está seteada
- CUANDO se llama a `jwt.NewJWT()`
- ENTONCES retorna `nil` y un error que menciona `"JWT_SECRET"`

#### Escenario: JWT_RECOVERY_PASS_SECRET ausente

- DADO que `JWT_SECRET` está seteada pero `JWT_RECOVERY_PASS_SECRET` no
- CUANDO se llama a `jwt.NewJWT()`
- ENTONCES retorna `nil` y un error que menciona `"JWT_RECOVERY_PASS_SECRET"`

#### Escenario: Secret vacío equivale a ausente

- DADO que `JWT_SECRET` está seteada como cadena vacía
- CUANDO se llama a `jwt.NewJWT()`
- ENTONCES retorna error (vacío no es válido)

---

## Dominio: Auth — Logs no sensibles

### Requisito: Logs de OAuth sin datos sensibles

El handler de OAuth (`OAuthGoogle`) MUST NOT loggear el struct completo de
`OAuthUserInfo` ya que puede contener tokens, IDs de proveedor y otros
datos sensibles.

Si se requiere logging de diagnóstico, SHOULD loggear únicamente el email
del usuario y el proveedor OAuth, sin tokens ni IDs internos.

#### Escenario: Login OAuth no loggea datos sensibles

- DADO que un usuario inicia sesión con Google
- CUANDO el handler `OAuthGoogle` procesa la request
- ENTONCES ningún log contiene el struct completo de `userInfo` (`%+v`)
- Y si se loggea algo, solo incluye campos no sensibles (email, provider)

---

### Requisito: Sin fmt.Println en producción

El código de producción MUST NOT usar `fmt.Println` para debug o logging.
Toda salida de diagnóstico MUST usar `slog` con nivel y campos apropiados.

#### Escenario: Recovery password no imprime a stdout

- DADO que se solicita recovery de password para un email registrado
- CUANDO el handler `RecoveryPasswordRequest` procesa la request
- ENTONCES no hay escritura directa a stdout (`fmt.Println`)
- Y si ocurre un error al enviar el mail, se registra con `slog.Error`

#### Escenario: Error de mailer loggeado correctamente

- DADO que el servicio de mail falla al enviar el email de recovery
- CUANDO el handler maneja el error
- ENTONCES el error queda registrado en el log estructurado con nivel `error`
- Y la respuesta al cliente es la respuesta genérica (sin revelar el error interno)

---

### Requisito: URL de recovery no expuesta en logs

El token de recovery de password y la URL completa que lo contiene
MUST NOT ser loggeados, ya que constituyen credenciales temporales.

#### Escenario: Token de recovery no aparece en logs

- DADO que se genera un token de recovery para un usuario
- CUANDO el handler construye la URL de reset
- ENTONCES la URL completa (que contiene el token) NO aparece en ningún log
- Y el email del usuario NO aparece en logs en el mismo contexto que el token

---

## Dominio: Auth — Consistencia de respuestas HTTP

### Requisito: Todos los handlers usan utils.WriteSuccess / utils.WriteError

Todos los handlers del módulo `auth` MUST usar `utils.WriteSuccess` y
`utils.WriteError` para escribir respuestas HTTP.

MUST NOT usar `json.NewEncoder(w).Encode(...)` directamente para respuestas
de éxito, ya que omite headers (`Content-Type`) y no respeta el contrato
de respuesta estándar de la API.

#### Escenario: UnlockUser responde con formato estándar

- DADO que se hace una request válida a `UnlockUser`
- CUANDO el handler procesa la request exitosamente
- ENTONCES la respuesta tiene `Content-Type: application/json`
- Y el body sigue el formato estándar de `utils.WriteSuccess`
- Y el status code es `200 OK`
