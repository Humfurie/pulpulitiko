# Build stage
FROM node:25-alpine AS builder

WORKDIR /src

# Build args for Nuxt public env vars
ARG NUXT_PUBLIC_API_URL=https://pulpulitiko.humfurie.org/api
ENV NUXT_PUBLIC_API_URL=${NUXT_PUBLIC_API_URL}

# Copy package files
COPY web/package.json web/package-lock.json ./

# Install dependencies
RUN npm ci

# Copy source code
COPY web/ ./

# Build Nuxt app
RUN npm run build

# Runtime stage
FROM node:25-alpine

WORKDIR /src

# Copy built application
COPY --from=builder /src/.output ./.output

ENV HOST=0.0.0.0
ENV PORT=3000

EXPOSE 3000

CMD ["node", ".output/server/index.mjs"]
