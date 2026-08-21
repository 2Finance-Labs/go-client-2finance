import 'package:test/test.dart';

// Ajuste o import pro caminho real do seu arquivo.
import 'package:two_finance_blockchain/blockchain/utils/decimals.dart';

void main() {
  group('DecimalRescaler.rescaleString', () {
    test('normalizes comma to dot', () {
      // "1,23" with from=2 => combined "123"
      expect(rescaleDecimalString('1,23', 2, 2), '123');
    });

    test('handles integers (no fractional part)', () {
      // value "5" with from=2 => "500"
      expect(rescaleDecimalString('5', 2, 2), '500');
    });

    test('trims leading zeros in integer part', () {
      // "00012.34" from=2 => int "12" frac "34" => "1234"
      expect(rescaleDecimalString('00012.34', 2, 2), '1234');
    });

    test('pads fractional part when shorter than fromDecimals', () {
      // "1.2" from=3 => frac "200" => combined "1200"
      expect(rescaleDecimalString('1.2', 3, 3), '1200');
    });

    test('truncates fractional part when longer than fromDecimals', () {
      // "1.23456" from=2 => frac "23" => "123"
      expect(rescaleDecimalString('1.23456', 2, 2), '123');
    });

    test('removes all leading zeros after combining', () {
      // "0.0001" from=4 => combined "00001" => trim => "1"
      expect(rescaleDecimalString('0.0001', 4, 4), '1');
    });

    test('diff==0 returns combined', () {
      expect(rescaleDecimalString('10.50', 2, 2), '1050');
    });

    test('upscales when toDecimals > fromDecimals (append zeros)', () {
      // "1.23" from=2 => "123" ; to=5 => append 3 zeros => "123000"
      expect(rescaleDecimalString('1.23', 2, 5), '123000');
    });

    test('downscales when toDecimals < fromDecimals (truncate right)', () {
      // "1.23" from=2 => "123" ; to=1 => remove 1 digit => "12"
      expect(rescaleDecimalString('1.23', 2, 1), '12');

      // "123.45" from=2 => "12345" ; to=0 => remove 2 digits => "123"
      expect(rescaleDecimalString('123.45', 2, 0), '123');
    });

    test('downscale underflow returns 0 when combined length <= -diff', () {
      // "0.01" from=2 => "1" (after trimming) ; to=0 => diff=-2
      // combined length=1 <= 2 => underflow => "0"
      expect(rescaleDecimalString('0.01', 2, 0), '0');

      // "0.0000001" from=7 => "1"; to=0 => diff=-7 => underflow => "0"
      expect(rescaleDecimalString('0.0000001', 7, 0), '0');
    });

    test('invalid numeric input throws FormatException', () {
      expect(
        () => rescaleDecimalString('12a.34', 2, 2),
        throwsA(isA<FormatException>()),
      );

      expect(
        () => rescaleDecimalString(
          '1.2.3',
          2,
          2,
        ), // split => "1" + "2" but extra dot breaks numeric
        throwsA(isA<FormatException>()),
      );
    });

    test('handles empty fractional part', () {
      // "7." => frac "" padded to from=3 => "000" => "7000"
      expect(rescaleDecimalString('7.', 3, 3), '7000');
    });

    test('handles leading zeros only', () {
      // "0000" from=4 => int "0", frac "" padded => "0000" => combined "0"
      expect(rescaleDecimalString('0000', 4, 4), '0');
    });

    test('throws FormatException for multiple decimal separators (1.2.3)', () {
      expect(
        () => rescaleDecimalString('1.2.3', 2, 2),
        throwsA(isA<FormatException>()),
      );

      expect(
        () => rescaleDecimalString('1,2,3', 2, 2),
        throwsA(isA<FormatException>()),
      );
    });
    test(
      'invalid input "1.2.3" throws FormatException and surfaces the error',
      () {
        expect(
          () => rescaleDecimalString('1.2.3', 2, 2),
          throwsA(
            isA<FormatException>().having(
              (e) => e.message,
              'message',
              contains('Invalid numeric input'),
            ),
          ),
        );
      },
    );
    test('invalid input "1,2,3" throws FormatException', () {
      expect(
        () => rescaleDecimalString('1,2,3', 2, 2),
        throwsA(isA<FormatException>()),
      );
    });
  });
}
