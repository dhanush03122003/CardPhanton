import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../data/repositories/gmail_repository_impl.dart';
import '../../domain/usecases/gmail_usecase.dart';

final gmailUseCaseProvider = Provider<GmailUseCase>((ref) {
  throw UnimplementedError('Gmail dependency graph to be bound in composition root.');
});

final gmailRepositoryProvider = Provider<GmailRepositoryImpl>((ref) {
  throw UnimplementedError('Gmail repository binding must be provided by feature bootstrap.');
});
