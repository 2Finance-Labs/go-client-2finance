import 'package:test/test.dart';
import 'package:uuid/uuid.dart';

import 'package:two_finance_blockchain/blockchain/utils/uuid.dart';

void main() {
  group('newUUID7', () {
    test('generates a valid UUIDv7 string', () {
      final id = newUUID7();

      expect(id, isNotEmpty);

      // parseable UUID
      expect(() => UuidValue.fromString(id), returnsNormally);

      // passes your validator
      expect(() => validateUUID7(id), returnsNormally);
      expect(() => validateUUID7Strict(id), returnsNormally);
    });

    test('generates different UUIDs (very likely)', () {
      final a = newUUID7();
      final b = newUUID7();

      expect(a, isNot(b));
    });
  });

  group('validateUUID7', () {
    test('accepts UUIDv7', () {
      final id = const Uuid().v7();
      expect(() => validateUUID7(id), returnsNormally);
    });

    test('rejects invalid format', () {
      expect(() => validateUUID7('not-a-uuid'), throwsA(isA<Exception>()));

      expect(() => validateUUID7(''), throwsA(isA<Exception>()));
    });

    test('rejects non-v7 UUID (e.g. v4)', () {
      final id = const Uuid().v4();

      expect(() => validateUUID7(id), throwsA(isA<Exception>()));
    });

    test('rejects wrong variant (non-RFC4122)', () {
      // start from a valid v7, then flip variant bits to something != 2
      final id = const Uuid().v7();
      final v = UuidValue.fromString(id);
      final bytes = v.toBytes();

      // variant is top 2 bits of byte[8]. Set to 0b00 (NCS) instead of 0b10.
      bytes[8] = bytes[8] & 0x3f; // clears top two bits -> 00xxxxxx

      final mutated = Uuid.unparse(bytes);

      expect(() => validateUUID7(mutated), throwsA(isA<Exception>()));
    });
  });

  group('validateUUID7Strict', () {
    test('accepts UUIDv7 generated now', () {
      final id = const Uuid().v7();
      expect(() => validateUUID7Strict(id), returnsNormally);
    });

    test('rejects invalid format', () {
      expect(
        () => validateUUID7Strict('not-a-uuid'),
        throwsA(isA<Exception>()),
      );
    });

    test('rejects non-v7 UUID (e.g. v4)', () {
      final id = const Uuid().v4();
      expect(() => validateUUID7Strict(id), throwsA(isA<Exception>()));
    });

    test('rejects wrong variant (non-RFC4122)', () {
      final id = const Uuid().v7();
      final v = UuidValue.fromString(id);
      final bytes = v.toBytes();

      // Set variant to 0b11 (future/reserved) instead of 0b10
      bytes[8] = (bytes[8] & 0x3f) | 0xc0; // 11xxxxxx

      final mutated = Uuid.unparse(bytes);

      expect(() => validateUUID7Strict(mutated), throwsA(isA<Exception>()));
    });

    test('rejects UUIDv7 timestamp too far in the future', () {
      // Build a UUIDv7-like byte array but with an absurdly-future timestamp.
      final id = const Uuid().v7();
      final v = UuidValue.fromString(id);
      final bytes = v.toBytes();

      final now = DateTime.now().millisecondsSinceEpoch;
      const maxFutureDriftMs = 5 * 60 * 1000;
      final futureMs = now + maxFutureDriftMs + 60 * 1000; // +1 min past drift

      // Write future timestamp into first 6 bytes (big-endian 48 bits)
      bytes[0] = (futureMs >> 40) & 0xff;
      bytes[1] = (futureMs >> 32) & 0xff;
      bytes[2] = (futureMs >> 24) & 0xff;
      bytes[3] = (futureMs >> 16) & 0xff;
      bytes[4] = (futureMs >> 8) & 0xff;
      bytes[5] = (futureMs) & 0xff;

      // Ensure version is 7 (high nibble of byte[6])
      bytes[6] = (bytes[6] & 0x0f) | 0x70;

      // Ensure RFC4122 variant (10xxxxxx)
      bytes[8] = (bytes[8] & 0x3f) | 0x80;

      final mutated = Uuid.unparse(bytes);

      expect(() => validateUUID7Strict(mutated), throwsA(isA<Exception>()));
    });
  });
}
