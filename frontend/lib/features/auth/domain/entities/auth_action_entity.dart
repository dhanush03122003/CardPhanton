class AuthActionEntity {
  const AuthActionEntity({
    required this.success,
    required this.message,
    this.username,
  });

  final bool success;
  final String message;
  final String? username;
}