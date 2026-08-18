import '../entities/dashboard_entity.dart';
import '../repositories/dashboard_repository.dart';

class DashboardUseCase {
  const DashboardUseCase(this._repository);

  final DashboardRepository _repository;

  Future<DashboardEntity> call() {
    return _repository.fetch();
  }
}
