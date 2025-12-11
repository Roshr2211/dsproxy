package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/yourname/dsproxy/pkg/batcher"
	"github.com/yourname/dsproxy/pkg/cache"
	"github.com/yourname/dsproxy/pkg/db"
)

func TestWriteHandler(t *testing.T) {
	// Setup mock components (in real tests, use mocks or test containers)
	ctx := context.Background()
	testDB, err := db.New(ctx, "postgres://postgres:postgres@localhost:5432/mydb?sslmode=disable")
	if err != nil {
		t.Skip("Skipping test: database not available")
	}
	defer testDB.Close(ctx)

	testCache := cache.New("localhost:6379")
	testBatcher := batcher.New(testDB, 10, 1*time.Second)
	go testBatcher.Run(ctx)

	h := New(testDB, testCache, testBatcher)

	tests := []struct {
		name       string
		body       string
		wantStatus int
	}{
		{
			name:       "valid request",
			body:       `{"user_id":"test1","value":"hello"}`,
			wantStatus: http.StatusAccepted,
		},
		{
			name:       "valid request with timestamp",
			body:       `{"user_id":"test2","value":"world","ts":1234567890}`,
			wantStatus: http.StatusAccepted,
		},
		{
			name:       "invalid json",
			body:       `{invalid}`,
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/write", bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			h.writeHandler(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("writeHandler() status = %v, want %v", w.Code, tt.wantStatus)
			}
		})
	}
}

func TestReadHandler(t *testing.T) {
	ctx := context.Background()
	testDB, err := db.New(ctx, "postgres://postgres:postgres@localhost:5432/mydb?sslmode=disable")
	if err != nil {
		t.Skip("Skipping test: database not available")
	}
	defer testDB.Close(ctx)

	testCache := cache.New("localhost:6379")
	testBatcher := batcher.New(testDB, 10, 1*time.Second)

	h := New(testDB, testCache, testBatcher)

	tests := []struct {
		name       string
		userID     string
		wantStatus int
	}{
		{
			name:       "missing user_id",
			userID:     "",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "user not found",
			userID:     "nonexistent",
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/read?user_id="+tt.userID, nil)
			w := httptest.NewRecorder()

			h.readHandler(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("readHandler() status = %v, want %v", w.Code, tt.wantStatus)
			}
		})
	}
}

func TestWriteRequest_JSON(t *testing.T) {
	tests := []struct {
		name    string
		json    string
		wantErr bool
	}{
		{
			name:    "valid json",
			json:    `{"user_id":"u1","value":"v1"}`,
			wantErr: false,
		},
		{
			name:    "valid json with timestamp",
			json:    `{"user_id":"u1","value":"v1","ts":123}`,
			wantErr: false,
		},
		{
			name:    "invalid json",
			json:    `{invalid}`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var req WriteReq
			err := json.Unmarshal([]byte(tt.json), &req)
			if (err != nil) != tt.wantErr {
				t.Errorf("JSON unmarshal error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
