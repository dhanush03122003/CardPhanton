class Validators {
  Validators._();

  static String? username(String? value) {
    final trimmed = value?.trim() ?? '';
    if (trimmed.isEmpty) return 'Please enter a username.';
    if (trimmed.length < 3) return 'Username must be at least 3 characters.';
    return null;
  }
}
