# dsProxy

A high-performance Go proxy service with Redis caching, PostgreSQL batching, and Prometheus metrics.



## Features

- **Write-through caching**: Writes are cached in Redis for fast reads
- **Batch processing**: Database writes are batched (50 records or 2 seconds) for optimal throughput
- **Read optimization**: Reads check Redis first, fall back to PostgreSQL
- **Metrics**: Prometheus endpoint at `/metrics` for monitoring
- **Configurable**: Environment-based configuration

## Architecture

**Current Architecture (Single Instance):**

```
Client Request
    ↓
HTTP Handler (Single Instance)
    ↓
├── Write: Cache in Redis → In-Memory Queue → Batch to PostgreSQL
└── Read:  Check Redis → Fallback to PostgreSQL
```

**Planned Distributed Architecture:**

```
Load Balancer
    ↓
Multiple App Instances → Message Queue (Kafka/RabbitMQ) → Worker Pool
    ↓                                                          ↓
Redis Cache (Shared)                                    PostgreSQL
```

## Quick Start

### Prerequisites

- Go 1.22+
- Docker (for PostgreSQL and Redis)

### 1. Start Dependencies

```powershell
# Start PostgreSQL
docker run -d --name pg -e POSTGRES_PASSWORD=postgres -p 5432:5432 postgres:15

# Create database
docker exec -it pg psql -U postgres -c "CREATE DATABASE mydb;"

# Start Redis
docker run -d --name redis -p 6379:6379 redis:7
```

### 2. Configure Environment

Copy `.env.example` to `.env` and adjust settings:

```bash
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASS=postgres
DB_NAME=mydb
PROXY_PORT=8081
REDIS_ADDR=localhost:6379
```

### 3. Run the Application

```powershell
# Set environment variables
$env:DB_HOST="localhost"
$env:DB_PORT="5432"
$env:DB_USER="postgres"
$env:DB_PASS="postgres"
$env:DB_NAME="mydb"
$env:PROXY_PORT="8081"
$env:REDIS_ADDR="localhost:6379"

# Run
go run cmd/dsproxy/main.go
```

## API Endpoints

### Write Data

```powershell
Invoke-RestMethod -Uri "http://localhost:8081/write" -Method POST `
  -Body '{"user_id":"user1","value":"hello world"}' `
  -ContentType "application/json"
```

**Request:**
```json
{
  "user_id": "user1",
  "value": "hello world",
  "ts": 1234567890  // optional, auto-generated if omitted
}
```

**Response:** `accepted` (202)

### Read Data

```powershell
Invoke-RestMethod -Uri "http://localhost:8081/read?user_id=user1"
```

**Response:** Latest value for the user

### Metrics

```powershell
Invoke-RestMethod -Uri "http://localhost:8081/metrics"
```

Returns Prometheus-formatted metrics including Go runtime stats, goroutines, memory usage, etc.

## Project Structure

```
dsproxy/
├── cmd/dsproxy/main.go          # Application entry point
├── pkg/
│   ├── ai/claude.go             # Claude AI integration (placeholder)
│   ├── batcher/
│   │   ├── batcher.go           # Batch write handler
│   │   └── batcher_test.go      # Batcher unit tests
│   ├── cache/
│   │   ├── cache.go             # Redis cache wrapper
│   │   └── cache_test.go        # Cache unit tests
│   ├── db/
│   │   ├── db.go                # PostgreSQL connection
│   │   └── db_test.go           # Database unit tests
│   └── handler/
│       ├── handler.go           # HTTP handlers
│       └── handler_test.go      # Handler unit tests
├── test-integration.ps1         # Full integration test suite
├── test-quick.ps1               # Quick smoke tests
├── .env                         # Environment configuration (create from .env.example)
├── .env.example                 # Environment template
├── Dockerfile                   # Container build
├── docker-compose.yml           # Multi-container setup
├── go.mod                       # Go dependencies
├── go.sum                       # Go dependency checksums
├── Makefile                     # Build automation
└── README.md                    # This file
```

## Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `DB_HOST` | localhost | PostgreSQL host |
| `DB_PORT` | 5432 | PostgreSQL port |
| `DB_USER` | postgres | Database user |
| `DB_PASS` | postgres | Database password |
| `DB_NAME` | mydb | Database name |
| `PROXY_PORT` | 8080 | HTTP server port |
| `REDIS_ADDR` | localhost:6379 | Redis connection string |

## Batching Configuration

Edit `cmd/dsproxy/main.go` to adjust batching parameters:

```go
b := batcher.New(pg, 50, 2*time.Second)  // 50 records or 2 seconds
```

## Verify Data

**Check PostgreSQL:**
```powershell
docker exec -it pg psql -U postgres -d mydb -c "SELECT * FROM user_data;"
```

**Check Redis:**
```powershell
docker exec -it redis redis-cli GET user1
```

**View all Redis keys:**
```powershell
docker exec redis redis-cli KEYS "*"
```

**Check database record count:**
```powershell
docker exec pg psql -U postgres -d mydb -c "SELECT COUNT(*) FROM user_data;"
```

**Clean up test data:**
```powershell
docker exec pg psql -U postgres -d mydb -c "DELETE FROM user_data WHERE user_id LIKE 'test%';"
```

## Testing

### Unit Tests

Run all unit tests with coverage:

```powershell
go test ./pkg/... -cover
```

**Test Coverage:**
- Handler: 60.5%
- Batcher: 91.7%
- Cache: 85.7%
- Database: 81.5%

Run specific package tests:
```powershell
go test ./pkg/handler -v
go test ./pkg/batcher -v
go test ./pkg/cache -v
go test ./pkg/db -v
```

### Integration Tests

**Quick Test** - Verify basic functionality (30 seconds):
```powershell
.\test-quick.ps1
```

**Full Integration Suite** - Comprehensive testing (2 minutes):
```powershell
.\test-integration.ps1
```

The integration test suite covers:
- ✅ Write/Read endpoints
- ✅ Redis caching
- ✅ Time-based batching (2s)
- ✅ Size-based batching (50 records)
- ✅ Database fallback
- ✅ Prometheus metrics
- ✅ Error handling

**Prerequisites for integration tests:**
- Application running on port 8081
- PostgreSQL container running (`pg`)
- Redis container running (`redis`)

## Development

**Build:**
```powershell
go build -o dsproxy.exe ./cmd/dsproxy
```

**Run all tests:**
```powershell
go test ./pkg/... -v -cover
```

**Clean build cache:**
```powershell
go clean -cache
```

**Format code:**
```powershell
go fmt ./...
```

**Lint (requires golangci-lint):**
```powershell
golangci-lint run
```

## Production Notes

- Use proper secret management (not plain text environment variables)
- Configure connection pooling for PostgreSQL
- Set Redis TTL based on your use case (default: 5 minutes)
- Enable TLS for database connections
- Set up proper logging and alerting
- Use Docker Compose or Kubernetes for orchestration
- Monitor batch queue size and flush times
- Set up health check endpoints
- Configure proper backup strategies for PostgreSQL
- Use Redis persistence (RDB or AOF) if needed

## Troubleshooting

**Application won't start:**
- Check if PostgreSQL and Redis are running: `docker ps`
- Verify environment variables are set correctly
- Check port 8081 is available: `Get-NetTCPConnection -LocalPort 8081`

**Tests failing:**
- Ensure PostgreSQL is accessible: `docker exec pg psql -U postgres -c "SELECT 1;"`
- Ensure Redis is accessible: `docker exec redis redis-cli PING`
- Run `go clean -cache` and retry

**Batch not flushing:**
- Check logs for errors
- Verify database connection in logs
- Check batch size (default: 50) and interval (default: 2s)

## Performance Tuning

**Batch Configuration:**
```go
// Adjust in cmd/dsproxy/main.go
b := batcher.New(pg, 100, 5*time.Second)  // 100 records or 5 seconds
```

**Redis TTL:**
```go
// Adjust in pkg/cache/cache.go
return c.client.Set(ctx, key, val, 10*time.Minute).Err()  // 10 minutes
```

**PostgreSQL Connection Pool:**
```go
// Add in pkg/db/db.go after pgxpool.NewWithConfig
pool.Config().MaxConns = 20  // Adjust based on load
```

## Future Roadmap

### Distributed Architecture (Planned)

To enable horizontal scaling and high availability, the following enhancements are planned:

**Phase 1: Message Queue Integration**
- Replace in-memory queue with Kafka or RabbitMQ
- Separate API servers from batch processors
- Enable multiple instances without conflicts

**Phase 2: Distributed Coordination**
- Implement distributed locks using Redis
- Add leader election for batch processors
- Coordinate batch flushes across instances

**Phase 3: High Availability**
- Health checks and auto-recovery
- Circuit breakers for external dependencies
- Graceful shutdown and state persistence

**Phase 4: Advanced Features**
- Dead letter queue for failed batches
- Retry mechanisms with exponential backoff
- Real-time metrics dashboard
- Distributed tracing (OpenTelemetry)

### Current Limitations

- **Single instance only**: Running multiple instances will cause duplicate writes
- **In-memory state**: Batch queue is lost on crash or restart
- **No failover**: No automatic recovery if instance fails

Contributions and suggestions for the distributed implementation are welcome!

## License

MIT
