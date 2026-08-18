import '../entities/auth_entity.dart';
import '../repositories/auth_repository.dart';

class AuthUseCase {
  const AuthUseCase(this._repository);

  final AuthRepository _repository;

  Future<AuthEntity> call() {
    return _repository.fetch();
  }
}
