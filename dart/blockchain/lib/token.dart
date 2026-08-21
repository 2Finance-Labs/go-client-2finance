part of 'two_finance_blockchain.dart';

extension Token on TwoFinanceBlockchain {
  Future<ContractOutput> addToken({
    required String address, // token contract address (deployed)
    required String symbol,
    required String name,
    required int decimals,
    required String totalSupply,
    required String description,
    required String owner,
    required String image,
    required String website,
    required Map<String, String> tagsSocialMedia,
    required Map<String, String> tagsCategory,
    required Map<String, String> tags,
    required String creator,
    required String creatorWebsite,
    required Map<String, bool> allowedUsers,
    required Map<String, bool> blockedUsers,
    required Map<String, dynamic> frozenAccounts,
    required List<Map<String, dynamic>> feeTiersList,
    required String feeAddress,
    required bool freezeAuthorityRevoked,
    required bool mintAuthorityRevoked,
    required bool updateAuthorityRevoked,
    required bool paused,
    required DateTime expiredAt,
    required String assetGlbUri,
    required String tokenType,
    required bool transferable,
    String assetType = '',
    bool stablecoin = false,
  }) async {
    if (symbol.isEmpty) throw ArgumentError('symbol not set');
    if (name.isEmpty) throw ArgumentError('name not set');
    if (totalSupply.isEmpty) throw ArgumentError('total supply not set');
    if (owner.isEmpty) throw ArgumentError('owner not set');
    if (creator.isEmpty) throw ArgumentError('creator not set');
    if (creatorWebsite.isEmpty) throw ArgumentError('creator website not set');
    if (image.isEmpty) throw ArgumentError('image not set');
    if (website.isEmpty) throw ArgumentError('website not set');
    if (feeAddress.isEmpty) throw ArgumentError('fee address not set');
    if (assetGlbUri.isEmpty) throw ArgumentError('asset GLB URI not set');
    if (tokenType.isEmpty) throw ArgumentError('token type not set');

    KeyManager.validateEDDSAPublicKeyHex(feeAddress);

    validateUserMap(allowedUsers, 'allowed users');
    validateUserMap(blockedUsers, 'blocked users');

    final from = publicKeyHex ?? '';
    if (from.isEmpty) throw ArgumentError('from address not set');
    KeyManager.validateEDDSAPublicKeyHex(from);

    final uuid7 = newUUID7();
    const int version = 1;

    final JsonMessage data = {
      "symbol": symbol,
      "name": name,
      "decimals": decimals,
      "total_supply": totalSupply,
      "description": description,
      "owner": owner,
      "fee_tiers_list": feeTiersList,
      "fee_address": feeAddress,
      "image": image,
      "website": website,
      "tags_social_media": tagsSocialMedia,
      "tags_category": tagsCategory,
      "tags": tags,
      "creator": creator,
      "creator_website": creatorWebsite,
      "allowed_users": allowedUsers,
      "blocked_users": blockedUsers,
      "frozen_accounts": frozenAccounts,
      "freeze_authority_revoked": freezeAuthorityRevoked,
      "mint_authority_revoked": mintAuthorityRevoked,
      "update_authority_revoked": updateAuthorityRevoked,
      "paused": paused,
      "expired_at": expiredAt.toIso8601String(),
      "asset_glb_uri": assetGlbUri,
      "token_type": tokenType,
      "transferable": transferable,
      "asset_type": assetType,
    };

    return signAndSendTransaction(
      chainID: _chainID,
      from: from,
      to: address,
      method: METHOD_ADD_TOKEN,
      data: data,
      version: version,
      uuid7: uuid7,
    );
  }

  Future<ContractOutput> mintToken({
    required String tokenAddress,
    required String mintTo,
    required String amount,
    int decimals = 0,
    String tokenType = '',
  }) async {
    final from = publicKeyHex ?? '';
    if (tokenAddress.isEmpty) throw ArgumentError('token address not set');
    if (mintTo.isEmpty) throw ArgumentError('mint to address not set');
    if (amount.isEmpty) throw ArgumentError('amount not set');

    KeyManager.validateEDDSAPublicKeyHex(from);
    KeyManager.validateEDDSAPublicKeyHex(tokenAddress);
    KeyManager.validateEDDSAPublicKeyHex(mintTo);

    var amountScaled = amount;
    if (decimals != 0) {
      amountScaled = rescaleDecimalString(amount, 0, decimals);
    }

    final uuid7 = newUUID7();
    const int version = 1;

    final JsonMessage data = {"mint_to": mintTo, "amount": amountScaled};

    return signAndSendTransaction(
      chainID: _chainID,
      from: from,
      to: tokenAddress,
      method: METHOD_MINT_TOKEN,
      data: data,
      version: version,
      uuid7: uuid7,
    );
  }

  Future<ContractOutput> burnToken({
    required String tokenAddress,
    String amount = '',
    List<String> tokenUUIDList = const <String>[],
    int decimals = 0,
    String tokenType = '',
    String uuid = '',
  }) async {
    final from = publicKeyHex ?? '';
    if (tokenAddress.isEmpty) throw ArgumentError('token address not set');

    KeyManager.validateEDDSAPublicKeyHex(from);
    KeyManager.validateEDDSAPublicKeyHex(tokenAddress);

    final uuids = tokenUUIDList.isNotEmpty
        ? tokenUUIDList
        : (uuid.isNotEmpty ? <String>[uuid] : const <String>[]);
    var amountScaled = amount;
    if (decimals != 0) {
      amountScaled = rescaleDecimalString(amount, 0, decimals);
    }
    if (amountScaled.isEmpty && uuids.isEmpty) {
      throw ArgumentError('amount not set or token UUID not set');
    }

    final uuid7 = newUUID7();
    const int version = 1;

    final JsonMessage data = {};

    if (uuids.isNotEmpty) {
      data["uuids"] = uuids;
    }
    if (amountScaled.isNotEmpty) {
      data["amount"] = amountScaled;
    }

    return signAndSendTransaction(
      chainID: _chainID,
      from: from,
      to: tokenAddress,
      method: METHOD_BURN_TOKEN,
      data: data,
      version: version,
      uuid7: uuid7,
    );
  }

  Future<ContractOutput> transferToken({
    required String tokenAddress,
    required String transferTo,
    String amount = '',
    List<String> tokenUUIDList = const <String>[],
    int decimals = 0,
    String tokenType = '',
    String uuid = '',
  }) async {
    final from = publicKeyHex ?? '';
    if (transferTo.isEmpty) throw ArgumentError('to address not set');
    if (tokenAddress.isEmpty) throw ArgumentError('token address not set');
    if (from == transferTo) {
      throw ArgumentError('from and to addresses are the same');
    }

    KeyManager.validateEDDSAPublicKeyHex(from);
    KeyManager.validateEDDSAPublicKeyHex(transferTo);
    KeyManager.validateEDDSAPublicKeyHex(tokenAddress);

    var amountScaled = amount;
    if (decimals != 0) {
      amountScaled = rescaleDecimalString(amount, 0, decimals);
    }

    final uuid7 = newUUID7();
    const int version = 1;

    final uuids = tokenUUIDList.isNotEmpty
        ? tokenUUIDList
        : (uuid.isNotEmpty ? <String>[uuid] : const <String>[]);
    final JsonMessage data = {"transfer_to": transferTo};

    if (amountScaled.isNotEmpty) {
      data["amount"] = amountScaled;
    }
    if (uuids.isNotEmpty) {
      data["token_uuid_list"] = uuids;
    }

    return signAndSendTransaction(
      chainID: _chainID,
      from: from,
      to: tokenAddress,
      method: METHOD_TRANSFER_TOKEN,
      data: data,
      version: version,
      uuid7: uuid7,
    );
  }

  Future<ContractOutput> allowUsers(
    String tokenAddress,
    Map<String, bool> users,
  ) async {
    final from = publicKeyHex ?? '';
    if (tokenAddress.isEmpty) throw ArgumentError('token address not set');
    if (users.isEmpty) throw ArgumentError('users map is empty');

    KeyManager.validateEDDSAPublicKeyHex(from);
    KeyManager.validateEDDSAPublicKeyHex(tokenAddress);
    validateUserMap(users, 'allow users');

    final uuid7 = newUUID7();
    const int version = 1;

    return signAndSendTransaction(
      chainID: _chainID,
      from: from,
      to: tokenAddress,
      method: METHOD_ADD_ALLOWED_USERS,
      data: {"allowed_users": users},
      version: version,
      uuid7: uuid7,
    );
  }

  Future<ContractOutput> disallowUsers(
    String tokenAddress,
    Map<String, bool> users,
  ) async {
    final from = publicKeyHex ?? '';
    if (tokenAddress.isEmpty) throw ArgumentError('token address not set');
    if (users.isEmpty) throw ArgumentError('users map is empty');

    KeyManager.validateEDDSAPublicKeyHex(from);
    KeyManager.validateEDDSAPublicKeyHex(tokenAddress);
    validateUserMap(users, 'disallow users');

    final uuid7 = newUUID7();
    const int version = 1;

    return signAndSendTransaction(
      chainID: _chainID,
      from: from,
      to: tokenAddress,
      method: METHOD_REMOVE_ALLOWED_USERS,
      data: {"allowed_users": users},
      version: version,
      uuid7: uuid7,
    );
  }

  Future<ContractOutput> blockUsers(
    String tokenAddress,
    Map<String, bool> users,
  ) async {
    final from = publicKeyHex ?? '';
    if (tokenAddress.isEmpty) throw ArgumentError('token address not set');
    if (users.isEmpty) throw ArgumentError('users map is empty');

    KeyManager.validateEDDSAPublicKeyHex(from);
    KeyManager.validateEDDSAPublicKeyHex(tokenAddress);
    validateUserMap(users, 'block users');

    final uuid7 = newUUID7();
    const int version = 1;

    return signAndSendTransaction(
      chainID: _chainID,
      from: from,
      to: tokenAddress,
      method: METHOD_ADD_BLOCKED_USERS,
      data: {"blocked_users": users},
      version: version,
      uuid7: uuid7,
    );
  }

  Future<ContractOutput> unblockUsers(
    String tokenAddress,
    Map<String, bool> users,
  ) async {
    final from = publicKeyHex ?? '';
    if (tokenAddress.isEmpty) throw ArgumentError('token address not set');
    if (users.isEmpty) throw ArgumentError('users map is empty');

    KeyManager.validateEDDSAPublicKeyHex(from);
    KeyManager.validateEDDSAPublicKeyHex(tokenAddress);
    validateUserMap(users, 'unblock users');

    final uuid7 = newUUID7();
    const int version = 1;

    return signAndSendTransaction(
      chainID: _chainID,
      from: from,
      to: tokenAddress,
      method: METHOD_REMOVE_BLOCKED_USERS,
      data: {"blocked_users": users},
      version: version,
      uuid7: uuid7,
    );
  }

  Future<ContractOutput> revokeFreezeAuthority(
    String tokenAddress,
    bool revoke,
  ) async {
    final from = publicKeyHex ?? '';
    if (tokenAddress.isEmpty) throw ArgumentError('token address not set');
    KeyManager.validateEDDSAPublicKeyHex(from);
    KeyManager.validateEDDSAPublicKeyHex(tokenAddress);

    final uuid7 = newUUID7();
    const int version = 1;

    return signAndSendTransaction(
      chainID: _chainID,
      from: from,
      to: tokenAddress,
      method: METHOD_REVOKE_FREEZE_AUTHORITY,
      data: {"revoked": revoke},
      version: version,
      uuid7: uuid7,
    );
  }

  Future<ContractOutput> revokeMintAuthority(
    String tokenAddress,
    bool revoke,
  ) async {
    final from = publicKeyHex ?? '';
    if (tokenAddress.isEmpty) throw ArgumentError('token address not set');
    KeyManager.validateEDDSAPublicKeyHex(from);
    KeyManager.validateEDDSAPublicKeyHex(tokenAddress);

    final uuid7 = newUUID7();
    const int version = 1;

    return signAndSendTransaction(
      chainID: _chainID,
      from: from,
      to: tokenAddress,
      method: METHOD_REVOKE_MINT_AUTHORITY,
      data: {"revoked": revoke},
      version: version,
      uuid7: uuid7,
    );
  }

  Future<ContractOutput> revokeUpdateAuthority(
    String tokenAddress,
    bool revoke,
  ) async {
    final from = publicKeyHex ?? '';
    if (tokenAddress.isEmpty) throw ArgumentError('token address not set');
    KeyManager.validateEDDSAPublicKeyHex(from);
    KeyManager.validateEDDSAPublicKeyHex(tokenAddress);

    final uuid7 = newUUID7();
    const int version = 1;

    return signAndSendTransaction(
      chainID: _chainID,
      from: from,
      to: tokenAddress,
      method: METHOD_REVOKE_UPDATE_AUTHORITY,
      data: {"revoked": revoke},
      version: version,
      uuid7: uuid7,
    );
  }

  Future<ContractOutput> updateMetadata({
    required String tokenAddress,
    required String symbol,
    required String name,
    required int decimals,
    required String description,
    required String image,
    required String website,
    required Map<String, String> tagsSocialMedia,
    required Map<String, String> tagsCategory,
    required Map<String, String> tags,
    required String creator,
    required String creatorWebsite,
    required DateTime expiredAt,
  }) async {
    final from = publicKeyHex ?? '';
    if (tokenAddress.isEmpty) throw ArgumentError('token address not set');
    if (symbol.isEmpty) throw ArgumentError('symbol not set');
    if (name.isEmpty) throw ArgumentError('name not set');
    if (description.isEmpty) throw ArgumentError('description not set');
    if (image.isEmpty) throw ArgumentError('image not set');
    if (website.isEmpty) throw ArgumentError('website not set');
    if (creator.isEmpty) throw ArgumentError('creator not set');
    if (creatorWebsite.isEmpty) throw ArgumentError('creator website not set');

    KeyManager.validateEDDSAPublicKeyHex(from);
    KeyManager.validateEDDSAPublicKeyHex(tokenAddress);

    final uuid7 = newUUID7();
    const int version = 1;

    final data = <String, dynamic>{
      "symbol": symbol,
      "name": name,
      "decimals": decimals,
      "description": description,
      "image": image,
      "website": website,
      "tags_social_media": tagsSocialMedia,
      "tags_category": tagsCategory,
      "tags": tags,
      "creator": creator,
      "creator_website": creatorWebsite,
      "expired_at": expiredAt.toIso8601String(),
    };

    return signAndSendTransaction(
      chainID: _chainID,
      from: from,
      to: tokenAddress,
      method: METHOD_UPDATE_METADATA,
      data: data,
      version: version,
      uuid7: uuid7,
    );
  }

  Future<ContractOutput> freezeWallet(
    String tokenAddress,
    String wallet,
  ) async {
    final from = publicKeyHex ?? '';
    if (tokenAddress.isEmpty) throw ArgumentError('token address not set');
    if (wallet.isEmpty) throw ArgumentError('wallet not set');

    KeyManager.validateEDDSAPublicKeyHex(from);
    KeyManager.validateEDDSAPublicKeyHex(tokenAddress);
    KeyManager.validateEDDSAPublicKeyHex(wallet);

    final uuid7 = newUUID7();
    const int version = 1;

    return signAndSendTransaction(
      chainID: _chainID,
      from: from,
      to: tokenAddress,
      method: METHOD_FREEZE_WALLET,
      data: {"wallet": wallet},
      version: version,
      uuid7: uuid7,
    );
  }

  Future<ContractOutput> unfreezeWallet(
    String tokenAddress,
    String wallet,
  ) async {
    final from = publicKeyHex ?? '';
    if (tokenAddress.isEmpty) throw ArgumentError('token address not set');
    if (wallet.isEmpty) throw ArgumentError('wallet not set');

    KeyManager.validateEDDSAPublicKeyHex(from);
    KeyManager.validateEDDSAPublicKeyHex(tokenAddress);
    KeyManager.validateEDDSAPublicKeyHex(wallet);

    final uuid7 = newUUID7();
    const int version = 1;

    return signAndSendTransaction(
      chainID: _chainID,
      from: from,
      to: tokenAddress,
      method: METHOD_UNFREEZE_WALLET,
      data: {"wallet": wallet},
      version: version,
      uuid7: uuid7,
    );
  }

  Future<ContractOutput> pauseToken(
    String tokenAddress, [
    bool paused = true,
  ]) async {
    final from = publicKeyHex ?? '';
    if (tokenAddress.isEmpty) throw ArgumentError('token address not set');
    if (!paused) throw ArgumentError('paused must be true to pause token');
    KeyManager.validateEDDSAPublicKeyHex(from);
    KeyManager.validateEDDSAPublicKeyHex(tokenAddress);

    final uuid7 = newUUID7();
    const int version = 1;

    return signAndSendTransaction(
      chainID: _chainID,
      from: from,
      to: tokenAddress,
      method: METHOD_PAUSE_TOKEN,
      data: {"paused": paused},
      version: version,
      uuid7: uuid7,
    );
  }

  Future<ContractOutput> unpauseToken(
    String tokenAddress, [
    bool paused = false,
  ]) async {
    final from = publicKeyHex ?? '';
    if (tokenAddress.isEmpty) throw ArgumentError('token address not set');
    if (paused) throw ArgumentError('paused must be false to unpause token');
    KeyManager.validateEDDSAPublicKeyHex(from);
    KeyManager.validateEDDSAPublicKeyHex(tokenAddress);

    final uuid7 = newUUID7();
    const int version = 1;

    return signAndSendTransaction(
      chainID: _chainID,
      from: from,
      to: tokenAddress,
      method: METHOD_UNPAUSE_TOKEN,
      data: {"paused": paused},
      version: version,
      uuid7: uuid7,
    );
  }

  Future<ContractOutput> updateFeeTiers(
    String tokenAddress,
    List<Map<String, dynamic>> feeTiersList,
  ) async {
    final from = publicKeyHex ?? '';
    if (tokenAddress.isEmpty) throw ArgumentError('token address not set');
    if (feeTiersList.isEmpty) throw ArgumentError('fee tiers list is empty');

    KeyManager.validateEDDSAPublicKeyHex(from);
    KeyManager.validateEDDSAPublicKeyHex(tokenAddress);

    final uuid7 = newUUID7();
    const int version = 1;

    return signAndSendTransaction(
      chainID: _chainID,
      from: from,
      to: tokenAddress,
      method: METHOD_UPDATE_FEE_TIERS,
      data: {"fee_tiers_list": feeTiersList},
      version: version,
      uuid7: uuid7,
    );
  }

  Future<ContractOutput> updateFeeAddress(
    String tokenAddress,
    String feeAddress,
  ) async {
    final from = publicKeyHex ?? '';
    if (tokenAddress.isEmpty) throw ArgumentError('token address not set');
    if (feeAddress.isEmpty) throw ArgumentError('fee address not set');

    KeyManager.validateEDDSAPublicKeyHex(from);
    KeyManager.validateEDDSAPublicKeyHex(tokenAddress);
    KeyManager.validateEDDSAPublicKeyHex(feeAddress);

    final uuid7 = newUUID7();
    const int version = 1;

    return signAndSendTransaction(
      chainID: _chainID,
      from: from,
      to: tokenAddress,
      method: METHOD_UPDATE_FEE_ADDRESS,
      data: {"fee_address": feeAddress},
      version: version,
      uuid7: uuid7,
    );
  }

  Future<ContractOutput> updateGlbFile(
    String tokenAddress,
    String newAssetGlbUri,
  ) async {
    final from = publicKeyHex ?? '';
    if (tokenAddress.isEmpty) throw ArgumentError('token address not set');
    if (newAssetGlbUri.isEmpty)
      throw ArgumentError('new asset GLB URI not set');

    KeyManager.validateEDDSAPublicKeyHex(from);
    KeyManager.validateEDDSAPublicKeyHex(tokenAddress);

    final uuid7 = newUUID7();
    const int version = 1;

    return signAndSendTransaction(
      chainID: _chainID,
      from: from,
      to: tokenAddress,
      method: METHOD_UPDATE_GLB_FILE,
      data: {"new_asset_glb_uri": newAssetGlbUri},
      version: version,
      uuid7: uuid7,
    );
  }

  Future<ContractOutput> transferableToken(
    String tokenAddress,
    bool transferable,
  ) async {
    final from = publicKeyHex ?? '';
    if (tokenAddress.isEmpty) throw ArgumentError('token address not set');
    KeyManager.validateEDDSAPublicKeyHex(from);
    KeyManager.validateEDDSAPublicKeyHex(tokenAddress);

    final uuid7 = newUUID7();
    const int version = 1;

    return signAndSendTransaction(
      chainID: _chainID,
      from: from,
      to: tokenAddress,
      method: METHOD_TRANSFERABLE_TOKEN,
      data: {"transferable": transferable},
      version: version,
      uuid7: uuid7,
    );
  }

  Future<ContractOutput> untransferableToken(
    String tokenAddress,
    bool transferable,
  ) async {
    final from = publicKeyHex ?? '';
    if (tokenAddress.isEmpty) throw ArgumentError('token address not set');

    KeyManager.validateEDDSAPublicKeyHex(from);
    KeyManager.validateEDDSAPublicKeyHex(tokenAddress);

    final uuid7 = newUUID7();
    const int version = 1;

    return signAndSendTransaction(
      chainID: _chainID,
      from: from,
      to: tokenAddress,
      method: METHOD_UNTRANSFERABLE_TOKEN,
      data: {"transferable": transferable},
      version: version,
      uuid7: uuid7,
    );
  }

  // -------------------- READ CALLS (getState) --------------------

  Future<ContractOutput> getToken({
    String tokenAddress = '',
    String symbol = '',
    String name = '',
  }) async {
    final from = publicKeyHex ?? '';
    KeyManager.validateEDDSAPublicKeyHex(from);

    if (tokenAddress.isEmpty && symbol.isEmpty && name.isEmpty) {
      throw ArgumentError('token address, symbol or name must be set');
    }

    final data = <String, dynamic>{"symbol": symbol, "name": name};

    // Se tiver endereço, consulta “no token”
    if (tokenAddress.isNotEmpty) {
      KeyManager.validateEDDSAPublicKeyHex(tokenAddress);
      return getState(to: tokenAddress, method: METHOD_GET_TOKEN, data: data);
    }

    data["contract_version"] = TOKEN_CONTRACT_V2;

    // Caso contrário, query global (to:'')
    return getState(to: '', method: METHOD_GET_TOKEN, data: data);
  }

  Future<ContractOutput> listTokens({
    String ownerAddress = '',
    String symbol = '',
    String name = '',
    String tokenType = '',
    int page = 0,
    int limit = 20,
    bool ascending = true,
  }) async {
    final from = publicKeyHex ?? '';
    KeyManager.validateEDDSAPublicKeyHex(from);

    if (ownerAddress.isNotEmpty) {
      KeyManager.validateEDDSAPublicKeyHex(ownerAddress);
    }

    final data = <String, dynamic>{
      "owner": ownerAddress,
      "symbol": symbol,
      "name": name,
      "page": page,
      "limit": limit,
      "ascending": ascending,
      "contract_version": TOKEN_CONTRACT_V2,
    };
    if (tokenType.isNotEmpty) {
      data["token_type"] = tokenType;
    }

    return getState(to: '', method: METHOD_LIST_TOKENS, data: data);
  }

  Future<ContractOutput> getTokenBalance({
    required String tokenAddress,
    required String ownerAddress,
  }) async {
    final from = publicKeyHex ?? '';
    if (tokenAddress.isEmpty) throw ArgumentError('token address not set');
    if (ownerAddress.isEmpty) throw ArgumentError('owner address not set');

    KeyManager.validateEDDSAPublicKeyHex(from);
    KeyManager.validateEDDSAPublicKeyHex(tokenAddress);
    KeyManager.validateEDDSAPublicKeyHex(ownerAddress);

    final data = <String, dynamic>{"owner_address": ownerAddress};
    return getState(
      to: tokenAddress,
      method: METHOD_GET_TOKEN_BALANCE,
      data: data,
    );
  }

  Future<ContractOutput> getTokenBalanceNFT({
    required String tokenAddress,
    required String ownerAddress,
    required String uuid,
  }) async {
    final from = publicKeyHex ?? '';
    if (tokenAddress.isEmpty) throw ArgumentError('token address not set');
    if (ownerAddress.isEmpty) throw ArgumentError('owner address not set');
    if (uuid.isEmpty) throw ArgumentError('uuid not set');

    KeyManager.validateEDDSAPublicKeyHex(from);
    KeyManager.validateEDDSAPublicKeyHex(tokenAddress);
    KeyManager.validateEDDSAPublicKeyHex(ownerAddress);

    final data = <String, dynamic>{
      "owner_address": ownerAddress,
      "token_uuid": uuid,
    };
    return getState(
      to: tokenAddress,
      method: METHOD_GET_TOKEN_BALANCE_NFT,
      data: data,
    );
  }

  Future<ContractOutput> listTokenBalances({
    String tokenAddress = '',
    String ownerAddress = '',
    String tokenType = '',
    int page = 0,
    int limit = 20,
    bool ascending = true,
  }) async {
    final from = publicKeyHex ?? '';
    KeyManager.validateEDDSAPublicKeyHex(from);

    if (tokenAddress.isNotEmpty)
      KeyManager.validateEDDSAPublicKeyHex(tokenAddress);
    if (ownerAddress.isNotEmpty)
      KeyManager.validateEDDSAPublicKeyHex(ownerAddress);

    final data = <String, dynamic>{
      "address": tokenAddress,
      "owner_address": ownerAddress,
      "page": page,
      "limit": limit,
      "ascending": ascending,
      "contract_version": TOKEN_CONTRACT_V2,
    };
    if (tokenType.isNotEmpty) {
      data["token_type"] = tokenType;
    }

    return getState(to: '', method: METHOD_LIST_TOKEN_BALANCES, data: data);
  }
}
