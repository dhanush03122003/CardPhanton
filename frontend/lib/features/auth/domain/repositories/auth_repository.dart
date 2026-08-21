import '../entities/auth_action_entity.dart';
import '../entities/auth_session_entity.dart';

abstract interface class AuthRepository {
  Future<Map<String, dynamic>> generateRegistrationOptions(String username);

  Future<AuthActionEntity> verifyRegistration({
    required String username,
    required Map<String, dynamic> credential,
    String? deviceNickname,
  });

  Future<Map<String, dynamic>> generateAuthenticationOptions(String username);

  Future<AuthActionEntity> verifyAuthentication({
    required String username,
    required Map<String, dynamic> credential,
  });

  Future<Map<String, dynamic>> generateConditionalOptions();

  Future<AuthSessionEntity> verifyToken();
}
