# Inventory web — built Nitro server on a minimal Node runtime.
# Build context is the repo root.
FROM oven/bun:1-alpine AS bun

FROM node:24-alpine AS builder
COPY --from=bun /usr/local/bin/bun /usr/local/bin/bun
WORKDIR /app
ENV NODE_ENV=production

COPY web/package.json web/bun.lock ./
RUN bun install --frozen-lockfile

COPY web ./
RUN bun run build

FROM node:24-alpine
WORKDIR /app
ENV NODE_ENV=production

COPY --from=builder --chown=node:node /app/.output ./.output

USER node
EXPOSE 3000

CMD ["node", ".output/server/index.mjs"]
