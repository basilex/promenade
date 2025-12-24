# Arquitectura del Sistema de Licencias

Promenade utiliza un sistema de licencias basado en firmas para módulos comerciales. Este documento describe la arquitectura, implementación y patrones de uso.

---

## Visión General

### Principios de Diseño

1. **Independencia del Módulo**: Cada módulo gestiona su propia licencia
2. **Seguridad Basada en Firmas**: HMAC-SHA256 previene la manipulación
3. **Degradación Gradual**: Período de gracia para licencias expiradas
4. **Específico por Entorno**: Reglas de validación diferentes por entorno
5. **Legible para Humanos**: Las claves de licencia están estructuradas y son parseables

### Formato de Licencia

```
PROMENADE-{MODULE}-{TIER}-{EXPIRY}-{SIGNATURE}
```

Ejemplo:

```
PROMENADE-ANALYTICS-PRO-20261231-K8mF3pL9qT2xN7vR4zW6yH1jC5eS8bD0aG3hM6nP9
```

Componentes:

- **Prefijo**: Siempre `PROMENADE`
- **Módulo**: Nombre del módulo en MAYÚSCULAS (ANALYTICS, WAREHOUSE, AUDITLOG)
- **Nivel**: Nivel de licencia (BASIC, PRO, ENTERPRISE)
- **Expiración**: Fecha en formato YYYYMMDD
- **Firma**: HMAC-SHA256 de las primeras 4 partes, codificada en base64 URL

---

## Arquitectura

### Componentes del Sistema

**Flujo de Validación de Licencia:**

1. **Inicio de Aplicación**
   - Registro de Módulos (`pkg/module`) se inicializa
2. **Para Cada Módulo Habilitado**
   - Se llama a `IModule.Initialize()`
3. **Validador de Licencia** (`module/license/`)
   - `Parse()` - Extraer componentes de licencia
   - `Validate()` - Verificar firma y vencimiento
   - `HealthCheck()` - Validación continua
4. **Estado de Licencia** (resultado)
   - Válida - Módulo opera normalmente
   - Expirada (Gracia) - Advertencia emitida
   - Inválida - Módulo deshabilitado

### Flujo de Validación

1. **Parsear**: Extraer componentes de la cadena de licencia
2. **Verificar Módulo**: Asegurar que la licencia coincide con el nombre del módulo
3. **Verificar Firma**: Validación HMAC-SHA256
4. **Verificar Expiración**: Validar fecha + período de gracia
5. **Devolver Estado**: Válida, expirada (gracia) o inválida

### Almacenamiento

Las licencias se almacenan como variables de entorno:

- `{MODULE}_LICENSE_KEY`: La clave de licencia
- `LICENSE_SECRET`: Secreto para verificación de firma (producción)

Ejemplo:

```bash
export ANALYTICS_LICENSE_KEY="PROMENADE-ANALYTICS-PRO-20261231-..."
export LICENSE_SECRET="your-production-secret-key"
```

---

## Implementación

### Integración de Módulo

Cada módulo comercial implementa validación de licencia en su método `Initialize()`:

```go
func (m *AnalyticsModule) Initialize(cfg interface{}) error {
 // Cargar configuración del módulo
 config, ok := cfg.(*Config)
 if !ok {
 return ErrInvalidConfig
 }

 // Validar licencia si es necesario
 if config.LicenseRequired {
 if err := m.validateLicense(config); err != nil {
 return fmt.Errorf("license validation failed: %w", err)
 }
 }

 return nil
}

func (m *AnalyticsModule) validateLicense(config *Config) error {
 licenseKey := config.LicenseKey
 if licenseKey == "" {
 licenseKey = os.Getenv("ANALYTICS_LICENSE_KEY")
 }

 if licenseKey == "" {
 return ErrLicenseRequired
 }

 secret := os.Getenv("LICENSE_SECRET")
 if secret == "" {
 secret = "default-dev-secret"
 }

 license, err := ParseLicense(licenseKey)
 if err != nil {
 return err
 }

 validationOptions := ValidationOptions{
 ValidateExpiry: config.ValidateExpiry,
 ValidateSignature: config.ValidateSignature,
 GracePeriodDays: config.GracePeriodDays,
 }

 return license.Validate("ANALYTICS", secret, validationOptions)
}
```

### Validación de Licencia

El paquete license proporciona la lógica central de validación:

```go
// Parse extrae componentes de la cadena de licencia
func ParseLicense(licenseKey string) (*License, error)

// Validate verifica la validez de la licencia
func (l *License) Validate(
 expectedModule string,
 secret string,
 options ValidationOptions,
) error

// Generate crea una nueva licencia (para pruebas/herramientas)
func GenerateLicense(
 module, tier, expiry, secret string,
) (string, error)
```

### Verificaciones de Salud

La verificación de salud de cada módulo incluye el estado de la licencia:

```go
func (m *AnalyticsModule) HealthCheck(ctx context.Context) error {
 if m.license != nil {
 if m.license.IsExpired() && !m.license.InGracePeriod(7) {
 return ErrLicenseExpired
 }
 }
 return nil
}
```

---

## Configuración

### Configuraciones Específicas por Entorno

#### Desarrollo (`config.dev.yaml`)

```yaml
license_required: false # Opcional en desarrollo
validate_expiry: false # Ignorar expiración
validate_signature: true # Aún verificar firmas
grace_period_days: 30 # Período de gracia largo
```

#### Prueba (`config.test.yaml`)

```yaml
license_required: false # Sin licencia en CI/CD
validate_expiry: false
validate_signature: false
grace_period_days: 0
```

#### Producción (`config.prod.yaml`)

```yaml
license_required: true # Obligatoria
validate_expiry: true # Expiración estricta
validate_signature: true # Seguridad completa
grace_period_days: 7 # Gracia limitada
validate_on_request: true # Validación opcional por solicitud
```

### Configuración de Módulo

En `config/modules.yaml`:

```yaml
modules:
  analytics:
  enabled: true
  version: "1.0.0"
  description: "Advanced analytics and reporting"
  license_key: "" # Definir vía variable de entorno
  settings:
  metrics_retention_days: 90
  max_reports_per_user: 10
  max_dashboards_per_user: 5
```

---

## Niveles y Características

### Nivel BASIC

- Recolección básica de métricas (100 métricas/día)
- Informes básicos (JSON, CSV)
- 5 dashboards por usuario
- 30 días de retención de datos
- Soporte por email

### Nivel PRO

- Métricas ilimitadas
- Informes avanzados (PDF, XLSX)
- Dashboards ilimitados
- 90 días de retención de datos
- Informes programados
- Soporte prioritario

### Nivel ENTERPRISE

- Todas las características PRO
- Retención personalizada (1+ años)
- Soporte multi-tenant
- Marca personalizada
- Acceso a API
- Soporte dedicado
- Despliegue on-premise

---

## Uso

### Generando Licencias

Use el script proporcionado:

```bash
# Generar licencia PRO válida por 365 días
./scripts/generate-license.sh analytics PRO 365

# Salida:
# PROMENADE-ANALYTICS-PRO-20261231-K8mF3pL9qT2xN7vR4zW6yH1jC5eS8bD0aG3hM6nP9
```

Definir variable de entorno:

```bash
export ANALYTICS_LICENSE_KEY="PROMENADE-ANALYTICS-PRO-20261231-..."
export LICENSE_SECRET="your-production-secret"
```

### Probando Validación de Licencia

El módulo analytics incluye pruebas completas:

```bash
# Ejecutar pruebas de validación de licencia
go test ./internal/modules/analytics/license/... -v

# Escenarios de prueba:
# - Parsear licencia válida
# - Detectar formato inválido
# - Verificación de firma
# - Manejo de expiración
# - Lógica de período de gracia
# - Detección de incompatibilidad de módulo
```

### Herramienta de Generación de Licencia

Construir el generador de licencias:

```bash
make build-license-generator

# O manualmente:
go build -o ./bin/license-generator ./cmd/license-generator/main.go
```

Generar licencia programáticamente:

```bash
./bin/license-generator \
 -module=ANALYTICS \
 -tier=PRO \
 -expiry=20261231 \
 -secret="your-secret-key"
```

---

## Seguridad

### Firmas HMAC-SHA256

- **Algoritmo**: HMAC con SHA-256
- **Clave**: Almacenada en la variable de entorno `LICENSE_SECRET`
- **Codificación**: Codificación base64 URL (URL-safe, sin padding)
- **Verificación**: Comparación de tiempo constante para prevenir ataques de temporización

### Gestión de Secretos

**Desarrollo**:

```bash
# Secreto de prueba por defecto (aceptable para desarrollo local)
LICENSE_SECRET="default-dev-secret-change-in-production"
```

**Producción**:

```bash
# Generar secreto fuerte (32+ bytes)
openssl rand -base64 32

# Almacenar en vault seguro (AWS Secrets Manager, HashiCorp Vault, etc.)
# Definir vía variable de entorno (nunca hacer commit en git)
```

### Estrategia de Rotación

1. Generar nuevo secreto
2. Firmar nuevas licencias con nuevo secreto
3. Soportar ambos secretos antiguo y nuevo durante transición
4. Deprecar secreto antiguo después del período de gracia

---

## Manejo de Errores

### Tipos de Error

```go
var (
 ErrLicenseRequired = errors.New("license key is required")
 ErrInvalidFormat = errors.New("invalid license format")
 ErrInvalidSignature = errors.New("invalid license signature")
 ErrLicenseExpired = errors.New("license has expired")
 ErrModuleMismatch = errors.New("license module mismatch")
 ErrInvalidTier = errors.New("invalid license tier")
)
```

### Degradación Gradual

1. **Licencia Ausente** (dev/test): Módulo carga con características reducidas
2. **Licencia Expirada**: Período de gracia permite operación continua con advertencias
3. **Licencia Inválida**: Inicialización del módulo falla en producción, registra en dev

### Mensajes al Usuario

```go
// Error de producción (estricto)
return fmt.Errorf("analytics module requires valid license: %w", err)

// Advertencia de desarrollo (permisivo)
logger.Warn("Analytics license validation failed, continuing in dev mode",
 "error", err)
```

---

## Estrategia de Monetización

### Estado Actual

| Módulo        | Estado         | Nivel              | Motivo                                 |
| ------------- | -------------- | ------------------ | -------------------------------------- |
| posts         | Gratuito       | N/A                | Características sociales principales   |
| profiles      | Gratuito       | N/A                | Características principales de usuario |
| **analytics** | ** Comercial** | **PRO/ENTERPRISE** | **Activo: Analytics como premium**     |
| warehouse     | 🔮 Planificado | TBD                | Futuro: Gestión de inventario          |

### Hoja de Ruta

**Fase 1 (Actual - Activo)**:

- Módulo Analytics es **comercial** (niveles PRO/ENTERPRISE)
- Enfoque en métricas de negocio, informes, dashboards
- Objetivo: PYMEs, empresas con necesidad de insights de datos
- Estado: Implementado y activado

**Fase 2 (Q2 2026 - Planificado)**:

- Lanzamiento del **Módulo Audit Log** (comercial)
- Características: Cumplimiento, GDPR, pistas de auditoría detalladas
- Objetivo: Industrias reguladas, empresas

**Fase 3 (Después de Audit Log)**:

- **Migrar Analytics a nivel gratuito**
- Analytics se vuelve accesible para todos los usuarios
- Audit Log permanece comercial para necesidades de cumplimiento

### Justificación

1. **Analytics Primero**: Funciona con datos existentes, valor inmediato
2. **Audit Log Premium**: Cumplimiento es requisito empresarial
3. **Analytics Gratuito**: Adopción más amplia, upsell a Audit Log

---

## Monitoreo y Observabilidad

### Métricas de Licencia

Rastrear uso de licencia a través de logs:

```go
logger.Info("License validated",
 "module", "analytics",
 "tier", license.Tier,
 "expiry", license.ExpiryDate,
 "days_remaining", daysRemaining)
```

### Endpoint de Salud

Estado de licencia disponible vía verificación de salud:

```bash
curl http://localhost:8080/api/v1/analytics/health

# Respuesta:
{
 "status": "healthy",
 "license": {
 "tier": "PRO",
 "expiry": "2026-12-31",
 "days_remaining": 365,
 "in_grace_period": false
 }
}
```

### Alertas

Configurar monitoreo para:

- Licencias expiradas (período de gracia terminando)
- Intentos de licencia inválida
- Errores de verificación de licencia

---

## Pruebas

### Pruebas Unitarias

Cada módulo incluye pruebas de licencia completas:

```bash
# Ejecutar todas las pruebas de licencia
go test ./internal/modules/*/license/... -v

# Ejecutar pruebas de módulo específico
go test ./internal/modules/analytics/license/... -v
```

Cobertura de pruebas:

- Parsear licencia válida
- Detectar formato inválido
- Verificación de firma
- Manejo de expiración
- Lógica de período de gracia
- Incompatibilidad de módulo
- Validación de nivel

### Pruebas de Integración

Probar inicialización de módulo con diferentes estados de licencia:

```go
func TestModuleInitialization(t *testing.T) {
 tests := []struct {
 name string
 licenseKey string
 expectError bool
 }{
 {"valid license", validLicense, false},
 {"expired license", expiredLicense, true},
 {"invalid signature", tamperedLicense, true},
 {"missing license", "", true},
 }
 // ...
}
```

---

## Solución de Problemas

### Problemas Comunes

**1. Error de Licencia Requerida**

```
Error: license validation failed: license key is required
```

Solución:

```bash
export ANALYTICS_LICENSE_KEY="PROMENADE-ANALYTICS-PRO-20261231-..."
```

**2. Firma Inválida**

```
Error: invalid license signature
```

Causas:

- `LICENSE_SECRET` incorrecto
- Clave de licencia manipulada
- Clave copiada incorrectamente

Solución: Regenerar licencia con secreto correcto.

**3. Licencia Expirada**

```
Warning: License expired 3 days ago, grace period active
```

Solución: Renovar licencia antes del fin del período de gracia (predeterminado 7 días).

**4. Incompatibilidad de Módulo**

```
Error: license module mismatch: expected ANALYTICS, got WAREHOUSE
```

Solución: Usar licencia correcta para cada módulo.

### Modo de Depuración

Habilitar registro detallado de licencia:

```yaml
# config/app.dev.yaml
log:
  level: debug

modules:
  analytics:
  settings:
  debug_license: true # Registrar todas las etapas de validación
```

### Herramienta de Validación de Licencia

Probar validez de licencia manualmente:

```bash
# Validar licencia
go run ./cmd/license-generator/main.go \
 -validate \
 -key="PROMENADE-ANALYTICS-PRO-20261231-..." \
 -secret="your-secret-key"

# Salida:
# License valid
# IModule: ANALYTICS
# Tier: PRO
# Expiry: 2026-12-31
# Days remaining: 365
```

---

## Mejores Prácticas

### Para Desarrolladores

1. **Nunca Hacer Commit de Secretos**: Usar archivos `.env` (gitignored)
2. **Probar con Licencias Inválidas**: Garantizar manejo de errores
3. **Documentar Características por Nivel**: Matriz clara de características por nivel
4. **Degradación Gradual**: No fallar en problemas de licencia en dev
5. **Registro Estructurado**: Registrar eventos de licencia para monitoreo

### Para Operadores

1. **Rotar Secretos Regularmente**: Actualizar `LICENSE_SECRET` trimestralmente
2. **Monitorear Fechas de Expiración**: Alertar 30 días antes de expiración
3. **Almacenamiento Seguro de Secretos**: Usar soluciones de vault (AWS, HashiCorp)
4. **Rastrear Uso de Licencia**: Monitorear qué niveles están activos
5. **Planificar Renovaciones**: Renovar antes del fin del período de gracia

### Para Autores de Módulos

1. **Seguir Patrones**: Usar módulo analytics existente como plantilla
2. **Documentar Características**: Documentación clara de características basadas en nivel
3. **Probar Completamente**: Pruebas completas de validación de licencia
4. **Verificaciones de Salud**: Incluir estado de licencia en endpoints de salud
5. **Mensajes de Error**: Mensajes amigables y accionables para el usuario

---

## Referencias

- [README del Módulo Analytics](../internal/modules/analytics/README.es.md)
- [Paquete License](../internal/modules/analytics/license/)
- [Guía de Desarrollo de Módulos](./MODULE_DEVELOPMENT.es.md)
- [Arquitectura de Configuración](./MODULE_CONFIG_ARCHITECTURE.es.md)

---

## Apéndice

### Especificación de Formato de Licencia

```
Formato: PROMENADE-{MODULE}-{TIER}-{EXPIRY}-{SIGNATURE}

Restricciones:
- MODULE: [A-Z0-9]+ (mayúsculas, sin espacios)
- TIER: BASIC|PRO|ENTERPRISE
- EXPIRY: YYYYMMDD (fecha válida)
- SIGNATURE: [A-Za-z0-9_-]+ (base64 URL-encoded)

Longitud máxima: 256 caracteres
Longitud mínima: 50 caracteres
```

### Implementación HMAC-SHA256

```go
func generateSignature(data, secret string) string {
 h := hmac.New(sha256.New, []byte(secret))
 h.Write([]byte(data))
 signature := base64.URLEncoding.EncodeToString(h.Sum(nil))
 return strings.TrimRight(signature, "=") // Eliminar padding
}

func verifySignature(data, signature, secret string) bool {
 expected := generateSignature(data, secret)
 return hmac.Equal([]byte(expected), []byte(signature))
}
```

### Lista de Verificación de Pruebas

- [ ] Parsear licencia válida
- [ ] Rechazar formato inválido
- [ ] Verificar firma
- [ ] Verificar expiración
- [ ] Probar período de gracia
- [ ] Detectar incompatibilidad de módulo
- [ ] Validar niveles
- [ ] Probar licencia ausente
- [ ] Probar licencia manipulada
- [ ] Integración con módulo
- [ ] Verificación de salud incluye estado
- [ ] Mensajes de error claros

---

**Última Actualización**: 2024-12-19
**Versión**: 1.0.0
**Estado**: Listo para Producción
