class AdminDto {
  const AdminDto({required this.id});

  final String id;

  factory AdminDto.fromJson(Map<String, dynamic> json) {
    return AdminDto(id: json['id'] as String? ?? '');
  }

  Map<String, dynamic> toJson() => <String, dynamic>{'id': id};
}
