// test/blockchain/utils/json_test.dart
import 'dart:convert';
import 'dart:typed_data';

import 'package:test/test.dart';

// Adjust import to your real file path.
import 'package:two_finance_blockchain/blockchain/utils/json.dart';

void main() {
  group('JsonMessage + canonicalJsonEncode', () {
    // Helpers
    JsonMessage decodeObject(String s) {
      final decoded = jsonDecode(s);
      expect(decoded, isA<Map>());
      return Map<String, dynamic>.from(decoded as Map);
    }

    test(
      'canonicalJsonEncode encodes JsonMessage to valid JSON object string',
      () {
        final JsonMessage msg = {'a': 1, 'b': 'x', 'c': true};

        final s = canonicalJsonEncode(msg);

        // Must be valid JSON and decode back to same structure
        expect(() => jsonDecode(s), returnsNormally);
        expect(decodeObject(s), msg);
      },
    );

    test(
      'canonicalJsonEncode is deterministic regardless of insertion order (top-level keys)',
      () {
        final JsonMessage m1 = {'b': 2, 'a': 1};
        final JsonMessage m2 = {'a': 1, 'b': 2};

        final s1 = canonicalJsonEncode(m1);
        final s2 = canonicalJsonEncode(m2);

        expect(s1, s2);
        expect(s1, r'{"a":1,"b":2}');
      },
    );

    test('canonicalJsonEncode sorts nested object keys deterministically', () {
      final JsonMessage msg = {
        'z': {'b': 2, 'a': 1},
        'a': 1,
      };

      final s = canonicalJsonEncode(msg);

      // top-level order: a, z ; nested z order: a, b
      expect(s, r'{"a":1,"z":{"a":1,"b":2}}');
      expect(decodeObject(s), msg);
    });

    test(
      'canonicalJsonEncode preserves list order but canonicalizes list items',
      () {
        final JsonMessage msg = {
          'arr': [
            {'b': 2, 'a': 1},
            {'d': 4, 'c': 3},
          ],
        };

        final s = canonicalJsonEncode(msg);

        // arr is a list => order preserved; objects inside list are canonicalized
        expect(s, r'{"arr":[{"a":1,"b":2},{"c":3,"d":4}]}');
        expect(decodeObject(s), msg);
      },
    );

    test('canonicalJsonEncode handles primitives and null stably', () {
      final JsonMessage msg = {'s': 'ok', 'n': 10, 'b': false, 'nullv': null};

      final s = canonicalJsonEncode(msg);

      // key order: b, n, nullv, s
      expect(s, r'{"b":false,"n":10,"nullv":null,"s":"ok"}');
      expect(decodeObject(s), msg);
    });

    test(
      'canonicalJsonEncode output is stable across repeated calls (same input instance)',
      () {
        final JsonMessage msg = {
          'a': 1,
          'b': {'c': 3},
          'd': [1, 2, 3],
        };

        final s1 = canonicalJsonEncode(msg);
        final s2 = canonicalJsonEncode(msg);

        expect(s1, s2);
        expect(decodeObject(s1), msg);
      },
    );

    test(
      'canonicalJsonEncode output is stable across equivalent inputs (deep)',
      () {
        final JsonMessage m1 = {
          'outer': {'b': 2, 'a': 1},
          'list': [
            {'y': 2, 'x': 1},
          ],
        };

        final JsonMessage m2 = {
          'list': [
            {'x': 1, 'y': 2},
          ],
          'outer': {'a': 1, 'b': 2},
        };

        final s1 = canonicalJsonEncode(m1);
        final s2 = canonicalJsonEncode(m2);

        expect(s1, s2);
        expect(decodeObject(s1), m1);
        expect(decodeObject(s2), m2);
      },
    );

    test('JsonMessage should be a JSON object (Map) — not String/bytes', () {
      final JsonMessage msg = {'contract_version': 'walletV2'};
      expect(msg, isA<Map<String, dynamic>>());
      expect(msg['contract_version'], 'walletV2');
    });
  });

  group('canonicalJsonEncode', () {
    test('sorts object keys lexicographically (top-level)', () {
      final m = {'b': 2, 'a': 1, 'c': 3};
      expect(canonicalJsonEncode(m), '{"a":1,"b":2,"c":3}');
    });

    test('sorts nested object keys lexicographically (recursive)', () {
      final m = {
        'z': {'b': 2, 'a': 1},
        'a': 0,
      };
      expect(canonicalJsonEncode(m), '{"a":0,"z":{"a":1,"b":2}}');
    });

    test('keeps list order stable', () {
      final v = [3, 2, 1];
      expect(canonicalJsonEncode(v), '[3,2,1]');
    });

    test('encodes list of objects with canonical objects inside', () {
      final v = [
        {'b': 2, 'a': 1},
        {'d': 4, 'c': 3},
      ];
      expect(canonicalJsonEncode(v), '[{"a":1,"b":2},{"c":3,"d":4}]');
    });

    test('encodes primitives via jsonEncode (string/number/bool/null)', () {
      expect(canonicalJsonEncode('x'), jsonEncode('x'));
      expect(canonicalJsonEncode(10), '10');
      expect(canonicalJsonEncode(true), 'true');
      expect(canonicalJsonEncode(null), 'null');
    });

    test('produces stable output regardless of insertion order', () {
      final m1 = {
        'b': 2,
        'a': 1,
        'c': {'y': 2, 'x': 1},
      };
      final m2 = {
        'c': {'x': 1, 'y': 2},
        'a': 1,
        'b': 2,
      };

      expect(canonicalJsonEncode(m1), canonicalJsonEncode(m2));
      expect(canonicalJsonEncode(m1), '{"a":1,"b":2,"c":{"x":1,"y":2}}');
    });

    test('no whitespace in canonical output', () {
      final m = {'b': 2, 'a': 1};
      final s = canonicalJsonEncode(m);

      expect(s.contains(' '), isFalse);
      expect(s.contains('\n'), isFalse);
      expect(s.contains('\t'), isFalse);
    });
  });
}
