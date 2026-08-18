class CardsDto {
  const CardsDto({required this.id});

  final String id;

  factory CardsDto.fromJson(Map<String, dynamic> json) {
    return CardsDto(id: json['id'] as String? ?? '');
  }

  Map<String, dynamic> toJson() => <String, dynamic>{'id': id};
}
