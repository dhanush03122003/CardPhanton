import 'package:dio/dio.dart';

class AppLogger {
  AppLogger._();

  static void networkResponse(Response<dynamic> response) {
    // Replace with your preferred logging sink in production.
    // ignore: avoid_print
    print('[API] ${response.requestOptions.method} ${response.requestOptions.path} -> ${response.statusCode}');
  }

  static void networkError(DioException error) {
    // ignore: avoid_print
    print('[API][ERROR] ${error.requestOptions.method} ${error.requestOptions.path} -> ${error.message}');
  }
}
