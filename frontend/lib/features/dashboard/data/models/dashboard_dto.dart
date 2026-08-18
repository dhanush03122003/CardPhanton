class DashboardDto {
  const DashboardDto({required this.id});

  final String id;

  factory DashboardDto.fromJson(Map<String, dynamic> json) {
    return DashboardDto(id: json['id'] as String? ?? '');
  }

  Map<String, dynamic> toJson() => <String, dynamic>{'id': id};
}
