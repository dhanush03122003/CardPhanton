import 'package:dio/dio.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_secure_storage/flutter_secure_storage.dart';
import 'package:shared_preferences/shared_preferences.dart';

import '../../app/env.dart';
import '../services/preferences_service.dart';
import '../services/secure_storage_service.dart';
import 'interceptors/auth_interceptor.dart';
import 'interceptors/error_interceptor.dart';
import 'interceptors/logging_interceptor.dart';
import 'interceptors/retry_interceptor.dart';
import 'token_refresh_service.dart';
import 'web_dio_adapter.dart';

final flutterSecureStorageProvider = Provider<FlutterSecureStorage>((ref) {
  return const FlutterSecureStorage();
});

final sharedPreferencesProvider = FutureProvider<SharedPreferences>((ref) async {
  return SharedPreferences.getInstance();
});

final secureStorageServiceProvider = Provider<SecureStorageService>((ref) {
  final storage = ref.watch(flutterSecureStorageProvider);
  return SecureStorageService(storage);
});

final preferencesServiceProvider = Provider<PreferencesService?>((ref) {
  final sharedPrefsAsync = ref.watch(sharedPreferencesProvider);
  return sharedPrefsAsync.maybeWhen(
    data: (sharedPrefs) => PreferencesService(sharedPrefs),
    orElse: () => null,
  );
});

final tokenRefreshServiceProvider = Provider<TokenRefreshService>((ref) {
  final secureStorageService = ref.watch(secureStorageServiceProvider);
  return TokenRefreshService(secureStorageService);
});

final dioProvider = Provider<Dio>((ref) {
  final secureStorageService = ref.watch(secureStorageServiceProvider);
  final tokenRefreshService = ref.watch(tokenRefreshServiceProvider);

  final dio = Dio(
    BaseOptions(
      baseUrl: AppEnv.apiBaseUrl,
      connectTimeout: const Duration(seconds: 15),
      receiveTimeout: const Duration(seconds: 15),
      sendTimeout: const Duration(seconds: 15),
      contentType: Headers.jsonContentType,
      responseType: ResponseType.json,
    ),
  );

  configureBrowserCredentials(dio);

  dio.interceptors.addAll(<Interceptor>[
    AuthInterceptor(
      secureStorageService: secureStorageService,
      tokenRefreshService: tokenRefreshService,
    ),
    RetryInterceptor(),
    ErrorInterceptor(),
    LoggingInterceptor(),
  ]);

  return dio;
});
