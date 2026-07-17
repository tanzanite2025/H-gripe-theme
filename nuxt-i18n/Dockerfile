FROM node:24-alpine AS build
WORKDIR /app
COPY package*.json ./
RUN npm ci
COPY . .
ARG NUXT_PUBLIC_API_BASE=/api/v1
ARG NITRO_PRESET=node-server
ARG API_INTERNAL_ORIGIN=http://api:9000
ARG CDN_URL=
ENV NUXT_PUBLIC_API_BASE=${NUXT_PUBLIC_API_BASE}
ENV NITRO_PRESET=${NITRO_PRESET}
ENV API_INTERNAL_ORIGIN=${API_INTERNAL_ORIGIN}
ENV CDN_URL=${CDN_URL}
RUN npm run build

FROM node:24-alpine
WORKDIR /app
COPY --from=build /app/.output ./.output
ENV NODE_ENV=production
ENV HOST=0.0.0.0
ENV PORT=3000
ENV NUXT_PUBLIC_API_BASE=/api/v1
USER node
EXPOSE 3000
HEALTHCHECK --interval=30s --timeout=10s --start-period=30s --retries=5 \
  CMD wget -q --spider http://127.0.0.1:3000/ || exit 1
CMD ["node", ".output/server/index.mjs"]
