import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../data/repositories/admin_repository_impl.dart';
import '../../domain/usecases/admin_usecase.dart';

final adminUseCaseProvider = Provider<AdminUseCase>((ref) {
  throw UnimplementedError('Admin dependency graph to be bound in composition root.');
});

final adminRepositoryProvider = Provider<AdminRepositoryImpl>((ref) {
  throw UnimplementedError('Admin repository binding must be provided by feature bootstrap.');
});
