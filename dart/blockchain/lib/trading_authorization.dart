import 'dart:collection';
import 'dart:convert';
import 'dart:typed_data';

import 'package:cryptography/cryptography.dart';

import 'blockchain/keys/keys.dart';

const signedMessagePreambleV1 = '2FINANCE-SIGNED-MESSAGE-V1';
const tradingSessionDomainV1 = '2finance.trading_session.v1';
const tradingSessionRevokeDomainV1 = '2finance.trading_session.revoke.v1';
const orderAddDomainV1 = '2finance.order.add.v1';
const orderCancelDomainV1 = '2finance.order.cancel.v1';
const orderCancelAllDomainV1 = '2finance.order.cancel_all.v1';

enum OrderAuthorizationMode { wallet, session }

String canonicalJsonV1(Object? value) => jsonEncode(_canonicalValue(value));

Uint8List canonicalSigningBytesV1(Map<String, dynamic> payload) {
  final domain = payload['domain']?.toString() ?? '';
  if (domain.isEmpty) {
    throw const FormatException('domain is required');
  }
  return Uint8List.fromList(
    utf8.encode(
      '$signedMessagePreambleV1\n$domain\n${canonicalJsonV1(payload)}',
    ),
  );
}

Object? _canonicalValue(Object? value) {
  if (value is Map) {
    final sorted = SplayTreeMap<String, Object?>();
    for (final entry in value.entries) {
      sorted[entry.key.toString()] = _canonicalValue(entry.value);
    }
    return sorted;
  }
  if (value is List) {
    return value.map(_canonicalValue).toList(growable: false);
  }
  if (value == null || value is String || value is bool || value is int) {
    return value;
  }
  throw FormatException(
    'canonical payloads allow only objects, arrays, strings, booleans, integers and null; got ${value.runtimeType}',
  );
}

String canonicalDecimal(String input, {bool allowZero = true}) {
  var value = input.trim();
  if (!RegExp(r'^(?:0|[1-9][0-9]*)(?:\.[0-9]+)?$').hasMatch(value)) {
    throw FormatException('invalid unsigned decimal: $input');
  }
  if (value.contains('.')) {
    value = value.replaceFirst(RegExp(r'0+$'), '');
    value = value.replaceFirst(RegExp(r'\.$'), '');
  }
  if (!allowZero && value == '0') {
    throw const FormatException('decimal must be greater than zero');
  }
  return value;
}

String canonicalNonce(Object nonce) {
  final value = nonce.toString();
  if (!RegExp(r'^[1-9][0-9]{0,19}$').hasMatch(value)) {
    throw FormatException('nonce must be a positive canonical uint64: $nonce');
  }
  final parsed = BigInt.parse(value);
  if (parsed > BigInt.parse('18446744073709551615')) {
    throw FormatException('nonce exceeds uint64: $nonce');
  }
  return value;
}

String canonicalChallengeNonce(String nonce) {
  final value = nonce.trim();
  if (!RegExp(r'^[0-9a-f]{64}$').hasMatch(value) ||
      RegExp(r'^0{64}$').hasMatch(value)) {
    throw const FormatException(
      'session nonce must be a non-zero lowercase 256-bit hex string',
    );
  }
  return value;
}

List<String> canonicalStringSet(Iterable<String> values) {
  final sorted = values.map((value) => value.trim()).toSet().toList()..sort();
  if (sorted.isEmpty || sorted.any((value) => value.isEmpty)) {
    throw const FormatException('permission lists cannot be empty');
  }
  return List.unmodifiable(sorted);
}

final class TradingSessionAuthorizationPayload {
  TradingSessionAuthorizationPayload({
    required this.chainId,
    required this.account,
    required this.sessionId,
    required this.sessionPublicKey,
    required Iterable<String> allowedSymbols,
    required Iterable<String> allowedActions,
    required Iterable<String> allowedOrderTypes,
    required String maxOrderAmount,
    required String maxOrderNotional,
    required String maxSessionNotional,
    required this.issuedAtMs,
    required this.expiresAtMs,
    required Object nonce,
    required this.origin,
  }) : allowedSymbols = canonicalStringSet(allowedSymbols),
       allowedActions = canonicalStringSet(allowedActions),
       allowedOrderTypes = canonicalStringSet(allowedOrderTypes),
       maxOrderAmount = canonicalDecimal(maxOrderAmount, allowZero: false),
       maxOrderNotional = canonicalDecimal(maxOrderNotional, allowZero: false),
       maxSessionNotional = canonicalDecimal(
         maxSessionNotional,
         allowZero: false,
       ),
       nonce = canonicalChallengeNonce(nonce.toString()) {
    _requireText(chainId, 'chain_id');
    KeyManager.validateEDDSAPublicKeyHex(account);
    _requireText(sessionId, 'session_id');
    KeyManager.validateEDDSAPublicKeyHex(sessionPublicKey);
    if (issuedAtMs <= 0 || expiresAtMs <= issuedAtMs) {
      throw const FormatException('session expiration must follow issue time');
    }
    final uri = Uri.tryParse(origin);
    if (uri == null || !uri.hasScheme || uri.host.isEmpty) {
      throw const FormatException('origin must be an absolute URI');
    }
  }

  final String chainId;
  final String account;
  final String sessionId;
  final String sessionPublicKey;
  final List<String> allowedSymbols;
  final List<String> allowedActions;
  final List<String> allowedOrderTypes;
  final String maxOrderAmount;
  final String maxOrderNotional;
  final String maxSessionNotional;
  final int issuedAtMs;
  final int expiresAtMs;
  final String nonce;
  final String origin;

  Map<String, dynamic> toJson() => {
    'domain': tradingSessionDomainV1,
    'chain_id': chainId,
    'account': account,
    'session_id': sessionId,
    'session_public_key': sessionPublicKey,
    'allowed_symbols': allowedSymbols,
    'allowed_actions': allowedActions,
    'allowed_order_types': allowedOrderTypes,
    'max_order_amount': maxOrderAmount,
    'max_order_notional': maxOrderNotional,
    'max_session_notional': maxSessionNotional,
    'issued_at_ms': issuedAtMs,
    'expires_at_ms': expiresAtMs,
    'nonce': nonce,
    'origin': origin,
  };
}

final class OrderCommandPayload {
  OrderCommandPayload({
    required this.chainId,
    required this.account,
    required this.engineId,
    required this.symbol,
    required this.symbolId,
    required this.routeEpoch,
    required String operation,
    required this.authorizationMode,
    required String side,
    required String type,
    required String amount,
    required String price,
    required String timeInForce,
    this.orderId = '',
    required this.clientOrderId,
    required this.idempotencyKey,
    this.sessionId = '',
    required Object nonce,
    required this.expiresAtMs,
  }) : operation = operation.trim().toUpperCase(),
       side = side.trim().toUpperCase(),
       type = type.trim().toUpperCase(),
       amount = canonicalDecimal(amount),
       price = canonicalDecimal(price),
       timeInForce = timeInForce.trim().toUpperCase(),
       nonce = canonicalNonce(nonce) {
    _requireText(chainId, 'chain_id');
    _requireText(account, 'account');
    KeyManager.validateEDDSAPublicKeyHex(account);
    _requireText(engineId, 'engine_id');
    if (!RegExp(r'^[A-Z0-9._-]{1,32}/[A-Z0-9._-]{1,32}$').hasMatch(symbol)) {
      throw const FormatException('symbol must be canonical BASE/QUOTE');
    }
    if (symbolId <= 0 || routeEpoch <= 0 || expiresAtMs <= 0) {
      throw const FormatException(
        'symbol_id, route_epoch and expires_at_ms must be positive',
      );
    }
    _requireText(clientOrderId, 'client_order_id');
    _requireText(idempotencyKey, 'idempotency_key');
    _validateOperationShape();
  }

  final String chainId;
  final String account;
  final String engineId;
  final String symbol;
  final int symbolId;
  final int routeEpoch;
  final String operation;
  final OrderAuthorizationMode authorizationMode;
  final String side;
  final String type;
  final String amount;
  final String price;
  final String timeInForce;
  final String orderId;
  final String clientOrderId;
  final String idempotencyKey;
  final String sessionId;
  final String nonce;
  final int expiresAtMs;

  String get domain => switch (operation) {
    'ADD' => orderAddDomainV1,
    'CANCEL' => orderCancelDomainV1,
    'CANCEL_ALL' => orderCancelAllDomainV1,
    _ => throw FormatException('unsupported operation: $operation'),
  };

  String get requiredAction => switch (operation) {
    'ADD' => 'order:submit',
    'CANCEL' => 'order:cancel',
    'CANCEL_ALL' => 'order:cancel_all',
    _ => throw FormatException('unsupported operation: $operation'),
  };

  void _validateOperationShape() {
    if (operation == 'ADD') {
      if (side != 'BUY' && side != 'SELL') {
        throw const FormatException('ADD side must be BUY or SELL');
      }
      if (type != 'LIMIT' && type != 'MARKET') {
        throw const FormatException('ADD type must be LIMIT or MARKET');
      }
      canonicalDecimal(amount, allowZero: false);
      if (type == 'LIMIT') {
        canonicalDecimal(price, allowZero: false);
      } else if (price != '0') {
        throw const FormatException('MARKET price must be zero');
      }
      if (!const {'GTC', 'IOC', 'FOK'}.contains(timeInForce)) {
        throw const FormatException('unsupported time_in_force');
      }
      if (orderId.isNotEmpty) {
        throw const FormatException('ADD must not contain order_id');
      }
      return;
    }
    if (operation != 'CANCEL' && operation != 'CANCEL_ALL') {
      throw FormatException('unsupported operation: $operation');
    }
    if (side != 'NONE' ||
        type != 'NONE' ||
        amount != '0' ||
        price != '0' ||
        timeInForce != 'NONE') {
      throw const FormatException(
        'cancellation uses NONE side/type/TIF and zero amount/price',
      );
    }
    if ((operation == 'CANCEL') != orderId.isNotEmpty) {
      throw const FormatException(
        'CANCEL requires and CANCEL_ALL forbids order_id',
      );
    }
  }

  Map<String, dynamic> toJson() => {
    'domain': domain,
    'chain_id': chainId,
    'account': account,
    'engine_id': engineId,
    'symbol': symbol,
    'symbol_id': symbolId,
    'route_epoch': routeEpoch,
    'operation': operation,
    'authorization_mode': authorizationMode.name.toUpperCase(),
    'side': side,
    'type': type,
    'amount': amount,
    'price': price,
    'time_in_force': timeInForce,
    'order_id': orderId,
    'client_order_id': clientOrderId,
    'idempotency_key': idempotencyKey,
    'session_id': sessionId,
    'nonce': nonce,
    'expires_at_ms': expiresAtMs,
  };
}

final class TradingSessionRevocationPayload {
  TradingSessionRevocationPayload({
    required this.chainId,
    required this.account,
    required this.sessionId,
    required Object nonce,
    required this.expiresAtMs,
  }) : nonce = canonicalChallengeNonce(nonce.toString()) {
    _requireText(chainId, 'chain_id');
    KeyManager.validateEDDSAPublicKeyHex(account);
    _requireText(sessionId, 'session_id');
    if (expiresAtMs <= 0) {
      throw const FormatException('expires_at_ms must be positive');
    }
  }

  final String chainId;
  final String account;
  final String sessionId;
  final String nonce;
  final int expiresAtMs;

  Map<String, dynamic> toJson() => {
    'domain': tradingSessionRevokeDomainV1,
    'chain_id': chainId,
    'account': account,
    'session_id': sessionId,
    'nonce': nonce,
    'expires_at_ms': expiresAtMs,
  };
}

final class SignedAuthorizationEnvelope {
  const SignedAuthorizationEnvelope({
    required this.schema,
    required this.payload,
    required this.mode,
    required this.publicKey,
    required this.signature,
  });

  factory SignedAuthorizationEnvelope.fromJson(Map<String, dynamic> json) {
    final rawPayload = json['payload'];
    final rawAuthorization = json['authorization'];
    if (rawPayload is! Map || rawAuthorization is! Map) {
      throw const FormatException('invalid signed authorization envelope');
    }
    final mode = switch (rawAuthorization['mode']) {
      'WALLET' => OrderAuthorizationMode.wallet,
      'SESSION' => OrderAuthorizationMode.session,
      _ => throw const FormatException('unsupported authorization mode'),
    };
    final schema = json['schema'];
    final publicKey = rawAuthorization['public_key'];
    final signature = rawAuthorization['signature'];
    if (schema is! String ||
        publicKey is! String ||
        signature is! String ||
        !RegExp(r'^[0-9a-f]{64}$').hasMatch(publicKey) ||
        !RegExp(r'^[0-9a-f]{128}$').hasMatch(signature)) {
      throw const FormatException('invalid authorization encoding');
    }
    return SignedAuthorizationEnvelope(
      schema: schema,
      payload: Map<String, dynamic>.unmodifiable(
        Map<String, dynamic>.from(rawPayload),
      ),
      mode: mode,
      publicKey: publicKey,
      signature: signature,
    );
  }

  final String schema;
  final Map<String, dynamic> payload;
  final OrderAuthorizationMode mode;
  final String publicKey;
  final String signature;

  Map<String, dynamic> toJson() => {
    'schema': schema,
    'payload': payload,
    'authorization': {
      'mode': mode.name.toUpperCase(),
      'public_key': publicKey,
      'signature': signature,
    },
  };
}

final class Ed25519CanonicalSigner {
  Ed25519CanonicalSigner._(this._keyPair, this.publicKey);

  final SimpleKeyPair _keyPair;
  final String publicKey;

  static Future<Ed25519CanonicalSigner> generate() async {
    final keyPair = await Ed25519().newKeyPair();
    final publicKey = await keyPair.extractPublicKey();
    return Ed25519CanonicalSigner._(
      keyPair,
      KeyManager.bytesToHex(publicKey.bytes),
    );
  }

  static Future<Ed25519CanonicalSigner> fromPrivateKeyHex(
    String privateKeyHex,
  ) async {
    final privateBytes = KeyManager.hexToBytes(privateKeyHex);
    if (privateBytes.length < 32) {
      throw const FormatException(
        'Ed25519 private key needs at least 32 bytes',
      );
    }
    final keyPair = await Ed25519().newKeyPairFromSeed(
      privateBytes.sublist(0, 32),
    );
    final publicKey = await keyPair.extractPublicKey();
    return Ed25519CanonicalSigner._(
      keyPair,
      KeyManager.bytesToHex(publicKey.bytes),
    );
  }

  Future<SignedAuthorizationEnvelope> sign({
    required String schema,
    required Map<String, dynamic> payload,
    required OrderAuthorizationMode mode,
  }) async {
    final signature = await Ed25519().sign(
      canonicalSigningBytesV1(payload),
      keyPair: _keyPair,
    );
    return SignedAuthorizationEnvelope(
      schema: schema,
      payload: Map.unmodifiable(payload),
      mode: mode,
      publicKey: publicKey,
      signature: KeyManager.bytesToHex(signature.bytes),
    );
  }

  Future<SignedAuthorizationEnvelope> signTradingSession(
    TradingSessionAuthorizationPayload payload,
  ) => sign(
    schema: '2finance.trading_session_authorization.v1',
    payload: payload.toJson(),
    mode: OrderAuthorizationMode.wallet,
  );

  Future<SignedAuthorizationEnvelope> signOrder(
    OrderCommandPayload payload, {
    required OrderAuthorizationMode mode,
  }) {
    if (payload.authorizationMode != mode) {
      throw ArgumentError(
        'signed authorization_mode does not match envelope mode',
      );
    }
    return sign(
      schema: 'matchengine.order_command.v3',
      payload: payload.toJson(),
      mode: mode,
    );
  }

  Future<SignedAuthorizationEnvelope> signRevocation(
    TradingSessionRevocationPayload payload,
  ) => sign(
    schema: '2finance.trading_session_revocation.v1',
    payload: payload.toJson(),
    mode: OrderAuthorizationMode.wallet,
  );
}

Future<bool> verifyCanonicalEnvelope(SignedAuthorizationEnvelope envelope) {
  return Ed25519().verify(
    canonicalSigningBytesV1(envelope.payload),
    signature: Signature(
      KeyManager.hexToBytes(envelope.signature),
      publicKey: SimplePublicKey(
        KeyManager.hexToBytes(envelope.publicKey),
        type: KeyPairType.ed25519,
      ),
    ),
  );
}

void _requireText(String value, String field) {
  if (value.trim().isEmpty) {
    throw FormatException('$field is required');
  }
}
