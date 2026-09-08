import 'package:test/test.dart';
import 'package:two_finance_blockchain/trading_authorization.dart';

void main() {
  test('canonical JSON and domain preamble are deterministic', () {
    final bytes = canonicalSigningBytesV1({
      'z': 2,
      'domain': orderAddDomainV1,
      'nested': {'b': true, 'a': 'x'},
    });
    expect(
      String.fromCharCodes(bytes),
      '$signedMessagePreambleV1\n$orderAddDomainV1\n'
      '{"domain":"$orderAddDomainV1","nested":{"a":"x","b":true},"z":2}',
    );
  });

  test('session key signs a valid canonical order', () async {
    final signer = await Ed25519CanonicalSigner.fromPrivateKeyHex(
      '000102030405060708090a0b0c0d0e0f'
      '101112131415161718191a1b1c1d1e1f',
    );
    final payload = OrderCommandPayload(
      chainId: '2finance-testnet',
      account:
          'aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa',
      engineId: 'engine-btc',
      symbol: 'BTC/USDT',
      symbolId: 1,
      routeEpoch: 7,
      operation: 'ADD',
      authorizationMode: OrderAuthorizationMode.session,
      side: 'BUY',
      type: 'LIMIT',
      amount: '0.0100',
      price: '64000.00',
      timeInForce: 'GTC',
      clientOrderId: 'order-1',
      idempotencyKey: 'order-1-add',
      sessionId: 'session-1',
      nonce: 42,
      expiresAtMs: 1788725000000,
    );
    final envelope = await signer.signOrder(
      payload,
      mode: OrderAuthorizationMode.session,
    );

    expect(payload.amount, '0.01');
    expect(payload.price, '64000');
    expect(
      signer.publicKey,
      '03a107bff3ce10be1d70dd18e74bc09967e4d6309ba50d5f1ddc8664125531b8',
    );
    expect(
      envelope.signature,
      'e35ced593f24ef68142a0e2313470616d1ae7ed702ca68e998629646ba73a1185'
      'cd5edf77bffd584b5c3aa8779e5718f0934f6db0ece6ee62998ef1de6db5208',
    );
    expect(envelope.payload['domain'], orderAddDomainV1);
    expect(await verifyCanonicalEnvelope(envelope), isTrue);
  });

  test('cancel-all is domain-separated and has a strict empty order id', () {
    final payload = OrderCommandPayload(
      chainId: '2finance-testnet',
      account:
          'aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa',
      engineId: 'engine-btc',
      symbol: 'BTC/USDT',
      symbolId: 1,
      routeEpoch: 7,
      operation: 'CANCEL_ALL',
      authorizationMode: OrderAuthorizationMode.session,
      side: 'NONE',
      type: 'NONE',
      amount: '0',
      price: '0',
      timeInForce: 'NONE',
      clientOrderId: 'cancel-all-1',
      idempotencyKey: 'cancel-all-1',
      sessionId: 'session-1',
      nonce: '43',
      expiresAtMs: 1788725000000,
    );

    expect(payload.domain, orderCancelAllDomainV1);
    expect(payload.requiredAction, 'order:cancel_all');
  });
}
