import '../entities/cards_entity.dart';
import '../repositories/cards_repository.dart';

class CardsUseCase {
  const CardsUseCase(this._repository);

  final CardsRepository _repository;

  Future<CardsEntity> call() {
    return _repository.fetch();
  }
}
