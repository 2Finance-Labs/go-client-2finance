import 'package:test/test.dart';
import 'package:two_finance_blockchain/blockchain/log/log.dart';

void main() {
  final validTxHash = 'a' * 64;
  const validContractAddress = '0xcontract';
  const validContractVersion = 'walletV2';

  group('Log constructor & getters', () {
    test('creates log and exposes getters correctly', () {
      final log = Log(
        logType: 'event',
        logIndex: 1,
        transactionHash: validTxHash,
        event: 'eyJrZXkiOiAidmFsdWUifQ==', // base64 encoded '{"key": "value"}'
        contractVersion: validContractVersion,
        contractAddress: validContractAddress,
      );

      expect(log.getLogType(), equals('event'));
      expect(log.getLogIndex(), equals(1));
      expect(log.getTransactionHash(), equals(validTxHash));
      expect(log.getEvent(), equals('eyJrZXkiOiAidmFsdWUifQ=='));
      expect(log.getContractVersion(), equals(validContractVersion));
      expect(log.getContractAddress(), equals(validContractAddress));
    });
  });

  group('Log.fromJson / toJson', () {
    test('fromJson parses all fields correctly', () {
      final json = {
        'log_type': 'event',
        'log_index': 2,
        'transaction_hash': validTxHash,
        'event': 'eyJhbW91bnQiOiAxMH0=', // base64 encoded '{"amount": 10}'
        'contract_version': validContractVersion,
        'contract_address': validContractAddress,
      };

      final log = Log.fromJson(json);

      expect(log.logType, equals('event'));
      expect(log.logIndex, equals(2));
      expect(log.transactionHash, equals(validTxHash));
      expect(log.event, equals('eyJhbW91bnQiOiAxMH0='));
      expect(log.contractVersion, equals(validContractVersion));
      expect(log.contractAddress, equals(validContractAddress));
    });

    test('fromJson throws when event missing', () {
      final json = {
        'log_type': 'event',
        'log_index': 2,
        'transaction_hash': validTxHash,
        // no 'event'
        'contract_version': validContractVersion,
        'contract_address': validContractAddress,
      };

      expect(
        () => Log.fromJson(json),
        throwsA(
          predicate((e) => e is Exception && e.toString().contains('event')),
        ),
      );
    });

    test('toJson serializes all fields correctly', () {
      final log = Log(
        logType: 'event',
        logIndex: 3,
        transactionHash: validTxHash,
        event: 'eyJ4IjogMX0=', // base64 encoded '{"x": 1}'
        contractVersion: validContractVersion,
        contractAddress: validContractAddress,
      );

      final json = log.toJson();

      expect(
        json,
        equals({
          'log_type': 'event',
          'log_index': 3,
          'transaction_hash': validTxHash,
          'event': 'eyJ4IjogMX0=', // base64 encoded '{"x": 1}'
          'contract_version': validContractVersion,
          'contract_address': validContractAddress,
        }),
      );
    });

    test('roundtrip toJson -> fromJson -> toJson', () {
      final original = Log(
        logType: 'event',
        logIndex: 4,
        transactionHash: validTxHash,
        event: 'eyJmb28iOiAiYmFyIn0=', // base64 encoded '{"foo": "bar"}'
        contractVersion: validContractVersion,
        contractAddress: validContractAddress,
      );

      final json = original.toJson();
      final parsed = Log.fromJson(json);

      expect(parsed.toJson(), equals(json));
    });
  });

  group('Log.validateLog', () {
    Log validLog() => Log(
      logType: 'event',
      logIndex: 1,
      transactionHash: validTxHash,
      event: 'eyJvayI6IHRydWV9', // base64 encoded '{"ok": true}'
      contractVersion: validContractVersion,
      contractAddress: validContractAddress,
    );

    test('does not throw for valid log', () {
      final log = validLog();
      expect(() => log.validateLog(), returnsNormally);
    });

    test('throws if logType is empty', () {
      final log = Log(
        logType: '',
        logIndex: 1,
        transactionHash: validTxHash,
        event: 'eyJvayI6IHRydWV9', // base64 encoded '{"ok": true}'
        contractVersion: validContractVersion,
        contractAddress: validContractAddress,
      );
      expect(() => log.validateLog(), throwsException);
    });

    test('throws if logIndex is zero', () {
      final log = Log(
        logType: 'event',
        logIndex: 0,
        transactionHash: validTxHash,
        event: 'eyJvayI6IHRydWV9', // base64 encoded '{"ok": true}'
        contractVersion: validContractVersion,
        contractAddress: validContractAddress,
      );
      expect(() => log.validateLog(), throwsException);
    });

    test('throws if transactionHash length is invalid', () {
      final log = Log(
        logType: 'event',
        logIndex: 1,
        transactionHash: 'abc',
        event: "",
        contractVersion: validContractVersion,
        contractAddress: validContractAddress,
      );
      expect(() => log.validateLog(), throwsException);
    });

    test('throws if event is empty', () {
      final log = Log(
        logType: 'event',
        logIndex: 1,
        transactionHash: validTxHash,
        event: "",
        contractVersion: validContractVersion,
        contractAddress: validContractAddress,
      );
      expect(() => log.validateLog(), throwsException);
    });

    test('throws if contractAddress is empty', () {
      final log = Log(
        logType: 'event',
        logIndex: 1,
        transactionHash: validTxHash,
        event: "",
        contractVersion: validContractVersion,
        contractAddress: '',
      );
      expect(() => log.validateLog(), throwsException);
    });

    test('throws if contractVersion is empty', () {
      final log = Log(
        logType: 'event',
        logIndex: 1,
        transactionHash: validTxHash,
        event: "",
        contractVersion: '',
        contractAddress: validContractAddress,
      );
      expect(() => log.validateLog(), throwsException);
    });
  });

  group('newLog factory', () {
    test('creates a valid log', () {
      final log = newLog(
        logType: 'event',
        logIndex: 10,
        transactionHash: validTxHash,
        event:
            'eyJoZWxsbyI6ICJ3b3JsZCJ9', // base64 encoded '{"hello": "world"}'
        contractVersion: validContractVersion,
        contractAddress: validContractAddress,
      );

      expect(log.logType, equals('event'));
      expect(log.logIndex, equals(10));
      expect(log.transactionHash, equals(validTxHash));
      expect(log.event, equals('eyJoZWxsbyI6ICJ3b3JsZCJ9'));
      expect(log.contractVersion, equals(validContractVersion));
      expect(log.contractAddress, equals(validContractAddress));

      expect(() => log.validateLog(), returnsNormally);
    });
  });
}
