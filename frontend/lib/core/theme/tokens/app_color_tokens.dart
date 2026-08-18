import 'package:flutter/material.dart';

class AppColorTokens {
  const AppColorTokens._({
    required this.background,
    required this.surface,
    required this.primary,
    required this.onPrimary,
    required this.error,
    required this.onError,
  });

  const AppColorTokens.light()
      : this._(
          background: const Color(0xFFF7F8FA),
          surface: Colors.white,
          primary: const Color(0xFF0C63E7),
          onPrimary: Colors.white,
          error: const Color(0xFFB3261E),
          onError: Colors.white,
        );

  const AppColorTokens.dark()
      : this._(
          background: const Color(0xFF0F1115),
          surface: const Color(0xFF1A1E25),
          primary: const Color(0xFF7FA8FF),
          onPrimary: const Color(0xFF001A4D),
          error: const Color(0xFFF2B8B5),
          onError: const Color(0xFF601410),
        );

  final Color background;
  final Color surface;
  final Color primary;
  final Color onPrimary;
  final Color error;
  final Color onError;

  ColorScheme get colorScheme {
    return ColorScheme.fromSeed(
      seedColor: primary,
      brightness: background.computeLuminance() > 0.5 ? Brightness.light : Brightness.dark,
      surface: surface,
      error: error,
      onError: onError,
    ).copyWith(
      primary: primary,
      onPrimary: onPrimary,
      surface: surface,
    );
  }
}
