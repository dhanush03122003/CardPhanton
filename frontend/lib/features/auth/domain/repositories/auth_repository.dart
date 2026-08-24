import '../entities/user.dart';

abstract interface class AuthRepository {
  Future<User> login(String username);
  Future<User> register(String username);
  Future<User> getCurrentUser();
}
