# GOTRS Development Roadmap

*Last updated: December 23, 2025*

## 🚀 Current Status

**Version**: Pre-release (targeting 0.5.0 MVP on January 18, 2026)

GOTRS is a modern, open-source ticketing system built with Go and HTMX, designed as an OTRS-compatible replacement. The core ticketing functionality is complete, and we're in the final stabilization phase before MVP release.

### What Works Today
- ✅ **Agent Interface**: Full ticket management (create, view, update, assign, transfer)
- ✅ **Customer Portal**: Login, ticket submission, viewing, replies, and closure
- ✅ **Email Integration**: Inbound (POP3/IMAP) and outbound with RFC-compliant threading
- ✅ **Database**: Full MySQL/MariaDB and PostgreSQL compatibility
- ✅ **Search**: Ticket search by number or title with pagination
- ✅ **Queue Management**: Real-time statistics and ticket filtering

### What's Missing for MVP
- ⏳ 48-hour stability burn-in test
- ⏳ Staging environment deployment

---

## 📊 Metrics (December 23, 2025)

| Metric | Status | Target |
|--------|--------|--------|
| Core Ticket Functionality | **90%** | 100% |
| Customer Portal | **85%** | Basic |
| Email Integration | **80%** | Basic |
| Admin Modules | ~60% | 80% |
| Test Coverage | ~65% | 50% |
| Production Readiness | **60%** | 70% |
| Days to MVP | **26** | - |

---

## ✅ Delivered (August - December 2025)

### Foundation (August 2025)
- OTRS-compatible database schema (116 tables)
- Docker/Podman containerization with rootless support
- Go/Gin backend with HTMX + Alpine.js frontend
- JWT-based authentication system
- Baseline schema with `make synthesize` credential generation

### Database Compatibility (August - September 2025)
- Full MySQL/MariaDB compatibility via `ConvertPlaceholders` pattern
- PostgreSQL support with RETURNING clauses
- OTRS MySQL database import capability
- 500+ SQL placeholder conversion fixes

### Core Ticketing (September - October 2025)
- Ticket creation with proper number generation (DateChecksum format)
- Ticket zoom with live articles, history, and customer context
- Article system (public and internal replies)
- Status transitions, agent assignment, queue transfer
- Basic search with pagination

### Agent Features (September - November 2025)
- Queue detail pages with real-time statistics
- Ticket history rendering across PostgreSQL/MariaDB
- Reminder toasts with snooze actions
- Per-queue ticket counts and filtering

### Email Integration (November - December 2025)
- RFC-compliant Message-ID, In-Reply-To, References headers
- Outbound threaded notifications on ticket create/reply
- Inbound POP3 connector with postmaster processor
- IMAP connector with folder metadata propagation
- SMTP4Dev integration test suite

### Customer Portal (December 2025)
- Customer login with stateless JWT
- Ticket submission with Tiptap rich text editor
- View own tickets with filtering
- Add replies to own tickets
- Close ticket functionality
- Full i18n support (English and German) for all 12 portal templates

### Quality & Testing
- Template validation at startup
- Comprehensive API/HTMX test suites
- DB-less fallbacks for CI
- Playwright E2E harness

---

## 🎯 MVP Release (0.5.0) - Target: January 18, 2026

### Success Criteria
- ✅ Agents can create and manage tickets
- ✅ Customers can submit tickets via web form
- ✅ Customers can view and reply to their tickets
- ✅ Customers can close their own tickets
- ✅ Basic ticket workflow (new → open → closed)
- ✅ Comments/articles on tickets work
- ✅ Email notifications sent on ticket events
- ✅ Search tickets by number or title
- ✅ 5+ test tickets successfully processed
- ⏹️ System stable for 48 hours without crashes
- ✅ Basic documentation for setup and usage

### Remaining Work (Week 4: January 5-18, 2026)
- ⏹️ Fix critical bugs discovered in testing
- ⏹️ Performance verification (48-hour burn-in)
- ⏹️ Deploy to staging environment

---

## 🔮 Future Roadmap

### Post-MVP Stabilization (Q1 2026)
- Complete admin module audit and fixes
- Comprehensive test coverage (80%+)
- Performance optimization for 1000+ concurrent users
- Complete API documentation
- OTRS migration tools tested with real data

### Enhancement Phase (Q2 2026)
- Advanced reporting and analytics
- Workflow automation engine
- Knowledge base integration
- Multi-language support (i18n) for customer portal
- REST API v2 with GraphQL option
- Kubernetes deployment manifests

### Innovation Phase (2026+)
- Mobile applications (iOS/Android)
- AI-powered ticket classification and routing
- Predictive analytics for SLA management
- Plugin marketplace for extensions
- Enterprise integrations (Salesforce, ServiceNow, Slack)
- Multi-tenant white-label theming

---

## 🏗️ Technical Debt & Known Issues

### Admin Modules Status
| Module | Status | Notes |
|--------|--------|-------|
| Users | 🟡 UI exists | CRUD not fully verified |
| Groups | 🟡 UI exists | CRUD not fully verified |
| Customer Users | ❌ 404 | Not implemented |
| Customer Companies | 🟡 UI exists | POST route restored |
| Queues | ✅ Working | Detail pages with stats |
| Priorities | 🟡 UI exists | CRUD not fully verified |
| States | 🟡 UI exists | CRUD not fully verified |
| Types | ✅ Working | List/create/update/delete |
| SLA | 🟡 UI exists | CRUD not fully verified |
| Services | ❌ 404 | Not implemented |
| Roles/Permissions | ❌ | Not implemented |
| Dynamic Fields | ❌ | Not implemented |
| Templates | ❌ | Not implemented |
| Signatures | ❌ | Not implemented |

### Deferred Items
- Multi-tenant theming (requires schema changes)
- Customer portal i18n
- Inbound email auto-ticket creation rules UI
- SLA engine
- Dynamic fields system
- Bulk ticket actions
- Attachment handling in customer portal

---

## 📈 Success Criteria for 1.0 (Future)

**Production Release (TBD after MVP proven):**
- ⏹️ All core OTRS features implemented
- ⏹️ <200ms response time (p95)
- ⏹️ Support for 1000+ concurrent users
- ⏹️ 80%+ test coverage
- ⏹️ Zero critical security issues
- ⏹️ Complete documentation
- ⏹️ Migration tools tested with real OTRS data
- ⏹️ 5+ production deployments validated

---

## 🤝 How to Contribute

We welcome contributions! Priority areas:
1. Testing and bug reports
2. Documentation improvements
3. Translation (i18n)
4. Frontend UI/UX enhancements
5. Performance optimization

See [CONTRIBUTING.md](CONTRIBUTING.md) for details.

---

## 🎖️ Version History

| Version | Date | Milestone |
|---------|------|-----------|
| 0.1.0 | Aug 17, 2025 | Database schema, initial structure |
| 0.2.0 | Aug 24, 2025 | Admin UI scaffolding |
| 0.3.0 | Aug 27, 2025 | Schema migrations, MySQL compatibility |
| 0.4.0 | Oct 20, 2025 | Ticket vertical slice, queue automation |
| 0.5.0 | Jan 18, 2026 | **MVP Target** - Full ticketing system |

---

## 📜 Historical Development Log

<details>
<summary>Click to expand detailed development history</summary>

### December 2025
- Customer portal: login, ticket CRUD, replies, close functionality
- Inbound email: POP3/IMAP connectors, postmaster processor
- SMTP4Dev integration test suite
- Admin mail account poll status API

### November 2025
- Email threading support (RFC-compliant headers)
- Outbound notifications on ticket create/reply

### October 2025
- Ticket creation vertical slice complete
- Real ticket zoom with live articles
- Ticket number generator (DateChecksum format)
- Status transitions and assignment workflow
- Queue statistics and filtering

### September 2025
- MySQL compatibility layer (ConvertPlaceholders)
- Template system robustness
- Code quality improvements
- Queue detail enhancements

### August 2025
- Foundation: schema, Docker, authentication
- OTRS import capability
- Dynamic credential generation

</details>
