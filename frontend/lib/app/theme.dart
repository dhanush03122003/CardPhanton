import 'package:flutter/material.dart';

import '../core/theme/tokens/app_color_tokens.dart';
import '../core/theme/tokens/app_motion_tokens.dart';
import '../core/theme/tokens/app_radius_tokens.dart';
import '../core/theme/tokens/app_spacing_tokens.dart';
import '../core/theme/tokens/app_typography_tokens.dart';
import '../core/theme/extensions/theme_tokens_extension.dart';

class AppTheme {
  AppTheme._();

  static ThemeData get light {
    const colors = AppColorTokens.light();

    return ThemeData(
      useMaterial3: true,
      brightness: Brightness.light,
      colorScheme: colors.colorScheme,
      scaffoldBackgroundColor: colors.background,
      textTheme: AppTypographyTokens.textTheme,
      extensions: <ThemeExtension<dynamic>>[
        const ThemeTokensExtension(
          spacing: AppSpacingTokens(),
          radius: AppRadiusTokens(),
          motion: AppMotionTokens(),
        ),
      ],
    );
  }

  static ThemeData get dark {
    const colors = AppColorTokens.dark();

    return ThemeData(
      useMaterial3: true,
      brightness: Brightness.dark,
      colorScheme: colors.colorScheme,
      scaffoldBackgroundColor: colors.background,
      textTheme: AppTypographyTokens.textTheme,
      extensions: <ThemeExtension<dynamic>>[
        const ThemeTokensExtension(
          spacing: AppSpacingTokens(),
          radius: AppRadiusTokens(),
          motion: AppMotionTokens(),
        ),
      ],
    );
  }
}
