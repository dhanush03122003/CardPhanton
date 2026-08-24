import '../../domain/entities/user.dart';

class UserModel extends User {
  const UserModel({
    required super.username,
  });

  factory UserModel.fromJson(Map<String, dynamic> json) {
    return UserModel(
      username: json['username'] as String? ?? '',
    );
  }
}