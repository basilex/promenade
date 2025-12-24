[🇬🇧 English](../CREDENTIALS.md) | [🇺🇦 Українська](../uk/CREDENTIALS.uk.md) | [🇩🇪 Deutsch](../de/CREDENTIALS.de.md) | [🇵🇹 Português](../pt/CREDENTIALS.pt.md) | 🇪🇸 **Español**

---

# Credenciales de Desarrollo

Referencia rápida para usuarios predeterminados y credenciales de acceso con roles RBAC.

## Usuarios Predeterminados

### 1. Administrador del Sistema (Admin Principal - Estilo Oracle)

```
Email:    system@promenade.com
Password: passw0rd
Role:     admin
Access:   *:* (acceso completo al sistema)
```

**Usar para:**

- Inicialización y bootstrap del sistema
- Gestión RBAC (roles, permisos)
- Gestión de usuarios (banear, suspender, asignar roles)
- Todas las operaciones administrativas

### 2. Administrador

```
Email:    admin@promenade.com
Password: passw0rd
Role:     admin
Access:   users:*, posts:*, comments:*, profiles:*, roles:read|list|assign
```

**Usar para:**

- Gestión de usuarios (crear, actualizar, eliminar, banear)
- Gestión de contenido (posts, comentarios)
- Asignación de roles a usuarios
- Probar permisos de nivel administrativo

### 4. Moderador

```
Email:    moderator@promenade.com
Password: passw0rd
Role:     moderator
Access:   posts:read|update|delete|list, comments:*, profiles:read|list
```

**Usar para:**

- Moderación de contenido (posts, comentarios)
- Gestión de comentarios (aprobar, eliminar)
- Probar flujos de trabajo de moderación
- Visibilidad limitada de usuarios (solo lectura)

### 5. Usuario Regular

```
Email:    alexander.vasilenko@gmail.com
Password: 03041965
Role:     user
Access:   posts:create|read, comments:create|read, profiles:create|read
```

**Usar para:**

- Probar flujos de trabajo de usuario regular
- Creación de contenido propio (posts, comentarios)
- Gestión de perfil
- Operaciones básicas de usuario

### Acceso de Invitado (no autenticado)

Sin necesidad de inicio de sesión:

```
Role:     guest (implícito)
Access:   posts:read, comments:read, profiles:read
```

**Usar para:**

- Navegación de contenido público
- Probar acceso no autenticado
- Operaciones de solo lectura

## Ejemplos de Inicio de Sesión Rápido

```bash
# Inicio de sesión como Administrador del Sistema (admin con acceso completo)
curl -X POST http://localhost:8081/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"system@promenade.com","password":"passw0rd"}' | jq

# Inicio de sesión como Administrador (admin)
curl -X POST http://localhost:8081/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@promenade.com","password":"passw0rd"}' | jq

# Inicio de sesión como Moderador (moderator)
curl -X POST http://localhost:8081/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"moderator@promenade.com","password":"passw0rd"}' | jq

# Inicio de sesión como Usuario Regular (user)
curl -X POST http://localhost:8081/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"alexander.vasilenko@gmail.com","password":"03041965"}' | jq

# Guardar token para reutilización (administrador del sistema)
export TOKEN=$(curl -s -X POST http://localhost:8081/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"system@promenade.com","password":"passw0rd"}' | jq -r '.data.access_token')

# Usar token en solicitudes
curl -X GET http://localhost:8081/api/v1/auth/me \
  -H "Authorization: Bearer $TOKEN" | jq
```

## Probando Permisos RBAC

```bash
# Probar acceso de administrador del sistema (debe funcionar - acceso completo)
curl -X GET http://localhost:8081/api/v1/admin/users \
  -H "Authorization: Bearer $SYSTEM_TOKEN" | jq

# Probar acceso de administrador (debe funcionar - tiene users:*)
curl -X GET http://localhost:8081/api/v1/admin/users \
  -H "Authorization: Bearer $ADMIN_TOKEN" | jq

# Probar acceso de moderador (debe fallar - sin permiso users:list)
curl -X GET http://localhost:8081/api/v1/admin/users \
  -H "Authorization: Bearer $MODERATOR_TOKEN" | jq

# Probar acceso de usuario (debe fallar - sin permisos de admin)
curl -X GET http://localhost:8081/api/v1/admin/users \
  -H "Authorization: Bearer $USER_TOKEN" | jq
```

## Acceso a la Base de Datos

```bash
# Conectar a la base de datos de desarrollo
psql -h localhost -p 5432 -U system -d promenade_dev
# Password: passw0rd

# Ver todos los usuarios con sus roles
SELECT
    u.email,
    u.name,
    r.name as role,
    r.description
FROM users u
LEFT JOIN user_roles ur ON u.id = ur.user_id
LEFT JOIN roles r ON ur.role_id = r.id
ORDER BY u.email;

# Ver permisos de roles
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

## [!] Advertencia de Seguridad

**¡Estas credenciales son SOLO PARA DESARROLLO!**

Antes de implementar en producción:

1. Cambie todas las contraseñas predeterminadas
2. Elimine o desactive cuentas de administrador predeterminadas
3. Use credenciales específicas del entorno
4. Active la gestión adecuada de secrets
5. Configure proveedores de autenticación adecuados

## ¿Necesita Ayuda?

- [Guía de Autorización](docs/AUTHORIZATION.md) - Documentación del sistema RBAC
- [Guía de Pruebas](docs/TESTING_GUIDE.md) - Cómo probar con autenticación
- [Documentación de la API](README.md#api-examples) - Referencia completa de la API
