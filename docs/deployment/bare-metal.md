# Bare Metal Installation Guide

## Coming Soon

This guide will provide comprehensive instructions for installing GOTRS directly on Linux servers without containerization.

## Planned Content

- System requirements
- Supported operating systems (Ubuntu, RHEL, Debian, CentOS)
- Prerequisites installation
  - Go runtime
  - PostgreSQL database
  - Redis cache
  - Nginx reverse proxy
- Building from source
- Binary installation
- System service configuration (systemd)
- Database setup and migrations
- Configuration files
- User and permission setup
- Firewall configuration
- SSL/TLS certificate setup
- Log rotation
- Backup configuration
- Monitoring setup
- Performance tuning
- Clustering and high availability
- Upgrade procedures
- Troubleshooting

## Quick Preview

```bash
# Install prerequisites
sudo apt-get update
sudo apt-get install postgresql redis nginx

# Download and install GOTRS
wget https://github.com/gotrs/gotrs/releases/latest/gotrs-linux-amd64.tar.gz
tar -xzf gotrs-linux-amd64.tar.gz
sudo mv gotrs /usr/local/bin/

# Initialize database
gotrs migrate up

# Start service
sudo systemctl start gotrs
sudo systemctl enable gotrs
```

## System Requirements

- **OS**: Linux (64-bit)
- **RAM**: Minimum 4GB, Recommended 8GB+
- **CPU**: 2+ cores
- **Disk**: 20GB+ available space
- **Database**: PostgreSQL 14+
- **Cache**: Redis 6+

## See Also

- [Docker Deployment](docker.md)
- [Kubernetes Deployment](kubernetes.md)
- [Architecture Overview](../../ARCHITECTURE.md)

---

*Full documentation coming soon. For development setup, see [docs/development/MVP.md](../development/MVP.md)*