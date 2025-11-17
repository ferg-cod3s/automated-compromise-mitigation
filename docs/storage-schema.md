# ACM Phase III - Storage Schema Documentation

**Version:** 1.0
**Date:** 2025-11-17
**Status:** Design Complete
**Document Type:** Technical Specification

---

## Table of Contents

1. [Overview](#overview)
2. [Schema Design Principles](#schema-design-principles)
3. [Table Structures](#table-structures)
4. [Indexes and Performance](#indexes-and-performance)
5. [Data Integrity](#data-integrity)
6. [Migration Strategy](#migration-strategy)
7. [Backup and Recovery](#backup-and-recovery)
8. [Security Considerations](#security-considerations)

---

## Overview

Phase III introduces SQLite-based persistent storage to replace the in-memory storage used in Phases I and II. This enables:

- **Data Persistence**: CRCs, evidence chains, and audit logs survive service restarts
- **Scalability**: Handle thousands of credentials and compliance validations
- **Auditability**: Cryptographically-signed audit trails with chain integrity
- **Performance**: Indexed queries for fast lookups and reporting

### Database Location

- **Default Path**: `~/.acm/data/acm.db`
- **Configurable**: Via `ACM_DATABASE_PATH` environment variable
- **Permissions**: `0600` (owner read/write only)
- **Backup Path**: `~/.acm/data/backups/acm-{timestamp}.db`

---

## Schema Design Principles

### 1. SQLite-Friendly Design

- **Use TEXT for IDs**: SQLite has excellent TEXT performance
- **Store JSON for Complex Data**: Leverage SQLite's JSON1 extension
- **Store Enums as INTEGER**: With mapping tables for readability
- **Use INTEGER for Timestamps**: Unix timestamps for simplicity

### 2. Normalization Strategy

- **Moderate Normalization**: Balance query performance with data integrity
- **Enum Tables**: Separate tables for enum values (reference data)
- **Denormalization**: Store JSON for complex nested structures (rules, metadata)

### 3. Performance Optimization

- **Strategic Indexes**: Index frequently queried columns (site, timestamp, credential_id)
- **Composite Indexes**: For common query patterns (e.g., site + expires_at)
- **Write-Ahead Logging (WAL)**: Enable concurrent reads during writes
- **Incremental Vacuum**: Prevent fragmentation without blocking

### 4. Security and Integrity

- **Foreign Key Constraints**: Ensure referential integrity (CRC → Evidence)
- **Triggers**: Validate evidence chain linkage on insert
- **Signatures**: Store Ed25519 signatures for tamper detection
- **Encryption at Rest**: User responsibility (filesystem/OS-level encryption)

---

## Table Structures

### 3.1 `crcs` - Compliance Rule Sets

Stores parsed Terms of Service compliance rules for websites.

| Column | Type | Description |
|--------|------|-------------|
| `id` | TEXT PRIMARY KEY | Unique CRC identifier (e.g., `CRC-github.com-a1b2c3d4`) |
| `site` | TEXT NOT NULL | Domain name (e.g., `github.com`) |
| `tos_url` | TEXT NOT NULL | URL of the Terms of Service document |
| `tos_version` | TEXT NOT NULL | ToS version (date or version string) |
| `tos_hash` | TEXT NOT NULL | SHA-256 hash of ToS content |
| `parsed_at` | INTEGER NOT NULL | Unix timestamp when ToS was analyzed |
| `expires_at` | INTEGER NOT NULL | Unix timestamp when CRC expires |
| `stored_at` | INTEGER NOT NULL | Unix timestamp when cached |
| `recommendation` | INTEGER NOT NULL | Overall recommendation (enum) |
| `reasoning` | TEXT | Human-readable reasoning for recommendation |
| `rules_json` | TEXT NOT NULL | JSON array of `ComplianceRule` objects |
| `signature` | TEXT NOT NULL | Ed25519 signature (hex-encoded) |
| `created_at` | INTEGER NOT NULL | Auto-generated creation timestamp |

**Indexes:**
- `idx_crcs_site` (site)
- `idx_crcs_expires_at` (expires_at)
- `idx_crcs_site_expires` (site, expires_at)

**Sample Query:**
```sql
-- Get valid CRC for a site
SELECT * FROM crcs
WHERE site = 'github.com'
  AND expires_at > strftime('%s', 'now')
ORDER BY parsed_at DESC
LIMIT 1;
```

---

### 3.2 `evidence_entries` - Evidence Chain

Stores cryptographically-linked evidence chain entries for compliance auditing.

| Column | Type | Description |
|--------|------|-------------|
| `id` | TEXT PRIMARY KEY | Unique entry ID (e.g., `EVD-1700000000-a1b2c3d4`) |
| `timestamp` | INTEGER NOT NULL | Unix timestamp of event |
| `event_type` | INTEGER NOT NULL | Event type (enum: validation, rotation, etc.) |
| `site` | TEXT NOT NULL | Target site domain |
| `credential_id_hash` | TEXT NOT NULL | SHA-256 hash of credential ID |
| `action_type` | INTEGER NOT NULL | Action type (enum) |
| `action_method` | INTEGER NOT NULL | Automation method (enum) |
| `action_context_json` | TEXT | JSON map of additional action context |
| `validation_result` | INTEGER NOT NULL | Validation result (enum) |
| `crc_id` | TEXT | Foreign key to `crcs.id` (nullable) |
| `applied_rule_ids_json` | TEXT | JSON array of rule IDs |
| `evidence_data_json` | TEXT NOT NULL | JSON blob with evidence details |
| `previous_entry_id` | TEXT | Links to previous entry (NULL for first) |
| `chain_hash` | TEXT NOT NULL | SHA-256 hash linking to previous entry |
| `signature` | TEXT NOT NULL | Ed25519 signature (hex-encoded) |
| `created_at` | INTEGER NOT NULL | Auto-generated creation timestamp |

**Indexes:**
- `idx_evidence_timestamp` (timestamp)
- `idx_evidence_site` (site)
- `idx_evidence_credential_hash` (credential_id_hash)
- `idx_evidence_event_type` (event_type)
- `idx_evidence_crc_id` (crc_id)
- `idx_evidence_previous_entry` (previous_entry_id)
- `idx_evidence_credential_timestamp` (credential_id_hash, timestamp)

**Sample Query:**
```sql
-- Export evidence chain for a credential
SELECT * FROM evidence_entries
WHERE credential_id_hash = 'abc123...'
ORDER BY timestamp ASC;
```

---

### 3.3 `audit_events` - Audit Logs

Stores audit events with cryptographic signatures for tamper-evidence.

| Column | Type | Description |
|--------|------|-------------|
| `id` | INTEGER PRIMARY KEY | Auto-increment ID |
| `event_id` | TEXT NOT NULL UNIQUE | UUID or generated ID |
| `timestamp` | INTEGER NOT NULL | Unix timestamp |
| `event_type` | TEXT NOT NULL | Event type (rotation, detection, compliance, him, auth, system) |
| `status` | TEXT NOT NULL | Status (success, failure, pending, skipped) |
| `credential_id` | TEXT NOT NULL | Hashed credential identifier |
| `site` | TEXT | Website/service domain |
| `username` | TEXT | Username (may be encrypted) |
| `message` | TEXT | Additional context message |
| `metadata_json` | TEXT | JSON map of metadata |
| `signature` | BLOB NOT NULL | Ed25519 signature (raw bytes) |
| `created_at` | INTEGER NOT NULL | Auto-generated creation timestamp |

**Indexes:**
- `idx_audit_timestamp` (timestamp)
- `idx_audit_event_type` (event_type)
- `idx_audit_credential_id` (credential_id)
- `idx_audit_site` (site)
- `idx_audit_status` (status)
- `idx_audit_type_timestamp` (event_type, timestamp)

**Sample Query:**
```sql
-- Get recent failed rotations
SELECT * FROM audit_events
WHERE event_type = 'rotation'
  AND status = 'failure'
  AND timestamp > strftime('%s', 'now', '-7 days')
ORDER BY timestamp DESC;
```

---

### 3.4 Enum Mapping Tables

Reference tables for enum value lookups:

- `enum_compliance_recommendation` - ComplianceRecommendation enum
- `enum_evidence_event_type` - EvidenceEventType enum
- `enum_action_type` - ActionType enum
- `enum_automation_method` - AutomationMethod enum
- `enum_validation_result` - ValidationResult enum

**Purpose:**
- Human-readable enum names in queries
- Validation of enum values
- Documentation of valid values

**Sample Query:**
```sql
-- Join with enum table for readable output
SELECT
    e.id,
    e.site,
    t.name as event_type_name,
    e.timestamp
FROM evidence_entries e
JOIN enum_evidence_event_type t ON e.event_type = t.value;
```

---

## Indexes and Performance

### Index Strategy

1. **Single-Column Indexes**: For high-selectivity columns (site, credential_id_hash)
2. **Composite Indexes**: For common multi-column filters (site + expires_at)
3. **Timestamp Indexes**: Enable efficient time-range queries
4. **Foreign Key Indexes**: Improve join performance (crc_id)

### Performance Targets

| Operation | Target | Notes |
|-----------|--------|-------|
| CRC lookup by site | < 10ms | With valid index hit |
| Evidence chain query (10K entries) | < 50ms | Indexed by credential_id |
| Audit log query (100K entries) | < 100ms | Indexed by timestamp |
| Evidence chain verification | < 500ms | Full chain traversal |
| Database startup | < 200ms | Schema verification |

### SQLite Optimizations

Configured in `schema.sql`:

```sql
PRAGMA foreign_keys = ON;           -- Enable FK constraints
PRAGMA journal_mode = WAL;          -- Write-Ahead Logging
PRAGMA synchronous = NORMAL;        -- Balance speed/safety
PRAGMA cache_size = -64000;         -- 64MB cache
PRAGMA auto_vacuum = INCREMENTAL;   -- Prevent fragmentation
```

---

## Data Integrity

### Foreign Key Constraints

- **Evidence → CRC**: `evidence_entries.crc_id` references `crcs.id`
  - `ON DELETE SET NULL`: If CRC deleted, evidence remains but CRC link is nullified

### Triggers

**`validate_evidence_chain`**: Ensures previous entry exists before inserting new entry

```sql
CREATE TRIGGER validate_evidence_chain
BEFORE INSERT ON evidence_entries
FOR EACH ROW
WHEN NEW.previous_entry_id IS NOT NULL
BEGIN
    SELECT RAISE(ABORT, 'Previous entry does not exist')
    WHERE NOT EXISTS (
        SELECT 1 FROM evidence_entries WHERE id = NEW.previous_entry_id
    );
END;
```

### Cryptographic Signatures

All evidence entries and CRCs are signed with Ed25519:

- **CRC Signature**: Signs `id || site || tos_hash || parsed_at || recommendation`
- **Evidence Signature**: Signs `id || timestamp || site || credential_id_hash || event_type || validation_result || chain_hash`

Signature verification detects tampering in exported data.

---

## Migration Strategy

### Migration System Design

**File Structure:**
```
internal/acvs/storage/migrations/
├── 001_initial_schema.sql
├── 002_add_rotation_state.sql (future)
└── 003_add_performance_indexes.sql (future)
```

**Migration Workflow:**

1. **Service Startup**: Check current schema version
2. **Compare**: Determine migrations needed (current → latest)
3. **Apply**: Execute migrations in order with transactions
4. **Rollback**: Automatic rollback on failure
5. **Verify**: Run integrity checks after migration

**Schema Version Tracking:**

```sql
CREATE TABLE schema_version (
    version INTEGER PRIMARY KEY,
    applied_at INTEGER NOT NULL,
    description TEXT NOT NULL,
    checksum TEXT NOT NULL
);
```

**Migration Checklist:**

- [ ] Create migration SQL file (`00X_description.sql`)
- [ ] Include both UP and DOWN (rollback) migrations
- [ ] Test migration on copy of production database
- [ ] Document breaking changes in `CHANGELOG.md`
- [ ] Update this document with schema changes

---

## Backup and Recovery

### Automatic Backups

**Backup Strategy:**
- **Pre-Migration**: Automatic backup before applying migrations
- **Daily**: Scheduled backup at 02:00 local time (configurable)
- **Pre-Upgrade**: Backup before ACM service upgrade

**Backup Location:**
```
~/.acm/data/backups/
├── acm-20250117-020000.db          # Daily backup
├── acm-20250117-migration-v2.db    # Pre-migration backup
└── acm-20250116-020000.db          # Previous daily backup
```

**Backup Retention:**
- Daily backups: 30 days
- Migration backups: Permanent (until manually deleted)
- Upgrade backups: 90 days

### Manual Backup

```bash
# Create backup
sqlite3 ~/.acm/data/acm.db ".backup ~/.acm/data/backups/manual-backup.db"

# Verify backup integrity
sqlite3 ~/.acm/data/backups/manual-backup.db "PRAGMA integrity_check;"
```

### Restore Procedure

```bash
# 1. Stop ACM service
systemctl stop acm-service

# 2. Restore from backup
cp ~/.acm/data/backups/acm-20250117-020000.db ~/.acm/data/acm.db

# 3. Verify integrity
sqlite3 ~/.acm/data/acm.db "PRAGMA integrity_check;"

# 4. Restart service
systemctl start acm-service
```

### Disaster Recovery

**Corruption Detection:**
- Service startup: Run `PRAGMA integrity_check`
- Scheduled: Weekly integrity verification

**Recovery Steps:**
1. Detect corruption (integrity check fails)
2. Alert user via TUI/logs
3. Attempt SQLite recovery (`PRAGMA quick_check`)
4. If recovery fails, restore from latest backup
5. Log incident in audit trail

---

## Security Considerations

### Encryption at Rest

**Database File:**
- ACM does **not** provide database encryption
- Users should use **filesystem-level encryption** (LUKS, FileVault, BitLocker)
- Or **OS-level encryption** (eCryptfs, fscrypt)

**Sensitive Data:**
- Credential IDs are **SHA-256 hashed** before storage
- Usernames and sites stored in plaintext (needed for queries)
- Passwords **never stored** (zero-knowledge principle)

### File Permissions

```bash
~/.acm/data/acm.db         # 0600 (owner read/write only)
~/.acm/data/acm.db-wal     # 0600 (WAL file)
~/.acm/data/acm.db-shm     # 0600 (shared memory)
```

### SQL Injection Prevention

All queries use **parameterized statements**:

```go
// SAFE: Parameterized query
stmt, err := db.Prepare("SELECT * FROM crcs WHERE site = ?")
result := stmt.QueryRow(site)

// UNSAFE: String concatenation (NEVER DO THIS)
query := fmt.Sprintf("SELECT * FROM crcs WHERE site = '%s'", site)
```

### Signature Verification

**On Export:**
1. Retrieve data from SQLite
2. Verify Ed25519 signatures
3. Check evidence chain integrity
4. Generate export report with verification status

**Evidence Chain Integrity:**
- Verify each entry's signature
- Verify chain hash linkage
- Detect broken links or tampering

---

## Maintenance Operations

### Database Vacuuming

```sql
-- Full vacuum (requires exclusive lock)
VACUUM;

-- Incremental vacuum (no lock required)
PRAGMA incremental_vacuum(100);
```

**Schedule:** Monthly via cron or systemd timer

### Analyzing Statistics

```sql
-- Update query planner statistics
ANALYZE;
```

**Schedule:** Weekly or after large data imports

### Cleanup Expired CRCs

```sql
DELETE FROM crcs WHERE expires_at < strftime('%s', 'now');
```

**Schedule:** Daily at 03:00 local time

### Database Size Monitoring

```bash
# Check database size
du -h ~/.acm/data/acm.db

# Check table sizes
sqlite3 ~/.acm/data/acm.db "
    SELECT name, (pgsize * pgcount) / 1024 / 1024 as size_mb
    FROM dbstat
    GROUP BY name
    ORDER BY size_mb DESC;
"
```

---

## Future Enhancements (Phase IV)

- **Full-Text Search**: Enable FTS5 for searching evidence data
- **Partitioning**: Separate old audit logs into archive tables
- **Replication**: Export to PostgreSQL for enterprise deployments
- **Encryption**: SQLCipher integration for encrypted databases
- **Sharding**: Separate databases per password manager vault

---

## References

- [SQLite Documentation](https://www.sqlite.org/docs.html)
- [SQLite JSON1 Extension](https://www.sqlite.org/json1.html)
- [SQLite Write-Ahead Logging](https://www.sqlite.org/wal.html)
- [ACM Threat Model](../acm-threat-model.md)
- [ACM Phase I Implementation](../PHASE1_IMPLEMENTATION_SUMMARY.md)
- [ACM Phase II Implementation](../PHASE2_IMPLEMENTATION_SUMMARY.md)

---

**Document Status:** Complete
**Next Step:** Implement SQLite CRC storage (Task 1.2)
