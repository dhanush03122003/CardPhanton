import '../../domain/entities/gmail_entity.dart';
import '../../domain/repositories/gmail_repository.dart';
import '../datasources/gmail_remote_data_source.dart';

class GmailRepositoryImpl implements GmailRepository {
  GmailRepositoryImpl(this._remoteDataSource);

  final GmailRemoteDataSource _remoteDataSource;

  @override
  Future<GmailEntity> fetch() async {
    final dto = await _remoteDataSource.fetch();
    return GmailEntity(id: dto.id);
  }
}
