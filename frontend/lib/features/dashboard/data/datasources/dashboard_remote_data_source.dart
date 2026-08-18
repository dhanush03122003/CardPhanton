import '../models/dashboard_dto.dart';

abstract interface class DashboardRemoteDataSource {
  Future<DashboardDto> fetch();
}
