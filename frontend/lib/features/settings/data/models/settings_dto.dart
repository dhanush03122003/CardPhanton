class SettingsDto {
  const SettingsDto({required this.id});

  final String id;

  factory SettingsDto.fromJson(Map<String, dynamic> json) {
    return SettingsDto(id: json['id'] as String? ?? '');
  }

  Map<String, dynamic> toJson() => <String, dynamic>{'id': id};
}
