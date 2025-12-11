package cache

import (
	"context"
	"testing"
)

func TestCache_SetAndGet(t *testing.T) {
	c := New("localhost:6379")
	ctx := context.Background()

	tests := []struct {
		name    string
		key     string
		value   string
		wantErr bool
	}{
		{
			name:    "set and get simple value",
			key:     "test_key_1",
			value:   "test_value_1",
			wantErr: false,
		},
		{
			name:    "set and get with special characters",
			key:     "test:key:2",
			value:   "value with spaces",
			wantErr: false,
		},
		{
			name:    "set and get empty value",
			key:     "test_key_3",
			value:   "",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set value
			err := c.Set(ctx, tt.key, tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("Set() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// Get value
			got, err := c.Get(ctx, tt.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if got != tt.value {
				t.Errorf("Get() = %v, want %v", got, tt.value)
			}
		})
	}
}

func TestCache_GetNonExistent(t *testing.T) {
	c := New("localhost:6379")
	ctx := context.Background()

	_, err := c.Get(ctx, "nonexistent_key_12345")
	if err == nil {
		t.Error("Expected error when getting non-existent key, got nil")
	}
}

func TestCache_Expiration(t *testing.T) {
	c := New("localhost:6379")
	ctx := context.Background()

	key := "test_expiration_key"
	value := "test_value"

	// Set value
	err := c.Set(ctx, key, value)
	if err != nil {
		t.Fatalf("Set() error = %v", err)
	}

	// Get immediately (should work)
	got, err := c.Get(ctx, key)
	if err != nil {
		t.Errorf("Get() error = %v", err)
	}
	if got != value {
		t.Errorf("Get() = %v, want %v", got, value)
	}

	// Note: Actual expiration test would require waiting 5 minutes or mocking
	// This is just a basic smoke test
}

func TestCache_Update(t *testing.T) {
	c := New("localhost:6379")
	ctx := context.Background()

	key := "test_update_key"
	value1 := "original_value"
	value2 := "updated_value"

	// Set initial value
	err := c.Set(ctx, key, value1)
	if err != nil {
		t.Fatalf("Set() error = %v", err)
	}

	// Update value
	err = c.Set(ctx, key, value2)
	if err != nil {
		t.Fatalf("Set() update error = %v", err)
	}

	// Get updated value
	got, err := c.Get(ctx, key)
	if err != nil {
		t.Errorf("Get() error = %v", err)
	}
	if got != value2 {
		t.Errorf("Get() = %v, want %v", got, value2)
	}
}

func TestCache_ConcurrentAccess(t *testing.T) {
	c := New("localhost:6379")
	ctx := context.Background()

	numGoroutines := 10
	key := "concurrent_test_key"

	// Concurrent writes
	done := make(chan bool, numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			err := c.Set(ctx, key, "value_from_goroutine")
			if err != nil {
				t.Errorf("Concurrent Set() error = %v", err)
			}
			done <- true
		}(i)
	}

	// Wait for all writes
	for i := 0; i < numGoroutines; i++ {
		<-done
	}

	// Should be able to read
	_, err := c.Get(ctx, key)
	if err != nil {
		t.Errorf("Get() after concurrent writes error = %v", err)
	}
}

func TestCache_New(t *testing.T) {
	tests := []struct {
		name string
		addr string
	}{
		{
			name: "localhost",
			addr: "localhost:6379",
		},
		{
			name: "with IP",
			addr: "127.0.0.1:6379",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := New(tt.addr)
			if c == nil {
				t.Error("New() returned nil")
			}
			if c.client == nil {
				t.Error("New() client is nil")
			}
		})
	}
}
