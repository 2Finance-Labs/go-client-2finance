import 'dart:convert';
import 'package:test/test.dart';
import 'package:two_finance_blockchain/blockchain/utils/marshal.dart';
import 'package:two_finance_blockchain/blockchain/contract/walletV2/models/wallet.dart';
import 'package:two_finance_blockchain/blockchain/contract/walletV2/domain/wallet.dart';
import 'package:two_finance_blockchain/blockchain/log/log.dart';

void main() {
  group('Marshal Tests', () {
    group('unmarshalState', () {
      test('Map<String, dynamic> -> WalletState (sucesso)', () {
        final obj = <String, dynamic>{
          'address': 'pub1',
          'public_key': 'pub2',
          'created_at': '2026-03-03T10:00:00.000Z',
          'updated_at': '2026-03-03T11:00:00.000Z',
        };

        final out = unmarshalState<WalletState>(obj, WalletState.fromJson);

        expect(out.address, equals('pub1'));
        expect(out.publicKey, equals('pub2'));
        expect(
          out.createdAt.toIso8601String(),
          equals('2026-03-03T10:00:00.000Z'),
        );
        expect(
          out.updatedAt.toIso8601String(),
          equals('2026-03-03T11:00:00.000Z'),
        );
      });

      test(
        'Map<dynamic, dynamic> -> WalletState (converte keys para String)',
        () {
          final Map<dynamic, dynamic> obj = {
            'address': 'pub1',
            'public_key': 'pub2',
            'created_at': '2026-03-03T12:00:00.000Z',
            'updated_at': '2026-03-03T13:00:00.000Z',
            1: 'ignored',
          };

          final out = unmarshalState<WalletState>(obj, WalletState.fromJson);

          expect(out.address, equals('pub1'));
          expect(out.publicKey, equals('pub2'));
          expect(
            out.createdAt.toIso8601String(),
            equals('2026-03-03T12:00:00.000Z'),
          );
          expect(
            out.updatedAt.toIso8601String(),
            equals('2026-03-03T13:00:00.000Z'),
          );
        },
      );

      test('JSON string -> WalletState (sucesso)', () {
        final obj =
            '{"address":"pub1","public_key":"pub2","created_at":"2026-03-03T14:00:00.000Z","updated_at":"2026-03-03T15:00:00.000Z"}';

        final out = unmarshalState<WalletState>(obj, WalletState.fromJson);

        expect(out.address, equals('pub1'));
        expect(out.publicKey, equals('pub2'));
        expect(
          out.createdAt.toIso8601String(),
          equals('2026-03-03T14:00:00.000Z'),
        );
        expect(
          out.updatedAt.toIso8601String(),
          equals('2026-03-03T15:00:00.000Z'),
        );
      });

      test('JSON string com espaços -> WalletState (sucesso)', () {
        final obj =
            '  {"address":"pub1","public_key":"pub2","created_at":"2026-03-03T16:00:00.000Z","updated_at":"2026-03-03T17:00:00.000Z"}  ';

        final out = unmarshalState<WalletState>(obj, WalletState.fromJson);

        expect(out.address, equals('pub1'));
        expect(out.publicKey, equals('pub2'));
        expect(
          out.createdAt.toIso8601String(),
          equals('2026-03-03T16:00:00.000Z'),
        );
        expect(
          out.updatedAt.toIso8601String(),
          equals('2026-03-03T17:00:00.000Z'),
        );
      });

      test('UTF-8 bytes (List<int>) -> WalletState (sucesso)', () {
        final bytes = utf8.encode(
          '{"address":"pub1","public_key":"pub2","created_at":"2026-03-03T18:00:00.000Z","updated_at":"2026-03-03T19:00:00.000Z"}',
        );

        final out = unmarshalState<WalletState>(bytes, WalletState.fromJson);

        expect(out.address, equals('pub1'));
        expect(out.publicKey, equals('pub2'));
        expect(
          out.createdAt.toIso8601String(),
          equals('2026-03-03T18:00:00.000Z'),
        );
        expect(
          out.updatedAt.toIso8601String(),
          equals('2026-03-03T19:00:00.000Z'),
        );
      });

      test('obj == null -> erro', () {
        expect(
          () => unmarshalState<WalletState>(null, WalletState.fromJson),
          throwsA(
            predicate(
              (e) =>
                  e is Exception &&
                  e.toString().contains('unmarshalState: object is null'),
            ),
          ),
        );
      });

      test('JSON string vazia -> erro', () {
        expect(
          () => unmarshalState<WalletState>('   ', WalletState.fromJson),
          throwsA(
            predicate(
              (e) =>
                  e is Exception &&
                  e.toString().contains('unmarshalState: empty JSON string'),
            ),
          ),
        );
      });

      test('JSON string nao-objeto (ex: lista) -> erro', () {
        expect(
          () => unmarshalState<WalletState>('[1,2,3]', WalletState.fromJson),
          throwsA(
            predicate(
              (e) =>
                  e is Exception &&
                  e.toString().contains('JSON did not decode to object'),
            ),
          ),
        );
      });

      test('bytes JSON nao-objeto (ex: numero) -> erro', () {
        final bytes = utf8.encode('123');

        expect(
          () => unmarshalState<WalletState>(bytes, WalletState.fromJson),
          throwsA(
            predicate(
              (e) =>
                  e is Exception &&
                  e.toString().contains('bytes JSON did not decode to object'),
            ),
          ),
        );
      });

      test('tipo nao suportado -> erro', () {
        expect(
          () => unmarshalState<WalletState>(42, WalletState.fromJson),
          throwsA(
            predicate(
              (e) =>
                  e is Exception &&
                  e.toString().contains('unsupported state object type'),
            ),
          ),
        );
      });
    });

    group('unmarshalLog', () {
      test('obj map -> Log (sucesso)', () {
        final obj = {
          'log_type': 'TOKEN_MINT',
          'log_index': 10,
          'transaction_hash': '0xabc',
          // event precisa ser String base64
          'event': base64Encode(utf8.encode(jsonEncode({'amount': '100'}))),
          'contract_version': 'walletV2',
          'contract_address': '0xtoken',
        };

        final out = unmarshalLog<Log>(obj, Log.fromJson);

        expect(out.logType, equals('TOKEN_MINT'));
        expect(out.logIndex, equals(10));
        expect(out.transactionHash, equals('0xabc'));
        expect(out.contractVersion, equals('walletV2'));
        expect(out.contractAddress, equals('0xtoken'));

        final ev = unmarshalEvent<Map<String, dynamic>>(
          out.event,
          (m) => m, // identity
        );

        expect(ev['amount'], equals('100'));
      });

      test('obj com chave nao-string -> Log (jsonEncode converte)', () {
        final obj = {
          1: 'ignored',
          'log_type': 'TRANSFER',
          'log_index': 11,
          'transaction_hash': '0xdef',
          'event': base64Encode(
            utf8.encode(jsonEncode({'from': '0x1', 'to': '0x2'})),
          ),
          'contract_version': 'v2',
          'contract_address': '0xcontract',
        };

        // Obs: jsonEncode vai converter a key 1 -> "1", mas isso nao afeta o parse do Log
        final out = unmarshalLog<Log>(obj, Log.fromJson);

        expect(out.logType, equals('TRANSFER'));
        expect(out.logIndex, equals(11));
        expect(out.transactionHash, equals('0xdef'));
        expect(out.contractVersion, equals('v2'));
        expect(out.contractAddress, equals('0xcontract'));

        final ev = unmarshalEvent<Map<String, dynamic>>(
          out.event,
          (m) => m, // identity
        );
        expect(ev['from'], equals('0x1'));
        expect(ev['to'], equals('0x2'));
      });

      test('obj que nao vira Map no decode (ex: lista) -> erro', () {
        expect(
          () => unmarshalLog<Log>([1, 2, 3], Log.fromJson),
          throwsA(
            predicate(
              (e) =>
                  e is Exception &&
                  e.toString().contains('marshal/unmarshal log'),
            ),
          ),
        );
      });

      test('obj nao-serializavel -> erro', () {
        final obj = Object(); // jsonEncode(Object()) geralmente falha
        expect(
          () => unmarshalLog<Log>(obj, Log.fromJson),
          throwsA(
            predicate(
              (e) =>
                  e is Exception &&
                  e.toString().contains('marshal/unmarshal log'),
            ),
          ),
        );
      });
    });

    group('unmarshalEvent', () {
      test('event base64 -> Wallet (sucesso)', () {
        final event = base64Encode(
          utf8.encode('{"address":"0xabc","public_key":"pub123"}'),
        );

        final out = unmarshalEvent<Wallet>(event, Wallet.fromJson);

        expect(out.address, equals('0xabc'));
        expect(out.publicKey, equals('pub123'));
      });

      test('event vazio -> erro', () {
        expect(
          () => unmarshalEvent<Wallet>('', Wallet.fromJson),
          throwsA(
            predicate(
              (e) => e is Exception && e.toString().contains('empty event'),
            ),
          ),
        );
      });

      test('event com JSON invalido -> erro', () {
        final event = base64Encode(utf8.encode('{"address":'));

        expect(
          () => unmarshalEvent<Wallet>(event, Wallet.fromJson),
          throwsA(
            predicate(
              (e) => e is Exception && e.toString().contains('unmarshal event'),
            ),
          ),
        );
      });

      test('event JSON nao-objeto (ex: lista) -> erro', () {
        final bytes = utf8.encode('[1,2,3]');

        expect(
          () => unmarshalEvent<Wallet>(base64Encode(bytes), Wallet.fromJson),
          throwsA(
            predicate(
              (e) => e is Exception && e.toString().contains('unmarshal event'),
            ),
          ),
        );
      });
    });
  });
}
