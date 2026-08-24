import 'package:flutter/foundation.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:web_authn_web/web_authn_web.dart';
import 'package:dio/dio.dart';

import '../../../../core/network/dio_client.dart';
import '../../data/datasources/auth_remote_datasource.dart';
import '../../data/repositories/auth_repository_impl.dart';
import '../../domain/entities/user.dart';
import '../../domain/repositories/auth_repository.dart';
import '../../domain/usecases/login.dart';
import '../../domain/usecases/register.dart';
import '../../domain/usecases/get_current_user.dart';

final dioClientProvider = Provider<DioClient>((ref) => DioClient());

final webAuthnProvider = Provider<WebAuthnWeb>((ref) => WebAuthnWeb());

final authRemoteDataSourceProvider = Provider<AuthRemoteDataSource>((ref) {
  return AuthRemoteDataSource(ref.watch(dioClientProvider).dio);
});

final authRepositoryProvider = Provider<AuthRepository>((ref) {
  return AuthRepositoryImpl(
    remoteDataSource: ref.watch(authRemoteDataSourceProvider),
    webAuthn: ref.watch(webAuthnProvider),
  );
});

final loginUseCaseProvider = Provider<Login>(
  (ref) => Login(ref.watch(authRepositoryProvider)),
);

final registerUseCaseProvider = Provider<Register>(
  (ref) => Register(ref.watch(authRepositoryProvider)),
);

final getCurrentUserUseCaseProvider = Provider<GetCurrentUser>(
  (ref) => GetCurrentUser(ref.watch(authRepositoryProvider)),
);

class AuthState {
  final bool loading;
  final User? user;
  final String? error;

  /// True until the startup session check (`restoreSession`) resolves.
  /// Defaults to true so the router doesn't treat "haven't checked yet"
  /// as "logged out" during the brief window before that check completes.
  final bool initializing;

  const AuthState({
    this.loading = false,
    this.user,
    this.error,
    this.initializing = true,
  });

  AuthState copyWith({
    bool? loading,
    User? user,
    String? error,
    bool? initializing,
  }) {
    return AuthState(
      loading: loading ?? this.loading,
      user: user ?? this.user,
      error: error,
      initializing: initializing ?? this.initializing,
    );
  }
}

class AuthController extends Notifier<AuthState> {
  @override
  AuthState build() => const AuthState();

  void clearError() {
    state = AuthState(user: state.user, initializing: false);
  }

  /// Call once at app startup to check for an existing session cookie.
  Future<void> restoreSession() async {
    try {
      final user = await ref.read(getCurrentUserUseCaseProvider).call();
      state = AuthState(user: user, initializing: false);
    } catch (_) {
      state = const AuthState(initializing: false); // no valid session
    }
  }

  Future<void> login(String username) async {
    state = AuthState(loading: true, initializing: false);
    try {
      final user = await ref.read(loginUseCaseProvider).call(username);
      state = AuthState(user: user, initializing: false);
    } catch (e, st) {
      debugPrint('Auth error: $e\n$st');
      state = AuthState(error: _message(e), initializing: false);
    }
  }

  Future<void> register(String username) async {
    state = AuthState(loading: true, initializing: false);
    try {
      final user = await ref.read(registerUseCaseProvider).call(username);
      state = AuthState(user: user, initializing: false);
    } catch (e, st) {
      debugPrint('Auth error: $e\n$st');
      state = AuthState(error: _message(e), initializing: false);
    }
  }

  Future<void> logout() async {
    // If your backend exposes a logout route that clears the session
    // cookie, call it here via the repository before resetting state.
    state = const AuthState(initializing: false);
  }

  String _message(Object error) {
    if (error is DioException) {
      final data = error.response?.data;

      if (data is Map && data['error'] is String) {
        return data['error'] as String;
      }

      if (error.response?.statusCode == 401) {
        return 'Your session has expired. Please try again.';
      }
      if (error.response?.statusCode == 404) {
        return 'User not found.';
      }
      if (error.response?.statusCode == 400) {
        return 'Invalid request. Please try again.';
      }

      return 'Something went wrong. Please try again.';
    }

    if (error is WebAuthnWebException) {
      final cause = error.cause?.toString() ?? '';

      if (cause.contains('InvalidStateError')) {
        return 'This authenticator is already registered. Please try logging in.';
      }
      if (cause.contains('NotAllowedError')) {
        return 'The operation was cancelled or not allowed.';
      }
      if (cause.contains('SecurityError')) {
        return 'WebAuthn is not available for this website.';
      }

      return error.message;
    }

    return 'Something went wrong. Please try again.';
  }
}

final authControllerProvider = NotifierProvider<AuthController, AuthState>(
  AuthController.new,
);
