import 'dart:convert';

import 'trading_authorization.dart';

const walletConnectRequestSchemaV1 = '2finance.wallet_connect_request.v1';
const walletConnectResponseSchemaV1 = '2finance.wallet_connect_response.v1';
const walletDeepLinkScheme = 'twofinance';
const walletUniversalLinkHost = 'wallet.2finance.com';

/// Public, self-contained request transported by QR, app link or deep link.
/// It deliberately contains no bearer token, cookie, seed or private key.
final class TradingSessionConnectRequest {
  TradingSessionConnectRequest({
    required this.challengeId,
    required this.challengeExpiresAt,
    required this.responseUri,
    required this.payload,
  }) {
    _validateIdentifier(challengeId, 'challenge_id');
    _validateOriginAndResponseUri(
      origin: payload.origin,
      responseUri: responseUri,
      challengeId: challengeId,
    );
  }

  final String challengeId;
  final DateTime challengeExpiresAt;
  final Uri responseUri;
  final TradingSessionAuthorizationPayload payload;

  String get origin => payload.origin;

  Map<String, dynamic> toJson() => {
    'schema': walletConnectRequestSchemaV1,
    'challenge_id': challengeId,
    'challenge_expires_at': challengeExpiresAt.toUtc().toIso8601String(),
    'response_uri': responseUri.toString(),
    'payload': payload.toJson(),
  };

  String encode() =>
      base64Url.encode(utf8.encode(jsonEncode(toJson()))).replaceAll('=', '');

  Uri toDeepLink() => Uri(
    scheme: walletDeepLinkScheme,
    host: 'exchange',
    path: '/connect',
    fragment: Uri(queryParameters: {'request': encode()}).query,
  );

  Uri toUniversalLink({String host = walletUniversalLinkHost}) => Uri(
    scheme: 'https',
    host: host,
    path: '/connect/exchange',
    fragment: Uri(queryParameters: {'request': encode()}).query,
  );

  factory TradingSessionConnectRequest.parse(
    String input, {
    required Iterable<String> trustedOrigins,
    DateTime? now,
  }) {
    final uri = Uri.tryParse(input.trim());
    if (uri == null || !_isSupportedWalletLink(uri)) {
      throw const FormatException('unsupported 2Finance Wallet link');
    }
    final fragment = Uri.splitQueryString(uri.fragment);
    final encoded = fragment['request'];
    if (encoded == null ||
        encoded.isEmpty ||
        fragment.length != 1 ||
        uri.query.isNotEmpty) {
      throw const FormatException(
        'wallet request must be carried in the URL fragment',
      );
    }
    return TradingSessionConnectRequest.decode(
      encoded,
      trustedOrigins: trustedOrigins,
      now: now,
    );
  }

  factory TradingSessionConnectRequest.decode(
    String encoded, {
    required Iterable<String> trustedOrigins,
    DateTime? now,
  }) {
    if (encoded.length > 16384 ||
        !RegExp(r'^[A-Za-z0-9_-]+$').hasMatch(encoded)) {
      throw const FormatException('invalid wallet request encoding');
    }
    Object? decoded;
    try {
      decoded = jsonDecode(
        utf8.decode(base64Url.decode(base64Url.normalize(encoded))),
      );
    } catch (_) {
      throw const FormatException('invalid wallet request encoding');
    }
    if (decoded is! Map) {
      throw const FormatException('wallet request must be a JSON object');
    }
    return TradingSessionConnectRequest.fromJson(
      Map<String, dynamic>.from(decoded),
      trustedOrigins: trustedOrigins,
      now: now,
    );
  }

  factory TradingSessionConnectRequest.fromJson(
    Map<String, dynamic> json, {
    required Iterable<String> trustedOrigins,
    DateTime? now,
  }) {
    _requireExactKeys(json, const {
      'schema',
      'challenge_id',
      'challenge_expires_at',
      'response_uri',
      'payload',
    });
    if (json['schema'] != walletConnectRequestSchemaV1) {
      throw const FormatException('unsupported wallet request schema');
    }
    final rawPayload = json['payload'];
    if (rawPayload is! Map) {
      throw const FormatException('wallet request payload is required');
    }
    final payload = _parseTradingSessionPayload(
      Map<String, dynamic>.from(rawPayload),
    );
    final challengeId = _requiredString(json, 'challenge_id');
    final challengeExpiresAt = DateTime.tryParse(
      _requiredString(json, 'challenge_expires_at'),
    )?.toUtc();
    final responseUri = Uri.tryParse(_requiredString(json, 'response_uri'));
    if (challengeExpiresAt == null || responseUri == null) {
      throw const FormatException('invalid challenge expiry or response URI');
    }

    final current = (now ?? DateTime.now()).toUtc();
    final issuedAt = DateTime.fromMillisecondsSinceEpoch(
      payload.issuedAtMs,
      isUtc: true,
    );
    final sessionExpiry = DateTime.fromMillisecondsSinceEpoch(
      payload.expiresAtMs,
      isUtc: true,
    );
    if (!current.isBefore(challengeExpiresAt) ||
        challengeExpiresAt.isAfter(sessionExpiry) ||
        issuedAt.isAfter(current.add(const Duration(minutes: 2))) ||
        sessionExpiry.isAfter(issuedAt.add(const Duration(hours: 24)))) {
      throw const FormatException(
        'wallet request is expired or has an unsafe lifetime',
      );
    }

    final normalizedTrusted = trustedOrigins
        .map(_normalizeOrigin)
        .where((value) => value.isNotEmpty)
        .toSet();
    if (!normalizedTrusted.contains(_normalizeOrigin(payload.origin))) {
      throw const FormatException('exchange origin is not trusted');
    }
    return TradingSessionConnectRequest(
      challengeId: challengeId,
      challengeExpiresAt: challengeExpiresAt,
      responseUri: responseUri,
      payload: payload,
    );
  }

  factory TradingSessionConnectRequest.fromChallengeResponse(
    Map<String, dynamic> challenge, {
    required Uri exchangeOrigin,
    DateTime? now,
  }) {
    if (challenge['schema'] != '2finance.trading_session_challenge.v1' ||
        challenge['status'] != 'pending') {
      throw const FormatException('invalid or inactive exchange challenge');
    }
    final rawPayload = challenge['payload'];
    if (rawPayload is! Map) {
      throw const FormatException('exchange challenge payload is required');
    }
    final challengeId = _requiredString(challenge, 'challenge_id');
    final challengeExpiry = DateTime.tryParse(
      _requiredString(challenge, 'challenge_expires_at'),
    );
    if (challengeExpiry == null) {
      throw const FormatException('invalid exchange challenge expiry');
    }
    final payload = _parseTradingSessionPayload(
      Map<String, dynamic>.from(rawPayload),
    );
    final origin = Uri.parse(payload.origin);
    if (_normalizeOrigin(exchangeOrigin.origin) !=
        _normalizeOrigin(origin.origin)) {
      throw const FormatException(
        'challenge origin does not match exchange origin',
      );
    }
    final request = TradingSessionConnectRequest(
      challengeId: challengeId,
      challengeExpiresAt: challengeExpiry.toUtc(),
      responseUri: origin.replace(
        path:
            '/api/v2/exchange/session/challenges/${Uri.encodeComponent(challengeId)}/responses',
        query: '',
        fragment: '',
      ),
      payload: payload,
    );
    // Reuse the wallet-side lifetime checks before displaying a QR.
    return TradingSessionConnectRequest.fromJson(
      request.toJson(),
      trustedOrigins: [payload.origin],
      now: now,
    );
  }
}

final class TradingSessionConnectResponse {
  const TradingSessionConnectResponse({
    required this.challengeId,
    required this.envelope,
  });

  final String challengeId;
  final SignedAuthorizationEnvelope envelope;

  Map<String, dynamic> toJson() => {
    'schema': walletConnectResponseSchemaV1,
    'challenge_id': challengeId,
    'envelope': envelope.toJson(),
  };
}

bool _isSupportedWalletLink(Uri uri) {
  if (uri.scheme == walletDeepLinkScheme) {
    return uri.host == 'exchange' &&
        uri.path == '/connect' &&
        uri.userInfo.isEmpty &&
        !uri.hasPort;
  }
  return uri.scheme == 'https' &&
      uri.host == walletUniversalLinkHost &&
      uri.path == '/connect/exchange' &&
      uri.userInfo.isEmpty &&
      (!uri.hasPort || uri.port == 443);
}

TradingSessionAuthorizationPayload _parseTradingSessionPayload(
  Map<String, dynamic> json,
) {
  _requireExactKeys(json, const {
    'domain',
    'chain_id',
    'account',
    'session_id',
    'session_public_key',
    'allowed_symbols',
    'allowed_actions',
    'allowed_order_types',
    'max_order_amount',
    'max_order_notional',
    'max_session_notional',
    'issued_at_ms',
    'expires_at_ms',
    'nonce',
    'origin',
  });
  if (json['domain'] != tradingSessionDomainV1) {
    throw const FormatException('unsupported trading session domain');
  }
  return TradingSessionAuthorizationPayload(
    chainId: _requiredString(json, 'chain_id'),
    account: _requiredString(json, 'account'),
    sessionId: _requiredString(json, 'session_id'),
    sessionPublicKey: _requiredString(json, 'session_public_key'),
    allowedSymbols: _requiredStringList(json, 'allowed_symbols'),
    allowedActions: _requiredStringList(json, 'allowed_actions'),
    allowedOrderTypes: _requiredStringList(json, 'allowed_order_types'),
    maxOrderAmount: _requiredString(json, 'max_order_amount'),
    maxOrderNotional: _requiredString(json, 'max_order_notional'),
    maxSessionNotional: _requiredString(json, 'max_session_notional'),
    issuedAtMs: _requiredInt(json, 'issued_at_ms'),
    expiresAtMs: _requiredInt(json, 'expires_at_ms'),
    nonce: _requiredString(json, 'nonce'),
    origin: _requiredString(json, 'origin'),
  );
}

void _validateOriginAndResponseUri({
  required String origin,
  required Uri responseUri,
  required String challengeId,
}) {
  final originUri = Uri.parse(origin);
  final loopback =
      originUri.host == 'localhost' ||
      originUri.host == '127.0.0.1' ||
      originUri.host == '::1';
  if ((originUri.scheme != 'https' &&
          !(loopback && originUri.scheme == 'http')) ||
      originUri.userInfo.isNotEmpty ||
      (originUri.path.isNotEmpty && originUri.path != '/') ||
      originUri.query.isNotEmpty ||
      originUri.fragment.isNotEmpty ||
      _normalizeOrigin(origin) != _normalizeOrigin(responseUri.origin)) {
    throw const FormatException('unsafe exchange origin or response URI');
  }
  final expectedPath =
      '/api/v2/exchange/session/challenges/${Uri.encodeComponent(challengeId)}/responses';
  if (responseUri.path != expectedPath ||
      responseUri.query.isNotEmpty ||
      responseUri.fragment.isNotEmpty ||
      responseUri.userInfo.isNotEmpty) {
    throw const FormatException('response URI is not bound to the challenge');
  }
}

String _normalizeOrigin(String value) {
  final uri = Uri.tryParse(value.trim());
  if (uri == null || !uri.hasScheme || uri.host.isEmpty) return '';
  final defaultPort =
      (uri.scheme == 'https' && uri.port == 443) ||
      (uri.scheme == 'http' && uri.port == 80);
  return '${uri.scheme.toLowerCase()}://${uri.host.toLowerCase()}'
      '${defaultPort || !uri.hasPort ? '' : ':${uri.port}'}';
}

void _validateIdentifier(String value, String field) {
  if (!RegExp(r'^[A-Za-z0-9][A-Za-z0-9._:-]{0,254}$').hasMatch(value)) {
    throw FormatException('$field has an unsupported format');
  }
}

void _requireExactKeys(Map<String, dynamic> json, Set<String> expected) {
  if (json.keys.toSet().difference(expected).isNotEmpty ||
      expected.difference(json.keys.toSet()).isNotEmpty) {
    throw const FormatException(
      'wallet request contains missing or unknown fields',
    );
  }
}

String _requiredString(Map<String, dynamic> json, String key) {
  final value = json[key];
  if (value is! String || value.trim().isEmpty || value != value.trim()) {
    throw FormatException('$key must be a non-empty canonical string');
  }
  return value;
}

int _requiredInt(Map<String, dynamic> json, String key) {
  final value = json[key];
  if (value is! int || value <= 0) {
    throw FormatException('$key must be a positive integer');
  }
  return value;
}

List<String> _requiredStringList(Map<String, dynamic> json, String key) {
  final value = json[key];
  if (value is! List || value.isEmpty || value.any((item) => item is! String)) {
    throw FormatException('$key must be a non-empty string list');
  }
  final strings = value.cast<String>();
  final canonical = canonicalStringSet(strings);
  if (jsonEncode(strings) != jsonEncode(canonical)) {
    throw FormatException('$key must be unique and sorted');
  }
  return canonical;
}
