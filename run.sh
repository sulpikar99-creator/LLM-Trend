#!/bin/bash

MODE=${1:-dev}

if [ "$MODE" = "dev" ]; then
    echo "🚀 Starting DEVELOPMENT mode..."
    echo "Backend will run on: http://localhost:8080"
    echo "Frontend will run on: http://localhost:5173"
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
    echo "🌐 Open browser: http://localhost:5173"
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
    echo "🚀 Starting backend on: http://localhost:8080"
    echo ""
    go run main.go

else
    echo "Usage: ./run.sh [dev|prod]"
    echo ""
    echo "Modes:"
    echo "  dev  - Development mode (frontend on 5173, backend on 8080)"
    echo "  prod - Production mode (all on 8080)"
    echo ""
    echo "Examples:"
    echo "  ./run.sh dev    # Start development mode"
    echo "  ./run.sh prod   # Start production mode"
    exit 1
fi
