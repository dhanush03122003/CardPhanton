import 'package:dio/dio.dart';

class AuthRemoteDataSource {
  final Dio dio;

  AuthRemoteDataSource(this.dio);

  Future<Map<String, dynamic>> generateAuthenticationOptions(
    String username,
  ) async {
    final response = await dio.get(
      '/api/auth/generate-authentication-options',
      queryParameters: {'username': username},
    );

    return Map<String, dynamic>.from(response.data as Map);
  }

  Future<Map<String, dynamic>> generateConditionalOptions() async {
    final response = await dio.get('/api/auth/generate-conditional-options');

    return Map<String, dynamic>.from(response.data as Map);
  }

  Future<Map<String, dynamic>> verifyAuthentication({
    String? username,
    required Map<String, dynamic> verification,
  }) async {
    final response = await dio.post(
      '/api/auth/verify-authentication',
      data: {
        if (username != null) 'username': username,
        'verification': verification,
      },
    );

    return Map<String, dynamic>.from(response.data as Map);
  }

  Future<Map<String, dynamic>> generateRegistrationOptions(
    String username,
  ) async {
    final response = await dio.get(
      '/api/auth/generate-registration-options',
      queryParameters: {'username': username},
    );

    return Map<String, dynamic>.from(response.data as Map);
  }

  Future<Map<String, dynamic>> verifyRegistration({
    required String username,
    required Map<String, dynamic> verification,
  }) async {
    final response = await dio.post(
      '/api/auth/verify-registration',
      data: {'username': username, 'verification': verification},
    );

    return Map<String, dynamic>.from(response.data as Map);
  }

  /// Reads the current session via the httpOnly session cookie and returns
  /// the logged-in user; throws (e.g. 401) if there's no valid session.
  /// NOTE: adjust the path to match your actual backend route.
  Future<Map<String, dynamic>> getCurrentUser() async {
    final response = await dio.get('/api/auth/me');

    return Map<String, dynamic>.from(response.data as Map);
  }
}
