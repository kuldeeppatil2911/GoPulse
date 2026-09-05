$ErrorActionPreference = "Stop"

Write-Host "Starting Postgres and Redis..."
docker compose up -d postgres redis

Write-Host "Waiting a moment for DBs..."
Start-Sleep -Seconds 5

Write-Host "Starting GoPulse Server..."
$env:PORT="8080"
$env:DATABASE_URL="postgres://gopulse:gopulse_password@127.0.0.1:55432/gopulse?sslmode=disable"
$env:REDIS_URL="redis://127.0.0.1:16379/0"
$env:ENVIRONMENT="development"
$env:LOG_LEVEL="info"

go run cmd/server/main.go
