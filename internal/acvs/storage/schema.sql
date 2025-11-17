-- ACM Phase III - SQLite Storage Schema
-- This schema supports persistent storage for:
-- 1. Compliance Rule Sets (CRCs)
-- 2. Evidence Chain Entries
-- 3. Audit Events (Phase I compatibility)
--
-- Design Principles:
-- - Use TEXT for IDs and JSON for complex data (SQLite-friendly)
-- - Index frequently queried columns (site, timestamp, credential_id_hash)
-- - Store enums as INTEGER for efficiency with mapping tables
-- - Support cryptographic signatures (Ed25519) as HEX strings
-- - Enable foreign key constraints for data integrity

-- ============================================================================
-- Schema Version Management
-- ============================================================================

CREATE TABLE IF NOT EXISTS schema_version (
    version INTEGER PRIMARY KEY,
    applied_at INTEGER NOT NULL,  -- Unix timestamp
    description TEXT NOT NULL,
    checksum TEXT NOT NULL        -- SHA-256 hash of migration script
);

-- ============================================================================
-- Compliance Rule Sets (CRCs)
-- ============================================================================

CREATE TABLE IF NOT EXISTS crcs (
    -- Primary identifier
    id TEXT PRIMARY KEY,

    -- Site information
    site TEXT NOT NULL,
    tos_url TEXT NOT NULL,
    tos_version TEXT NOT NULL,
    tos_hash TEXT NOT NULL,  -- SHA-256 hash of ToS content

    -- Timestamps
    parsed_at INTEGER NOT NULL,   -- Unix timestamp
    expires_at INTEGER NOT NULL,  -- Unix timestamp
    stored_at INTEGER NOT NULL,   -- Unix timestamp (when cached)

    -- Compliance data
    recommendation INTEGER NOT NULL,  -- Maps to ComplianceRecommendation enum
    reasoning TEXT,
    rules_json TEXT NOT NULL,  -- JSON array of ComplianceRule objects

    -- Cryptographic signature
    signature TEXT NOT NULL,  -- Ed25519 signature (hex-encoded)

    -- Metadata
    created_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now'))
);

-- CRC Indexes for performance
CREATE INDEX IF NOT EXISTS idx_crcs_site ON crcs(site);
CREATE INDEX IF NOT EXISTS idx_crcs_expires_at ON crcs(expires_at);
CREATE INDEX IF NOT EXISTS idx_crcs_site_expires ON crcs(site, expires_at);

-- ============================================================================
-- Evidence Chain Entries
-- ============================================================================

CREATE TABLE IF NOT EXISTS evidence_entries (
    -- Primary identifier
    id TEXT PRIMARY KEY,

    -- Timestamp
    timestamp INTEGER NOT NULL,  -- Unix timestamp

    -- Event classification
    event_type INTEGER NOT NULL,  -- Maps to EvidenceEventType enum

    -- Credential and site information
    site TEXT NOT NULL,
    credential_id_hash TEXT NOT NULL,  -- SHA-256 hash of credential ID

    -- Action details (stored as JSON)
    action_type INTEGER NOT NULL,       -- Maps to ActionType enum
    action_method INTEGER NOT NULL,     -- Maps to AutomationMethod enum
    action_context_json TEXT,           -- JSON map of additional context

    -- Validation result
    validation_result INTEGER NOT NULL,  -- Maps to ValidationResult enum

    -- CRC reference
    crc_id TEXT,  -- Foreign key to crcs.id (nullable, may not have CRC)
    applied_rule_ids_json TEXT,  -- JSON array of rule IDs

    -- Evidence data
    evidence_data_json TEXT NOT NULL,  -- JSON blob with evidence details

    -- Chain linkage
    previous_entry_id TEXT,  -- Links to previous entry (NULL for first entry)
    chain_hash TEXT NOT NULL,  -- SHA-256 hash linking to previous entry

    -- Cryptographic signature
    signature TEXT NOT NULL,  -- Ed25519 signature (hex-encoded)

    -- Metadata
    created_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),

    -- Foreign key constraint
    FOREIGN KEY (crc_id) REFERENCES crcs(id) ON DELETE SET NULL
);

-- Evidence Entry Indexes for performance
CREATE INDEX IF NOT EXISTS idx_evidence_timestamp ON evidence_entries(timestamp);
CREATE INDEX IF NOT EXISTS idx_evidence_site ON evidence_entries(site);
CREATE INDEX IF NOT EXISTS idx_evidence_credential_hash ON evidence_entries(credential_id_hash);
CREATE INDEX IF NOT EXISTS idx_evidence_event_type ON evidence_entries(event_type);
CREATE INDEX IF NOT EXISTS idx_evidence_crc_id ON evidence_entries(crc_id);
CREATE INDEX IF NOT EXISTS idx_evidence_previous_entry ON evidence_entries(previous_entry_id);
CREATE INDEX IF NOT EXISTS idx_evidence_credential_timestamp ON evidence_entries(credential_id_hash, timestamp);

-- ============================================================================
-- Audit Events (Phase I Compatibility)
-- ============================================================================

CREATE TABLE IF NOT EXISTS audit_events (
    -- Primary identifier (auto-increment for Phase I compatibility)
    id INTEGER PRIMARY KEY AUTOINCREMENT,

    -- Event identifier
    event_id TEXT NOT NULL UNIQUE,  -- UUID or generated ID

    -- Timestamp
    timestamp INTEGER NOT NULL,  -- Unix timestamp

    -- Event classification
    event_type TEXT NOT NULL,  -- rotation, detection, compliance, him, auth, system
    status TEXT NOT NULL,      -- success, failure, pending, skipped

    -- Credential information
    credential_id TEXT NOT NULL,  -- Hashed credential identifier
    site TEXT,
    username TEXT,

    -- Event details
    message TEXT,
    metadata_json TEXT,  -- JSON map of additional metadata

    -- Cryptographic signature
    signature BLOB NOT NULL,  -- Ed25519 signature (raw bytes)

    -- Metadata
    created_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now'))
);

-- Audit Event Indexes for performance
CREATE INDEX IF NOT EXISTS idx_audit_timestamp ON audit_events(timestamp);
CREATE INDEX IF NOT EXISTS idx_audit_event_type ON audit_events(event_type);
CREATE INDEX IF NOT EXISTS idx_audit_credential_id ON audit_events(credential_id);
CREATE INDEX IF NOT EXISTS idx_audit_site ON audit_events(site);
CREATE INDEX IF NOT EXISTS idx_audit_status ON audit_events(status);
CREATE INDEX IF NOT EXISTS idx_audit_type_timestamp ON audit_events(event_type, timestamp);

-- ============================================================================
-- Enum Mapping Tables (for reference and validation)
-- ============================================================================

-- ComplianceRecommendation enum mapping
CREATE TABLE IF NOT EXISTS enum_compliance_recommendation (
    value INTEGER PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    description TEXT
);

INSERT OR IGNORE INTO enum_compliance_recommendation (value, name, description) VALUES
    (0, 'UNSPECIFIED', 'Unspecified recommendation'),
    (1, 'ALLOWED', 'Automation is safe'),
    (2, 'ALLOWED_WITH_API', 'Use API if available'),
    (3, 'HIM_REQUIRED', 'Human interaction needed'),
    (4, 'BLOCKED', 'Automation prohibited'),
    (5, 'UNCERTAIN', 'Unable to determine');

-- EvidenceEventType enum mapping
CREATE TABLE IF NOT EXISTS enum_evidence_event_type (
    value INTEGER PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    description TEXT
);

INSERT OR IGNORE INTO enum_evidence_event_type (value, name, description) VALUES
    (0, 'UNSPECIFIED', 'Unspecified event type'),
    (1, 'VALIDATION', 'Pre-flight validation'),
    (2, 'ROTATION', 'Credential rotation'),
    (3, 'HIM_PROMPT', 'Human intervention'),
    (4, 'CRC_UPDATE', 'CRC cache update'),
    (5, 'ACVS_ENABLED', 'ACVS opt-in'),
    (6, 'ACVS_DISABLED', 'ACVS opt-out');

-- ActionType enum mapping
CREATE TABLE IF NOT EXISTS enum_action_type (
    value INTEGER PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    description TEXT
);

INSERT OR IGNORE INTO enum_action_type (value, name, description) VALUES
    (0, 'UNSPECIFIED', 'Unspecified action type'),
    (1, 'CREDENTIAL_ROTATION', 'Credential rotation'),
    (2, 'PASSWORD_CHANGE', 'Password change'),
    (3, 'MFA_SETUP', 'MFA setup'),
    (4, 'ACCOUNT_RECOVERY', 'Account recovery'),
    (5, 'DATA_EXPORT', 'Data export');

-- AutomationMethod enum mapping
CREATE TABLE IF NOT EXISTS enum_automation_method (
    value INTEGER PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    description TEXT
);

INSERT OR IGNORE INTO enum_automation_method (value, name, description) VALUES
    (0, 'UNSPECIFIED', 'Unspecified method'),
    (1, 'API', 'Official API'),
    (2, 'UI_SCRIPT', 'Browser automation'),
    (3, 'CLI', 'Command-line tool'),
    (4, 'MANUAL', 'Human-in-the-Middle');

-- ValidationResult enum mapping
CREATE TABLE IF NOT EXISTS enum_validation_result (
    value INTEGER PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    description TEXT
);

INSERT OR IGNORE INTO enum_validation_result (value, name, description) VALUES
    (0, 'UNSPECIFIED', 'Unspecified result'),
    (1, 'ALLOWED', 'Action is permitted'),
    (2, 'HIM_REQUIRED', 'Requires human interaction'),
    (3, 'BLOCKED', 'Action is prohibited'),
    (4, 'RATE_LIMITED', 'Rate limit would be exceeded'),
    (5, 'DISABLED', 'ACVS not enabled');

-- ============================================================================
-- Database Statistics and Maintenance
-- ============================================================================

-- View for CRC statistics
CREATE VIEW IF NOT EXISTS v_crc_stats AS
SELECT
    COUNT(*) as total_crcs,
    COUNT(CASE WHEN expires_at > strftime('%s', 'now') THEN 1 END) as valid_crcs,
    COUNT(CASE WHEN expires_at <= strftime('%s', 'now') THEN 1 END) as expired_crcs,
    COUNT(DISTINCT site) as unique_sites,
    MIN(parsed_at) as oldest_parse,
    MAX(parsed_at) as newest_parse
FROM crcs;

-- View for evidence chain statistics
CREATE VIEW IF NOT EXISTS v_evidence_stats AS
SELECT
    COUNT(*) as total_entries,
    COUNT(DISTINCT site) as unique_sites,
    COUNT(DISTINCT credential_id_hash) as unique_credentials,
    MIN(timestamp) as first_entry,
    MAX(timestamp) as last_entry,
    (SELECT COUNT(*) FROM evidence_entries WHERE previous_entry_id IS NULL) as chain_starts
FROM evidence_entries;

-- View for audit event statistics
CREATE VIEW IF NOT EXISTS v_audit_stats AS
SELECT
    COUNT(*) as total_events,
    COUNT(CASE WHEN status = 'success' THEN 1 END) as success_count,
    COUNT(CASE WHEN status = 'failure' THEN 1 END) as failure_count,
    COUNT(CASE WHEN event_type = 'rotation' THEN 1 END) as rotation_count,
    COUNT(CASE WHEN event_type = 'compliance' THEN 1 END) as compliance_count,
    COUNT(DISTINCT credential_id) as unique_credentials,
    MIN(timestamp) as first_event,
    MAX(timestamp) as last_event
FROM audit_events;

-- ============================================================================
-- Data Integrity Triggers
-- ============================================================================

-- Trigger to prevent orphaned evidence entries if CRC is deleted
-- (This is handled by ON DELETE SET NULL in foreign key, but documented here)

-- Trigger to validate evidence chain linkage
CREATE TRIGGER IF NOT EXISTS validate_evidence_chain
BEFORE INSERT ON evidence_entries
FOR EACH ROW
WHEN NEW.previous_entry_id IS NOT NULL
BEGIN
    SELECT RAISE(ABORT, 'Previous entry does not exist')
    WHERE NOT EXISTS (
        SELECT 1 FROM evidence_entries WHERE id = NEW.previous_entry_id
    );
END;

-- ============================================================================
-- Cleanup and Maintenance Queries (for reference)
-- ============================================================================

-- Delete expired CRCs (run periodically):
-- DELETE FROM crcs WHERE expires_at < strftime('%s', 'now');

-- Vacuum database (reclaim space after deletions):
-- VACUUM;

-- Analyze database (update query planner statistics):
-- ANALYZE;

-- ============================================================================
-- Initial Schema Version
-- ============================================================================

INSERT OR IGNORE INTO schema_version (version, applied_at, description, checksum)
VALUES (
    1,
    strftime('%s', 'now'),
    'Initial Phase III schema with CRCs, Evidence Chain, and Audit Events',
    '0000000000000000000000000000000000000000000000000000000000000000'
);

-- Enable foreign key constraints (SQLite default is OFF)
PRAGMA foreign_keys = ON;

-- Enable Write-Ahead Logging for better concurrency
PRAGMA journal_mode = WAL;

-- Set synchronous mode to NORMAL for better performance (still crash-safe)
PRAGMA synchronous = NORMAL;

-- Cache size (negative = KB, positive = pages). 64MB cache.
PRAGMA cache_size = -64000;

-- Enable auto-vacuum to prevent database fragmentation
PRAGMA auto_vacuum = INCREMENTAL;
