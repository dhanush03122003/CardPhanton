import '../services/secure_storage_service.dart';

class TokenRefreshService {
  TokenRefreshService(this._secureStorageService);

  final SecureStorageService _secureStorageService;

  Future<String?> refreshAccessToken() async {
    // Placeholder: integrate refresh endpoint when backend contract is finalized.
    final currentRefreshToken = await _secureStorageService.getRefreshToken();
    if (currentRefreshToken == null || currentRefreshToken.isEmpty) {
      return null;
    }

    return null;
  }
}
