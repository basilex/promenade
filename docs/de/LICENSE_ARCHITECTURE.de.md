# Architektur des Lizenzsystems

Promenade verwendet ein signaturbasiertes Lizenzsystem für kommerzielle IModule. Dieses Dokument beschreibt die Architektur, Implementierung und Nutzungsmuster.

---

## Überblick

### Designprinzipien

1. **Modulunabhängigkeit**: Jedes Modul verwaltet seine eigene Lizenzierung
2. **Signaturbasierte Sicherheit**: HMAC-SHA256 verhindert Manipulation
3. **Sanfte Degradierung**: Gnadenfrist für abgelaufene Lizenzen
4. **Umgebungsspezifisch**: Unterschiedliche Validierungsregeln pro Umgebung
5. **Menschenlesbar**: Lizenzschlüssel sind strukturiert und parsbar

### Lizenzformat

```
PROMENADE-{MODULE}-{TIER}-{EXPIRY}-{SIGNATURE}
```

Beispiel:

```
PROMENADE-ANALYTICS-PRO-20261231-K8mF3pL9qT2xN7vR4zW6yH1jC5eS8bD0aG3hM6nP9
```

Komponenten:

- **Präfix**: Immer `PROMENADE`
- **Modul**: Modulname in GROSSBUCHSTABEN (ANALYTICS, WAREHOUSE, AUDITLOG)
- **Tier**: Lizenzstufe (BASIC, PRO, ENTERPRISE)
- **Ablauf**: Datum im Format YYYYMMDD
- **Signatur**: HMAC-SHA256 der ersten 4 Teile, base64 URL-kodiert

---

## Architektur

### Systemkomponenten

**Lizenzvalidierungsablauf:**

1. **Anwendungsstart**
   - Modul-Registry (`pkg/module`) wird initialisiert
2. **Für Jedes Aktivierte Modul**
   - `IModule.Initialize()` wird aufgerufen
3. **Lizenz-Validator** (`module/license/`)
   - `Parse()` - Lizenzkomponenten extrahieren
   - `Validate()` - Signatur und Ablauf prüfen
   - `HealthCheck()` - Laufende Validierung
4. **Lizenzstatus** (Ergebnis)
   - Gültig - Modul funktioniert normal
   - Abgelaufen (Gnade) - Warnung ausgegeben
   - Ungültig - Modul deaktiviert

### Validierungsablauf

1. **Parsen**: Komponenten aus Lizenzstring extrahieren
2. **Modul verifizieren**: Sicherstellen, dass Lizenz zum Modulnamen passt
3. **Signatur verifizieren**: HMAC-SHA256 Validierung
4. **Ablauf prüfen**: Datum + Gnadenfrist validieren
5. **Status zurückgeben**: Gültig, abgelaufen (Gnade) oder ungültig

### Speicherung

Lizenzen werden als Umgebungsvariablen gespeichert:

- `{MODULE}_LICENSE_KEY`: Der Lizenzschlüssel
- `LICENSE_SECRET`: Geheimnis für Signaturverifizierung (Produktion)

Beispiel:

```bash
export ANALYTICS_LICENSE_KEY="PROMENADE-ANALYTICS-PRO-20261231-..."
export LICENSE_SECRET="your-production-secret-key"
```

---

## Implementierung

### Modulintegration

Jedes kommerzielle Modul implementiert Lizenzvalidierung in seiner `Initialize()` Methode:

```go
func (m *AnalyticsModule) Initialize(cfg interface{}) error {
 // Modulkonfiguration laden
 config, ok := cfg.(*Config)
 if !ok {
 return ErrInvalidConfig
 }

 // Lizenz validieren, falls erforderlich
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

### Lizenzvalidierung

Das license-Paket bietet die zentrale Validierungslogik:

```go
// Parse extrahiert Komponenten aus Lizenzstring
func ParseLicense(licenseKey string) (*License, error)

// Validate prüft Lizenzgültigkeit
func (l *License) Validate(
 expectedModule string,
 secret string,
 options ValidationOptions,
) error

// Generate erstellt neue Lizenz (für Tests/Tools)
func GenerateLicense(
 module, tier, expiry, secret string,
) (string, error)
```

### Gesundheitschecks

Der Gesundheitscheck jedes Moduls enthält den Lizenzstatus:

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

## Konfiguration

### Umgebungsspezifische Einstellungen

#### Entwicklung (`config.dev.yaml`)

```yaml
license_required: false # Optional in der Entwicklung
validate_expiry: false # Ablauf ignorieren
validate_signature: true # Signaturen dennoch prüfen
grace_period_days: 30 # Lange Gnadenfrist
```

#### Test (`config.test.yaml`)

```yaml
license_required: false # Keine Lizenz in CI/CD
validate_expiry: false
validate_signature: false
grace_period_days: 0
```

#### Produktion (`config.prod.yaml`)

```yaml
license_required: true # Verpflichtend
validate_expiry: true # Strikter Ablauf
validate_signature: true # Volle Sicherheit
grace_period_days: 7 # Begrenzte Gnade
validate_on_request: true # Optionale Pro-Request-Validierung
```

### Modulkonfiguration

In `config/modules.yaml`:

```yaml
modules:
  analytics:
  enabled: true
  version: "1.0.0"
  description: "Advanced analytics and reporting"
  license_key: "" # Über Umgebungsvariable setzen
  settings:
  metrics_retention_days: 90
  max_reports_per_user: 10
  max_dashboards_per_user: 5
```

---

## Stufen und Funktionen

### BASIC Stufe

- Basis-Metrikerfassung (100 Metriken/Tag)
- Basisberichte (JSON, CSV)
- 5 Dashboards pro Benutzer
- 30 Tage Datenspeicherung
- E-Mail-Support

### PRO Stufe

- Unbegrenzte Metriken
- Erweiterte Berichte (PDF, XLSX)
- Unbegrenzte Dashboards
- 90 Tage Datenspeicherung
- Geplante Berichte
- Prioritäts-Support

### ENTERPRISE Stufe

- Alle PRO-Funktionen
- Benutzerdefinierte Speicherung (1+ Jahre)
- Multi-Tenant-Unterstützung
- Eigenes Branding
- API-Zugriff
- Dedizierter Support
- On-Premise-Deployment

---

## Verwendung

### Lizenzen Generieren

Verwenden Sie das bereitgestellte Skript:

```bash
# PRO-Lizenz für 365 Tage generieren
./scripts/generate-license.sh analytics PRO 365

# Ausgabe:
# PROMENADE-ANALYTICS-PRO-20261231-K8mF3pL9qT2xN7vR4zW6yH1jC5eS8bD0aG3hM6nP9
```

Umgebungsvariable setzen:

```bash
export ANALYTICS_LICENSE_KEY="PROMENADE-ANALYTICS-PRO-20261231-..."
export LICENSE_SECRET="your-production-secret"
```

### Lizenzvalidierung Testen

Das analytics-Modul enthält umfassende Tests:

```bash
# Lizenzvalidierungstests ausführen
go test ./internal/modules/analytics/license/... -v

# Testszenarien:
# - Gültiges Lizenzparsing
# - Ungültige Formaterkennung
# - Signaturverifizierung
# - Ablaufbehandlung
# - Gnadenfristen-Logik
# - Modulabweichungserkennung
```

### Lizenzgenerierungstool

Lizenzgenerator bauen:

```bash
make build-license-generator

# Oder manuell:
go build -o ./bin/license-generator ./cmd/license-generator/main.go
```

Lizenz programmatisch generieren:

```bash
./bin/license-generator \
 -module=ANALYTICS \
 -tier=PRO \
 -expiry=20261231 \
 -secret="your-secret-key"
```

---

## Sicherheit

### HMAC-SHA256 Signaturen

- **Algorithmus**: HMAC mit SHA-256
- **Schlüssel**: Gespeichert in `LICENSE_SECRET` Umgebungsvariable
- **Kodierung**: Base64 URL-Kodierung (URL-sicher, ohne Padding)
- **Verifizierung**: Konstant-Zeit-Vergleich zur Verhinderung von Timing-Angriffen

### Geheimnismanagement

**Entwicklung**:

```bash
# Standard-Testgeheimnis (akzeptabel für lokale Entwicklung)
LICENSE_SECRET="default-dev-secret-change-in-production"
```

**Produktion**:

```bash
# Starkes Geheimnis generieren (32+ Bytes)
openssl rand -base64 32

# In sicherem Vault speichern (AWS Secrets Manager, HashiCorp Vault, etc.)
# Über Umgebungsvariable setzen (nie in git committen)
```

### Rotationsstrategie

1. Neues Geheimnis generieren
2. Neue Lizenzen mit neuem Geheimnis signieren
3. Beide alte und neue Geheimnisse während Übergangsphase unterstützen
4. Altes Geheimnis nach Gnadenfrist verwerfen

---

## Fehlerbehandlung

### Fehlertypen

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

### Sanfte Degradierung

1. **Fehlende Lizenz** (dev/test): Modul lädt mit reduzierten Funktionen
2. **Abgelaufene Lizenz**: Gnadenfrist erlaubt Weiterbetrieb mit Warnungen
3. **Ungültige Lizenz**: Modul-Initialisierung schlägt in Produktion fehl, loggt in dev

### Benutzernachrichten

```go
// Produktionsfehler (strikt)
return fmt.Errorf("analytics module requires valid license: %w", err)

// Entwicklungswarnung (permissiv)
logger.Warn("Analytics license validation failed, continuing in dev mode",
 "error", err)
```

---

## Monetarisierungsstrategie

### Aktueller Stand

| Modul         | Status           | Stufe              | Grund                            |
| ------------- | ---------------- | ------------------ | -------------------------------- |
| posts         | Kostenlos        | N/A                | Kern-Social-Features             |
| profiles      | Kostenlos        | N/A                | Kern-Benutzer-Features           |
| **analytics** | ** Kommerziell** | **PRO/ENTERPRISE** | **Aktiv: Analytics als Premium** |
| warehouse     |  Geplant       | TBD                | Zukunft: Bestandsverwaltung      |

### Roadmap

**Phase 1 (Aktuell - Aktiv)**:

- Analytics-Modul ist **kommerziell** (PRO/ENTERPRISE Stufen)
- Fokus auf Business-Metriken, Berichten, Dashboards
- Ziel: KMUs, Unternehmen mit Dateneinsichten-Bedarf
- Status: Implementiert und aktiviert

**Phase 2 (Q2 2026 - Geplant)**:

- Release **Audit Log Modul** (kommerziell)
- Features: Compliance, GDPR, detaillierte Audit-Trails
- Ziel: Regulierte Branchen, Unternehmen

**Phase 3 (Nach Audit Log)**:

- **Analytics auf kostenlose Stufe** verschieben
- Analytics wird für alle Benutzer zugänglich
- Audit Log bleibt kommerziell für Compliance-Bedürfnisse

### Begründung

1. **Analytics Zuerst**: Funktioniert mit vorhandenen Daten, sofortiger Wert
2. **Audit Log Premium**: Compliance ist Unternehmensanforderung
3. **Kostenloses Analytics**: Breitere Akzeptanz, Upsell zu Audit Log

---

## Monitoring und Observability

### Lizenzmetriken

Lizenznutzung über Logs verfolgen:

```go
logger.Info("License validated",
 "module", "analytics",
 "tier", license.Tier,
 "expiry", license.ExpiryDate,
 "days_remaining", daysRemaining)
```

### Health Endpoint

Lizenzstatus über Gesundheitscheck verfügbar:

```bash
curl http://localhost:8080/api/v1/analytics/health

# Antwort:
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

### Alerts

Monitoring einrichten für:

- Abgelaufene Lizenzen (Gnadenfrist endet)
- Ungültige Lizenzversuche
- Lizenzverifizierungsfehler

---

## Testen

### Unit-Tests

Jedes Modul enthält umfassende Lizenztests:

```bash
# Alle Lizenztests ausführen
go test ./internal/modules/*/license/... -v

# Spezifische Modultests ausführen
go test ./internal/modules/analytics/license/... -v
```

Testabdeckung:

- Gültiges Lizenzparsing
- Ungültige Formaterkennung
- Signaturverifizierung
- Ablaufbehandlung
- Gnadenfristen-Logik
- Modulabweichung
- Stufenvalidierung

### Integrationstests

Modulinitialisierung mit verschiedenen Lizenzzuständen testen:

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

## Fehlerbehebung

### Häufige Probleme

**1. Lizenz Erforderlich Fehler**

```
Error: license validation failed: license key is required
```

Lösung:

```bash
export ANALYTICS_LICENSE_KEY="PROMENADE-ANALYTICS-PRO-20261231-..."
```

**2. Ungültige Signatur**

```
Error: invalid license signature
```

Ursachen:

- Falsches `LICENSE_SECRET`
- Manipulierter Lizenzschlüssel
- Schlüssel falsch kopiert

Lösung: Lizenz mit korrektem Geheimnis neu generieren.

**3. Abgelaufene Lizenz**

```
Warning: License expired 3 days ago, grace period active
```

Lösung: Lizenz vor Ablauf der Gnadenfrist erneuern (Standard 7 Tage).

**4. Modulabweichung**

```
Error: license module mismatch: expected ANALYTICS, got WAREHOUSE
```

Lösung: Korrekte Lizenz für jedes Modul verwenden.

### Debug-Modus

Ausführliches Lizenzlogging aktivieren:

```yaml
# config/app.dev.yaml
log:
  level: debug

modules:
  analytics:
  settings:
  debug_license: true # Alle Validierungsschritte loggen
```

### Lizenzvalidierungstool

Lizenzgültigkeit manuell testen:

```bash
# Lizenz validieren
go run ./cmd/license-generator/main.go \
 -validate \
 -key="PROMENADE-ANALYTICS-PRO-20261231-..." \
 -secret="your-secret-key"

# Ausgabe:
# License valid
# IModule: ANALYTICS
# Tier: PRO
# Expiry: 2026-12-31
# Days remaining: 365
```

---

## Best Practices

### Für Entwickler

1. **Niemals Geheimnisse committen**: `.env` Dateien verwenden (gitignored)
2. **Mit ungültigen Lizenzen testen**: Fehlerbehandlung sicherstellen
3. **Stufenfunktionen dokumentieren**: Klare Feature-Matrix pro Stufe
4. **Sanfte Degradierung**: Nicht bei Lizenzproblemen in dev crashen
5. **Strukturiertes Logging**: Lizenzereignisse für Monitoring loggen

### Für Betreiber

1. **Geheimnisse regelmäßig rotieren**: `LICENSE_SECRET` quartalsweise aktualisieren
2. **Ablaufdaten überwachen**: 30 Tage vor Ablauf alarmieren
3. **Sichere Geheimnisspeicherung**: Vault-Lösungen verwenden (AWS, HashiCorp)
4. **Lizenznutzung verfolgen**: Überwachen, welche Stufen aktiv sind
5. **Erneuerungen planen**: Vor Ablauf der Gnadenfrist erneuern

### Für Modulautoren

1. **Muster folgen**: Bestehendes analytics-Modul als Vorlage verwenden
2. **Features dokumentieren**: Klare stufenbasierte Feature-Dokumentation
3. **Gründlich testen**: Umfassende Lizenzvalidierungstests
4. **Gesundheitschecks**: Lizenzstatus in Health-Endpoints einschließen
5. **Fehlermeldungen**: Benutzerfreundliche, umsetzbare Fehlermeldungen

---

## Referenzen

- [Analytics Modul README](../internal/modules/analytics/README.de.md)
- [License Package](../internal/modules/analytics/license/)
- [Modulentwicklungsleitfaden](./MODULE_DEVELOPMENT.de.md)
- [Konfigurationsarchitektur](./MODULE_CONFIG_ARCHITECTURE.de.md)

---

## Anhang

### Lizenzformat-Spezifikation

```
Format: PROMENADE-{MODULE}-{TIER}-{EXPIRY}-{SIGNATURE}

Einschränkungen:
- MODULE: [A-Z0-9]+ (Großbuchstaben, keine Leerzeichen)
- TIER: BASIC|PRO|ENTERPRISE
- EXPIRY: YYYYMMDD (gültiges Datum)
- SIGNATURE: [A-Za-z0-9_-]+ (base64 URL-kodiert)

Max. Länge: 256 Zeichen
Min. Länge: 50 Zeichen
```

### HMAC-SHA256 Implementierung

```go
func generateSignature(data, secret string) string {
 h := hmac.New(sha256.New, []byte(secret))
 h.Write([]byte(data))
 signature := base64.URLEncoding.EncodeToString(h.Sum(nil))
 return strings.TrimRight(signature, "=") // Padding entfernen
}

func verifySignature(data, signature, secret string) bool {
 expected := generateSignature(data, secret)
 return hmac.Equal([]byte(expected), []byte(signature))
}
```

### Test-Checkliste

- [ ] Gültige Lizenz parsen
- [ ] Ungültiges Format ablehnen
- [ ] Signatur verifizieren
- [ ] Ablauf prüfen
- [ ] Gnadenfrist testen
- [ ] Modulabweichung erkennen
- [ ] Stufen validieren
- [ ] Fehlende Lizenz testen
- [ ] Manipulierte Lizenz testen
- [ ] Integration mit Modul
- [ ] Gesundheitscheck enthält Status
- [ ] Fehlermeldungen sind klar

---

**Letzte Aktualisierung**: 2024-12-19
**Version**: 1.0.0
**Status**: Produktionsbereit
