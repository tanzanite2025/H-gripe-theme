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
STORAGE_BASE_URL=http://localhost:9200
```

Use object storage for deployed environments:

```env
STORAGE_TYPE=s3
STORAGE_BUCKET=my-bucket
STORAGE_REGION=us-west-2
STORAGE_ACCESS_KEY_ID=...
STORAGE_SECRET_ACCESS_KEY=...
# Optional S3-compatible endpoint, for example https://minio.internal:9000
STORAGE_ENDPOINT=
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

## Showcase Object Policy

Picture Warehouse submissions use two prefixes:

- `showcase/pending/` contains files waiting for moderation and must never be
  publicly readable.
- `showcase/approved/` contains files copied after approval and may be served
  publicly through the configured public URL.

For S3, OSS, and compatible object storage, keep the bucket private and grant
the API service account only the object operations it needs. Do not grant
anonymous `GetObject` access to `showcase/pending/*`, and do not rely on an
ACL comment in application code as a substitute for bucket policy. The
application API still checks the database moderation status before serving
local public uploads, but a public bucket or public CDN origin can bypass that
check.

Before production rollout, verify directly against the storage provider that:

1. Anonymous reads of a pending object return `403` or `404`.
2. The API service account can upload, copy, and delete objects.
3. Approved objects are exposed only through the intended public origin.

## Customer-Service Avatar Policy

Public Chat agent avatars are intentionally not media-library assets. They
live under `customer-service/avatars/` and are referenced only by the current
customer-service agent profile. Each replacement gets a new opaque object key,
so the storage adapters attach `Cache-Control: public, max-age=31536000,
immutable` to avatar objects and local upload serving returns the same header.

For S3, OSS, and CDN deployments, preserve this object metadata when copying
or proxying the prefix. The storage bucket/CDN must permit public reads for the
current avatar delivery path while the application remains responsible for
deleting replaced objects through its outbox cleanup worker.

## Tests

```powershell
cd go-backend
go test ./internal/pkg/storage
```
