class ProfileDto {
  const ProfileDto({required this.id});

  final String id;

  factory ProfileDto.fromJson(Map<String, dynamic> json) {
    return ProfileDto(id: json['id'] as String? ?? '');
  }

  Map<String, dynamic> toJson() => <String, dynamic>{'id': id};
}
