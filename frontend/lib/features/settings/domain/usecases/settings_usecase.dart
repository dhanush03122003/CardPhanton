import '../entities/settings_entity.dart';
import '../repositories/settings_repository.dart';

class SettingsUseCase {
  const SettingsUseCase(this._repository);

  final SettingsRepository _repository;

  Future<SettingsEntity> call() {
    return _repository.fetch();
  }
}
