import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../data/repositories/settings_repository_impl.dart';
import '../../domain/usecases/settings_usecase.dart';

final settingsUseCaseProvider = Provider<SettingsUseCase>((ref) {
  throw UnimplementedError('Settings dependency graph to be bound in composition root.');
});

final settingsRepositoryProvider = Provider<SettingsRepositoryImpl>((ref) {
  throw UnimplementedError('Settings repository binding must be provided by feature bootstrap.');
});
