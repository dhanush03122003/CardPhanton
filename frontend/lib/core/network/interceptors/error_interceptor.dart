import 'package:dio/dio.dart';

import '../../errors/failures.dart';
import '../../exceptions/app_exceptions.dart';

class ErrorInterceptor extends Interceptor {
  @override
  void onError(DioException err, ErrorInterceptorHandler handler) {
    final statusCode = err.response?.statusCode;

    if (statusCode == 401) {
      handler.reject(
        DioException(
          requestOptions: err.requestOptions,
          response: err.response,
          type: err.type,
          error: UnauthorizedException(const UnauthorizedFailure('Unauthorized').message),
          message: err.message,
        ),
      );
      return;
    }

    handler.next(err);
  }
}
