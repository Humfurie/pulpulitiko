# Development Dockerfile for Nuxt with hot reload
FROM node:25-alpine

WORKDIR /src

# Install dependencies first (cached layer)
COPY web/package.json web/package-lock.json ./
RUN npm ci

# Source code will be mounted as volume
# Don't copy source here for dev

ENV HOST=0.0.0.0
ENV PORT=3000

EXPOSE 3000

CMD ["npm", "run", "dev"]
