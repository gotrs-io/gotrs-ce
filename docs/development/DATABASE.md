# GOTRS Database Schema Design

## Overview

The GOTRS database schema is designed as a superset of OTRS's schema to ensure compatibility while adding modern features and optimizations. The schema follows PostgreSQL best practices with support for other databases planned.

## Design Principles

1. **OTRS Compatibility**: Core tables maintain OTRS field compatibility
2. **Extensibility**: JSONB fields for flexible custom data
3. **Performance**: Optimized indexes and partitioning strategies
4. **Audit Trail**: Comprehensive history tracking
5. **Multi-tenancy Ready**: Designed for future tenant isolation

## Core Schema

### Users and Authentication

```sql
-- Users table (compatible with OTRS users table)
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    login VARCHAR(200) UNIQUE NOT NULL,  -- OTRS compatibility
    email VARCHAR(255) UNIQUE NOT NULL,
    pw VARCHAR(255),  -- OTRS compatibility field name
    title VARCHAR(50),
    first_name VARCHAR(100),
    last_name VARCHAR(100),
    phone VARCHAR(50),
    mobile VARCHAR(50),
    valid_id INTEGER DEFAULT 1,  -- OTRS: 1=valid, 2=invalid
    create_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    create_by UUID REFERENCES users(id),
    change_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    change_by UUID REFERENCES users(id),
    
    -- GOTRS additions
    password_hash VARCHAR(255),  -- Modern password storage
    mfa_secret VARCHAR(255),
    mfa_enabled BOOLEAN DEFAULT FALSE,
    language VARCHAR(10) DEFAULT 'en',
    timezone VARCHAR(50) DEFAULT 'UTC',
    avatar_url TEXT,
    last_login TIMESTAMP,
    failed_login_attempts INTEGER DEFAULT 0,
    locked_until TIMESTAMP,
    metadata JSONB DEFAULT '{}',
    
    -- Indexes
    INDEX idx_users_email (email),
    INDEX idx_users_valid (valid_id),
    INDEX idx_users_login (login)
);

-- User preferences (OTRS compatible)
CREATE TABLE user_preferences (
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    preferences_key VARCHAR(150) NOT NULL,
    preferences_value TEXT,
    PRIMARY KEY (user_id, preferences_key)
);

-- Sessions
CREATE TABLE sessions (
    id VARCHAR(255) PRIMARY KEY,
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    session_data TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP,
    ip_address INET,
    user_agent TEXT,
    
    INDEX idx_sessions_user (user_id),
    INDEX idx_sessions_expires (expires_at)
);

-- API Keys (GOTRS addition)
CREATE TABLE api_keys (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    key_hash VARCHAR(255) UNIQUE NOT NULL,
    permissions JSONB,
    last_used TIMESTAMP,
    expires_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    INDEX idx_api_keys_user (user_id)
);
```

### Groups and Permissions

```sql
-- Groups (OTRS compatible)
CREATE TABLE groups (
    id SERIAL PRIMARY KEY,
    name VARCHAR(200) UNIQUE NOT NULL,
    comments VARCHAR(250),
    valid_id INTEGER DEFAULT 1,
    create_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    create_by UUID REFERENCES users(id),
    change_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    change_by UUID REFERENCES users(id),
    
    -- GOTRS additions
    parent_id INTEGER REFERENCES groups(id),
    permissions JSONB DEFAULT '{}',
    
    INDEX idx_groups_valid (valid_id)
);

-- Group-User relations (OTRS compatible)
CREATE TABLE group_user (
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    group_id INTEGER REFERENCES groups(id) ON DELETE CASCADE,
    permission_key VARCHAR(20) NOT NULL,  -- ro, move_into, create, note, owner, priority, rw
    permission_value SMALLINT NOT NULL,   -- 1=yes, 0=no
    create_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    create_by UUID REFERENCES users(id),
    change_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    change_by UUID REFERENCES users(id),
    
    PRIMARY KEY (user_id, group_id, permission_key),
    INDEX idx_group_user_group (group_id),
    INDEX idx_group_user_user (user_id)
);

-- Roles (OTRS compatible)
CREATE TABLE roles (
    id SERIAL PRIMARY KEY,
    name VARCHAR(200) UNIQUE NOT NULL,
    comments VARCHAR(250),
    valid_id INTEGER DEFAULT 1,
    create_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    create_by UUID REFERENCES users(id),
    change_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    change_by UUID REFERENCES users(id),
    
    INDEX idx_roles_valid (valid_id)
);

-- Role-User relations
CREATE TABLE role_user (
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    role_id INTEGER REFERENCES roles(id) ON DELETE CASCADE,
    create_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    create_by UUID REFERENCES users(id),
    change_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    change_by UUID REFERENCES users(id),
    
    PRIMARY KEY (user_id, role_id)
);

-- Role-Group relations
CREATE TABLE role_group (
    role_id INTEGER REFERENCES roles(id) ON DELETE CASCADE,
    group_id INTEGER REFERENCES groups(id) ON DELETE CASCADE,
    permission_key VARCHAR(20) NOT NULL,
    permission_value SMALLINT NOT NULL,
    create_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    create_by UUID REFERENCES users(id),
    change_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    change_by UUID REFERENCES users(id),
    
    PRIMARY KEY (role_id, group_id, permission_key)
);
```

### Customers

```sql
-- Customer companies (OTRS compatible)
CREATE TABLE customer_company (
    customer_id VARCHAR(150) PRIMARY KEY,
    name VARCHAR(200) NOT NULL,
    street VARCHAR(200),
    zip VARCHAR(20),
    city VARCHAR(200),
    country VARCHAR(200),
    url VARCHAR(200),
    comments VARCHAR(250),
    valid_id INTEGER DEFAULT 1,
    create_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    create_by UUID REFERENCES users(id),
    change_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    change_by UUID REFERENCES users(id),
    
    -- GOTRS additions
    parent_id VARCHAR(150) REFERENCES customer_company(customer_id),
    industry VARCHAR(100),
    size VARCHAR(50),
    sla_id INTEGER,
    metadata JSONB DEFAULT '{}',
    
    INDEX idx_customer_company_valid (valid_id),
    INDEX idx_customer_company_name (name)
);

-- Customer users (OTRS compatible)
CREATE TABLE customer_user (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    login VARCHAR(200) UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    customer_id VARCHAR(150) REFERENCES customer_company(customer_id),
    pw VARCHAR(255),  -- OTRS compatibility
    title VARCHAR(50),
    first_name VARCHAR(100),
    last_name VARCHAR(100),
    phone VARCHAR(50),
    mobile VARCHAR(50),
    street VARCHAR(150),
    zip VARCHAR(20),
    city VARCHAR(200),
    country VARCHAR(200),
    valid_id INTEGER DEFAULT 1,
    create_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    create_by UUID REFERENCES users(id),
    change_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    change_by UUID REFERENCES users(id),
    
    -- GOTRS additions
    password_hash VARCHAR(255),
    language VARCHAR(10) DEFAULT 'en',
    timezone VARCHAR(50) DEFAULT 'UTC',
    avatar_url TEXT,
    last_login TIMESTAMP,
    preferences JSONB DEFAULT '{}',
    
    INDEX idx_customer_user_email (email),
    INDEX idx_customer_user_customer (customer_id),
    INDEX idx_customer_user_valid (valid_id)
);
```

### Tickets

```sql
-- Ticket states (OTRS compatible)
CREATE TABLE ticket_state (
    id SERIAL PRIMARY KEY,
    name VARCHAR(200) UNIQUE NOT NULL,
    comments VARCHAR(250),
    type_id INTEGER NOT NULL,  -- 1=new, 2=open, 3=pending, 4=closed, 5=removed
    valid_id INTEGER DEFAULT 1,
    create_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    create_by UUID REFERENCES users(id),
    change_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    change_by UUID REFERENCES users(id),
    
    -- GOTRS additions
    color VARCHAR(7),  -- Hex color for UI
    icon VARCHAR(50),  -- Icon identifier
    auto_transition_to INTEGER REFERENCES ticket_state(id),
    auto_transition_time INTERVAL,
    
    INDEX idx_ticket_state_type (type_id)
);

-- Ticket priorities (OTRS compatible)
CREATE TABLE ticket_priority (
    id SERIAL PRIMARY KEY,
    name VARCHAR(200) UNIQUE NOT NULL,
    color VARCHAR(7),  -- GOTRS addition
    valid_id INTEGER DEFAULT 1,
    create_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    create_by UUID REFERENCES users(id),
    change_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    change_by UUID REFERENCES users(id)
);

-- Ticket types (OTRS compatible)
CREATE TABLE ticket_type (
    id SERIAL PRIMARY KEY,
    name VARCHAR(200) UNIQUE NOT NULL,
    comments VARCHAR(250),
    valid_id INTEGER DEFAULT 1,
    create_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    create_by UUID REFERENCES users(id),
    change_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    change_by UUID REFERENCES users(id),
    
    -- GOTRS additions
    icon VARCHAR(50),
    default_priority_id INTEGER REFERENCES ticket_priority(id)
);

-- Queues (OTRS compatible)
CREATE TABLE queue (
    id SERIAL PRIMARY KEY,
    name VARCHAR(200) UNIQUE NOT NULL,
    group_id INTEGER REFERENCES groups(id),
    system_address_id INTEGER,
    salutation_id INTEGER,
    signature_id INTEGER,
    follow_up_id INTEGER,
    follow_up_lock SMALLINT DEFAULT 0,
    unlock_timeout INTEGER,
    comments VARCHAR(250),
    valid_id INTEGER DEFAULT 1,
    create_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    create_by UUID REFERENCES users(id),
    change_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    change_by UUID REFERENCES users(id),
    
    -- GOTRS additions
    parent_id INTEGER REFERENCES queue(id),
    default_priority_id INTEGER REFERENCES ticket_priority(id),
    sla_id INTEGER,
    auto_response_config JSONB DEFAULT '{}',
    
    INDEX idx_queue_group (group_id),
    INDEX idx_queue_valid (valid_id)
);

-- Main tickets table (OTRS compatible with extensions)
CREATE TABLE ticket (
    id BIGSERIAL PRIMARY KEY,
    tn VARCHAR(50) UNIQUE NOT NULL,  -- Ticket number
    title VARCHAR(255),
    queue_id INTEGER REFERENCES queue(id),
    ticket_lock_id SMALLINT DEFAULT 1,  -- 1=unlocked, 2=locked
    type_id INTEGER REFERENCES ticket_type(id),
    service_id INTEGER,
    sla_id INTEGER,
    user_id UUID REFERENCES users(id),  -- Current owner
    responsible_user_id UUID REFERENCES users(id),
    ticket_priority_id INTEGER REFERENCES ticket_priority(id),
    ticket_state_id INTEGER REFERENCES ticket_state(id),
    customer_id VARCHAR(150),
    customer_user_id UUID REFERENCES customer_user(id),
    timeout INTEGER DEFAULT 0,
    until_time INTEGER DEFAULT 0,
    escalation_time INTEGER DEFAULT 0,
    escalation_update_time INTEGER DEFAULT 0,
    escalation_response_time INTEGER DEFAULT 0,
    escalation_solution_time INTEGER DEFAULT 0,
    archive_flag SMALLINT DEFAULT 0,
    create_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    create_by UUID REFERENCES users(id),
    change_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    change_by UUID REFERENCES users(id),
    
    -- GOTRS additions
    uuid UUID DEFAULT gen_random_uuid() UNIQUE,
    source VARCHAR(50) DEFAULT 'web',  -- web, email, phone, api, chat
    merged_to_id BIGINT REFERENCES ticket(id),
    split_from_id BIGINT REFERENCES ticket(id),
    tags TEXT[],
    custom_fields JSONB DEFAULT '{}',
    metadata JSONB DEFAULT '{}',
    
    -- Indexes for performance
    INDEX idx_ticket_tn (tn),
    INDEX idx_ticket_queue (queue_id),
    INDEX idx_ticket_state (ticket_state_id),
    INDEX idx_ticket_priority (ticket_priority_id),
    INDEX idx_ticket_customer (customer_id),
    INDEX idx_ticket_customer_user (customer_user_id),
    INDEX idx_ticket_user (user_id),
    INDEX idx_ticket_responsible (responsible_user_id),
    INDEX idx_ticket_create_time (create_time DESC),
    INDEX idx_ticket_escalation (escalation_time),
    INDEX idx_ticket_archive (archive_flag)
);

-- Ticket history (OTRS compatible)
CREATE TABLE ticket_history (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(200),
    history_type_id INTEGER NOT NULL,
    ticket_id BIGINT REFERENCES ticket(id) ON DELETE CASCADE,
    article_id BIGINT,
    type_id INTEGER,
    queue_id INTEGER REFERENCES queue(id),
    owner_id UUID REFERENCES users(id),
    priority_id INTEGER REFERENCES ticket_priority(id),
    state_id INTEGER REFERENCES ticket_state(id),
    create_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    create_by UUID REFERENCES users(id),
    change_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    change_by UUID REFERENCES users(id),
    
    -- GOTRS additions
    old_value TEXT,
    new_value TEXT,
    comments TEXT,
    
    INDEX idx_ticket_history_ticket (ticket_id),
    INDEX idx_ticket_history_create_time (create_time),
    INDEX idx_ticket_history_type (history_type_id)
);
```

### Articles/Messages

```sql
-- Article types (OTRS compatible)
CREATE TABLE article_type (
    id SERIAL PRIMARY KEY,
    name VARCHAR(200) UNIQUE NOT NULL,
    comments VARCHAR(250),
    valid_id INTEGER DEFAULT 1,
    create_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    create_by UUID REFERENCES users(id),
    change_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    change_by UUID REFERENCES users(id)
);

-- Articles/Messages (OTRS compatible, renamed in GOTRS)
CREATE TABLE article (
    id BIGSERIAL PRIMARY KEY,
    ticket_id BIGINT REFERENCES ticket(id) ON DELETE CASCADE,
    article_type_id INTEGER REFERENCES article_type(id),
    article_sender_type_id INTEGER,  -- 1=agent, 2=system, 3=customer
    from_address TEXT,
    to_address TEXT,
    cc_address TEXT,
    subject VARCHAR(255),
    message_id VARCHAR(255),
    message_id_md5 VARCHAR(32),
    in_reply_to TEXT,
    references TEXT,
    content_type VARCHAR(250),
    body TEXT,
    incoming_time INTEGER DEFAULT 0,
    is_visible_for_customer SMALLINT DEFAULT 1,
    create_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    create_by UUID REFERENCES users(id),
    change_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    change_by UUID REFERENCES users(id),
    
    -- GOTRS additions
    uuid UUID DEFAULT gen_random_uuid() UNIQUE,
    html_body TEXT,
    plain_body TEXT,
    internal_note BOOLEAN DEFAULT FALSE,
    delivery_status VARCHAR(50),
    read_status BOOLEAN DEFAULT FALSE,
    metadata JSONB DEFAULT '{}',
    
    INDEX idx_article_ticket (ticket_id),
    INDEX idx_article_message_id (message_id_md5),
    INDEX idx_article_create_time (create_time),
    INDEX idx_article_sender_type (article_sender_type_id)
);

-- Attachments
CREATE TABLE article_attachment (
    id BIGSERIAL PRIMARY KEY,
    article_id BIGINT REFERENCES article(id) ON DELETE CASCADE,
    filename VARCHAR(250),
    content_size VARCHAR(30),
    content_type VARCHAR(250),
    content_id VARCHAR(250),
    content_alternative VARCHAR(50),
    disposition VARCHAR(15),
    content BYTEA,  -- For small files
    create_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    create_by UUID REFERENCES users(id),
    change_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    change_by UUID REFERENCES users(id),
    
    -- GOTRS additions
    storage_backend VARCHAR(50) DEFAULT 'database',  -- database, filesystem, s3
    storage_path TEXT,  -- For external storage
    checksum VARCHAR(64),  -- SHA256
    virus_scan_status VARCHAR(50),
    virus_scan_time TIMESTAMP,
    
    INDEX idx_article_attachment_article (article_id),
    INDEX idx_article_attachment_filename (filename)
);
```

### SLA and Services

```sql
-- Services (OTRS compatible)
CREATE TABLE service (
    id SERIAL PRIMARY KEY,
    name VARCHAR(200) UNIQUE NOT NULL,
    valid_id INTEGER DEFAULT 1,
    comments VARCHAR(250),
    create_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    create_by UUID REFERENCES users(id),
    change_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    change_by UUID REFERENCES users(id),
    
    -- GOTRS additions
    parent_id INTEGER REFERENCES service(id),
    criticality VARCHAR(50),
    
    INDEX idx_service_valid (valid_id)
);

-- SLA definitions (OTRS compatible)
CREATE TABLE sla (
    id SERIAL PRIMARY KEY,
    name VARCHAR(200) UNIQUE NOT NULL,
    calendar_name VARCHAR(100),
    first_response_time INTEGER,  -- minutes
    first_response_notify INTEGER,  -- percent
    update_time INTEGER,
    update_notify INTEGER,
    solution_time INTEGER,
    solution_notify INTEGER,
    valid_id INTEGER DEFAULT 1,
    comments VARCHAR(250),
    create_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    create_by UUID REFERENCES users(id),
    change_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    change_by UUID REFERENCES users(id),
    
    -- GOTRS additions
    escalation_levels INTEGER DEFAULT 3,
    business_hours JSONB DEFAULT '{}',
    holidays JSONB DEFAULT '[]',
    
    INDEX idx_sla_valid (valid_id)
);

-- Service-SLA relations
CREATE TABLE service_sla (
    service_id INTEGER REFERENCES service(id) ON DELETE CASCADE,
    sla_id INTEGER REFERENCES sla(id) ON DELETE CASCADE,
    PRIMARY KEY (service_id, sla_id)
);
```

### Workflows and Automation

```sql
-- Generic Interface (GOTRS addition)
CREATE TABLE workflows (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) UNIQUE NOT NULL,
    description TEXT,
    trigger_type VARCHAR(50),  -- time, event, manual
    trigger_config JSONB,
    conditions JSONB,
    actions JSONB,
    enabled BOOLEAN DEFAULT TRUE,
    valid_id INTEGER DEFAULT 1,
    execution_order INTEGER DEFAULT 0,
    last_execution TIMESTAMP,
    execution_count INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    INDEX idx_workflows_enabled (enabled),
    INDEX idx_workflows_trigger_type (trigger_type)
);

-- Workflow execution history
CREATE TABLE workflow_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workflow_id UUID REFERENCES workflows(id) ON DELETE CASCADE,
    ticket_id BIGINT REFERENCES ticket(id),
    status VARCHAR(50),  -- success, failure, partial
    started_at TIMESTAMP,
    completed_at TIMESTAMP,
    error_message TEXT,
    actions_taken JSONB,
    
    INDEX idx_workflow_history_workflow (workflow_id),
    INDEX idx_workflow_history_ticket (ticket_id),
    INDEX idx_workflow_history_started (started_at DESC)
);
```

### Audit and Compliance

```sql
-- Audit log (GOTRS addition)
CREATE TABLE audit_log (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    user_id UUID REFERENCES users(id),
    session_id VARCHAR(255),
    ip_address INET,
    user_agent TEXT,
    action VARCHAR(100) NOT NULL,
    object_type VARCHAR(50),
    object_id VARCHAR(255),
    old_values JSONB,
    new_values JSONB,
    result VARCHAR(50),  -- success, failure, partial
    error_message TEXT,
    metadata JSONB,
    
    INDEX idx_audit_log_timestamp (timestamp DESC),
    INDEX idx_audit_log_user (user_id),
    INDEX idx_audit_log_action (action),
    INDEX idx_audit_log_object (object_type, object_id)
);

-- GDPR compliance
CREATE TABLE data_retention_policy (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    table_name VARCHAR(100) NOT NULL,
    column_name VARCHAR(100),
    retention_days INTEGER NOT NULL,
    action VARCHAR(50) NOT NULL,  -- delete, anonymize, archive
    conditions JSONB,
    last_run TIMESTAMP,
    next_run TIMESTAMP,
    enabled BOOLEAN DEFAULT TRUE
);
```

### Performance Optimizations

```sql
-- Partitioning for large tables (PostgreSQL 11+)
-- Partition tickets by year
CREATE TABLE ticket_2024 PARTITION OF ticket
    FOR VALUES FROM ('2024-01-01') TO ('2025-01-01');

CREATE TABLE ticket_2025 PARTITION OF ticket
    FOR VALUES FROM ('2025-01-01') TO ('2026-01-01');

-- Partition audit_log by month
CREATE TABLE audit_log_2024_01 PARTITION OF audit_log
    FOR VALUES FROM ('2024-01-01') TO ('2024-02-01');

-- Materialized views for reporting
CREATE MATERIALIZED VIEW ticket_statistics AS
SELECT 
    DATE_TRUNC('day', create_time) as date,
    queue_id,
    ticket_state_id,
    ticket_priority_id,
    COUNT(*) as ticket_count,
    AVG(EXTRACT(EPOCH FROM (change_time - create_time))/3600)::INT as avg_resolution_hours
FROM ticket
WHERE create_time > CURRENT_DATE - INTERVAL '90 days'
GROUP BY 1, 2, 3, 4;

CREATE INDEX idx_ticket_statistics_date ON ticket_statistics(date DESC);

-- Refresh materialized view daily
CREATE OR REPLACE FUNCTION refresh_ticket_statistics()
RETURNS void AS $$
BEGIN
    REFRESH MATERIALIZED VIEW CONCURRENTLY ticket_statistics;
END;
$$ LANGUAGE plpgsql;
```

### Custom Fields

```sql
-- Dynamic custom field definitions
CREATE TABLE custom_field_definitions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    object_type VARCHAR(50) NOT NULL,  -- ticket, customer, user
    field_name VARCHAR(100) NOT NULL,
    field_type VARCHAR(50) NOT NULL,  -- text, number, date, select, multiselect
    field_config JSONB,  -- options, validation rules, etc.
    required BOOLEAN DEFAULT FALSE,
    visible_for_customer BOOLEAN DEFAULT TRUE,
    searchable BOOLEAN DEFAULT TRUE,
    position INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    UNIQUE(object_type, field_name),
    INDEX idx_custom_field_object_type (object_type)
);

-- Custom field values
CREATE TABLE custom_field_values (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    field_id UUID REFERENCES custom_field_definitions(id) ON DELETE CASCADE,
    object_id VARCHAR(255) NOT NULL,  -- ticket.id, user.id, etc.
    value_text TEXT,
    value_number NUMERIC,
    value_date TIMESTAMP,
    value_json JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    INDEX idx_custom_field_values_field (field_id),
    INDEX idx_custom_field_values_object (object_id)
);
```

## Migration Support

```sql
-- OTRS compatibility views
CREATE VIEW otrs_ticket AS
SELECT 
    id,
    tn,
    title,
    queue_id,
    ticket_lock_id AS lock_id,
    type_id,
    ticket_state_id AS state_id,
    ticket_priority_id AS priority_id,
    create_time,
    change_time
FROM ticket;

-- Migration tracking
CREATE TABLE migration_log (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    source_system VARCHAR(50),
    source_version VARCHAR(50),
    migration_type VARCHAR(50),
    table_name VARCHAR(100),
    records_processed INTEGER,
    records_failed INTEGER,
    started_at TIMESTAMP,
    completed_at TIMESTAMP,
    status VARCHAR(50),
    error_log TEXT
);
```

## Database Maintenance

```sql
-- Maintenance procedures
CREATE OR REPLACE FUNCTION cleanup_old_sessions()
RETURNS void AS $$
BEGIN
    DELETE FROM sessions WHERE expires_at < CURRENT_TIMESTAMP;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION archive_old_tickets()
RETURNS void AS $$
BEGIN
    UPDATE ticket 
    SET archive_flag = 1 
    WHERE ticket_state_id IN (
        SELECT id FROM ticket_state WHERE type_id = 4  -- closed
    ) 
    AND change_time < CURRENT_DATE - INTERVAL '90 days'
    AND archive_flag = 0;
END;
$$ LANGUAGE plpgsql;

-- Scheduled maintenance
SELECT cron.schedule('cleanup-sessions', '0 * * * *', 'SELECT cleanup_old_sessions()');
SELECT cron.schedule('archive-tickets', '0 2 * * *', 'SELECT archive_old_tickets()');
```

## Performance Indexes

```sql
-- Additional performance indexes
CREATE INDEX idx_ticket_fulltext ON ticket USING gin(to_tsvector('english', title || ' ' || COALESCE(title, '')));
CREATE INDEX idx_article_fulltext ON article USING gin(to_tsvector('english', COALESCE(subject, '') || ' ' || COALESCE(body, '')));
CREATE INDEX idx_ticket_custom_fields ON ticket USING gin(custom_fields);
CREATE INDEX idx_ticket_tags ON ticket USING gin(tags);

-- Partial indexes for common queries
CREATE INDEX idx_ticket_open ON ticket(id) WHERE ticket_state_id IN (1, 2);  -- new, open
CREATE INDEX idx_ticket_unassigned ON ticket(id) WHERE user_id IS NULL;
CREATE INDEX idx_ticket_escalated ON ticket(escalation_time) WHERE escalation_time > 0;
```

---

*Database schema version 1.0*
*Last updated: August 2025*