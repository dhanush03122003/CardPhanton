import '../models/profile_dto.dart';

abstract interface class ProfileRemoteDataSource {
  Future<ProfileDto> fetch();
}
