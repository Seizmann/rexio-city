#!/bin/bash
# Restart RexiO City backend with latest code

export PATH="$PATH:/usr/local/go/bin"

echo "Stopping old backend..."
sudo pkill -f "/home/sijan/SijansP/rexio-city/backend/go/api" 2>/dev/null || true
sleep 2

echo "Building new backend..."
cd /home/sijan/SijansP/rexio-city/backend/go
go build -o api ./cmd/api/
chmod +x api

echo "Starting new backend..."
export DATABASE_URL="$(grep DATABASE_URL .env | cut -d= -f2-)"
nohup ./api > /tmp/rexio-backend.log 2>&1 &
sleep 3

echo "Checking backend..."
curl -s http://localhost:10888/api/health && echo " ✓ Backend started!" || echo " ✗ Backend failed"
echo "Logs:"
tail -20 /tmp/rexio-backend.log
