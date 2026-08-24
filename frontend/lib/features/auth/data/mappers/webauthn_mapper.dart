import 'package:web_authn_web/web_authn_web.dart';

/// Converts between server JSON and web_authn_web's typed option/credential
/// objects, isolated from AuthRepositoryImpl so the repository only
/// orchestrates the flow.
class WebAuthnMapper {
  WebAuthnMapper._();

  static PublicKeyCredentialRequestOptions requestOptionsFromJson(Map<String, dynamic> json) {
    return PublicKeyCredentialRequestOptions(
      challenge: json['challenge'] as String,
      timeout: (json['timeout'] as num?)?.toInt(),
      rpId: json['rpId'] as String?,
      userVerification: json['userVerification'] as String?,
    );
  }

  static PublicKeyCredentialCreationOptions creationOptionsFromJson(Map<String, dynamic> json) {
    final rp = Map<String, dynamic>.from(json['rp'] as Map);
    final user = Map<String, dynamic>.from(json['user'] as Map);

    final pubKeyCredParams = (json['pubKeyCredParams'] as List<dynamic>? ?? []).map((item) {
      final value = Map<String, dynamic>.from(item as Map);
      return PubKeyCredParam(
        type: value['type'] as String? ?? 'public-key',
        alg: (value['alg'] as num?)?.toInt() ?? -7,
      );
    }).toList();

    return PublicKeyCredentialCreationOptions(
      rp: RpEntity(name: rp['name'] as String, id: rp['id'] as String?),
      user: UserEntity(
        name: user['name'] as String,
        id: user['id'] as String,
        displayName: user['displayName'] as String? ?? user['name'] as String,
      ),
      challenge: json['challenge'] as String,
      pubKeyCredParams: pubKeyCredParams,
      timeout: (json['timeout'] as num?)?.toInt(),
      attestation: json['attestation'] as String?,
    );
  }

  static Map<String, dynamic> credentialToJson(dynamic credential) {
    if (credential is! Map) {
      throw Exception('WebAuthn returned an invalid credential.');
    }

    final result = Map<String, dynamic>.from(credential);

    if (result['rawId'] is String) {
      result['rawId'] = _normalizeBase64Url(result['rawId'] as String);
    }

    if (result['response'] is Map) {
      final response = Map<String, dynamic>.from(result['response'] as Map);
      for (final key in ['clientDataJSON', 'attestationObject', 'authenticatorData', 'signature']) {
        if (response[key] is String) {
          response[key] = _normalizeBase64Url(response[key] as String);
        }
      }
      result['response'] = response;
    }

    return result;
  }

  static String _normalizeBase64Url(String value) =>
      value.replaceAll('+', '-').replaceAll('/', '_').replaceAll('=', '');
}