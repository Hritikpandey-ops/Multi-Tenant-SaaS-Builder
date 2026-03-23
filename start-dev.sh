#!/bin/bash

# Multi-Tenant SaaS - Development Startup Script

echo "Starting Multi-Tenant SaaS Platform..."
echo ""

# Check if we're in the right directory
if [ ! -f "docker-compose.yml" ]; then
    echo "Error: Please run this script from the project root directory"
    exit 1
fi

# Step 1: Start Docker services (PostgreSQL, Redis)
echo "Starting Docker services..."
docker-compose up -d postgres redis nats

# Wait for services to be ready
echo "Waiting for services to be ready..."
sleep 5

# Step 2: Start backend services
echo ""
echo "Starting backend services..."

# Kill any existing processes on ports 3000, 3001
lsof -ti:3000 | xargs kill -9 2>/dev/null || true
lsof -ti:3001 | xargs kill -9 2>/dev/null || true

# Start auth service
echo "  → Auth Service (port 3001)"
DATABASE_PORT=5433 nohup ./bin/auth > /tmp/auth.log 2>&1 &
AUTH_PID=$!

# Wait for auth service to start
sleep 2

# Start gateway
echo "  → API Gateway (port 3000)"
DATABASE_PORT=5433 nohup ./bin/gateway > /tmp/gateway.log 2>&1 &
GATEWAY_PID=$!

# Step 3: Start frontend
echo ""
echo "Starting frontend..."
cd web

# Kill any existing process on port 3001 (frontend uses 3001)
lsof -ti:3001 | xargs kill -9 2>/dev/null || true

# Start Next.js dev server on port 3001
# Note: Using 3002 to avoid conflict with auth service
PORT=3002 nohup npm run dev > /tmp/web.log 2>&1 &
WEB_PID=$!

cd ..

# Save PIDs for cleanup
echo $AUTH_PID > /tmp/multitenant-auth.pid
echo $GATEWAY_PID > /tmp/multitenant-gateway.pid
echo $WEB_PID > /tmp/multitenant-web.pid

echo ""
echo "All services started!"
echo ""
echo "Access the application:"
echo "  → Frontend:        http://localhost:3002"
echo "  → API Gateway:      http://localhost:3000"
echo "  → Auth Service:     http://localhost:3001"
echo "  → PostgreSQL:      localhost:5433"
echo "  → Redis:            localhost:6379"
echo ""
echo "Logs:"
echo "  → Auth:     tail -f /tmp/auth.log"
echo "  → Gateway:  tail -f /tmp/gateway.log"
echo "  → Web:      tail -f /tmp/web.log"
echo ""
echo "To stop all services, run: ./stop-dev.sh"
echo ""
