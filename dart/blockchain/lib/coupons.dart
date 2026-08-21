part of 'two_finance_blockchain.dart';

extension CouponClient on TwoFinanceBlockchain {
  Future<ContractOutput> addCoupon({
    required String address,
    required String discountType,
    required String percentageBPS,
    required String fixedAmount,
    required String minOrder,
    required DateTime startAt,
    required DateTime expiredAt,
    required bool paused,
    required bool stackable,
    required int maxRedemptions,
    required int perUserLimit,
    required String passcodeHash,
    required String voucherOwner,
    required String symbol,
    required String name,
    required String amount,
    required String description,
    required String image,
    required String website,
    required Map<String, String> tagsSocialMedia,
    required Map<String, String> tagsCategory,
    required Map<String, String> tags,
    required String creator,
    required String creatorWebsite,
    required String assetGlbUri,
  }) async {
    final from = publicKeyHex ?? '';
    final to = address;
    const int version = 1;
    final uuid7 = newUUID7();
    const method = METHOD_ADD_COUPON;

    if (from.isEmpty) throw ArgumentError('from address not set');
    KeyManager.validateEDDSAPublicKeyHex(from);

    if (address.isEmpty) throw ArgumentError('address not set');
    KeyManager.validateEDDSAPublicKeyHex(address);

    if (!(discountType == 'percentage' || discountType == 'fixed-amount')) {
      throw ArgumentError('invalid discount_type: $discountType');
    }

    if (discountType == 'percentage' && percentageBPS.isEmpty) {
      throw ArgumentError(
        'percentage_bps must be set for discount_type=percentage',
      );
    }

    if (discountType == 'fixed-amount' && fixedAmount.isEmpty) {
      throw ArgumentError(
        'fixed_amount must be set for discount_type=fixed-amount',
      );
    }

    if (voucherOwner.isEmpty) throw ArgumentError('voucherOwner must be set');
    if (symbol.isEmpty) throw ArgumentError('symbol must be set');
    if (name.isEmpty) throw ArgumentError('name must be set');
    if (amount.isEmpty) throw ArgumentError('amount must be set');
    if (description.isEmpty) throw ArgumentError('description must be set');
    if (image.isEmpty) throw ArgumentError('image must be set');
    if (website.isEmpty) throw ArgumentError('website must be set');
    if (tagsSocialMedia.isEmpty) {
      throw ArgumentError('tagsSocialMedia must be set');
    }
    if (tagsCategory.isEmpty) throw ArgumentError('tagsCategory must be set');
    if (tags.isEmpty) throw ArgumentError('tags must be set');
    if (creator.isEmpty) throw ArgumentError('creator must be set');
    if (creatorWebsite.isEmpty) {
      throw ArgumentError('creatorWebsite must be set');
    }
    if (assetGlbUri.isEmpty) throw ArgumentError('assetGlbUri must be set');

    final JsonMessage data = {
      'address': address,
      'discount_type': discountType,
      'percentage_bps': percentageBPS,
      'fixed_amount': fixedAmount,
      'min_order': minOrder,
      'start_at': startAt.toIso8601String(),
      'expired_at': expiredAt.toIso8601String(),
      'paused': paused,
      'stackable': stackable,
      'max_redemptions': maxRedemptions,
      'per_user_limit': perUserLimit,
      'passcode_hash': passcodeHash,
      'voucher_owner': voucherOwner,
      'symbol': symbol,
      'name': name,
      'amount': amount,
      'description': description,
      'image': image,
      'website': website,
      'tags_social_media': tagsSocialMedia,
      'tags_category': tagsCategory,
      'tags': tags,
      'creator': creator,
      'creator_website': creatorWebsite,
      'asset_glb_uri': assetGlbUri,
    };

    return signAndSendTransaction(
      chainID: _chainID,
      from: from,
      to: to,
      method: method,
      data: data,
      version: version,
      uuid7: uuid7,
    );
  }

  Future<ContractOutput> updateCoupon({
    required String address,
    required String tokenAddress,
    required String discountType,
    required String percentageBPS,
    required String fixedAmount,
    required String minOrder,
    required DateTime startAt,
    required DateTime expiredAt,
    required bool stackable,
    required int maxRedemptions,
    required int perUserLimit,
    required String passcodeHash,
  }) async {
    if (address.isEmpty) throw ArgumentError('address not set');
    KeyManager.validateEDDSAPublicKeyHex(address);

    if (tokenAddress.isNotEmpty) {
      KeyManager.validateEDDSAPublicKeyHex(tokenAddress);
    }

    if (discountType.isNotEmpty &&
        !(discountType == 'percentage' || discountType == 'fixed-amount')) {
      throw ArgumentError('invalid discount_type: $discountType');
    }

    final from = publicKeyHex ?? '';
    final to = address;
    const int version = 1;
    final uuid7 = newUUID7();
    const method = METHOD_UPDATE_COUPON;

    if (from.isEmpty) throw ArgumentError('from address not set');
    KeyManager.validateEDDSAPublicKeyHex(from);

    final JsonMessage data = {
      'address': address,
      'token_address': tokenAddress,
      'discount_type': discountType,
      'percentage_bps': percentageBPS,
      'fixed_amount': fixedAmount,
      'min_order': minOrder,
      'start_at': startAt.toIso8601String(),
      'expired_at': expiredAt.toIso8601String(),
      'stackable': stackable,
      'max_redemptions': maxRedemptions,
      'per_user_limit': perUserLimit,
      'passcode_hash': passcodeHash,
    };

    return signAndSendTransaction(
      chainID: _chainID,
      from: from,
      to: to,
      method: method,
      data: data,
      version: version,
      uuid7: uuid7,
    );
  }

  Future<ContractOutput> issueVoucher({
    required String address,
    required String toAddress,
    required String amount,
  }) async {
    if (address.isEmpty) throw ArgumentError('address not set');
    KeyManager.validateEDDSAPublicKeyHex(address);

    if (toAddress.isEmpty) throw ArgumentError('to_address not set');
    KeyManager.validateEDDSAPublicKeyHex(toAddress);

    if (amount.isEmpty) throw ArgumentError('amount not set');

    final from = publicKeyHex ?? '';
    if (from.isEmpty) throw ArgumentError('from address not set');
    KeyManager.validateEDDSAPublicKeyHex(from);

    final to = address;
    const int version = 1;
    final uuid7 = newUUID7();
    const method = METHOD_ISSUE_VOUCHER;

    final JsonMessage data = {
      'address': address,
      'to_address': toAddress,
      'amount': amount,
    };

    return signAndSendTransaction(
      chainID: _chainID,
      from: from,
      to: to,
      method: method,
      data: data,
      version: version,
      uuid7: uuid7,
    );
  }

  Future<ContractOutput> redeemVoucher({
    required String address,
    required String orderAmount,
    required String passcode,
    required String voucherUUID,
  }) async {
    if (address.isEmpty) throw ArgumentError('address not set');
    KeyManager.validateEDDSAPublicKeyHex(address);

    if (orderAmount.isEmpty) throw ArgumentError('order_amount not set');
    if (passcode.isEmpty) {
      throw ArgumentError('passcode (preimage) not set');
    }
    if (voucherUUID.isEmpty) {
      throw ArgumentError('voucher_uuid must be set for non-fungible tokens');
    }

    final from = publicKeyHex ?? '';
    if (from.isEmpty) throw ArgumentError('from address not set');
    KeyManager.validateEDDSAPublicKeyHex(from);

    final to = address;
    const int version = 1;
    final uuid7 = newUUID7();
    const method = METHOD_REDEEM_VOUCHER;

    final JsonMessage data = {
      'address': address,
      'order_amount': orderAmount,
      'passcode': passcode,
      'voucher_uuid': voucherUUID,
    };

    return signAndSendTransaction(
      chainID: _chainID,
      from: from,
      to: to,
      method: method,
      data: data,
      version: version,
      uuid7: uuid7,
    );
  }

  Future<ContractOutput> pauseCoupon({
    required String address,
    required bool pause,
  }) async {
    if (address.isEmpty) throw ArgumentError('address not set');
    KeyManager.validateEDDSAPublicKeyHex(address);

    if (!pause) throw ArgumentError('pause must be true: paused=$pause');

    final from = publicKeyHex ?? '';
    if (from.isEmpty) throw ArgumentError('from address not set');
    KeyManager.validateEDDSAPublicKeyHex(from);

    final to = address;
    const int version = 1;
    final uuid7 = newUUID7();
    const method = METHOD_PAUSE_COUPON;

    final JsonMessage data = {'address': address, 'paused': pause};

    return signAndSendTransaction(
      chainID: _chainID,
      from: from,
      to: to,
      method: method,
      data: data,
      version: version,
      uuid7: uuid7,
    );
  }

  Future<ContractOutput> unpauseCoupon({
    required String address,
    required bool pause,
  }) async {
    if (address.isEmpty) throw ArgumentError('address not set');
    KeyManager.validateEDDSAPublicKeyHex(address);

    if (pause) throw ArgumentError('pause must be false: paused=$pause');

    final from = publicKeyHex ?? '';
    if (from.isEmpty) throw ArgumentError('from address not set');
    KeyManager.validateEDDSAPublicKeyHex(from);

    final to = address;
    const int version = 1;
    final uuid7 = newUUID7();
    const method = METHOD_UNPAUSE_COUPON;

    final JsonMessage data = {'address': address, 'paused': pause};

    return signAndSendTransaction(
      chainID: _chainID,
      from: from,
      to: to,
      method: method,
      data: data,
      version: version,
      uuid7: uuid7,
    );
  }

  Future<ContractOutput> getCoupon({required String address}) async {
    final from = publicKeyHex ?? '';
    if (from.isEmpty) throw ArgumentError('from address not set');
    KeyManager.validateEDDSAPublicKeyHex(from);

    if (address.isEmpty) throw ArgumentError('coupon address must be set');
    KeyManager.validateEDDSAPublicKeyHex(address);

    const method = METHOD_GET_COUPON;

    return getState(to: address, method: method, data: {'address': address});
  }

  Future<ContractOutput> listCoupons({
    required String owner,
    required String tokenAddress,
    required String programType,
    bool? paused,
    required int page,
    required int limit,
    required bool ascending,
  }) async {
    final from = publicKeyHex ?? '';
    if (from.isEmpty) throw ArgumentError('from address not set');
    KeyManager.validateEDDSAPublicKeyHex(from);

    if (owner.isNotEmpty) {
      KeyManager.validateEDDSAPublicKeyHex(owner);
    }

    if (tokenAddress.isNotEmpty) {
      KeyManager.validateEDDSAPublicKeyHex(tokenAddress);
    }

    if (programType.isNotEmpty &&
        !(programType == 'percentage' || programType == 'fixed-amount')) {
      throw ArgumentError('invalid program_type: $programType');
    }

    if (page < 1) throw ArgumentError('page must be greater than 0');
    if (limit < 1) throw ArgumentError('limit must be greater than 0');

    const method = METHOD_LIST_COUPONS;

    final JsonMessage data = {
      'owner': owner,
      'token_address': tokenAddress,
      'program_type': programType,
      'paused': paused,
      'page': page,
      'limit': limit,
      'ascending': ascending,
      'contract_version': COUPON_CONTRACT_V2,
    };

    return getState(to: '', method: method, data: data);
  }
}
