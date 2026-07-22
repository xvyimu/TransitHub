# Frontend image: builds web-console (Vue3) and serves via non-root Nginx with same-origin API proxy.
# Build context MUST be the repository root.
#
# Default React frontend remains deploy/separated/Dockerfile.frontend until cutover gate.
#
# NGINX_IMAGE may be a tag or repo@sha256 digest.

ARG NGINX_IMAGE=nginxinc/nginx-unprivileged@sha256:65e3e85dbaed8ba248841d9d58a899b6197106c23cb0ff1a132b7bfe0547e4c0

FROM node:22-bookworm AS builder

RUN corepack enable && corepack prepare pnpm@11.5.0 --activate

WORKDIR /build/web-console
COPY web-console/package.json web-console/pnpm-lock.yaml ./
RUN pnpm install --frozen-lockfile
COPY web-console/ ./
RUN pnpm run build

FROM ${NGINX_IMAGE}

USER root
RUN apk add --no-cache curl gettext \
    && mkdir -p /etc/nginx/templates /var/log/nginx \
    && chown -R nginx:nginx /etc/nginx /var/cache/nginx /var/log/nginx /usr/share/nginx/html

COPY --from=builder /build/web-console/dist /usr/share/nginx/html
COPY deploy/separated/nginx.conf.template /etc/nginx/templates/nginx.conf.template
COPY deploy/separated/docker-entrypoint.sh /docker-entrypoint.sh
RUN chmod 755 /docker-entrypoint.sh \
    && chown -R nginx:nginx /usr/share/nginx/html

USER nginx
ENV BACKEND_UPSTREAM=backend:3000 \
    DNS_RESOLVER=127.0.0.11 \
    NGINX_PORT=8080 \
    SERVER_NAME=_ \
    CLIENT_MAX_BODY_SIZE=100m \
    PROXY_CONNECT_TIMEOUT=60s \
    PROXY_SEND_TIMEOUT=3600s \
    PROXY_READ_TIMEOUT=3600s

EXPOSE 8080
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
  CMD curl -fsS "http://127.0.0.1:${NGINX_PORT}/frontend-healthz" | grep -q '"status":"ok"'

ENTRYPOINT ["/docker-entrypoint.sh"]
