$ErrorActionPreference = "Stop"

Write-Host "Cleaning up previous containers and volumes..."
docker compose down -v

Write-Host "Starting Postgres and Redis..."
docker compose up -d postgres redis

Write-Host "Waiting for database to be ready (up to 30s)..."
$ready = $false
for ($i=1; $i -le 15; $i++) {
    $status = docker exec $(docker compose ps -q postgres) pg_isready -U gopulse -d gopulse 2>&1
    if ($status -match "accepting connections") {
        $ready = $true
        break
    }
    Start-Sleep -Seconds 2
}

if (-not $ready) {
    Write-Host "Database failed to start in time!"
    exit 1
}

Write-Host "Applying database schema..."
$sql = Get-Content -Path migrations\000001_init_schema.up.sql -Raw
$sql | docker exec -i $(docker compose ps -q postgres) psql -U gopulse -d gopulse

Write-Host "Building GoPulse Server..."
go build -o server.exe cmd/server/main.go

Write-Host "Starting GoPulse Server..."
$env:PORT="8080"
$env:DATABASE_URL="postgres://gopulse:gopulse_password@127.0.0.1:55432/gopulse?sslmode=disable"
$env:REDIS_URL="redis://127.0.0.1:16379/0"
$env:ENVIRONMENT="development"
$env:LOG_LEVEL="info"

$serverProcess = Start-Process -FilePath ".\server.exe" -PassThru -NoNewWindow
Start-Sleep -Seconds 3

try {
    Write-Host "1. Creating a monitored Service..." -ForegroundColor Cyan
    $servicePayload = @{
        name = "PaymentService"
        description = "Core payment processing API"
    } | ConvertTo-Json
    
    $serviceResponse = Invoke-RestMethod -Uri "http://localhost:8080/api/services" -Method Post -Body $servicePayload -ContentType "application/json"
    $apiKey = $serviceResponse.data.apiKey
    $serviceId = $serviceResponse.data.id
    Write-Host "Created Service: PaymentService with API Key: $apiKey"

    Write-Host "`n2. Sending API Request Metrics..." -ForegroundColor Cyan
    
    # Send some successful requests
    for ($i=1; $i -le 3; $i++) {
        $metricPayload = @{
            requestId = [guid]::NewGuid().ToString()
            method = "POST"
            path = "/v1/charge"
            statusCode = 200
            latencyMs = 120 + (Get-Random -Maximum 50)
            timestamp = (Get-Date).ToString("o")
        } | ConvertTo-Json
        
        Invoke-RestMethod -Uri "http://localhost:8080/api/metrics" -Method Post -Body $metricPayload -ContentType "application/json" -Headers @{"X-API-Key"=$apiKey} | Out-Null
        Write-Host "Sent 200 OK metric for /v1/charge"
    }

    # Send a failed request
    $metricPayload = @{
        requestId = [guid]::NewGuid().ToString()
        method = "POST"
        path = "/v1/charge"
        statusCode = 500
        latencyMs = 350
        timestamp = (Get-Date).ToString("o")
    } | ConvertTo-Json
    Invoke-RestMethod -Uri "http://localhost:8080/api/metrics" -Method Post -Body $metricPayload -ContentType "application/json" -Headers @{"X-API-Key"=$apiKey} | Out-Null
    Write-Host "Sent 500 ERROR metric for /v1/charge"

    Start-Sleep -Seconds 1 # Wait for metrics to process/cache

    Write-Host "`n3. Querying GraphQL Analytics..." -ForegroundColor Cyan
    $graphqlQuery = @{
        query = @"
query {
  serviceMetrics(serviceId: "$serviceId", timeRange: null) {
    totalRequests
    averageLatency
  }
}
"@
    } | ConvertTo-Json

    $gqlResponse = Invoke-RestMethod -Uri "http://localhost:8080/graphql" -Method Post -Body $graphqlQuery -ContentType "application/json"
    Write-Host "GraphQL Response:"
    $gqlResponse | ConvertTo-Json -Depth 5

    Write-Host "`n4. Testing Error Handling (Invalid API Key)..." -ForegroundColor Cyan
    try {
        Invoke-RestMethod -Uri "http://localhost:8080/api/metrics" -Method Post -Body "{}" -ContentType "application/json" -Headers @{"X-API-Key"="invalid-key"} | Out-Null
    } catch {
        Write-Host "Caught expected error: $_"
    }

    Write-Host "`nDemo completed successfully!" -ForegroundColor Green
} finally {
    Write-Host "Cleaning up server process..."
    Stop-Process -Id $serverProcess.Id -Force
}
