import '../models/settings_dto.dart';

abstract interface class SettingsRemoteDataSource {
  Future<SettingsDto> fetch();
}
