import 'package:flutter/foundation.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

enum UserRole {
  user,
  admin,
}

class AuthSession extends ChangeNotifier {
  bool _isAuthenticated = false;
  UserRole _role = UserRole.user;
  String? _username;

  bool get isAuthenticated => _isAuthenticated;
  UserRole get role => _role;
  String? get username => _username;

  void updateAuth({
    required bool isAuthenticated,
    required UserRole role,
    String? username,
  }) {
    _isAuthenticated = isAuthenticated;
    _role = role;
    _username = username;
    notifyListeners();
  }

  void clear() {
    _isAuthenticated = false;
    _role = UserRole.user;
    _username = null;
    notifyListeners();
  }
}

final authSessionProvider = ChangeNotifierProvider<AuthSession>((ref) {
  return AuthSession();
});