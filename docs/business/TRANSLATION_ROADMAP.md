# Business Documentation Translation Roadmap

**Multi-language strategy** for business decision-makers with clear separation from technical documentation.

---

## Translation Status

| Priority | Language | ISO Code | File | Status | Lines | Target Date |
|----------|----------|----------|------|--------|-------|-------------|
| 1 | English | EN | BUSINESS_OVERVIEW.md | ✅ Complete | 500+ | Jan 6, 2026 |
| 1 | Ukrainian | UK | BUSINESS_OVERVIEW_UK.md | ✅ Complete | 500+ | Jan 6, 2026 |
| 1 | German | DE | BUSINESS_OVERVIEW_DE.md | ✅ Complete | 622 | Jan 6, 2026 |
| 2 | French | FR | BUSINESS_OVERVIEW_FR.md | 🔄 Planned | ~600 | Q1 2026 |
| 2 | Spanish | ES | BUSINESS_OVERVIEW_ES.md | 🔄 Planned | ~600 | Q1 2026 |
| 3 | Portuguese | PT | BUSINESS_OVERVIEW_PT.md | 🔄 Planned | ~600 | Q2 2026 |
| 3 | Korean | KR | BUSINESS_OVERVIEW_KR.md | 🔄 Planned | ~600 | Q2 2026 |
| 4 | Japanese | JP | BUSINESS_OVERVIEW_JP.md | 🔄 Planned | ~600 | Q2 2026 |
| 4 | Chinese (Trad) | ZH | BUSINESS_OVERVIEW_ZH.md | 🔄 Planned | ~600 | Q2 2026 |

**Total Progress**: 3/9 languages (33%)

---

## Priority Rationale

**Priority 1** (✅ Complete):
- **English**: Universal business language, primary audience
- **Ukrainian**: Developer's native language, local investors/partners
- **German**: Major European economy, strong tech sector

**Priority 2** (Q1 2026):
- **French**: EU official language, strong African market presence
- **Spanish**: 2nd most spoken language globally, Latin America + Spain markets

**Priority 3** (Q2 2026):
- **Portuguese**: Brazil (large emerging market), Portugal (EU gateway)
- **Korean**: Advanced tech market, high startup activity

**Priority 4** (Q2 2026):
- **Japanese**: High-value market, advanced tech adoption
- **Chinese (Traditional)**: Taiwan/Hong Kong markets, tech hubs

---

## Language-Specific Considerations

### French (FR)
- **Market**: France, Belgium, Switzerland, Canada, African countries
- **Business Culture**: Formal tone, emphasis on quality and craftsmanship
- **Terminology**: "Gestion de la relation client" (CRM), "Logiciel en tant que service" (SaaS)
- **Currency**: Keep USD (international), note EUR equivalents in examples

### Spanish (ES)
- **Market**: Spain, Latin America (Mexico, Argentina, Chile, Colombia)
- **Business Culture**: Relationship-focused, personal connections matter
- **Terminology**: "Gestión de relaciones con clientes" (CRM), "Software como servicio" (SaaS)
- **Regional Variants**: Use neutral Spanish (not country-specific)

### Portuguese (PT)
- **Market**: Brazil (primary), Portugal, Angola, Mozambique
- **Business Culture**: Warm, relationship-oriented, growing tech scene
- **Terminology**: "Gestão de relacionamento com o cliente" (CRM), "Software como serviço" (SaaS)
- **Variant**: Brazilian Portuguese (PT-BR) recommended for broader reach

### Korean (KR)
- **Market**: South Korea (advanced tech market, high startup density)
- **Business Culture**: Hierarchical, detail-oriented, high tech adoption
- **Terminology**: "고객 관계 관리" (CRM), "서비스형 소프트웨어" (SaaS)
- **Formality**: Use formal business Korean (존댓말)

### Japanese (JP)
- **Market**: Japan (high-value market, quality-focused)
- **Business Culture**: Detail-oriented, long decision cycles, relationship-based
- **Terminology**: "顧客関係管理" (CRM), "サービスとしてのソフトウェア" (SaaS)
- **Formality**: Use polite business Japanese (敬語)

### Chinese Traditional (ZH)
- **Market**: Taiwan, Hong Kong, Macau (tech hubs, international business)
- **Business Culture**: Pragmatic, network-oriented, fast-paced
- **Terminology**: "客戶關係管理" (CRM), "軟體即服務" (SaaS)
- **Script**: Traditional characters (not Simplified) for target markets

---

## Translation Guidelines

### Content to Translate

**Translate fully**:
- Executive Summary
- Business Value Proposition
- Core Capabilities descriptions
- Use case narratives and company profiles
- Deployment options explanations
- Security & Compliance descriptions
- Roadmap feature descriptions
- Getting Started instructions
- Conclusion and call to action

**Localize moderately**:
- Company names in use cases (use local-sounding names)
- Industry examples (adjust to resonate with target market)
- Contact information (keep email, add local business hours if relevant)

**Keep in English or original**:
- Technical terms: API, REST, JWT, RBAC, PostgreSQL, Redis, etc.
- Product name: "Promenade"
- Currency amounts: USD (add local currency note if helpful)
- URLs and email addresses
- Code snippets and technical specifications
- Version numbers and dates

### Quality Standards

1. **Professional Translation**
   - Use native speakers or professional translation services
   - Business terminology must be appropriate for executives/managers
   - Avoid literal translations that sound unnatural

2. **Cultural Adaptation**
   - Examples should resonate with target market business culture
   - Success metrics framed in culturally relevant terms
   - Relationship to technology reflects local market maturity

3. **Consistency**
   - All translations follow same 15-section structure
   - Metrics and statistics remain identical across languages
   - Technical terms standardized (create glossary)

4. **Verification**
   - Native speaker review required before publication
   - Business professional validation (not just language check)
   - Test with target audience sample if possible

---

## Translation Process

### Step 1: Preparation
1. Create glossary of key terms (CRM, RBAC, etc.) with approved translations
2. Identify native speaker translators or professional service
3. Set up translation tracking (this file)

### Step 2: Translation
1. Translate all 15 sections maintaining structure
2. Localize examples where appropriate
3. Keep technical terms and metrics consistent
4. Add translator credit if external contributor

### Step 3: Review
1. Native speaker business professional review
2. Technical accuracy check (ensure no tech details changed)
3. Cultural appropriateness assessment
4. Compare with English original for completeness

### Step 4: Publication
1. Create file: `docs/business/BUSINESS_OVERVIEW_{ISO}.md`
2. Update `docs/business/README.md` table
3. Update main `README.md` multi-language table
4. Update `docs/INDEX.md` multi-language table
5. Commit with descriptive message
6. Announce to community if open-source contributions

### Step 5: Maintenance
1. Include in monthly review cycle
2. Update when English version updates
3. Sync version numbers across all languages
4. Note language-specific updates in commit messages

---

## Contribution Guidelines

**For Community Translators**:

1. **Check Translation Status** above to see which languages need work
2. **Native Speaker Preferred** - We want natural, professional business language
3. **Follow Structure** - Use existing translations as template
4. **Create Glossary Entry** - Define key terms you translate
5. **Submit Pull Request** with:
   - New `BUSINESS_OVERVIEW_{ISO}.md` file
   - Updated `docs/business/README.md` table
   - Updated main `README.md` table
   - Updated `docs/INDEX.md` table
   - Brief translation notes (if any localization decisions made)

**Translation Credit**:
- Add your name/attribution at bottom of translated document
- Format: "Translated by: [Your Name] (Native [Language] Speaker)"
- Optional: Link to LinkedIn/website if professional translator

**Review Process**:
- Maintainers will review for structural completeness
- Native speaker maintainer (if available) will review language quality
- May request changes or clarifications
- Merge when approved

---

## Technical Documentation Policy

**IMPORTANT**: Technical documentation remains **English only**.

**Rationale**:
- **Developers** are expected to read English technical docs (industry standard)
- **Maintenance burden**: Technical docs update frequently, translation overhead too high
- **Terminology precision**: Technical terms often don't translate well
- **API/code examples**: Must stay in English for copy-paste usability
- **Community**: English enables broader community participation

**English-only technical docs**:
- `/docs/guides/` - All developer guides
- `/docs/concepts/` - Architecture concepts
- `/docs/reference/` - API reference, technical specs
- `/pkg/` READMEs - Package documentation
- `/internal/` READMEs - Internal documentation
- Code comments and inline documentation

**Exception**: README.md main introduction can have brief multi-language section pointing to business docs, but detailed technical content stays English.

---

## Glossary (Template)

Create language-specific glossaries as translations progress:

| English | French | Spanish | Portuguese | Korean | Japanese | Chinese (Trad) |
|---------|--------|---------|------------|--------|----------|----------------|
| CRM | Gestion de la relation client | Gestión de relaciones con clientes | Gestão de relacionamento com o cliente | 고객 관계 관리 | 顧客関係管理 | 客戶關係管理 |
| SaaS | Logiciel en tant que service | Software como servicio | Software como serviço | 서비스형 소프트웨어 | サービスとしてのソフトウェア | 軟體即服務 |
| Order Management | Gestion des commandes | Gestión de pedidos | Gestão de pedidos | 주문 관리 | 注文管理 | 訂單管理 |
| Warehouse | Entrepôt | Almacén | Armazém | 창고 | 倉庫 | 倉庫 |
| Billing | Facturation | Facturación | Faturamento | 청구 | 請求 | 計費 |
| Role-Based Access Control | Contrôle d'accès basé sur les rôles | Control de acceso basado en roles | Controle de acesso baseado em funções | 역할 기반 접근 제어 | ロールベースアクセス制御 | 基於角色的存取控制 |

*(Expand as translations progress)*

---

## Success Metrics

**Translation Quality**:
- Native speaker approval required
- Business professional validation passed
- No technical inaccuracies introduced
- Natural language flow (not robotic translation)

**Coverage**:
- All 15 sections translated completely
- All use cases localized appropriately
- All deployment options explained clearly
- Getting Started section actionable

**Maintenance**:
- Updated within 1 week of English version changes
- Version numbers synced across languages
- Monthly review completed for all languages

**Community Engagement** (if open-source):
- Translation contributions acknowledged
- Language-specific GitHub Discussions/Issues
- Multi-language README badges

---

**Last Updated**: January 6, 2026  
**Current Focus**: French & Spanish (Q1 2026)  
**Maintained by**: Promenade Product Team  
**Contact**: alexander.vasilenko@gmail.com
