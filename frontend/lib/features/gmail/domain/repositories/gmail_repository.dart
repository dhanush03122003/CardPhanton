import '../entities/gmail_entity.dart';

abstract interface class GmailRepository {
  Future<GmailEntity> fetch();
}
