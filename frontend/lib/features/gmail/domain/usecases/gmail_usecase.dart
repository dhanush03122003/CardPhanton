import '../entities/gmail_entity.dart';
import '../repositories/gmail_repository.dart';

class GmailUseCase {
  const GmailUseCase(this._repository);

  final GmailRepository _repository;

  Future<GmailEntity> call() {
    return _repository.fetch();
  }
}
