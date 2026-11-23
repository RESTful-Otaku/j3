# Build context is the repository root so generated docs and API types stay
# available to the web build. Targets: deps, build, prod, dev.
FROM oven/bun:1.4-alpine AS deps
WORKDIR /app
COPY web/package.json web/bun.lock ./
RUN bun install --frozen-lockfile

FROM oven/bun:1.4-alpine AS build
WORKDIR /app
COPY --from=deps /app/node_modules ./node_modules
COPY web/ ./
RUN bun run build

FROM nginx:1.29-alpine AS prod
COPY web/nginx.conf /etc/nginx/conf.d/default.conf
COPY --from=build /app/build /srv/j3
EXPOSE 8080
CMD ["nginx", "-g", "daemon off;"]

FROM oven/bun:1.4-alpine AS dev
WORKDIR /app
COPY --from=deps /app/node_modules ./node_modules
COPY web/ ./
ENV VITE_API_TARGET=http://j3-qa-api:8787
EXPOSE 5173
CMD ["bun", "run", "dev"]
