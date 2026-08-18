import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../data/repositories/auth_repository_impl.dart';
import '../../domain/usecases/auth_usecase.dart';

final authUseCaseProvider = Provider<AuthUseCase>((ref) {
  throw UnimplementedError('Auth dependency graph to be bound in composition root.');
});

final authRepositoryProvider = Provider<AuthRepositoryImpl>((ref) {
  throw UnimplementedError('Auth repository binding must be provided by feature bootstrap.');
});
