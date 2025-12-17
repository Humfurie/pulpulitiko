# Pulpulitiko

A high-performance, viral-ready blog platform for Philippine political news and commentary.

## Tech Stack

| Layer | Technology | Purpose |
|-------|------------|---------|
| Backend | Go (Chi) | API, business logic, high concurrency |
| Frontend | Nuxt 3 | SSR for SEO, reactive UI |
| Database | PostgreSQL | Primary data store |
| Cache | Redis | Hot article caching, sessions, rate limiting |
| Storage | MinIO | Media files (images, documents) |
| CDN | Cloudflare | Caching, DDoS protection, SSL |

## Project Structure

```
pulpulitiko/
├── api/                    # Go backend
│   ├── cmd/
│   │   ├── server/         # API entry point
│   │   ├── migrate/        # Database migrations
│   │   └── seed/           # Database seeding
│   ├── internal/
│   │   ├── handlers/       # HTTP handlers
│   │   ├── services/       # Business logic
│   │   ├── repository/     # Database access
│   │   ├── models/         # Data structures
│   │   ├── middleware/     # Auth, logging, rate limiting
│   │   └── config/         # Configuration
│   ├── pkg/
│   │   ├── storage/        # MinIO client
│   │   └── cache/          # Redis client
│   └── migrations/         # SQL migrations
│
├── web/                    # Nuxt 3 frontend
│   ├── app/
│   │   ├── pages/          # File-based routing
│   │   ├── components/     # Vue components
│   │   ├── composables/    # Reusable logic
│   │   ├── layouts/        # Page layouts
│   │   └── assets/         # Styles
│   └── public/             # Static assets
│
└── docker/                 # Docker configuration
```

## Getting Started

### Prerequisites

- Docker & Docker Compose
- Node.js 20+ (for local frontend development)
- Go 1.21+ (for local API development)

### Quick Start (Docker)

1. Clone the repository

2. Set up environment variables:
   ```bash
   cp .env.example .env
   ```

3. (Optional) Customize admin credentials in `.env`:
   ```env
   ADMIN_EMAIL=your-email@example.com
   ADMIN_PASSWORD=your-secure-password
   ADMIN_NAME=Your Name
   ```

4. Build and start all services:
   ```bash
   docker compose build
   docker compose up -d
   ```

5. Run database migrations and seed:
   ```bash
   # Run migrations
   docker compose exec api migrate -path ./migrations -database "postgres://politics:localdev@postgres:5432/politics_db?sslmode=disable" up

   # Seed the database (creates admin user, roles, categories, tags, sample articles)
   docker compose exec api ./seed
   ```

6. Access the application:
   - **Frontend**: http://localhost:3000
   - **Admin Panel**: http://localhost:3000/admin
   - **API**: http://localhost:8080
   - **MinIO Console**: http://localhost:9001 (minioadmin/minioadmin)

### Default Admin Credentials

If you didn't customize the `.env` file:
- **Email**: `admin@pulpulitiko.com`
- **Password**: `changeme`

### Database Commands

**Using Makefile (Recommended - Laravel-style):**
```bash
cd api

# Run migrations
make migrate-up

# Fresh migration (drop all tables and re-migrate)
make migrate-fresh

# Fresh migration + seed
make migrate-fresh-seed

# Check migration version
make migrate-version

# Rollback one migration
make migrate-down

# Build seed binary
make build
```

**Using Docker (Full commands):**
```bash
# Run migrations
docker compose exec api migrate -path ./migrations -database "postgres://politics:localdev@postgres:5432/politics_db?sslmode=disable" up

# Fresh migration (drop all tables and re-migrate)
docker compose exec api migrate -path ./migrations -database "postgres://politics:localdev@postgres:5432/politics_db?sslmode=disable" down -all
docker compose exec api migrate -path ./migrations -database "postgres://politics:localdev@postgres:5432/politics_db?sslmode=disable" up

# Seed the database
docker compose exec api ./seed

# Check migration version
docker compose exec api migrate -path ./migrations -database "postgres://politics:localdev@postgres:5432/politics_db?sslmode=disable" version

# Rollback one migration
docker compose exec api migrate -path ./migrations -database "postgres://politics:localdev@postgres:5432/politics_db?sslmode=disable" down 1
```

### Local Development (without Docker)

1. Start infrastructure services only:
   ```bash
   docker-compose up -d postgres redis minio
   ```

2. Set up environment variables:
   ```bash
   cp .env.example .env
   cp api/.env.example api/.env
   ```

3. Run database migrations:
   ```bash
   cd api && make migrate-up
   ```

4. Start the API server:
   ```bash
   cd api && make dev
   ```

5. Start the frontend:
   ```bash
   cd web && npm install && npm run dev
   ```

## API Endpoints

### Public

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/articles` | List articles (paginated) |
| GET | `/api/articles/:slug` | Single article |
| GET | `/api/articles/trending` | Trending articles |
| GET | `/api/categories` | List categories |
| GET | `/api/categories/:slug` | Articles by category |
| GET | `/api/tags/:slug` | Articles by tag |
| GET | `/api/search?q=` | Search articles |

### Admin (Auth Required)

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/admin/articles` | Create article |
| PUT | `/api/admin/articles/:id` | Update article |
| DELETE | `/api/admin/articles/:id` | Delete article |
| POST | `/api/admin/upload` | Upload media |

## Environment Variables

```env
# Database
DATABASE_URL=postgres://user:pass@localhost:5432/politics_db

# Redis
REDIS_URL=redis://localhost:6379

# MinIO
MINIO_ENDPOINT=localhost:9000
MINIO_ACCESS_KEY=minioadmin
MINIO_SECRET_KEY=minioadmin
MINIO_BUCKET=politics-media
MINIO_USE_SSL=false

# App
APP_ENV=development
APP_PORT=8080
JWT_SECRET=your-secret-key

# Frontend
NUXT_PUBLIC_API_URL=http://localhost:8080/api
```

## Caching Strategy

| Data | TTL | Storage |
|------|-----|---------|
| Article list (homepage) | 5 minutes | Redis |
| Single article | 15 minutes | Redis |
| Trending articles | 10 minutes | Redis |
| Category lists | 30 minutes | Redis |
| Static assets | 1 year | Cloudflare |

## License

MIT
