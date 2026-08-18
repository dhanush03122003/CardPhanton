class NotificationsDto {
  const NotificationsDto({required this.id});

  final String id;

  factory NotificationsDto.fromJson(Map<String, dynamic> json) {
    return NotificationsDto(id: json['id'] as String? ?? '');
  }

  Map<String, dynamic> toJson() => <String, dynamic>{'id': id};
}
