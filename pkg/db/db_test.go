package db

import (
	"context"
	"testing"
)

func TestDB_New(t *testing.T) {
	ctx := context.Background()
	
	tests := []struct {
		name    string
		url     string
		wantErr bool
	}{
		{
			name:    "valid connection",
			url:     "postgres://postgres:postgres@localhost:5432/mydb?sslmode=disable",
			wantErr: false,
		},
		{
			name:    "invalid url format",
			url:     "invalid://url",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, err := New(ctx, tt.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && db != nil {
				defer db.Close(ctx)
			}
		})
	}
}

func TestDB_InsertBatch(t *testing.T) {
	ctx := context.Background()
	db, err := New(ctx, "postgres://postgres:postgres@localhost:5432/mydb?sslmode=disable")
	if err != nil {
		t.Skip("Skipping test: database not available")
	}
	defer db.Close(ctx)

	tests := []struct {
		name    string
		records []Record
		wantErr bool
	}{
		{
			name: "insert single record",
			records: []Record{
				{UserID: "test_user_1", Value: "test_value_1", Ts: 1000},
			},
			wantErr: false,
		},
		{
			name: "insert multiple records",
			records: []Record{
				{UserID: "test_user_2", Value: "test_value_2", Ts: 2000},
				{UserID: "test_user_3", Value: "test_value_3", Ts: 3000},
			},
			wantErr: false,
		},
		{
			name:    "insert empty batch",
			records: []Record{},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := db.InsertBatch(ctx, tt.records)
			if (err != nil) != tt.wantErr {
				t.Errorf("InsertBatch() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDB_GetLatest(t *testing.T) {
	ctx := context.Background()
	db, err := New(ctx, "postgres://postgres:postgres@localhost:5432/mydb?sslmode=disable")
	if err != nil {
		t.Skip("Skipping test: database not available")
	}
	defer db.Close(ctx)

	// Insert test data
	testUser := "get_latest_test_user"
	records := []Record{
		{UserID: testUser, Value: "old_value", Ts: 1000},
		{UserID: testUser, Value: "latest_value", Ts: 2000},
	}
	err = db.InsertBatch(ctx, records)
	if err != nil {
		t.Fatalf("Failed to insert test data: %v", err)
	}

	// Test GetLatest
	rec, err := db.GetLatest(ctx, testUser)
	if err != nil {
		t.Errorf("GetLatest() error = %v", err)
		return
	}

	if rec.Value != "latest_value" {
		t.Errorf("GetLatest() value = %v, want %v", rec.Value, "latest_value")
	}
	if rec.Ts != 2000 {
		t.Errorf("GetLatest() ts = %v, want %v", rec.Ts, 2000)
	}
}

func TestDB_GetLatest_NotFound(t *testing.T) {
	ctx := context.Background()
	db, err := New(ctx, "postgres://postgres:postgres@localhost:5432/mydb?sslmode=disable")
	if err != nil {
		t.Skip("Skipping test: database not available")
	}
	defer db.Close(ctx)

	_, err = db.GetLatest(ctx, "nonexistent_user_xyz123")
	if err == nil {
		t.Error("Expected error for non-existent user, got nil")
	}
}

func TestDB_Close(t *testing.T) {
	ctx := context.Background()
	db, err := New(ctx, "postgres://postgres:postgres@localhost:5432/mydb?sslmode=disable")
	if err != nil {
		t.Skip("Skipping test: database not available")
	}

	// Should not panic
	db.Close(ctx)
}

func TestRecord_Struct(t *testing.T) {
	rec := Record{
		UserID: "test_user",
		Value:  "test_value",
		Ts:     123456,
	}

	if rec.UserID != "test_user" {
		t.Errorf("UserID = %v, want %v", rec.UserID, "test_user")
	}
	if rec.Value != "test_value" {
		t.Errorf("Value = %v, want %v", rec.Value, "test_value")
	}
	if rec.Ts != 123456 {
		t.Errorf("Ts = %v, want %v", rec.Ts, 123456)
	}
}
