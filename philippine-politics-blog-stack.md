# Philippine Politics Blog — Tech Stack

## Overview

A high-performance, viral-ready blog platform for Philippine political news and commentary.

## Tech Stack

| Layer | Technology | Purpose |
|-------|------------|---------|
| Backend | Go (Chi or Echo) | API, business logic, high concurrency |
| Frontend | Vue 3 / Nuxt 3 | SSR for SEO, reactive UI |
| Database | PostgreSQL | Primary data store |
| Cache | Redis | Hot article caching, sessions, rate limiting |
| Storage | MinIO | Media files (images, documents) |
| CDN | Cloudflare | Caching, DDoS protection, SSL |

## Project Structure

```
philippine-politics/
├── api/                    # Go backend
│   ├── cmd/
│   │   └── server/
│   │       └── main.go     # Entry point
│   ├── internal/
│   │   ├── handlers/       # HTTP handlers
│   │   ├── services/       # Business logic
│   │   ├── repository/     # Database access
│   │   ├── models/         # Data structures
│   │   └── middleware/     # Auth, logging, rate limiting
│   ├── pkg/
│   │   ├── storage/        # MinIO client
│   │   └── cache/          # Redis client
│   ├── migrations/         # SQL migrations
│   ├── go.mod
│   └── go.sum
│
├── web/                    # Nuxt frontend
│   ├── pages/              # File-based routing
│   │   ├── index.vue
│   │   ├── article/
│   │   │   └── [slug].vue
│   │   ├── category/
│   │   │   └── [name].vue
│   │   └── search.vue
│   ├── components/
│   │   ├── ArticleCard.vue
│   │   ├── Header.vue
│   │   ├── Footer.vue
│   │   └── ShareButtons.vue
│   ├── composables/        # Reusable logic
│   ├── assets/
│   ├── public/
│   ├── nuxt.config.ts
│   └── package.json
│
├── docker/
│   ├── docker-compose.yml
│   ├── docker-compose.prod.yml
│   ├── api.Dockerfile
│   └── web.Dockerfile
│
├── scripts/                # Deployment, maintenance
├── .env.example
└── README.md
```

## Database Schema (Initial)

```sql
-- Articles
CREATE TABLE articles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    slug VARCHAR(255) UNIQUE NOT NULL,
    title VARCHAR(500) NOT NULL,
    summary TEXT,
    content TEXT NOT NULL,
    featured_image VARCHAR(500),
    author_id UUID REFERENCES authors(id),
    category_id UUID REFERENCES categories(id),
    status VARCHAR(20) DEFAULT 'draft',  -- draft, published, archived
    published_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Categories
CREATE TABLE categories (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    slug VARCHAR(100) UNIQUE NOT NULL,
    description TEXT
);

-- Authors
CREATE TABLE authors (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(200) NOT NULL,
    slug VARCHAR(200) UNIQUE NOT NULL,
    bio TEXT,
    avatar VARCHAR(500),
    email VARCHAR(255) UNIQUE
);

-- Tags
CREATE TABLE tags (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    slug VARCHAR(100) UNIQUE NOT NULL
);

-- Article Tags (many-to-many)
CREATE TABLE article_tags (
    article_id UUID REFERENCES articles(id) ON DELETE CASCADE,
    tag_id UUID REFERENCES tags(id) ON DELETE CASCADE,
    PRIMARY KEY (article_id, tag_id)
);

-- Indexes for performance
CREATE INDEX idx_articles_published_at ON articles(published_at DESC);
CREATE INDEX idx_articles_category ON articles(category_id);
CREATE INDEX idx_articles_status ON articles(status);
CREATE INDEX idx_articles_slug ON articles(slug);
```

## Go Dependencies

```
go get github.com/go-chi/chi/v5          # Router
go get github.com/jackc/pgx/v5           # PostgreSQL driver
go get github.com/redis/go-redis/v9      # Redis client
go get github.com/minio/minio-go/v7      # MinIO/S3 client
go get github.com/golang-jwt/jwt/v5      # JWT auth
go get github.com/rs/zerolog             # Logging
go get github.com/go-playground/validator/v10  # Validation
```

## API Endpoints

```
GET    /api/articles              # List articles (paginated)
GET    /api/articles/:slug        # Single article
GET    /api/articles/trending     # Trending articles (Redis-backed)
GET    /api/categories            # List categories
GET    /api/categories/:slug      # Articles by category
GET    /api/tags/:slug            # Articles by tag
GET    /api/search?q=             # Search articles

POST   /api/admin/articles        # Create article (auth required)
PUT    /api/admin/articles/:id    # Update article (auth required)
DELETE /api/admin/articles/:id    # Delete article (auth required)
POST   /api/admin/upload          # Upload media to MinIO (auth required)
```

## Caching Strategy

| Data | TTL | Storage |
|------|-----|---------|
| Article list (homepage) | 5 minutes | Redis |
| Single article | 15 minutes | Redis |
| Trending articles | 10 minutes | Redis |
| Category lists | 30 minutes | Redis |
| Static assets | 1 year | Cloudflare |

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

## Docker Compose (Development)

```yaml
version: '3.8'

services:
  postgres:
    image: postgres:16-alpine
    environment:
      POSTGRES_DB: politics_db
      POSTGRES_USER: politics
      POSTGRES_PASSWORD: localdev
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"

  minio:
    image: minio/minio
    command: server /data --console-address ":9001"
    environment:
      MINIO_ROOT_USER: minioadmin
      MINIO_ROOT_PASSWORD: minioadmin
    ports:
      - "9000:9000"
      - "9001:9001"
    volumes:
      - minio_data:/data

volumes:
  postgres_data:
  minio_data:
```

## Deployment Checklist

- [ ] Set up Cloudflare DNS and proxy
- [ ] Configure SSL (Cloudflare handles this)
- [ ] Set up production PostgreSQL (Supabase, Neon, or self-hosted)
- [ ] Set up production Redis (Upstash, Redis Cloud, or self-hosted)
- [ ] Configure MinIO or switch to Cloudflare R2/S3
- [ ] Set up CI/CD (GitHub Actions)
- [ ] Configure rate limiting for API
- [ ] Set up monitoring (Grafana, Prometheus, or simple uptime checks)
- [ ] Implement backup strategy for database
- [ ] Add social sharing meta tags (Open Graph, Twitter Cards)

## SEO Essentials for Viral Reach

- Server-side rendering via Nuxt
- Proper Open Graph tags for Facebook/Twitter sharing
- Schema.org markup for articles
- Fast load times (target < 2s)
- Mobile-first responsive design
- XML sitemap generation
- robots.txt configuration
