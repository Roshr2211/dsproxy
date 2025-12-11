package handler

import (
    // "context"
    "encoding/json"
    "net/http"
    "time"

    "github.com/jackc/pgx/v5"
    "github.com/yourname/dsproxy/pkg/cache"
    "github.com/yourname/dsproxy/pkg/db"
    "github.com/yourname/dsproxy/pkg/batcher"
    "github.com/prometheus/client_golang/prometheus/promhttp"
)

type Handler struct {
    db      *db.DB
    cache   *cache.Cache
    batcher *batcher.Batcher
}

func New(d *db.DB, c *cache.Cache, b *batcher.Batcher) *Handler {
    return &Handler{db: d, cache: c, batcher: b}
}

func (h *Handler) Routes() http.Handler {
    mux := http.NewServeMux()
    mux.HandleFunc("/write", h.writeHandler)
    mux.HandleFunc("/read", h.readHandler)
    mux.Handle("/metrics", promhttp.Handler())
    return mux
}

type WriteReq struct {
    UserID string `json:"user_id"`
    Value  string `json:"value"`
    Ts     int64  `json:"ts,omitempty"`
}

func (h *Handler) writeHandler(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    if r.Method != http.MethodPost {
        http.Error(w, "method", http.StatusMethodNotAllowed)
        return
    }
    var req WriteReq
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }
    if req.Ts == 0 {
        req.Ts = time.Now().Unix()
    }

    // write-through to cache
    _ = h.cache.Set(ctx, req.UserID, req.Value)

    // enqueue to batcher
    h.batcher.Enqueue(req.UserID, req.Value, req.Ts)

    w.WriteHeader(http.StatusAccepted)
    _, _ = w.Write([]byte("accepted"))
}

func (h *Handler) readHandler(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    user := r.URL.Query().Get("user_id")
    if user == "" {
        http.Error(w, "missing user_id", http.StatusBadRequest)
        return
    }
    // check cache first
    if val, err := h.cache.Get(ctx, user); err == nil && val != "" {
        w.WriteHeader(http.StatusOK)
        _, _ = w.Write([]byte(val))
        return
    }

    // fallback to DB
    rec, err := h.db.GetLatest(ctx, user)
    if err == pgx.ErrNoRows {
        http.Error(w, "not found", http.StatusNotFound)
        return
    } else if err != nil {
        http.Error(w, "db error", http.StatusInternalServerError)
        return
    }
    w.WriteHeader(http.StatusOK)
    _, _ = w.Write([]byte(rec.Value))
}
