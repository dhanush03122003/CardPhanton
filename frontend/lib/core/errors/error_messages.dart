class ErrorMessages {
  ErrorMessages._();

  static const sessionExpired = 'Your session has expired. Please try again.';
  static const userNotFound = 'User not found.';
  static const invalidRequest = 'Invalid request. Please try again.';
  static const generic = 'Something went wrong. Please try again.';
  static const authenticatorAlreadyRegistered =
      'This authenticator is already registered. Please try logging in.';
  static const operationCancelled = 'The operation was cancelled or not allowed.';
  static const webAuthnUnavailable = 'WebAuthn is not available for this website.';
}