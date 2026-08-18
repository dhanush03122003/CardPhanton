class AuthDto {
  const AuthDto({required this.id});

  final String id;

  factory AuthDto.fromJson(Map<String, dynamic> json) {
    return AuthDto(id: json['id'] as String? ?? '');
  }

  Map<String, dynamic> toJson() => <String, dynamic>{'id': id};
}
