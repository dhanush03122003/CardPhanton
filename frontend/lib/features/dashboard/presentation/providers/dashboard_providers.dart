import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../data/repositories/dashboard_repository_impl.dart';
import '../../domain/usecases/dashboard_usecase.dart';

final dashboardUseCaseProvider = Provider<DashboardUseCase>((ref) {
  throw UnimplementedError('Dashboard dependency graph to be bound in composition root.');
});

final dashboardRepositoryProvider = Provider<DashboardRepositoryImpl>((ref) {
  throw UnimplementedError('Dashboard repository binding must be provided by feature bootstrap.');
});
