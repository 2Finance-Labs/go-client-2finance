part of 'two_finance_blockchain.dart';

extension DropClient on TwoFinanceBlockchain {
  static const int _txVersion = 1;

  String _requireDropFromAddress() {
    final from = publicKeyHex ?? '';
    if (from.isEmpty) {
      throw ArgumentError('from address not set');
    }

    KeyManager.validateEDDSAPublicKeyHex(from);
    return from;
  }

  String _newDropUuid7() => newUUID7();

  DateTime _goZeroDropDateTime() => DateTime.utc(1);

  void _validateDropAddress(String value, String label) {
    try {
      KeyManager.validateEDDSAPublicKeyHex(value);
    } catch (e) {
      throw ArgumentError('invalid $label: $e');
    }
  }

  Future<ContractOutput> newDrop({
    required String address,
    required String programAddress,
    required String tokenAddress,
    required String owner,
    required String title,
    required String description,
    required String shortDescription,
    String imageUrl = '',
    String bannerUrl = '',
    required Map<String, bool> categories,
    Map<String, bool> socialRequirements = const <String, bool>{},
    Map<String, bool> postLinks = const <String, bool>{},
    required String verificationType,
    required DateTime startAt,
    required DateTime expireAt,
    required int requestLimit,
    required String claimAmount,
    required int claimIntervalSeconds,
  }) async {
    final from = _requireDropFromAddress();

    if (address.isEmpty) {
      throw ArgumentError('drop address not set');
    }
    if (owner.isEmpty) {
      throw ArgumentError('owner not set');
    }
    if (title.isEmpty) {
      throw ArgumentError('title not set');
    }
    if (verificationType.isEmpty) {
      throw ArgumentError('verification type not set');
    }

    _validateDropAddress(owner, 'owner address');

    return signAndSendTransaction(
      chainID: _chainID,
      from: from,
      to: address,
      method: METHOD_NEW_DROP,
      data: {
        'address': address,
        'program_address': programAddress,
        'token_address': tokenAddress,
        'owner': owner,
        'title': title,
        'description': description,
        'short_description': shortDescription,
        'image_url': imageUrl,
        'banner_url': bannerUrl,
        'categories': categories,
        'social_requirements': socialRequirements,
        'post_links': postLinks,
        'verification_type': verificationType,
        'start_at': startAt.toUtc().toIso8601String(),
        'expire_at': expireAt.toUtc().toIso8601String(),
        'request_limit': requestLimit,
        'claim_amount': claimAmount,
        'claim_interval_seconds': claimIntervalSeconds,
      },
      version: _txVersion,
      uuid7: _newDropUuid7(),
    );
  }

  Future<ContractOutput> updateDropMetadata({
    required String address,
    String programAddress = '',
    String tokenAddress = '',
    String title = '',
    String description = '',
    String shortDescription = '',
    String imageUrl = '',
    String bannerUrl = '',
    Map<String, bool> categories = const <String, bool>{},
    Map<String, bool> socialRequirements = const <String, bool>{},
    Map<String, bool> postLinks = const <String, bool>{},
    String verificationType = '',
    DateTime? startAt,
    DateTime? expireAt,
    int? requestLimit,
    String claimAmount = '',
    int? claimIntervalSeconds,
  }) async {
    if (address.isEmpty) {
      throw ArgumentError('drop address not set');
    }

    final from = _requireDropFromAddress();
    final effectiveStartAt = startAt ?? _goZeroDropDateTime();
    final effectiveExpireAt = expireAt ?? _goZeroDropDateTime();

    final data = <String, dynamic>{
      'address': address,
      'program_address': programAddress,
      'token_address': tokenAddress,
      'title': title,
      'description': description,
      'short_description': shortDescription,
      'image_url': imageUrl,
      'banner_url': bannerUrl,
      'categories': categories,
      'social_requirements': socialRequirements,
      'post_links': postLinks,
      'verification_type': verificationType,
      'start_at': effectiveStartAt.toUtc().toIso8601String(),
      'expire_at': effectiveExpireAt.toUtc().toIso8601String(),
      'request_limit': requestLimit ?? 0,
      'claim_amount': claimAmount,
      'claim_interval_seconds': claimIntervalSeconds ?? 0,
    };

    return signAndSendTransaction(
      chainID: _chainID,
      from: from,
      to: address,
      method: METHOD_UPDATE_DROP_METADATA,
      data: data,
      version: _txVersion,
      uuid7: _newDropUuid7(),
    );
  }

  Future<ContractOutput> allowOracles({
    required String address,
    required Map<String, bool> oracles,
  }) async {
    final from = _requireDropFromAddress();

    if (address.isEmpty) {
      throw ArgumentError('drop address not set');
    }
    if (oracles.isEmpty) {
      throw ArgumentError('oracles map is empty');
    }

    _validateDropAddress(address, 'drop address');
    validateUserMap(oracles, 'oracles');

    return signAndSendTransaction(
      chainID: _chainID,
      from: from,
      to: address,
      method: METHOD_ALLOW_ORACLES,
      data: {'oracles': oracles},
      version: _txVersion,
      uuid7: _newDropUuid7(),
    );
  }

  Future<ContractOutput> disallowOracles({
    required String address,
    required Map<String, bool> oracles,
  }) async {
    final from = _requireDropFromAddress();

    if (address.isEmpty) {
      throw ArgumentError('drop address not set');
    }
    if (oracles.isEmpty) {
      throw ArgumentError('oracles map is empty');
    }

    _validateDropAddress(address, 'drop address');
    validateUserMap(oracles, 'oracles');

    return signAndSendTransaction(
      chainID: _chainID,
      from: from,
      to: address,
      method: METHOD_DISALLOW_ORACLES,
      data: {'oracles': oracles},
      version: _txVersion,
      uuid7: _newDropUuid7(),
    );
  }

  Future<ContractOutput> depositDrop({
    required String address,
    required String programAddress,
    required String tokenAddress,
    required String amount,
    List<String> uuids = const <String>[],
  }) async {
    final from = _requireDropFromAddress();

    if (address.isEmpty) {
      throw ArgumentError('drop address not set');
    }
    if (amount.isEmpty) {
      throw ArgumentError('amount not set');
    }

    return signAndSendTransaction(
      chainID: _chainID,
      from: from,
      to: address,
      method: METHOD_DEPOSIT_DROP,
      data: {
        'program_address': programAddress,
        'token_address': tokenAddress,
        'amount': amount,
        'uuids': uuids,
      },
      version: _txVersion,
      uuid7: _newDropUuid7(),
    );
  }

  Future<ContractOutput> claimDrop({required String address}) async {
    final from = _requireDropFromAddress();

    if (address.isEmpty) {
      throw ArgumentError('drop address not set');
    }

    _validateDropAddress(address, 'drop address');

    return signAndSendTransaction(
      chainID: _chainID,
      from: from,
      to: address,
      method: METHOD_CLAIM_DROP,
      data: {'address': address, 'wallet': from},
      version: _txVersion,
      uuid7: _newDropUuid7(),
    );
  }

  Future<ContractOutput> withdrawDrop({
    required String address,
    required String programAddress,
    required String tokenAddress,
    required String amount,
    List<String> uuids = const <String>[],
  }) async {
    final from = _requireDropFromAddress();

    if (address.isEmpty) {
      throw ArgumentError('drop address not set');
    }
    if (amount.isEmpty) {
      throw ArgumentError('amount not set');
    }

    return signAndSendTransaction(
      chainID: _chainID,
      from: from,
      to: address,
      method: METHOD_WITHDRAW_DROP,
      data: {
        'program_address': programAddress,
        'token_address': tokenAddress,
        'amount': amount,
        'uuids': uuids,
      },
      version: _txVersion,
      uuid7: _newDropUuid7(),
    );
  }

  Future<ContractOutput> pauseDrop(String address) async {
    if (address.isEmpty) {
      throw ArgumentError('drop address not set');
    }
    _validateDropAddress(address, 'drop address');

    final from = _requireDropFromAddress();

    return signAndSendTransaction(
      chainID: _chainID,
      from: from,
      to: address,
      method: METHOD_PAUSE_DROP,
      data: {'address': address, 'paused': true},
      version: _txVersion,
      uuid7: _newDropUuid7(),
    );
  }

  Future<ContractOutput> unpauseDrop(String address) async {
    if (address.isEmpty) {
      throw ArgumentError('drop address not set');
    }
    _validateDropAddress(address, 'drop address');

    final from = _requireDropFromAddress();

    return signAndSendTransaction(
      chainID: _chainID,
      from: from,
      to: address,
      method: METHOD_UNPAUSE_DROP,
      data: {'address': address, 'paused': false},
      version: _txVersion,
      uuid7: _newDropUuid7(),
    );
  }

  Future<ContractOutput> attestParticipantEligibility({
    required String address,
    required String wallet,
    required bool approved,
  }) async {
    final from = _requireDropFromAddress();

    if (address.isEmpty) {
      throw ArgumentError('drop address not set');
    }
    if (wallet.isEmpty) {
      throw ArgumentError('wallet not set');
    }

    _validateDropAddress(address, 'drop address');
    _validateDropAddress(wallet, 'wallet address');

    return signAndSendTransaction(
      chainID: _chainID,
      from: from,
      to: address,
      method: METHOD_ATTEST_ELIGIBILITY,
      data: {'wallet': wallet, 'approved': approved},
      version: _txVersion,
      uuid7: _newDropUuid7(),
    );
  }

  Future<ContractOutput> manuallyAttestParticipantEligibility({
    required String address,
    required String wallet,
    required bool approved,
  }) async {
    final from = _requireDropFromAddress();

    if (address.isEmpty) {
      throw ArgumentError('drop address not set');
    }
    if (wallet.isEmpty) {
      throw ArgumentError('wallet not set');
    }

    _validateDropAddress(address, 'drop address');
    _validateDropAddress(wallet, 'wallet address');

    return signAndSendTransaction(
      chainID: _chainID,
      from: from,
      to: address,
      method: METHOD_MANUAL_ATTEST_ELIGIBILITY,
      data: {'wallet': wallet, 'approved': approved},
      version: _txVersion,
      uuid7: _newDropUuid7(),
    );
  }

  Future<ContractOutput> getDrop({required String address}) async {
    final from = _requireDropFromAddress();

    if (address.isEmpty) {
      throw ArgumentError('drop address must be set');
    }

    _validateDropAddress(from, 'from address');
    _validateDropAddress(address, 'drop address');

    return getState(
      to: address,
      method: METHOD_GET_DROP,
      data: const <String, dynamic>{},
    );
  }

  Future<ContractOutput> lastClaimed({
    required String address,
    required String wallet,
  }) async {
    final from = _requireDropFromAddress();

    if (address.isEmpty) {
      throw ArgumentError('drop address must be set');
    }
    if (wallet.isEmpty) {
      throw ArgumentError('wallet must be set');
    }

    _validateDropAddress(from, 'from address');
    _validateDropAddress(address, 'drop address');
    _validateDropAddress(wallet, 'wallet address');

    return getState(
      to: address,
      method: METHOD_LAST_CLAIMED_DROP,
      data: {'wallet': wallet},
    );
  }

  Future<ContractOutput> listDrops({
    String owner = '',
    int page = 1,
    int limit = 10,
    bool ascending = false,
  }) async {
    final from = _requireDropFromAddress();
    _validateDropAddress(from, 'from address');

    if (owner.isNotEmpty) {
      _validateDropAddress(owner, 'owner address');
    }
    if (page < 1) {
      throw ArgumentError('page must be greater than 0');
    }
    if (limit < 1) {
      throw ArgumentError('limit must be greater than 0');
    }

    return getState(
      to: '',
      method: METHOD_LIST_DROPS,
      data: {
        'owner': owner,
        'page': page,
        'limit': limit,
        'ascending': ascending,
        'contract_version': DROP_CONTRACT_V2,
      },
    );
  }
}
