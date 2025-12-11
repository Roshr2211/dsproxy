# dsProxy

A high-performance Go proxy service with Redis caching, PostgreSQL batching, and Prometheus metrics.

## Features

- **Write-through caching**: Writes are cached in Redis for fast reads
- **Batch processing**: Database writes are batched (50 records or 2 seconds) for optimal throughput
- **Read optimization**: Reads check Redis first, fall back to PostgreSQL
- **Metrics**: Prometheus endpoint at `/metrics` for monitoring
- **Configurable**: Environment-based configuration

## Architecture

```
Client Request
    ↓
HTTP Handler
    ↓
├── Write: Cache in Redis → Queue for batch insert → PostgreSQL
└── Read:  Check Redis → Fallback to PostgreSQL
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
├── cmd/dsproxy/main.go       # Application entry point
├── pkg/
│   ├── ai/claude.go          # Claude AI integration
│   ├── batcher/batcher.go    # Batch write handler
│   ├── cache/cache.go        # Redis cache wrapper
│   ├── db/db.go              # PostgreSQL connection
│   └── handler/handler.go    # HTTP handlers
├── .env.example              # Environment template
├── Dockerfile                # Container build
├── docker-compose.yml        # Multi-container setup
└── go.mod                    # Go dependencies
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

## Development

**Build:**
```powershell
go build -o dsproxy.exe ./cmd/dsproxy
```

**Run tests:**
```powershell
go test ./...
```

**Clean build cache:**
```powershell
go clean -cache
```

## Production Notes

- Use proper secret management (not plain text environment variables)
- Configure connection pooling for PostgreSQL
- Set Redis TTL based on your use case (default: 5 minutes)
- Enable TLS for database connections
- Set up proper logging and alerting
- Use Docker Compose or Kubernetes for orchestration

## License

MIT
