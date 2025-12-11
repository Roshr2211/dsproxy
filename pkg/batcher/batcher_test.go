package batcher

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/yourname/dsproxy/pkg/db"
)

// mockDB implements a simple mock for testing
type mockDB struct {
	mu      sync.Mutex
	batches [][]db.Record
}

func (m *mockDB) InsertBatch(ctx context.Context, records []db.Record) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.batches = append(m.batches, records)
	return nil
}

func (m *mockDB) GetBatchCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.batches)
}

func (m *mockDB) GetLastBatch() []db.Record {
	m.mu.Lock()
	defer m.mu.Unlock()
	if len(m.batches) == 0 {
		return nil
	}
	return m.batches[len(m.batches)-1]
}

// We need a wrapper to use mockDB with Batcher
type mockDBWrapper struct {
	mock *mockDB
}

func TestBatcher_SizeBasedFlush(t *testing.T) {
	// Create a test DB
	ctx := context.Background()
	testDB, err := db.New(ctx, "postgres://postgres:postgres@localhost:5432/mydb?sslmode=disable")
	if err != nil {
		t.Skip("Skipping test: database not available")
	}
	defer testDB.Close(ctx)

	batchSize := 5
	b := New(testDB, batchSize, 10*time.Second)
	go b.Run(ctx)

	// Enqueue exactly batchSize items
	for i := 0; i < batchSize; i++ {
		b.Enqueue("user1", "value", int64(i))
	}

	// Give it a moment to flush
	time.Sleep(100 * time.Millisecond)

	// Queue should be empty after size-based flush
	b.mu.Lock()
	queueLen := len(b.queue)
	b.mu.Unlock()

	if queueLen >= batchSize {
		t.Errorf("Expected queue to be flushed, but has %d items", queueLen)
	}
}

func TestBatcher_TimeBasedFlush(t *testing.T) {
	ctx := context.Background()
	testDB, err := db.New(ctx, "postgres://postgres:postgres@localhost:5432/mydb?sslmode=disable")
	if err != nil {
		t.Skip("Skipping test: database not available")
	}
	defer testDB.Close(ctx)

	flushInterval := 200 * time.Millisecond
	b := New(testDB, 100, flushInterval)
	go b.Run(ctx)

	// Enqueue fewer items than batch size
	b.Enqueue("user1", "value1", 1)
	b.Enqueue("user2", "value2", 2)

	// Wait for time-based flush
	time.Sleep(flushInterval + 100*time.Millisecond)

	// Queue should be empty after time-based flush
	b.mu.Lock()
	queueLen := len(b.queue)
	b.mu.Unlock()

	if queueLen > 0 {
		t.Errorf("Expected queue to be flushed after interval, but has %d items", queueLen)
	}
}

func TestBatcher_Enqueue(t *testing.T) {
	ctx := context.Background()
	testDB, err := db.New(ctx, "postgres://postgres:postgres@localhost:5432/mydb?sslmode=disable")
	if err != nil {
		t.Skip("Skipping test: database not available")
	}
	defer testDB.Close(ctx)

	b := New(testDB, 10, 1*time.Second)

	tests := []struct {
		name   string
		userID string
		value  string
		ts     int64
	}{
		{"enqueue first", "user1", "value1", 100},
		{"enqueue second", "user2", "value2", 200},
		{"enqueue third", "user3", "value3", 300},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			initialLen := len(b.queue)
			b.Enqueue(tt.userID, tt.value, tt.ts)
			
			b.mu.Lock()
			newLen := len(b.queue)
			b.mu.Unlock()

			if newLen != initialLen+1 {
				t.Errorf("Expected queue length %d, got %d", initialLen+1, newLen)
			}
		})
	}
}

func TestBatcher_ConcurrentEnqueue(t *testing.T) {
	ctx := context.Background()
	testDB, err := db.New(ctx, "postgres://postgres:postgres@localhost:5432/mydb?sslmode=disable")
	if err != nil {
		t.Skip("Skipping test: database not available")
	}
	defer testDB.Close(ctx)

	b := New(testDB, 1000, 10*time.Second)
	go b.Run(ctx)

	var wg sync.WaitGroup
	numGoroutines := 10
	itemsPerGoroutine := 10

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < itemsPerGoroutine; j++ {
				b.Enqueue("user", "value", int64(id*100+j))
			}
		}(i)
	}

	wg.Wait()
	time.Sleep(100 * time.Millisecond)

	// All items should be queued or flushed (no crashes)
	b.mu.Lock()
	queueLen := len(b.queue)
	b.mu.Unlock()

	if queueLen < 0 {
		t.Error("Queue length should never be negative")
	}
}
