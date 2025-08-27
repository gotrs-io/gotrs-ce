# GOTRS Architecture

## Overview

GOTRS employs a modular, microservices-based architecture designed for scalability, maintainability, and security. The system evolves from an initial modular monolith (MVP) to a full microservices architecture as scaling needs arise.

## Architecture Evolution

### Phase 1: Modular Monolith (MVP)
```
┌─────────────────────────────────────────────┐
│             Load Balancer                    │
└─────────────────┬───────────────────────────┘
                  │
┌─────────────────▼───────────────────────────┐
│           GOTRS Application                  │
│  ┌────────────────────────────────────┐    │
│  │         Web Server (Gin)           │    │
│  └────────────────┬───────────────────┘    │
│  ┌────────────────▼───────────────────┐    │
│  │         Service Layer              │    │
│  │  ┌──────┐ ┌──────┐ ┌──────┐      │    │
│  │  │Auth  │ │Ticket│ │User  │ ...  │    │
│  │  └──────┘ └──────┘ └──────┘      │    │
│  └────────────────┬───────────────────┘    │
│  ┌────────────────▼───────────────────┐    │
│  │      Data Access Layer             │    │
│  └────────────────────────────────────┘    │
└─────────────────────────────────────────────┘
           │              │
    ┌──────▼──────┐ ┌─────▼─────┐
    │ PostgreSQL  │ │   Valkey   │
    └─────────────┘ └───────────┘
```

### Phase 2: Microservices Architecture (Production)
```
┌─────────────────────────────────────────────┐
│            API Gateway (Kong/Traefik)        │
└───────┬─────────┬──────────┬────────────────┘
        │         │          │
┌───────▼──┐ ┌───▼───┐ ┌────▼────┐
│   Auth   │ │Ticket │ │  User   │
│ Service  │ │Service│ │ Service │
└───────┬──┘ └───┬───┘ └────┬────┘
        │         │          │
┌───────▼─────────▼──────────▼────┐
│       Message Bus (NATS)        │
└───────┬─────────┬────────────────┘
        │         │
┌───────▼──┐ ┌───▼────────┐
│Workflow  │ │Notification│
│ Engine   │ │  Service   │
└──────────┘ └────────────┘
```

## Core Components

### 1. API Gateway
- **Technology**: Kong or Traefik
- **Responsibilities**:
  - Request routing
  - Authentication/authorization
  - Rate limiting
  - SSL termination
  - Request/response transformation
  - API versioning

### 2. Core Services

#### Auth Service
```go
// Handles all authentication and authorization
type AuthService interface {
    Login(credentials) (*Token, error)
    Validate(token) (*Claims, error)
    Refresh(refreshToken) (*Token, error)
    Logout(token) error
    // OAuth2, SAML, LDAP integrations
}
```

#### Ticket Service
```go
// Core ticketing functionality
type TicketService interface {
    Create(ticket) (*Ticket, error)
    Update(id, updates) (*Ticket, error)
    Get(id) (*Ticket, error)
    List(filters) ([]*Ticket, error)
    AddComment(ticketId, comment) error
    ChangeStatus(ticketId, status) error
}
```

#### User Service
```go
// User and organization management
type UserService interface {
    CreateUser(user) (*User, error)
    GetUser(id) (*User, error)
    UpdateUser(id, updates) (*User, error)
    AssignRole(userId, role) error
    ManageOrganization(org) error
}
```

### 3. Data Layer

#### Primary Database (PostgreSQL)
```sql
-- OTRS-Compatible Schema (100% compatibility maintained)
-- Uses INTEGER primary keys, NOT UUIDs
ticket
├── id (BIGSERIAL)           -- BIGINT auto-increment
├── tn (VARCHAR(50))         -- Ticket number, unique
├── title (VARCHAR(255))
├── ticket_state_id (SMALLINT, FK)
├── ticket_priority_id (SMALLINT, FK)
├── queue_id (INTEGER, FK)
├── customer_id (VARCHAR(150))      -- Company ID
├── customer_user_id (VARCHAR(250)) -- Customer email/login
├── user_id (INTEGER, FK)           -- Owner
├── responsible_user_id (INTEGER)   -- Assigned agent
├── create_time (TIMESTAMP)
├── change_time (TIMESTAMP)
└── (116 total OTRS tables)

-- Indexes for performance
CREATE INDEX idx_ticket_state_id ON ticket(ticket_state_id);
CREATE INDEX idx_ticket_queue_id ON ticket(queue_id);
CREATE INDEX idx_ticket_customer_id ON ticket(customer_id);
CREATE INDEX idx_ticket_create_time ON ticket(create_time DESC);
```

#### Cache Layer (Valkey)
```
Key Patterns:
- session:{session_id} - User sessions
- cache:ticket:{id} - Ticket cache
- rate:{ip} - Rate limiting
- queue:{name} - Job queues
- pubsub:{channel} - Real-time events
```

#### Search Layer (Zinc)
```
Elasticsearch-compatible API:
- Full-text search across tickets
- Faceted search and filtering
- Auto-complete suggestions
- Index: gotrs_tickets
- Compatible with ES clients
```

#### Workflow Engine (Temporal)
```
Workflow Patterns:
- Ticket lifecycle management
- SLA enforcement
- Escalation rules
- Automated notifications
- Approval chains
```

### 4. Communication Patterns

#### Synchronous Communication
- REST API for JSON data exchange
- HTMX endpoints returning HTML fragments
- gRPC for internal service-to-service calls

#### Asynchronous Communication
- Temporal for workflow orchestration
- Server-Sent Events (SSE) for real-time updates
- Background job processing via Temporal activities

### 5. Security Architecture

```yaml
Security Layers:
  Network:
    - TLS 1.3 for all communications
    - Network segmentation
    - Firewall rules
    
  Application:
    - JWT with short TTL
    - API key management
    - Rate limiting per endpoint
    - Input validation
    
  Data:
    - Encryption at rest
    - Field-level encryption for PII
    - Audit logging
    
  Access Control:
    - RBAC with fine-grained permissions
    - Multi-factor authentication
    - Session management
```

## Deployment Architecture

### Development Environment
```yaml
# docker-compose.yml (Podman-compatible, no version specified)
services:
  backend:
    build: .
    ports:
      - "8080:8080"
    environment:
      - ENV=development
    volumes:
      - ./:/app
      
  postgres:
    image: postgres:15-alpine
    environment:
      - POSTGRES_DB=gotrs
      
  valkey:
    image: valkey/valkey:7-alpine
    
  temporal:
    image: temporalio/auto-setup:1.22.4
    ports:
      - "7233:7233"
      
  zinc:
    image: public.ecr.aws/zinclabs/zinc:0.4.9
    ports:
      - "4080:4080"
    environment:
      - ZINC_FIRST_ADMIN_USER=admin
      - ZINC_FIRST_ADMIN_PASSWORD=admin
      
  mailhog:
    image: mailhog/mailhog:latest
    ports:
      - "8025:8025"
```

### Production Kubernetes
```yaml
# Kubernetes deployment structure
namespaces/
├── gotrs-prod/
│   ├── deployments/
│   │   ├── auth-service
│   │   ├── ticket-service
│   │   └── user-service
│   ├── services/
│   ├── configmaps/
│   └── secrets/
└── gotrs-monitoring/
    ├── prometheus
    ├── grafana
    └── elasticsearch
```

## Scalability Patterns

### Horizontal Scaling
```
┌──────────────┐
│Load Balancer │
└──────┬───────┘
       │
┌──────▼───────────────────────┐
│   Service Instances (N)      │
│ ┌────┐ ┌────┐ ┌────┐ ┌────┐│
│ │Pod1│ │Pod2│ │Pod3│ │PodN││
│ └────┘ └────┘ └────┘ └────┘│
└──────────────────────────────┘
```

### Database Scaling
- Read replicas for query distribution
- Connection pooling (pgBouncer)
- Partitioning for large tables
- Archival strategy for old tickets

### Caching Strategy
1. **Application Cache**: In-memory caching
2. **Valkey Cache**: Shared cache across instances
3. **CDN**: Static assets and attachments
4. **Database Query Cache**: PostgreSQL query optimization

## Monitoring and Observability

### Metrics (Prometheus)
```go
// Key metrics to track
var (
    RequestDuration = prometheus.NewHistogramVec(...)
    RequestCount = prometheus.NewCounterVec(...)
    ErrorRate = prometheus.NewGaugeVec(...)
    QueueDepth = prometheus.NewGaugeVec(...)
)
```

### Logging (ELK Stack)
```json
{
  "timestamp": "2024-01-15T10:30:00Z",
  "level": "INFO",
  "service": "ticket-service",
  "trace_id": "abc123",
  "user_id": "user456",
  "action": "ticket.create",
  "duration_ms": 45
}
```

### Tracing (OpenTelemetry)
- Distributed tracing across services
- Performance bottleneck identification
- Request flow visualization

## Technology Stack

### Backend
- **Language**: Go 1.21+
- **Web Framework**: Gin or Fiber
- **ORM**: GORM with raw SQL for complex queries
- **Authentication**: JWT with refresh tokens
- **Validation**: go-playground/validator

### Frontend
- **Framework**: HTMX for hypermedia-driven architecture
- **JavaScript**: Alpine.js for lightweight interactivity
- **CSS**: Tailwind CSS with standalone CLI
- **Templates**: Go HTML templates with layouts
- **Real-time**: Server-Sent Events (SSE)
- **Build Tool**: Tailwind CLI (no JavaScript bundler needed)
- **Testing**: Go template tests + E2E tests

### Infrastructure
- **Container Runtime**: Docker or Podman (auto-detected)
- **Orchestration**: Kubernetes/OpenShift
- **CI/CD**: GitHub Actions
- **Monitoring**: Prometheus + Grafana
- **Logging**: ELK Stack

## Development Principles

1. **Domain-Driven Design**: Clear bounded contexts
2. **Clean Architecture**: Separation of concerns
3. **12-Factor App**: Cloud-native best practices
4. **API-First**: Design APIs before implementation
5. **Test-Driven Development**: Comprehensive test coverage
6. **Security by Design**: Security at every layer

## Performance Targets

- **API Response Time**: < 200ms (p95)
- **Page Load Time**: < 2 seconds
- **Concurrent Users**: 10,000+
- **Tickets/Second**: 100+ creates
- **Uptime**: 99.9% SLA
- **RPO**: < 1 hour
- **RTO**: < 4 hours

## Extension Architecture

### Plugin System
```go
type Plugin interface {
    Name() string
    Version() string
    Init(container *Container) error
    RegisterRoutes(router *gin.RouterGroup)
    RegisterHooks(hooks *HookRegistry)
    Shutdown() error
}
```

### Extension Points
- Custom ticket fields
- Workflow actions
- Notification channels
- Report generators
- Authentication providers
- Storage backends

## Future Considerations

1. **Multi-tenancy**: Isolated customer environments
2. **Edge Computing**: Regional deployments
3. **AI Integration**: ML models for ticket classification
4. **Blockchain**: Audit trail integrity
5. **IoT Support**: Device monitoring integration
6. **Voice Integration**: Phone support channel