import 'package:shared_preferences/shared_preferences.dart';

import '../constants/storage_keys.dart';

class PreferencesService {
  PreferencesService(this._preferences);

  final SharedPreferences _preferences;

  Future<void> setThemeMode(String mode) async {
    await _preferences.setString(StorageKeys.themeMode, mode);
  }

  String? getThemeMode() {
    return _preferences.getString(StorageKeys.themeMode);
  }

  Future<void> setUserPreferencesJson(String value) async {
    await _preferences.setString(StorageKeys.userPreferences, value);
  }

  String? getUserPreferencesJson() {
    return _preferences.getString(StorageKeys.userPreferences);
  }
}
