import '../../domain/entities/auth_action_entity.dart';
import '../../domain/entities/auth_session_entity.dart';
import '../../domain/repositories/auth_repository.dart';
import '../datasources/auth_remote_data_source.dart';

class AuthRepositoryImpl implements AuthRepository {
  AuthRepositoryImpl(this._remoteDataSource);

  final AuthRemoteDataSource _remoteDataSource;

  @override
Future<Map<String, dynamic>> generateRegistrationOptions(
  String username,
) {
  return _remoteDataSource.generateRegistrationOptions(
    username.trim(),
  );
}

  @override
  Future<AuthActionEntity> verifyRegistration({
    required String username,
    required Map<String, dynamic> credential,
    String? deviceNickname,
  }) async {
    final dto = await _remoteDataSource.verifyRegistration(
      username: username,
      credential: credential,
      deviceNickname: deviceNickname,
    );
    return AuthActionEntity(
      success: dto.success,
      message: dto.message,
      username: dto.username ?? username,
    );
  }

  @override
  Future<Map<String, dynamic>> generateAuthenticationOptions(String username) {
    return _remoteDataSource.generateAuthenticationOptions(username);
  }

  @override
  Future<AuthActionEntity> verifyAuthentication({
    required String username,
    required Map<String, dynamic> credential,
  }) async {
    final dto = await _remoteDataSource.verifyAuthentication(
      username: username,
      credential: credential,
    );
    return AuthActionEntity(
      success: dto.success,
      message: dto.message,
      username: dto.username ?? username,
    );
  }

  @override
  Future<Map<String, dynamic>> generateConditionalOptions() {
    return _remoteDataSource.generateConditionalOptions();
  }

  @override
  Future<AuthSessionEntity> verifyToken() async {
    final dto = await _remoteDataSource.verifyToken();
    return AuthSessionEntity(valid: dto.valid, username: dto.username);
  }
}
