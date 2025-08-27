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

### 1. Start with Docker/Podman Compose

```bash
# Clone repository
git clone https://github.com/gotrs-io/gotrs-ce.git
cd gotrs-ce

# Copy environment template
cp .env.example .env

# Start services (auto-detects docker/podman)
make up

# Or manually:
# docker-compose up -d
# podman-compose up -d

# Check status
make logs
```

### 2. Access GOTRS

- **Frontend**: http://localhost
- **Backend API**: http://localhost:8080/api/v1/status
- **API Documentation**: http://localhost:8080/api/docs (coming soon)
- **Mailhog (Email Testing)**: http://localhost:8025
- **Database Admin**: http://localhost:8090 (Adminer, optional)

### 3. Development Environment

Once services are running:

```bash
# View logs
make logs

# Stop services
make down

# Restart services
make up

# Clean restart (removes volumes)
make clean && make up
```

## Troubleshooting

### Common Issues

#### Permission Denied Errors

If you see permission errors:

```bash
# Fix file ownership (rarely needed)
sudo chown -R $USER:$USER ./

# Then restart containers
make down && make up
```

**Note**: This is typically only needed if files were created by root or copied from another system.

#### Docker Hub Authentication Issues

If you see `unauthorized: incorrect username or password`:

```bash
# Pull images manually first
podman pull docker.io/library/postgres:15-alpine
podman pull docker.io/valkey/valkey:7-alpine
podman pull docker.io/library/nginx:alpine

# Then start services
make up
```

#### Container Build Failures

```bash
# Clean rebuild all containers
make clean
podman system prune -f
make up
```

#### Port Already in Use

```bash
# Check what's using the port
sudo netstat -tulpn | grep :3000

# Stop conflicting services or change ports in .env
```

#### Frontend Not Loading

```bash
# Check nginx container status
podman logs gotrs-nginx

# Check backend container status
podman logs gotrs-backend

# Restart services
make down && make up
```

### Development Setup Verification

Check if everything is working:

```bash
# Health checks
curl http://localhost:8080/health          # Backend health
curl http://localhost                     # Frontend app
curl http://localhost:8025                # Mailhog UI

# Database connection
make db-shell  # Should connect to PostgreSQL
```

## Available Guides

- [Docker Deployment](../deployment/docker.md)
- [Administrator Guide](../admin-guide/README.md)
- [Agent Manual](../agent-manual/README.md)
- [Podman Setup](../PODMAN.md)

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