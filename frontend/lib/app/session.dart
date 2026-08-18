import 'package:flutter/foundation.dart';

enum UserRole {
  user,
  admin,
}

class AuthSession extends ChangeNotifier {
  bool _isAuthenticated = false;
  UserRole _role = UserRole.user;

  bool get isAuthenticated => _isAuthenticated;
  UserRole get role => _role;

  void updateAuth({
    required bool isAuthenticated,
    required UserRole role,
  }) {
    _isAuthenticated = isAuthenticated;
    _role = role;
    notifyListeners();
  }

  void clear() {
    _isAuthenticated = false;
    _role = UserRole.user;
    notifyListeners();
  }
}