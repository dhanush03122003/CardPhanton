import 'package:flutter/foundation.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../features/auth/presentation/pages/login_page.dart';
import '../features/auth/presentation/pages/register_page.dart';
import '../features/auth/presentation/providers/auth_providers.dart';
import '../features/home/presentation/pages/home_page.dart';
import 'router_paths.dart';

/// Bridges Riverpod's [authControllerProvider] into a [Listenable] so
/// go_router re-evaluates its `redirect` callback whenever auth state
/// changes (login, logout, session restore) — not just on navigation.
class _AuthRefreshListenable extends ChangeNotifier {
  _AuthRefreshListenable(Ref ref) {
    ref.listen(authControllerProvider, (_, __) => notifyListeners());
  }
}

/// The app's router. Read via `ref.watch(routerProvider)` from the widget
/// that builds `MaterialApp.router` (see `app.dart`).
final routerProvider = Provider<GoRouter>((ref) {
  final refreshListenable = _AuthRefreshListenable(ref);
  ref.onDispose(refreshListenable.dispose);

  return GoRouter(
    initialLocation: RouterPaths.login,
    refreshListenable: refreshListenable,
    redirect: (context, state) {
      final auth = ref.read(authControllerProvider);

      // Still checking for an existing session cookie on startup — don't
      // redirect yet, or an authenticated user deep-linking to /homepage
      // would get bounced to /login before the check even resolves.
      if (auth.initializing) return null;

      final loggedIn = auth.user != null;
      final onAuthPage =
          state.matchedLocation == RouterPaths.login ||
          state.matchedLocation == RouterPaths.register;

      if (!loggedIn && !onAuthPage) return RouterPaths.login;
      if (loggedIn && onAuthPage) return RouterPaths.homepage;

      return null;
    },
    routes: [
      GoRoute(
        path: RouterPaths.login,
        builder: (context, state) => const LoginPage(),
      ),
      GoRoute(
        path: RouterPaths.register,
        builder: (context, state) => const RegisterPage(),
      ),
      GoRoute(
        path: RouterPaths.homepage,
        builder: (context, state) => const HomePage(),
      ),
    ],
  );
});
