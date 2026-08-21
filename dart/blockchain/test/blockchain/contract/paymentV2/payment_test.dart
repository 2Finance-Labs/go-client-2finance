import 'dart:convert';

import 'package:test/test.dart';

import 'package:two_finance_blockchain/blockchain/contract/paymentV2/constants.dart';
import 'package:two_finance_blockchain/blockchain/contract/paymentV2/domain/payment.dart';
import 'package:two_finance_blockchain/blockchain/contract/paymentV2/models/payment.dart';
import 'package:two_finance_blockchain/blockchain/contract/tokenV2/constants.dart';
import 'package:two_finance_blockchain/blockchain/types/types.dart';
import 'package:two_finance_blockchain/blockchain/utils/event.dart';
import 'package:two_finance_blockchain/blockchain/utils/marshal.dart';
import 'package:two_finance_blockchain/two_finance_blockchain.dart';

import '../../../helpers/helpers.dart';

String _addAmounts(String a, String b) =>
    (BigInt.parse(a) + BigInt.parse(b)).toString();
String _subAmounts(String a, String b) =>
    (BigInt.parse(a) - BigInt.parse(b)).toString();

Future<PaymentState> _getPaymentState(
  TwoFinanceBlockchain c,
  String address,
) async {
  final out = await c.getPayment(address: address);
  expect(out.states, isNotNull);
  expect(out.states!, isNotEmpty);

  return unmarshalState(
    out.states!.first.object,
    (json) => PaymentState.fromJson(json),
  );
}

List<PaymentState> _parsePaymentListState(Object? obj) {
  if (obj == null) {
    throw StateError('payment list state is null');
  }

  dynamic decoded = obj;
  if (decoded is String) {
    decoded = jsonDecode(decoded);
  }

  if (decoded is Map) {
    decoded =
        decoded['payments'] ?? decoded['items'] ?? decoded['data'] ?? decoded;
  }

  if (decoded is! List) {
    throw StateError(
      'unexpected payment list state type: ${decoded.runtimeType}',
    );
  }

  return List<dynamic>.from(decoded)
      .map(
        (item) => PaymentState.fromJson(Map<String, dynamic>.from(item as Map)),
      )
      .toList();
}

void _expectDateClose(
  DateTime? actual,
  DateTime expected, {
  int toleranceSeconds = 2,
}) {
  expect(actual, isNotNull);
  expect(
    actual!.toUtc().difference(expected.toUtc()).inSeconds.abs(),
    lessThanOrEqualTo(toleranceSeconds),
  );
}

void main() {
  final testTimeout = Timeout(Duration(minutes: 4));

  group('PaymentV2 E2E', () {
    test(
      'E2E: create + pause/unpause + authorize/capture/refund + void + direct pay + list',
      () async {
        final c = await setupClient();

        final ownerUser = await newTestUser('owner');
        final payerUser = await newTestUser('payer');
        final payeeUser = await newTestUser('payee');

        // ------------------
        //      TOKEN
        // ------------------

        final tokenAddress = await createBasicToken(
          c,
          ownerPublicKey: ownerUser.publicKey,
          ownerPrivateKey: ownerUser.privateKey,
          decimals: 6,
          requireFee: false,
          tokenType: TOKEN_TYPE_FUNGIBLE,
          stablecoin: false,
        );

        await c.setPrivateKey(ownerUser.privateKey);
        final deployed = await c.deployContract1(PAYMENT_CONTRACT_V2);
        expect(deployed.logs, isNotNull);
        expect(deployed.logs!, isNotEmpty);

        final paymentAddress = deployed.logs!.first.contractAddress;
        expect(paymentAddress, isNotEmpty);

        // ------------------
        //   ALLOW USERS
        // ------------------

        final outAllow = await c.allowUsers(tokenAddress, {
          ownerUser.publicKey: true,
          payerUser.publicKey: true,
          payeeUser.publicKey: true,
          paymentAddress: true,
        });
        expect(outAllow.logs, isNotNull);
        expect(outAllow.logs!, isNotEmpty);

        // ------------------
        //  FUND THE PAYER
        // ------------------

        const fundAmount = '500';
        final outFund = await c.transferToken(
          tokenAddress: tokenAddress,
          transferTo: payerUser.publicKey,
          amount: fundAmount,
          decimals: 0,
          tokenType: TOKEN_TYPE_FUNGIBLE,
          uuid: '',
        );
        expect(outFund.logs, isNotNull);
        expect(outFund.logs!, isNotEmpty);

        // ------------------
        //   INITIAL BALANCE
        // ------------------
        final payerBalanceBefore = await getFtBalanceAmount(
          c,
          tokenAddress: tokenAddress,
          ownerAddress: payerUser.publicKey,
        );

        String payeeBalanceBefore;
        try {
          payeeBalanceBefore = await getFtBalanceAmount(
            c,
            tokenAddress: tokenAddress,
            ownerAddress: payeeUser.publicKey,
          );
        } catch (e) {
          final message = e.toString();
          expect(
            message,
            anyOf(
              contains('record not found'),
              contains('Expected: non-empty'),
            ),
          );
          payeeBalanceBefore = '0';
        }

        const orderId = 'order-payment-e2e-001';
        const amount = '300';
        final expiredAt = DateTime.now().toUtc().add(const Duration(hours: 2));

        // ------------------
        //   CREATE PAYMENT
        // ------------------

        await c.setPrivateKey(ownerUser.privateKey);
        final createPaymentOut = await c.createPayment(
          address: paymentAddress,
          owner: ownerUser.publicKey,
          tokenAddress: tokenAddress,
          orderId: orderId,
          payer: payerUser.publicKey,
          payee: payeeUser.publicKey,
          amount: amount,
          expiredAt: expiredAt,
        );
        expect(createPaymentOut, isA<ContractOutput>());
        expect(createPaymentOut.logs, isNotNull);
        expect(createPaymentOut.logs!, isNotEmpty);

        final createLog = createPaymentOut.logs!.first;
        expect(createLog.logType, equals(PAYMENT_CREATED_LOG));
        expect(createLog.contractAddress, equals(paymentAddress));

        final createEvent = decodeEvent(createLog.event);
        expect(createEvent['address'], equals(paymentAddress));
        expect(createEvent['owner'], equals(ownerUser.publicKey));
        expect(createEvent['token_address'], equals(tokenAddress));
        expect(createEvent['order_id'], equals(orderId));
        expect(createEvent['payer'], equals(payerUser.publicKey));
        expect(createEvent['payee'], equals(payeeUser.publicKey));
        expect(createEvent['amount'], equals(amount));
        expect(createEvent['status'], equals(STATUS_CREATED));
        expect(createEvent['paused'], anyOf(isNull, isFalse));

        if (createEvent['expired_at'] != null &&
            createEvent['expired_at'].toString().isNotEmpty) {
          _expectDateClose(
            DateTime.parse(createEvent['expired_at'].toString()),
            expiredAt,
          );
        }

        final paymentStateAfterCreate = await _getPaymentState(
          c,
          paymentAddress,
        );
        expect(paymentStateAfterCreate.address, equals(paymentAddress));
        expect(paymentStateAfterCreate.owner, equals(ownerUser.publicKey));
        expect(paymentStateAfterCreate.tokenAddress, equals(tokenAddress));
        expect(paymentStateAfterCreate.orderId, equals(orderId));
        expect(paymentStateAfterCreate.payer, equals(payerUser.publicKey));
        expect(paymentStateAfterCreate.payee, equals(payeeUser.publicKey));
        expect(paymentStateAfterCreate.amount, equals(amount));
        expect(paymentStateAfterCreate.status, equals(STATUS_CREATED));
        expect(paymentStateAfterCreate.paused, isFalse);
        expect(paymentStateAfterCreate.hash, isNotEmpty);
        if (paymentStateAfterCreate.expiredAt != null &&
            paymentStateAfterCreate.expiredAt!.year > 1) {
          _expectDateClose(paymentStateAfterCreate.expiredAt, expiredAt);
        }

        // ------------------
        //       PAUSE
        // ------------------

        await c.setPrivateKey(payerUser.privateKey);
        final pauseOut = await c.pausePayment(
          address: paymentAddress,
          paused: true,
        );
        expect(pauseOut.logs, isNotNull);
        expect(pauseOut.logs!, isNotEmpty);
        expect(pauseOut.logs!.first.logType, equals(PAYMENT_PAUSED_LOG));

        final pauseEvent = decodeEvent(pauseOut.logs!.first.event);
        expect(pauseEvent['address'], equals(paymentAddress));
        expect(pauseEvent['paused'], isTrue);

        final paymentStateAfterPause = await _getPaymentState(
          c,
          paymentAddress,
        );
        expect(paymentStateAfterPause.paused, isTrue);

        // ------------------
        //      UNPAUSE
        // ------------------

        final unpauseOut = await c.unpausePayment(
          address: paymentAddress,
          paused: false,
        );
        expect(unpauseOut.logs, isNotNull);
        expect(unpauseOut.logs!, isNotEmpty);
        expect(unpauseOut.logs!.first.logType, equals(PAYMENT_UNPAUSED_LOG));

        final unpauseEvent = decodeEvent(unpauseOut.logs!.first.event);
        expect(unpauseEvent['address'], equals(paymentAddress));
        expect(unpauseEvent['paused'], anyOf(isNull, isFalse));

        final paymentStateAfterUnpause = await _getPaymentState(
          c,
          paymentAddress,
        );
        expect(paymentStateAfterUnpause.paused, isFalse);

        // ------------------
        //     AUTHORIZE
        // ------------------

        final authorizeOut = await c.authorizePayment(address: paymentAddress);
        expect(authorizeOut.logs, isNotNull);
        expect(authorizeOut.logs!, isNotEmpty);
        expect(
          authorizeOut.logs!.first.logType,
          equals(PAYMENT_AUTHORIZED_LOG),
        );

        final authorizeEvent = decodeEvent(authorizeOut.logs!.first.event);
        expect(authorizeEvent['address'], equals(paymentAddress));
        expect(authorizeEvent['status'], equals(STATUS_AUTHORIZED));

        final paymentStateAfterAuthorize = await _getPaymentState(
          c,
          paymentAddress,
        );
        expect(paymentStateAfterAuthorize.status, equals(STATUS_AUTHORIZED));

        // ------------------
        //      CAPTURE
        // ------------------

        await c.setPrivateKey(payeeUser.privateKey);
        final captureOut = await c.capturePayment(address: paymentAddress);
        expect(captureOut.logs, isNotNull);
        expect(captureOut.logs!, isNotEmpty);
        expect(captureOut.logs!.first.logType, equals(PAYMENT_CAPTURED_LOG));

        final captureEvent = decodeEvent(captureOut.logs!.first.event);
        expect(captureEvent['address'], equals(paymentAddress));
        expect(captureEvent['status'], equals(STATUS_CAPTURED));

        final paymentStateAfterCapture = await _getPaymentState(
          c,
          paymentAddress,
        );
        expect(paymentStateAfterCapture.status, equals(STATUS_CAPTURED));
        expect(paymentStateAfterCapture.capturedAmount, equals(amount));

        final payerBalanceAfterCapture = await getFtBalanceAmount(
          c,
          tokenAddress: tokenAddress,
          ownerAddress: payerUser.publicKey,
        );
        final payeeBalanceAfterCapture = await getFtBalanceAmount(
          c,
          tokenAddress: tokenAddress,
          ownerAddress: payeeUser.publicKey,
        );

        expect(
          payerBalanceAfterCapture,
          equals(_subAmounts(payerBalanceBefore, amount)),
        );
        expect(
          payeeBalanceAfterCapture,
          equals(_addAmounts(payeeBalanceBefore, amount)),
        );

        // ------------------
        //       REFUND
        // ------------------

        const refundAmount = '100';
        final refundOut = await c.refundPayment(
          address: paymentAddress,
          amount: refundAmount,
        );
        expect(refundOut.logs, isNotNull);
        expect(refundOut.logs!, isNotEmpty);
        expect(refundOut.logs!.first.logType, equals(PAYMENT_REFUNDED_LOG));

        final refundEvent = decodeEvent(refundOut.logs!.first.event);
        expect(refundEvent['address'], equals(paymentAddress));
        expect(refundEvent['status'], equals(STATUS_REFUNDED));
        expect(refundEvent['refunded_amount'], equals(refundAmount));

        final paymentStateAfterRefund = await _getPaymentState(
          c,
          paymentAddress,
        );
        expect(paymentStateAfterRefund.status, equals(STATUS_REFUNDED));
        expect(paymentStateAfterRefund.refundedAmount, equals(refundAmount));

        final payerBalanceAfterRefund = await getFtBalanceAmount(
          c,
          tokenAddress: tokenAddress,
          ownerAddress: payerUser.publicKey,
        );
        final payeeBalanceAfterRefund = await getFtBalanceAmount(
          c,
          tokenAddress: tokenAddress,
          ownerAddress: payeeUser.publicKey,
        );

        expect(
          payerBalanceAfterRefund,
          equals(_addAmounts(payerBalanceAfterCapture, refundAmount)),
        );
        expect(
          payeeBalanceAfterRefund,
          equals(_subAmounts(payeeBalanceAfterCapture, refundAmount)),
        );

        // ------------------
        //     VOID FLOW
        // ------------------

        final deployedVoid = await c.deployContract1(PAYMENT_CONTRACT_V2);
        final voidPaymentAddress = deployedVoid.logs!.first.contractAddress;
        expect(voidPaymentAddress, isNotEmpty);

        await c.setPrivateKey(ownerUser.privateKey);
        final outAllowVoid = await c.allowUsers(tokenAddress, {
          voidPaymentAddress: true,
        });
        expect(outAllowVoid.logs, isNotNull);
        expect(outAllowVoid.logs!, isNotEmpty);

        const voidOrderId = 'order-payment-e2e-void-001';
        const voidAmount = '80';
        await c.createPayment(
          address: voidPaymentAddress,
          owner: ownerUser.publicKey,
          tokenAddress: tokenAddress,
          orderId: voidOrderId,
          payer: payerUser.publicKey,
          payee: payeeUser.publicKey,
          amount: voidAmount,
          expiredAt: expiredAt,
        );

        await c.setPrivateKey(payerUser.privateKey);
        await c.authorizePayment(address: voidPaymentAddress);
        // ------------------
        //        VOID
        // ------------------
        final voidOut = await c.voidPayment(address: voidPaymentAddress);
        expect(voidOut.logs, isNotNull);
        expect(voidOut.logs!, isNotEmpty);
        expect(voidOut.logs!.first.logType, equals(PAYMENT_VOIDED_LOG));

        final voidEvent = decodeEvent(voidOut.logs!.first.event);
        expect(voidEvent['address'], equals(voidPaymentAddress));
        expect(voidEvent['status'], equals(STATUS_VOIDED));

        final voidState = await _getPaymentState(c, voidPaymentAddress);
        expect(voidState.address, equals(voidPaymentAddress));
        expect(voidState.owner, equals(ownerUser.publicKey));
        expect(voidState.tokenAddress, equals(tokenAddress));
        expect(voidState.orderId, equals(voidOrderId));
        expect(voidState.payer, equals(payerUser.publicKey));
        expect(voidState.payee, equals(payeeUser.publicKey));
        expect(voidState.amount, equals(voidAmount));
        expect(voidState.status, equals(STATUS_VOIDED));
        expect(voidState.paused, isFalse);
        expect(voidState.hash, isNotEmpty);

        // ------------------
        //     DIRECT PAY
        // ------------------

        final deployedDirect = await c.deployContract1(PAYMENT_CONTRACT_V2);
        final directPaymentAddress = deployedDirect.logs!.first.contractAddress;
        expect(directPaymentAddress, isNotEmpty);

        await c.setPrivateKey(ownerUser.privateKey);
        final outAllowDirect = await c.allowUsers(tokenAddress, {
          directPaymentAddress: true,
        });
        expect(outAllowDirect.logs, isNotNull);
        expect(outAllowDirect.logs!, isNotEmpty);

        final payerBalanceBeforeDirect = await getFtBalanceAmount(
          c,
          tokenAddress: tokenAddress,
          ownerAddress: payerUser.publicKey,
        );
        final payeeBalanceBeforeDirect = await getFtBalanceAmount(
          c,
          tokenAddress: tokenAddress,
          ownerAddress: payeeUser.publicKey,
        );

        const directPayAmount = '50';
        const directPayOrderId = 'order-payment-e2e-direct-001';

        await c.setPrivateKey(payerUser.privateKey);
        final directPayOut = await c.directPay(
          address: directPaymentAddress,
          owner: payerUser.publicKey,
          tokenAddress: tokenAddress,
          orderId: directPayOrderId,
          payer: payerUser.publicKey,
          payee: payeeUser.publicKey,
          amount: directPayAmount,
          expiredAt: DateTime.now().toUtc().add(const Duration(hours: 2)),
        );

        expect(directPayOut.logs, isNotNull);
        expect(directPayOut.logs, hasLength(3));

        expect(directPayOut.logs![0].logType, equals(PAYMENT_CREATED_LOG));
        expect(directPayOut.logs![1].logType, equals(PAYMENT_AUTHORIZED_LOG));
        expect(directPayOut.logs![2].logType, equals(PAYMENT_CAPTURED_LOG));

        final directCreatedEvent = decodeEvent(directPayOut.logs![0].event);
        expect(directCreatedEvent['address'], equals(directPaymentAddress));
        expect(directCreatedEvent['owner'], equals(payerUser.publicKey));
        expect(directCreatedEvent['token_address'], equals(tokenAddress));
        expect(directCreatedEvent['order_id'], equals(directPayOrderId));
        expect(directCreatedEvent['payer'], equals(payerUser.publicKey));
        expect(directCreatedEvent['payee'], equals(payeeUser.publicKey));
        expect(directCreatedEvent['amount'], equals(directPayAmount));
        expect(directCreatedEvent['status'], equals(STATUS_CREATED));
        expect(directCreatedEvent['paused'], anyOf(isNull, isFalse));

        final directAuthorizedEvent = decodeEvent(directPayOut.logs![1].event);
        expect(directAuthorizedEvent['address'], equals(directPaymentAddress));
        expect(directAuthorizedEvent['status'], equals(STATUS_AUTHORIZED));

        final directCapturedEvent = decodeEvent(directPayOut.logs![2].event);
        expect(directCapturedEvent['address'], equals(directPaymentAddress));
        expect(directCapturedEvent['captured_amount'], equals(directPayAmount));
        expect(directCapturedEvent['status'], equals(STATUS_CAPTURED));

        final directPaymentState = await _getPaymentState(
          c,
          directPaymentAddress,
        );
        expect(directPaymentState.address, equals(directPaymentAddress));
        expect(directPaymentState.owner, equals(payerUser.publicKey));
        expect(directPaymentState.tokenAddress, equals(tokenAddress));
        expect(directPaymentState.orderId, equals(directPayOrderId));
        expect(directPaymentState.payer, equals(payerUser.publicKey));
        expect(directPaymentState.payee, equals(payeeUser.publicKey));
        expect(directPaymentState.amount, equals(directPayAmount));
        expect(directPaymentState.capturedAmount, equals(directPayAmount));
        expect(directPaymentState.refundedAmount, equals('0'));
        expect(directPaymentState.status, equals(STATUS_CAPTURED));
        expect(directPaymentState.paused, isFalse);
        expect(directPaymentState.hash, isNotEmpty);

        final payerBalanceAfterDirect = await getFtBalanceAmount(
          c,
          tokenAddress: tokenAddress,
          ownerAddress: payerUser.publicKey,
        );
        final payeeBalanceAfterDirect = await getFtBalanceAmount(
          c,
          tokenAddress: tokenAddress,
          ownerAddress: payeeUser.publicKey,
        );

        expect(
          payerBalanceAfterDirect,
          equals(_subAmounts(payerBalanceBeforeDirect, directPayAmount)),
        );
        expect(
          payeeBalanceAfterDirect,
          equals(_addAmounts(payeeBalanceBeforeDirect, directPayAmount)),
        );

        // ------------------
        //    LIST PAYMENTS
        // ------------------

        final listPaymentsOut = await c.listPayments(
          tokenAddress: tokenAddress,
          payer: payerUser.publicKey,
          payee: payeeUser.publicKey,
          page: 1,
          limit: 10,
          ascending: true,
        );
        expect(listPaymentsOut.states, isNotNull);
        expect(listPaymentsOut.states!, isNotEmpty);

        final payments = _parsePaymentListState(
          listPaymentsOut.states!.first.object,
        );
        expect(payments, isNotEmpty);

        PaymentState? foundCreatedFlow;
        PaymentState? foundVoidFlow;
        PaymentState? foundDirectFlow;

        for (final payment in payments) {
          if (payment.address == paymentAddress) {
            foundCreatedFlow = payment;
          }
          if (payment.address == voidPaymentAddress) {
            foundVoidFlow = payment;
          }
          if (payment.address == directPaymentAddress) {
            foundDirectFlow = payment;
          }
        }

        expect(foundCreatedFlow, isNotNull);
        expect(foundCreatedFlow!.orderId, equals(orderId));
        expect(foundCreatedFlow.payer, equals(payerUser.publicKey));
        expect(foundCreatedFlow.payee, equals(payeeUser.publicKey));
        expect(foundCreatedFlow.amount, equals(amount));
        expect(foundCreatedFlow.status, equals(STATUS_REFUNDED));
        expect(foundCreatedFlow.capturedAmount, equals(amount));
        expect(foundCreatedFlow.refundedAmount, equals(refundAmount));
        expect(foundCreatedFlow.createdAt, isNotNull);
        expect(foundCreatedFlow.createdAt!.year, greaterThan(1));
        expect(foundCreatedFlow.updatedAt, isNotNull);
        expect(foundCreatedFlow.updatedAt!.year, greaterThan(1));

        expect(foundVoidFlow, isNotNull);
        expect(foundVoidFlow!.orderId, equals(voidOrderId));
        expect(foundVoidFlow.payer, equals(payerUser.publicKey));
        expect(foundVoidFlow.payee, equals(payeeUser.publicKey));
        expect(foundVoidFlow.amount, equals(voidAmount));
        expect(foundVoidFlow.status, equals(STATUS_VOIDED));
        expect(foundVoidFlow.createdAt, isNotNull);
        expect(foundVoidFlow.createdAt!.year, greaterThan(1));
        expect(foundVoidFlow.updatedAt, isNotNull);
        expect(foundVoidFlow.updatedAt!.year, greaterThan(1));

        expect(foundDirectFlow, isNotNull);
        expect(foundDirectFlow!.orderId, equals(directPayOrderId));
        expect(foundDirectFlow.payer, equals(payerUser.publicKey));
        expect(foundDirectFlow.payee, equals(payeeUser.publicKey));
        expect(foundDirectFlow.amount, equals(directPayAmount));
        expect(foundDirectFlow.capturedAmount, equals(directPayAmount));
        expect(foundDirectFlow.refundedAmount, equals('0'));
        expect(foundDirectFlow.status, equals(STATUS_CAPTURED));
        expect(foundDirectFlow.createdAt, isNotNull);
        expect(foundDirectFlow.createdAt!.year, greaterThan(1));
        expect(foundDirectFlow.updatedAt, isNotNull);
        expect(foundDirectFlow.updatedAt!.year, greaterThan(1));
      },
      timeout: testTimeout,
    );

    test('E2E: authorize + void', () async {
      final c = await setupClient();

      // ------------------
      //      WALLETS
      // ------------------
      final ownerUser = await newTestUser('owner');
      final payerUser = await newTestUser('payer');
      final payeeUser = await newTestUser('payee');

      // ------------------
      //      TOKEN
      // ------------------
      final tokenAddress = await createBasicToken(
        c,
        ownerPublicKey: ownerUser.publicKey,
        ownerPrivateKey: ownerUser.privateKey,
        decimals: 6,
        requireFee: false,
        tokenType: TOKEN_TYPE_FUNGIBLE,
        stablecoin: false,
      );

      // ------------------
      //   DEPLOY PAYMENT
      // ------------------
      await c.setPrivateKey(ownerUser.privateKey);
      final deployed = await c.deployContract1(PAYMENT_CONTRACT_V2);
      expect(deployed.logs, isNotNull);
      expect(deployed.logs!, isNotEmpty);

      final paymentAddress = deployed.logs!.first.contractAddress;
      expect(paymentAddress, isNotEmpty);

      // ------------------
      //   ALLOW USERS
      // ------------------
      final outAllow = await c.allowUsers(tokenAddress, {
        ownerUser.publicKey: true,
        payerUser.publicKey: true,
        payeeUser.publicKey: true,
        paymentAddress: true,
      });
      expect(outAllow.logs, isNotNull);
      expect(outAllow.logs!, isNotEmpty);

      // ------------------
      //    FUND PAYER
      // ------------------
      const fundAmount = '500';
      final outFund = await c.transferToken(
        tokenAddress: tokenAddress,
        transferTo: payerUser.publicKey,
        amount: fundAmount,
        decimals: 0,
        tokenType: TOKEN_TYPE_FUNGIBLE,
        uuid: '',
      );
      expect(outFund.logs, isNotNull);
      expect(outFund.logs!, isNotEmpty);

      // ------------------
      //   INITIAL BALANCE
      // ------------------
      final payerBalanceBefore = await getFtBalanceAmount(
        c,
        tokenAddress: tokenAddress,
        ownerAddress: payerUser.publicKey,
      );

      String payeeBalanceBefore;
      try {
        payeeBalanceBefore = await getFtBalanceAmount(
          c,
          tokenAddress: tokenAddress,
          ownerAddress: payeeUser.publicKey,
        );
      } catch (e) {
        final message = e.toString();
        expect(
          message,
          anyOf(contains('record not found'), contains('Expected: non-empty')),
        );
        payeeBalanceBefore = '0';
      }

      // ------------------
      //   CREATE PAYMENT
      // ------------------
      const orderId = 'order-payment-auth-void-e2e-001';
      const amount = '300';
      final expiredAt = DateTime.now().toUtc().add(const Duration(hours: 2));

      final createPaymentOut = await c.createPayment(
        address: paymentAddress,
        owner: ownerUser.publicKey,
        tokenAddress: tokenAddress,
        orderId: orderId,
        payer: payerUser.publicKey,
        payee: payeeUser.publicKey,
        amount: amount,
        expiredAt: expiredAt,
      );
      expect(createPaymentOut.logs, isNotNull);
      expect(createPaymentOut.logs!, isNotEmpty);

      final createLog = createPaymentOut.logs!.first;
      expect(createLog.logType, equals(PAYMENT_CREATED_LOG));

      final createEvent = decodeEvent(createLog.event);
      expect(createEvent['address'], equals(paymentAddress));
      expect(createEvent['owner'], equals(ownerUser.publicKey));
      expect(createEvent['token_address'], equals(tokenAddress));
      expect(createEvent['order_id'], equals(orderId));
      expect(createEvent['payer'], equals(payerUser.publicKey));
      expect(createEvent['payee'], equals(payeeUser.publicKey));
      expect(createEvent['amount'], equals(amount));
      expect(createEvent['status'], equals(STATUS_CREATED));
      expect(createEvent['paused'], anyOf(isNull, isFalse));

      final paymentStateAfterCreate = await _getPaymentState(c, paymentAddress);
      expect(paymentStateAfterCreate.address, equals(paymentAddress));
      expect(paymentStateAfterCreate.owner, equals(ownerUser.publicKey));
      expect(paymentStateAfterCreate.tokenAddress, equals(tokenAddress));
      expect(paymentStateAfterCreate.orderId, equals(orderId));
      expect(paymentStateAfterCreate.payer, equals(payerUser.publicKey));
      expect(paymentStateAfterCreate.payee, equals(payeeUser.publicKey));
      expect(paymentStateAfterCreate.amount, equals(amount));
      expect(paymentStateAfterCreate.status, equals(STATUS_CREATED));
      expect(paymentStateAfterCreate.paused, isFalse);
      expect(paymentStateAfterCreate.hash, isNotEmpty);

      // ------------------
      //     AUTHORIZE
      // ------------------
      await c.setPrivateKey(payerUser.privateKey);
      final authorizeOut = await c.authorizePayment(address: paymentAddress);
      expect(authorizeOut.logs, isNotNull);
      expect(authorizeOut.logs!, isNotEmpty);
      expect(authorizeOut.logs!.first.logType, equals(PAYMENT_AUTHORIZED_LOG));

      final authorizeEvent = decodeEvent(authorizeOut.logs!.first.event);
      expect(authorizeEvent['address'], equals(paymentAddress));
      expect(authorizeEvent['status'], equals(STATUS_AUTHORIZED));

      final paymentStateAfterAuthorize = await _getPaymentState(
        c,
        paymentAddress,
      );
      expect(paymentStateAfterAuthorize.address, equals(paymentAddress));
      expect(paymentStateAfterAuthorize.status, equals(STATUS_AUTHORIZED));
      expect(paymentStateAfterAuthorize.amount, equals(amount));
      expect(paymentStateAfterAuthorize.paused, isFalse);

      // ------------------
      //        VOID
      // ------------------
      final voidOut = await c.voidPayment(address: paymentAddress);
      expect(voidOut.logs, isNotNull);
      expect(voidOut.logs!, isNotEmpty);
      expect(voidOut.logs!.first.logType, equals(PAYMENT_VOIDED_LOG));

      final voidEvent = decodeEvent(voidOut.logs!.first.event);
      expect(voidEvent['address'], equals(paymentAddress));
      expect(voidEvent['status'], equals(STATUS_VOIDED));

      final paymentStateAfterVoid = await _getPaymentState(c, paymentAddress);
      expect(paymentStateAfterVoid.address, equals(paymentAddress));
      expect(paymentStateAfterVoid.owner, equals(ownerUser.publicKey));
      expect(paymentStateAfterVoid.tokenAddress, equals(tokenAddress));
      expect(paymentStateAfterVoid.orderId, equals(orderId));
      expect(paymentStateAfterVoid.payer, equals(payerUser.publicKey));
      expect(paymentStateAfterVoid.payee, equals(payeeUser.publicKey));
      expect(paymentStateAfterVoid.amount, equals(amount));
      expect(paymentStateAfterVoid.status, equals(STATUS_VOIDED));
      expect(paymentStateAfterVoid.paused, isFalse);
      expect(paymentStateAfterVoid.hash, isNotEmpty);

      // ------------------
      //    BALANCE CHECK
      // ------------------
      final payerBalanceAfterVoid = await getFtBalanceAmount(
        c,
        tokenAddress: tokenAddress,
        ownerAddress: payerUser.publicKey,
      );

      String payeeBalanceAfterVoid;
      try {
        payeeBalanceAfterVoid = await getFtBalanceAmount(
          c,
          tokenAddress: tokenAddress,
          ownerAddress: payeeUser.publicKey,
        );
      } catch (e) {
        final message = e.toString();
        expect(
          message,
          anyOf(contains('record not found'), contains('Expected: non-empty')),
        );
        payeeBalanceAfterVoid = '0';
      }

      expect(payerBalanceAfterVoid, equals(payerBalanceBefore));
      expect(payeeBalanceAfterVoid, equals(payeeBalanceBefore));

      // ------------------
      //    LIST PAYMENTS
      // ------------------
      final listPaymentsOut = await c.listPayments(
        orderId: orderId,
        tokenAddress: tokenAddress,
        status: const [STATUS_VOIDED],
        payer: payerUser.publicKey,
        payee: payeeUser.publicKey,
        page: 1,
        limit: 10,
        ascending: true,
      );
      expect(listPaymentsOut.states, isNotNull);
      expect(listPaymentsOut.states!, isNotEmpty);

      final payments = _parsePaymentListState(
        listPaymentsOut.states!.first.object,
      );
      expect(payments, isNotEmpty);

      PaymentState? foundPayment;
      for (final payment in payments) {
        if (payment.address == paymentAddress) {
          foundPayment = payment;
          break;
        }
      }

      expect(foundPayment, isNotNull);
      expect(foundPayment!.owner, equals(ownerUser.publicKey));
      expect(foundPayment.tokenAddress, equals(tokenAddress));
      expect(foundPayment.orderId, equals(orderId));
      expect(foundPayment.payer, equals(payerUser.publicKey));
      expect(foundPayment.payee, equals(payeeUser.publicKey));
      expect(foundPayment.amount, equals(amount));
      expect(foundPayment.status, equals(STATUS_VOIDED));
      expect(foundPayment.createdAt, isNotNull);
      expect(foundPayment.updatedAt, isNotNull);
    }, timeout: testTimeout);
  });
}
