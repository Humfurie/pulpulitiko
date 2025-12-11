# Pulpulitiko Project

## Running with Docker (Production)

```bash
# Start all services
docker compose -f docker-compose.prod.yml up -d

# Run migrations
docker exec pulpulitiko-api /usr/local/bin/migrate -path /app/migrations -database "postgres://politics:localdev@postgres:5432/politics_db?sslmode=disable" up

# Run seed (set your own admin credentials)
docker exec -e ADMIN_EMAIL=<email> -e ADMIN_PASSWORD=<password> -e ADMIN_NAME=<name> pulpulitiko-api /app/seed

# View logs
docker logs -f pulpulitiko-api
docker logs -f pulpulitiko-web

# Stop all services
docker compose -f docker-compose.prod.yml down
```

## Services

| Service | Container | Port | Description |
|---------|-----------|------|-------------|
| PostgreSQL | pulpulitiko-postgres | 5432 (internal) | Database |
| Redis | pulpulitiko-redis | 6379 (internal) | Cache |
| API | pulpulitiko-api | 8080 (internal) | Go backend |
| Web | pulpulitiko-web | 3000 (internal) | Nuxt frontend |

## URLs (via Traefik)

- Frontend: https://pulpulitiko.humfurie.org
- API: https://pulpulitiko.humfurie.org/api

## External Dependencies

- MinIO: minio.humfurie.org (file storage)
- Traefik: External reverse proxy with SSL (uses `proxy` network)

## Environment Variables

Key variables in `.env`:
- `POSTGRES_PASSWORD` - Database password
- `MINIO_ACCESS_KEY` / `MINIO_SECRET_KEY` - MinIO credentials
- `JWT_SECRET` - Authentication secret
- `RESEND_API_KEY` - Email service (optional)

