part of two_finance_blockchain;

extension PaymentClient on TwoFinanceBlockchain {
  bool _isZeroLikeDateTime(DateTime value) {
    final isUnixEpoch = value.millisecondsSinceEpoch == 0;
    final isGoZeroLike =
        value.year <= 1 &&
        value.month == 1 &&
        value.day == 1 &&
        value.hour == 0 &&
        value.minute == 0 &&
        value.second == 0 &&
        value.millisecond == 0 &&
        value.microsecond == 0;

    return isUnixEpoch || isGoZeroLike;
  }

  String _requireFromAddress() {
    final from = publicKeyHex ?? '';
    if (from.isEmpty) {
      throw ArgumentError('from address not set');
    }
    KeyManager.validateEDDSAPublicKeyHex(from);
    return from;
  }

  Future<ContractOutput> createPayment({
    required String address,
    required String owner,
    required String tokenAddress,
    required String orderId,
    required String payer,
    required String payee,
    required String amount,
    required DateTime expiredAt,
  }) async {
    final from = _requireFromAddress();

    if (address.isEmpty) throw ArgumentError('address not set');
    if (owner.isEmpty) throw ArgumentError('owner not set');
    if (tokenAddress.isEmpty) throw ArgumentError('token address not set');
    if (payer.isEmpty) throw ArgumentError('payer not set');
    if (payee.isEmpty) throw ArgumentError('payee not set');
    if (payer == payee) {
      throw ArgumentError(
        'payee and payer cannot be the same: $payee - $payer',
      );
    }
    if (orderId.isEmpty) throw ArgumentError('order_id not set');
    if (amount.isEmpty) throw ArgumentError('amount not set');
    if (_isZeroLikeDateTime(expiredAt)) {
      throw ArgumentError('expired_at not set');
    }

    KeyManager.validateEDDSAPublicKeyHex(address);
    KeyManager.validateEDDSAPublicKeyHex(owner);
    KeyManager.validateEDDSAPublicKeyHex(tokenAddress);
    KeyManager.validateEDDSAPublicKeyHex(payer);
    KeyManager.validateEDDSAPublicKeyHex(payee);

    return signAndSendTransaction(
      chainID: _chainID,
      from: from,
      to: address,
      method: METHOD_CREATE_PAYMENT,
      data: {
        'address': address,
        'owner': owner,
        'token_address': tokenAddress,
        'order_id': orderId,
        'payer': payer,
        'payee': payee,
        'amount': amount,
        'expired_at': expiredAt.toUtc().toIso8601String(),
      },
      version: 1,
      uuid7: newUUID7(),
    );
  }

  Future<ContractOutput> directPay({
    required String address,
    required String owner,
    required String tokenAddress,
    required String orderId,
    required String payer,
    required String payee,
    required String amount,
    required DateTime expiredAt,
  }) async {
    final from = _requireFromAddress();

    if (address.isEmpty) throw ArgumentError('address not set');
    if (owner.isEmpty) throw ArgumentError('owner not set');
    if (tokenAddress.isEmpty) throw ArgumentError('token address not set');
    if (payer.isEmpty) throw ArgumentError('payer not set');
    if (payee.isEmpty) throw ArgumentError('payee not set');
    if (payer == payee) {
      throw ArgumentError(
        'payee and payer cannot be the same: $payee - $payer',
      );
    }
    if (orderId.isEmpty) throw ArgumentError('order_id not set');
    if (amount.isEmpty) throw ArgumentError('amount not set');
    if (_isZeroLikeDateTime(expiredAt)) {
      throw ArgumentError('expired_at not set');
    }

    KeyManager.validateEDDSAPublicKeyHex(address);
    KeyManager.validateEDDSAPublicKeyHex(owner);
    KeyManager.validateEDDSAPublicKeyHex(tokenAddress);
    KeyManager.validateEDDSAPublicKeyHex(payer);
    KeyManager.validateEDDSAPublicKeyHex(payee);

    return signAndSendTransaction(
      chainID: _chainID,
      from: from,
      to: address,
      method: METHOD_DIRECT_PAY,
      data: {
        'address': address,
        'owner': owner,
        'token_address': tokenAddress,
        'order_id': orderId,
        'payer': payer,
        'payee': payee,
        'amount': amount,
        'expired_at': expiredAt.toUtc().toIso8601String(),
      },
      version: 1,
      uuid7: newUUID7(),
    );
  }

  Future<ContractOutput> authorizePayment({required String address}) async {
    final from = _requireFromAddress();

    if (address.isEmpty) throw ArgumentError('address not set');
    KeyManager.validateEDDSAPublicKeyHex(address);

    return signAndSendTransaction(
      chainID: _chainID,
      from: from,
      to: address,
      method: METHOD_AUTHORIZE_PAYMENT,
      data: {'address': address},
      version: 1,
      uuid7: newUUID7(),
    );
  }

  Future<ContractOutput> capturePayment({required String address}) async {
    final from = _requireFromAddress();

    if (address.isEmpty) throw ArgumentError('address not set');
    KeyManager.validateEDDSAPublicKeyHex(address);

    return signAndSendTransaction(
      chainID: _chainID,
      from: from,
      to: address,
      method: METHOD_CAPTURE_PAYMENT,
      data: {'address': address},
      version: 1,
      uuid7: newUUID7(),
    );
  }

  Future<ContractOutput> refundPayment({
    required String address,
    required String amount,
  }) async {
    final from = _requireFromAddress();

    if (address.isEmpty) throw ArgumentError('address not set');
    if (amount.isEmpty) throw ArgumentError('amount not set');
    KeyManager.validateEDDSAPublicKeyHex(address);

    return signAndSendTransaction(
      chainID: _chainID,
      from: from,
      to: address,
      method: METHOD_REFUND_PAYMENT,
      data: {'address': address, 'amount': amount},
      version: 1,
      uuid7: newUUID7(),
    );
  }

  Future<ContractOutput> voidPayment({required String address}) async {
    final from = _requireFromAddress();

    if (address.isEmpty) throw ArgumentError('address not set');
    KeyManager.validateEDDSAPublicKeyHex(address);

    return signAndSendTransaction(
      chainID: _chainID,
      from: from,
      to: address,
      method: METHOD_VOID_PAYMENT,
      data: {'address': address},
      version: 1,
      uuid7: newUUID7(),
    );
  }

  Future<ContractOutput> pausePayment({
    required String address,
    required bool paused,
  }) async {
    final from = _requireFromAddress();

    if (address.isEmpty) throw ArgumentError('address not set');
    if (!paused) throw ArgumentError('paused must be true: Pause: $paused');
    KeyManager.validateEDDSAPublicKeyHex(address);

    return signAndSendTransaction(
      chainID: _chainID,
      from: from,
      to: address,
      method: METHOD_PAUSE_PAYMENT,
      data: {'address': address, 'paused': paused},
      version: 1,
      uuid7: newUUID7(),
    );
  }

  Future<ContractOutput> unpausePayment({
    required String address,
    required bool paused,
  }) async {
    final from = _requireFromAddress();

    if (address.isEmpty) throw ArgumentError('address not set');
    if (paused) throw ArgumentError('paused must be false: Pause: $paused');
    KeyManager.validateEDDSAPublicKeyHex(address);

    return signAndSendTransaction(
      chainID: _chainID,
      from: from,
      to: address,
      method: METHOD_UNPAUSE_PAYMENT,
      data: {'address': address, 'paused': paused},
      version: 1,
      uuid7: newUUID7(),
    );
  }

  Future<ContractOutput> getPayment({required String address}) async {
    final from = _requireFromAddress();

    if (address.isEmpty) throw ArgumentError('payment address must be set');
    KeyManager.validateEDDSAPublicKeyHex(address);
    KeyManager.validateEDDSAPublicKeyHex(from);

    return getState(
      to: address,
      method: METHOD_GET_PAYMENT,
      data: {'address': address},
    );
  }

  Future<ContractOutput> listPayments({
    String orderId = '',
    String tokenAddress = '',
    List<String> status = const [],
    String payer = '',
    String payee = '',
    int page = 1,
    int limit = 10,
    bool ascending = false,
  }) async {
    final from = _requireFromAddress();
    KeyManager.validateEDDSAPublicKeyHex(from);

    if (tokenAddress.isNotEmpty) {
      KeyManager.validateEDDSAPublicKeyHex(tokenAddress);
    }
    if (payer.isNotEmpty) {
      KeyManager.validateEDDSAPublicKeyHex(payer);
    }
    if (payee.isNotEmpty) {
      KeyManager.validateEDDSAPublicKeyHex(payee);
    }
    if (page < 1) throw ArgumentError('page must be greater than 0');
    if (limit < 1) throw ArgumentError('limit must be greater than 0');

    return getState(
      to: '',
      method: METHOD_LIST_PAYMENTS,
      data: {
        'order_id': orderId,
        'token_address': tokenAddress,
        'status': status,
        'payer': payer,
        'payee': payee,
        'page': page,
        'limit': limit,
        'ascending': ascending,
        'contract_version': PAYMENT_CONTRACT_V2,
      },
    );
  }
}
