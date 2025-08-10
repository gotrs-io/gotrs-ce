# GOTRS Quick Start Guide

## Coming Soon

This guide will help you get GOTRS up and running quickly for evaluation and development.

## Planned Content

- Prerequisites check
- Installation options overview
- Docker Compose quick start
- Initial configuration
- Creating your first admin user
- Basic system configuration
- Email setup
- Creating queues and groups
- Adding agents and customers
- Creating your first ticket
- Basic workflow configuration
- Testing email integration
- Accessing different portals (admin, agent, customer)
- Basic troubleshooting
- Next steps

## Quick Preview

### 1. Start with Docker Compose

```bash
# Clone repository
git clone https://github.com/gotrs/gotrs.git
cd gotrs

# Copy environment template
cp .env.example .env

# Start services
docker-compose up -d

# Check status
docker-compose ps
```

### 2. Access GOTRS

- **Web Interface**: http://localhost:8080
- **API Documentation**: http://localhost:8080/api/docs
- **Default Admin**: admin@localhost / admin123

### 3. Initial Setup

1. Log in as admin
2. Configure email settings
3. Create your first queue
4. Add agents
5. Create a test ticket

## Available Guides

- [Docker Deployment](../deployment/docker.md)
- [Administrator Guide](../admin-guide/README.md)
- [Agent Manual](../agent-manual/README.md)

## Demo Instance

Try GOTRS without installation:
- URL: https://try.gotrs.io
- See [Demo Guide](../DEMO.md) for credentials

## Getting Help

- GitHub Issues: [Report problems]
- Discord: [Community support]
- Documentation: [Full docs]

---

*Full quick start guide coming soon. For now, see [MVP Development Guide](../development/MVP.md)*