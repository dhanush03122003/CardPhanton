import 'package:web_authn_web/web_authn_web.dart';

import '../../domain/entities/user.dart';
import '../../domain/repositories/auth_repository.dart';
import '../datasources/auth_remote_datasource.dart';
import '../mappers/webauthn_mapper.dart';
import '../models/user_model.dart';

class AuthRepositoryImpl implements AuthRepository {
  final AuthRemoteDataSource remoteDataSource;
  final WebAuthnWeb webAuthn;

  AuthRepositoryImpl({required this.remoteDataSource, required this.webAuthn});

  @override
  Future<User> login(String username) async {
    final optionsResponse = await remoteDataSource
        .generateAuthenticationOptions(username);
    final publicKey = Map<String, dynamic>.from(
      optionsResponse['publicKey'] as Map,
    );
    final options = WebAuthnMapper.requestOptionsFromJson(publicKey);

    final assertion = await webAuthn.sign(options);
    final verification = WebAuthnMapper.credentialToJson(assertion);

    final verifyResponse = await remoteDataSource.verifyAuthentication(
      username: username,
      verification: verification,
    );

    if (verifyResponse['success'] != true) {
      throw Exception('Authentication verification failed.');
    }
    return User(username: verifyResponse['username'] as String? ?? username);
  }

  @override
  Future<User> register(String username) async {
    final optionsResponse = await remoteDataSource.generateRegistrationOptions(
      username,
    );
    final publicKey = Map<String, dynamic>.from(
      optionsResponse['publicKey'] as Map,
    );
    final options = WebAuthnMapper.creationOptionsFromJson(publicKey);

    final credential = await webAuthn.register(options);
    final verification = WebAuthnMapper.credentialToJson(credential);

    final verifyResponse = await remoteDataSource.verifyRegistration(
      username: username,
      verification: verification,
    );

    if (verifyResponse['success'] != true) {
      throw Exception('Registration verification failed.');
    }
    return User(username: verifyResponse['username'] as String? ?? username);
  }

  @override
  Future<User> getCurrentUser() async {
    final response = await remoteDataSource.getCurrentUser();
    return UserModel.fromJson(Map<String, dynamic>.from(response));
  }
}
