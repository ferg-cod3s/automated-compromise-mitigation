package logging

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestDefaultPerformanceThresholds(t *testing.T) {
	thresholds := DefaultPerformanceThresholds()

	if thresholds.DBQuerySlow != 100*time.Millisecond {
		t.Errorf("DBQuerySlow = %v, want %v", thresholds.DBQuerySlow, 100*time.Millisecond)
	}
	if thresholds.APICallSlow != 1000*time.Millisecond {
		t.Errorf("APICallSlow = %v, want %v", thresholds.APICallSlow, 1000*time.Millisecond)
	}
}

func TestGetThreshold(t *testing.T) {
	thresholds := DefaultPerformanceThresholds()

	tests := []struct {
		opType   OperationType
		expected time.Duration
	}{
		{OpTypeDB, 100 * time.Millisecond},
		{OpTypeDBTx, 500 * time.Millisecond},
		{OpTypeAPI, 1000 * time.Millisecond},
		{OpTypeHTTP, 2000 * time.Millisecond},
		{OpTypeGRPC, 500 * time.Millisecond},
		{OpTypeFileRead, 200 * time.Millisecond},
		{OpTypeFileWrite, 500 * time.Millisecond},
		{OpTypeCrypto, 100 * time.Millisecond},
		{OpTypeVault, 2000 * time.Millisecond},
		{OpTypeGeneric, 100 * time.Millisecond},
	}

	for _, tt := range tests {
		t.Run(string(tt.opType), func(t *testing.T) {
			result := thresholds.GetThreshold(tt.opType)
			if result != tt.expected {
				t.Errorf("GetThreshold(%v) = %v, want %v", tt.opType, result, tt.expected)
			}
		})
	}
}

func TestPerformanceTracker(t *testing.T) {
	tracker := NewPerformanceTracker()

	// Track first operation
	tracker.Track("test_op", 50*time.Millisecond, nil)

	metrics := tracker.GetMetrics("test_op")
	if metrics == nil {
		t.Fatal("Expected metrics, got nil")
	}
	if metrics.Count != 1 {
		t.Errorf("Count = %d, want 1", metrics.Count)
	}
	if metrics.MinDuration != 50*time.Millisecond {
		t.Errorf("MinDuration = %v, want %v", metrics.MinDuration, 50*time.Millisecond)
	}
	if metrics.MaxDuration != 50*time.Millisecond {
		t.Errorf("MaxDuration = %v, want %v", metrics.MaxDuration, 50*time.Millisecond)
	}

	// Track second operation (faster)
	tracker.Track("test_op", 25*time.Millisecond, nil)

	metrics = tracker.GetMetrics("test_op")
	if metrics.Count != 2 {
		t.Errorf("Count = %d, want 2", metrics.Count)
	}
	if metrics.MinDuration != 25*time.Millisecond {
		t.Errorf("MinDuration = %v, want %v", metrics.MinDuration, 25*time.Millisecond)
	}
	if metrics.MaxDuration != 50*time.Millisecond {
		t.Errorf("MaxDuration = %v, want %v", metrics.MaxDuration, 50*time.Millisecond)
	}

	// Track third operation (slower)
	tracker.Track("test_op", 100*time.Millisecond, nil)

	metrics = tracker.GetMetrics("test_op")
	if metrics.Count != 3 {
		t.Errorf("Count = %d, want 3", metrics.Count)
	}
	if metrics.MaxDuration != 100*time.Millisecond {
		t.Errorf("MaxDuration = %v, want %v", metrics.MaxDuration, 100*time.Millisecond)
	}
	if metrics.MinDuration != 25*time.Millisecond {
		t.Errorf("MinDuration = %v, want %v", metrics.MinDuration, 25*time.Millisecond)
	}
}

func TestPerformanceTrackerErrors(t *testing.T) {
	tracker := NewPerformanceTracker()

	// Track successful operation
	tracker.Track("test_op", 10*time.Millisecond, nil)

	// Track failed operation
	tracker.Track("test_op", 20*time.Millisecond, errors.New("test error"))

	metrics := tracker.GetMetrics("test_op")
	if metrics.Count != 2 {
		t.Errorf("Count = %d, want 2", metrics.Count)
	}
	if metrics.ErrorCount != 1 {
		t.Errorf("ErrorCount = %d, want 1", metrics.ErrorCount)
	}
}

func TestMarkSlow(t *testing.T) {
	tracker := NewPerformanceTracker()

	// Track operation
	tracker.Track("test_op", 50*time.Millisecond, nil)

	// Mark as slow
	tracker.MarkSlow("test_op")

	metrics := tracker.GetMetrics("test_op")
	if metrics.SlowCount != 1 {
		t.Errorf("SlowCount = %d, want 1", metrics.SlowCount)
	}

	// Mark as slow again
	tracker.MarkSlow("test_op")

	metrics = tracker.GetMetrics("test_op")
	if metrics.SlowCount != 2 {
		t.Errorf("SlowCount = %d, want 2", metrics.SlowCount)
	}
}

func TestGetAllMetrics(t *testing.T) {
	tracker := NewPerformanceTracker()

	tracker.Track("op1", 10*time.Millisecond, nil)
	tracker.Track("op2", 20*time.Millisecond, nil)
	tracker.Track("op3", 30*time.Millisecond, nil)

	allMetrics := tracker.GetAllMetrics()
	if len(allMetrics) != 3 {
		t.Errorf("GetAllMetrics returned %d metrics, want 3", len(allMetrics))
	}

	if _, exists := allMetrics["op1"]; !exists {
		t.Error("Expected op1 in metrics")
	}
	if _, exists := allMetrics["op2"]; !exists {
		t.Error("Expected op2 in metrics")
	}
	if _, exists := allMetrics["op3"]; !exists {
		t.Error("Expected op3 in metrics")
	}
}

func TestReset(t *testing.T) {
	tracker := NewPerformanceTracker()

	tracker.Track("test_op", 10*time.Millisecond, nil)

	metrics := tracker.GetMetrics("test_op")
	if metrics == nil {
		t.Fatal("Expected metrics before reset")
	}

	tracker.Reset()

	metrics = tracker.GetMetrics("test_op")
	if metrics != nil {
		t.Error("Expected nil metrics after reset")
	}
}

func TestGetMemoryStats(t *testing.T) {
	stats := GetMemoryStats()

	if stats.Alloc == 0 {
		t.Error("Expected non-zero Alloc")
	}
	if stats.TotalAlloc == 0 {
		t.Error("Expected non-zero TotalAlloc")
	}
	if stats.Sys == 0 {
		t.Error("Expected non-zero Sys")
	}
}

func TestGoroutineStats(t *testing.T) {
	count := GoroutineStats()

	if count == 0 {
		t.Error("Expected non-zero goroutine count")
	}

	// Should be at least 1 (the current goroutine)
	if count < 1 {
		t.Errorf("Expected at least 1 goroutine, got %d", count)
	}
}

func TestTakeSnapshot(t *testing.T) {
	tracker := NewPerformanceTracker()

	tracker.Track("op1", 10*time.Millisecond, nil)
	tracker.Track("op2", 20*time.Millisecond, nil)

	snapshot := tracker.TakeSnapshot()

	if snapshot.Timestamp.IsZero() {
		t.Error("Expected non-zero timestamp")
	}
	if snapshot.Memory.Alloc == 0 {
		t.Error("Expected non-zero memory allocation")
	}
	if snapshot.Goroutines == 0 {
		t.Error("Expected non-zero goroutine count")
	}
	if len(snapshot.Operations) != 2 {
		t.Errorf("Expected 2 operations in snapshot, got %d", len(snapshot.Operations))
	}
}

func TestLogMemoryStats(t *testing.T) {
	logger := NewLogger("test")

	// Should not panic
	logger.LogMemoryStats(context.Background())
}

func TestLogGoroutineStats(t *testing.T) {
	logger := NewLogger("test")

	// Should not panic
	logger.LogGoroutineStats(context.Background())
}

func TestLogSnapshot(t *testing.T) {
	logger := NewLogger("test")
	tracker := NewPerformanceTracker()

	tracker.Track("test_op", 10*time.Millisecond, nil)
	tracker.MarkSlow("test_op")

	snapshot := tracker.TakeSnapshot()

	// Should not panic
	logger.LogSnapshot(context.Background(), snapshot)
}

func TestTrackOperationWithType(t *testing.T) {
	logger := NewLogger("test")
	ctx := context.Background()

	tests := []struct {
		name       string
		opType     OperationType
		fn         func() error
		expectWarn bool
	}{
		{
			name:   "fast operation",
			opType: OpTypeDB,
			fn: func() error {
				time.Sleep(1 * time.Millisecond)
				return nil
			},
			expectWarn: false,
		},
		{
			name:   "operation with error",
			opType: OpTypeAPI,
			fn: func() error {
				return errors.New("test error")
			},
			expectWarn: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := logger.TrackOperationWithType(ctx, tt.opType, tt.name, tt.fn)
			if (err != nil) != (tt.fn() != nil) {
				t.Errorf("Unexpected error status")
			}
		})
	}
}

func TestTrackMemoryOperation(t *testing.T) {
	logger := NewLogger("test")
	ctx := context.Background()

	// Simple operation that should not significantly increase memory
	err := logger.TrackMemoryOperation(ctx, "test_op", func() error {
		// Allocate a small amount of memory
		_ = make([]byte, 1024)
		return nil
	})

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

func TestToMB(t *testing.T) {
	tests := []struct {
		bytes    uint64
		expected float64
	}{
		{0, 0},
		{1024 * 1024, 1.0},
		{2 * 1024 * 1024, 2.0},
		{512 * 1024, 0.5},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			result := toMB(tt.bytes)
			if result != tt.expected {
				t.Errorf("toMB(%d) = %f, want %f", tt.bytes, result, tt.expected)
			}
		})
	}
}

func BenchmarkPerformanceTrackerTrack(b *testing.B) {
	tracker := NewPerformanceTracker()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tracker.Track("bench_op", 10*time.Millisecond, nil)
	}
}

func BenchmarkGetMemoryStats(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = GetMemoryStats()
	}
}

func BenchmarkGoroutineStats(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = GoroutineStats()
	}
}
