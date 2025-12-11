package cache

import (
    "context"
    "time"
    "log"

    "github.com/go-redis/redis/v8"
)

type Cache struct {
    client *redis.Client
}

func New(addr string) *Cache {
    opt := &redis.Options{
        Addr: addr,
    }
    client := redis.NewClient(opt)
    // simple ping
    if err := client.Ping(context.Background()).Err(); err != nil {
        log.Printf("redis ping error: %v", err)
    }
    return &Cache{client: client}
}

func (c *Cache) Set(ctx context.Context, key, val string) error {
    return c.client.Set(ctx, key, val, 5*time.Minute).Err()
}

func (c *Cache) Get(ctx context.Context, key string) (string, error) {
    return c.client.Get(ctx, key).Result()
}
