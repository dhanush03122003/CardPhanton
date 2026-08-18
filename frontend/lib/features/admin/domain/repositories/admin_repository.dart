import '../entities/admin_entity.dart';

abstract interface class AdminRepository {
  Future<AdminEntity> fetch();
}
