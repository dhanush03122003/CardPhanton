import 'package:flutter/material.dart';

import '../theme/extensions/theme_tokens_extension.dart';

extension ThemeContextX on BuildContext {
  ThemeTokensExtension get tokens {
    return Theme.of(this).extension<ThemeTokensExtension>()!;
  }
}
