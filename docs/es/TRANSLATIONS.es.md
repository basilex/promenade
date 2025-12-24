# Traducciones de la Documentación de Promenade

[🇬🇧 English](../TRANSLATIONS.md) | [🇺🇦 Українська](../uk/TRANSLATIONS.uk.md) | [🇩🇪 Deutsch](../de/TRANSLATIONS.de.md) | [🇵🇹 Português](../pt/TRANSLATIONS.pt.md) | 🇪🇸 **Español**

---

 **Available Languages / Доступні мови / Verfügbare Sprachen / Idiomas Disponíveis**

La documentación de Promenade está disponible en varios idiomas para hacerla accesible a desarrolladores de todo el mundo.

---

## Traducciones Disponibles

### 🇬🇧 English (Principal)

- Estado: **Completo**
- Ubicación: `/docs/` y `README.md`
- Mantenedor: Equipo principal

### 🇺🇦 Українська (Ukrainian)

- Estado: **Activo** 🟢
- Ubicación: `/docs/uk/` y `README.uk.md`
- Cobertura: README (100%), Documentos principales (en progreso)
- [Explorar documentos ucranianos →](uk/README.md)

### 🇩🇪 Deutsch (German)

- Estado: **Planificado** 📋
- Ubicación: `/docs/de/` y `README.de.md`
- Cobertura: Estructura lista, se necesitan traducciones
- [Explorar documentos alemanes →](de/README.md)

### 🇵🇹 Português (Portuguese)

- Estado: **Planificado** 📋
- Ubicación: `/docs/pt/` y `README.pt.md`
- Cobertura: Estructura lista, se necesitan traducciones
- [Explorar documentos portugueses →](pt/README.md)

### 🇪🇸 Español (Spanish)

- Estado: **Planificado** 📋
- Ubicación: `/docs/es/` y `README.es.md`
- Cobertura: Estructura lista, se necesitan traducciones
- [Explorar documentos españoles →](es/README.md)

---

## Prioridad de Traducción

Los documentos se traducen en el siguiente orden de prioridad:

###  Alta Prioridad (Esencial para comenzar)

1. `README.md` - Descripción general del proyecto e inicio rápido
2. `docs/ARCHITECTURE_QUICKREF.md` - Referencia rápida de arquitectura
3. `docs/MODULE_DEVELOPMENT.md` - Guía de desarrollo de módulos

### 🔶 Prioridad Media (Importante para desarrollo)

4. `docs/TESTING_GUIDE.md` - Mejores prácticas de pruebas
5. `docs/MODULE_INDEPENDENCE.md` - Principios de módulos
6. `docs/ARCHITECTURE_OVERVIEW.md` - Arquitectura detallada

### 🔷 Prioridad Baja (Temas avanzados)

7. `docs/PURGE_ARCHITECTURE.md` - Sistema de purga
8. `docs/REDIS_BUS_TESTING.md` - Pruebas del event bus
9. Documentos de referencia técnica

---

## Convención de Nomenclatura de Archivos

Todos los documentos traducidos siguen un patrón de nomenclatura consistente:

```
Original:     docs/FILENAME.md
Ukrainian:    docs/uk/FILENAME.uk.md
German:       docs/de/FILENAME.de.md
Portuguese:   docs/pt/FILENAME.pt.md
Spanish:      docs/es/FILENAME.es.md
```

**Códigos de idioma** siguen el estándar ISO 639-1:

- `uk` - Ukrainian (українська)
- `de` - German (Deutsch)
- `pt` - Portuguese (Português)
- `es` - Spanish (Español)

---

## Cómo Agregar una Traducción

### 1. Elija un Documento

Elija un documento no traducido de la lista de prioridades anterior.

### 2. Cree el Archivo de Traducción

```bash
# Example: Translate ARCHITECTURE_QUICKREF.md to Ukrainian
touch docs/uk/ARCHITECTURE_QUICKREF.uk.md
```

### 3. Agregue el Selector de Idioma

Al principio del **documento original en inglés**, agregue:

```markdown
🇬🇧 **English** | [🇺🇦 Українська](uk/FILENAME.uk.md) | [🇩🇪 Deutsch](de/FILENAME.de.md) | [🇵🇹 Português](pt/FILENAME.pt.md) | [🇪🇸 Español](es/FILENAME.es.md)
```

Al principio del **documento traducido**, agregue:

```markdown
🇬🇧 [English](../FILENAME.md) | 🇪🇸 **Español**
```

### 4. Actualice el Progreso

Actualice el README del idioma correspondiente (`docs/{lang}/README.md`) para marcar el documento como completo.

### 5. Envíe un Pull Request

- Título: `docs: Add [Language] translation for [Document]`
- Ejemplo: `docs: Add Ukrainian translation for ARCHITECTURE_QUICKREF`

---

## Pautas de Traducción

###  HAGA:

- Traduzca términos técnicos de manera consistente (use el glosario a continuación)
- Mantenga los ejemplos de código sin cambios (el código es universal)
- Preserve todos los enlaces (actualice rutas a versiones traducidas cuando estén disponibles)
- Mantenga la misma estructura del documento
- Use convenciones del idioma nativo (por ejemplo, formatos de fecha, comillas)

###  NO HAGA:

- No traduzca nombres de archivos o rutas en el código
- No cambie ejemplos de código o salidas de comandos
- No elimine ni omita secciones
- No traduzca nombres de marcas (Promenade, PostgreSQL, Redis)
- No traduzca palabras clave de programación (`func`, `type`, `interface`)

---

## Glosario de Términos Técnicos

Para garantizar consistencia en las traducciones:

| English            | 🇺🇦 Українська            | 🇩🇪 Deutsch         | 🇵🇹 Português      | 🇪🇸 Español          |
| ------------------ | ------------------------ | ------------------ | ----------------- | ------------------- |
| Module             | Модуль                   | Modul              | Módulo            | Módulo              |
| Repository         | Репозиторій              | Repository         | Repositório       | Repositorio         |
| Use Case           | Use Case / Бізнес-логіка | Anwendungsfall     | Caso de Uso       | Caso de Uso         |
| Handler            | Хендлер                  | Handler            | Manipulador       | Manejador           |
| Entity             | Сутність                 | Entität            | Entidade          | Entidad             |
| Migration          | Міграція                 | Migration          | Migração          | Migración           |
| Testing            | Тестування               | Testen             | Teste             | Pruebas             |
| Clean Architecture | Clean Architecture       | Clean Architecture | Arquitetura Limpa | Arquitectura Limpia |

---

## Colaboradores

Agradecimiento especial a los colaboradores de traducción:

- 🇺🇦 Ukrainian: [@basilex](https://github.com/basilex) and AI Assistant
- 🇩🇪 German: _¡Se buscan colaboradores!_
- 🇵🇹 Portuguese: _¡Se buscan colaboradores!_
- 🇪🇸 Spanish: _¡Se buscan colaboradores!_

**¿Quiere colaborar?** ¡Consulte la [Guía de Contribución](../README.md#contributing) y elija un documento de la lista de prioridades anterior!

---

## Soporte

- **Problemas de Documentación**: [GitHub Issues](https://github.com/basilex/promenade/issues)
- **Preguntas sobre Traducción**: alexander.vasilenko@gmail.com
- **Comunidad**: ¡Únase a las discusiones en su idioma!

---

**Haciendo Promenade accesible para desarrolladores de todo el mundo**
