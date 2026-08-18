import '../models/gmail_dto.dart';

abstract interface class GmailRemoteDataSource {
  Future<GmailDto> fetch();
}
