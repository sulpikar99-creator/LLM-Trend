@echo off
setlocal enabledelayedexpansion

set MODE=%1
if "%MODE%"=="" set MODE=dev

if "%MODE%"=="dev" (
    echo.
    echo ========================================
    echo   Starting DEVELOPMENT mode
    echo ========================================
    echo.
    echo Backend will run on: http://localhost:8080
    echo Frontend will run on: http://localhost:5173
    echo.
    echo Starting backend...
    start "LLM-Trend Backend" cmd /c "go run main.go"

    timeout /t 3 /nobreak >nul

    echo Starting frontend...
    start "LLM-Trend Frontend" cmd /c "cd web && npm run dev"

    echo.
    echo ========================================
    echo   Started!
    echo ========================================
    echo.
    echo   Backend: http://localhost:8080
    echo   Frontend: http://localhost:5173
    echo.
    echo   Open your browser to: http://localhost:5173
    echo.
    echo   Close the terminal windows to stop
    echo ========================================

) else if "%MODE%"=="prod" (
    echo.
    echo ========================================
    echo   Starting PRODUCTION mode
    echo ========================================
    echo.
    echo Building frontend...
    cd web
    call npm run build
    if errorlevel 1 (
        echo.
        echo ERROR: Frontend build failed!
        exit /b 1
    )
    cd ..

    echo.
    echo Frontend built successfully
    echo.
    echo Starting backend on: http://localhost:8080
    echo.
    go run main.go

) else (
    echo.
    echo Usage: run.bat [dev^|prod]
    echo.
    echo Modes:
    echo   dev  - Development mode ^(frontend on 5173, backend on 8080^)
    echo   prod - Production mode ^(all on 8080^)
    echo.
    echo Examples:
    echo   run.bat dev    # Start development mode
    echo   run.bat prod   # Start production mode
    echo.
    exit /b 1
)
