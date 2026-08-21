import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../features/admin/presentation/pages/admin_page.dart';
import '../features/auth/presentation/pages/auth_page.dart';
import '../features/cards/presentation/pages/cards_page.dart';
import '../features/dashboard/presentation/pages/dashboard_page.dart';
import '../features/gmail/presentation/pages/gmail_page.dart';
import '../features/notifications/presentation/pages/notifications_page.dart';
import '../features/profile/presentation/pages/profile_page.dart';
import '../features/settings/presentation/pages/settings_page.dart';
import 'session.dart';
import 'router_paths.dart';

final appRouterProvider = Provider<GoRouter>((ref) {
  final authSession = ref.watch(authSessionProvider);

  return GoRouter(
    initialLocation: RouterPaths.login,
    routes: <RouteBase>[
      GoRoute(
        path: RouterPaths.login,
        name: 'login',
        builder: (context, state) {
          return AuthPage(
            mode: AuthMode.login,
            initialUsername: state.uri.queryParameters['username'],
            registrationSuccess: state.uri.queryParameters['registered'] == '1',
          );
        },
      ),
      GoRoute(
        path: RouterPaths.register,
        name: 'register',
        builder: (context, state) {
          return const AuthPage(mode: AuthMode.register);
        },
      ),
      ShellRoute(
        builder: (context, state, child) => child,
        routes: <RouteBase>[
          GoRoute(
            path: RouterPaths.dashboard,
            name: 'dashboard',
            builder: (context, state) => const DashboardPage(),
            routes: <RouteBase>[
              GoRoute(
                path: 'cards',
                name: 'cards',
                builder: (context, state) => const CardsPage(),
              ),
              GoRoute(
                path: 'profile',
                name: 'profile',
                builder: (context, state) => const ProfilePage(),
              ),
              GoRoute(
                path: 'settings',
                name: 'settings',
                builder: (context, state) => const SettingsPage(),
              ),
            ],
          ),
          GoRoute(
            path: RouterPaths.admin,
            name: 'admin',
            builder: (context, state) => const AdminPage(),
          ),
          GoRoute(
            path: RouterPaths.notifications,
            name: 'notifications',
            builder: (context, state) => const NotificationsPage(),
          ),
          GoRoute(
            path: RouterPaths.gmail,
            name: 'gmail',
            builder: (context, state) => const GmailPage(),
          ),
        ],
      ),
    ],
    redirect: (context, state) {
      final isLoggingIn = state.matchedLocation == RouterPaths.login;
      final isRegistering = state.matchedLocation == RouterPaths.register;
      final isAuthed = authSession.isAuthenticated;
      final isAuthRoute = isLoggingIn || isRegistering;

      if (!isAuthed && !isAuthRoute) {
        return RouterPaths.login;
      }

      if (isAuthed && isAuthRoute) {
        return RouterPaths.dashboard;
      }

      final isAdminRoute = state.matchedLocation.startsWith(RouterPaths.admin);
      if (isAdminRoute && authSession.role != UserRole.admin) {
        return RouterPaths.dashboard;
      }

      return null;
    },
  );
});
