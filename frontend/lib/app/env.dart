enum AppEnvironment {
  dev,
  stage,
  prod;

  static AppEnvironment fromString(String raw) {
    switch (raw.toLowerCase()) {
      case 'stage':
        return AppEnvironment.stage;
      case 'prod':
        return AppEnvironment.prod;
      case 'dev':
      default:
        return AppEnvironment.dev;
    }
  }
}

class AppEnv {
  AppEnv._();

  static const String _env = String.fromEnvironment('APP_ENV', defaultValue: 'dev');
  static const String _baseUrl = String.fromEnvironment(
    'API_BASE_URL',
    defaultValue: 'http://localhost:8080',
  );

  static AppEnvironment get environment => AppEnvironment.fromString(_env);
  static String get apiBaseUrl => _baseUrl;
  static bool get isProduction => environment == AppEnvironment.prod;
}
