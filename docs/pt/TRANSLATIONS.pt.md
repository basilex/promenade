# Traduções da Documentação Promenade

[🇬🇧 English](../TRANSLATIONS.md) | [🇺🇦 Українська](../uk/TRANSLATIONS.uk.md) | [🇩🇪 Deutsch](../de/TRANSLATIONS.de.md) | 🇵🇹 **Português** | [🇪🇸 Español](../es/TRANSLATIONS.es.md)

---

🌍 **Available Languages / Доступні мови / Verfügbare Sprachen / Idiomas Disponíveis**

A documentação do Promenade está disponível em vários idiomas para torná-la acessível a desenvolvedores em todo o mundo.

---

## Traduções Disponíveis

### 🇬🇧 English (Primário)

- Status: **Completo** ✅
- Localização: `/docs/` e `README.md`
- Mantenedor: Equipe principal

### 🇺🇦 Українська (Ukrainian)

- Status: **Ativo** 🟢
- Localização: `/docs/uk/` e `README.uk.md`
- Cobertura: README (100%), Documentos principais (em progresso)
- [Navegar documentos ucranianos →](uk/README.md)

### 🇩🇪 Deutsch (German)

- Status: **Planejado** 📋
- Localização: `/docs/de/` e `README.de.md`
- Cobertura: Estrutura pronta, traduções necessárias
- [Navegar documentos alemães →](de/README.md)

### 🇵🇹 Português (Portuguese)

- Status: **Planejado** 📋
- Localização: `/docs/pt/` e `README.pt.md`
- Cobertura: Estrutura pronta, traduções necessárias
- [Navegar documentos portugueses →](pt/README.md)

### 🇪🇸 Español (Spanish)

- Status: **Planejado** 📋
- Localização: `/docs/es/` e `README.es.md`
- Cobertura: Estrutura pronta, traduções necessárias
- [Navegar documentos espanhóis →](es/README.md)

---

## Prioridade de Tradução

Os documentos são traduzidos na seguinte ordem de prioridade:

### 🔥 Alta Prioridade (Essencial para começar)

1. `README.md` - Visão geral do projeto e início rápido
2. `docs/ARCHITECTURE_QUICKREF.md` - Referência rápida de arquitetura
3. `docs/MODULE_DEVELOPMENT.md` - Guia de desenvolvimento de módulos

### 🔶 Prioridade Média (Importante para desenvolvimento)

4. `docs/TESTING_GUIDE.md` - Melhores práticas de testes
5. `docs/MODULE_INDEPENDENCE.md` - Princípios de módulos
6. `docs/ARCHITECTURE_OVERVIEW.md` - Arquitetura detalhada

### 🔷 Prioridade Baixa (Tópicos avançados)

7. `docs/PURGE_ARCHITECTURE.md` - Sistema de limpeza
8. `docs/REDIS_BUS_TESTING.md` - Testes do event bus
9. Documentos de referência técnica

---

## Convenção de Nomenclatura de Arquivos

Todos os documentos traduzidos seguem um padrão consistente de nomenclatura:

```
Original:     docs/FILENAME.md
Ukrainian:    docs/uk/FILENAME.uk.md
German:       docs/de/FILENAME.de.md
Portuguese:   docs/pt/FILENAME.pt.md
Spanish:      docs/es/FILENAME.es.md
```

**Códigos de idioma** seguem o padrão ISO 639-1:

- `uk` - Ukrainian (українська)
- `de` - German (Deutsch)
- `pt` - Portuguese (Português)
- `es` - Spanish (Español)

---

## Como Adicionar uma Tradução

### 1. Escolha um Documento

Escolha um documento não traduzido da lista de prioridades acima.

### 2. Crie o Arquivo de Tradução

```bash
# Example: Translate ARCHITECTURE_QUICKREF.md to Ukrainian
touch docs/uk/ARCHITECTURE_QUICKREF.uk.md
```

### 3. Adicione o Seletor de Idioma

No início do **documento original em inglês**, adicione:

```markdown
🇬🇧 **English** | [🇺🇦 Українська](uk/FILENAME.uk.md) | [🇩🇪 Deutsch](de/FILENAME.de.md) | [🇵🇹 Português](pt/FILENAME.pt.md) | [🇪🇸 Español](es/FILENAME.es.md)
```

No início do **documento traduzido**, adicione:

```markdown
🇬🇧 [English](../FILENAME.md) | 🇵🇹 **Português**
```

### 4. Atualize o Progresso

Atualize o README do idioma relevante (`docs/{lang}/README.md`) para marcar o documento como completo.

### 5. Envie um Pull Request

- Título: `docs: Add [Language] translation for [Document]`
- Exemplo: `docs: Add Ukrainian translation for ARCHITECTURE_QUICKREF`

---

## Diretrizes de Tradução

### ✅ FAÇA:

- Traduza termos técnicos de forma consistente (use o glossário abaixo)
- Mantenha os exemplos de código inalterados (código é universal)
- Preserve todos os links (atualize caminhos para versões traduzidas quando disponíveis)
- Mantenha a mesma estrutura do documento
- Use convenções do idioma nativo (por exemplo, formatos de data, aspas)

### ❌ NÃO FAÇA:

- Não traduza nomes de arquivos ou caminhos no código
- Não altere exemplos de código ou saídas de comandos
- Não remova ou pule seções
- Não traduza nomes de marcas (Promenade, PostgreSQL, Redis)
- Não traduza palavras-chave de programação (`func`, `type`, `interface`)

---

## Glossário de Termos Técnicos

Para garantir consistência nas traduções:

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

## Contribuidores

Agradecimentos especiais aos contribuidores de tradução:

- 🇺🇦 Ukrainian: [@basilex](https://github.com/basilex) and AI Assistant
- 🇩🇪 German: _Contribuidores são bem-vindos!_
- 🇵🇹 Portuguese: _Contribuidores são bem-vindos!_
- 🇪🇸 Spanish: _Contribuidores são bem-vindos!_

**Quer contribuir?** Confira o [Guia de Contribuição](../README.md#contributing) e escolha um documento da lista de prioridades acima!

---

## Suporte

- **Problemas de Documentação**: [GitHub Issues](https://github.com/basilex/promenade/issues)
- **Perguntas sobre Tradução**: alexander.vasilenko@gmail.com
- **Comunidade**: Participe de discussões no seu idioma!

---

**Tornando o Promenade acessível a desenvolvedores em todo o mundo** 🌍
