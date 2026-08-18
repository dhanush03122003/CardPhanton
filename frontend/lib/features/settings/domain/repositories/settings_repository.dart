import '../entities/settings_entity.dart';

abstract interface class SettingsRepository {
  Future<SettingsEntity> fetch();
}
