# Promenade Repository Structure & Licensing

**Last Updated**: January 23, 2026  
**Version**: v0.2.0-dev planning

---

## 🎯 Overview

Promenade follows an **Open Core** business model with repository separation:

- **Core Backend** (public, MIT) - Foundation for everyone
- **UI/UX** (private, commercial) - Protected from theft
- **Enterprise Features** (private, commercial) - Advanced capabilities

This structure balances open-source values with business sustainability and protection from unauthorized use.

---

## 📦 Repository Structure

### 1. `github.com/basilex/promenade` (PUBLIC, MIT)

**What's included**:

- Backend API with RESTful endpoints
- 11 bounded contexts (DDD architecture)
- PostgreSQL database layer (pgx driver)
- Event-driven architecture (bus pattern)
- Complete documentation
- Test suites (unit, smoke, integration)
- Docker Compose for development
- Makefile workflows

**What's NOT included**:

- ❌ Frontend applications
- ❌ Multi-tenancy implementation
- ❌ Bank integrations
- ❌ Advanced analytics

**License**: MIT  
**Access**: Public, anyone can fork/use  
**Version**: v0.1.0-dev (completed), v0.2.0-dev (in progress)

**Purpose**:

- Reference implementation of DDD in Go
- Foundation for custom implementations
- Learning resource
- Community contributions welcome

---

### 2. `github.com/basilex/promenade-ui` (PRIVATE, COMMERCIAL)

**What's included**:

- Next.js 14 Admin Panel (TypeScript, Tailwind CSS)
- Flutter Mobile App (iOS/Android)
- shadcn/ui component library
- Authentication flows
- CRUD interfaces for all contexts
- Analytics dashboards
- Responsive design
- White-label customization

**Technology Stack**:

- Frontend: Next.js 14, React 18, TypeScript
- Mobile: Flutter 3.16+, Dart 3.0+
- State: React Query, Zustand
- UI: Tailwind CSS, shadcn/ui
- Forms: Zod validation
- Charts: Recharts, fl_chart

**License**: Proprietary (Commercial)  
**Access**: Private repository, paid customers only  
**Deployment**: Self-hosted or managed

**Why private?**:

- 💰 Significant investment in UI/UX design
- 🎨 Unique design system and workflows
- 📱 Mobile app development costs
- 🛡️ Protection from clone sites
- 🇺🇦 Ukraine-specific considerations (war, weak IP enforcement)

---

### 3. `github.com/basilex/promenade-enterprise` (PRIVATE, COMMERCIAL)

**What's included**:

- Multi-tenancy module
  - Organization isolation
  - Subscription plans
  - Usage limits and quotas
- Advanced Analytics
  - BI dashboards
  - Custom report builder
  - Data visualization
- Bank Integrations
  - Monobank API
  - PrivatBank API
  - Universal Bank API
  - Automatic reconciliation
- Advanced Billing
  - Usage-based pricing
  - Revenue recognition
  - Dunning management
- Email/SMS Service
  - Transactional emails
  - Marketing campaigns
  - SMS notifications
- Enterprise Support
  - Priority bug fixes
  - Custom development
  - Training and onboarding

**Technology Stack**:

- Backend: Go modules extending core
- Database: PostgreSQL with multi-tenant schemas
- Integrations: REST/SOAP clients
- Notifications: SendGrid, Twilio
- Analytics: ClickHouse, Redis

**License**: Proprietary (Enterprise Agreement)  
**Access**: Private repository, enterprise customers only  
**Deployment**: Kubernetes clusters

**Why private?**:

- 💎 High-value enterprise features
- 🏦 Bank integrations (compliance, certifications)
- 🔐 Multi-tenancy security (critical for SaaS)
- 🇺🇦 Market-specific integrations (Monobank, etc.)
- 🛡️ Prevention of competing SaaS offerings

---

## 🔒 Access Control

### Core Backend (MIT)

- ✅ Anyone can clone, fork, modify
- ✅ No restrictions on commercial use
- ✅ Can use as foundation for own projects
- ✅ Attribution appreciated but not required

### UI/UX (Commercial)

- 🔐 Access via paid subscription ($99/month+)
- 🔐 License key required for build/deployment
- 🔐 Updates and support included
- 🔐 White-label customization available

### Enterprise (Commercial)

- 🔐 Access via enterprise agreement (custom pricing)
- 🔐 Dedicated support channel
- 🔐 Custom SLA (99.9% uptime guaranteed)
- 🔐 Onboarding and training included

---

## 💰 Business Model

### Revenue Streams

**1. UI/UX Subscriptions** ($99-$299/month)

- Professional: $99/mo (up to 3 orgs, web + mobile)
- Business: $199/mo (up to 10 orgs, priority support)
- Premium: $299/mo (unlimited orgs, white-label)

**2. Enterprise Licenses** ($499-$2999/month)

- Starter: $499/mo (basic enterprise features)
- Growth: $999/mo (all features, 5 deployments)
- Enterprise: $2999/mo (unlimited, SLA, dedicated support)

**3. Professional Services**

- Custom development: $75/hour
- Training: $500/day
- Consulting: $100/hour
- Implementation: project-based

**4. Managed Hosting** (future)

- Small: $199/mo (up to 1000 users)
- Medium: $499/mo (up to 10k users)
- Large: custom (unlimited users)

### Free Tier (Core Backend)

- ✅ Builds community
- ✅ Drives adoption
- ✅ Generates leads
- ✅ Showcases architecture quality
- ✅ Attracts contributors

---

## 🎓 Development Workflow

### Contributing to Core Backend (MIT)

**Anyone can contribute**:

1. Fork `promenade` repository
2. Create feature branch
3. Submit pull request
4. Code review by maintainers
5. Merge if approved

**Areas for contributions**:

- Bug fixes
- Performance improvements
- New bounded contexts
- Documentation
- Test coverage
- Language translations

**CLA**: Contributors retain copyright, grant MIT license

### Commercial Modules

**Only paid customers can contribute**:

- Must have active subscription
- Sign Contributor License Agreement (CLA)
- Copyright transfers to maintainer
- Eligible for revenue sharing (if agreed)

---

## 🛡️ Protection Strategy

### Why Separate Repositories?

**Legal protection** (weak in Ukraine):

- ❌ Proprietary license in public repo = easily ignored
- ❌ Terms of Service = not enforceable during war
- ❌ Copyright claims = years in court, expensive

**Technical protection** (strong):

- ✅ Private repo = code not accessible
- ✅ License keys = runtime verification
- ✅ Obfuscation = reverse engineering harder
- ✅ SaaS deployment = no source code exposure

### Additional Protections

**UI/UX Repository**:

- Obfuscated JavaScript builds
- License key verification at runtime
- Regular security audits
- Watermarks in demo versions

**Enterprise Repository**:

- Customer-specific builds
- License server integration
- Usage tracking and analytics
- Automatic license expiration

**Legal Measures** (when possible):

- Trademark registration ("Promenade")
- Copyright registration
- Terms of Service
- DMCA takedown process (for copies)

---

## 📋 Customer Journey

### 1. Discovery

- Find `promenade` on GitHub (MIT)
- Read documentation
- Try core backend locally
- Realize need for UI/UX

### 2. Evaluation

- Request demo access to UI
- 14-day free trial
- Evaluate features
- Check pricing

### 3. Purchase

- Subscribe to Professional ($99/mo)
- Receive access to `promenade-ui`
- Deploy self-hosted or use managed
- Start using product

### 4. Growth

- Need multi-tenancy? Upgrade to Enterprise
- Need bank integration? Add module
- Need custom features? Professional services
- Refer other customers? Revenue share

---

## 🌍 Target Markets

### Primary (Ukraine)

- Small businesses (10-50 employees)
- Mid-market (50-500 employees)
- Industry: retail, wholesale, services

### Secondary (Eastern Europe)

- Poland, Romania, Moldova
- Russian-speaking markets (Belarus, Kazakhstan)
- Similar business practices

### Future (Global)

- English localization
- Multi-currency support
- International bank integrations
- Global deployment

---

## 🔮 Future Plans

### v0.3.0 (Q2 2026)

- More enterprise features
- Additional bank integrations
- Mobile app enhancements
- API marketplace (3rd-party integrations)

### v1.0.0 (Q4 2026)

- Production-ready
- SaaS offering launch
- Partner program
- Reseller network

### Beyond

- Open source advanced features after 2-3 years
- Transition successful modules to MIT
- Build sustainable open-source ecosystem

---

## 📞 Contact & Inquiries

**Sales & Licensing**:  
Email: alexander.vasilenko@gmail.com  
Subject: "Promenade Commercial Inquiry"

**Technical Support** (paid customers):  
Email: support@promenade.dev  
Discord: Private channel (access after purchase)

**Partnerships**:  
Email: partnerships@promenade.dev

**General Questions**:  
GitHub Discussions: `promenade` repository

---

## ⚖️ Legal Disclaimer

This document describes the licensing structure for Promenade Platform.

- Core backend (MIT license) has NO restrictions
- Commercial modules require paid licenses
- Unauthorized redistribution of commercial modules is prohibited
- All prices subject to change
- Terms of Service apply to paid subscriptions

**Important**: This structure is designed to protect commercial investments
while maintaining an open-source core. Given the current situation in Ukraine
(war, weak IP enforcement), technical protection (private repositories) is
prioritized over legal protection (licenses).

---

**Last Updated**: January 23, 2026  
**Next Review**: March 2026 (after v0.2.0 release)
