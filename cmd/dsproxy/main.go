package main

import (
    "context"
    "log"
    "net/http"
    "os"
    "time"

    "github.com/yourname/dsproxy/pkg/batcher"
    "github.com/yourname/dsproxy/pkg/cache"
    "github.com/yourname/dsproxy/pkg/db"
    "github.com/yourname/dsproxy/pkg/handler"
)

func main() {
    ctx := context.Background()
    
    // Read individual env vars and construct DATABASE_URL
    dbHost := os.Getenv("DB_HOST")
    if dbHost == "" {
        dbHost = "localhost"
    }
    dbPort := os.Getenv("DB_PORT")
    if dbPort == "" {
        dbPort = "5432"
    }
    dbUser := os.Getenv("DB_USER")
    if dbUser == "" {
        dbUser = "postgres"
    }
    dbPass := os.Getenv("DB_PASS")
    if dbPass == "" {
        dbPass = "postgres"
    }
    dbName := os.Getenv("DB_NAME")
    if dbName == "" {
        dbName = "mydb"
    }
    
    dbURL := "postgres://" + dbUser + ":" + dbPass + "@" + dbHost + ":" + dbPort + "/" + dbName + "?sslmode=disable"
    
    redisAddr := os.Getenv("REDIS_ADDR")
    if redisAddr == "" {
        redisAddr = "localhost:6379"
    }
    
    proxyPort := os.Getenv("PROXY_PORT")
    if proxyPort == "" {
        proxyPort = "8080"
    }

    pg, err := db.New(ctx, dbURL)
    if err != nil {
        log.Fatalf("failed connect db: %v", err)
    }
    defer pg.Close(ctx)

    cacheClient := cache.New(redisAddr)

    b := batcher.New(pg, 50, 2*time.Second)
    go b.Run(ctx)

    h := handler.New(pg, cacheClient, b)

    srv := &http.Server{
        Addr:    ":" + proxyPort,
        Handler: h.Routes(),
    }

    log.Println("dsProxy running on :" + proxyPort)
    if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
        log.Fatalf("server error: %v", err)
    }

    // graceful shutdown omitted for brevity
    _ = ctx
}
