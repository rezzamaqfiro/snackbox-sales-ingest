# SnackBox Sales Ingestion & Inventory Service

Production-ready implementation fixing concurrency panics, race conditions, and N+1 query bottlenecks in `sales-ingest`.

## Key Improvements
- **Thread-safe Cache:** `StockCache` uses `sync.RWMutex` to protect concurrent map reads/writes.
- **Worker Pool Architecture:** Fixed-size worker pool running asynchronously from application startup instead of spawning unbounded goroutines per request.
- **Atomic SQL Decrements:** Stock decrements use atomic `UPDATE ... SET qty = GREATEST(0, qty - $1) RETURNING qty` queries.
- **Batch Processing:** Replaced loop querying with PostgreSQL `ANY($1)` batch lookups.
- **Idempotency:** Enforced `ON CONFLICT (event_id) DO NOTHING` on sale record ingestion.

## Running Tests & App
```bash
go test -v ./...
go run ./...