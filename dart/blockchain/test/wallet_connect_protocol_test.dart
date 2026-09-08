import 'package:test/test.dart';
import 'package:two_finance_blockchain/trading_authorization.dart';
import 'package:two_finance_blockchain/wallet_connect_protocol.dart';

void main() {
  final now = DateTime.utc(2026, 9, 6, 12);
  const account =
      'aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa';
  const sessionKey =
      'bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb';

  TradingSessionConnectRequest request() => TradingSessionConnectRequest(
    challengeId: 'challenge-1',
    challengeExpiresAt: now.add(const Duration(minutes: 5)),
    responseUri: Uri.parse(
      'https://exchange.example/api/v2/exchange/session/challenges/challenge-1/responses',
    ),
    payload: TradingSessionAuthorizationPayload(
      chainId: '2finance-testnet',
      account: account,
      sessionId: 'session-1',
      sessionPublicKey: sessionKey,
      allowedSymbols: const ['BTC/USDT'],
      allowedActions: const ['order:cancel', 'order:submit'],
      allowedOrderTypes: const ['LIMIT'],
      maxOrderAmount: '1',
      maxOrderNotional: '100',
      maxSessionNotional: '500',
      issuedAtMs: now.millisecondsSinceEpoch,
      expiresAtMs: now.add(const Duration(hours: 1)).millisecondsSinceEpoch,
      nonce: '1' * 64,
      origin: 'https://exchange.example',
    ),
  );

  test('round-trips QR, deep link and universal link requests', () {
    final source = request();
    for (final link in [source.toDeepLink(), source.toUniversalLink()]) {
      final parsed = TradingSessionConnectRequest.parse(
        link.toString(),
        trustedOrigins: const ['https://exchange.example'],
        now: now,
      );
      expect(parsed.challengeId, source.challengeId);
      expect(parsed.payload.account, account);
      expect(parsed.payload.allowedActions, ['order:cancel', 'order:submit']);
      expect(parsed.responseUri, source.responseUri);
    }
  });

  test('rejects phishing origin and response endpoint substitution', () {
    final json = request().toJson();
    expect(
      () => TradingSessionConnectRequest.fromJson(
        json,
        trustedOrigins: const ['https://attacker.example'],
        now: now,
      ),
      throwsFormatException,
    );

    final changed = Map<String, dynamic>.from(json)
      ..['response_uri'] =
          'https://attacker.example/api/v2/exchange/session/challenges/challenge-1/responses';
    expect(
      () => TradingSessionConnectRequest.fromJson(
        changed,
        trustedOrigins: const ['https://exchange.example'],
        now: now,
      ),
      throwsFormatException,
    );
  });

  test('rejects expired, query-carried and non-canonical requests', () {
    expect(
      () => TradingSessionConnectRequest.fromJson(
        request().toJson(),
        trustedOrigins: const ['https://exchange.example'],
        now: now.add(const Duration(minutes: 6)),
      ),
      throwsFormatException,
    );
    final queryLink = request().toUniversalLink().replace(
      queryParameters: {'request': request().encode()},
      fragment: '',
    );
    expect(
      () => TradingSessionConnectRequest.parse(
        queryLink.toString(),
        trustedOrigins: const ['https://exchange.example'],
        now: now,
      ),
      throwsFormatException,
    );
    final json = request().toJson()..['unexpected'] = true;
    expect(
      () => TradingSessionConnectRequest.fromJson(
        json,
        trustedOrigins: const ['https://exchange.example'],
        now: now,
      ),
      throwsFormatException,
    );
  });
}
