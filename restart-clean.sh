#!/bin/bash

echo "🐳 Stopping all containers..."
docker compose down -v

echo "🧹 Removing dangling volumes (optional)..."
docker volume prune -f

echo "🔧 Rebuilding images..."
docker compose build --no-cache

echo "⬆️ Starting fresh containers..."
docker compose up -d

echo "Backend logs (follow):"
docker compose logs -f backend
