#!/bin/bash

MODE=${1:-dev}

# Load environment variables from .env
if [ -f .env ]; then
    export $(grep -v '^#' .env | xargs)
fi

# Get ports from environment or use defaults
FRONTEND_PORT=${NOFX_FRONTEND_PORT:-3000}
BACKEND_PORT=${NOFX_BACKEND_PORT:-8080}

if [ "$MODE" = "dev" ]; then
    echo "🚀 Starting DEVELOPMENT mode..."
    echo "Backend will run on: http://localhost:$BACKEND_PORT"
    echo "Frontend will run on: http://localhost:$FRONTEND_PORT"
    echo ""

    # Start backend
    echo "Starting backend..."
    go run main.go &
    BACKEND_PID=$!

    # Wait for backend to start
    sleep 3

    # Start frontend
    echo "Starting frontend..."
    cd web && npm run dev &
    FRONTEND_PID=$!

    echo ""
    echo "✅ Started!"
    echo "   Backend PID: $BACKEND_PID"
    echo "   Frontend PID: $FRONTEND_PID"
    echo ""
    echo "🌐 Open browser: http://localhost:$FRONTEND_PORT"
    echo ""
    echo "Press Ctrl+C to stop..."

    # Wait for interrupt
    trap "echo 'Stopping...'; kill $BACKEND_PID $FRONTEND_PID 2>/dev/null; exit 0" SIGINT SIGTERM
    wait

elif [ "$MODE" = "prod" ]; then
    echo "🚀 Starting PRODUCTION mode..."
    echo ""

    # Build frontend
    echo "📦 Building frontend..."
    cd web && npm run build && cd ..

    if [ $? -ne 0 ]; then
        echo "❌ Frontend build failed!"
        exit 1
    fi

    echo "✅ Frontend built successfully"
    echo ""

    # Start backend (will serve static files)
    echo "🚀 Starting backend on: http://localhost:$BACKEND_PORT"
    echo ""
    go run main.go

else
    echo "Usage: ./run.sh [dev|prod]"
    echo ""
    echo "Modes:"
    echo "  dev  - Development mode (frontend on $FRONTEND_PORT, backend on $BACKEND_PORT)"
    echo "  prod - Production mode (all on $BACKEND_PORT)"
    echo ""
    echo "Port Configuration (.env):"
    echo "  NOFX_FRONTEND_PORT=$FRONTEND_PORT"
    echo "  NOFX_BACKEND_PORT=$BACKEND_PORT"
    echo ""
    echo "Examples:"
    echo "  ./run.sh dev    # Start development mode"
    echo "  ./run.sh prod   # Start production mode"
    exit 1
fi
