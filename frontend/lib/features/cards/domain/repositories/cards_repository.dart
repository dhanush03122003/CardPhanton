import '../entities/cards_entity.dart';

abstract interface class CardsRepository {
  Future<CardsEntity> fetch();
}
