part of two_finance_blockchain;

extension RaffleClient on TwoFinanceBlockchain {
  static const int _txVersion = 1;

  void _validatePublicKeyHex(String value, String label) {
    try {
      KeyManager.validateEDDSAPublicKeyHex(value);
    } catch (e) {
      throw ArgumentError('invalid $label: $e');
    }
  }

  String _requireFromAddress() {
    final from = publicKeyHex ?? '';
    if (from.isEmpty) {
      throw ArgumentError('from address not set');
    }

    _validatePublicKeyHex(from, 'from address');
    return from;
  }

  String _newUuid7OrThrow() {
    try {
      return newUUID7();
    } catch (e) {
      throw StateError('failed to generate UUIDv7: $e');
    }
  }

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

  Future<ContractOutput> addRaffle({
    required String address,
    required String owner,
    required String tokenAddress,
    required String ticketPrice,
    required int maxEntries,
    required int maxEntriesPerUser,
    required DateTime startAt,
    required DateTime expiredAt,
    bool paused = false,
    String seedCommitHex = '',
    Map<String, String> metadata = const <String, String>{},
  }) async {
    final from = _requireFromAddress();

    if (address.isEmpty) {
      throw ArgumentError('address not set');
    }
    _validatePublicKeyHex(address, 'raffle address');

    if (owner.isEmpty) {
      throw ArgumentError('owner not set');
    }
    _validatePublicKeyHex(owner, 'owner address');

    if (tokenAddress.isEmpty) {
      throw ArgumentError('token address not set');
    }
    _validatePublicKeyHex(tokenAddress, 'token address');

    if (ticketPrice.isEmpty) {
      throw ArgumentError('ticket_price not set');
    }
    if (maxEntries <= 0) {
      throw ArgumentError('max_entries must be > 0');
    }
    if (maxEntriesPerUser <= 0) {
      throw ArgumentError('max_entries_per_user must be > 0');
    }
    if (maxEntriesPerUser > maxEntries) {
      throw ArgumentError('max_entries_per_user cannot exceed max_entries');
    }
    if (_isZeroLikeDateTime(startAt)) {
      throw ArgumentError('start_at not set');
    }
    if (_isZeroLikeDateTime(expiredAt)) {
      throw ArgumentError('expired_at not set');
    }

    final JsonMessage data = {
      'address': address,
      'owner': owner,
      'token_address': tokenAddress,
      'ticket_price': ticketPrice,
      'max_entries': maxEntries,
      'max_entries_per_user': maxEntriesPerUser,
      'start_at': startAt.toIso8601String(),
      'expired_at': expiredAt.toIso8601String(),
      'paused': paused,
      'seed_commit_hex': seedCommitHex,
      'metadata': metadata,
    };

    return signAndSendTransaction(
      chainID: _chainID,
      from: from,
      to: address,
      method: METHOD_ADD_RAFFLE,
      data: data,
      version: _txVersion,
      uuid7: _newUuid7OrThrow(),
    );
  }

  Future<ContractOutput> updateRaffle({
    required String address,
    String tokenAddress = '',
    String ticketPrice = '',
    int maxEntries = 0,
    int maxEntriesPerUser = 0,
    DateTime? startAt,
    DateTime? expiredAt,
    String seedCommitHex = '',
    Map<String, String> metadata = const <String, String>{},
  }) async {
    if (address.isEmpty) {
      throw ArgumentError('address not set');
    }
    _validatePublicKeyHex(address, 'address');

    final from = _requireFromAddress();

    if (tokenAddress.isNotEmpty) {
      _validatePublicKeyHex(tokenAddress, 'token address');
    }

    if (ticketPrice.isEmpty &&
        maxEntries == 0 &&
        maxEntriesPerUser == 0 &&
        startAt == null &&
        expiredAt == null &&
        seedCommitHex.isEmpty &&
        metadata.isEmpty) {
      throw ArgumentError('no fields to update');
    }

    if (maxEntries < 0 || maxEntriesPerUser < 0) {
      throw ArgumentError('max entries must be >= 0');
    }

    if (maxEntries > 0 && maxEntriesPerUser > maxEntries) {
      throw ArgumentError('max_entries_per_user cannot exceed max_entries');
    }

    final JsonMessage data = {
      'address': address,
      'token_address': tokenAddress,
      'ticket_price': ticketPrice,
      'max_entries': maxEntries,
      'max_entries_per_user': maxEntriesPerUser,
      'seed_commit_hex': seedCommitHex,
      'metadata': metadata,
    };

    if (startAt != null) {
      data['start_at'] = startAt.toIso8601String();
    }

    if (expiredAt != null) {
      data['expired_at'] = expiredAt.toIso8601String();
    }

    return signAndSendTransaction(
      chainID: _chainID,
      from: from,
      to: address,
      method: METHOD_UPDATE_RAFFLE,
      data: data,
      version: _txVersion,
      uuid7: _newUuid7OrThrow(),
    );
  }

  Future<ContractOutput> pauseRaffle(String address, bool paused) async {
    if (address.isEmpty) {
      throw ArgumentError('address not set');
    }
    _validatePublicKeyHex(address, 'address');

    if (!paused) {
      throw ArgumentError('paused must be true: Pause: $paused');
    }

    final from = _requireFromAddress();

    final JsonMessage data = {'address': address, 'paused': paused};

    return signAndSendTransaction(
      chainID: _chainID,
      from: from,
      to: address,
      method: METHOD_PAUSE_RAFFLE,
      data: data,
      version: _txVersion,
      uuid7: _newUuid7OrThrow(),
    );
  }

  Future<ContractOutput> unpauseRaffle(String address, bool paused) async {
    if (address.isEmpty) {
      throw ArgumentError('address not set');
    }
    _validatePublicKeyHex(address, 'address');

    if (paused) {
      throw ArgumentError('paused must be false: Pause: $paused');
    }

    final from = _requireFromAddress();

    final JsonMessage data = {'address': address, 'paused': paused};

    return signAndSendTransaction(
      chainID: _chainID,
      from: from,
      to: address,
      method: METHOD_UNPAUSE_RAFFLE,
      data: data,
      version: _txVersion,
      uuid7: _newUuid7OrThrow(),
    );
  }

  Future<ContractOutput> enterRaffle({
    required String address,
    required int tickets,
    required String payTokenAddress,
    required String tokenType,
    required String uuid,
  }) async {
    final from = _requireFromAddress();

    if (address.isEmpty) {
      throw ArgumentError('address is required');
    }
    _validatePublicKeyHex(address, 'raffle address');

    if (tickets <= 0) {
      throw ArgumentError('tickets must be > 0');
    }

    if (payTokenAddress.isEmpty) {
      throw ArgumentError('pay_token_address is required');
    }
    _validatePublicKeyHex(payTokenAddress, 'pay_token_address');

    if (tokenType.isEmpty) {
      throw ArgumentError('tokenType not set');
    }

    if (tokenType == TOKEN_TYPE_NON_FUNGIBLE && uuid.isEmpty) {
      throw ArgumentError('uuid must be set for non-fungible tokens');
    }

    final JsonMessage data = {
      'address': address,
      'entrant': from,
      'tickets': tickets,
      'pay_token_address': payTokenAddress,
      'token_type': tokenType,
      'uuid': uuid,
    };

    return signAndSendTransaction(
      chainID: _chainID,
      from: from,
      to: address,
      method: METHOD_ENTER_RAFFLE,
      data: data,
      version: _txVersion,
      uuid7: _newUuid7OrThrow(),
    );
  }

  Future<ContractOutput> drawRaffle(String address, String revealSeed) async {
    if (address.isEmpty) {
      throw ArgumentError('address not set');
    }
    _validatePublicKeyHex(address, 'address');

    if (revealSeed.isEmpty) {
      throw ArgumentError('reveal_seed not set');
    }

    final from = _requireFromAddress();

    final JsonMessage data = {'address': address, 'reveal_seed': revealSeed};

    return signAndSendTransaction(
      chainID: _chainID,
      from: from,
      to: address,
      method: METHOD_DRAW_RAFFLE,
      data: data,
      version: _txVersion,
      uuid7: _newUuid7OrThrow(),
    );
  }

  Future<ContractOutput> claimRaffle({
    required String address,
    String prizeUUID = '',
    String winner = '',
    String tokenType = '',
    String uuid = '',
  }) async {
    if (address.isEmpty) {
      throw ArgumentError('address not set');
    }
    _validatePublicKeyHex(address, 'address');

    final effectivePrizeUUID = prizeUUID.isNotEmpty ? prizeUUID : uuid;

    if (effectivePrizeUUID.isEmpty) {
      throw ArgumentError('prizeUUID not set');
    }

    final from = _requireFromAddress();

    final JsonMessage data = {
      'raffle_address': address,
      'prize_uuid': effectivePrizeUUID,
    };

    return signAndSendTransaction(
      chainID: _chainID,
      from: from,
      to: address,
      method: METHOD_CLAIM_RAFFLE,
      data: data,
      version: _txVersion,
      uuid7: _newUuid7OrThrow(),
    );
  }

  Future<ContractOutput> withdrawRaffle({
    required String address,
    required String tokenAddress,
    required String amount,
    required String tokenType,
    required String uuid,
  }) async {
    if (address.isEmpty) {
      throw ArgumentError('address not set');
    }
    _validatePublicKeyHex(address, 'address');

    if (tokenAddress.isEmpty) {
      throw ArgumentError('token address not set');
    }
    _validatePublicKeyHex(tokenAddress, 'token address');

    if (amount.isEmpty) {
      throw ArgumentError('amount not set');
    }

    if (tokenType.isEmpty) {
      throw ArgumentError('tokenType not set');
    }

    if (tokenType == TOKEN_TYPE_NON_FUNGIBLE && uuid.isEmpty) {
      throw ArgumentError('uuid must be set for non-fungible tokens');
    }

    final from = _requireFromAddress();

    final JsonMessage data = {
      'address': address,
      'token_address': tokenAddress,
      'amount': amount,
      'token_type': tokenType,
      'uuid': uuid,
    };

    return signAndSendTransaction(
      chainID: _chainID,
      from: from,
      to: address,
      method: METHOD_WITHDRAW_RAFFLE,
      data: data,
      version: _txVersion,
      uuid7: _newUuid7OrThrow(),
    );
  }

  Future<ContractOutput> addRafflePrize({
    required String raffleAddress,
    required String tokenAddress,
    String amount = '',
    List<String> uuidNFTs = const <String>[],
    String tokenType = '',
    String uuid = '',
  }) async {
    if (raffleAddress.isEmpty) {
      throw ArgumentError('raffle address not set');
    }
    _validatePublicKeyHex(raffleAddress, 'raffle address');

    if (tokenAddress.isEmpty) {
      throw ArgumentError('token address not set');
    }
    _validatePublicKeyHex(tokenAddress, 'token address');

    final effectiveUUIDs = uuidNFTs.isNotEmpty
        ? uuidNFTs
        : (tokenType == TOKEN_TYPE_NON_FUNGIBLE && uuid.isNotEmpty
              ? <String>[uuid]
              : const <String>[]);

    if (amount.isEmpty && effectiveUUIDs.isEmpty) {
      throw ArgumentError('amount not set or uuidNFTs not set');
    }

    final from = _requireFromAddress();

    final JsonMessage data = {
      'amount': amount,
      'raffle_address': raffleAddress,
      'token_address': tokenAddress,
      'uuid_nfts': effectiveUUIDs,
    };

    return signAndSendTransaction(
      chainID: _chainID,
      from: from,
      to: raffleAddress,
      method: METHOD_ADD_RAFFLE_PRIZE,
      data: data,
      version: _txVersion,
      uuid7: _newUuid7OrThrow(),
    );
  }

  Future<ContractOutput> removeRafflePrize({
    required String raffleAddress,
    required String uuid,
    String tokenType = '',
  }) async {
    if (raffleAddress.isEmpty) {
      throw ArgumentError('raffle address not set');
    }
    _validatePublicKeyHex(raffleAddress, 'raffle address');

    if (uuid.isEmpty) {
      throw ArgumentError('uuid not set');
    }

    final from = _requireFromAddress();

    final JsonMessage data = {'raffle_address': raffleAddress, 'uuid': uuid};

    return signAndSendTransaction(
      chainID: _chainID,
      from: from,
      to: raffleAddress,
      method: METHOD_REMOVE_RAFFLE_PRIZE,
      data: data,
      version: _txVersion,
      uuid7: _newUuid7OrThrow(),
    );
  }

  Future<ContractOutput> getRaffle(String address) async {
    _requireFromAddress();

    if (address.isEmpty) {
      throw ArgumentError('raffle address must be set');
    }
    _validatePublicKeyHex(address, 'raffle address');

    return getState(
      to: address,
      method: METHOD_GET_RAFFLE,
      data: const <String, dynamic>{},
    );
  }

  Future<ContractOutput> listRaffles({
    String owner = '',
    String tokenAddress = '',
    bool? paused,
    bool? activeOnly,
    int page = 1,
    int limit = 10,
    bool ascending = false,
  }) async {
    _requireFromAddress();

    if (owner.isNotEmpty) {
      _validatePublicKeyHex(owner, 'owner address');
    }

    if (tokenAddress.isNotEmpty) {
      _validatePublicKeyHex(tokenAddress, 'token address');
    }

    if (page < 1) {
      throw ArgumentError('page must be greater than 0');
    }

    if (limit < 1) {
      throw ArgumentError('limit must be greater than 0');
    }

    final JsonMessage data = {
      'owner': owner,
      'page': page,
      'limit': limit,
      'ascending': ascending,
      'token_address': tokenAddress,
    };

    if (paused != null) {
      data['paused'] = paused;
    }

    if (activeOnly != null) {
      data['active_only'] = activeOnly;
    }

    return getState(to: '', method: METHOD_LIST_RAFFLES, data: data);
  }

  Future<ContractOutput> listPrizes({
    required String raffleAddress,
    int page = 1,
    int limit = 10,
    bool ascending = false,
  }) async {
    _requireFromAddress();

    if (raffleAddress.isEmpty) {
      throw ArgumentError('raffle address must be set');
    }
    _validatePublicKeyHex(raffleAddress, 'raffle address');

    if (page < 1) {
      throw ArgumentError('page must be greater than 0');
    }

    if (limit < 1) {
      throw ArgumentError('limit must be greater than 0');
    }

    final JsonMessage data = {
      'raffle_address': raffleAddress,
      'page': page,
      'limit': limit,
      'ascending': ascending,
      'contract_version': RAFFLE_CONTRACT_V2,
    };

    return getState(to: '', method: METHOD_LIST_PRIZES, data: data);
  }

  Future<ContractOutput> getPrize({
    required String address,
    required String prizeUUID,
  }) async {
    _requireFromAddress();

    if (prizeUUID.isEmpty) {
      throw ArgumentError('prize UUID must be set');
    }

    final JsonMessage data = {
      'raffle_address': address,
      'prize_uuid': prizeUUID,
      'contract_version': RAFFLE_CONTRACT_V2,
    };

    return getState(to: address, method: METHOD_GET_PRIZE, data: data);
  }
}
