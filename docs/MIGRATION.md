# OTRS to GOTRS Migration Guide

## Overview

GOTRS provides comprehensive migration tools and compatibility layers to ensure a smooth transition from OTRS. Our migration strategy prioritizes data integrity, minimal downtime, and preservation of all ticket history.

## Migration Compatibility

### Supported OTRS Versions
- ✅ OTRS 6.x (Full support)
- ✅ OTRS 5.x (Full support)
- ⚠️ OTRS 4.x (Limited support, requires pre-migration upgrade)
- ⚠️ OTRS 3.x (Requires intermediate migration to OTRS 5+)

### Supported Databases
- PostgreSQL 9.6+
- MySQL 5.7+ / MariaDB 10.2+
- Oracle 12c+ (Enterprise Edition only)

## Pre-Migration Checklist

### 1. System Requirements
- [ ] GOTRS system requirements met
- [ ] Sufficient disk space (2x current OTRS database size)
- [ ] Database backup completed
- [ ] Network connectivity between OTRS and GOTRS servers
- [ ] Migration tool downloaded and verified

### 2. OTRS Preparation
- [ ] OTRS system health check passed
- [ ] All OTRS packages documented
- [ ] Custom fields identified
- [ ] Workflows documented
- [ ] Integrations cataloged
- [ ] User permissions exported

### 3. GOTRS Preparation
- [ ] GOTRS installed and tested
- [ ] Database connection verified
- [ ] Email configuration ready
- [ ] Storage configuration completed
- [ ] SSL certificates installed

## Migration Process

### Phase 1: Analysis & Planning

#### 1.1 Run Migration Analyzer
```bash
./gotrs-migrate analyze \
  --source-type otrs \
  --source-db "postgres://otrs_user:pass@otrs-server/otrs" \
  --report-output migration-report.html
```

#### 1.2 Review Migration Report
```
Migration Analysis Report
========================
Total Tickets: 45,328
Total Users: 1,247
Total Customers: 5,823
Total Attachments: 127,491 (42.3 GB)
Custom Fields: 23
Workflows: 12
Estimated Migration Time: 2.5 hours
Compatibility Score: 98%

Warnings:
- 2 custom modules require manual migration
- 15 users have duplicate email addresses
- 3 workflows use unsupported conditions
```

### Phase 2: Test Migration

#### 2.1 Create Test Environment
```bash
# Clone OTRS database for testing
pg_dump otrs_production > otrs_test.sql
psql otrs_test < otrs_test.sql

# Set up test GOTRS instance
docker-compose -f docker-compose.test.yml up -d
```

#### 2.2 Run Test Migration
```bash
./gotrs-migrate test \
  --source-db "postgres://user:pass@localhost/otrs_test" \
  --target-db "postgres://user:pass@localhost/gotrs_test" \
  --validate
```

#### 2.3 Validation Tests
```bash
# Run automated validation
./gotrs-migrate validate \
  --target-db "postgres://user:pass@localhost/gotrs_test" \
  --validation-level comprehensive

# Manual validation checklist
- [ ] Random ticket content verification (10 samples)
- [ ] User login testing
- [ ] Permission verification
- [ ] Attachment accessibility
- [ ] Workflow execution
- [ ] Report generation
```

### Phase 3: Production Migration

#### 3.1 Maintenance Mode
```bash
# Enable OTRS maintenance mode
su - otrs
./bin/otrs.Console.pl Maint::Config::Set MaintenanceMode 1

# Display maintenance message
./bin/otrs.Console.pl Maint::Config::Set MaintenanceMessage \
  "System migration in progress. Expected completion: 2 hours"
```

#### 3.2 Final Backup
```bash
# Backup OTRS database
pg_dump otrs_production > otrs_final_backup_$(date +%Y%m%d_%H%M%S).sql

# Backup OTRS attachments
tar -czf otrs_attachments_$(date +%Y%m%d_%H%M%S).tar.gz /opt/otrs/var/article
```

#### 3.3 Execute Migration
```bash
./gotrs-migrate run \
  --source-db "postgres://otrs_user:pass@otrs-server/otrs" \
  --target-db "postgres://gotrs_user:pass@gotrs-server/gotrs" \
  --batch-size 1000 \
  --parallel-workers 4 \
  --progress \
  --log-file migration_$(date +%Y%m%d_%H%M%S).log
```

### Phase 4: Post-Migration

#### 4.1 Data Verification
```bash
./gotrs-migrate verify \
  --source-db "postgres://otrs_user:pass@otrs-server/otrs" \
  --target-db "postgres://gotrs_user:pass@gotrs-server/gotrs" \
  --checksum \
  --report post-migration-report.html
```

#### 4.2 System Configuration
```bash
# Update email settings
./gotrs-cli config set email.smtp.host mail.example.com
./gotrs-cli config set email.smtp.port 587
./gotrs-cli config set email.smtp.auth true

# Configure integrations
./gotrs-cli integration enable ldap
./gotrs-cli integration configure ldap --import-from otrs
```

## Data Mapping

### Database Schema Mapping

| OTRS Table | GOTRS Table | Notes |
|------------|-------------|-------|
| ticket | tickets | Direct mapping with extended fields |
| ticket_history | ticket_history | Full history preserved |
| article | ticket_messages | Renamed for clarity |
| users | users | Extended with OAuth fields |
| customer_user | customers | Restructured for organizations |
| queue | queues | Enhanced with SLA fields |
| ticket_type | ticket_types | Direct mapping |
| ticket_state | ticket_states | Workflow states added |
| ticket_priority | ticket_priorities | Direct mapping |
| groups | groups | RBAC permissions added |

### Field Mapping

```yaml
# Custom field mapping configuration
field_mapping:
  ticket:
    tn: number
    title: title
    ticket_state_id: status_id
    ticket_priority_id: priority_id
    queue_id: queue_id
    customer_id: customer_id
    customer_user_id: customer_user_id
    owner_id: agent_id
    responsible_user_id: responsible_id
    
  custom_fields:
    - otrs_field: TicketFreeText1
      gotrs_field: custom_field_1
      type: string
    - otrs_field: TicketFreeTime1
      gotrs_field: custom_date_1
      type: datetime
```

### Status Mapping

| OTRS Status | GOTRS Status | Auto-Transition |
|-------------|--------------|-----------------|
| new | new | No |
| open | open | No |
| pending reminder | pending | Yes (on date) |
| pending auto close+ | pending_positive | Yes (auto-close) |
| pending auto close- | pending_negative | Yes (auto-close) |
| closed successful | resolved | No |
| closed unsuccessful | closed | No |
| merged | merged | No |
| removed | deleted | No |

## Migration Tools

### Command Line Interface

```bash
# Basic migration
./gotrs-migrate run --config migration.yaml

# Advanced options
./gotrs-migrate run \
  --source-db $SOURCE_DB \
  --target-db $TARGET_DB \
  --include-tables tickets,users,customers \
  --exclude-tables sessions,cache \
  --from-date 2020-01-01 \
  --to-date 2024-01-01 \
  --dry-run
```

### Configuration File

```yaml
# migration.yaml
source:
  type: otrs
  version: 6.0
  database:
    driver: postgres
    host: otrs-server.example.com
    port: 5432
    name: otrs
    user: otrs_user
    password: ${OTRS_DB_PASSWORD}
    
target:
  type: gotrs
  database:
    driver: postgres
    host: gotrs-server.example.com
    port: 5432
    name: gotrs
    user: gotrs_user
    password: ${GOTRS_DB_PASSWORD}
    
options:
  batch_size: 1000
  parallel_workers: 4
  skip_validation: false
  preserve_ids: true
  convert_encoding: UTF-8
  
mappings:
  custom_fields: field_mappings.yaml
  workflows: workflow_mappings.yaml
  
excluded:
  tables:
    - sessions
    - cache_*
  users:
    - test_*
    - demo_*
```

### Rollback Procedure

```bash
# If migration fails, rollback GOTRS
psql gotrs < gotrs_pre_migration_backup.sql

# Restore OTRS to operational state
./bin/otrs.Console.pl Maint::Config::Set MaintenanceMode 0

# Investigate issues
grep ERROR migration_*.log
```

## Special Considerations

### Large Installations (>100k tickets)

```bash
# Use incremental migration
./gotrs-migrate run \
  --mode incremental \
  --cutoff-date 2023-01-01 \
  --sync-recent 30d

# After initial migration, sync recent changes
./gotrs-migrate sync \
  --source-db $SOURCE_DB \
  --target-db $TARGET_DB \
  --since-last-sync
```

### Custom OTRS Modules

1. **GenericInterface**: Requires API endpoint mapping
2. **Custom Packages**: Manual code migration needed
3. **Process Management**: Workflow redesign in GOTRS
4. **Stats Module**: Report recreation required

### Performance Optimization

```sql
-- Pre-migration optimizations
VACUUM ANALYZE;
REINDEX DATABASE otrs;

-- Post-migration optimizations
ANALYZE tickets;
ANALYZE ticket_history;
ANALYZE users;
```

## Compatibility Mode

### Enable OTRS Compatibility

```yaml
# config/compatibility.yaml
compatibility:
  otrs:
    enabled: true
    version: 6.0
    features:
      legacy_api: true
      old_urls: true
      classic_ui: false
```

### API Compatibility Layer

```nginx
# nginx.conf - URL redirects
location /otrs/index.pl {
    rewrite ^/otrs/index.pl(.*)$ /api/v1/legacy$1 permanent;
}

location /otrs/customer.pl {
    rewrite ^/otrs/customer.pl(.*)$ /customer$1 permanent;
}
```

## Post-Migration Tasks

### 1. User Training
- [ ] Agent training sessions scheduled
- [ ] Customer notification sent
- [ ] Documentation updated
- [ ] Video tutorials created

### 2. System Optimization
- [ ] Database indexes rebuilt
- [ ] Cache warming completed
- [ ] Performance baseline established
- [ ] Monitoring configured

### 3. Integration Updates
- [ ] Email templates updated
- [ ] Webhook URLs changed
- [ ] API keys regenerated
- [ ] Third-party systems notified

### 4. Cleanup
- [ ] Old OTRS system decommissioned
- [ ] Temporary migration files removed
- [ ] Backup retention policy applied
- [ ] Documentation archived

## Troubleshooting

### Common Issues

#### Character Encoding Problems
```bash
# Fix UTF-8 encoding issues
./gotrs-migrate fix-encoding \
  --target-db $TARGET_DB \
  --encoding UTF-8 \
  --tables tickets,ticket_history
```

#### Duplicate Key Errors
```sql
-- Reset sequences after migration
SELECT setval('tickets_id_seq', (SELECT MAX(id) FROM tickets));
SELECT setval('users_id_seq', (SELECT MAX(id) FROM users));
```

#### Missing Attachments
```bash
# Verify and re-sync attachments
./gotrs-migrate sync-attachments \
  --source-path /opt/otrs/var/article \
  --target-path /var/gotrs/attachments \
  --verify-checksums
```

## Support

### Migration Assistance

- **Documentation**: https://gotrs.io/docs/migration
- **Community Forum**: https://community.gotrs.io/migration
- **Professional Services**: migration@gotrs.io
- **Emergency Support**: +1-555-GOTRS-911

### Reporting Issues

When reporting migration issues, include:
1. Migration report (HTML/JSON)
2. Error logs
3. OTRS version and configuration
4. GOTRS version
5. Database type and version

---

*Migration guide version 1.0*
*Last updated: August 2025*