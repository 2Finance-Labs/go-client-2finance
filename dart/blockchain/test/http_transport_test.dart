import 'dart:convert';

import 'package:http/http.dart' as http;
import 'package:http/testing.dart';
import 'package:test/test.dart';
import 'package:two_finance_blockchain/infra/http/http_transport.dart';

void main() {
  test(
    'HttpFinanceNetworkTransport posts to virtual-machine endpoint',
    () async {
      late Uri receivedUri;
      late Map<String, String> receivedHeaders;
      late Map<String, dynamic> receivedBody;

      final transport = HttpFinanceNetworkTransport(
        baseUrl: 'http://2finance-network:9095/',
        tokenProvider: () async => 'test-access-token',
        httpClient: MockClient((request) async {
          receivedUri = request.url;
          receivedHeaders = request.headers;
          receivedBody = jsonDecode(request.body) as Map<String, dynamic>;
          return http.Response(
            jsonEncode({
              'code': 200,
              'msg': 'Successfully',
              'data': {'states': <dynamic>[], 'logs': <dynamic>[]},
            }),
            200,
            headers: {'content-type': 'application/json'},
          );
        }),
      );

      final data = await transport.sendRequest('get_state', {
        'to': 'wallet',
        'method': 'get',
      }, 'reply-id');

      expect(
        receivedUri.toString(),
        'http://2finance-network:9095/v2/2finance-network/query',
      );
      expect(receivedBody['method'], 'get_state');
      expect(receivedBody['params'], {'to': 'wallet', 'method': 'get'});
      expect(receivedHeaders['Authorization'], 'Bearer test-access-token');
      expect(data, {'states': <dynamic>[], 'logs': <dynamic>[]});
    },
  );

  test('HttpFinanceNetworkTransport works without a token provider', () async {
    late Map<String, String> receivedHeaders;
    final transport = HttpFinanceNetworkTransport(
      baseUrl: 'http://2finance-network:9095/',
      httpClient: MockClient((request) async {
        receivedHeaders = request.headers;
        return http.Response(
          jsonEncode({
            'code': 200,
            'msg': 'Successfully',
            'data': {'ok': true},
          }),
          200,
        );
      }),
    );

    final data = await transport.sendRequest(
      'get_state',
      <String, dynamic>{},
      'reply-id',
    );

    expect(receivedHeaders.containsKey('Authorization'), isFalse);
    expect(data, {'ok': true});
  });

  test('HttpFinanceNetworkTransport ignores empty bearer tokens', () async {
    late Map<String, String> receivedHeaders;
    final transport = HttpFinanceNetworkTransport(
      baseUrl: 'http://2finance-network:9095/',
      tokenProvider: () async => '   ',
      httpClient: MockClient((request) async {
        receivedHeaders = request.headers;
        return http.Response(
          jsonEncode({
            'code': 200,
            'msg': 'Successfully',
            'data': {'ok': true},
          }),
          200,
        );
      }),
    );

    await transport.sendRequest('get_state', <String, dynamic>{}, 'reply-id');

    expect(receivedHeaders.containsKey('Authorization'), isFalse);
  });

  test(
    'HttpFinanceNetworkTransport keeps preformatted bearer tokens',
    () async {
      late Map<String, String> receivedHeaders;
      final transport = HttpFinanceNetworkTransport(
        baseUrl: 'http://2finance-network:9095/',
        tokenProvider: () async => 'Bearer already-prefixed',
        httpClient: MockClient((request) async {
          receivedHeaders = request.headers;
          return http.Response(
            jsonEncode({
              'code': 200,
              'msg': 'Successfully',
              'data': {'ok': true},
            }),
            200,
          );
        }),
      );

      await transport.sendRequest('get_state', <String, dynamic>{}, 'reply-id');

      expect(receivedHeaders['Authorization'], 'Bearer already-prefixed');
    },
  );

  test(
    'HttpFinanceNetworkTransport does not include bearer token in errors',
    () {
      final transport = HttpFinanceNetworkTransport(
        baseUrl: 'http://2finance-network:9095/',
        tokenProvider: () async => 'sensitive-token-value',
        httpClient: MockClient((request) async {
          return http.Response('upstream denied', 503);
        }),
      );

      expect(
        () =>
            transport.sendRequest('get_state', <String, dynamic>{}, 'reply-id'),
        throwsA(
          isA<Exception>().having(
            (error) => error.toString(),
            'message',
            allOf(
              contains('2finance-network HTTP 503'),
              isNot(contains('sensitive-token-value')),
            ),
          ),
        ),
      );
    },
  );

  test('HttpFinanceNetworkTransport submits and queries protocol v2', () async {
    final hash = 'a' * 64;
    final signature = 'b' * 128;
    final requests = <http.Request>[];
    final transport = HttpFinanceNetworkTransport(
      baseUrl: 'http://2finance-network:9095/',
      httpClient: MockClient((request) async {
        requests.add(request);
        final signed = <String, dynamic>{
          'chain_id': 2,
          'from': '1' * 64,
          'to': '0' * 64,
          'method': 'deploy_contract',
          'data': {
            'runtime': 'native_go',
            'version': 1,
            'payload': {'contract_version': 'walletV2'},
          },
          'version': 1,
          'uuid7': '019fba91-223d-7c2e-9825-65eb7c3607e0',
          'hash': hash,
          'signature': signature,
        };
        final status = <String, dynamic>{
          'protocol_version': 2,
          'chain_id': '2finance-dev-v2',
          'transaction_hash': hash,
          'signed_transaction': signed,
          'commit': {
            'sequence': 1,
            'admission_height': 2,
            'status': 'committed',
            'committed_at': '2026-08-01T01:02:03Z',
          },
          'artifact': {
            'runtime': 'native_go',
            'runtime_version': 1,
            'execution_spec_hash': '1' * 64,
            'receipt_hash': '2' * 64,
            'artifact_hash': '3' * 64,
            'log_root': '4' * 64,
            'write_set_root': '5' * 64,
            'state_root_before': '6' * 64,
            'state_root_after': '7' * 64,
            'timestamp_hash': '8' * 64,
          },
        };
        return http.Response(
          jsonEncode({
            'code': 200,
            'msg': 'Successfully',
            'data': request.method == 'POST'
                ? {
                    'execution_output': <String, dynamic>{},
                    'transaction': status,
                  }
                : status,
          }),
          200,
        );
      }),
    );
    final body = <String, dynamic>{
      'chain_id': 2,
      'hash': hash,
      'signature': signature,
    };

    final submitted = await transport.submitSignedTransactionV2(body);
    final queried = await transport.transactionFinalityV2(hash.toUpperCase());

    expect(requests, hasLength(2));
    expect(requests[0].method, 'POST');
    expect(
      requests[0].url.toString(),
      'http://2finance-network:9095/v2/2finance-network/transactions',
    );
    expect(jsonDecode(requests[0].body), body);
    expect(requests[1].method, 'GET');
    expect(
      requests[1].url.toString(),
      'http://2finance-network:9095/v2/2finance-network/transactions/$hash/finality',
    );
    expect((submitted['transaction'] as Map)['transaction_hash'], hash);
    expect(queried['signed_transaction'], isA<Map<String, dynamic>>());
  });

  test('protocol v2 finality rejects malformed hashes before HTTP', () async {
    final transport = HttpFinanceNetworkTransport(
      baseUrl: 'http://2finance-network:9095/',
      httpClient: MockClient((request) async {
        fail('HTTP must not be called');
      }),
    );
    await expectLater(
      transport.transactionFinalityV2('bad'),
      throwsArgumentError,
    );
  });

  test('HttpFinanceNetworkTransport reads canonical v2 logs', () async {
    late http.Request received;
    final transport = HttpFinanceNetworkTransport(
      baseUrl: 'http://2finance-network:9095/',
      httpClient: MockClient((request) async {
        received = request;
        return http.Response(
          jsonEncode({
            'code': 200,
            'data': {
              'page': 1,
              'limit': 100,
              'logs': [
                {'runtime': 'evm', 'contract_version': 'evmV2'},
              ],
            },
          }),
          200,
        );
      }),
    );

    final page = await transport.executionLogsV2({
      'runtime': 'evm',
      'transaction_hash': '',
    });

    expect(received.method, 'GET');
    expect(
      received.url.toString(),
      'http://2finance-network:9095/v2/2finance-network/logs?runtime=evm',
    );
    expect((page['logs'] as List).single['contract_version'], 'evmV2');
  });

  test(
    'protocol v2 retries transient conflict with identical signed body',
    () async {
      final hash = 'a' * 64;
      final body = <String, dynamic>{
        'chain_id': 2,
        'uuid7': '019fba91-223d-7c2e-9825-65eb7c3607e0',
        'hash': hash,
        'signature': 'b' * 128,
      };
      final requestBodies = <String>[];
      final transport = HttpFinanceNetworkTransport(
        baseUrl: 'http://2finance-network:9095',
        httpClient: MockClient((request) async {
          requestBodies.add(request.body);
          if (requestBodies.length == 1) {
            return http.Response(
              jsonEncode({
                'code': 422,
                'data': "can't serialize access for this transaction",
              }),
              422,
            );
          }
          return http.Response(
            jsonEncode({
              'code': 200,
              'data': {
                'execution_output': <String, dynamic>{},
                'transaction': {'transaction_hash': hash},
              },
            }),
            200,
          );
        }),
      );

      final submitted = await transport.submitSignedTransactionV2(body);

      expect(requestBodies, hasLength(2));
      expect(requestBodies[1], requestBodies[0]);
      expect((submitted['transaction'] as Map)['transaction_hash'], hash);
    },
  );
}
