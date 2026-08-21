class AuthActionDto {
  const AuthActionDto({
    required this.success,
    required this.message,
    this.username,
  });

  final bool success;
  final String message;
  final String? username;

  factory AuthActionDto.fromJson(
    Map<String, dynamic> json, {
    String? fallbackUsername,
  }) {
    return AuthActionDto(
      success: json['success'] == true,
      message: json['message'] as String? ?? '',
      username: json['username'] as String? ?? fallbackUsername,
    );
  }
}

class AuthSessionDto {
  const AuthSessionDto({
    required this.valid,
    this.username,
  });

  final bool valid;
  final String? username;

  factory AuthSessionDto.fromJson(Map<String, dynamic> json) {
    return AuthSessionDto(
      valid: json['valid'] == true,
      username: json['username'] as String?,
    );
  }
}
