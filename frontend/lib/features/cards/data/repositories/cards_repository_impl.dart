import '../../domain/entities/cards_entity.dart';
import '../../domain/repositories/cards_repository.dart';
import '../datasources/cards_remote_data_source.dart';

class CardsRepositoryImpl implements CardsRepository {
  CardsRepositoryImpl(this._remoteDataSource);

  final CardsRemoteDataSource _remoteDataSource;

  @override
  Future<CardsEntity> fetch() async {
    final dto = await _remoteDataSource.fetch();
    return CardsEntity(id: dto.id);
  }
}
