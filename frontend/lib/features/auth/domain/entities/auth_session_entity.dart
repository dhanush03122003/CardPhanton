class AuthSessionEntity {
  const AuthSessionEntity({
    required this.valid,
    this.username,
  });

  final bool valid;
  final String? username;
}