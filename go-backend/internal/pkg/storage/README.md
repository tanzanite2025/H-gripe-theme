# Storage Package

This package provides backend storage adapters behind a common interface.

## Implementations

- `storage.go` - local filesystem storage
- `s3.go` - S3-compatible storage
- `oss.go` - Aliyun OSS storage

## Configuration

Use local storage for development:

```env
STORAGE_TYPE=local
STORAGE_LOCAL_PATH=./uploads
STORAGE_BASE_URL=http://localhost:9000
```

Use object storage for deployed environments:

```env
STORAGE_TYPE=s3
STORAGE_BUCKET=my-bucket
STORAGE_REGION=us-west-2
STORAGE_ACCESS_KEY_ID=...
STORAGE_SECRET_ACCESS_KEY=...
STORAGE_BASE_URL=https://cdn.example.com
```

```env
STORAGE_TYPE=oss
STORAGE_BUCKET=my-bucket
STORAGE_REGION=cn-hangzhou
STORAGE_ACCESS_KEY_ID=...
STORAGE_SECRET_ACCESS_KEY=...
STORAGE_ENDPOINT=https://oss-cn-hangzhou.aliyuncs.com
STORAGE_BASE_URL=https://cdn.example.com
```

Do not commit real credentials.

## Safety Rules

- Validate upload size and content type before calling storage adapters.
- Never trust a client-provided filename as a storage path.
- Keep generated object keys opaque and collision-resistant.
- Prefer private buckets plus signed URLs unless a file is intentionally public.
- Treat local storage as development infrastructure, not durable production storage.

## Tests

```powershell
cd go-backend
go test ./internal/pkg/storage
```
