import 'package:dio/dio.dart';

import '../../services/secure_storage_service.dart';
import '../token_refresh_service.dart';

class AuthInterceptor extends Interceptor {
  AuthInterceptor({
    required SecureStorageService secureStorageService,
    required TokenRefreshService tokenRefreshService,
  })  : _secureStorageService = secureStorageService,
        _tokenRefreshService = tokenRefreshService;

  final SecureStorageService _secureStorageService;
  final TokenRefreshService _tokenRefreshService;

  @override
  Future<void> onRequest(
    RequestOptions options,
    RequestInterceptorHandler handler,
  ) async {
    final token = await _secureStorageService.getAccessToken();
    if (token != null && token.isNotEmpty) {
      options.headers['Authorization'] = 'Bearer $token';
    }

    handler.next(options);
  }

  @override
  Future<void> onError(
    DioException err,
    ErrorInterceptorHandler handler,
  ) async {
    final isUnauthorized = err.response?.statusCode == 401;
    if (!isUnauthorized) {
      handler.next(err);
      return;
    }

    final newToken = await _tokenRefreshService.refreshAccessToken();
    if (newToken == null || newToken.isEmpty) {
      handler.next(err);
      return;
    }

    final requestOptions = err.requestOptions;
    requestOptions.headers['Authorization'] = 'Bearer $newToken';

    try {
      final response = await Dio().fetch<dynamic>(requestOptions);
      handler.resolve(response);
    } on DioException catch (retryError) {
      handler.next(retryError);
    }
  }
}
