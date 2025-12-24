[🇬🇧 English](../CREDENTIALS.md) | [🇺🇦 Українська](../uk/CREDENTIALS.uk.md) | [🇩🇪 Deutsch](../de/CREDENTIALS.de.md) | 🇵🇹 **Português** | [🇪🇸 Español](CREDENTIALS.es.md)

---

# Credenciais de Desenvolvimento

Referência rápida para usuários padrão e credenciais de acesso com funções RBAC.

## Usuários Padrão

### 1. Administrador do Sistema (Admin Principal - Estilo Oracle)

```
Email:    system@promenade.com
Password: passw0rd
Role:     admin
Access:   *:* (acesso completo ao sistema)
```

**Use para:**

- Inicialização e bootstrap do sistema
- Gerenciamento RBAC (funções, permissões)
- Gerenciamento de usuários (banir, suspender, atribuir funções)
- Todas as operações administrativas

### 2. Administrador

```
Email:    admin@promenade.com
Password: passw0rd
Role:     admin
Access:   users:*, posts:*, comments:*, profiles:*, roles:read|list|assign
```

**Use para:**

- Gerenciamento de usuários (criar, atualizar, excluir, banir)
- Gerenciamento de conteúdo (posts, comentários)
- Atribuição de funções a usuários
- Testar permissões de nível administrativo

### 4. Moderador

```
Email:    moderator@promenade.com
Password: passw0rd
Role:     moderator
Access:   posts:read|update|delete|list, comments:*, profiles:read|list
```

**Use para:**

- Moderação de conteúdo (posts, comentários)
- Gerenciamento de comentários (aprovar, excluir)
- Testar fluxos de trabalho de moderação
- Visibilidade limitada de usuários (somente leitura)

### 5. Usuário Regular

```
Email:    alexander.vasilenko@gmail.com
Password: 03041965
Role:     user
Access:   posts:create|read, comments:create|read, profiles:create|read
```

**Use para:**

- Testar fluxos de trabalho de usuário regular
- Criação de conteúdo próprio (posts, comentários)
- Gerenciamento de perfil
- Operações básicas de usuário

### Acesso de Convidado (não autenticado)

Sem necessidade de login:

```
Role:     guest (implícito)
Access:   posts:read, comments:read, profiles:read
```

**Use para:**

- Navegação de conteúdo público
- Testar acesso não autenticado
- Operações somente leitura

## Exemplos de Login Rápido

```bash
# Login como Administrador do Sistema (admin com acesso completo)
curl -X POST http://localhost:8081/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"system@promenade.com","password":"passw0rd"}' | jq

# Login como Administrador (admin)
curl -X POST http://localhost:8081/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@promenade.com","password":"passw0rd"}' | jq

# Login como Moderador (moderator)
curl -X POST http://localhost:8081/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"moderator@promenade.com","password":"passw0rd"}' | jq

# Login como Usuário Regular (user)
curl -X POST http://localhost:8081/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"alexander.vasilenko@gmail.com","password":"03041965"}' | jq

# Salvar token para reutilização (administrador do sistema)
export TOKEN=$(curl -s -X POST http://localhost:8081/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"system@promenade.com","password":"passw0rd"}' | jq -r '.data.access_token')

# Usar token em requisições
curl -X GET http://localhost:8081/api/v1/auth/me \
  -H "Authorization: Bearer $TOKEN" | jq
```

## Testando Permissões RBAC

```bash
# Testar acesso do administrador do sistema (deve funcionar - acesso completo)
curl -X GET http://localhost:8081/api/v1/admin/users \
  -H "Authorization: Bearer $SYSTEM_TOKEN" | jq

# Testar acesso do administrador (deve funcionar - tem users:*)
curl -X GET http://localhost:8081/api/v1/admin/users \
  -H "Authorization: Bearer $ADMIN_TOKEN" | jq

# Testar acesso do moderador (deve falhar - sem permissão users:list)
curl -X GET http://localhost:8081/api/v1/admin/users \
  -H "Authorization: Bearer $MODERATOR_TOKEN" | jq

# Testar acesso do usuário (deve falhar - sem permissões de admin)
curl -X GET http://localhost:8081/api/v1/admin/users \
  -H "Authorization: Bearer $USER_TOKEN" | jq
```

## Acesso ao Banco de Dados

```bash
# Conectar ao banco de dados de desenvolvimento
psql -h localhost -p 5432 -U system -d promenade_dev
# Password: passw0rd

# Visualizar todos os usuários com suas funções
SELECT
    u.email,
    u.name,
    r.name as role,
    r.description
FROM users u
LEFT JOIN user_roles ur ON u.id = ur.user_id
LEFT JOIN roles r ON ur.role_id = r.id
ORDER BY u.email;

# Visualizar permissões de funções
SELECT
    r.name as role,
    p.resource,
    p.action,
    p.description
FROM roles r
JOIN role_permissions rp ON r.id = rp.role_id
JOIN permissions p ON rp.permission_id = p.id
ORDER BY r.name, p.resource, p.action;
```

## [!] Aviso de Segurança

**Estas credenciais são APENAS PARA DESENVOLVIMENTO!**

Antes de implantar em produção:

1. Altere todas as senhas padrão
2. Remova ou desative contas de administrador padrão
3. Use credenciais específicas do ambiente
4. Ative o gerenciamento adequado de secrets
5. Configure provedores de autenticação adequados

## Precisa de Ajuda?

- [Guia de Autorização](docs/AUTHORIZATION.md) - Documentação do sistema RBAC
- [Guia de Testes](docs/TESTING_GUIDE.md) - Como testar com autenticação
- [Documentação da API](README.md#api-examples) - Referência completa da API
