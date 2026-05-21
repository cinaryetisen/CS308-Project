# Medieval Store — Backend

## Environment variables

| Var | Purpose |
|-----|---------|
| `POSTGRES_DSN` | Postgres connection string |
| `MONGO_URI` | MongoDB connection string |
| `JWT_SECRET` | HMAC-SHA256 signing key for JWTs |
| `DATA_ENC_KEY` | 32-byte AES-256-GCM key for PII at rest, base64-encoded |
| `SMTP_HOST` / `SMTP_PORT` / `SMTP_USERNAME` / `SMTP_PASSWORD` / `SMTP_FROM_EMAIL` | Outbound email (Mailtrap in dev) |
| `IMAGE_BASE_URL` | Base URL for product images (defaults to `http://localhost:8080/images`) |

## PII encryption

`User.TaxID`, `User.HomeAddress`, and `Order.DeliveryAddress` are AES-256-GCM encrypted at rest by GORM hooks in `models/user.go` and `models/order.go`. The cipher uses `DATA_ENC_KEY` from the env.

- **Generate a key:** `openssl rand -base64 32`
- **Format:** the encrypted column stores `enc:<base64(nonce || sealed)>`; rows without that prefix are treated as legacy plaintext and pass through unchanged so a one-time migration isn't required.
- **Rotation:** there is no automated rotation. To rotate, dump+re-insert affected rows under a new key. Tests fall back to an all-zero key when `DATA_ENC_KEY` is unset.
