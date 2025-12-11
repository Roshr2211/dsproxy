package db

import (
    "context"
    "log"

    "github.com/jackc/pgx/v5/pgxpool"
)

type DB struct {
    pool *pgxpool.Pool
}

type Record struct {
    UserID string
    Value  string
    Ts     int64
}

func New(ctx context.Context, url string) (*DB, error) {
    cfg, err := pgxpool.ParseConfig(url)
    if err != nil {
        return nil, err
    }
    pool, err := pgxpool.NewWithConfig(ctx, cfg)
    if err != nil {
        return nil, err
    }
    // init table
    if _, err := pool.Exec(ctx, `CREATE TABLE IF NOT EXISTS user_data (
        user_id TEXT,
        value TEXT,
        ts BIGINT
    );`); err != nil {
        log.Printf("failed create table: %v", err)
    }

    return &DB{pool: pool}, nil
}

func (d *DB) Close(ctx context.Context) {
    d.pool.Close()
}

func (d *DB) InsertBatch(ctx context.Context, rows []Record) error {
    if len(rows) == 0 {
        return nil
    }
    // simple batch insert using COPY or tx
    tx, err := d.pool.Begin(ctx)
    if err != nil {
        return err
    }
    defer tx.Rollback(ctx)

    for _, r := range rows {
        if _, err := tx.Exec(ctx, `INSERT INTO user_data (user_id,value,ts) VALUES ($1,$2,$3)`, r.UserID, r.Value, r.Ts); err != nil {
            return err
        }
    }
    if err := tx.Commit(ctx); err != nil {
        return err
    }
    return nil
}

func (d *DB) GetLatest(ctx context.Context, user string) (*Record, error) {
    row := d.pool.QueryRow(ctx, `SELECT user_id, value, ts FROM user_data WHERE user_id=$1 ORDER BY ts DESC LIMIT 1`, user)
    var r Record
    if err := row.Scan(&r.UserID, &r.Value, &r.Ts); err != nil {
        return nil, err
    }
    return &r, nil
}
