import 'package:dio/dio.dart';
import 'package:flutter/foundation.dart';
import 'package:web_authn_web/web_authn_web.dart';

import 'error_messages.dart';

/// Maps low-level exceptions from the auth flow into user-facing messages.
/// Always logs the raw error first so failures are never silently swallowed.
class AuthErrorMapper {
  AuthErrorMapper._();

  static String map(Object error, [StackTrace? stackTrace]) {
    debugPrint('Auth failure: $error');
    if (stackTrace != null) debugPrint(stackTrace.toString());

    if (error is DioException) return _fromDio(error);
    if (error is WebAuthnWebException) return _fromWebAuthn(error);
    return ErrorMessages.generic;
  }

  static String _fromDio(DioException error) {
    final data = error.response?.data;
    if (data is Map && data['error'] is String) return data['error'] as String;

    switch (error.response?.statusCode) {
      case 401:
        return ErrorMessages.sessionExpired;
      case 404:
        return ErrorMessages.userNotFound;
      case 400:
        return ErrorMessages.invalidRequest;
      default:
        return ErrorMessages.generic;
    }
  }

  static String _fromWebAuthn(WebAuthnWebException error) {
    final cause = error.cause?.toString() ?? '';
    if (cause.contains('InvalidStateError'))
      return ErrorMessages.authenticatorAlreadyRegistered;
    if (cause.contains('NotAllowedError'))
      return ErrorMessages.operationCancelled;
    if (cause.contains('SecurityError'))
      return ErrorMessages.webAuthnUnavailable;
    return error.message;
  }
}
