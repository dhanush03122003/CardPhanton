import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../features/auth/presentation/providers/auth_providers.dart';
import 'router.dart';
import 'theme.dart';

class CardPhantonApp extends ConsumerStatefulWidget {
  const CardPhantonApp({super.key});

  @override
  ConsumerState<CardPhantonApp> createState() => _CardPhantonAppState();
}

class _CardPhantonAppState extends ConsumerState<CardPhantonApp> {
  @override
  void initState() {
    super.initState();
    // Check for an existing session cookie once, at startup. Router
    // redirects wait on `auth.initializing` until this resolves.
    Future.microtask(
      () => ref.read(authControllerProvider.notifier).restoreSession(),
    );
  }

  @override
  Widget build(BuildContext context) {
    final router = ref.watch(routerProvider);

    return MaterialApp.router(
      title: 'CardPhanton',
      debugShowCheckedModeBanner: false,
      routerConfig: router,
      theme: lightTheme,
      darkTheme: darkTheme,
      themeMode: ThemeMode.system,
    );
  }
}
