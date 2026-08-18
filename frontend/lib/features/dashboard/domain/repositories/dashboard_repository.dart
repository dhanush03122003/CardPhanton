import '../entities/dashboard_entity.dart';

abstract interface class DashboardRepository {
  Future<DashboardEntity> fetch();
}
