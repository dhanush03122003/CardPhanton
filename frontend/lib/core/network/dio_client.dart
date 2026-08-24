import 'package:dio/dio.dart';
import 'package:dio/browser.dart';
import 'package:flutter/foundation.dart';
import 'package:dio_web_adapter/dio_web_adapter.dart';

import '../../app/env.dart';

class DioClient {
  DioClient() {
    _dio = Dio(
      BaseOptions(
        baseUrl: Env.apiBaseUrl,
        connectTimeout: const Duration(seconds: 15),
        receiveTimeout: const Duration(seconds: 15),
        headers: {'Content-Type': 'application/json'},
      ),
    );

    dio.httpClientAdapter = BrowserHttpClientAdapter(withCredentials: true);
  }

  late final Dio _dio;

  Dio get dio => _dio;
}
