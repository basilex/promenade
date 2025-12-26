---
title: "Clean Architecture"
description: "Strikte Schichtenarchitektur mit klarer Trennung der Zuständigkeiten"
weight: 1
---

## Clean Architecture

Promenade folgt den Prinzipien der **Clean Architecture von Uncle Bob** mit strikten Abhängigkeitsregeln und klaren Schichtgrenzen.

### Vier Schichten

```
Domain (Entitäten, Interfaces)
    ↓
Use Case (Geschäftslogik)
    ↓
Adapter (Repos, Handler)
    ↓
Infrastructure (DB, HTTP, Events)
```

### Abhängigkeitsregel

**Innere Schichten hängen niemals von äußeren ab**. Das gewährleistet:

- Geschäftslogik ist Framework-agnostisch
- Einfaches Testen mit Mock-Implementierungen
- Flexible Infrastrukturänderungen
- Klare Grenzen zwischen Komponenten

### Beispielstruktur

```go
// Domain-Schicht - reine Geschäftsentitäten
type User struct {
    ID        uuid.UUID
    Email     string
    Status    UserStatus
}

// Use Case - nur Geschäftslogik
type UserUseCase interface {
    RegisterUser(ctx, email, password) (*User, error)
}

// Adapter - Framework-Integration
type UserHandler struct {
    usecase UserUseCase
}

func (h *UserHandler) Register(c *gin.Context) {
    // HTTP-spezifischer Code hier
}
```

### Vorteile

✅ **Testbar** - Jede Schicht unabhängig mocken  
✅ **Wartbar** - Änderungen isoliert auf eine Schicht  
✅ **Skalierbar** - Features hinzufügen ohne Code zu brechen  
✅ **Framework-unabhängig** - Geschäftslogik kennt HTTP oder DB nicht

[Mehr erfahren →](/promenade/docs/ARCHITECTURE_OVERVIEW)
