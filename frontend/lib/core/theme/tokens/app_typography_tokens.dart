import 'package:flutter/material.dart';

class AppTypographyTokens {
  AppTypographyTokens._();

  static const String _fontFamily = 'NotoSans';

  static const TextTheme textTheme = TextTheme(
    headlineLarge: TextStyle(fontFamily: _fontFamily, fontWeight: FontWeight.w700, fontSize: 32),
    headlineMedium: TextStyle(fontFamily: _fontFamily, fontWeight: FontWeight.w700, fontSize: 28),
    titleLarge: TextStyle(fontFamily: _fontFamily, fontWeight: FontWeight.w600, fontSize: 22),
    titleMedium: TextStyle(fontFamily: _fontFamily, fontWeight: FontWeight.w600, fontSize: 18),
    bodyLarge: TextStyle(fontFamily: _fontFamily, fontWeight: FontWeight.w400, fontSize: 16),
    bodyMedium: TextStyle(fontFamily: _fontFamily, fontWeight: FontWeight.w400, fontSize: 14),
    labelLarge: TextStyle(fontFamily: _fontFamily, fontWeight: FontWeight.w600, fontSize: 14),
  );
}
