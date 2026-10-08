# Inventory web — development: Vite dev server with HMR.
# Build context is the repo root; Tilt live-syncs web/src and web/public.
FROM oven/bun:1-alpine AS bun

FROM node:24-alpine
COPY --from=bun /usr/local/bin/bun /usr/local/bin/bun
WORKDIR /app

COPY web/package.json web/bun.lock ./
RUN bun install --frozen-lockfile

COPY web ./

EXPOSE 3000

CMD ["bun", "run", "dev", "--host", "0.0.0.0", "--port", "3000"]
