class AppException implements Exception {
  AppException(this.message);

  final String message;

  @override
  String toString() => 'AppException: $message';
}

class UnauthorizedException extends AppException {
  UnauthorizedException(super.message);
}

class ParsingException extends AppException {
  ParsingException(super.message);
}
