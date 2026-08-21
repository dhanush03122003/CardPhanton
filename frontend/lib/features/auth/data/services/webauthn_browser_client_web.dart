import 'package:web_authn_web/web_authn_web.dart';

class WebAuthnBrowserClient {
  const WebAuthnBrowserClient();

  Future<Map<String, dynamic>> createCredential(
  Map<String, dynamic> publicKey,
) async {
  try {
    final options = _createRegistrationOptions(publicKey);

print('========== WebAuthn Registration Options ==========');
print(publicKey);
print('===================================================');
final webAuthn = WebAuthnWeb();
final result = await webAuthn.register(options);

    return _normalizeResult(result);
  } catch (e, stackTrace) {
    print('========== WebAuthn Registration Error ==========');
    print('Error type: ${e.runtimeType}');
    print('Error: $e');
    print('Stack trace: $stackTrace');
    print('=================================================');

    throw StateError(
      'WebAuthn registration failed: $e',
    );
  }
}

  Future<Map<String, dynamic>> getAssertion(
    Map<String, dynamic> publicKey, {
    bool conditional = false,
  }) async {
    try {
      final webAuthn = WebAuthnWeb();

      final options = _createAuthenticationOptions(publicKey);

      final result = await webAuthn.sign(options);

      return _normalizeResult(result);
    } on WebAuthnWebException catch (e) {
      throw StateError(
        e.message.isNotEmpty
            ? e.message
            : 'WebAuthn authentication failed.',
      );
    } catch (e) {
      throw StateError(
        'WebAuthn authentication failed: $e',
      );
    }
  }

  PublicKeyCredentialCreationOptions _createRegistrationOptions(
    Map<String, dynamic> publicKey,
  ) {
    final rp = _asMap(
      publicKey['rp'],
      fieldName: 'rp',
    );

    final user = _asMap(
      publicKey['user'],
      fieldName: 'user',
    );

    final pubKeyCredParams =
        _asList(publicKey['pubKeyCredParams'])
            .map(
              (dynamic item) {
                final value = _asMap(
                  item,
                  fieldName: 'pubKeyCredParams item',
                );

                return PubKeyCredParam(
                  type: _asString(
                    value['type'],
                    fieldName: 'pubKeyCredParams.type',
                  ),
                  alg: _asInt(
                    value['alg'],
                    fieldName: 'pubKeyCredParams.alg',
                  ),
                );
              },
            )
            .toList();

    return PublicKeyCredentialCreationOptions(
      rp: RpEntity(
        name: _asString(
          rp['name'],
          fieldName: 'rp.name',
        ),
        id: rp['id'] as String?,
      ),
      user: UserEntity(
        name: _asString(
          user['name'],
          fieldName: 'user.name',
        ),
        id: _asString(
          user['id'],
          fieldName: 'user.id',
        ),
        displayName: _asString(
          user['displayName'] ?? user['name'],
          fieldName: 'user.displayName',
        ),
      ),
      challenge: _asString(
        publicKey['challenge'],
        fieldName: 'challenge',
      ),
      pubKeyCredParams: pubKeyCredParams,
      timeout: _asIntOrNull(
        publicKey['timeout'],
      ),
      attestation: publicKey['attestation'] as String?,
      authenticatorSelection:
          _createAuthenticatorSelection(
        publicKey['authenticatorSelection'],
      ),
      excludeCredentials:
          _createCredentialDescriptors(
        publicKey['excludeCredentials'],
      ),
    );
  }

  PublicKeyCredentialRequestOptions _createAuthenticationOptions(
    Map<String, dynamic> publicKey,
  ) {
    return PublicKeyCredentialRequestOptions(
      challenge: _asString(
        publicKey['challenge'],
        fieldName: 'challenge',
      ),
      rpId: publicKey['rpId'] as String?,
      timeout: _asIntOrNull(
        publicKey['timeout'],
      ),
      userVerification:
          publicKey['userVerification'] as String?,
      allowCredentials:
          _createCredentialDescriptors(
        publicKey['allowCredentials'],
      ),
    );
  }

  AuthenticatorSelectionCriteria?
      _createAuthenticatorSelection(
    dynamic value,
  ) {
    if (value == null) {
      return null;
    }

    final selection = _asMap(
      value,
      fieldName: 'authenticatorSelection',
    );

    return AuthenticatorSelectionCriteria(
      authenticatorAttachment:
          selection['authenticatorAttachment'] as String?,
      residentKey:
          selection['residentKey'] as String?,
      userVerification:
          selection['userVerification'] as String?,
      requireResidentKey:
          selection['requireResidentKey'] as bool?,
    );
  }

  List<CredentialDescriptor>?
      _createCredentialDescriptors(
    dynamic value,
  ) {
    if (value == null) {
      return null;
    }

    final list = _asList(value);

    if (list.isEmpty) {
      return <CredentialDescriptor>[];
    }

    return list.map(
      (dynamic item) {
        final credential = _asMap(
          item,
          fieldName: 'credential descriptor',
        );

        return CredentialDescriptor(
          id: _asString(
            credential['id'],
            fieldName: 'credential.id',
          ),
          type: credential['type'] as String? ?? 'public-key',
        );
      },
    ).toList();
  }

  Map<String, dynamic> _normalizeResult(
    dynamic result,
  ) {
    if (result == null) {
      throw StateError(
        'WebAuthn returned no credential.',
      );
    }

    if (result is Map<String, dynamic>) {
      return result;
    }

    if (result is Map) {
      return Map<String, dynamic>.from(result);
    }

    throw StateError(
      'Unsupported WebAuthn result type: '
      '${result.runtimeType}',
    );
  }

  Map<String, dynamic> _asMap(
    dynamic value, {
    required String fieldName,
  }) {
    if (value is Map<String, dynamic>) {
      return value;
    }

    if (value is Map) {
      return Map<String, dynamic>.from(value);
    }

    throw StateError(
      'WebAuthn field "$fieldName" must be an object. '
      'Received ${value.runtimeType}.',
    );
  }

  List<dynamic> _asList(dynamic value) {
    if (value is List) {
      return value;
    }

    throw StateError(
      'Expected a list but received ${value.runtimeType}.',
    );
  }

  String _asString(
    dynamic value, {
    required String fieldName,
  }) {
    if (value is String && value.isNotEmpty) {
      return value;
    }

    throw StateError(
      'WebAuthn field "$fieldName" must be a non-empty string.',
    );
  }

  int _asInt(
    dynamic value, {
    required String fieldName,
  }) {
    if (value is int) {
      return value;
    }

    if (value is num) {
      return value.toInt();
    }

    throw StateError(
      'WebAuthn field "$fieldName" must be an integer.',
    );
  }

  int? _asIntOrNull(dynamic value) {
    if (value == null) {
      return null;
    }

    if (value is int) {
      return value;
    }

    if (value is num) {
      return value.toInt();
    }

    return null;
  }
}