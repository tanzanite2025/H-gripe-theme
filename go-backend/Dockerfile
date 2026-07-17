# Multi-stage production build.
FROM golang:1.25.1-alpine AS builder

ARG VERSION=dev
ARG BUILD_TIME=unknown

# 安装必要的工具
RUN apk add --no-cache git ca-certificates tzdata

# 设置工作目录
WORKDIR /app

# 复制go mod文件
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制源代码
COPY . .

# 构建应用
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s -X main.Version=${VERSION} -X main.BuildTime=${BUILD_TIME}" \
    -o /app/server \
    ./cmd/server

# Runtime image is pinned to avoid an unreviewed latest tag.
FROM alpine:3.22

# 安装ca证书和时区数据
RUN apk --no-cache add ca-certificates tzdata

# 创建非root用户
RUN addgroup -g 1000 app && \
    adduser -D -u 1000 -G app app

# 设置工作目录
WORKDIR /app

# 从构建阶段复制二进制文件
COPY --from=builder /app/server /app/server

# Production configuration and SQL migrations are runtime inputs.
COPY --from=builder /app/config /app/config
COPY --from=builder /app/migrations /app/migrations

# 创建必要的目录
RUN mkdir -p /app/logs /app/uploads && \
    chown -R app:app /app

# 切换到非root用户
USER app

# 暴露端口
EXPOSE 9000

# 健康检查
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:9000/health || exit 1

ENTRYPOINT ["/app/server"]
