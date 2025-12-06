docker compose down --volumes --remove-orphans
docker system prune -a --volumes -f
docker compose up --build -d
