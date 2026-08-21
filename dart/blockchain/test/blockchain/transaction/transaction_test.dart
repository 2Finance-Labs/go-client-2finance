import 'dart:convert';

import 'package:test/test.dart';
import '../../helpers/helpers.dart';

import 'package:two_finance_blockchain/blockchain/transaction/transaction.dart'; // <-- ajuste o caminho
import 'package:two_finance_blockchain/blockchain/types/types.dart';
import 'package:two_finance_blockchain/blockchain/keys/keys.dart';
import 'package:two_finance_blockchain/blockchain/utils/json.dart';

JsonMessage _validData() {
  // mapToJsonRawMessage espera Map<String,dynamic>
  return {'amount': '10', 'memo': 'hello'};
}

Future<Transaction> _buildSignedValidTx({
  int chainId = 1,
  String? from,
  String? to,
  String method = 'transfer',
  int version = 1,
  String? uuid7,
}) async {
  final pair = await validKeyPair();

  final f = from ?? pair.publicKey;
  final t = to ?? await validPublicKeyHex();
  final id =
      uuid7 ??
      // use um UUIDv7 válido no seu projeto. Se você tiver um gerador, use aqui.
      // Caso não tenha, substitua por um UUIDv7 real que seu validateUUID7 aceite.
      '018f4c3e-9c2a-7b8e-8f14-6c4c2a1d9a33';

  final tx = Transaction.create(
    chainID: chainId,
    from: f,
    to: t,
    method: method,
    data: _validData(),
    version: version,
    uuid7: id,
  );

  // assina e popula hash/signature
  return signTransaction(pair.privateKey, tx);
}

void main() {
  group('Transaction.calculateHash()', () {
    test('returns 64-hex string and is deterministic', () async {
      final tx = Transaction.create(
        chainID: 1,
        from: await validPublicKeyHex(),
        to: await validPublicKeyHex(),
        method: 'transfer',
        data: _validData(),
        version: 1,
        uuid7: '018f4c3e-9c2a-7b8e-8f14-6c4c2a1d9a33',
      );

      final h1 = await tx.calculateHash();
      final h2 = await tx.calculateHash();

      expect(h1, equals(h2));
      expect(h1.length, equals(64));
      // hex lowercase/uppercase depende do seu bytesToHex; aqui só valida hex.
      expect(RegExp(r'^[0-9a-fA-F]{64}$').hasMatch(h1), isTrue);
    });

    test('ignores current hash/signature fields when hashing', () async {
      final tx = Transaction.create(
        chainID: 1,
        from: await validPublicKeyHex(),
        to: await validPublicKeyHex(),
        method: 'transfer',
        data: _validData(),
        version: 1,
        uuid7: '018f4c3e-9c2a-7b8e-8f14-6c4c2a1d9a33',
      );

      final base = await tx.calculateHash();

      tx.hash = repeatHex('c', 64);
      tx.signature = repeatHex('d', 128);

      final again = await tx.calculateHash();
      expect(again, equals(base));
    });
  });

  group('Transaction.validateHash()', () {
    test('passes when hash matches computed', () async {
      final tx = await _buildSignedValidTx();
      await tx.validateHash(); // should not throw
    });

    test('throws when hash mismatches computed', () async {
      final tx = await _buildSignedValidTx();
      tx.hash = repeatHex('e', 64);

      expect(() async => tx.validateHash(), throwsA(isA<Exception>()));
    });
  });

  group('signTransaction()', () {
    test('sets hash (64 hex) and signature (128 hex)', () async {
      final pair = await validKeyPair();
      final tx = Transaction.create(
        chainID: 1,
        from: await validPublicKeyHex(),
        to: await validPublicKeyHex(),
        method: 'transfer',
        data: _validData(),
        version: 1,
        uuid7: '018f4c3e-9c2a-7b8e-8f14-6c4c2a1d9a33',
      );

      final signed = await signTransaction(pair.privateKey, tx);

      expect(signed.hash.length, equals(64));
      expect(RegExp(r'^[0-9a-fA-F]{64}$').hasMatch(signed.hash), isTrue);

      expect(signed.signature.length, equals(128));
      expect(RegExp(r'^[0-9a-fA-F]{128}$').hasMatch(signed.signature), isTrue);
    });

    test('throws if private key < 32 bytes (64 hex chars)', () async {
      final tx = Transaction.create(
        chainID: 1,
        from: await validPublicKeyHex(),
        to: await validPublicKeyHex(),
        method: 'transfer',
        data: _validData(),
        version: 1,
        uuid7: '018f4c3e-9c2a-7b8e-8f14-6c4c2a1d9a33',
      );

      expect(() async => signTransaction('aa', tx), throwsA(isA<Exception>()));
    });
  });

  group('Transaction.validateTransaction()', () {
    test('passes for a fully valid, signed transaction', () async {
      final tx = await _buildSignedValidTx();
      await tx.validateTransaction(); // should not throw
    });

    test('rejects chainID <= 0', () async {
      final tx = await _buildSignedValidTx(chainId: 0);
      expect(() async => tx.validateTransaction(), throwsA(isA<Exception>()));
    });

    test('rejects unsupported chainID > 2', () async {
      final tx = await _buildSignedValidTx(chainId: 3);
      expect(() async => tx.validateTransaction(), throwsA(isA<Exception>()));
    });

    test('rejects empty from', () async {
      final tx = await _buildSignedValidTx(from: '');
      expect(() async => tx.validateTransaction(), throwsA(isA<Exception>()));
    });

    test('rejects empty to', () async {
      final tx = await _buildSignedValidTx(to: '');
      expect(() async => tx.validateTransaction(), throwsA(isA<Exception>()));
    });

    test('rejects from == to', () async {
      final same = await validPublicKeyHex();
      final tx = await _buildSignedValidTx(from: same, to: same);
      expect(() async => tx.validateTransaction(), throwsA(isA<Exception>()));
    });

    test('rejects empty method', () async {
      final tx = await _buildSignedValidTx(method: '');
      expect(() async => tx.validateTransaction(), throwsA(isA<Exception>()));
    });

    test('rejects version == 0', () async {
      final tx = await _buildSignedValidTx(version: 0);
      expect(() async => tx.validateTransaction(), throwsA(isA<Exception>()));
    });

    test('rejects wrong hash length', () async {
      final tx = await _buildSignedValidTx();
      tx.hash = 'abc'; // not 64
      expect(() async => tx.validateTransaction(), throwsA(isA<Exception>()));
    });

    test('rejects empty signature', () async {
      final tx = await _buildSignedValidTx();
      tx.signature = '';
      expect(() async => tx.validateTransaction(), throwsA(isA<Exception>()));
    });

    test('rejects wrong signature length', () async {
      final tx = await _buildSignedValidTx();
      tx.signature = repeatHex('a', 127);
      expect(() async => tx.validateTransaction(), throwsA(isA<Exception>()));
    });

    test('rejects invalid hash (computed != provided)', () async {
      final tx = await _buildSignedValidTx();
      tx.hash = repeatHex('f', 64);
      expect(() async => tx.validateTransaction(), throwsA(isA<Exception>()));
    });

    test('allows DEPLOY_CONTRACT_ADDRESS without pubkey validation', () async {
      final tx = await _buildSignedValidTx(to: DEPLOY_CONTRACT_ADDRESS);
      await tx
          .validateTransaction(); // should not throw if deploy address bypass works
    });
  });

  group('Transaction.fromJson()', () {
    test('throws when data is null', () {
      final json = <String, dynamic>{
        'chain_id': 1,
        'from': validPublicKeyHex(),
        'to': validPublicKeyHex(),
        'method': 'transfer',
        'data': null,
        'version': 1,
        'uuid7': '018f4c3e-9c2a-7b8e-8f14-6c4c2a1d9a33',
        'hash': repeatHex('a', 64),
        'signature': repeatHex('b', 128),
      };

      expect(() => Transaction.fromJson(json), throwsA(isA<Exception>()));
    });

    test('roundtrip toJson -> fromJson preserves core fields', () async {
      final tx = await _buildSignedValidTx();
      final json = tx.toJson();

      final parsed = Transaction.fromJson(json);

      expect(parsed.chainID, tx.chainID);
      expect(parsed.from, tx.from);
      expect(parsed.to, tx.to);
      expect(parsed.method, tx.method);
      expect(parsed.version, tx.version);
      expect(parsed.uuid7, tx.uuid7);
      expect(parsed.hash, tx.hash);
      expect(parsed.signature, tx.signature);

      // data: garante que é serializável e equivalente
      expect(jsonEncode(parsed.data), equals(jsonEncode(tx.data)));
    });
  });

  group('Transaction.calculateHash() JCS invariants', () {
    /// If your calculateHash is truly JCS, changing map insertion order in the
    /// *semantic* data should not change the hash.
    test('hash is invariant to data map key insertion order', () async {
      final pair = await validKeyPair();

      // Same semantic content, different insertion order
      final JsonMessage dataA = {'amount': '10', 'memo': 'hello'};

      final JsonMessage dataB = {'memo': 'hello', 'amount': '10'};

      final from = pair.publicKey;
      final to = await validPublicKeyHex(); // different key from `from`

      final tx1 = Transaction.create(
        chainID: 1,
        from: from,
        to: to,
        method: 'transfer',
        data: dataA,
        version: 1,
        uuid7: '018f4c3e-9c2a-7b8e-8f14-6c4c2a1d9a33',
      );

      final tx2 = Transaction.create(
        chainID: 1,
        from: from,
        to: to,
        method: 'transfer',
        data: dataB,
        version: 1,
        uuid7: '018f4c3e-9c2a-7b8e-8f14-6c4c2a1d9a33',
      );

      final h1 = await tx1.calculateHash();
      final h2 = await tx2.calculateHash();

      expect(h1, equals(h2));
    });

    test('hash is always 64 hex chars', () async {
      final tx = await _buildSignedValidTx();
      final h = await tx.calculateHash();
      expect(h.length, 64);
      expect(RegExp(r'^[0-9a-fA-F]{64}$').hasMatch(h), isTrue);
    });
  });
}
