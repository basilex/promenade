# Índice de Documentação do Promenade

[🇬🇧 English](../INDEX.md) | [🇺🇦 Українська](../uk/INDEX.uk.md) | [🇩🇪 Deutsch](../de/INDEX.de.md) | 🇵🇹 **Português** | [🇪🇸 Español](../es/INDEX.es.md)

Este diretório contém documentação abrangente sobre a arquitetura da aplicação Promenade, fluxos de trabalho de desenvolvimento e melhores práticas.

>  **Novo!** A documentação está agora disponível em vários idiomas. Veja [TRANSLATIONS.md](../TRANSLATIONS.md) para o status da tradução e diretrizes de contribuição.

---

## Comece Aqui

### Novo no Promenade?

1. **[ARCHITECTURE_OVERVIEW.md](ARCHITECTURE_OVERVIEW.pt.md)** - Diagramas visuais de arquitetura e visão geral dos componentes
2. **[ARCHITECTURE_QUICKREF.md](ARCHITECTURE_QUICKREF.pt.md)** - Guia de referência rápida para desenvolvedores
3. **[README.md](../../README.md)** - README principal do projeto

### Revisão de Arquitetura

- **[ARCHITECTURE_AUDIT.md](ARCHITECTURE_AUDIT.pt.md)** - Auditoria completa de conformidade de arquitetura (Core vs Módulos)

---

## Conceitos Principais

### Arquitetura & Design

| Documento                                               | Descrição                                 | Quando Ler                         |
| ------------------------------------------------------- | ----------------------------------------- | ---------------------------------- |
| [ARCHITECTURE_OVERVIEW.md](ARCHITECTURE_OVERVIEW.pt.md) | Arquitetura visual completa com diagramas | Entender estrutura do sistema      |
| [ARCHITECTURE_AUDIT.md](ARCHITECTURE_AUDIT.pt.md)       | Relatório de conformidade de arquitetura  | Verificar princípios de design     |
| [ARCHITECTURE_QUICKREF.md](ARCHITECTURE_QUICKREF.pt.md) | Referência rápida para padrões comuns     | Desenvolvimento diário             |
| [../../internal/CORE.md](../../internal/CORE.md)        | Documentação dos componentes Core         | Entender responsabilidades do core |

### Módulos

| Documento                                                            | Descrição                                  | Quando Ler                    |
| -------------------------------------------------------------------- | ------------------------------------------ | ----------------------------- |
| [MODULE_DEVELOPMENT.md](MODULE_DEVELOPMENT.pt.md)                    | Guia completo para criação de módulos      | Construir novos módulos       |
| [MODULE_INDEPENDENCE.md](MODULE_INDEPENDENCE.pt.md)                  | Princípios de independência de módulos     | Entender limites dos módulos  |
| [MODULE_CONFIG_ARCHITECTURE.md](MODULE_CONFIG_ARCHITECTURE.pt.md)    | Gerenciamento de configuração para módulos | Configurar configs de módulo  |
| [../../internal/modules/README.md](../../internal/modules/README.md) | Estrutura de diretórios de módulos         | Visão geral rápida de módulos |

---

## Guias Técnicos

### Banco de Dados & Persistência

| Documento                               | Descrição                           | Quando Ler                     |
| --------------------------------------- | ----------------------------------- | ------------------------------ |
| [UUID_V7_GUIDE.md](UUID_V7_GUIDE.pt.md) | Usando UUIDs ordenados por tempo    | Trabalhar com chaves primárias |
| [SOFT_DELETE.md](../SOFT_DELETE.md)     | Padrões e armadilhas de soft delete | Implementar soft delete        |

### Infraestrutura

| Documento                                         | Descrição                               | Quando Ler                           |
| ------------------------------------------------- | --------------------------------------- | ------------------------------------ |
| [PURGE_ARCHITECTURE.md](PURGE_ARCHITECTURE.pt.md) | Design do sistema de purge automatizado | Implementar políticas de retenção    |
| [REDIS_BUS_TESTING.md](../REDIS_BUS_TESTING.md)   | Testes com Redis Event IBus              | Testar recursos orientados a eventos |
| [LOGGING.md](../LOGGING.md)                       | Logging estruturado com contexto        | Adicionar logging ao código          |

### Segurança & Autenticação

| Documento                                             | Descrição                                           | Quando Ler                       |
| ----------------------------------------------------- | --------------------------------------------------- | -------------------------------- |
| [AUTH_SCHEMA.md](AUTH_SCHEMA.pt.md)                   | Sistema de autenticação (registro, login, JWT)      | Entender fluxo de autenticação   |
| [AUTHORIZATION.md](AUTHORIZATION.pt.md)               | Sistema de permissões RBAC                          | Implementar autorização          |
| [CREDENTIALS.md](../CREDENTIALS.md)                   | Usuários e papéis padrão para dev/test              | Testar com usuários predefinidos |
| [LICENSE_ARCHITECTURE.md](LICENSE_ARCHITECTURE.pt.md) | Sistema de licenciamento de módulos com HMAC-SHA256 | Implementar módulos comerciais   |

---

## Testes

| Documento                                                     | Descrição                                | Quando Ler                         |
| ------------------------------------------------------------- | ---------------------------------------- | ---------------------------------- |
| [TESTING_GUIDE.md](TESTING_GUIDE.pt.md)                       | Estratégia completa de testes            | Escrever testes                    |
| [TESTING_INFRASTRUCTURE.md](TESTING_INFRASTRUCTURE.pt.md)     | Configuração de infraestrutura de testes | Configurar ambiente de teste       |
| [MOCK_GENERATION_STANDARD.md](../MOCK_GENERATION_STANDARD.md) | Abordagem unificada de geração de mocks  | Trabalhar com mocks de repositório |
| [MOCK_STANDARDIZATION.md](../MOCK_STANDARDIZATION.md)         | Resumo de padronização de mocks          | Entender unificação de mocks       |
| [../../test/README.md](../../test/README.md)                  | Estrutura de diretórios de testes        | Entender organização de testes     |

---

## Fluxos de Trabalho de Desenvolvimento

### Build & Deploy

| Documento                                               | Descrição                            | Quando Ler              |
| ------------------------------------------------------- | ------------------------------------ | ----------------------- |
| [MAKEFILE_ARCHITECTURE.md](../MAKEFILE_ARCHITECTURE.md) | Documentação do sistema Makefile     | Usar comandos make      |
| [../../docker/README.md](../../docker/README.md)        | Configuração e deployment com Docker | Containerizar aplicação |

### Validação & Qualidade

| Documento                         | Descrição                       | Quando Ler                    |
| --------------------------------- | ------------------------------- | ----------------------------- |
| [VALIDATION.md](../VALIDATION.md) | Padrões de validação de entrada | Adicionar regras de validação |

---

## Documentação da API

### API V1

- [v1/v1_docs.go](../v1/v1_docs.go) - Documentação da API V1
- [v1/v1_swagger.yaml](../v1/v1_swagger.yaml) - Especificação Swagger V1 (YAML)
- [v1/v1_swagger.json](../v1/v1_swagger.json) - Especificação Swagger V1 (JSON)

### API V2

- [v2/v2_docs.go](../v2/v2_docs.go) - Documentação da API V2
- [v2/v2_swagger.yaml](../v2/v2_swagger.yaml) - Especificação Swagger V2 (YAML)
- [v2/v2_swagger.json](../v2/v2_swagger.json) - Especificação Swagger V2 (JSON)

---

## Por Tópico

### Arquitetura Core

- [ARCHITECTURE_OVERVIEW.md](ARCHITECTURE_OVERVIEW.pt.md) - Visão geral visual
- [ARCHITECTURE_AUDIT.md](ARCHITECTURE_AUDIT.pt.md) - Revisão de conformidade
- [ARCHITECTURE_QUICKREF.md](ARCHITECTURE_QUICKREF.pt.md) - Referência rápida
- [../../internal/CORE.md](../../internal/CORE.md) - Componentes Core

### Sistema de Módulos

- [MODULE_DEVELOPMENT.md](MODULE_DEVELOPMENT.pt.md) - Guia de desenvolvimento
- [MODULE_INDEPENDENCE.md](MODULE_INDEPENDENCE.pt.md) - Princípios de independência
- [MODULE_CONFIG_ARCHITECTURE.md](MODULE_CONFIG_ARCHITECTURE.pt.md) - Configuração
- [../../internal/modules/README.md](../../internal/modules/README.md) - Índice de módulos

### Gerenciamento de Dados

- [UUID_V7_GUIDE.md](UUID_V7_GUIDE.pt.md) - Chaves primárias
- [SOFT_DELETE.md](../SOFT_DELETE.md) - Padrões de soft delete
- [PURGE_ARCHITECTURE.md](PURGE_ARCHITECTURE.pt.md) - Purge automatizado

### Segurança & Auth

- [AUTH_SCHEMA.md](AUTH_SCHEMA.pt.md) - Autenticação
- [AUTHORIZATION.md](AUTHORIZATION.pt.md) - Sistema RBAC
- [CREDENTIALS.md](../CREDENTIALS.md) - Tratamento de credenciais

### Testes & Qualidade

- [TESTING_GUIDE.md](TESTING_GUIDE.pt.md) - Estratégia de testes
- [TESTING_INFRASTRUCTURE.md](TESTING_INFRASTRUCTURE.pt.md) - Configuração de testes
- [VALIDATION.md](../VALIDATION.md) - Validação de entrada
- [REDIS_BUS_TESTING.md](../REDIS_BUS_TESTING.md) - Testes de event bus

### Infraestrutura

- [LOGGING.md](../LOGGING.md) - Logging estruturado
- [MAKEFILE_ARCHITECTURE.md](../MAKEFILE_ARCHITECTURE.md) - Sistema de build
- [../../docker/README.md](../../docker/README.md) - Configuração Docker

---

## Caminhos de Aprendizado

### Caminho 1: Entendendo o Sistema (Novo Desenvolvedor)

1. [README.md](../../README.md) - Visão geral do projeto
2. [ARCHITECTURE_OVERVIEW.md](ARCHITECTURE_OVERVIEW.pt.md) - Arquitetura do sistema
3. [ARCHITECTURE_QUICKREF.md](ARCHITECTURE_QUICKREF.pt.md) - Padrões comuns
4. [MODULE_DEVELOPMENT.md](MODULE_DEVELOPMENT.pt.md) - Construir recursos
5. [TESTING_GUIDE.md](TESTING_GUIDE.pt.md) - Testar seu código

### Caminho 2: Construindo um Novo Módulo

1. [MODULE_DEVELOPMENT.md](MODULE_DEVELOPMENT.pt.md) - Guia de criação de módulos
2. [MODULE_INDEPENDENCE.md](MODULE_INDEPENDENCE.pt.md) - Princípios de design
3. [MODULE_CONFIG_ARCHITECTURE.md](MODULE_CONFIG_ARCHITECTURE.pt.md) - Configuração
4. [UUID_V7_GUIDE.md](UUID_V7_GUIDE.pt.md) - Chaves primárias
5. [SOFT_DELETE.md](../SOFT_DELETE.md) - Se usar soft delete
6. [PURGE_ARCHITECTURE.md](PURGE_ARCHITECTURE.pt.md) - Se implementar purge

### Caminho 3: Revisão de Arquitetura (Technical Lead)

1. [ARCHITECTURE_AUDIT.md](ARCHITECTURE_AUDIT.pt.md) - Análise do estado atual
2. [ARCHITECTURE_OVERVIEW.md](ARCHITECTURE_OVERVIEW.pt.md) - Diagramas visuais
3. [../../internal/CORE.md](../../internal/CORE.md) - Limites do Core
4. [MODULE_INDEPENDENCE.md](MODULE_INDEPENDENCE.pt.md) - Isolamento de módulos

### Caminho 4: Implementação de Segurança

1. [AUTH_SCHEMA.md](AUTH_SCHEMA.pt.md) - Fluxos de autenticação
2. [AUTHORIZATION.md](AUTHORIZATION.pt.md) - Permissões RBAC
3. [CREDENTIALS.md](../CREDENTIALS.md) - Segurança de credenciais

### Caminho 5: Testes & Qualidade

1. [TESTING_GUIDE.md](TESTING_GUIDE.pt.md) - Estratégia de testes
2. [TESTING_INFRASTRUCTURE.md](TESTING_INFRASTRUCTURE.pt.md) - Configuração de testes
3. [VALIDATION.md](../VALIDATION.md) - Validação de entrada
4. [REDIS_BUS_TESTING.md](../REDIS_BUS_TESTING.md) - Testes de eventos

---

## Buscas Rápidas

### Como faço para...

**...criar um novo módulo?**
→ [MODULE_DEVELOPMENT.md](MODULE_DEVELOPMENT.pt.md)

**...adicionar configuração ao meu módulo?**
→ [MODULE_CONFIG_ARCHITECTURE.md](MODULE_CONFIG_ARCHITECTURE.pt.md)

**...implementar soft delete?**
→ [SOFT_DELETE.md](../SOFT_DELETE.md)

**...adicionar políticas de retenção?**
→ [PURGE_ARCHITECTURE.md](PURGE_ARCHITECTURE.pt.md)

**...usar UUIDs corretamente?**
→ [UUID_V7_GUIDE.md](UUID_V7_GUIDE.pt.md)

**...implementar autenticação?**
→ [AUTH_SCHEMA.md](AUTH_SCHEMA.pt.md)

**...adicionar permissões RBAC?**
→ [AUTHORIZATION.md](AUTHORIZATION.pt.md)

**...escrever testes?**
→ [TESTING_GUIDE.md](TESTING_GUIDE.pt.md)

**...adicionar logging?**
→ [LOGGING.md](../LOGGING.md)

**...validar entrada?**
→ [VALIDATION.md](../VALIDATION.md)

**...entender a arquitetura?**
→ [ARCHITECTURE_OVERVIEW.md](ARCHITECTURE_OVERVIEW.pt.md)

**...verificar se meu código segue os princípios?**
→ [ARCHITECTURE_AUDIT.md](ARCHITECTURE_AUDIT.pt.md)

**...obter uma referência rápida?**
→ [ARCHITECTURE_QUICKREF.md](ARCHITECTURE_QUICKREF.pt.md)

---

## Atualizações Recentes

### 22 de Dezembro de 2025

- Criada documentação abrangente de arquitetura:
- [ARCHITECTURE_AUDIT.md](ARCHITECTURE_AUDIT.pt.md) - Revisão completa de conformidade
- [ARCHITECTURE_OVERVIEW.md](ARCHITECTURE_OVERVIEW.pt.md) - Diagramas visuais
- [ARCHITECTURE_QUICKREF.md](ARCHITECTURE_QUICKREF.pt.md) - Referência rápida
- Documentação do sistema de purge atualizada:
- [PURGE_ARCHITECTURE.md](PURGE_ARCHITECTURE.pt.md) - Abordagem baseada em registro
- Independência de módulos verificada (15.000+ linhas removidas do core)

---

## 🤝 Contribuindo

Ao adicionar nova documentação:

1. **Escolha o tipo certo:**

   - `ARCHITECTURE_*.md` - Arquitetura e padrões de design
   - `MODULE_*.md` - Documentação do sistema de módulos
   - `*_GUIDE.md` - Guias práticos e tutoriais
   - `*_SCHEMA.md` - Esquemas e estruturas de dados
   - `README.md` - Visões gerais de diretórios

2. **Atualize este índice:**

   - Adicione à seção relevante
   - Atualize "Atualizações Recentes"
   - Adicione a "Buscas Rápidas" se aplicável

3. **Faça referências cruzadas:**

   - Link para documentos relacionados
   - Atualize docs relacionados com links de volta

4. **Mantenha atualizado:**
   - Atualize quando a arquitetura mudar
   - Arquive docs desatualizados com prefixo `DEPRECATED_`

---

## 📞 Suporte

- **Perguntas sobre arquitetura?** → Leia [ARCHITECTURE_OVERVIEW.md](ARCHITECTURE_OVERVIEW.pt.md)
- **Perguntas sobre módulos?** → Leia [MODULE_DEVELOPMENT.md](MODULE_DEVELOPMENT.pt.md)
- **Perguntas sobre testes?** → Leia [TESTING_GUIDE.md](TESTING_GUIDE.pt.md)
- **Outras perguntas?** → Verifique este índice ou o [README.md](../../README.md) principal

---

**Versão da Documentação:** 2.0
**Última Atualização:** 22 de Dezembro de 2025
**Mantido por:** Equipe de Desenvolvimento Promenade
