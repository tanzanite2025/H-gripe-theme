FROM nginx:1.30.4-alpine3.24

RUN apk upgrade --no-cache \
    && for pkg in nginx-module-image-filter gd libgd tiff curl libcurl; do if apk info -e "$pkg" >/dev/null 2>&1; then apk del --no-network "$pkg"; fi; done \
    && rm -f /etc/nginx/conf.d/default.conf \
    && mkdir -p /tmp/client_temp /tmp/proxy_temp /tmp/fastcgi_temp /tmp/uwsgi_temp /tmp/scgi_temp \
    && chown -R nginx:nginx /var/cache/nginx /tmp

COPY nginx/theme-web.conf /etc/nginx/nginx.conf

USER nginx

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD wget -q --spider http://127.0.0.1:8080/healthz || exit 1

CMD ["nginx", "-g", "daemon off;"]
