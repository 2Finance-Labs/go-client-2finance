import 'dart:convert';

import 'package:crypto/crypto.dart';
import 'package:test/test.dart';
import 'package:two_finance_blockchain/blockchain/types/types.dart';
import 'package:two_finance_blockchain/blockchain/contract/couponsV2/constants.dart';
import 'package:two_finance_blockchain/blockchain/utils/marshal.dart';
import 'package:two_finance_blockchain/two_finance_blockchain.dart';

import '../../../helpers/helpers.dart';

typedef JsonMap = Map<String, dynamic>;

Future<JsonMap> _getCouponState(TwoFinanceBlockchain c, String address) async {
  final out = await c.getCoupon(address: address);
  expect(out.states, isNotNull);
  expect(out.states!, isNotEmpty);

  return unmarshalState(
    out.states!.first.object,
    (json) => Map<String, dynamic>.from(json as Map),
  );
}

void _expectDateClose(
  String actualIso,
  DateTime expected, {
  int toleranceSeconds = 1,
}) {
  final actual = DateTime.parse(actualIso).toUtc();
  final expectedUtc = expected.toUtc();

  expect(
    actual.difference(expectedUtc).inSeconds.abs(),
    lessThanOrEqualTo(toleranceSeconds),
  );
}

String _stringOrEmpty(dynamic value) {
  if (value == null) return '';
  return value as String;
}

bool _boolOrFalse(dynamic value) {
  if (value == null) return false;
  return value as bool;
}

int _intOrZero(dynamic value) {
  if (value == null) return 0;
  return value as int;
}

void _expectCouponSnapshot(
  JsonMap coupon, {
  required String address,
  String? tokenAddress,
  required String discountType,
  required String percentageBPS,
  required String fixedAmount,
  required String minOrder,
  required bool paused,
  required bool stackable,
  required int maxRedemptions,
  required int perUserLimit,
  required String passcodeHash,
  required DateTime startAt,
  required DateTime expiredAt,
}) {
  expect(_stringOrEmpty(coupon['address']), equals(address));

  final actualTokenAddress = _stringOrEmpty(coupon['token_address']);
  if (tokenAddress != null) {
    expect(actualTokenAddress, equals(tokenAddress));
  } else {
    expect(actualTokenAddress.isNotEmpty, isTrue);
  }

  expect(_stringOrEmpty(coupon['discount_type']), equals(discountType));
  expect(_stringOrEmpty(coupon['percentage_bps']), equals(percentageBPS));
  expect(_stringOrEmpty(coupon['fixed_amount']), equals(fixedAmount));
  expect(_stringOrEmpty(coupon['min_order']), equals(minOrder));
  expect(_boolOrFalse(coupon['paused']), equals(paused));
  expect(_boolOrFalse(coupon['stackable']), equals(stackable));
  expect(_intOrZero(coupon['max_redemptions']), equals(maxRedemptions));
  expect(_intOrZero(coupon['per_user_limit']), equals(perUserLimit));
  expect(_stringOrEmpty(coupon['passcode_hash']), equals(passcodeHash));

  _expectDateClose(coupon['start_at'] as String, startAt);
  _expectDateClose(coupon['expired_at'] as String, expiredAt);
}

void main() {
  group('CouponV2 E2E', () {
    test(
      'E2E: deploy coupon contract + addCoupon + updateCoupon + issue/redeem voucher + pause/unpause + get/list',
      () async {
        final c = await setupClient();

        addTearDown(() => teardownClient(c));

        final ownerUser = await newTestUser('owner');
        await c.setPrivateKey(ownerUser.privateKey);

        final deployed = await c.deployContract1(COUPON_CONTRACT_V2);

        expect(deployed, isA<ContractOutput>());
        expect(deployed.logs, isNotNull);
        expect(deployed.logs!, isNotEmpty);

        final couponAddress = deployed.logs!.first.contractAddress;
        expect(couponAddress, isNotEmpty);

        final startAt = DateTime.now().toUtc().add(const Duration(seconds: 1));
        final expiredAt = DateTime.now().toUtc().add(
          const Duration(minutes: 25),
        );

        final passcodeHash = sha256
            .convert(utf8.encode('e2e-passcode'))
            .toString();

        const discountType = 'percentage';
        const percentageBPS = '1000';
        const fixedAmount = '';
        const minOrder = '50';
        const paused = false;
        const stackable = true;
        const maxRedemptions = 100;
        const perUserLimit = 5;

        final voucherOwner = ownerUser.publicKey;
        final symbol = 'TST${randSuffix(4)}';
        const name = 'Test Token';
        const amount = '1000';
        const description = 'This is a test token';
        const image = 'https://example.com/image.png';
        const website = 'https://example.com';
        final tagsSocialMedia = <String, String>{
          'twitter': '@example',
          'discord': 'example#1234',
        };
        final tagsCategory = <String, String>{
          'type': 'utility',
          'industry': 'gaming',
        };
        final tags = <String, String>{'tag1': 'value1', 'tag2': 'value2'};
        const creator = 'Example Creator';
        const creatorWebsite = 'https://creator.example.com';
        const assetGlbUri = 'https://example.com/asset.glb';

        final outAddCoupon = await c.addCoupon(
          address: couponAddress,
          discountType: discountType,
          percentageBPS: percentageBPS,
          fixedAmount: fixedAmount,
          minOrder: minOrder,
          startAt: startAt,
          expiredAt: expiredAt,
          paused: paused,
          stackable: stackable,
          maxRedemptions: maxRedemptions,
          perUserLimit: perUserLimit,
          passcodeHash: passcodeHash,
          voucherOwner: voucherOwner,
          symbol: symbol,
          name: name,
          amount: amount,
          description: description,
          image: image,
          website: website,
          tagsSocialMedia: tagsSocialMedia,
          tagsCategory: tagsCategory,
          tags: tags,
          creator: creator,
          creatorWebsite: creatorWebsite,
          assetGlbUri: assetGlbUri,
        );

        expect(outAddCoupon, isA<ContractOutput>());
        expect(outAddCoupon.logs, isNotNull);
        expect(outAddCoupon.logs!, isNotEmpty);

        final addLog = outAddCoupon.logs!.first;
        expect(addLog.contractAddress, equals(couponAddress));
        expect(addLog.logType, equals('Coupon_Created'));

        final couponEvent = unmarshalEvent<JsonMap>(
          addLog.event,
          (json) => Map<String, dynamic>.from(json as Map),
        );

        _expectCouponSnapshot(
          couponEvent,
          address: couponAddress,
          discountType: discountType,
          percentageBPS: percentageBPS,
          fixedAmount: fixedAmount,
          minOrder: minOrder,
          paused: paused,
          stackable: stackable,
          maxRedemptions: maxRedemptions,
          perUserLimit: perUserLimit,
          passcodeHash: passcodeHash,
          startAt: startAt,
          expiredAt: expiredAt,
        );

        final couponTokenAddress = couponEvent['token_address'] as String;

        expect(outAddCoupon.delegatedCall, isNotNull);
        expect(outAddCoupon.delegatedCall!, isNotEmpty);
        expect(outAddCoupon.delegatedCall!.length, greaterThanOrEqualTo(2));

        final delegatedCreatedToken = outAddCoupon.delegatedCall![0];
        expect(delegatedCreatedToken.logs, isNotNull);
        expect(delegatedCreatedToken.logs!, isNotEmpty);
        expect(
          delegatedCreatedToken.logs!.first.logType,
          equals('Token_Created'),
        );

        final delegatedTokenTransfer = outAddCoupon.delegatedCall![1];
        expect(delegatedTokenTransfer.logs, isNotNull);
        expect(delegatedTokenTransfer.logs!, isNotEmpty);
        expect(
          delegatedTokenTransfer.logs!.first.logType,
          equals('Token_Transferred_NFT'),
        );

        final outBalance = await c.listTokenBalances(
          tokenAddress: couponTokenAddress,
          ownerAddress: voucherOwner,
          page: 1,
          limit: 3,
          ascending: true,
        );

        expect(outBalance.states, isNotNull);
        expect(outBalance.states!, isNotEmpty);

        final rawBalances = outBalance.states!.first.object;
        final balanceStates = List<Map<String, dynamic>>.from(
          (rawBalances as List).map((e) => Map<String, dynamic>.from(e as Map)),
        );

        expect(balanceStates, isNotEmpty);

        final firstBalance = balanceStates.first;
        expect(firstBalance['token_address'], equals(couponTokenAddress));
        expect(firstBalance['owner_address'], equals(voucherOwner));
        expect(firstBalance['amount'], equals('1'));
        expect(firstBalance['token_uuid'], isNotNull);
        expect((firstBalance['token_uuid'] as String).isNotEmpty, isTrue);

        final couponState = await _getCouponState(c, couponAddress);

        _expectCouponSnapshot(
          couponState,
          address: couponAddress,
          tokenAddress: couponTokenAddress,
          discountType: discountType,
          percentageBPS: percentageBPS,
          fixedAmount: fixedAmount,
          minOrder: minOrder,
          paused: paused,
          stackable: stackable,
          maxRedemptions: maxRedemptions,
          perUserLimit: perUserLimit,
          passcodeHash: passcodeHash,
          startAt: startAt,
          expiredAt: expiredAt,
        );

        // ------------------
        // UPDATE COUPON
        // ------------------
        final updatedStartAt = DateTime.now().toUtc().add(
          const Duration(seconds: 1),
        );
        final updatedExpiredAt = DateTime.now().toUtc().add(
          const Duration(minutes: 40),
        );

        final updatedPasscodeHash = sha256
            .convert(utf8.encode('e2e-passcode-2'))
            .toString();

        const updatedDiscountType = 'fixed-amount';
        const updatedPercentageBPS = '';
        const updatedFixedAmount = '250';
        const updatedMinOrder = '100';
        const updatedStackable = false;
        const updatedMaxRedemptions = 10;
        const updatedPerUserLimit = 1;

        final outUpdateCoupon = await c.updateCoupon(
          address: couponAddress,
          tokenAddress: couponTokenAddress,
          discountType: updatedDiscountType,
          percentageBPS: updatedPercentageBPS,
          fixedAmount: updatedFixedAmount,
          minOrder: updatedMinOrder,
          startAt: updatedStartAt,
          expiredAt: updatedExpiredAt,
          stackable: updatedStackable,
          maxRedemptions: updatedMaxRedemptions,
          perUserLimit: updatedPerUserLimit,
          passcodeHash: updatedPasscodeHash,
        );

        expect(outUpdateCoupon, isA<ContractOutput>());
        expect(outUpdateCoupon.logs, isNotNull);
        expect(outUpdateCoupon.logs!, isNotEmpty);

        final updateLog = outUpdateCoupon.logs!.first;
        expect(updateLog.contractAddress, equals(couponAddress));
        expect(updateLog.logType, equals('Coupon_Updated'));

        final updateEvent = unmarshalEvent<JsonMap>(
          updateLog.event,
          (json) => Map<String, dynamic>.from(json as Map),
        );

        _expectCouponSnapshot(
          updateEvent,
          address: couponAddress,
          tokenAddress: couponTokenAddress,
          discountType: updatedDiscountType,
          percentageBPS: updatedPercentageBPS,
          fixedAmount: updatedFixedAmount,
          minOrder: updatedMinOrder,
          paused: paused,
          stackable: updatedStackable,
          maxRedemptions: updatedMaxRedemptions,
          perUserLimit: updatedPerUserLimit,
          passcodeHash: updatedPasscodeHash,
          startAt: updatedStartAt,
          expiredAt: updatedExpiredAt,
        );

        final couponStateAfterUpdate = await _getCouponState(c, couponAddress);

        _expectCouponSnapshot(
          couponStateAfterUpdate,
          address: couponAddress,
          tokenAddress: couponTokenAddress,
          discountType: updatedDiscountType,
          percentageBPS: updatedPercentageBPS,
          fixedAmount: updatedFixedAmount,
          minOrder: updatedMinOrder,
          paused: paused,
          stackable: updatedStackable,
          maxRedemptions: updatedMaxRedemptions,
          perUserLimit: updatedPerUserLimit,
          passcodeHash: updatedPasscodeHash,
          startAt: updatedStartAt,
          expiredAt: updatedExpiredAt,
        );

        // ------------------
        // ISSUE VOUCHER
        // ------------------
        const issueAmount = '101';

        final outIssueVoucher = await c.issueVoucher(
          address: couponAddress,
          toAddress: ownerUser.publicKey,
          amount: issueAmount,
        );

        expect(outIssueVoucher, isA<ContractOutput>());
        expect(outIssueVoucher.logs, isNotNull);
        expect(outIssueVoucher.logs!, isNotEmpty);

        final issueLog = outIssueVoucher.logs!.first;
        expect(issueLog.contractAddress, equals(couponAddress));
        expect(issueLog.logType, equals('Voucher_Issued'));

        final issueEvent = unmarshalEvent<JsonMap>(
          issueLog.event,
          (json) => Map<String, dynamic>.from(json as Map),
        );

        expect(issueEvent['Address'], equals(couponAddress));
        expect(issueEvent['ToAddress'], equals(ownerUser.publicKey));
        expect(issueEvent['Amount'], equals(issueAmount));

        expect(outIssueVoucher.delegatedCall, isNotNull);
        expect(outIssueVoucher.delegatedCall!, isNotEmpty);

        final delegatedIssue = outIssueVoucher.delegatedCall!.first;
        expect(delegatedIssue.logs, isNotNull);
        expect(delegatedIssue.logs!, isNotEmpty);
        expect(delegatedIssue.logs!.first.logType, equals('Token_Minted_NFT'));

        final outVoucherBalances = await c.listTokenBalances(
          tokenAddress: couponTokenAddress,
          ownerAddress: ownerUser.publicKey,
          page: 1,
          limit: 10,
          ascending: true,
        );

        expect(outVoucherBalances.states, isNotNull);
        expect(outVoucherBalances.states!, isNotEmpty);

        final rawVoucherBalances = outVoucherBalances.states!.first.object;
        final voucherBalances = List<Map<String, dynamic>>.from(
          (rawVoucherBalances as List).map(
            (e) => Map<String, dynamic>.from(e as Map),
          ),
        );

        expect(voucherBalances, isNotEmpty);

        final issuedVoucher = voucherBalances.firstWhere(
          (balance) =>
              balance['token_address'] == couponTokenAddress &&
              balance['owner_address'] == ownerUser.publicKey,
        );

        expect(issuedVoucher['token_address'], equals(couponTokenAddress));
        expect(issuedVoucher['owner_address'], equals(ownerUser.publicKey));
        expect(issuedVoucher['amount'], equals('1'));
        expect(issuedVoucher['token_uuid'], isNotNull);
        expect((issuedVoucher['token_uuid'] as String).isNotEmpty, isTrue);

        final voucherUUID = issuedVoucher['token_uuid'] as String;

        await Future.delayed(const Duration(seconds: 1));

        // ------------------
        // REDEEM VOUCHER
        // ------------------
        await c.setPrivateKey(ownerUser.privateKey);

        final outRedeemVoucher = await c.redeemVoucher(
          address: couponAddress,
          orderAmount: '100',
          passcode: 'e2e-passcode-2',
          voucherUUID: voucherUUID,
        );

        expect(outRedeemVoucher, isA<ContractOutput>());
        expect(outRedeemVoucher.logs, isNotNull);
        expect(outRedeemVoucher.logs!, isNotEmpty);

        final redeemLog = outRedeemVoucher.logs!.first;
        expect(redeemLog.contractAddress, equals(couponAddress));
        expect(redeemLog.logType, equals('Voucher_Redeemed'));

        final redeemEvent = unmarshalEvent<JsonMap>(
          redeemLog.event,
          (json) => Map<String, dynamic>.from(json as Map),
        );
        expect(redeemEvent['coupon_address'], equals(couponAddress));
        expect(redeemEvent['token_address'], equals(couponTokenAddress));
        expect(redeemEvent['user_address'], equals(ownerUser.publicKey));
        expect(redeemEvent['order_amount'], equals('100'));
        expect(redeemEvent['voucher_uuid'], equals(voucherUUID));
        expect(redeemEvent['discount_amount'], isNotNull);
        expect(redeemEvent['discount_amount'].toString(), isNotEmpty);

        // ------------------
        // PAUSE COUPON
        // ------------------
        await c.setPrivateKey(ownerUser.privateKey);

        final outPauseCoupon = await c.pauseCoupon(
          address: couponAddress,
          pause: true,
        );

        expect(outPauseCoupon, isA<ContractOutput>());
        expect(outPauseCoupon.logs, isNotNull);
        expect(outPauseCoupon.logs!, isNotEmpty);

        final pauseLog = outPauseCoupon.logs!.first;
        expect(pauseLog.contractAddress, equals(couponAddress));
        expect(pauseLog.logType, equals('Coupon_Paused'));

        final pauseEvent = unmarshalEvent<JsonMap>(
          pauseLog.event,
          (json) => Map<String, dynamic>.from(json as Map),
        );

        expect(pauseEvent['address'], equals(couponAddress));
        expect(pauseEvent['paused'], equals(true));

        // ------------------
        // UNPAUSE COUPON
        // ------------------
        final outUnpauseCoupon = await c.unpauseCoupon(
          address: couponAddress,
          pause: false,
        );

        expect(outUnpauseCoupon, isA<ContractOutput>());
        expect(outUnpauseCoupon.logs, isNotNull);
        expect(outUnpauseCoupon.logs!, isNotEmpty);

        final unpauseLog = outUnpauseCoupon.logs!.first;
        expect(unpauseLog.contractAddress, equals(couponAddress));
        expect(unpauseLog.logType, equals('Coupon_Unpaused'));

        final unpauseEvent = unmarshalEvent<JsonMap>(
          unpauseLog.event,
          (json) => Map<String, dynamic>.from(json as Map),
        );

        expect(unpauseEvent['address'], equals(couponAddress));

        // ------------------
        // GET COUPON
        // ------------------
        final finalCouponState = await _getCouponState(c, couponAddress);

        _expectCouponSnapshot(
          finalCouponState,
          address: couponAddress,
          tokenAddress: couponTokenAddress,
          discountType: updatedDiscountType,
          percentageBPS: updatedPercentageBPS,
          fixedAmount: updatedFixedAmount,
          minOrder: updatedMinOrder,
          paused: false,
          stackable: updatedStackable,
          maxRedemptions: updatedMaxRedemptions,
          perUserLimit: updatedPerUserLimit,
          passcodeHash: updatedPasscodeHash,
          startAt: updatedStartAt,
          expiredAt: updatedExpiredAt,
        );

        expect(finalCouponState['total_redemptions'], equals(1));
        expect(finalCouponState['created_at'], isNotNull);
        expect(finalCouponState['updated_at'], isNotNull);

        // ------------------
        // LIST COUPONS
        // ------------------
        final outListCoupons = await c.listCoupons(
          owner: '',
          tokenAddress: couponTokenAddress,
          programType: '',
          paused: null,
          page: 1,
          limit: 10,
          ascending: true,
        );

        expect(outListCoupons.states, isNotNull);
        expect(outListCoupons.states!, isNotEmpty);

        final rawCouponsList = outListCoupons.states!.first.object;
        final couponsList = List<Map<String, dynamic>>.from(
          (rawCouponsList as List).map(
            (e) => Map<String, dynamic>.from(e as Map),
          ),
        );

        expect(couponsList, isNotEmpty);

        final listedCoupon = couponsList.firstWhere(
          (coupon) => coupon['address'] == couponAddress,
          orElse: () => <String, dynamic>{},
        );

        expect(listedCoupon, isNotEmpty);

        _expectCouponSnapshot(
          listedCoupon,
          address: couponAddress,
          tokenAddress: couponTokenAddress,
          discountType: updatedDiscountType,
          percentageBPS: updatedPercentageBPS,
          fixedAmount: updatedFixedAmount,
          minOrder: updatedMinOrder,
          paused: false,
          stackable: updatedStackable,
          maxRedemptions: updatedMaxRedemptions,
          perUserLimit: updatedPerUserLimit,
          passcodeHash: updatedPasscodeHash,
          startAt: updatedStartAt,
          expiredAt: updatedExpiredAt,
        );

        expect(listedCoupon['total_redemptions'], equals(1));
        expect(listedCoupon['created_at'], isNotNull);
        expect(listedCoupon['updated_at'], isNotNull);
      },
    );
  });
}
