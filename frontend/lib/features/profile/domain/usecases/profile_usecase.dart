import '../entities/profile_entity.dart';
import '../repositories/profile_repository.dart';

class ProfileUseCase {
  const ProfileUseCase(this._repository);

  final ProfileRepository _repository;

  Future<ProfileEntity> call() {
    return _repository.fetch();
  }
}
