import '../entities/notifications_entity.dart';

abstract interface class NotificationsRepository {
  Future<NotificationsEntity> fetch();
}
