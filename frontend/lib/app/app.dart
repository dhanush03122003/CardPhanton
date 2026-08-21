import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../features/auth/presentation/providers/auth_providers.dart';
import 'session.dart';
import 'router.dart';
import 'theme.dart';

class CardPhantonApp extends ConsumerStatefulWidget {
  const CardPhantonApp({super.key});

  @override
  ConsumerState<CardPhantonApp> createState() => _CardPhantonAppState();
}

class _CardPhantonAppState extends ConsumerState<CardPhantonApp> {
  bool _sessionChecked = false;

  @override
  void didChangeDependencies() {
    super.didChangeDependencies();
    if (_sessionChecked) {
      return;
    }
    _sessionChecked = true;
    WidgetsBinding.instance.addPostFrameCallback((_) {
      if (mounted && kIsWeb) {
        _restoreSession();
      }
    });
  }

  Future<void> _restoreSession() async {
    final authUseCase = ref.read(authUseCaseProvider);
    final authSession = ref.read(authSessionProvider);

    try {
      final session = await authUseCase.verifyToken();
      if (session.valid) {
        authSession.updateAuth(
          isAuthenticated: true,
          role: UserRole.user,
          username: session.username,
        );
      }
    } catch (_) {
      authSession.clear();
    }
  }

  @override
  Widget build(BuildContext context) {
    final router = ref.watch(appRouterProvider);

    return MaterialApp.router(
      title: 'CardPhanton',
      debugShowCheckedModeBanner: false,
      theme: AppTheme.light,
      darkTheme: AppTheme.dark,
      themeMode: ThemeMode.system,
      routerConfig: router,
    );
  }
}
