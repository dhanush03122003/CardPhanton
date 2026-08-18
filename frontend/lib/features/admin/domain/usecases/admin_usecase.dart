import '../entities/admin_entity.dart';
import '../repositories/admin_repository.dart';

class AdminUseCase {
  const AdminUseCase(this._repository);

  final AdminRepository _repository;

  Future<AdminEntity> call() {
    return _repository.fetch();
  }
}
