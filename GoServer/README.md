# WebAuthn Server API Documentation

This document provides a complete guide for frontend developers to integrate with the WebAuthn authentication server.

## Base URL

```
http://localhost:8080/api
```

## Environment Variables

```env
# Server
PORT=8080
CLIENT_ORIGIN=http://localhost:5173

# JWT
JWT_SECRET=your-secret-key

# Cookie Settings (in seconds)
AUTH_TOKEN_EXPIRY=86400          # 24 hours
WEBAUTHN_SESSION_EXPIRY=300     # 5 minutes
```

## Authentication

This API uses cookie-based authentication. After login, the server sets two cookies:

| Cookie | Purpose | Expiry |
|--------|---------|--------|
| `auth_token` | JWT for authenticated requests | 24 hours |
| `webauthn_session` | WebAuthn flow session | 5 minutes |

Both cookies are `Secure` and `HttpOnly` for security.

---

## Endpoints

### Authentication Endpoints (Public)

#### 1. Generate Registration Options
Start the WebAuthn registration flow.

**Endpoint:** `GET /api/auth/generate-registration-options`

**Request:**
```
No body required
```

**Response:**
```json
{
  "publicKey": {
    "challenge": "base64-encoded-challenge",
    "rp": {
      "name": "WebAuthn App",
      "id": "localhost"
    },
    "user": {
      "id": "base64-encoded-user-id",
      "name": "username",
      "displayName": "Display Name"
    },
    "pubKeyCredParams": [
      {"type": "public-key", "alg": -7}
    ],
    "timeout": 60000,
    "excludeCredentials": [],
    "authenticatorSelection": {
      "userVerification": "preferred"
    }
  }
}
```

---

#### 2. Verify Registration
Complete the WebAuthn registration.

**Endpoint:** `POST /api/auth/verify-registration`

**Request Body:**
```json
{
  "username": "johndoe",
  "credential": {
    "id": "base64-encoded-credential-id",
    "rawId": "base64-encoded-credential-id",
    "type": "public-key",
    "response": {
      "attestationObject": "base64-encoded-attestation",
      "clientDataJSON": "base64-encoded-client-data"
    }
  },
  "deviceNickname": "My Laptop"
}
```

**Response:**
```json
{
  "success": true,
  "message": "Registration successful"
}
```

---

#### 3. Generate Authentication Options
Start the WebAuthn login flow.

**Endpoint:** `GET /api/auth/generate-authentication-options`

**Request:**
```
No body required
```

**Response:**
```json
{
  "publicKey": {
    "challenge": "base64-encoded-challenge",
    "timeout": 60000,
    "rpId": "localhost",
    "allowCredentials": [
      {"type": "public-key", "id": "base64-encoded-credential-id"}
    ],
    "userVerification": "preferred"
  }
}
```

---

#### 4. Verify Authentication
Complete the WebAuthn login.

**Endpoint:** `POST /api/auth/verify-authentication`

**Request Body:**
```json
{
  "username": "johndoe",
  "credential": {
    "id": "base64-encoded-credential-id",
    "rawId": "base64-encoded-credential-id",
    "type": "public-key",
    "response": {
      "authenticatorData": "base64-encoded-auth-data",
      "clientDataJSON": "base64-encoded-client-data",
      "signature": "base64-encoded-signature",
      "userHandle": "base64-encoded-user-handle"
    }
  }
}
```

**Response:**
```json
{
  "success": true,
  "message": "Login successful"
}
```

**Cookies Set:** `auth_token` (JWT)

---

#### 5. Verify Token
Check if current token is valid.

**Endpoint:** `POST /api/auth/verify-token`

**Request:**
```
No body required (uses auth_token cookie)
```

**Response:**
```json
{
  "valid": true,
  "username": "johndoe"
}
```

---

#### 6. Generate Conditional Options
For WebAuthn conditional mediation (autofill).

**Endpoint:** `GET /api/auth/generate-conditional-options`

**Request:**
```
No body required
```

**Response:**
Same as Generate Authentication Options.

---

### Authentication Endpoints (Protected)

All protected endpoints require the `auth_token` cookie to be sent.

#### 7. Get User Info

**Endpoint:** `GET /api/auth/user`

**Headers:**
```
Cookie: auth_token=<jwt-token>
```

**Response:**
```json
{
  "id": "user-uuid",
  "username": "johndoe",
  "displayName": "John Doe",
  "createdAt": "2024-01-15T10:00:00Z"
}
```

---

#### 8. Get Authenticators
List all registered passkeys/authenticators.

**Endpoint:** `GET /api/auth/authenticators`

**Response:**
```json
{
  "authenticators": [
    {
      "id": "authenticator-id",
      "nickname": "My Laptop",
      "deviceType": "platform",
      "backedUp": true,
      "transports": ["internal"],
      "createdAt": "2024-01-15T10:00:00Z"
    }
  ]
}
```

---

#### 9. Get Current User

**Endpoint:** `GET /api/auth/me`

**Response:**
```json
{
  "id": "user-uuid",
  "username": "johndoe",
  "displayName": "John Doe"
}
```

---

#### 10. Generate Additional Device Options
Add a new passkey to existing account.

**Endpoint:** `GET /api/auth/generate-additional-device-options`

**Response:**
Same as Generate Registration Options.

---

#### 11. Delete Authenticator
Remove a passkey.

**Endpoint:** `DELETE /api/auth/authenticator/:id`

**Response:**
```json
{
  "success": true,
  "message": "Authenticator deleted"
}
```

---

#### 12. Update Authenticator Nickname
Rename a passkey.

**Endpoint:** `PUT /api/auth/authenticator/:id/nickname`

**Request Body:**
```json
{
  "nickname": "New Nickname"
}
```

**Response:**
```json
{
  "success": true,
  "authenticator": {
    "id": "authenticator-id",
    "nickname": "New Nickname"
  }
}
```

---

#### 13. Logout
Clear authentication cookies.

**Endpoint:** `POST /api/auth/logout`

**Response:**
```json
{
  "success": true
}
```

**Cookies Cleared:** `auth_token`, `webauthn_session`

---

### Card Endpoints (Protected)

All card endpoints require the `auth_token` cookie.

#### 14. Get All Cards

**Endpoint:** `GET /api/auth/cards`

**Response:**
```json
{
  "cards": [
    {
      "id": "card-uuid",
      "user_id": "user-uuid",
      "pan": "4111111111111111",
      "cardholder_name": "John Doe",
      "bank_name": "Bank of America",
      "payment_method_type": "Credit",
      "card_brand": "Visa",
      "exp_month": 12,
      "exp_year": 2025,
      "cvv": 123,
      "created_at": "2024-01-15T10:00:00Z",
      "updated_at": "2024-01-15T10:00:00Z"
    }
  ]
}
```

---

#### 15. Create Card

**Endpoint:** `POST /api/auth/cards`

**Request Body:**
```json
{
  "pan": "4111111111111111",
  "cardholder_name": "John Doe",
  "bank_name": "Bank of America",
  "payment_method_type": "Credit",
  "card_brand": "Visa",
  "exp_month": 12,
  "exp_year": 2025,
  "cvv": 123
}
```

**Validation:**
- `pan`: Valid Visa, Mastercard, Amex, or Discover card number (15-16 digits)
- `payment_method_type`: "Credit" or "Debit"
- `card_brand`: "Visa", "Mastercard", "American Express", "Discover", "RuPay", or "Other"
- `exp_month`: 1-12
- `exp_year`: 2024 or later
- `cvv`: Exactly 3 digits

**Success Response:**
```json
{
  "card": {
    "id": "card-uuid",
    "user_id": "user-uuid",
    "pan": "4111111111111111",
    "cardholder_name": "John Doe",
    "bank_name": "Bank of America",
    "payment_method_type": "Credit",
    "card_brand": "Visa",
    "exp_month": 12,
    "exp_year": 2025,
    "cvv": 123,
    "created_at": "2024-01-15T10:00:00Z",
    "updated_at": "2024-01-15T10:00:00Z"
  }
}
```

**Error Response (Duplicate PAN):**
```json
{
  "error": "a card with this PAN already exists"
}
```
Status: `409 Conflict`

---

#### 16. Update Card

**Endpoint:** `PUT /api/auth/cards/:id`

**Request Body:**
```json
{
  "pan": "4111111111111111",
  "cardholder_name": "John Doe",
  "bank_name": "Bank of America",
  "payment_method_type": "Credit",
  "card_brand": "Visa",
  "exp_month": 12,
  "exp_year": 2026,
  "cvv": 456
}
```

**Success Response:**
```json
{
  "card": {
    "id": "card-uuid",
    "user_id": "user-uuid",
    "pan": "4111111111111111",
    "cardholder_name": "John Doe",
    "bank_name": "Bank of America",
    "payment_method_type": "Credit",
    "card_brand": "Visa",
    "exp_month": 12,
    "exp_year": 2026,
    "cvv": 456,
    "created_at": "2024-01-15T10:00:00Z",
    "updated_at": "2024-01-16T10:00:00Z"
  }
}
```

---

#### 17. Delete Card

**Endpoint:** `DELETE /api/auth/cards/:id`

**Response:**
```json
{
  "message": "card deleted successfully"
}
```

---

## Error Responses

| Status | Meaning |
|--------|---------|
| 400 | Bad Request - Invalid input |
| 401 | Unauthorized - Invalid/missing token |
| 403 | Forbidden - Not authorized |
| 404 | Not Found |
| 409 | Conflict - Duplicate PAN |
| 500 | Internal Server Error |

Error response format:
```json
{
  "error": "error message"
}
```

---

## Frontend Integration Tips

### WebAuthn Registration Flow
1. Call `GET /generate-registration-options`
2. Use WebAuthn API to create credential
3. Send credential to `POST /verify-registration`

### WebAuthn Login Flow
1. Call `GET /generate-authentication-options`
2. Use WebAuthn API to get assertion
3. Send assertion to `POST /verify-authentication`
4. Check for `auth_token` cookie in response

### Cookie Handling
- Cookies are set automatically by the server
- Include `withCredentials: true` in fetch/axios requests
- Server sets `SameSite=Strict` for security

### Card Validation
- PAN must pass Luhn algorithm check (server validates)
- CVV must be exactly 3 digits
- Cards with duplicate PAN are rejected with 409

---

## Running the Server

### Option 1: Binary
```bash
./myserver
```

### Option 2: Go Run
```bash
go run cmd/myserver/main.go
```

### Option 3: Build & Run
```bash
go build -o myserver ./cmd/myserver
./myserver
```

The server runs on port `8080` by default (configurable via `.env`).

---

## Testing with cURL

```bash
# Login (WebAuthn flow omitted - use browser)
# After login, check cards:
curl -X GET http://localhost:8080/api/auth/cards \
  -H "Cookie: auth_token=<your-token>"

# Create card
curl -X POST http://localhost:8080/api/auth/cards \
  -H "Content-Type: application/json" \
  -H "Cookie: auth_token=<your-token>" \
  -d '{
    "pan": "4111111111111111",
    "cardholder_name": "John Doe",
    "bank_name": "Bank of America",
    "payment_method_type": "Credit",
    "card_brand": "Visa",
    "exp_month": 12,
    "exp_year": 2025,
    "cvv": 123
  }'

# Logout
curl -X POST http://localhost:8080/api/auth/logout \
  -H "Cookie: auth_token=<your-token>"
```