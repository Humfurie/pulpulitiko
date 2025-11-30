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

- Go 1.21+
- Node.js 20+
- Docker & Docker Compose

### Development Setup

1. Clone the repository

2. Start the infrastructure services:
   ```bash
   docker-compose up -d
   ```

3. Set up environment variables:
   ```bash
   cp .env.example .env
   cp api/.env.example api/.env
   ```

4. Run database migrations:
   ```bash
   cd api && make migrate-up
   ```

5. Start the API server:
   ```bash
   cd api && make dev
   ```

6. Start the frontend:
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
