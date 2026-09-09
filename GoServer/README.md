# CardPhanton Go Server

Go backend for CardPhanton. It provides WebAuthn/passkey registration and login, cookie-based JWT authentication, card management, administrator workflows, and audit logging.

## Stack

- Go `1.25`
- Gin HTTP router
- WebAuthn via `github.com/go-webauthn/webauthn`
- JWT via `github.com/golang-jwt/jwt/v5`
- SQLite via `modernc.org/sqlite`
- Optional GeoIP lookup via `GeoLite2-City.mmdb`

## Requirements

- Go 1.25 or a compatible toolchain
- A WebAuthn-capable browser/device
- Write access to the server directory for SQLite data

The current implementation uses SQLite. Although the configuration variable is named `DATABASE_URL` and its example value resembles a PostgreSQL URL, empty values and values beginning with `postgres://` or `postgresql://` are mapped to the local `webauthn.db` SQLite file.

## Quick Start

Run these commands from `GoServer`:

```bash
go mod download
copy .env.example .env       # PowerShell: Copy-Item .env.example .env
go run ./cmd/myserver
```

The server loads `.env`, initializes `db/migrations/001_init.sql`, and listens on `http://localhost:8080` by default. The migration runs automatically at startup and adds compatibility columns to older databases when needed.

Build a binary with:

```bash
go build -o myserver ./cmd/myserver
./myserver
```

Start the process from `GoServer`; the migration path is relative to the working directory.

## Configuration

Copy `.env.example` to `.env` and change values as needed:

| Variable                  | Default                    | Purpose                                                                |
| ------------------------- | -------------------------- | ---------------------------------------------------------------------- |
| `DATABASE_URL`            | PostgreSQL-looking example | SQLite path; PostgreSQL-style values use `webauthn.db`.                |
| `JWT_SECRET`              | `default-secret`           | Signs application and WebAuthn-session JWTs.                           |
| `RP_ID`                   | `localhost`                | WebAuthn relying-party ID.                                             |
| `RP_NAME`                 | `WebAuthn App`             | WebAuthn relying-party display name.                                   |
| `RP_ORIGIN`               | `http://localhost:5173`    | WebAuthn relying-party origin.                                         |
| `PORT`                    | `8080`                     | HTTP listen port.                                                      |
| `CLIENT_ORIGIN`           | `http://localhost:5173`    | Comma-separated CORS origins.                                          |
| `ENVIRONMENT`             | `DEV`                      | `PROD` marks cookies as `Secure`; local development uses HTTP cookies. |
| `AUTH_TOKEN_EXPIRY`       | `86400`                    | Authentication-cookie lifetime in seconds.                             |
| `WEBAUTHN_SESSION_EXPIRY` | `300`                      | WebAuthn challenge-session lifetime in seconds.                        |
| `ADMIN_USERNAMES`         | empty                      | Comma-separated, case-insensitive usernames created as `ADMIN`.        |

If `GeoLite2-City.mmdb` is present in the working directory, authentication audit entries may include a location. Without it, location tracking is disabled and authentication still works.

## Authentication

The browser flow is: request options, call the WebAuthn API, then send the resulting credential JSON to the matching verification endpoint. The JSON must be in the `verification` property, not `credential`.

The server uses two HTTP-only cookies:

| Cookie             | Purpose                                              | Lifetime                  |
| ------------------ | ---------------------------------------------------- | ------------------------- |
| `webauthn_session` | Signed, short-lived WebAuthn challenge/session data. | `WEBAUTHN_SESSION_EXPIRY` |
| `auth_token`       | Signed JWT for protected requests.                   | `AUTH_TOKEN_EXPIRY`       |

Cookies use `/` as their path and are `Secure` only when `ENVIRONMENT=PROD`. Protected requests may use the `auth_token` cookie or `Authorization: Bearer <token>`. Browser clients must send credentials, such as `credentials: 'include'` with `fetch` or `withCredentials: true` with Axios.

User accounts use an explicit lifecycle with no ambiguous status values:

```text
PENDING_APPROVAL -> ACTIVE -> SUSPENDED
  |
  +-> rejected: user data is deleted and the admin audit history is retained
```

New normal users are created as `PENDING_APPROVAL`. An administrator can approve them to move them to `ACTIVE`, reject them to permanently remove their user data, or suspend an active account to move it to `SUSPENDED`. Only `ACTIVE` accounts can authenticate. Usernames in `ADMIN_USERNAMES` are created directly as `ACTIVE` administrators.

## API

The base URL is `http://localhost:8080/api` unless `PORT` is changed. JSON errors generally use `{ "error": "..." }`; some validation errors use a structured error with `code`, `title`, and `message`.

### Public authentication routes

| Method | Path                                                    | Notes                                                                                                    |
| ------ | ------------------------------------------------------- | -------------------------------------------------------------------------------------------------------- |
| `GET`  | `/auth/generate-registration-options?username=<name>`   | Requires a new username; sets `webauthn_session`.                                                        |
| `POST` | `/auth/verify-registration`                             | Body: `{ "username": "...", "verification": <PublicKeyCredential JSON>, "nickname": "..." }`.            |
| `GET`  | `/auth/generate-authentication-options?username=<name>` | Loads registered authenticators; sets `webauthn_session`.                                                |
| `POST` | `/auth/verify-authentication`                           | Body: `{ "username": "...", "verification": <PublicKeyCredential JSON> }`; sets `auth_token` on success. |
| `POST` | `/auth/verify-token`                                    | Body must contain `{ "token": "<jwt>" }`; validates the submitted token.                                 |
| `GET`  | `/auth/generate-conditional-options`                    | Starts discoverable/conditional login; sets `webauthn_session`.                                          |

### Authenticated user routes

All routes below require a valid `auth_token` cookie or bearer token.

| Method   | Path                                       | Purpose                                           |
| -------- | ------------------------------------------ | ------------------------------------------------- |
| `GET`    | `/auth/user`                               | Returns `{ "id", "username" }`.                   |
| `GET`    | `/auth/authenticators`                     | Lists the current user’s passkeys.                |
| `GET`    | `/auth/me`                                 | Returns the user and passkey metadata.            |
| `GET`    | `/auth/generate-additional-device-options` | Starts registration for another passkey.          |
| `DELETE` | `/auth/authenticator/:id`                  | Deletes one of the current user’s authenticators. |
| `PUT`    | `/auth/authenticator/:id/nickname`         | Body: `{ "nickname": "New name" }`.               |
| `POST`   | `/auth/logout`                             | Clears both authentication cookies.               |

### Card routes

All card routes require authentication. `GET /cards` returns all cards for authenticated users, while `GET /cards/mine` returns only the current user’s cards. Create, update, and delete operations are scoped to the authenticated user:

| Method   | Path          | Purpose                                                |
| -------- | ------------- | ------------------------------------------------------ |
| `GET`    | `/cards/mine` | Returns the authenticated user’s `{ "cards": [...] }`. |
| `POST`   | `/cards`      | Creates a card and returns HTTP `201`.                 |
| `PUT`    | `/cards/:id`  | Replaces an owned card.                                |
| `DELETE` | `/cards/:id`  | Deletes an owned card.                                 |

Create and update bodies use:

```json
{
  "pan": "4111111111111111",
  "cardholder_name": "John Doe",
  "bank_name": "Example Bank",
  "payment_method_type": "Credit",
  "card_brand": "Visa",
  "product_name": "Rewards",
  "linked_phone_number": "9876543210",
  "exp_month": 12,
  "exp_year": 2030,
  "cvv": 123
}
```

Validation requires a supported PAN pattern, `Credit` or `Debit` payment type, a brand of `Visa`, `Mastercard`, `American Express`, `Discover`, `RuPay`, or `Other`, a 10-digit phone number beginning with 6-9, a valid non-expired month/year (maximum 20 years ahead), and a three-digit CVV. Duplicate PANs return HTTP `409`. Card responses include all request fields plus `id`, `user_id`, `created_at`, and `updated_at`.

### Administrator routes

Every administrator route requires a valid token. Routes marked admin-only also require the authenticated user’s database role to be `ADMIN`.

| Method   | Path                                      | Access        | Purpose                                                                                                  |
| -------- | ----------------------------------------- | ------------- | -------------------------------------------------------------------------------------------------------- |
| `GET`    | `/admin/status`                           | Authenticated | Returns `{ "isAdmin": true\|false }`.                                                                    |
| `GET`    | `/admin/admins`                           | Authenticated | Lists administrator accounts.                                                                            |
| `GET`    | `/admin/users`                            | Admin-only    | Lists normal users; supports `search`, `page` (default 1), and `page_size` (default 10, max 100).        |
| `GET`    | `/admin/users/:id/details`                | Admin-only    | Returns the user, authenticators, user audit logs, admin audit logs, and cards.                          |
| `PUT`    | `/admin/users/:id/approve`                | Admin-only    | Changes `PENDING_APPROVAL` to `ACTIVE` and records `USER_APPROVED`.                                      |
| `DELETE` | `/admin/users/:id/reject`                 | Admin-only    | Deletes a pending user’s cards, activity logs, authenticators, and user record; retains `USER_REJECTED`. |
| `PUT`    | `/admin/users/:id/suspend`                | Admin-only    | Changes `ACTIVE` to `SUSPENDED` and records `USER_SUSPENDED`.                                            |
| `PUT`    | `/admin/users/:id/reactivate`             | Admin-only    | Changes `SUSPENDED` to `ACTIVE` and records `USER_REACTIVATED`.                                          |
| `DELETE` | `/admin/users/:id`                        | Admin-only    | Deletes an active or suspended user and retains `USER_DELETED`.                                          |
| `DELETE` | `/admin/users/:id/authenticators/:authId` | Admin-only    | Removes an authenticator, but never the last remaining one.                                              |
| `DELETE` | `/admin/users/:id/cards/:cardId`          | Admin-only    | Deletes one card belonging to the selected user and records `CARD_DELETED`.                              |
| `GET`    | `/cards`                                  | Authenticated | Returns all cards as `{ "cards": [...] }`.                                                               |

The paginated user response is:

```json
{
  "users": [],
  "search": "",
  "page": 1,
  "page_size": 10,
  "total": 0,
  "total_pages": 0
}
```

### User activity and audit events

User activity logs are stored in `audit_logs` and appear in the user-details activity timeline. A first-time registration records `ACCOUNT_CREATED` and `PASSKEY_ADDED`; a new normal user also records `SENT_FOR_APPROVAL`. Adding another passkey to an existing account records only `PASSKEY_ADDED`.

Administrative actions are stored separately in `admin_audit_logs`. Current actions include `APPROVAL_RECEIVED`, `USER_APPROVED`, `USER_REJECTED`, `USER_SUSPENDED`, `USER_REACTIVATED`, `USER_DELETED`, `AUTH_DELETED`, and `CARD_DELETED`. Records retain administrator and target IDs plus usernames, and target history is fetched by target username or ID so it survives re-registration with a new account ID.

## CORS and Database

`CLIENT_ORIGIN` is a comma-separated allowlist. The server enables `GET`, `POST`, `PUT`, `DELETE`, and `OPTIONS`, allows `Content-Type` and `Authorization`, and enables credentials for matching origins. The frontend origin must appear exactly in `CLIENT_ORIGIN`.

Startup creates `users`, `authenticators`, `audit_logs`, `admin_audit_logs`, and `cards` tables, plus indexes. SQLite foreign keys are enabled and the database uses one open connection. The database file is relative to the process working directory.

Administrative audit records store both IDs and usernames for the administrator and target user. User details load administrative history by `target_user_name`, so a user who registers again with the same username can still see historical administrator actions even though the new account has a different ID. User activity logs are separate and are removed when the user is deleted; administrative audit logs are retained.

Do not commit `.env`, `webauthn.db`, or real credential/card data.

## Example Requests

After obtaining an `auth_token` through the browser WebAuthn flow:

```bash
curl -X GET http://localhost:8080/api/cards \
  -H "Cookie: auth_token=<jwt>"
```

```bash
curl -X POST http://localhost:8080/api/cards \
  -H "Content-Type: application/json" \
  -H "Cookie: auth_token=<jwt>" \
  -d '{
    "pan": "4111111111111111",
    "cardholder_name": "John Doe",
    "bank_name": "Example Bank",
    "payment_method_type": "Credit",
    "card_brand": "Visa",
    "product_name": "Rewards",
    "linked_phone_number": "9876543210",
    "exp_month": 12,
    "exp_year": 2030,
    "cvv": 123
  }'
```

## Development Checks

```bash
go test ./...
go vet ./...
```

There is currently no HTTP end-to-end test suite in this directory. Exercise WebAuthn flows from the paired frontend in a real WebAuthn-capable browser as well.

## Security Notes

- Replace the default `JWT_SECRET` before deployment.
- Use `ENVIRONMENT=PROD` behind HTTPS so authentication cookies are marked `Secure`.
- Restrict `CLIENT_ORIGIN` and `RP_ORIGIN` to trusted origins.
- The current API stores and returns PAN and CVV values. This is sensitive payment data and is not production PCI-compliant storage without encryption, tokenization, and stronger access controls.
- Keep `GeoLite2-City.mmdb` and other operational data outside source control where appropriate.
