# Admin Account Recovery

Production backoffice passwords are not recoverable from the database. The `users.password` column stores bcrypt hashes only. When an operator cannot log in, rotate the password with `cmd/adminctl` instead of trying to retrieve plaintext.

## The Two-Step Flow

There are two separate cases:

1. **No backoffice account exists yet:** run `adminctl ensure-admin` once in the target environment. The login page and the admin UI cannot create the first account because they require an authenticated admin session.
2. **At least one admin can log in:** use `/ops/admin-accounts` in the admin console to list accounts and create or reset a backoffice account. Every page action is audited as an HTTP operation.

The page is available from **Operations Center -> Admin Accounts**. It never returns or stores a plaintext password.

## DEV Bootstrap

For a local Docker Compose environment, start the database first and run the one-off command against the same `app` image:

```powershell
cd go-backend
docker compose up -d postgres redis

$env:SERVER_MODE = "debug"
$env:ADMIN_EMAIL = "admin@example.com"
$env:ADMIN_PASSWORD = "replace-with-a-strong-dev-password"
docker compose run --rm --no-deps `
  -e SERVER_MODE `
  -e ADMIN_EMAIL `
  -e ADMIN_PASSWORD `
  --entrypoint /app/adminctl `
  app ensure-admin

Remove-Item Env:SERVER_MODE, Env:ADMIN_EMAIL, Env:ADMIN_PASSWORD
```

After this succeeds, open the DEV admin console, sign in once, and use **Operations Center -> Admin Accounts** for subsequent account changes.

## Production Reset Or Create

Prefer a mounted secret file for the password:

```sh
export SERVER_MODE=production
export ADMINCTL_CONFIRM=reset-production-admin
export ADMIN_EMAIL=admin@example.com
export ADMIN_PASSWORD_FILE=/run/secrets/backoffice_admin_password
export ADMIN_OPERATOR=ops@example.com

go run ./cmd/adminctl ensure-admin -config config/config.production.yaml
```

For local or emergency use, `ADMIN_PASSWORD` is also accepted:

```sh
export SERVER_MODE=production
export ADMINCTL_CONFIRM=reset-production-admin
export ADMIN_EMAIL=admin@example.com
export ADMIN_PASSWORD='replace-with-a-new-strong-password'

go run ./cmd/adminctl ensure-admin
```

For the production Compose deployment, run the command as a one-off task from the release directory. The database and the API image must be from the same deployment:

```sh
export SERVER_MODE=production
export ADMINCTL_CONFIRM=reset-production-admin
export ADMIN_EMAIL=admin@example.com
export ADMIN_PASSWORD='replace-with-a-new-strong-password'
export ADMIN_OPERATOR=ops@example.com

docker compose --env-file deployment/production.env -f compose.prod.yml run --rm --no-deps \
  -e SERVER_MODE \
  -e ADMINCTL_CONFIRM \
  -e ADMIN_EMAIL \
  -e ADMIN_PASSWORD \
  -e ADMIN_OPERATOR \
  --entrypoint /app/adminctl \
  api ensure-admin -config /app/config/config.production.yaml

unset SERVER_MODE ADMINCTL_CONFIRM ADMIN_EMAIL ADMIN_PASSWORD ADMIN_OPERATOR
```

The command:

- creates the account when `ADMIN_EMAIL` does not exist;
- resets the password and activates the account when `ADMIN_EMAIL` already exists;
- defaults `ADMIN_ROLE` to `admin`;
- writes an `audit_logs` entry without storing or printing the password;
- requires `ADMINCTL_CONFIRM=reset-production-admin` in `release`, `production`, or `prod` mode.

Optional fields:

```sh
export ADMIN_USERNAME=primary-admin
export ADMIN_ROLE=manager
export ADMIN_FIRST_NAME=Ops
export ADMIN_LAST_NAME=Admin
export ADMIN_LOCALE=en
```

After the command succeeds, log in to `/api/admin` or the Vue admin console with `ADMIN_EMAIL` and the new password.

## Container Usage

The production Docker image includes `/app/adminctl`. Run it as a one-off task or Kubernetes Job with the same database environment variables as the API container:

```sh
/app/adminctl ensure-admin -config /app/config/config.production.yaml
```

## Inspect Existing Backoffice Accounts

Use this only to identify accounts. It does not reveal passwords.

```sql
SELECT id, email, username, role, status, created_at
FROM users
WHERE role IN ('admin', 'manager', 'editor', 'support')
ORDER BY id;
```
