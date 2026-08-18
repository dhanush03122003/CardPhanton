import '../models/admin_dto.dart';

abstract interface class AdminRemoteDataSource {
  Future<AdminDto> fetch();
}
