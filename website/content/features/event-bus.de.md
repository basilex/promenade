---
title: "Ereignisgesteuerte Architektur"
description: "Duale Event-Bus-Adapter für asynchrone Kommunikation"
weight: 6
---

## Event-Bus-System

Promenade enthält einen **Event-Bus mit zwei Adaptern** für asynchrone, ereignisgesteuerte Kommunikation.

### Zwei Adapter

**Memory-Adapter** (Standard):

- In-Memory Pub/Sub
- Goroutines + Channels
- Perfekt für Dev/Test
- Null Abhängigkeiten

**Redis-Adapter** (Produktion):

- Verteilter Pub/Sub
- Multi-Instanz-Unterstützung
- Persistente Nachrichtenwarteschlange
- Fehlertoleranz

### Konfiguration

```yaml
# config/app.yaml
bus:
  adapter: "memory" # oder "redis"
  worker_pool_size: 4

  # Redis-spezifisch
  redis:
    host: "localhost"
    port: 6379
    password: ""
    db: 0
```

### Events Veröffentlichen

```go
// Event definieren
type UserRegisteredEvent struct {
    bus.BaseEvent
    UserID string
    Email  string
}

// Veröffentlichen
event := &UserRegisteredEvent{
    BaseEvent: bus.BaseEvent{ID: uuid.New().String()},
    UserID:    user.ID,
    Email:     user.Email,
}
eventBus.Publish(ctx, "user.registered", event)
```

### Events Abonnieren

```go
// Während Modulinitialisierung abonnieren
func (m *EmailModule) Initialize(ctx context.Context, core *module.Core) error {
    return core.EventBus.Subscribe(ctx, "user.registered",
        func(ctx context.Context, e bus.Event) error {
            evt := e.(*UserRegisteredEvent)
            return m.sendWelcomeEmail(ctx, evt.Email)
        },
    )
}
```

### Integrierte Events

- `user.registered` - Neue Benutzerregistrierung
- `user.deleted` - Benutzerkonto entfernt
- `post.created` - Neuer Beitrag veröffentlicht
- `comment.created` - Neuer Kommentar hinzugefügt
- `purge.completed` - Bereinigungsauftrag abgeschlossen

### Modulkommunikation

Module verwenden Events für **lose Kopplung**:

```
Posts Module                Email Module
    |                           |
    | user.registered           |
    |-------------------------->|
    |              Willkommens-E-Mail senden
    |                           |
```

### Vorteile

✅ **Asynchron** - Blockiert den Hauptthread nicht  
✅ **Entkoppelt** - Module kennen sich nicht gegenseitig  
✅ **Skalierbar** - Redis für verteilte Systeme  
✅ **Zuverlässig** - Automatische Wiederholungen und Fehlerbehandlung

[Ausführlicher Leitfaden →](/pkg/bus/README.md)
