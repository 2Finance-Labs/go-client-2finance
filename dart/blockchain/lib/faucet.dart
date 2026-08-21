part of 'two_finance_blockchain.dart';

extension Faucet on TwoFinanceBlockchain {
  // Future<ContractOutput> addFaucet(
  //   String address,
  //   String owner,
  //   String tokenAddress,
  //   DateTime startTime,
  //   DateTime expireTime,
  //   bool paused,
  //   int requestLimit,
  //   String claimAmount,
  //   Duration claimIntervalDuration,
  // ) async {
  //   final from = _publicKeyHex!;
  //   final to = address;
  //   final contractVersion = FAUCET_CONTRACT_V2;
  //   final method = METHOD_ADD_FAUCET;
  //   if (from.isEmpty) throw Exception("from address not set");
  //   KeyManager.validateEDDSAPublicKeyHex(from);

  //   if (address.isEmpty) throw Exception("address not set");
  //   KeyManager.validateEDDSAPublicKeyHex(address);

  //   if (owner.isEmpty) throw Exception("owner not set");
  //   if (tokenAddress.isEmpty) throw Exception("token address not set");

  //   KeyManager.validateEDDSAPublicKeyHex(owner);
  //   KeyManager.validateEDDSAPublicKeyHex(tokenAddress);

  //   if (claimAmount.isEmpty) throw Exception("amount not set");

  //   final data = {
  //     "address": address,
  //     "owner": owner,
  //     "token_address": tokenAddress,
  //     "start_time": startTime.toIso8601String(),
  //     "expire_time": expireTime.toIso8601String(),
  //     "paused": paused,
  //     "request_limit": requestLimit,
  //     "claim_amount": claimAmount,
  //     "claim_interval_duration": claimIntervalDuration.inSeconds,
  //   };

  //   try {
  //     return await signAndSendTransaction(chainID: chainID, from: from, to: to, method: method, data: data, version:version, uuid7:uuid7);
  //   } catch (e) {
  //     throw Exception("failed to send transaction: $e");
  //   }
  // }

  // Future<ContractOutput> updateFaucet(
  //   String address,
  //   DateTime startTime,
  //   DateTime expireTime,
  //   int requestLimit,
  //   Map<String, int> requestsByUser,
  //   String claimAmount,
  //   Duration claimIntervalDuration,
  //   Map<String, DateTime> lastClaimByUser,
  // ) async {
  //   final from = _publicKeyHex!;
  //   final to = address;
  //   const contractVersion = FAUCET_CONTRACT_V2;
  //   const method = METHOD_UPDATE_FAUCET;
  //   if (from.isEmpty) throw Exception("from address not set");
  //   KeyManager.validateEDDSAPublicKeyHex(from);

  //   if (address.isEmpty) throw Exception("address not set");
  //   final data = {
  //     "address": address,
  //     "start_time": startTime.toIso8601String(),
  //     "expire_time": expireTime.toIso8601String(),
  //     "request_limit": requestLimit,
  //     "requests_by_user": requestsByUser,
  //     "claim_amount": claimAmount,
  //     "claim_interval_duration": claimIntervalDuration.inSeconds,
  //     "last_claim_by_user": lastClaimByUser.map((k, v) => MapEntry(k, v.toIso8601String())),
  //   };

  //   try {
  //     return await signAndSendTransaction(chainID: chainID, from: from, to: to, method: method, data: data, version:version, uuid7:uuid7);
  //   } catch (e) {
  //     throw Exception("failed to send transaction: $e");
  //   }
  // }

  // Future<ContractOutput> pauseFaucet(String address, bool pause) async {
  //   if (address.isEmpty) throw Exception("address not set");
  //   KeyManager.validateEDDSAPublicKeyHex(address);

  //   if (!pause) throw Exception("pause must be true: Pause: $pause");

  //   final from = _publicKeyHex!;
  //   if (from.isEmpty) throw Exception("from address not set");
  //   KeyManager.validateEDDSAPublicKeyHex(from);

  //   final to = address;
  //   const contractVersion = FAUCET_CONTRACT_V2;
  //   const method = METHOD_PAUSE_FAUCET;

  //   final data = {
  //     "address": address,
  //     "paused": pause,
  //   };

  //   try {
  //     return await signAndSendTransaction(chainID: chainID, from: from, to: to, method: method, data: data, version:version, uuid7:uuid7);
  //   } catch (e) {
  //     throw Exception("failed to send transaction: $e");
  //   }
  // }

  // Future<ContractOutput> unpauseFaucet(String address, bool pause) async {
  //   if (address.isEmpty) throw Exception("address not set");
  //   KeyManager.validateEDDSAPublicKeyHex(address);

  //   if (pause) throw Exception("pause must be false: Pause: $pause");

  //   final from = _publicKeyHex!;
  //   if (from.isEmpty) throw Exception("from address not set");
  //   KeyManager.validateEDDSAPublicKeyHex(from);

  //   final to = address;
  //   const contractVersion = FAUCET_CONTRACT_V2;
  //   const method = METHOD_UNPAUSE_FAUCET;

  //   final data = {
  //     "address": address,
  //     "pause": pause,
  //   };

  //   try {
  //     return await signAndSendTransaction(chainID: chainID, from: from, to: to, method: method, data: data, version:version, uuid7:uuid7);
  //   } catch (e) {
  //     throw Exception("failed to send transaction: $e");
  //   }
  // }

  // Future<ContractOutput> depositFunds(String address, String tokenAddress, String amount) async {
  //   if (address.isEmpty) throw Exception("address not set");
  //   KeyManager.validateEDDSAPublicKeyHex(address);

  //   if (tokenAddress.isEmpty) throw Exception("address not set");
  //   KeyManager.validateEDDSAPublicKeyHex(tokenAddress);

  //   final from = _publicKeyHex!;
  //   if (from.isEmpty) throw Exception("from address not set");
  //   KeyManager.validateEDDSAPublicKeyHex(from);

  //   final to = address;
  //   const contractVersion = FAUCET_CONTRACT_V2;
  //   const method = METHOD_DEPOSIT_FUNDS;

  //   final data = {
  //     "address": address,
  //     "token_address": tokenAddress,
  //     "amount": amount,
  //   };

  //   try {
  //     return await signAndSendTransaction(chainID: chainID, from: from, to: to, method: method, data: data, version:version, uuid7:uuid7);
  //   } catch (e) {
  //     throw Exception("failed to send transaction: $e");
  //   }
  // }

  // Future<ContractOutput> withdrawFunds(String address, String tokenAddress, String amount) async {
  //   if (address.isEmpty) throw Exception("address not set");
  //   KeyManager.validateEDDSAPublicKeyHex(address);

  //   if (tokenAddress.isEmpty) throw Exception("address not set");
  //   KeyManager.validateEDDSAPublicKeyHex(tokenAddress);

  //   final from = _publicKeyHex!;
  //   if (from.isEmpty) throw Exception("from address not set");
  //   KeyManager.validateEDDSAPublicKeyHex(from);

  //   final to = address;
  //   const contractVersion = FAUCET_CONTRACT_V2;
  //   const method = METHOD_WITHDRAW_FUNDS;

  //   final data = {
  //     "address": address,
  //     "token_address": tokenAddress,
  //     "amount": amount,
  //   };

  //   try {
  //     return await signAndSendTransaction(chainID: chainID, from: from, to: to, method: method, data: data, version:version, uuid7:uuid7);
  //   } catch (e) {
  //     throw Exception("failed to send transaction: $e");
  //   }
  // }

  // Future<ContractOutput> updateRequestLimitPerUser(String address, int requestLimit) async {
  //   if (address.isEmpty) throw Exception("address not set");
  //   KeyManager.validateEDDSAPublicKeyHex(address);

  //   if (requestLimit < 0) throw Exception("request limit less than zero: $requestLimit");

  //   final from = _publicKeyHex!;
  //   if (from.isEmpty) throw Exception("from address not set");
  //   KeyManager.validateEDDSAPublicKeyHex(from);

  //   final to = address;
  //   const contractVersion = FAUCET_CONTRACT_V2;
  //   const method = METHOD_UPDATE_REQUEST_LIMIT_PER_USER;

  //   final data = {
  //     "address": address,
  //     "request_limit": requestLimit,
  //   };

  //   try {
  //     return await signAndSendTransaction(chainID: chainID, from: from, to: to, method: method, data: data, version:version, uuid7:uuid7);
  //   } catch (e) {
  //     throw Exception("failed to send transaction: $e");
  //   }
  // }

  // Future<ContractOutput> claimFunds(String address) async {
  //   if (address.isEmpty) throw Exception("address not set");
  //   KeyManager.validateEDDSAPublicKeyHex(address);

  //   final from = _publicKeyHex!;
  //   if (from.isEmpty) throw Exception("from address not set");
  //   KeyManager.validateEDDSAPublicKeyHex(from);

  //   const contractVersion = FAUCET_CONTRACT_V2;
  //   const method = METHOD_CLAIM_FUNDS;
  //   final to = address;

  //   final data = {
  //     "address": address,
  //   };

  //   try {
  //     return await signAndSendTransaction(chainID: chainID, from: from, to: to, method: method, data: data, version:version, uuid7:uuid7);
  //   } catch (e) {
  //     throw Exception("failed to send transaction: $e");
  //   }
  // }

  // Future<ContractOutput> getFaucet(String faucetAddress) async {
  //   final from = _publicKeyHex!;
  //   if (from.isEmpty) throw Exception("from address not set");

  //   if (faucetAddress.isEmpty) throw Exception("faucet address must be set");

  //   KeyManager.validateEDDSAPublicKeyHex(from);
  //   KeyManager.validateEDDSAPublicKeyHex(faucetAddress);

  //   const contractVersion = FAUCET_CONTRACT_V2;
  //   const method = METHOD_GET_FAUCET;
  //   final data = {
  //     "address": faucetAddress,
  //   };

  //   try {
  //     return await getState(contractVersion: contractVersion, method: method, data: data);
  //   } catch (e) {
  //     throw Exception("failed to get state: $e");
  //   }
  // }

  // Future<ContractOutput> listFaucets(
  //   String ownerAddress,
  //   int page,
  //   int limit,
  //   bool ascending,
  // ) async {
  //   final from = _publicKeyHex!;
  //   if (from.isEmpty) throw Exception("from address not set");
  //   KeyManager.validateEDDSAPublicKeyHex(from);

  //   if (ownerAddress.isNotEmpty) {
  //     KeyManager.validateEDDSAPublicKeyHex(ownerAddress);
  //       throw Exception("invalid owner address");
  //   }

  //   if (page < 1) throw Exception("page must be greater than 0");
  //   if (limit < 1) throw Exception("limit must be greater than 0");

  //   const contractVersion = FAUCET_CONTRACT_V2;
  //   const method = METHOD_LIST_FAUCETS;

  //   final data = {
  //     "owner": ownerAddress,
  //     "page": page,
  //     "limit": limit,
  //     "ascending": ascending,
  //   };

  //   try {
  //     return await getState(contractVersion: contractVersion, method: method, data: data);
  //   } catch (e) {
  //     throw Exception("failed to list faucet states: $e");
  //   }
  // }
}
