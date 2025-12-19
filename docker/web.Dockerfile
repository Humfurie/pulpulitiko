# Build stage
FROM node:22-alpine AS builder

WORKDIR /src

# Build args for Nuxt public env vars
ARG NUXT_PUBLIC_API_URL=https://pulpulitiko.humfurie.org/api
ENV NUXT_PUBLIC_API_URL=${NUXT_PUBLIC_API_URL}

# Copy package files
COPY web/package.json web/package-lock.json ./

# Install dependencies (ignore scripts since source code isn't copied yet)
RUN npm ci --ignore-scripts

# Copy source code
COPY web/ ./

# Run postinstall and build
RUN npm run postinstall && npm run build

# Runtime stage
FROM node:22-alpine

WORKDIR /src

# Copy built application
COPY --from=builder /src/.output ./.output

ENV HOST=0.0.0.0
ENV PORT=3000

EXPOSE 3000

CMD ["node", ".output/server/index.mjs"]
