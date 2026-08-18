# CardPhanton Frontend

Flutter Web frontend for CardPhanton. This codebase is intentionally lean and production-focused: it uses a small folder structure now, and it can grow module by module when the app needs it.

## 1. Project Architecture

This project follows Feature First + Clean Architecture.

- Feature First keeps each business area isolated.
- Clean Architecture keeps domain logic independent from Flutter and from the data layer.
- Riverpod manages state and dependency injection.
- go_router handles auth redirects, nested routes, and role-based routing.
- Dio handles API calls, interceptors, auth headers, logging, and error mapping.

### Current structure

```txt
lib/
├── app/
│   ├── app.dart
│   ├── env.dart
│   ├── router.dart
│   ├── router_paths.dart
│   ├── session.dart
│   └── theme.dart
├── core/
│   ├── api/
│   ├── constants/
│   ├── errors/
│   ├── exceptions/
│   ├── extensions/
│   ├── logger/
│   ├── network/
│   ├── services/
│   ├── theme/
│   ├── utils/
│   ├── validators/
│   └── widgets/
├── features/
│   ├── auth/
│   ├── cards/
│   ├── admin/
│   ├── dashboard/
│   ├── profile/
│   ├── gmail/
│   ├── notifications/
│   └── settings/
└── main.dart
```

## 2. Folder Responsibilities

### app/

Application-level wiring only:

- app bootstrap
- router and route paths
- app environment config
- app-level session state
- app theme setup

### core/

Cross-feature foundation used everywhere:

- `api/`: shared API client wrapper and providers
- `network/`: Dio setup, interceptors, and token refresh placeholder
- `errors/`: failure types and `Result<T>`
- `exceptions/`: exception classes
- `constants/`: storage keys and other shared constants
- `extensions/`: reusable extensions
- `logger/`: logging helpers
- `services/`: secure storage and preferences services
- `theme/`: design tokens and theme extensions
- `utils/`: general helpers
- `validators/`: shared form validation rules
- `widgets/`: reusable design-system widgets

### features/

One folder per feature, each split into:

- `data/`: DTOs, datasources, repository implementations
- `domain/`: entities, repository contracts, use cases
- `presentation/`: pages, widgets, Riverpod providers

## 3. What Is In Place Today

- A Flutter Web app shell in `frontend/`
- Riverpod setup for state and DI
- go_router with auth and role-based redirects
- Dio with storage, logging, error, auth, and retry placeholders
- Token-based theming with light and dark support
- Shared validators, error types, and reusable widgets
- Feature skeletons for auth, cards, admin, dashboard, profile, gmail, notifications, and settings
- A small app-level session model in `lib/app/session.dart` instead of a separate shared layer

## 4. Rules To Follow

- Keep feature code inside its feature folder.
- Do not let DTOs leak into UI.
- Keep UI composition-based and small.
- Keep reusable cross-feature code in `core/`.
- Keep route constants centralized.
- Keep auth/session state at the app level, not spread across features.
- Keep secrets out of source control.
- Use secure storage for JWT and refresh tokens.
- Use shared preferences for theme and user preferences.
- Keep the app responsive and mobile-web first.
- Do not hardcode colors, spacing, or radius in widgets.
- Prefer immutable models and explicit types.

## 5. How To Extend When The App Grows

### When a feature becomes real

Add files only inside that feature:

- `data/models/`
- `data/datasources/`
- `data/repositories/`
- `domain/entities/`
- `domain/repositories/`
- `domain/usecases/`
- `presentation/pages/`
- `presentation/widgets/`
- `presentation/providers/`

### When shared code becomes necessary

Add to `core/` only if many features need it:

- networking helpers
- failure/result handling
- common widgets
- validators
- utilities
- services
- theme tokens

### When routing becomes larger

- Keep login, protected, and admin routes separated.
- Split route groups only when the router gets hard to read.
- Keep route paths centralized.

### When design system grows

- Put reusable widgets in `core/widgets`.
- Keep feature-specific widgets inside the feature.
- Introduce more theme tokens only when needed.

## 6. Commands

Run these from `frontend/`:

```bash
flutter pub get
dart run build_runner build --delete-conflicting-outputs
flutter analyze
flutter test
flutter run -d chrome --web-port 5500
flutter build web --release
```

If web support is missing on a new machine:

```bash
flutter config --enable-web
flutter create --platforms web .
```

## 7. Future Extension Map

If the app grows, extend it in this order:

1. Add more files inside an existing feature.
2. Add a new feature folder under `features/`.
3. Add shared helpers in `core/` only when multiple features need them.
4. Split router logic only when route count becomes hard to manage.
5. Introduce more design tokens only when the UI system needs it.

## 8. Short Production Rule

Keep the structure small now, keep the boundaries strict, and only expand folders when real app growth justifies it.
