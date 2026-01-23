# v0.2.0-dev Week 1 Tasks (January 23-29, 2026)

**Sprint Goal**: Setup foundation for multi-tenancy and frontend development

---

## 🎯 Critical Path Tasks

### 1. Architecture & Planning

- [ ] **ADR-0004: Multi-tenancy Strategy**
  - Document decision: Application-level filtering vs Row-level security
  - Pros/cons analysis
  - Performance implications
  - Security considerations
  - **Assignee**: Backend Lead
  - **Estimate**: 4 hours
  - **Priority**: CRITICAL

- [ ] **Database Schema Design**
  - Design organization table schema
  - Plan migration strategy for existing tables
  - Document foreign key relationships
  - Create ER diagram
  - **Assignee**: Backend Lead
  - **Estimate**: 8 hours
  - **Priority**: CRITICAL

### 2. Multi-tenancy Foundation

- [ ] **Create shared/organization Context**
  - Create directory structure: `internal/contexts/shared/organization/`
  - Define Organization aggregate
    ```go
    type Organization struct {
        aggregate.BaseAggregate
        Name        string
        Slug        string // URL-friendly identifier
        Status      OrganizationStatus
        Settings    jsonstore.Field[OrganizationSettings]
        Limits      OrganizationLimits
        OwnerUserID uuidv7.UUID
    }
    ```
  - **Assignee**: Backend Developer
  - **Estimate**: 6 hours
  - **Priority**: CRITICAL

- [ ] **Organization Repository**
  - Interface definition
  - PostgreSQL implementation
  - CRUD operations
  - Query by slug
  - **Assignee**: Backend Developer
  - **Estimate**: 4 hours
  - **Priority**: CRITICAL

- [ ] **Organization Use Cases**
  - CreateOrganization
  - UpdateOrganization
  - GetOrganization (by ID and by slug)
  - SuspendOrganization
  - DeleteOrganization (soft delete)
  - **Assignee**: Backend Developer
  - **Estimate**: 6 hours
  - **Priority**: HIGH

- [ ] **Organization HTTP Handlers**
  - POST /api/v1/organizations
  - GET /api/v1/organizations/:id
  - PUT /api/v1/organizations/:id
  - DELETE /api/v1/organizations/:id
  - DTOs for request/response
  - **Assignee**: Backend Developer
  - **Estimate**: 4 hours
  - **Priority**: HIGH

### 3. Database Migration

- [ ] **Create Migration Script**
  - File: `migrations/postgres/000013_add_organization_support.up.sql`
  - Create organizations table
  - Add organization_id to all multi-tenant tables
  - Create indexes on organization_id columns
  - **Assignee**: Backend Developer
  - **Estimate**: 8 hours
  - **Priority**: CRITICAL

- [ ] **Migration Testing**
  - Test on fresh database
  - Test on database with existing data
  - Verify rollback works
  - Document manual steps if needed
  - **Assignee**: Backend Developer
  - **Estimate**: 4 hours
  - **Priority**: HIGH

### 4. Frontend Project Setup

- [ ] **Initialize Next.js Project**

  ```bash
  cd front/
  npx create-next-app@14 nextjs --typescript --tailwind --app --src-dir
  cd nextjs
  npm install @tanstack/react-query zustand zod
  npm install @radix-ui/react-dialog @radix-ui/react-dropdown-menu
  npm install lucide-react
  ```

  - **Assignee**: Frontend Developer
  - **Estimate**: 2 hours
  - **Priority**: CRITICAL

- [ ] **Project Structure**

  ```
  front/nextjs/
    src/
      app/              # Next.js 14 App Router
        (auth)/         # Auth group layout
        (dashboard)/    # Dashboard group layout
        api/            # API routes
      components/       # Shared components
        ui/            # shadcn/ui components
      lib/             # Utilities
        api/           # API client
        hooks/         # Custom hooks
        stores/        # Zustand stores
      types/           # TypeScript types
  ```

  - **Assignee**: Frontend Developer
  - **Estimate**: 2 hours
  - **Priority**: HIGH

- [ ] **Setup shadcn/ui**

  ```bash
  npx shadcn-ui@latest init
  npx shadcn-ui@latest add button
  npx shadcn-ui@latest add input
  npx shadcn-ui@latest add form
  npx shadcn-ui@latest add table
  npx shadcn-ui@latest add dialog
  ```

  - **Assignee**: Frontend Developer
  - **Estimate**: 1 hour
  - **Priority**: HIGH

- [ ] **API Client Setup**
  - Create axios instance with interceptors
  - JWT token management
  - Error handling
  - TypeScript types for API responses
  - **Assignee**: Frontend Developer
  - **Estimate**: 4 hours
  - **Priority**: CRITICAL

- [ ] **Authentication Flow**
  - Login page
  - Logout functionality
  - Protected route wrapper
  - Token refresh logic
  - **Assignee**: Frontend Developer
  - **Estimate**: 6 hours
  - **Priority**: CRITICAL

### 5. Development Environment

- [ ] **Update Docker Compose**
  - Add frontend service
  - Add hot reload volume mounts
  - Configure API proxy
  - **Assignee**: DevOps
  - **Estimate**: 2 hours
  - **Priority**: MEDIUM

- [ ] **Update Makefile**
  - Add `make frontend` target
  - Add `make frontend-install` target
  - Add `make frontend-build` target
  - **Assignee**: DevOps
  - **Estimate**: 1 hour
  - **Priority**: LOW

### 6. Testing

- [ ] **Organization Tests**
  - Unit tests for aggregate
  - Integration tests for repository
  - Smoke tests for handlers
  - **Assignee**: Backend Developer
  - **Estimate**: 4 hours
  - **Priority**: HIGH

- [ ] **Frontend Tests Setup**
  - Configure Jest + React Testing Library
  - Example component test
  - Example API mock test
  - **Assignee**: Frontend Developer
  - **Estimate**: 2 hours
  - **Priority**: LOW

### 7. Documentation

- [ ] **Update README.md**
  - Add frontend setup instructions
  - Update architecture diagram
  - Document new API endpoints
  - **Assignee**: Tech Lead
  - **Estimate**: 2 hours
  - **Priority**: MEDIUM

- [ ] **API Documentation**
  - Update Swagger annotations
  - Add organization endpoints
  - Regenerate swagger.json
  - **Assignee**: Backend Developer
  - **Estimate**: 1 hour
  - **Priority**: LOW

---

## 📊 Task Summary

| Category      | Tasks  | Estimate | Priority      |
| ------------- | ------ | -------- | ------------- |
| Architecture  | 2      | 12h      | CRITICAL      |
| Backend       | 8      | 36h      | CRITICAL/HIGH |
| Frontend      | 6      | 17h      | CRITICAL/HIGH |
| DevOps        | 2      | 3h       | MEDIUM/LOW    |
| Testing       | 2      | 6h       | HIGH/LOW      |
| Documentation | 2      | 3h       | MEDIUM/LOW    |
| **TOTAL**     | **22** | **77h**  | -             |

**Team Capacity**: 40h per person (1 week)
**Team Size**: 2 developers (backend + frontend)
**Total Capacity**: 80h
**Utilization**: 96% ✅

---

## 🎯 Definition of Done

**Task is done when**:

- ✅ Code written and working locally
- ✅ Tests written and passing
- ✅ Code reviewed and approved
- ✅ Documentation updated
- ✅ Merged to dev branch

**Sprint is done when**:

- ✅ Organization CRUD working
- ✅ Database migration tested
- ✅ Frontend project running
- ✅ Authentication flow works
- ✅ All tests passing
- ✅ Documentation complete

---

## 🚧 Blockers & Risks

### Identified Risks

1. **Multi-tenancy complexity** - May need more time for schema design
   - Mitigation: Start with ADR and get team buy-in early

2. **Frontend learning curve** - Next.js 14 App Router is new
   - Mitigation: Allocate time for learning/experimentation

3. **Migration on existing data** - Risk of data loss
   - Mitigation: Test thoroughly, create backup procedure

### Dependencies

- Frontend depends on backend API being ready
- Testing depends on migration being complete

---

## 📅 Daily Breakdown

### Monday (Jan 20)

- ✅ Complete pgx migration
- ✅ Create v0.1.0-dev tag
- ✅ Write v0.2.0-dev roadmap

### Tuesday (Jan 21)

- [ ] Write ADR-0004 (morning)
- [ ] Design database schema (afternoon)
- [ ] Initialize Next.js project (frontend dev)

### Wednesday (Jan 22)

- [ ] Create Organization context (backend)
- [ ] Setup shadcn/ui (frontend)
- [ ] Write migration script (backend)

### Thursday (Jan 23)

- [ ] Implement Organization repository (backend)
- [ ] Build API client (frontend)
- [ ] Test migration on dev environment

### Friday (Jan 24)

- [ ] Implement Organization use cases (backend)
- [ ] Build authentication flow (frontend)
- [ ] Write integration tests

---

## 🤝 Team Assignments

### Backend Developer

- Organization context implementation
- Database migration
- Repository and use cases
- Integration tests
- **Load**: 50h estimated

### Frontend Developer

- Next.js project setup
- API client
- Authentication flow
- Component library
- **Load**: 27h estimated

### Tech Lead (Part-time)

- Architecture decisions (ADR)
- Code reviews
- Documentation
- Unblock team
- **Load**: 15h estimated

---

## 📞 Communication

### Daily Standup

- **Time**: 10:00 AM Kyiv
- **Duration**: 15 min
- **Format**: Async in Slack (or sync call)

### Progress Updates

- Post updates in #dev channel
- Blockers escalated immediately
- Demo on Friday afternoon

### Questions

- Use GitHub Discussions
- Tag relevant people
- Response SLA: <4 hours

---

## 🎉 Success Celebration

**If we complete all critical tasks by Friday**:

- 🍕 Team pizza lunch
- 🏆 Recognition in team meeting
- 📸 Screenshot for portfolio

---

**Created**: January 23, 2026  
**Sprint**: Week 1 of v0.2.0-dev  
**Next Review**: January 26, 2026 (mid-week check-in)
