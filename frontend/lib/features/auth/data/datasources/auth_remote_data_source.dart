import '../../../../core/api/api_client.dart';
import '../models/auth_dto.dart';
import '../services/webauthn_browser_client.dart';

abstract interface class AuthRemoteDataSource {
  Future<Map<String, dynamic>> generateRegistrationOptions(String username);

  Future<AuthActionDto> verifyRegistration({
    required String username,
    required Map<String, dynamic> credential,
    String? deviceNickname,
  });

  Future<Map<String, dynamic>> generateAuthenticationOptions(String username);

  Future<AuthActionDto> verifyAuthentication({
    required String username,
    required Map<String, dynamic> credential,
  });

  Future<Map<String, dynamic>> generateConditionalOptions();

  Future<AuthSessionDto> verifyToken();
}

class AuthRemoteDataSourceImpl implements AuthRemoteDataSource {
  AuthRemoteDataSourceImpl(
    this._apiClient,
    this._browserClient,
  );

  final ApiClient _apiClient;
  final WebAuthnBrowserClient _browserClient;

  @override
Future<Map<String, dynamic>> generateRegistrationOptions(
  String username,
) async {
  final response = await _apiClient.get(
    '/auth/generate-registration-options',
    queryParameters: <String, dynamic>{
      'username': username.trim(),
    },
  );

  return _unwrapPublicKey(response.data);
}

  @override
  Future<AuthActionDto> verifyRegistration({
    required String username,
    required Map<String, dynamic> credential,
    String? deviceNickname,
  }) async {
    final response = await _apiClient.post(
  '/auth/verify-registration',
  data: <String, dynamic>{
    'username': username,
    'verification': credential,
    if (deviceNickname != null &&
        deviceNickname.trim().isNotEmpty)
      'deviceNickname': deviceNickname.trim(),
  },
);

    return AuthActionDto.fromJson(
      _asMap(response.data),
      fallbackUsername: username,
    );
  }

  @override
Future<Map<String, dynamic>> generateAuthenticationOptions(
  String username,
) async {
  final response = await _apiClient.get(
    '/auth/generate-authentication-options',
    queryParameters: <String, dynamic>{
      'username': username.trim(),
    },
  );

  return _unwrapPublicKey(response.data);
}

  @override
  Future<AuthActionDto> verifyAuthentication({
    required String username,
    required Map<String, dynamic> credential,
  }) async {
    final response = await _apiClient.post(
      '/auth/verify-authentication',
      data: <String, dynamic>{
        'username': username,
        'credential': credential,
      },
    );

    return AuthActionDto.fromJson(
      _asMap(response.data),
      fallbackUsername: username,
    );
  }

  @override
  Future<Map<String, dynamic>> generateConditionalOptions() async {
    final response = await _apiClient.get('/auth/generate-conditional-options');
    return _unwrapPublicKey(response.data);
  }

  @override
  Future<AuthSessionDto> verifyToken() async {
    final response = await _apiClient.post('/auth/verify-token');
    return AuthSessionDto.fromJson(_asMap(response.data));
  }

  Map<String, dynamic> _unwrapPublicKey(dynamic data) {
    final body = _asMap(data);
    final publicKey = body['publicKey'];
    if (publicKey is Map<String, dynamic>) {
      return publicKey;
    }
    if (publicKey is Map) {
      return Map<String, dynamic>.from(publicKey as Map);
    }
    return body;
  }

  Map<String, dynamic> _asMap(dynamic data) {
    if (data is Map<String, dynamic>) {
      return data;
    }
    if (data is Map) {
      return Map<String, dynamic>.from(data as Map);
    }
    throw StateError('Expected a JSON object but received ${data.runtimeType}.');
  }

  Future<Map<String, dynamic>> createCredential(Map<String, dynamic> publicKey) {
    return _browserClient.createCredential(publicKey);
  }

  Future<Map<String, dynamic>> getAssertion(
    Map<String, dynamic> publicKey, {
    bool conditional = false,
  }) {
    return _browserClient.getAssertion(
      publicKey,
      conditional: conditional,
    );
  }
}
