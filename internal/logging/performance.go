// Package logging provides performance instrumentation and metrics.
package logging

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"time"
)

// PerformanceThresholds defines what constitutes "slow" operations.
type PerformanceThresholds struct {
	// Database operations
	DBQuerySlow      time.Duration
	DBTransactionSlow time.Duration

	// API operations
	APICallSlow       time.Duration
	HTTPRequestSlow   time.Duration
	GRPCCallSlow      time.Duration

	// File operations
	FileReadSlow      time.Duration
	FileWriteSlow     time.Duration

	// Cryptographic operations
	CryptoOperationSlow time.Duration

	// Password manager operations
	VaultOperationSlow time.Duration
}

// DefaultPerformanceThresholds returns reasonable defaults for slow operation detection.
func DefaultPerformanceThresholds() PerformanceThresholds {
	return PerformanceThresholds{
		DBQuerySlow:         100 * time.Millisecond,
		DBTransactionSlow:   500 * time.Millisecond,
		APICallSlow:         1000 * time.Millisecond,
		HTTPRequestSlow:     2000 * time.Millisecond,
		GRPCCallSlow:        500 * time.Millisecond,
		FileReadSlow:        200 * time.Millisecond,
		FileWriteSlow:       500 * time.Millisecond,
		CryptoOperationSlow: 100 * time.Millisecond,
		VaultOperationSlow:  2000 * time.Millisecond,
	}
}

// OperationType categorizes operations for threshold selection.
type OperationType string

const (
	OpTypeDB          OperationType = "db"
	OpTypeDBTx        OperationType = "db_transaction"
	OpTypeAPI         OperationType = "api"
	OpTypeHTTP        OperationType = "http"
	OpTypeGRPC        OperationType = "grpc"
	OpTypeFileRead    OperationType = "file_read"
	OpTypeFileWrite   OperationType = "file_write"
	OpTypeCrypto      OperationType = "crypto"
	OpTypeVault       OperationType = "vault"
	OpTypeGeneric     OperationType = "generic"
)

// GetThreshold returns the appropriate threshold for an operation type.
func (t PerformanceThresholds) GetThreshold(opType OperationType) time.Duration {
	switch opType {
	case OpTypeDB:
		return t.DBQuerySlow
	case OpTypeDBTx:
		return t.DBTransactionSlow
	case OpTypeAPI:
		return t.APICallSlow
	case OpTypeHTTP:
		return t.HTTPRequestSlow
	case OpTypeGRPC:
		return t.GRPCCallSlow
	case OpTypeFileRead:
		return t.FileReadSlow
	case OpTypeFileWrite:
		return t.FileWriteSlow
	case OpTypeCrypto:
		return t.CryptoOperationSlow
	case OpTypeVault:
		return t.VaultOperationSlow
	default:
		return 100 * time.Millisecond
	}
}

// PerformanceTracker tracks operation performance metrics.
type PerformanceTracker struct {
	mu         sync.RWMutex
	thresholds PerformanceThresholds
	metrics    map[string]*OperationMetrics
}

// OperationMetrics holds performance statistics for an operation type.
type OperationMetrics struct {
	Count         int64
	TotalDuration time.Duration
	MinDuration   time.Duration
	MaxDuration   time.Duration
	SlowCount     int64
	ErrorCount    int64
}

// NewPerformanceTracker creates a new performance tracker.
func NewPerformanceTracker() *PerformanceTracker {
	return &PerformanceTracker{
		thresholds: DefaultPerformanceThresholds(),
		metrics:    make(map[string]*OperationMetrics),
	}
}

// Track tracks an operation and returns its duration.
func (pt *PerformanceTracker) Track(operation string, duration time.Duration, err error) {
	pt.mu.Lock()
	defer pt.mu.Unlock()

	metrics, exists := pt.metrics[operation]
	if !exists {
		metrics = &OperationMetrics{
			MinDuration: duration,
			MaxDuration: duration,
		}
		pt.metrics[operation] = metrics
	}

	metrics.Count++
	metrics.TotalDuration += duration

	if duration < metrics.MinDuration {
		metrics.MinDuration = duration
	}
	if duration > metrics.MaxDuration {
		metrics.MaxDuration = duration
	}

	if err != nil {
		metrics.ErrorCount++
	}
}

// MarkSlow marks an operation as slow.
func (pt *PerformanceTracker) MarkSlow(operation string) {
	pt.mu.Lock()
	defer pt.mu.Unlock()

	if metrics, exists := pt.metrics[operation]; exists {
		metrics.SlowCount++
	}
}

// GetMetrics returns metrics for a specific operation.
func (pt *PerformanceTracker) GetMetrics(operation string) *OperationMetrics {
	pt.mu.RLock()
	defer pt.mu.RUnlock()

	if metrics, exists := pt.metrics[operation]; exists {
		// Return a copy
		metricsCopy := *metrics
		return &metricsCopy
	}
	return nil
}

// GetAllMetrics returns all tracked metrics.
func (pt *PerformanceTracker) GetAllMetrics() map[string]*OperationMetrics {
	pt.mu.RLock()
	defer pt.mu.RUnlock()

	result := make(map[string]*OperationMetrics, len(pt.metrics))
	for k, v := range pt.metrics {
		metricsCopy := *v
		result[k] = &metricsCopy
	}
	return result
}

// Reset clears all metrics.
func (pt *PerformanceTracker) Reset() {
	pt.mu.Lock()
	defer pt.mu.Unlock()
	pt.metrics = make(map[string]*OperationMetrics)
}

// TrackOperationWithType tracks an operation with type-specific threshold checking.
func (l *Logger) TrackOperationWithType(
	ctx context.Context,
	opType OperationType,
	operation string,
	fn func() error,
) error {
	logger := l.WithContext(ctx)
	start := time.Now()

	logger.Debug("operation started",
		"operation", operation,
		"type", string(opType),
	)

	err := fn()
	duration := time.Since(start)

	// Determine threshold
	threshold := DefaultPerformanceThresholds().GetThreshold(opType)

	// Log completion
	if err != nil {
		logger.Error("operation failed",
			"operation", operation,
			"type", string(opType),
			"duration_ms", duration.Milliseconds(),
			"error", err,
		)
		return err
	}

	// Check if slow
	if duration > threshold {
		logger.Warn("slow operation detected",
			"operation", operation,
			"type", string(opType),
			"duration_ms", duration.Milliseconds(),
			"threshold_ms", threshold.Milliseconds(),
		)
	} else {
		logger.Info("operation completed",
			"operation", operation,
			"type", string(opType),
			"duration_ms", duration.Milliseconds(),
		)
	}

	return nil
}

// MemoryStats holds memory usage statistics.
type MemoryStats struct {
	Alloc        uint64  // Bytes allocated and still in use
	TotalAlloc   uint64  // Total bytes allocated (cumulative)
	Sys          uint64  // Bytes obtained from system
	NumGC        uint32  // Number of completed GC cycles
	HeapAlloc    uint64  // Bytes allocated on heap
	HeapInuse    uint64  // Bytes in in-use spans
	StackInuse   uint64  // Bytes in stack spans
	GCCPUPercent float64 // Percentage of CPU time used by GC
}

// GetMemoryStats returns current memory usage statistics.
func GetMemoryStats() MemoryStats {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	return MemoryStats{
		Alloc:        m.Alloc,
		TotalAlloc:   m.TotalAlloc,
		Sys:          m.Sys,
		NumGC:        m.NumGC,
		HeapAlloc:    m.HeapAlloc,
		HeapInuse:    m.HeapInuse,
		StackInuse:   m.StackInuse,
		GCCPUPercent: m.GCCPUFraction * 100,
	}
}

// LogMemoryStats logs current memory statistics.
func (l *Logger) LogMemoryStats(ctx context.Context) {
	stats := GetMemoryStats()
	logger := l.WithContext(ctx)

	logger.Info("memory statistics",
		"alloc_mb", toMB(stats.Alloc),
		"total_alloc_mb", toMB(stats.TotalAlloc),
		"sys_mb", toMB(stats.Sys),
		"num_gc", stats.NumGC,
		"heap_alloc_mb", toMB(stats.HeapAlloc),
		"heap_inuse_mb", toMB(stats.HeapInuse),
		"stack_inuse_mb", toMB(stats.StackInuse),
		"gc_cpu_percent", fmt.Sprintf("%.2f", stats.GCCPUPercent),
	)
}

// toMB converts bytes to megabytes.
func toMB(bytes uint64) float64 {
	return float64(bytes) / 1024 / 1024
}

// GoroutineStats returns the current number of goroutines.
func GoroutineStats() int {
	return runtime.NumGoroutine()
}

// LogGoroutineStats logs the current goroutine count.
func (l *Logger) LogGoroutineStats(ctx context.Context) {
	count := GoroutineStats()
	logger := l.WithContext(ctx)

	logger.Info("goroutine statistics",
		"count", count,
	)
}

// PerformanceSnapshot captures a point-in-time performance snapshot.
type PerformanceSnapshot struct {
	Timestamp   time.Time
	Memory      MemoryStats
	Goroutines  int
	Operations  map[string]*OperationMetrics
}

// TakeSnapshot captures a performance snapshot.
func (pt *PerformanceTracker) TakeSnapshot() PerformanceSnapshot {
	return PerformanceSnapshot{
		Timestamp:  time.Now(),
		Memory:     GetMemoryStats(),
		Goroutines: GoroutineStats(),
		Operations: pt.GetAllMetrics(),
	}
}

// LogSnapshot logs a performance snapshot.
func (l *Logger) LogSnapshot(ctx context.Context, snapshot PerformanceSnapshot) {
	logger := l.WithContext(ctx)

	logger.Info("performance snapshot",
		"timestamp", snapshot.Timestamp.Format(time.RFC3339),
		"memory_alloc_mb", toMB(snapshot.Memory.Alloc),
		"goroutines", snapshot.Goroutines,
		"operations_tracked", len(snapshot.Operations),
	)

	// Log slow operations if any
	for operation, metrics := range snapshot.Operations {
		if metrics.SlowCount > 0 {
			logger.Warn("slow operation summary",
				"operation", operation,
				"total_calls", metrics.Count,
				"slow_calls", metrics.SlowCount,
				"error_calls", metrics.ErrorCount,
				"avg_duration_ms", metrics.TotalDuration.Milliseconds()/metrics.Count,
				"max_duration_ms", metrics.MaxDuration.Milliseconds(),
			)
		}
	}
}

// TrackMemoryOperation tracks memory-sensitive operations and logs if memory spikes.
func (l *Logger) TrackMemoryOperation(ctx context.Context, operation string, fn func() error) error {
	logger := l.WithContext(ctx)

	// Capture initial memory
	initialMem := GetMemoryStats()

	// Run operation
	err := l.TimedOperation(ctx, operation, fn)

	// Capture final memory
	finalMem := GetMemoryStats()

	// Calculate delta
	allocDelta := int64(finalMem.Alloc) - int64(initialMem.Alloc)

	// Log if significant memory increase (>10MB)
	if allocDelta > 10*1024*1024 {
		logger.Warn("memory-intensive operation detected",
			"operation", operation,
			"memory_delta_mb", toMB(uint64(allocDelta)),
			"initial_mb", toMB(initialMem.Alloc),
			"final_mb", toMB(finalMem.Alloc),
		)
	}

	return err
}
