part of 'two_finance_blockchain.dart';

extension CashbackClient on TwoFinanceBlockchain {
  static const String _programTypeFixed = 'fixed-percentage';
  static const String _programTypeVariable = 'variable-percentage';

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

  void _validateProgramType(String programType) {
    if (programType != _programTypeFixed &&
        programType != _programTypeVariable) {
      throw ArgumentError('invalid program_type: $programType');
    }
  }

  Future<ContractOutput> addCashback({
    required String address,
    required String owner,
    required String tokenAddress,
    required String programType,
    required String percentage,
    required DateTime startAt,
    required DateTime expiredAt,
    required bool paused,
  }) async {
    final from = _requireFromAddress();

    if (address.isEmpty) throw ArgumentError('address not set');
    if (owner.isEmpty) throw ArgumentError('owner not set');
    if (tokenAddress.isEmpty) throw ArgumentError('token address not set');
    if (percentage.isEmpty) throw ArgumentError('percentage not set');
    if (_isZeroLikeDateTime(startAt)) {
      throw ArgumentError('start_at not set');
    }
    if (_isZeroLikeDateTime(expiredAt)) {
      throw ArgumentError('expired_at not set');
    }

    KeyManager.validateEDDSAPublicKeyHex(address);
    KeyManager.validateEDDSAPublicKeyHex(owner);
    KeyManager.validateEDDSAPublicKeyHex(tokenAddress);
    _validateProgramType(programType);

    if (owner != from) {
      throw ArgumentError('owner must match from address');
    }

    return signAndSendTransaction(
      chainID: _chainID,
      from: from,
      to: address,
      method: METHOD_ADD_CASHBACK,
      data: {
        'address': address,
        'owner': owner,
        'token_address': tokenAddress,
        'program_type': programType,
        'percentage': percentage,
        'start_at': startAt.toUtc().toIso8601String(),
        'expired_at': expiredAt.toUtc().toIso8601String(),
        'paused': paused,
      },
      version: 1,
      uuid7: newUUID7(),
    );
  }

  Future<ContractOutput> updateCashback({
    required String address,
    String tokenAddress = '',
    required String programType,
    required String percentage,
    required DateTime startAt,
    required DateTime expiredAt,
  }) async {
    final from = _requireFromAddress();

    if (address.isEmpty) throw ArgumentError('address not set');
    if (percentage.isEmpty) throw ArgumentError('percentage not set');
    if (_isZeroLikeDateTime(startAt)) {
      throw ArgumentError('start_at not set');
    }
    if (_isZeroLikeDateTime(expiredAt)) {
      throw ArgumentError('expired_at not set');
    }

    KeyManager.validateEDDSAPublicKeyHex(address);
    if (tokenAddress.isNotEmpty) {
      KeyManager.validateEDDSAPublicKeyHex(tokenAddress);
    }
    _validateProgramType(programType);

    return signAndSendTransaction(
      chainID: _chainID,
      from: from,
      to: address,
      method: METHOD_UPDATE_CASHBACK,
      data: {
        'address': address,
        'token_address': tokenAddress,
        'program_type': programType,
        'percentage': percentage,
        'start_at': startAt.toUtc().toIso8601String(),
        'expired_at': expiredAt.toUtc().toIso8601String(),
      },
      version: 1,
      uuid7: newUUID7(),
    );
  }

  Future<ContractOutput> pauseCashback({
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
      method: METHOD_PAUSE_CASHBACK,
      data: {'address': address, 'paused': paused},
      version: 1,
      uuid7: newUUID7(),
    );
  }

  Future<ContractOutput> unpauseCashback({
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
      method: METHOD_UNPAUSE_CASHBACK,
      data: {'address': address, 'paused': paused},
      version: 1,
      uuid7: newUUID7(),
    );
  }

  Future<ContractOutput> depositCashbackFunds({
    required String address,
    required String tokenAddress,
    required String amount,
    String tokenType = '',
    String uuid = '',
  }) async {
    final from = _requireFromAddress();

    if (address.isEmpty) throw ArgumentError('address not set');
    if (tokenAddress.isEmpty) throw ArgumentError('token address not set');
    if (amount.isEmpty) throw ArgumentError('amount not set');
    final effectiveTokenType = tokenType.isEmpty
        ? TOKEN_TYPE_FUNGIBLE
        : tokenType;
    if (effectiveTokenType == TOKEN_TYPE_NON_FUNGIBLE && uuid.isEmpty) {
      throw ArgumentError('uuid must be set for non-fungible tokens');
    }

    KeyManager.validateEDDSAPublicKeyHex(address);
    KeyManager.validateEDDSAPublicKeyHex(tokenAddress);

    return signAndSendTransaction(
      chainID: _chainID,
      from: from,
      to: address,
      method: METHOD_DEPOSIT_CASHBACK,
      data: {
        'address': address,
        'token_address': tokenAddress,
        'amount': amount,
        'token_type': effectiveTokenType,
        'uuid': uuid,
      },
      version: 1,
      uuid7: newUUID7(),
    );
  }

  Future<ContractOutput> withdrawCashbackFunds({
    required String address,
    required String tokenAddress,
    required String amount,
    String tokenType = '',
    String uuid = '',
  }) async {
    final from = _requireFromAddress();

    if (address.isEmpty) throw ArgumentError('address not set');
    if (tokenAddress.isEmpty) throw ArgumentError('token address not set');
    if (amount.isEmpty) throw ArgumentError('amount not set');
    final effectiveTokenType = tokenType.isEmpty
        ? TOKEN_TYPE_FUNGIBLE
        : tokenType;
    if (effectiveTokenType == TOKEN_TYPE_NON_FUNGIBLE && uuid.isEmpty) {
      throw ArgumentError('uuid must be set for non-fungible tokens');
    }

    KeyManager.validateEDDSAPublicKeyHex(address);
    KeyManager.validateEDDSAPublicKeyHex(tokenAddress);

    return signAndSendTransaction(
      chainID: _chainID,
      from: from,
      to: address,
      method: METHOD_WITHDRAW_CASHBACK,
      data: {
        'address': address,
        'token_address': tokenAddress,
        'amount': amount,
        'token_type': effectiveTokenType,
        'uuid': uuid,
      },
      version: 1,
      uuid7: newUUID7(),
    );
  }

  Future<ContractOutput> getCashback({required String address}) async {
    final from = _requireFromAddress();

    if (address.isEmpty) throw ArgumentError('cashback address must be set');
    KeyManager.validateEDDSAPublicKeyHex(address);
    KeyManager.validateEDDSAPublicKeyHex(from);

    return getState(
      to: address,
      method: METHOD_GET_CASHBACK,
      data: const <String, dynamic>{},
    );
  }

  Future<ContractOutput> listCashbacks({
    String owner = '',
    String tokenAddress = '',
    String programType = '',
    required bool paused,
    int page = 1,
    int limit = 10,
    bool ascending = false,
  }) async {
    final from = _requireFromAddress();
    KeyManager.validateEDDSAPublicKeyHex(from);

    if (owner.isNotEmpty) {
      KeyManager.validateEDDSAPublicKeyHex(owner);
    }
    if (tokenAddress.isNotEmpty) {
      KeyManager.validateEDDSAPublicKeyHex(tokenAddress);
    }
    if (programType.isNotEmpty) {
      _validateProgramType(programType);
    }
    if (page < 1) throw ArgumentError('page must be greater than 0');
    if (limit < 1) throw ArgumentError('limit must be greater than 0');

    return getState(
      to: '',
      method: METHOD_LIST_CASHBACKS,
      data: {
        'owner': owner,
        'program_type': programType,
        'paused': paused,
        'page': page,
        'limit': limit,
        'ascending': ascending,
        'token_address': tokenAddress,
        'contract_version': CASHBACK_CONTRACT_V2,
      },
    );
  }

  Future<ContractOutput> claimCashback({
    required String address,
    required String amount,
    String tokenType = '',
    String uuid = '',
  }) async {
    final from = _requireFromAddress();

    if (address.isEmpty) throw ArgumentError('address not set');
    if (amount.isEmpty) throw ArgumentError('amount not set');
    final effectiveTokenType = tokenType.isEmpty
        ? TOKEN_TYPE_FUNGIBLE
        : tokenType;
    if (effectiveTokenType == TOKEN_TYPE_NON_FUNGIBLE && uuid.isEmpty) {
      throw ArgumentError('uuid must be set for non-fungible tokens');
    }

    KeyManager.validateEDDSAPublicKeyHex(address);

    return signAndSendTransaction(
      chainID: _chainID,
      from: from,
      to: address,
      method: METHOD_CLAIM_CASHBACK,
      data: {
        'address': address,
        'amount': amount,
        'token_type': effectiveTokenType,
        'uuid': uuid,
      },
      version: 1,
      uuid7: newUUID7(),
    );
  }
}
