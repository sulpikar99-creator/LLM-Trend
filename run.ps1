# PowerShell startup script for LLM-Trend

param(
    [Parameter(Position=0)]
    [ValidateSet("dev", "prod")]
    [string]$Mode = "dev"
)

if ($Mode -eq "dev") {
    Write-Host ""
    Write-Host "========================================"
    Write-Host "  Starting DEVELOPMENT mode"
    Write-Host "========================================"
    Write-Host ""
    Write-Host "Backend will run on: http://localhost:8080"
    Write-Host "Frontend will run on: http://localhost:5173"
    Write-Host ""

    # Start backend
    Write-Host "Starting backend..."
    Start-Process powershell -ArgumentList "-NoExit", "-Command", "go run main.go" -WindowStyle Normal

    # Wait for backend to start
    Start-Sleep -Seconds 3

    # Start frontend
    Write-Host "Starting frontend..."
    Start-Process powershell -ArgumentList "-NoExit", "-Command", "cd web; npm run dev" -WindowStyle Normal

    Write-Host ""
    Write-Host "========================================"
    Write-Host "  Started!"
    Write-Host "========================================"
    Write-Host ""
    Write-Host "  Backend: http://localhost:8080"
    Write-Host "  Frontend: http://localhost:5173"
    Write-Host ""
    Write-Host "  Open your browser to: http://localhost:5173"
    Write-Host ""
    Write-Host "  Close the PowerShell windows to stop"
    Write-Host "========================================"
    Write-Host ""

} elseif ($Mode -eq "prod") {
    Write-Host ""
    Write-Host "========================================"
    Write-Host "  Starting PRODUCTION mode"
    Write-Host "========================================"
    Write-Host ""

    # Build frontend
    Write-Host "Building frontend..."
    Push-Location web
    npm run build
    if ($LASTEXITCODE -ne 0) {
        Write-Host ""
        Write-Host "ERROR: Frontend build failed!" -ForegroundColor Red
        Pop-Location
        exit 1
    }
    Pop-Location

    Write-Host ""
    Write-Host "Frontend built successfully" -ForegroundColor Green
    Write-Host ""
    Write-Host "Starting backend on: http://localhost:8080"
    Write-Host ""

    # Start backend
    go run main.go
}
