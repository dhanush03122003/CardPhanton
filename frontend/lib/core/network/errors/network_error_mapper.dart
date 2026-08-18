import 'package:dio/dio.dart';

import '../../errors/failures.dart';

class NetworkErrorMapper {
  NetworkErrorMapper._();

  static Failure map(DioException error) {
    final statusCode = error.response?.statusCode;

    if (statusCode == 401) {
      return const UnauthorizedFailure('Unauthorized request');
    }

    if (error.type == DioExceptionType.connectionTimeout ||
        error.type == DioExceptionType.receiveTimeout ||
        error.type == DioExceptionType.sendTimeout ||
        error.type == DioExceptionType.connectionError) {
      return const NetworkFailure('Network connection issue');
    }

    if (statusCode != null) {
      return ApiFailure('API request failed', statusCode: statusCode);
    }

    return const UnknownFailure('Unknown error');
  }
}
