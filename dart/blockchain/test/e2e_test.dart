import 'dart:async';
import 'dart:convert';
import 'dart:io';
import 'dart:math';
import 'dart:typed_data';
import 'package:crypto/crypto.dart';
import 'package:two_finance_blockchain/blockchain/keys/keys.dart';
import 'package:two_finance_blockchain/config/config.dart';
import 'package:two_finance_blockchain/infra/mqtt/mqtt.dart';
import 'package:two_finance_blockchain/two_finance_blockchain.dart';
import 'package:two_finance_blockchain/blockchain/utils/json.dart';
import 'package:test/test.dart';
import 'helpers/helpers.dart';
// ----------------------------------------------------------------------------
// Helpers
// ----------------------------------------------------------------------------

void main() {
  test('SetupClient', () async {
    expect(await setupClient(), isA<TwoFinanceBlockchain>());
  });

  test("Teste randSuffix", () {
    final s1 = randSuffix(4);
    final s2 = randSuffix(4);
    expect(s1.length, 4);
    expect(s2.length, 4);
    expect(s1 != s2, true);
  });

  test("Teste hexEncode", () {
    final bytes = Uint8List.fromList([0, 15, 255]);
    final hex = hexEncode(bytes);
    expect(hex, "000fff");
  });

  test("Teste amt", () {
    expect(amt(1, 6), "1000000");
    expect(amt(123, 2), "12300");
  });

  test("Teste waitUntil", () async {
    var count = 0;
    final future = waitUntil(Duration(seconds: 1), () => count >= 3);
    Timer.periodic(Duration(milliseconds: 100), (timer) {
      count++;
      if (count >= 3) timer.cancel();
    });
    await future;
    expect(count, 3);
  });
}

// ----------------------------------------------------------------------------
// Stub Client (substitua pelo real)
// ----------------------------------------------------------------------------
