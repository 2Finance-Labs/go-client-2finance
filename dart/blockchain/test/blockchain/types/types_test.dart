import 'package:test/test.dart';
import 'package:two_finance_blockchain/blockchain/types/types.dart';
import 'package:two_finance_blockchain/blockchain/utils/marshal.dart';
import 'dart:convert';

void main() {
  group('StateType', () {
    test('toJson omits null fields', () {
      final s = StateType();
      final json = s.toJson();

      expect(json.containsKey('type'), isFalse);
      expect(json.containsKey('object'), isFalse);
      expect(json, isEmpty);
    });

    test('toJson includes non-null fields', () {
      final s = StateType(
        type: 'balance_v2',
        object: {'address': 'abc', 'amount': '10'},
      );

      final json = s.toJson();
      expect(json['type'], 'balance_v2');
      expect(json['object'], {'address': 'abc', 'amount': '10'});
    });

    test('fromJson parses fields (including nulls)', () {
      final s1 = StateType.fromJson({
        'type': 'x',
        'object': {'a': 1},
      });
      expect(s1.type, 'x');
      expect(s1.object, {'a': 1});

      final s2 = StateType.fromJson({'type': null, 'object': null});
      expect(s2.type, isNull);
      expect(s2.object, isNull);
    });

    test('roundtrip fromJson -> toJson', () {
      final input = {
        'type': 'token_v2',
        'object': {'symbol': 'ADI', 'decimals': 18},
      };

      final s = StateType.fromJson(input);
      expect(s.toJson(), input);
    });
  });

  group('ContractOutput', () {
    test('toJson omits null fields', () {
      final out = ContractOutput();
      final json = out.toJson();

      expect(json.containsKey('states'), isFalse);
      expect(json.containsKey('logs'), isFalse);
      expect(json.containsKey('delegated_call'), isFalse);
      expect(json, isEmpty);
    });

    test('fromJson handles missing keys and null lists', () {
      final out1 = ContractOutput.fromJson({});
      expect(out1.states, isNull);
      expect(out1.logs, isNull);
      expect(out1.delegatedCall, isNull);

      final out2 = ContractOutput.fromJson({
        'states': null,
        'logs': null,
        'delegated_call': null,
      });
      expect(out2.states, isNull);
      expect(out2.logs, isNull);
      expect(out2.delegatedCall, isNull);
    });

    test('parses states list', () {
      final out = ContractOutput.fromJson({
        'states': [
          {
            'type': 'balance_v2',
            'object': {'address': 'a', 'amount': '1'},
          },
          {
            'type': 'mint_v2',
            'object': {'token': 'ADI', 'supply': '100'},
          },
        ],
      });

      expect(out.states, isNotNull);
      expect(out.states!.length, 2);
      expect(out.states![0].type, 'balance_v2');
      expect(out.states![0].object, {'address': 'a', 'amount': '1'});
      expect(out.states![1].type, 'mint_v2');
      expect(out.states![1].object, {'token': 'ADI', 'supply': '100'});
    });

    test('parses delegated_call recursively', () {
      final out = ContractOutput.fromJson({
        'delegated_call': [
          {
            'states': [
              {
                'type': 'x',
                'object': {'k': 'v'},
              },
            ],
          },
          {
            'delegated_call': [
              {
                'states': [
                  {
                    'type': 'y',
                    'object': {'n': 1},
                  },
                ],
              },
            ],
          },
        ],
      });

      expect(out.delegatedCall, isNotNull);
      expect(out.delegatedCall!.length, 2);

      // first delegated call has states
      final c1 = out.delegatedCall![0];
      expect(c1.states, isNotNull);
      expect(c1.states!.length, 1);
      expect(c1.states![0].type, 'x');
      expect(c1.states![0].object, {'k': 'v'});

      // second delegated call has nested delegated_call
      final c2 = out.delegatedCall![1];
      expect(c2.delegatedCall, isNotNull);
      expect(c2.delegatedCall!.length, 1);
      expect(c2.delegatedCall![0].states, isNotNull);
      expect(c2.delegatedCall![0].states![0].type, 'y');
      expect(c2.delegatedCall![0].states![0].object, {'n': 1});
    });

    test('toJson serializes states and delegated_call', () {
      final out = ContractOutput(
        states: [
          StateType(type: 'a', object: {'x': 1}),
        ],
        delegatedCall: [
          ContractOutput(
            states: [
              StateType(type: 'b', object: {'y': 2}),
            ],
          ),
        ],
      );

      final json = out.toJson();
      expect(json, {
        'states': [
          {
            'type': 'a',
            'object': {'x': 1},
          },
        ],
        'delegated_call': [
          {
            'states': [
              {
                'type': 'b',
                'object': {'y': 2},
              },
            ],
          },
        ],
      });
    });

    test('roundtrip toJson -> fromJson -> toJson (without logs)', () {
      final input = {
        'states': [
          {
            'type': 'token',
            'object': {'symbol': 'ADI'},
          },
        ],
        'delegated_call': [
          {
            'states': [
              {
                'type': 'balance',
                'object': {'address': 'a', 'amount': '10'},
              },
            ],
          },
        ],
      };

      final out = ContractOutput.fromJson(input);
      expect(out.toJson(), input);
    });

    test('logs: parses and serializes minimal log objects', () {
      final eventMap = {'name': 'TestEvent', 'value': 123};
      final eventBase64 = base64Encode(utf8.encode(jsonEncode(eventMap)));

      final input = {
        'logs': [
          {
            'log_type': 'event',
            'log_index': 1,
            'transaction_hash': 'a' * 64,
            'event': eventBase64,
            'contract_version': 'walletV2',
            'contract_address': DEPLOY_CONTRACT_ADDRESS,
          },
        ],
      };

      final out = ContractOutput.fromJson(input);

      expect(out.logs, isNotNull);
      expect(out.logs!.length, 1);

      final log = out.logs!.first;
      expect(log.logType, 'event');
      expect(log.logIndex, 1);
      expect(log.transactionHash, 'a' * 64);
      expect(log.contractVersion, 'walletV2');
      expect(log.contractAddress, DEPLOY_CONTRACT_ADDRESS);
      expect(log.event, eventBase64);
      final decoded = unmarshalEvent<Map<String, dynamic>>(log.event, (m) => m);
      expect(decoded['name'], equals('TestEvent'));
      expect(decoded['value'], equals(123));
      final json = out.toJson();
      expect(json, input);
    });
  });
}
