# Build stage
FROM node:22-alpine AS builder

WORKDIR /src

# Copy package files
COPY web/package.json web/package-lock.json ./

# Install dependencies
RUN npm ci

# Copy source code
COPY web/ ./

# Build Nuxt app
RUN npm run build

# Runtime stage
FROM node:22-alpine

WORKDIR /src

# Copy built application
COPY --from=builder /src/.output ./.output

ENV HOST=0.0.0.0
ENV PORT=3000

EXPOSE 3000

CMD ["node", ".output/server/index.mjs"]
