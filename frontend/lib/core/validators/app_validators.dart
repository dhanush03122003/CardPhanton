class AppValidators {
  AppValidators._();

  static String? requiredField(String? value, {String field = 'Field'}) {
    if (value == null || value.trim().isEmpty) {
      return '$field is required';
    }
    return null;
  }

  static String? email(String? value) {
    final requiredError = requiredField(value, field: 'Email');
    if (requiredError != null) return requiredError;

    final pattern = RegExp(r'^[^@\s]+@[^@\s]+\.[^@\s]+$');
    return pattern.hasMatch(value!.trim()) ? null : 'Invalid email address';
  }

  static String? password(String? value) {
    final requiredError = requiredField(value, field: 'Password');
    if (requiredError != null) return requiredError;

    if (value!.length < 8) return 'Password must be at least 8 characters';
    return null;
  }

  static String? cardNumber(String? value) {
    final requiredError = requiredField(value, field: 'Card number');
    if (requiredError != null) return requiredError;

    final digitsOnly = value!.replaceAll(RegExp(r'\s+'), '');
    final pattern = RegExp(r'^\d{12,19}$');
    return pattern.hasMatch(digitsOnly) ? null : 'Invalid card number';
  }

  static String? expiry(String? value) {
    final requiredError = requiredField(value, field: 'Expiry');
    if (requiredError != null) return requiredError;

    final pattern = RegExp(r'^(0[1-9]|1[0-2])\/(\d{2})$');
    return pattern.hasMatch(value!.trim()) ? null : 'Invalid expiry (MM/YY)';
  }

  static String? cvv(String? value) {
    final requiredError = requiredField(value, field: 'CVV');
    if (requiredError != null) return requiredError;

    final pattern = RegExp(r'^\d{3,4}$');
    return pattern.hasMatch(value!.trim()) ? null : 'Invalid CVV';
  }

  static String? phone(String? value) {
    final requiredError = requiredField(value, field: 'Phone');
    if (requiredError != null) return requiredError;

    final digitsOnly = value!.replaceAll(RegExp(r'[^\d]'), '');
    final pattern = RegExp(r'^\d{10,15}$');
    return pattern.hasMatch(digitsOnly) ? null : 'Invalid phone number';
  }
}
