import 'dart:convert';

import 'package:test/test.dart';

import 'package:two_finance_blockchain/blockchain/contract/cashbackV2/constants.dart';
import 'package:two_finance_blockchain/blockchain/contract/tokenV2/constants.dart';
import 'package:two_finance_blockchain/blockchain/types/types.dart';
import 'package:two_finance_blockchain/blockchain/utils/marshal.dart';
import 'package:two_finance_blockchain/two_finance_blockchain.dart';

import '../../../helpers/helpers.dart';

typedef JsonMap = Map<String, dynamic>;

String _addAmounts(String a, String b) =>
    (BigInt.parse(a) + BigInt.parse(b)).toString();
String _subAmounts(String a, String b) =>
    (BigInt.parse(a) - BigInt.parse(b)).toString();

Future<JsonMap> _getCashbackState(
  TwoFinanceBlockchain c,
  String address,
) async {
  final out = await c.getCashback(address: address);
  expect(out.states, isNotNull);
  expect(out.states!, isNotEmpty);

  return unmarshalState(
    out.states!.first.object,
    (json) => Map<String, dynamic>.from(json as Map),
  );
}

Future<String> _getFtBalanceAmountOrZero(
  TwoFinanceBlockchain c, {
  required String tokenAddress,
  required String ownerAddress,
}) async {
  try {
    return await getFtBalanceAmount(
      c,
      tokenAddress: tokenAddress,
      ownerAddress: ownerAddress,
    );
  } catch (e) {
    final message = e.toString();
    expect(
      message,
      anyOf(contains('record not found'), contains('Expected: non-empty')),
    );
    return '0';
  }
}

List<JsonMap> _parseCashbackListState(Object? obj) {
  if (obj == null) {
    throw StateError('cashback list state is null');
  }

  dynamic decoded = obj;
  if (decoded is String) {
    decoded = jsonDecode(decoded);
  }

  if (decoded is Map) {
    decoded =
        decoded['cashbacks'] ?? decoded['items'] ?? decoded['data'] ?? decoded;
  }

  if (decoded is! List) {
    throw StateError(
      'unexpected cashback list state type: ${decoded.runtimeType}',
    );
  }

  return List<dynamic>.from(
    decoded,
  ).map((item) => Map<String, dynamic>.from(item as Map)).toList();
}

void _expectDateClose(
  String actualIso,
  DateTime expected, {
  int toleranceSeconds = 2,
}) {
  final actual = DateTime.parse(actualIso).toUtc();
  final expectedUtc = expected.toUtc();

  expect(
    actual.difference(expectedUtc).inSeconds.abs(),
    lessThanOrEqualTo(toleranceSeconds),
  );
}

void _expectCashbackSnapshot(
  JsonMap cashback, {
  required String address,
  required String owner,
  required String tokenAddress,
  required String programType,
  required String percentage,
  required bool paused,
  required DateTime startAt,
  required DateTime expiredAt,
}) {
  expect(cashback['address'], equals(address));
  expect(cashback['owner'], equals(owner));
  expect(cashback['token_address'], equals(tokenAddress));
  expect(cashback['program_type'], equals(programType));
  expect(cashback['percentage'], equals(percentage));
  expect(cashback['paused'], equals(paused));
  _expectDateClose(cashback['start_at'] as String, startAt);
  _expectDateClose(cashback['expired_at'] as String, expiredAt);
}

void main() {
  group('CashbackV2 E2E', () {
    test(
      'E2E: add + get + update + pause/unpause + deposit + claim + withdraw + list',
      () async {
        final c = await setupClient();
        addTearDown(() => teardownClient(c));

        // ------------------
        // SETUP
        // ------------------
        final ownerUser = await newTestUser('owner');
        final customerUser = await newTestUser('customer');

        final tokenAddress = await createBasicToken(
          c,
          ownerPublicKey: ownerUser.publicKey,
          ownerPrivateKey: ownerUser.privateKey,
          decimals: 2,
          requireFee: false,
          tokenType: TOKEN_TYPE_FUNGIBLE,
          stablecoin: false,
        );

        await c.setPrivateKey(ownerUser.privateKey);
        final deployed = await c.deployContract1(CASHBACK_CONTRACT_V2);
        expect(deployed, isA<ContractOutput>());
        expect(deployed.logs, isNotNull);
        expect(deployed.logs!, isNotEmpty);

        final cashbackAddress = deployed.logs!.first.contractAddress;
        expect(cashbackAddress, isNotEmpty);

        final outAllow = await c.allowUsers(tokenAddress, {
          ownerUser.publicKey: true,
          customerUser.publicKey: true,
          cashbackAddress: true,
        });
        expect(outAllow.logs, isNotNull);
        expect(outAllow.logs!, isNotEmpty);

        final startAt = DateTime.now().toUtc().add(const Duration(seconds: 2));
        final expiredAt = startAt.add(const Duration(hours: 1));

        // ------------------
        // ADD CASHBACK
        // ------------------
        const initialProgramType = 'fixed-percentage';
        const initialPercentage = '500';

        final addOut = await c.addCashback(
          address: cashbackAddress,
          owner: ownerUser.publicKey,
          tokenAddress: tokenAddress,
          programType: initialProgramType,
          percentage: initialPercentage,
          startAt: startAt,
          expiredAt: expiredAt,
          paused: false,
        );
        expect(addOut, isA<ContractOutput>());
        expect(addOut.logs, isNotNull);
        expect(addOut.logs!, isNotEmpty);

        final addLog = addOut.logs!.first;
        expect(addLog.contractAddress, equals(cashbackAddress));
        expect(addLog.logType, equals('Cashback_Created'));

        final addEvent = unmarshalEvent<JsonMap>(
          addLog.event,
          (json) => Map<String, dynamic>.from(json as Map),
        );

        _expectCashbackSnapshot(
          addEvent,
          address: cashbackAddress,
          owner: ownerUser.publicKey,
          tokenAddress: tokenAddress,
          programType: initialProgramType,
          percentage: initialPercentage,
          paused: false,
          startAt: startAt,
          expiredAt: expiredAt,
        );

        final cashbackStateAfterAdd = await _getCashbackState(
          c,
          cashbackAddress,
        );
        _expectCashbackSnapshot(
          cashbackStateAfterAdd,
          address: cashbackAddress,
          owner: ownerUser.publicKey,
          tokenAddress: tokenAddress,
          programType: initialProgramType,
          percentage: initialPercentage,
          paused: false,
          startAt: startAt,
          expiredAt: expiredAt,
        );
        final stateHash = cashbackStateAfterAdd['hash'];
        if (stateHash != null) {
          expect(stateHash, isA<String>());
        }

        // ------------------
        // UPDATE CASHBACK
        // ------------------
        final updatedStartAt = DateTime.now().toUtc().add(
          const Duration(seconds: 4),
        );
        final updatedExpiredAt = updatedStartAt.add(const Duration(hours: 2));
        const updatedProgramType = 'variable-percentage';
        const updatedPercentage = '750';

        final updateOut = await c.updateCashback(
          address: cashbackAddress,
          tokenAddress: tokenAddress,
          programType: updatedProgramType,
          percentage: updatedPercentage,
          startAt: updatedStartAt,
          expiredAt: updatedExpiredAt,
        );
        expect(updateOut.logs, isNotNull);
        expect(updateOut.logs!, isNotEmpty);
        expect(updateOut.logs!.first.logType, equals('Cashback_Updated'));

        final updateEvent = unmarshalEvent<JsonMap>(
          updateOut.logs!.first.event,
          (json) => Map<String, dynamic>.from(json as Map),
        );

        expect(updateEvent['address'], equals(cashbackAddress));
        expect(updateEvent['token_address'], equals(tokenAddress));
        expect(updateEvent['program_type'], equals(updatedProgramType));
        expect(updateEvent['percentage'], equals(updatedPercentage));
        _expectDateClose(updateEvent['start_at'] as String, updatedStartAt);
        _expectDateClose(updateEvent['expired_at'] as String, updatedExpiredAt);

        final cashbackStateAfterUpdate = await _getCashbackState(
          c,
          cashbackAddress,
        );
        _expectCashbackSnapshot(
          cashbackStateAfterUpdate,
          address: cashbackAddress,
          owner: ownerUser.publicKey,
          tokenAddress: tokenAddress,
          programType: updatedProgramType,
          percentage: updatedPercentage,
          paused: false,
          startAt: updatedStartAt,
          expiredAt: updatedExpiredAt,
        );

        // ------------------
        // PAUSE CASHBACK
        // ------------------
        final pauseOut = await c.pauseCashback(
          address: cashbackAddress,
          paused: true,
        );
        expect(pauseOut.logs, isNotNull);
        expect(pauseOut.logs!, isNotEmpty);
        expect(pauseOut.logs!.first.logType, equals('Cashback_Paused'));

        final pauseEvent = unmarshalEvent<JsonMap>(
          pauseOut.logs!.first.event,
          (json) => Map<String, dynamic>.from(json as Map),
        );
        expect(pauseEvent['address'], equals(cashbackAddress));
        expect(pauseEvent['paused'], isTrue);

        final cashbackStateAfterPause = await _getCashbackState(
          c,
          cashbackAddress,
        );
        expect(cashbackStateAfterPause['paused'], isTrue);

        final listPausedOut = await c.listCashbacks(
          owner: ownerUser.publicKey,
          tokenAddress: tokenAddress,
          programType: updatedProgramType,
          paused: true,
          page: 1,
          limit: 10,
          ascending: true,
        );
        expect(listPausedOut.states, isNotNull);
        expect(listPausedOut.states!, isNotEmpty);

        final pausedCashbacks = _parseCashbackListState(
          listPausedOut.states!.first.object,
        );
        expect(pausedCashbacks, isA<List<JsonMap>>());

        // ------------------
        // UNPAUSE CASHBACK
        // ------------------
        final unpauseOut = await c.unpauseCashback(
          address: cashbackAddress,
          paused: false,
        );
        expect(unpauseOut.logs, isNotNull);
        expect(unpauseOut.logs!, isNotEmpty);
        expect(unpauseOut.logs!.first.logType, equals('Cashback_Unpaused'));

        final unpauseEvent = unmarshalEvent<JsonMap>(
          unpauseOut.logs!.first.event,
          (json) => Map<String, dynamic>.from(json as Map),
        );
        expect(unpauseEvent['address'], equals(cashbackAddress));
        expect(unpauseEvent['paused'], anyOf(isNull, isFalse));

        final cashbackStateAfterUnpause = await _getCashbackState(
          c,
          cashbackAddress,
        );
        expect(cashbackStateAfterUnpause['paused'], isFalse);

        // ------------------
        // DEPOSIT CASHBACK FUNDS
        // ------------------
        final ownerBalanceBeforeDeposit = await getFtBalanceAmount(
          c,
          tokenAddress: tokenAddress,
          ownerAddress: ownerUser.publicKey,
        );
        final customerBalanceBeforeClaim = await _getFtBalanceAmountOrZero(
          c,
          tokenAddress: tokenAddress,
          ownerAddress: customerUser.publicKey,
        );

        const depositAmount = '200';
        final depositOut = await c.depositCashbackFunds(
          address: cashbackAddress,
          tokenAddress: tokenAddress,
          amount: depositAmount,
        );
        expect(depositOut.logs, isNotNull);
        expect(depositOut.logs!, isNotEmpty);
        expect(depositOut.logs!.first.logType, equals('Cashback_Deposited'));
        expect(depositOut.delegatedCall, isNotNull);
        expect(depositOut.delegatedCall!, isNotEmpty);

        final depositEvent = unmarshalEvent<JsonMap>(
          depositOut.logs!.first.event,
          (json) => Map<String, dynamic>.from(json as Map),
        );
        expect(depositEvent['address'], equals(cashbackAddress));
        expect(depositEvent['token_address'], equals(tokenAddress));
        expect(depositEvent['amount'], equals(depositAmount));

        final ownerBalanceAfterDeposit = await getFtBalanceAmount(
          c,
          tokenAddress: tokenAddress,
          ownerAddress: ownerUser.publicKey,
        );
        final programBalanceAfterDeposit = await getFtBalanceAmount(
          c,
          tokenAddress: tokenAddress,
          ownerAddress: cashbackAddress,
        );

        expect(
          ownerBalanceAfterDeposit,
          equals(_subAmounts(ownerBalanceBeforeDeposit, depositAmount)),
        );
        expect(programBalanceAfterDeposit, equals(depositAmount));

        await waitUntil(
          const Duration(seconds: 10),
          () => DateTime.now().toUtc().isAfter(updatedStartAt),
        );

        // ------------------
        // CLAIM CASHBACK
        // ------------------
        await c.setPrivateKey(customerUser.privateKey);

        const purchaseAmount = '1000';
        const expectedClaimAmount = '75';

        final claimOut = await c.claimCashback(
          address: cashbackAddress,
          amount: purchaseAmount,
        );
        expect(claimOut.logs, isNotNull);
        expect(claimOut.logs!, isNotEmpty);
        expect(claimOut.logs!.first.logType, equals('Cashback_Claimed'));
        expect(claimOut.delegatedCall, isNotNull);
        expect(claimOut.delegatedCall!, isNotEmpty);

        final claimEvent = unmarshalEvent<JsonMap>(
          claimOut.logs!.first.event,
          (json) => Map<String, dynamic>.from(json as Map),
        );
        expect(claimEvent['address'], equals(cashbackAddress));
        expect(claimEvent['token_address'], equals(tokenAddress));
        expect(claimEvent['pay_to'], equals(customerUser.publicKey));
        expect(claimEvent['amount'], equals(expectedClaimAmount));

        final customerBalanceAfterClaim = await getFtBalanceAmount(
          c,
          tokenAddress: tokenAddress,
          ownerAddress: customerUser.publicKey,
        );
        final programBalanceAfterClaim = await getFtBalanceAmount(
          c,
          tokenAddress: tokenAddress,
          ownerAddress: cashbackAddress,
        );

        expect(
          customerBalanceAfterClaim,
          equals(_addAmounts(customerBalanceBeforeClaim, expectedClaimAmount)),
        );
        expect(
          programBalanceAfterClaim,
          equals(_subAmounts(programBalanceAfterDeposit, expectedClaimAmount)),
        );

        // ------------------
        // WITHDRAW CASHBACK FUNDS
        // ------------------
        await c.setPrivateKey(ownerUser.privateKey);

        const withdrawAmount = '125';
        final withdrawOut = await c.withdrawCashbackFunds(
          address: cashbackAddress,
          tokenAddress: tokenAddress,
          amount: withdrawAmount,
        );
        expect(withdrawOut.logs, isNotNull);
        expect(withdrawOut.logs!, isNotEmpty);
        expect(withdrawOut.logs!.first.logType, equals('Cashback_Withdrawn'));
        expect(withdrawOut.delegatedCall, isNotNull);
        expect(withdrawOut.delegatedCall!, isNotEmpty);

        final withdrawEvent = unmarshalEvent<JsonMap>(
          withdrawOut.logs!.first.event,
          (json) => Map<String, dynamic>.from(json as Map),
        );
        expect(withdrawEvent['address'], equals(cashbackAddress));
        expect(withdrawEvent['token_address'], equals(tokenAddress));
        expect(withdrawEvent['amount'], equals(withdrawAmount));

        final ownerBalanceAfterWithdraw = await getFtBalanceAmount(
          c,
          tokenAddress: tokenAddress,
          ownerAddress: ownerUser.publicKey,
        );
        final programBalanceAfterWithdraw = await _getFtBalanceAmountOrZero(
          c,
          tokenAddress: tokenAddress,
          ownerAddress: cashbackAddress,
        );

        expect(
          ownerBalanceAfterWithdraw,
          equals(_addAmounts(ownerBalanceAfterDeposit, withdrawAmount)),
        );
        expect(programBalanceAfterWithdraw, equals('0'));

        // ------------------
        // LIST CASHBACKS
        // ------------------
        final listOut = await c.listCashbacks(
          owner: ownerUser.publicKey,
          tokenAddress: tokenAddress,
          programType: updatedProgramType,
          paused: false,
          page: 1,
          limit: 10,
          ascending: true,
        );
        expect(listOut.states, isNotNull);
        expect(listOut.states!, isNotEmpty);

        final cashbacks = _parseCashbackListState(listOut.states!.first.object);
        expect(cashbacks, isA<List<JsonMap>>());
      },
    );
  });
}
