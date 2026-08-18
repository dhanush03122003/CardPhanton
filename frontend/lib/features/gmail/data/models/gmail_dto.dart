class GmailDto {
  const GmailDto({required this.id});

  final String id;

  factory GmailDto.fromJson(Map<String, dynamic> json) {
    return GmailDto(id: json['id'] as String? ?? '');
  }

  Map<String, dynamic> toJson() => <String, dynamic>{'id': id};
}
