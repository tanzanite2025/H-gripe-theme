FROM node:24-trixie-slim

ENV NODE_ENV=production \
    PORT=8080 \
    CHROME_PATH=/usr/bin/chromium \
    HOME=/tmp/runner-home \
    XDG_CONFIG_HOME=/tmp/runner-config \
    XDG_CACHE_HOME=/tmp/runner-cache \
    XDG_RUNTIME_DIR=/tmp/runner-runtime

RUN apt-get update \
    && apt-get upgrade -y \
    && apt-get install -y --no-install-recommends \
        ca-certificates \
        chromium \
        dumb-init \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY package.json package-lock.json ./
RUN npm ci --omit=dev --ignore-scripts \
    && rm -rf /usr/local/lib/node_modules/npm \
    && rm -f /usr/local/bin/npm /usr/local/bin/npx /usr/local/bin/corepack

COPY src ./src

RUN groupadd --gid 10001 runner \
    && useradd --uid 10001 --gid runner --create-home --shell /usr/sbin/nologin runner \
    && chown -R runner:runner /app

USER runner

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=5s --start-period=20s --retries=3 \
    CMD node -e "fetch('http://127.0.0.1:8080/healthz').then((r) => process.exit(r.ok ? 0 : 1)).catch(() => process.exit(1))"

ENTRYPOINT ["dumb-init", "--"]
CMD ["node", "src/server.mjs"]
