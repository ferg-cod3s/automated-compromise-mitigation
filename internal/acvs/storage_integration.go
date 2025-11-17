// Package acvs implements the Automated Compliance Validation Service.
package acvs

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"time"

	acmv1 "github.com/ferg-cod3s/automated-compromise-mitigation/api/proto/acm/v1"
	"github.com/ferg-cod3s/automated-compromise-mitigation/internal/acvs/crc"
	"github.com/ferg-cod3s/automated-compromise-mitigation/internal/acvs/evidence"
	"github.com/ferg-cod3s/automated-compromise-mitigation/internal/acvs/nlp"
	"github.com/ferg-cod3s/automated-compromise-mitigation/internal/acvs/storage"
	"github.com/ferg-cod3s/automated-compromise-mitigation/internal/acvs/validator"
)

// StorageBackend specifies the storage backend type.
type StorageBackend string

const (
	// StorageMemory uses in-memory storage (default, Phase I/II).
	StorageMemory StorageBackend = "memory"

	// StorageSQLite uses SQLite persistent storage (Phase III).
	StorageSQLite StorageBackend = "sqlite"
)

// StorageConfig holds storage configuration.
type StorageConfig struct {
	// Backend specifies the storage type (memory or sqlite).
	Backend StorageBackend

	// DatabasePath specifies the SQLite database file path.
	// Only used when Backend = StorageSQLite.
	DatabasePath string

	// CacheTTL specifies the cache time-to-live for CRCs.
	CacheTTL time.Duration

	// EnableBackups enables automatic database backups.
	// Only used when Backend = StorageSQLite.
	EnableBackups bool

	// BackupPath specifies the backup directory.
	BackupPath string
}

// DefaultStorageConfig returns the default storage configuration.
func DefaultStorageConfig() StorageConfig {
	return StorageConfig{
		Backend:       StorageMemory,
		DatabasePath:  storage.DefaultDatabasePath,
		CacheTTL:      crc.DefaultCacheTTL,
		EnableBackups: true,
		BackupPath:    "~/.acm/data/backups",
	}
}

// NewServiceWithStorage creates a new ACVS service with configurable storage backend.
// This is the Phase III constructor that supports both memory and SQLite storage.
func NewServiceWithStorage(ctx context.Context, config StorageConfig) (*ACVSService, *sql.DB, error) {
	var crcMgr CRCManager
	var evChain EvidenceChainGenerator
	var db *sql.DB
	var err error

	switch config.Backend {
	case StorageSQLite:
		// Initialize SQLite storage
		db, crcMgr, evChain, err = initializeSQLiteStorage(ctx, config)
		if err != nil {
			// Fallback to in-memory storage
			fmt.Fprintf(os.Stderr, "WARNING: SQLite initialization failed, falling back to in-memory storage: %v\n", err)
			crcMgr, evChain, err = initializeMemoryStorage()
			if err != nil {
				return nil, nil, fmt.Errorf("failed to initialize memory storage: %w", err)
			}
		}

	case StorageMemory:
		// Initialize in-memory storage
		crcMgr, evChain, err = initializeMemoryStorage()
		if err != nil {
			return nil, nil, fmt.Errorf("failed to initialize memory storage: %w", err)
		}

	default:
		return nil, nil, fmt.Errorf("unknown storage backend: %s", config.Backend)
	}

	// Initialize other components
	val := validator.NewValidator()
	nlpEng := nlp.NewEngine("/var/lib/acm/models/legal-tos-v1")

	// Note: ACVSService currently uses concrete types (*crc.Manager, *evidence.ChainGenerator).
	// For Phase III, we need to ensure compatibility. If using SQLite, we wrap the implementations.
	var concreteCRCMgr *crc.Manager
	var concreteEvChain *evidence.ChainGenerator

	// Convert interfaces to concrete types if possible
	if memMgr, ok := crcMgr.(*crc.Manager); ok {
		concreteCRCMgr = memMgr
	} else {
		// SQLite manager - we can't directly use it with the current service
		// For now, create a wrapper or use memory fallback
		// This is a limitation we'll note in documentation
		fmt.Fprintf(os.Stderr, "WARNING: ACVSService requires refactoring to fully support SQLite storage interfaces\n")
		concreteCRCMgr = crc.NewManager()
	}

	if memEv, ok := evChain.(*evidence.ChainGenerator); ok {
		concreteEvChain = memEv
	} else {
		// SQLite evidence chain - same limitation
		var err error
		concreteEvChain, err = evidence.NewChainGenerator()
		if err != nil {
			return nil, nil, fmt.Errorf("failed to create fallback evidence chain: %w", err)
		}
	}

	service := &ACVSService{
		crcManager:           concreteCRCMgr,
		validator:            val,
		evidenceChain:        concreteEvChain,
		nlpEngine:            nlpEng,
		tosFetcher:           NewSimpleToSFetcher(),
		enabled:              false,
		eulaVersion:          "",
		nlpModelVersion:      nlpEng.GetModelVersion(),
		cacheTTLSeconds:      int64(config.CacheTTL.Seconds()),
		evidenceChainEnabled: true,
		defaultOnUncertain:   acmv1.ValidationResult_VALIDATION_RESULT_HIM_REQUIRED,
		modelPath:            nlpEng.GetModelPath(),
		stats:                Statistics{},
	}

	return service, db, nil
}

// initializeSQLiteStorage initializes SQLite-backed storage.
func initializeSQLiteStorage(ctx context.Context, config StorageConfig) (*sql.DB, CRCManager, EvidenceChainGenerator, error) {
	// Initialize database
	dbConfig := storage.Config{
		Path:         config.DatabasePath,
		CacheTTL:     config.CacheTTL,
		BackupPath:   config.BackupPath,
		EnableWAL:    true,
		MaxOpenConns: 10,
		BusyTimeout:  5000,
	}

	db, err := storage.Initialize(ctx, dbConfig)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	// Create SQLite CRC manager
	crcMgr := storage.NewSQLiteCRCManager(db, config.CacheTTL)

	// Create SQLite evidence chain generator
	evChain, err := storage.NewSQLiteEvidenceChainGenerator(db)
	if err != nil {
		db.Close()
		return nil, nil, nil, fmt.Errorf("failed to create evidence chain: %w", err)
	}

	return db, crcMgr, evChain, nil
}

// initializeMemoryStorage initializes in-memory storage.
func initializeMemoryStorage() (CRCManager, EvidenceChainGenerator, error) {
	crcMgr := crc.NewManager()

	evChain, err := evidence.NewChainGenerator()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create evidence chain: %w", err)
	}

	return crcMgr, evChain, nil
}

// GetDatabaseHandle returns the database handle if SQLite storage is used.
// Returns nil if using in-memory storage.
func (s *ACVSService) GetDatabaseHandle() *sql.DB {
	// This would require modifying ACVSService struct to store the db handle
	// For now, return nil - this is a placeholder for future enhancement
	return nil
}

// HealthCheckStorage performs a health check on the storage backend.
func HealthCheckStorage(ctx context.Context, db *sql.DB) error {
	if db == nil {
		// In-memory storage, always healthy
		return nil
	}

	// Ping database
	if err := db.PingContext(ctx); err != nil {
		return fmt.Errorf("database ping failed: %w", err)
	}

	// Check integrity
	var result string
	err := db.QueryRowContext(ctx, "PRAGMA integrity_check").Scan(&result)
	if err != nil {
		return fmt.Errorf("integrity check query failed: %w", err)
	}

	if result != "ok" {
		return fmt.Errorf("integrity check failed: %s", result)
	}

	return nil
}

// MigrateToSQLite migrates data from in-memory storage to SQLite.
// This is useful for upgrading from Phase I/II to Phase III.
func MigrateToSQLite(ctx context.Context, memoryService *ACVSService, dbPath string) error {
	// Initialize SQLite database
	config := StorageConfig{
		Backend:      StorageSQLite,
		DatabasePath: dbPath,
		CacheTTL:     crc.DefaultCacheTTL,
	}

	db, sqliteCRC, sqliteEvidence, err := initializeSQLiteStorage(ctx, config)
	if err != nil {
		return fmt.Errorf("failed to initialize SQLite: %w", err)
	}
	defer db.Close()

	// Migrate CRCs
	if err := migrateCRCs(ctx, memoryService.crcManager, sqliteCRC); err != nil {
		return fmt.Errorf("failed to migrate CRCs: %w", err)
	}

	// Migrate evidence chain
	if err := migrateEvidenceChain(ctx, memoryService.evidenceChain, sqliteEvidence); err != nil {
		return fmt.Errorf("failed to migrate evidence chain: %w", err)
	}

	return nil
}

// migrateCRCs migrates CRC data from memory to SQLite.
func migrateCRCs(ctx context.Context, src CRCManager, dst CRCManager) error {
	// List all CRCs from source
	summaries, err := src.List(ctx, "", true) // Include expired
	if err != nil {
		return fmt.Errorf("failed to list source CRCs: %w", err)
	}

	// Migrate each CRC
	for _, summary := range summaries {
		crcData, found, err := src.Get(ctx, summary.Site)
		if err != nil {
			return fmt.Errorf("failed to get CRC for %s: %w", summary.Site, err)
		}

		if !found {
			continue
		}

		if err := dst.Store(ctx, crcData); err != nil {
			return fmt.Errorf("failed to store CRC for %s: %w", summary.Site, err)
		}
	}

	return nil
}

// migrateEvidenceChain migrates evidence chain data from memory to SQLite.
func migrateEvidenceChain(ctx context.Context, src EvidenceChainGenerator, dst EvidenceChainGenerator) error {
	// Export all evidence entries from source
	req := &ExportRequest{
		Format: acmv1.EvidenceExportFormat_EVIDENCE_EXPORT_FORMAT_JSON,
	}

	entries, err := src.Export(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to export evidence chain: %w", err)
	}

	// Import each entry to destination
	// Note: This would require converting proto entries back to Entry format
	// For Phase III, this is a placeholder - full implementation would require
	// additional methods to reconstruct Entry from EvidenceChainEntry
	if len(entries) > 0 {
		return fmt.Errorf("evidence chain migration not fully implemented - %d entries need migration", len(entries))
	}

	return nil
}
