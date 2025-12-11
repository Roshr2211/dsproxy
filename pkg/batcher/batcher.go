package batcher

import (
    "context"
    "sync"
    "time"

    "github.com/yourname/dsproxy/pkg/db"
)

type Batcher struct {
    db        *db.DB
    batchSize int
    interval  time.Duration

    mu    sync.Mutex
    queue []db.Record
    ch    chan struct{}
}

func New(d *db.DB, batchSize int, interval time.Duration) *Batcher {
    return &Batcher{
        db:        d,
        batchSize: batchSize,
        interval:  interval,
        queue:     make([]db.Record, 0, batchSize*2),
        ch:        make(chan struct{}, 1),
    }
}

func (b *Batcher) Enqueue(user, val string, ts int64) {
    b.mu.Lock()
    b.queue = append(b.queue, db.Record{UserID: user, Value: val, Ts: ts})
    shouldFlush := len(b.queue) >= b.batchSize
    b.mu.Unlock()
    if shouldFlush {
        select {
        case b.ch <- struct{}{}:
        default:
        }
    }
}

func (b *Batcher) Run(ctx context.Context) {
    ticker := time.NewTicker(b.interval)
    defer ticker.Stop()
    for {
        select {
        case <-ctx.Done():
            b.flush(ctx)
            return
        case <-b.ch:
            b.flush(ctx)
        case <-ticker.C:
            b.flush(ctx)
        }
    }
}

func (b *Batcher) flush(ctx context.Context) {
    b.mu.Lock()
    if len(b.queue) == 0 {
        b.mu.Unlock()
        return
    }
    toWrite := make([]db.Record, len(b.queue))
    copy(toWrite, b.queue)
    b.queue = b.queue[:0]
    b.mu.Unlock()
    _ = b.db.InsertBatch(ctx, toWrite)
}
