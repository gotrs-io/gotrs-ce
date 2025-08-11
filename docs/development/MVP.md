# GOTRS MVP Implementation Plan

## Overview

This document outlines the Minimum Viable Product (MVP) implementation for GOTRS, focusing on delivering core ticketing functionality within 6-8 weeks.

## MVP Scope

### Core Requirements
- ✅ Basic ticket CRUD operations
- ✅ User authentication and authorization
- ✅ Email integration (send/receive)
- ✅ Simple workflow (New → Open → Resolved → Closed)
- ✅ Agent and customer portals
- ✅ Basic search functionality
- ✅ Docker deployment

### Explicitly Excluded from MVP
- ❌ Complex workflows
- ❌ Advanced reporting
- ❌ Multiple languages
- ❌ Plugin system
- ❌ Mobile apps
- ❌ AI/ML features
- ❌ ITSM modules

## Technical Stack (Simplified)

### Backend
```yaml
language: Go 1.21
framework: Gin (HTTP router)
database: PostgreSQL 14
cache: Valkey 7 (sessions only)
email: SMTP/IMAP
auth: JWT with refresh tokens
```

### Frontend
```yaml
framework: React 18
ui_library: Material-UI
state: Redux Toolkit
build: Vite
styling: CSS Modules
```

### Infrastructure
```yaml
container: Docker
orchestration: Docker Compose
reverse_proxy: Nginx
monitoring: Basic health checks
```

## Project Structure

```
gotrs/
├── cmd/
│   └── server/
│       └── main.go          # Application entry point
├── internal/
│   ├── api/
│   │   ├── handlers/        # HTTP handlers
│   │   ├── middleware/      # Auth, CORS, logging
│   │   └── routes/          # Route definitions
│   ├── core/
│   │   ├── tickets/         # Ticket domain logic
│   │   ├── users/           # User domain logic
│   │   └── auth/            # Authentication logic
│   ├── data/
│   │   ├── postgres/        # Database implementation
│   │   └── valkey/          # Cache implementation
│   └── services/
│       ├── email/           # Email service
│       └── notification/    # Notification service
├── web/
│   ├── src/
│   │   ├── components/      # React components
│   │   ├── pages/          # Page components
│   │   ├── store/          # Redux store
│   │   └── services/       # API clients
│   └── public/             # Static assets
├── migrations/             # Database migrations
├── configs/               # Configuration files
├── docker/                # Docker configurations
└── docs/                  # Documentation
```

## Implementation Phases

### Week 1-2: Foundation

#### Database Schema
```sql
-- Core tables only
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    role VARCHAR(50) NOT NULL, -- admin, agent, customer
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE tickets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    number BIGSERIAL UNIQUE NOT NULL,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    status VARCHAR(50) DEFAULT 'new',
    priority VARCHAR(50) DEFAULT 'normal',
    customer_id UUID REFERENCES users(id),
    agent_id UUID REFERENCES users(id),
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE ticket_messages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    ticket_id UUID REFERENCES tickets(id) ON DELETE CASCADE,
    user_id UUID REFERENCES users(id),
    message TEXT NOT NULL,
    is_internal BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE attachments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    ticket_id UUID REFERENCES tickets(id) ON DELETE CASCADE,
    message_id UUID REFERENCES ticket_messages(id),
    filename VARCHAR(255) NOT NULL,
    size INTEGER,
    content_type VARCHAR(100),
    storage_path TEXT,
    created_at TIMESTAMP DEFAULT NOW()
);
```

#### API Endpoints
```yaml
# Authentication
POST   /api/auth/register
POST   /api/auth/login
POST   /api/auth/refresh
POST   /api/auth/logout

# Tickets
GET    /api/tickets          # List tickets
POST   /api/tickets          # Create ticket
GET    /api/tickets/:id      # Get ticket
PUT    /api/tickets/:id      # Update ticket
DELETE /api/tickets/:id      # Delete ticket

# Messages
GET    /api/tickets/:id/messages    # List messages
POST   /api/tickets/:id/messages    # Add message

# Users
GET    /api/users            # List users
GET    /api/users/:id        # Get user
PUT    /api/users/:id        # Update user

# Attachments
POST   /api/attachments      # Upload file
GET    /api/attachments/:id  # Download file
```

### Week 3-4: Core Features

#### Authentication Implementation
```go
// JWT token structure
type TokenClaims struct {
    UserID string `json:"user_id"`
    Email  string `json:"email"`
    Role   string `json:"role"`
    jwt.StandardClaims
}

// Basic auth middleware
func AuthMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        token := c.GetHeader("Authorization")
        if token == "" {
            c.AbortWithStatus(401)
            return
        }
        
        claims, err := ValidateToken(token)
        if err != nil {
            c.AbortWithStatus(401)
            return
        }
        
        c.Set("user", claims)
        c.Next()
    }
}
```

#### Ticket Service
```go
type TicketService struct {
    db *sql.DB
}

func (s *TicketService) Create(ticket *Ticket) error {
    query := `
        INSERT INTO tickets (title, description, customer_id, priority)
        VALUES ($1, $2, $3, $4)
        RETURNING id, number, created_at
    `
    err := s.db.QueryRow(query, 
        ticket.Title, 
        ticket.Description,
        ticket.CustomerID,
        ticket.Priority,
    ).Scan(&ticket.ID, &ticket.Number, &ticket.CreatedAt)
    return err
}

func (s *TicketService) List(filters TicketFilters) ([]Ticket, error) {
    // Basic filtering by status, agent, customer
    // Pagination support
    // Simple sorting
}
```

### Week 5-6: Frontend & Integration

#### React Components Structure
```typescript
// Core components
components/
├── Layout/
│   ├── Header.tsx
│   ├── Sidebar.tsx
│   └── Footer.tsx
├── Tickets/
│   ├── TicketList.tsx
│   ├── TicketDetail.tsx
│   ├── TicketForm.tsx
│   └── TicketMessage.tsx
├── Auth/
│   ├── LoginForm.tsx
│   ├── RegisterForm.tsx
│   └── ProtectedRoute.tsx
└── Common/
    ├── Loading.tsx
    ├── ErrorBoundary.tsx
    └── Pagination.tsx
```

#### State Management
```typescript
// Redux store structure
interface AppState {
    auth: {
        user: User | null;
        token: string | null;
        loading: boolean;
    };
    tickets: {
        list: Ticket[];
        current: Ticket | null;
        loading: boolean;
        filters: TicketFilters;
    };
    ui: {
        sidebar: boolean;
        theme: 'light' | 'dark';
    };
}
```

### Docker Configuration

```dockerfile
# Dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o gotrs cmd/server/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/gotrs .
COPY --from=builder /app/web/dist ./web/dist
EXPOSE 8080
CMD ["./gotrs"]
```

```yaml
# docker-compose.yml
version: '3.8'

services:
  app:
    build: .
    ports:
      - "8080:8080"
    environment:
      DATABASE_URL: postgres://gotrs:password@db:5432/gotrs
      VALKEY_URL: valkey://cache:6379
      JWT_SECRET: ${JWT_SECRET}
    depends_on:
      - db
      - cache
      
  db:
    image: postgres:14-alpine
    environment:
      POSTGRES_DB: gotrs
      POSTGRES_USER: gotrs
      POSTGRES_PASSWORD: password
    volumes:
      - postgres_data:/var/lib/postgresql/data
      
  cache:
    image: valkey:7-alpine
    
  nginx:
    image: nginx:alpine
    ports:
      - "80:80"
    volumes:
      - ./nginx.conf:/etc/nginx/nginx.conf
    depends_on:
      - app

volumes:
  postgres_data:
```

## Development Workflow

### Day 1-3: Setup
```bash
# Initialize project
go mod init github.com/gotrs/gotrs
npm create vite@latest web -- --template react-ts

# Install dependencies
go get github.com/gin-gonic/gin
go get github.com/golang-jwt/jwt/v5
go get github.com/lib/pq
go get github.com/valkey-io/valkey-go

# Frontend dependencies
cd web && npm install @mui/material @reduxjs/toolkit react-redux axios
```

### Day 4-10: Backend Development
- [ ] Database schema and migrations
- [ ] User authentication endpoints
- [ ] Ticket CRUD operations
- [ ] Email service integration
- [ ] Basic validation and error handling

### Day 11-20: Frontend Development
- [ ] Authentication flow
- [ ] Agent dashboard
- [ ] Ticket management UI
- [ ] Customer portal
- [ ] Basic search and filters

### Day 21-25: Integration
- [ ] Connect frontend to backend
- [ ] Email notifications
- [ ] File uploads
- [ ] Real-time updates (basic polling)

### Day 26-30: Testing & Deployment
- [ ] Unit tests for critical paths
- [ ] Integration tests
- [ ] Docker configuration
- [ ] Basic documentation
- [ ] Demo data generation

## MVP Features Checklist

### Authentication
- [x] User registration
- [x] Email/password login
- [x] JWT tokens
- [x] Role-based access (admin, agent, customer)
- [x] Password reset via email

### Tickets
- [x] Create ticket
- [x] View ticket list
- [x] View ticket details
- [x] Update ticket status
- [x] Add messages/comments
- [x] Assign to agent
- [x] Set priority
- [x] Basic search

### Email
- [x] Send ticket notifications
- [x] Receive tickets via email
- [x] Reply to ticket via email

### UI/UX
- [x] Responsive design
- [x] Agent dashboard
- [x] Customer portal
- [x] Login/logout
- [x] Basic theming

### Deployment
- [x] Docker containers
- [x] Docker Compose setup
- [x] Environment configuration
- [x] Basic health checks

## Success Criteria

### Functional
- Can create and manage tickets
- Email notifications work
- Users can login and see appropriate data
- Basic workflow functions

### Performance
- Page loads < 3 seconds
- API responses < 500ms
- Supports 100 concurrent users

### Quality
- No critical bugs
- Core features have tests
- Documentation exists
- Deployable with Docker

## Post-MVP Priorities

1. **Advanced Search**: Full-text search, saved filters
2. **Reporting**: Basic metrics and charts
3. **SLA Management**: Response time tracking
4. **Workflow Automation**: Triggers and actions
5. **API Documentation**: OpenAPI specification
6. **Performance Optimization**: Caching, query optimization
7. **Security Hardening**: Rate limiting, audit logs

## Risk Mitigation

### Technical Risks
- **Scope Creep**: Strictly enforce MVP boundaries
- **Integration Issues**: Test email early
- **Performance**: Use simple queries, add indexes
- **Security**: Use established libraries, no custom crypto

### Timeline Risks
- **Delays**: Have weekly checkpoints
- **Blockers**: Quick decisions, document for later
- **Testing**: Automated tests for critical paths only

## Resources

### Documentation
- [Gin Web Framework](https://gin-gonic.com/)
- [GORM](https://gorm.io/) (if using ORM)
- [React Documentation](https://react.dev/)
- [Material-UI](https://mui.com/)

### Tools
- **API Testing**: Postman/Insomnia
- **Database**: TablePlus/pgAdmin
- **Monitoring**: Health endpoint only
- **Logging**: Structured JSON logs

---

*MVP: Deliver value quickly, iterate based on feedback*
*Last updated: August 2025*