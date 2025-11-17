-- Migration 001: Initial Phase III Schema
-- Applied: Phase III Task 1.1
-- Description: Creates initial tables for CRCs, Evidence Chain, and Audit Events
--
-- This migration establishes the foundational schema for SQLite persistence
-- in ACM Phase III, replacing in-memory storage from Phases I and II.

-- Forward migration (UP)
-- This is a copy of the main schema.sql for migration tracking purposes.
-- See ../schema.sql for full schema documentation.

-- ============================================================================
-- Schema Version Management
-- ============================================================================

CREATE TABLE IF NOT EXISTS schema_version (
    version INTEGER PRIMARY KEY,
    applied_at INTEGER NOT NULL,
    description TEXT NOT NULL,
    checksum TEXT NOT NULL
);

-- ============================================================================
-- Core Tables
-- ============================================================================

-- (Tables are created by schema.sql)
-- This migration file documents the initial schema creation
-- for rollback and version tracking purposes.

-- ============================================================================
-- Rollback Migration (DOWN)
-- ============================================================================

-- To rollback this migration, execute:
-- DROP TABLE IF EXISTS audit_events;
-- DROP TABLE IF EXISTS evidence_entries;
-- DROP TABLE IF EXISTS crcs;
-- DROP TABLE IF EXISTS enum_compliance_recommendation;
-- DROP TABLE IF EXISTS enum_evidence_event_type;
-- DROP TABLE IF EXISTS enum_action_type;
-- DROP TABLE IF EXISTS enum_automation_method;
-- DROP TABLE IF EXISTS enum_validation_result;
-- DROP VIEW IF EXISTS v_crc_stats;
-- DROP VIEW IF EXISTS v_evidence_stats;
-- DROP VIEW IF EXISTS v_audit_stats;
-- DROP TRIGGER IF EXISTS validate_evidence_chain;
-- DELETE FROM schema_version WHERE version = 1;

-- Migration checksum (SHA-256 of this file)
-- This should be computed and verified by the migration system
-- Checksum: [TO BE COMPUTED BY MIGRATION SYSTEM]
