import 'package:dio/dio.dart';
import 'package:dio/browser.dart';

class ApiClient {
  ApiClient(this._dio) {
    _dio.options = BaseOptions(
      baseUrl: 'http://localhost:8080/api',
      connectTimeout: const Duration(seconds: 15),
      receiveTimeout: const Duration(seconds: 15),
      sendTimeout: const Duration(seconds: 15),
      headers: <String, dynamic>{
        'Content-Type': 'application/json',
        'Accept': 'application/json',
      },
    );

    _dio.httpClientAdapter = BrowserHttpClientAdapter(
      withCredentials: false,
    );
  }

  final Dio _dio;

  Future<Response<dynamic>> get(
    String path, {
    Map<String, dynamic>? queryParameters,
  }) {
    return _dio.get(
      path,
      queryParameters: queryParameters,
    );
  }

  Future<Response<dynamic>> post(
    String path, {
    dynamic data,
  }) {
    return _dio.post(
      path,
      data: data,
    );
  }

  Future<Response<dynamic>> put(
    String path, {
    dynamic data,
  }) {
    return _dio.put(
      path,
      data: data,
    );
  }

  Future<Response<dynamic>> delete(
    String path, {
    dynamic data,
  }) {
    return _dio.delete(
      path,
      data: data,
    );
  }
}