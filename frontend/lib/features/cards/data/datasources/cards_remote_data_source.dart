import '../models/cards_dto.dart';

abstract interface class CardsRemoteDataSource {
  Future<CardsDto> fetch();
}
