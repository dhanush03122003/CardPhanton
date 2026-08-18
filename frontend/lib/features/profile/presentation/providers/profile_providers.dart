import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../data/repositories/profile_repository_impl.dart';
import '../../domain/usecases/profile_usecase.dart';

final profileUseCaseProvider = Provider<ProfileUseCase>((ref) {
  throw UnimplementedError('Profile dependency graph to be bound in composition root.');
});

final profileRepositoryProvider = Provider<ProfileRepositoryImpl>((ref) {
  throw UnimplementedError('Profile repository binding must be provided by feature bootstrap.');
});
