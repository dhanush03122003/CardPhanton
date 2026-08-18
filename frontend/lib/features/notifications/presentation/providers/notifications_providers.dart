import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../data/repositories/notifications_repository_impl.dart';
import '../../domain/usecases/notifications_usecase.dart';

final notificationsUseCaseProvider = Provider<NotificationsUseCase>((ref) {
  throw UnimplementedError('Notifications dependency graph to be bound in composition root.');
});

final notificationsRepositoryProvider = Provider<NotificationsRepositoryImpl>((ref) {
  throw UnimplementedError('Notifications repository binding must be provided by feature bootstrap.');
});
