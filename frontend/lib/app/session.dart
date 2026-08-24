import 'package:flutter_riverpod/flutter_riverpod.dart';

class Session {
  final bool isAuthenticated;

  const Session({
    this.isAuthenticated = false,
  });
}

final sessionProvider = StateProvider<Session>(
  (ref) => const Session(),
);