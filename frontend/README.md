# CardPhanton — Frontend

Flutter Web client for CardPhanton, using passwordless authentication via
WebAuthn (passkeys) against a session-cookie-backed backend.

## Tech stack

| Concern            | Package                          |
|---------------------|-----------------------------------|
| State management    | `flutter_riverpod`                |
| Routing             | `go_router`                       |
| HTTP client         | `dio`                             |
| WebAuthn (passkeys) | `web_authn_web`                   |
| Local persistence   | `shared_preferences`              |

**Platform:** Flutter Web only (the auth flow depends on the browser's
WebAuthn API via `web_authn_web`).

---

## Getting started

### Prerequisites

- Flutter SDK (stable channel) — check the version pinned in `pubspec.yaml`
- A running instance of the CardPhanton backend (WebAuthn relying party +
  session store), reachable at the URL configured in `lib/app/env.dart`
- A Chromium-based browser for local development (`flutter run -d chrome`)
- WebAuthn requires a **secure context**: `localhost` works for local dev;
  anywhere else needs HTTPS

### Install dependencies

```bash
flutter pub get
```

### Configure environment

API base URL and other environment values live in `lib/app/env.dart`.
Confirm `apiBaseUrl` points at your backend (defaults to
`http://localhost:8080` for local development) and that the backend's CORS
config allows credentials from your frontend's origin — the app sends
`withCredentials: true` on every request since auth relies on an httpOnly
session cookie, not a bearer token.

### Run locally

```bash
flutter run -d chrome
```

### Build for production

```bash
flutter build web
```

Output is written to `build/web/`; serve it from any static host / reverse
proxy that sits on the same effective domain as your backend's configured
Relying Party ID (WebAuthn ties credentials to `rp.id`, so frontend and
backend must agree on the domain).

---

## Project structure

This project follows a **feature-first, layered (clean) architecture**.
Each feature under `lib/features/` is internally split into three layers,
and cross-cutting concerns live under `lib/core/`.

```
lib/
├── app/                  # App shell: routing, theming, env config, session bootstrap
│   ├── app.dart
│   ├── env.dart
│   ├── router.dart
│   ├── router_paths.dart
│   ├── session.dart
│   └── theme.dart
│
├── core/                 # Shared, feature-agnostic building blocks
│   ├── constants/        # App-wide constant values
│   ├── errors/           # Centralized error messages + exception → message mappers
│   ├── logger/           # Logging abstraction (use instead of raw print())
│   ├── network/          # Dio client setup (base URL, timeouts, credentials)
│   ├── services/         # Cross-feature services (e.g. platform integrations)
│   ├── utils/             # Small stateless helper functions
│   ├── validators/       # Reusable form-field validators
│   └── widgets/          # Shared, generic UI components (not feature-specific)
│
├── features/
│   └── <feature_name>/
│       ├── data/
│       │   ├── datasources/    # Raw API calls (Dio), no business logic
│       │   ├── mappers/        # JSON ⇄ typed-object conversion, isolated from repositories
│       │   ├── models/         # DTOs — extend domain entities, add fromJson/toJson
│       │   └── repositories/   # Implements the domain repository interface;
│       │                       # orchestrates datasources + mappers
│       ├── domain/
│       │   ├── entities/       # Plain business objects, no serialization logic
│       │   ├── repositories/   # Abstract interfaces — data layer depends on these,
│       │   │                   # not the other way around
│       │   └── usecases/       # One class per user action (Login, Register, ...);
│       │                       # thin wrappers around a repository call
│       └── presentation/
│           ├── pages/          # Screens — composition only, minimal logic
│           ├── providers/      # Riverpod providers + state notifiers for this feature
│           └── widgets/        # Widgets shared *within* this feature only
│
└── main.dart
```

### Why this structure

- **`domain/` has no Flutter or Dio imports.** It's pure Dart — entities and
  interfaces — so business rules are testable without mocking the framework.
- **`data/` depends on `domain/`, never the reverse.** Repositories implement
  the interfaces domain defines; swapping WebAuthn/Dio for something else
  later only touches this layer.
- **`presentation/` talks to `domain/` usecases, not `data/` directly.**
  Pages and providers never import a datasource or repository implementation.
- **Mappers are separated from repositories.** Anything that converts
  between raw JSON/SDK objects and typed Dart classes lives in
  `data/mappers/`, keeping repository classes focused on *orchestrating* a
  flow (call datasource → map → call SDK → map → return) rather than doing
  the marshalling inline.
- **Error messages are centralized**, not scattered as string literals
  across controllers. `core/errors/error_messages.dart` holds the copy;
  `core/errors/auth_error_mapper.dart` (and equivalents per feature, if
  needed) turns exceptions into one of those messages, and always logs the
  raw error first so failures are never silently swallowed.

---

## Authentication flow

Auth is passwordless, backed by WebAuthn (passkeys) and an httpOnly session
cookie — **there is no bearer token stored client-side.**

1. **Register/Login options** — the client asks the backend for a WebAuthn
   challenge (`generate-registration-options` / `generate-authentication-options`).
   The backend also sets a short-lived session cookie carrying that challenge.
2. **Browser ceremony** — `web_authn_web` calls the browser's native
   WebAuthn API (fingerprint / face / security key prompt).
3. **Verification** — the resulting credential is posted back
   (`verify-registration` / `verify-authentication`). The backend validates
   it against the session cookie's challenge and, on success, upgrades the
   session cookie to an authenticated one.
4. **Session state** — the frontend never sees a token. On success it just
   stores the returned `User` in memory (`AuthController` / Riverpod state).
   On app startup, `AuthController.restoreSession()` calls a "who am I"
   endpoint (`GET /api/auth/me`) that reads the session cookie server-side;
   this is what allows a page refresh to stay logged in.
5. **Logout** clears local state; if the backend exposes a logout route to
   invalidate the cookie server-side, that should be called too.

### Route protection

`/homepage` (and any other authenticated route) is guarded at the **router**
level via `GoRouter.redirect`, not inside the page itself — this ensures
direct URL access, refreshes, and back-navigation are all covered by the
same check, not just in-app navigation. See `lib/app/router.dart`.

---

## Code conventions

- No raw `print()` in shipped code — use the logger under `core/logger/`.
- No user-facing string literals inline in controllers/widgets — add them to
  the relevant `core/errors/error_messages.dart` (or feature-local
  equivalent) so copy changes don't require touching logic.
- Keep `data/repositories/*_impl.dart` files to orchestration only; if a
  method is doing JSON parsing or SDK-object construction inline, it belongs
  in a `data/mappers/` file instead.
- Shared widgets used by more than one page *within a feature* go in that
  feature's `presentation/widgets/`; widgets shared *across features* go in
  `core/widgets/`.
- Form validation logic goes in `core/validators/`, not inline in a page's
  `onPressed` handler.

---

## Troubleshooting

- **WebAuthn errors ("not allowed", "security error")**: confirm you're on
  `localhost` or HTTPS, and that `rp.id` returned by the backend matches the
  frontend's actual domain.
- **CORS / cookie not being sent**: confirm the backend's CORS config
  includes `Access-Control-Allow-Credentials: true` and an explicit (not
  wildcard) `Access-Control-Allow-Origin` matching the frontend's origin.
- **Stuck on login page after a successful passkey prompt**: check the
  browser console — auth failures are logged via `AuthErrorMapper.map()`
  before being converted to a user-facing message, so the raw exception is
  always visible there.