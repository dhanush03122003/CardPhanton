import 'package:dio/dio.dart';

class RetryInterceptor extends Interceptor {
  RetryInterceptor({this.maxRetryCount = 1});

  final int maxRetryCount;

  @override
  Future<void> onError(
    DioException err,
    ErrorInterceptorHandler handler,
  ) async {
    // Placeholder: wire a policy-based retry strategy when request idempotency rules are finalized.
    handler.next(err);
  }
}
