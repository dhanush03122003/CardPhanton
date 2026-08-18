import 'package:flutter/material.dart';

import '../tokens/app_motion_tokens.dart';
import '../tokens/app_radius_tokens.dart';
import '../tokens/app_spacing_tokens.dart';

class ThemeTokensExtension extends ThemeExtension<ThemeTokensExtension> {
  const ThemeTokensExtension({
    required this.spacing,
    required this.radius,
    required this.motion,
  });

  final AppSpacingTokens spacing;
  final AppRadiusTokens radius;
  final AppMotionTokens motion;

  @override
  ThemeExtension<ThemeTokensExtension> copyWith({
    AppSpacingTokens? spacing,
    AppRadiusTokens? radius,
    AppMotionTokens? motion,
  }) {
    return ThemeTokensExtension(
      spacing: spacing ?? this.spacing,
      radius: radius ?? this.radius,
      motion: motion ?? this.motion,
    );
  }

  @override
  ThemeExtension<ThemeTokensExtension> lerp(
    covariant ThemeExtension<ThemeTokensExtension>? other,
    double t,
  ) {
    return this;
  }
}
