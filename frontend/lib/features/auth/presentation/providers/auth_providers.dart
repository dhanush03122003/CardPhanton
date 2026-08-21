import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../../core/api/api_providers.dart';
import '../../data/datasources/auth_remote_data_source.dart';
import '../../data/repositories/auth_repository_impl.dart';
import '../../domain/usecases/auth_usecase.dart';
import '../../data/services/webauthn_browser_client.dart';

final authRemoteDataSourceProvider = Provider<AuthRemoteDataSource>((ref) {
  final apiClient = ref.watch(apiClientProvider);
  return AuthRemoteDataSourceImpl(apiClient, const WebAuthnBrowserClient());
});

final webAuthnBrowserClientProvider = Provider<WebAuthnBrowserClient>((ref) {
  return const WebAuthnBrowserClient();
});

final authRepositoryProvider = Provider<AuthRepositoryImpl>((ref) {
  final remoteDataSource = ref.watch(authRemoteDataSourceProvider);
  return AuthRepositoryImpl(remoteDataSource);
});

final authUseCaseProvider = Provider<AuthUseCase>((ref) {
  final repository = ref.watch(authRepositoryProvider);
  return AuthUseCase(repository);
});
