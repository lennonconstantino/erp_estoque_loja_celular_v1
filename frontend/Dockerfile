# ── build ──────────────────────────────────────────────────────────────────────
FROM node:20-alpine AS builder

RUN corepack enable && corepack prepare pnpm@9 --activate

WORKDIR /app

COPY package.json pnpm-lock.yaml pnpm-workspace.yaml ./
RUN pnpm install --frozen-lockfile

ARG VITE_API_BASE_URL=http://localhost:8080
ENV VITE_API_BASE_URL=${VITE_API_BASE_URL}

COPY . .
RUN pnpm build

# ── serve ──────────────────────────────────────────────────────────────────────
FROM nginx:alpine

COPY --from=builder /app/dist /usr/share/nginx/html
COPY nginx.conf /etc/nginx/conf.d/default.conf

EXPOSE 80

# O Railway injeta $PORT em runtime e faz o healthcheck nessa porta. O nginx.conf
# declara `listen 80;`; reescrevemos para $PORT no boot (fallback 80 p/ local/docker).
CMD ["/bin/sh","-c","sed -i \"s/listen 80;/listen ${PORT:-80};/\" /etc/nginx/conf.d/default.conf && exec nginx -g 'daemon off;'"]
