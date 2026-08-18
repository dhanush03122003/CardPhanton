import '../entities/notifications_entity.dart';
import '../repositories/notifications_repository.dart';

class NotificationsUseCase {
  const NotificationsUseCase(this._repository);

  final NotificationsRepository _repository;

  Future<NotificationsEntity> call() {
    return _repository.fetch();
  }
}
