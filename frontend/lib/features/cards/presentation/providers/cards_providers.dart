import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../data/repositories/cards_repository_impl.dart';
import '../../domain/usecases/cards_usecase.dart';

final cardsUseCaseProvider = Provider<CardsUseCase>((ref) {
  throw UnimplementedError('Cards dependency graph to be bound in composition root.');
});

final cardsRepositoryProvider = Provider<CardsRepositoryImpl>((ref) {
  throw UnimplementedError('Cards repository binding must be provided by feature bootstrap.');
});
