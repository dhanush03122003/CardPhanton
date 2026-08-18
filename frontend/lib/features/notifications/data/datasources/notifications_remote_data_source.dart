import '../models/notifications_dto.dart';

abstract interface class NotificationsRemoteDataSource {
  Future<NotificationsDto> fetch();
}
