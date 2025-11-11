@echo off
setlocal enabledelayedexpansion

set MODE=%1
if "%MODE%"=="" set MODE=dev

REM Load environment variables from .env file
if exist .env (
    for /f "usebackq tokens=1,* delims==" %%a in (".env") do (
        set "line=%%a"
        if not "!line:~0,1!"=="#" (
            set "%%a=%%b"
        )
    )
)

REM Set default ports if not in .env
if not defined NOFX_FRONTEND_PORT set NOFX_FRONTEND_PORT=3000
if not defined NOFX_BACKEND_PORT set NOFX_BACKEND_PORT=8080

if "%MODE%"=="dev" (
    echo.
    echo ========================================
    echo   Starting DEVELOPMENT mode
    echo ========================================
    echo.
    echo Backend will run on: http://localhost:%NOFX_BACKEND_PORT%
    echo Frontend will run on: http://localhost:%NOFX_FRONTEND_PORT%
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
    echo   Backend: http://localhost:%NOFX_BACKEND_PORT%
    echo   Frontend: http://localhost:%NOFX_FRONTEND_PORT%
    echo.
    echo   Open your browser to: http://localhost:%NOFX_FRONTEND_PORT%
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
    echo Starting backend on: http://localhost:%NOFX_BACKEND_PORT%
    echo.
    go run main.go

) else (
    echo.
    echo Usage: run.bat [dev^|prod]
    echo.
    echo Modes:
    echo   dev  - Development mode ^(frontend on %NOFX_FRONTEND_PORT%, backend on %NOFX_BACKEND_PORT%^)
    echo   prod - Production mode ^(all on %NOFX_BACKEND_PORT%^)
    echo.
    echo Port Configuration (.env):
    echo   NOFX_FRONTEND_PORT=%NOFX_FRONTEND_PORT%
    echo   NOFX_BACKEND_PORT=%NOFX_BACKEND_PORT%
    echo.
    echo Examples:
    echo   run.bat dev    # Start development mode
    echo   run.bat prod   # Start production mode
    echo.
    exit /b 1
)
