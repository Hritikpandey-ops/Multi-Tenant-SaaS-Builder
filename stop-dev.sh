#!/bin/bash

echo "Stopping Multi-Tenant SaaS Platform..."

# Stop backend services
if [ -f /tmp/multitenant-auth.pid ]; then
    AUTH_PID=$(cat /tmp/multitenant-auth.pid)
    kill $AUTH_PID 2>/dev/null && echo "  → Stopped Auth Service"
    rm /tmp/multitenant-auth.pid
fi

if [ -f /tmp/multitenant-gateway.pid ]; then
    GATEWAY_PID=$(cat /tmp/multitenant-gateway.pid)
    kill $GATEWAY_PID 2>/dev/null && echo "  → Stopped API Gateway"
    rm /tmp/multitenant-gateway.pid
fi

if [ -f /tmp/multitenant-web.pid ]; then
    WEB_PID=$(cat /tmp/multitenant-web.pid)
    kill $WEB_PID 2>/dev/null && echo "  → Stopped Web App"
    rm /tmp/multitenant-web.pid
fi

# Kill any remaining processes on ports
lsof -ti:3000 | xargs kill -9 2>/dev/null || true
lsof -ti:3001 | xargs kill -9 2>/dev/null || true
lsof -ti:3002 | xargs kill -9 2>/dev/null || true

# Stop Docker services
echo ""
echo "Stopping Docker services..."
docker-compose down

echo ""
echo "All services stopped!"
