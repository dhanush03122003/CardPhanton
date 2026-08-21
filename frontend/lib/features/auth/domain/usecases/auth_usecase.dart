import '../entities/auth_action_entity.dart';
import '../entities/auth_session_entity.dart';
import '../repositories/auth_repository.dart';

class AuthUseCase {
  const AuthUseCase(this._repository);

  final AuthRepository _repository;

  Future<Map<String, dynamic>> generateRegistrationOptions(String username) {
    return _repository.generateRegistrationOptions(username);
  }

  Future<AuthActionEntity> verifyRegistration({
    required String username,
    required Map<String, dynamic> credential,
    String? deviceNickname,
  }) {
    return _repository.verifyRegistration(
      username: username,
      credential: credential,
      deviceNickname: deviceNickname,
    );
  }

  Future<Map<String, dynamic>> generateAuthenticationOptions(String username) {
    return _repository.generateAuthenticationOptions(username);
  }

  Future<AuthActionEntity> verifyAuthentication({
    required String username,
    required Map<String, dynamic> credential,
  }) {
    return _repository.verifyAuthentication(
      username: username,
      credential: credential,
    );
  }

  Future<Map<String, dynamic>> generateConditionalOptions() {
    return _repository.generateConditionalOptions();
  }

  Future<AuthSessionEntity> verifyToken() {
    return _repository.verifyToken();
  }
}
