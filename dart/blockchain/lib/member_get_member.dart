part of 'two_finance_blockchain.dart';

extension MemberGetMember on TwoFinanceBlockchain {
  // Future<ContractOutput> addMgM({
  //   required String address,
  //   required String owner,
  //   required String tokenAddress,
  //   required String faucetAddress,
  //   required String amount,
  //   required DateTime startAt,
  //   required DateTime expireAt,
  //   required bool paused,
  // }) async {
  //   final from = _publicKeyHex!;
  //   if (from.isEmpty) throw Exception("from address not set");
  //   KeyManager.validateEDDSAPublicKeyHex(from);

  //   if (owner.isEmpty) throw Exception("owner not set");
  //   KeyManager.validateEDDSAPublicKeyHex(owner);

  //   if (tokenAddress.isEmpty) throw Exception("token address not set");
  //   KeyManager.validateEDDSAPublicKeyHex(tokenAddress);

  //   if (faucetAddress.isEmpty) throw Exception("faucet address not set");
  //   KeyManager.validateEDDSAPublicKeyHex(faucetAddress);

  //   if (address.isEmpty) throw Exception("address not set");
  //   KeyManager.validateEDDSAPublicKeyHex(address);

  //   final to = address;
  //   final contractVersion = MGM_CONTRACT_V2;
  //   final method = METHOD_ADD_MGM;

  //   final data = {
  //     "address": address,
  //     "owner": owner,
  //     "token_address": tokenAddress,
  //     "faucet_address": faucetAddress,
  //     "amount": amount,
  //     "start_at": startAt.toIso8601String(),
  //     "expire_at": expireAt.toIso8601String(),
  //     "paused": paused,
  //   };

  //   try {
  //     return await signAndSendTransaction(
  //       from: from,
  //       to: to,
  //       contractVersion: contractVersion,
  //       method: method,
  //       data: data,
  //     );
  //   } catch (e) {
  //     throw Exception("failed to add mgm: $e");
  //   }
  // }

  // Future<ContractOutput> updateMgM({
  //   required String mgmAddress,
  //   required String amount,
  //   required DateTime startAt,
  //   required DateTime expireAt,
  // }) async {
  //   if (mgmAddress.isEmpty) throw Exception("address not set");
  //   KeyManager.validateEDDSAPublicKeyHex(mgmAddress);

  //   final from = _publicKeyHex!;
  //   if (from.isEmpty) throw Exception("from address not set");
  //   KeyManager.validateEDDSAPublicKeyHex(from);

  //   final to = mgmAddress;
  //   final contractVersion = MGM_CONTRACT_V2;
  //   final method = METHOD_UPDATE_MGM;

  //   final data = {
  //     "mgm_address": mgmAddress,
  //     "amount": amount,
  //     "start_at": startAt.toIso8601String(),
  //     "expire_at": expireAt.toIso8601String(),
  //   };

  //   return signAndSendTransaction(
  //     from: from,
  //     to: to,
  //     contractVersion: contractVersion,
  //     method: method,
  //     data: data,
  //   );
  // }

  // Future<ContractOutput> pauseMgM(String mgmAddress, bool pause) async {
  //   if (mgmAddress.isEmpty) throw Exception("address not set");
  //   KeyManager.validateEDDSAPublicKeyHex(mgmAddress);

  //   if (!pause) throw Exception("pause must be true: Pause: $pause");

  //   final from = _publicKeyHex!;
  //   if (from.isEmpty) throw Exception("from address not set");
  //   KeyManager.validateEDDSAPublicKeyHex(from);

  //   final to = mgmAddress;
  //   final contractVersion = MGM_CONTRACT_V2;
  //   final method = METHOD_PAUSE_MGM;

  //   final data = {"mgm_address": mgmAddress, "paused": pause};

  //   return signAndSendTransaction(
  //     from: from,
  //     to: to,
  //     contractVersion: contractVersion,
  //     method: method,
  //     data: data,
  //   );
  // }

  // Future<ContractOutput> unpauseMgM(String mgmAddress, bool pause) async {
  //   if (mgmAddress.isEmpty) throw Exception("address not set");
  //   KeyManager.validateEDDSAPublicKeyHex(mgmAddress);

  //   if (pause) throw Exception("pause must be false: Pause: $pause");

  //   final from = _publicKeyHex!;
  //   if (from.isEmpty) throw Exception("from address not set");
  //   KeyManager.validateEDDSAPublicKeyHex(from);

  //   final to = mgmAddress;
  //   final contractVersion = MGM_CONTRACT_V2;
  //   final method = METHOD_UNPAUSE_MGM;

  //   final data = {"mgm_address": mgmAddress, "paused": pause};

  //   return signAndSendTransaction(
  //     from: from,
  //     to: to,
  //     contractVersion: contractVersion,
  //     method: method,
  //     data: data,
  //   );
  // }
}
