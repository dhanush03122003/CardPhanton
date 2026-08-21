class WebAuthnBrowserClient {
  const WebAuthnBrowserClient();

  Future<Map<String, dynamic>> createCredential(
    Map<String, dynamic> publicKey,
  ) {
    throw UnsupportedError(
      'WebAuthn is only available on the web platform.',
    );
  }

  Future<Map<String, dynamic>> getAssertion(
    Map<String, dynamic> publicKey, {
    bool conditional = false,
  }) {
    throw UnsupportedError(
      'WebAuthn is only available on the web platform.',
    );
  }
}