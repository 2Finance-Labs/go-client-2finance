import 'dart:async';
import 'dart:convert';
import 'dart:io';
import 'dart:math';
import 'dart:typed_data';

import 'package:crypto/crypto.dart';
import 'package:test/test.dart';
import 'package:two_finance_blockchain/blockchain/contract/tokenV2/constants.dart';
import 'package:two_finance_blockchain/blockchain/contract/tokenV2/models/balance.dart';
import 'package:two_finance_blockchain/blockchain/keys/keys.dart';
import 'package:two_finance_blockchain/blockchain/types/types.dart';
import 'package:two_finance_blockchain/blockchain/utils/marshal.dart';
import 'package:two_finance_blockchain/config/config.dart';
import 'package:two_finance_blockchain/infra/http/http_transport.dart';
import 'package:two_finance_blockchain/two_finance_blockchain.dart';

class TestUser {
  final String name;
  final String publicKey;
  final String privateKey;

  const TestUser({
    required this.name,
    required this.publicKey,
    required this.privateKey,
  });
}

class MintedNftPrize {
  final String tokenAddress;
  final String tokenUuid;
  final String symbol;

  const MintedNftPrize({
    required this.tokenAddress,
    required this.tokenUuid,
    required this.symbol,
  });
}

typedef JsonMap = Map<String, dynamic>;

String repeatHex(String hexChar, int len) => List.filled(len, hexChar).join();

Future<String> validPublicKeyHex() async {
  final km = KeyManager();
  final keyPair = await km.generateKeyEd25519();
  return keyPair.publicKey;
}

Future<KeyPair2Finance> validKeyPair() async {
  final km = KeyManager();
  return km.generateKeyEd25519();
}

Future<TestUser> newTestUser(String name) async {
  final kp = await validKeyPair();

  return TestUser(
    name: name,
    publicKey: kp.publicKey,
    privateKey: kp.privateKey,
  );
}

String sha256Hex(String input) {
  final bytes = utf8.encode(input);
  final digest = sha256.convert(bytes);
  return digest.bytes.map((b) => b.toRadixString(16).padLeft(2, '0')).join();
}

Future<TwoFinanceBlockchain> setupClient() async {
  try {
    await Config.loadConfig(
      env: 'dev',
      path: 'packages/two_finance_blockchain/assets/.env',
    );

    final networkUrl =
        Platform.environment['TWO_FINANCE_NETWORK_URL'] ??
        'http://127.0.0.1:19295';
    final transport = HttpFinanceNetworkTransport(baseUrl: networkUrl);
    print('Initializing TwoFinanceBlockchain plugin...');

    final keyManager = KeyManager();
    final chainID = int.tryParse(Platform.environment['CHAIN_ID'] ?? '2') ?? 2;
    if (chainID < 1 || chainID > 2) {
      throw ArgumentError('CHAIN_ID must be 1 or 2, got $chainID');
    }

    final plugin = TwoFinanceBlockchain(
      keyManager: keyManager,
      mqttClient: transport,
      chainID: chainID,
    );

    await plugin.initialize();
    return plugin;
  } catch (e) {
    print('⚠️ Error initializing TwoFinanceBlockchain: $e');
    rethrow;
  }
}

Future<void> teardownClient(TwoFinanceBlockchain c) async {
  // The shared HTTP transport has no session or subscription to tear down.
}

String randSuffix(int n) {
  final random = Random.secure();
  final values = List<int>.generate(n, (_) => random.nextInt(256));
  return hexEncode(Uint8List.fromList(values)).substring(0, n);
}

String hexEncode(Uint8List bytes) {
  final buffer = StringBuffer();
  for (final b in bytes) {
    buffer.write(b.toRadixString(16).padLeft(2, '0'));
  }
  return buffer.toString();
}

String amt(int unscaled, int decimals) {
  final p = BigInt.from(10).pow(decimals);
  final v = BigInt.from(unscaled) * p;
  return v.toString();
}

Future<void> waitUntil(Duration d, bool Function() pred) async {
  final completer = Completer<void>();
  final end = DateTime.now().add(d);

  Timer.periodic(const Duration(milliseconds: 20), (timer) {
    if (pred()) {
      timer.cancel();
      completer.complete();
    } else if (DateTime.now().isAfter(end)) {
      timer.cancel();
      completer.completeError(
        TimeoutException("timeout waiting for condition"),
      );
    }
  });

  return completer.future;
}

Future<void> expectFtBalance(
  TwoFinanceBlockchain c, {
  required String tokenAddress,
  required String ownerAddress,
  required String expectedAmount,
}) async {
  final out = await c.getTokenBalance(
    tokenAddress: tokenAddress,
    ownerAddress: ownerAddress,
  );

  expect(out.states, isNotNull);
  expect(out.states!, isNotEmpty);

  final balance = unmarshalState(
    out.states!.first.object,
    (json) => BalanceState.fromJson(json),
  );

  expect(balance.ownerAddress, equals(ownerAddress));
  expect(balance.tokenAddress, equals(tokenAddress));
  expect(balance.amount, equals(expectedAmount));
}

List<String> parseTokenUuidList(dynamic raw) {
  final items = List<dynamic>.from(raw as List);

  return items.map((item) {
    if (item is String) return item;

    final map = Map<String, dynamic>.from(item as Map);

    final value = map['token_uuid'] ?? map['uuid'];
    if (value == null) {
      throw StateError('token_uuid_list item sem token_uuid/uuid: $map');
    }

    return value.toString();
  }).toList();
}

Future<BalanceState> getFtBalanceState(
  TwoFinanceBlockchain c, {
  required String tokenAddress,
  required String ownerAddress,
}) async {
  final out = await c.getTokenBalance(
    tokenAddress: tokenAddress,
    ownerAddress: ownerAddress,
  );

  expect(out.states, isNotNull);
  expect(out.states!, isNotEmpty);

  final balance = unmarshalState(
    out.states!.first.object,
    (json) => BalanceState.fromJson(json),
  );

  expect(balance.ownerAddress, equals(ownerAddress));
  expect(balance.tokenAddress, equals(tokenAddress));

  return balance;
}

Future<String> getFtBalanceAmount(
  TwoFinanceBlockchain c, {
  required String tokenAddress,
  required String ownerAddress,
}) async {
  final balance = await getFtBalanceState(
    c,
    tokenAddress: tokenAddress,
    ownerAddress: ownerAddress,
  );

  expect(balance.amount, isNotNull);
  return balance.amount!;
}

Future<BalanceState> getNftBalanceState(
  TwoFinanceBlockchain c, {
  required String tokenAddress,
  required String ownerAddress,
  required String uuid,
}) async {
  final out = await c.getTokenBalanceNFT(
    tokenAddress: tokenAddress,
    ownerAddress: ownerAddress,
    uuid: uuid,
  );

  expect(out, isA<ContractOutput>());
  expect(out.states, isNotNull);
  expect(out.states!, isNotEmpty);

  final balance = unmarshalState(
    out.states!.first.object,
    (json) => BalanceState.fromJson(json),
  );

  expect(balance.tokenAddress, equals(tokenAddress));
  expect(balance.ownerAddress, equals(ownerAddress));
  expect(balance.tokenUuid, equals(uuid));

  return balance;
}

Future<void> expectNftBalance(
  TwoFinanceBlockchain c, {
  required String tokenAddress,
  required String ownerAddress,
  required String uuid,
  required String expectedAmount,
  required String expectedTokenType,
  bool? expectedBurned,
}) async {
  final balance = await getNftBalanceState(
    c,
    tokenAddress: tokenAddress,
    ownerAddress: ownerAddress,
    uuid: uuid,
  );

  expect(balance.amount, equals(expectedAmount));
  expect(balance.tokenType, equals(expectedTokenType));

  if (expectedBurned != null) {
    expect(balance.burned, equals(expectedBurned));
  }
}

Future<void> enterRaffleFtAndExpect(
  TwoFinanceBlockchain c, {
  required TestUser user,
  required String raffleAddress,
  required String payTokenAddress,
  required int tickets,
  required String expectedPaid,
  required String requestUuid,
  bool setSigner = true,
}) async {
  if (setSigner) {
    await c.setPrivateKey(user.privateKey);
  }

  final out = await c.enterRaffle(
    address: raffleAddress,
    tickets: tickets,
    payTokenAddress: payTokenAddress,
    tokenType: TOKEN_TYPE_FUNGIBLE,
    uuid: requestUuid,
  );

  expect(out, isA<ContractOutput>());
  expect(out.logs, isNotNull);
  expect(out.logs!, isNotEmpty);

  final log = out.logs!.first;
  expect(log.contractAddress, equals(raffleAddress));
  expect(log.logType, equals('Raffle_Entered'));

  final event = unmarshalEvent<Map<String, dynamic>>(
    log.event,
    (json) => Map<String, dynamic>.from(json as Map),
  );

  expect(event['raffle_address'], equals(raffleAddress));
  expect(event['entrant'], equals(user.publicKey));
  expect(event['tickets'], equals(tickets));
  expect(event['paid'], equals(expectedPaid));
  expect(event['pay_token_address'], equals(payTokenAddress));
  expect(event['uuid'], isA<String>());
  expect((event['uuid'] as String).isNotEmpty, isTrue);
}

Future<void> allowAndMintFtUsers(
  TwoFinanceBlockchain c, {
  required String tokenAddress,
  required List<TestUser> users,
  required String amount,
  required int decimals,
}) async {
  final outAllow = await c.allowUsers(tokenAddress, {
    for (final user in users) user.publicKey: true,
  });

  expect(outAllow, isA<ContractOutput>());
  expect(outAllow.logs, isNotNull);
  expect(outAllow.logs!, isNotEmpty);
  expect(outAllow.logs!.first.logType, equals('Token_AllowedUsersAdded'));
  expect(outAllow.logs!.first.contractAddress, equals(tokenAddress));

  final expectedAmount = (BigInt.parse(amount) * BigInt.from(10).pow(decimals))
      .toString();

  for (final user in users) {
    final outMint = await c.mintToken(
      tokenAddress: tokenAddress,
      mintTo: user.publicKey,
      amount: amount,
      decimals: decimals,
      tokenType: TOKEN_TYPE_FUNGIBLE,
    );

    expect(outMint, isA<ContractOutput>());
    expect(outMint.logs, isNotNull);
    expect(outMint.logs!, isNotEmpty);
    expect(outMint.logs!.length, equals(3));
    expect(outMint.logs![0].logType, equals('Token_Minted_FT'));
    expect(outMint.logs![1].logType, equals('Token_TotalSupplyIncreased'));
    expect(outMint.logs![2].logType, equals('Token_BalanceIncreased_FT'));

    await expectFtBalance(
      c,
      tokenAddress: tokenAddress,
      ownerAddress: user.publicKey,
      expectedAmount: expectedAmount,
    );
  }
}

Future<String> createBasicToken(
  TwoFinanceBlockchain c, {
  required String ownerPrivateKey,
  required String ownerPublicKey,
  required int decimals,
  required bool requireFee,
  required String tokenType,
  required bool stablecoin,
}) async {
  await c.setPrivateKey(ownerPrivateKey);

  final deployedToken = await c.deployContract1(TOKEN_CONTRACT_V2);

  expect(deployedToken, isA<ContractOutput>());
  expect(deployedToken.logs, isNotNull);
  expect(deployedToken.logs!, isNotEmpty);

  final deployLog = deployedToken.logs!.first;
  final tokenAddress = deployLog.contractAddress;

  expect(tokenAddress, isNotEmpty);

  final symbol =
      '2F${DateTime.now().microsecondsSinceEpoch.toRadixString(16)}${randSuffix(4).toUpperCase()}';
  final totalSupply = tokenType == TOKEN_TYPE_FUNGIBLE
      ? amt(1000000, decimals)
      : '1';

  final feeTiersList = requireFee
      ? <Map<String, dynamic>>[
          {
            'min_amount': '0',
            'max_amount': amt(10000, decimals),
            'min_volume': '0',
            'max_volume': amt(100000, decimals),
            'fee_bps': 50,
          },
        ]
      : <Map<String, dynamic>>[];

  final outAddToken = await c.addToken(
    address: tokenAddress,
    symbol: symbol,
    name: '2Finance',
    decimals: decimals,
    totalSupply: totalSupply,
    description: 'e2e token created by tests',
    owner: ownerPublicKey,
    image: 'https://example.com/image.png',
    website: 'https://example.com',
    tagsSocialMedia: {'twitter': 'https://twitter.com/2finance'},
    tagsCategory: {'category': 'DeFi'},
    tags: {'tag1': 'DeFi', 'tag2': 'Blockchain'},
    creator: '2Finance Test',
    creatorWebsite: 'https://creator.example',
    allowedUsers: {},
    blockedUsers: {},
    frozenAccounts: {},
    feeTiersList: feeTiersList,
    feeAddress: ownerPublicKey,
    freezeAuthorityRevoked: false,
    mintAuthorityRevoked: false,
    updateAuthorityRevoked: false,
    paused: false,
    expiredAt: DateTime.now().toUtc().add(const Duration(days: 365)),
    assetGlbUri: 'https://example.com/asset.glb',
    tokenType: tokenType,
    transferable: true,
    stablecoin: stablecoin,
  );

  expect(outAddToken, isA<ContractOutput>());
  expect(outAddToken.logs, isNotNull);
  expect(outAddToken.logs!, isNotEmpty);

  final addTokenLog = outAddToken.logs!.first;
  expect(addTokenLog.contractAddress, equals(tokenAddress));
  expect(addTokenLog.logType, equals('Token_Created'));

  final tokenEvent = unmarshalEvent<Map<String, dynamic>>(
    addTokenLog.event,
    (json) => Map<String, dynamic>.from(json as Map),
  );

  expect(tokenEvent['address'], equals(tokenAddress));
  expect(tokenEvent['token_type'], equals(tokenType));

  return tokenAddress;
}

Future<MintedNftPrize> createAndMintNftPrize(
  TwoFinanceBlockchain c, {
  required TestUser ownerUser,
  String name = 'Raffle Prize NFT',
  String description = 'raffle nft prize e2e',
  String image = 'https://example.com/raffle-prize-nft.png',
  int mintCount = 2,
}) async {
  await c.setPrivateKey(ownerUser.privateKey);

  final deployedPrize = await c.deployContract1(TOKEN_CONTRACT_V2);
  expect(deployedPrize, isA<ContractOutput>());
  expect(deployedPrize.logs, isNotNull);
  expect(deployedPrize.logs!, isNotEmpty);

  final prizeTokenAddress = deployedPrize.logs!.first.contractAddress;
  expect(prizeTokenAddress, isNotEmpty);

  final prizeFeeUser = await newTestUser('prize-fee');
  final prizeExpiredAt = DateTime.now().toUtc().add(const Duration(days: 365));
  final prizeSymbol =
      'RFNFT_${DateTime.now().millisecondsSinceEpoch.toString().substring(8)}';

  final outAddPrizeToken = await c.addToken(
    address: prizeTokenAddress,
    symbol: prizeSymbol,
    name: name,
    decimals: 0,
    totalSupply: '1',
    description: description,
    owner: ownerUser.publicKey,
    image: image,
    website: 'https://example.com',
    tagsSocialMedia: const {'twitter': 'https://twitter.com/2finance'},
    tagsCategory: const {'category': 'Collectibles'},
    tags: const {'tag1': 'NFT', 'tag2': 'Raffle'},
    creator: '2Finance Test',
    creatorWebsite: 'https://creator.example',
    allowedUsers: const <String, bool>{},
    blockedUsers: const <String, bool>{},
    frozenAccounts: const <String, dynamic>{},
    feeTiersList: const <Map<String, dynamic>>[],
    feeAddress: prizeFeeUser.publicKey,
    freezeAuthorityRevoked: false,
    mintAuthorityRevoked: false,
    updateAuthorityRevoked: false,
    paused: false,
    expiredAt: prizeExpiredAt,
    assetGlbUri: 'https://example.com/prize.glb',
    tokenType: TOKEN_TYPE_NON_FUNGIBLE,
    transferable: true,
    stablecoin: false,
  );

  expect(outAddPrizeToken, isA<ContractOutput>());
  expect(outAddPrizeToken.logs, isNotNull);
  expect(outAddPrizeToken.logs!, isNotEmpty);

  final prizeAddLog = outAddPrizeToken.logs!.first;
  expect(prizeAddLog.contractAddress, equals(prizeTokenAddress));
  expect(prizeAddLog.logType, equals('Token_Created'));

  final prizeTokenEvent = unmarshalEvent<Map<String, dynamic>>(
    prizeAddLog.event,
    (json) => Map<String, dynamic>.from(json as Map),
  );

  expect(prizeTokenEvent['address'], equals(prizeTokenAddress));
  expect(prizeTokenEvent['token_type'], equals(TOKEN_TYPE_NON_FUNGIBLE));

  final outMintPrizeToken = await c.mintToken(
    tokenAddress: prizeTokenAddress,
    mintTo: ownerUser.publicKey,
    amount: mintCount.toString(),
    decimals: 0,
    tokenType: TOKEN_TYPE_NON_FUNGIBLE,
  );

  expect(outMintPrizeToken, isA<ContractOutput>());
  expect(outMintPrizeToken.logs, isNotNull);
  expect(outMintPrizeToken.logs!, isNotEmpty);

  final mintPrizeLog = outMintPrizeToken.logs!.first;
  expect(mintPrizeLog.contractAddress, equals(prizeTokenAddress));
  expect(mintPrizeLog.logType, equals('Token_Minted_NFT'));

  final mintPrizeEvent = unmarshalEvent<Map<String, dynamic>>(
    mintPrizeLog.event,
    (json) => Map<String, dynamic>.from(json as Map),
  );

  expect(mintPrizeEvent['token_address'], equals(prizeTokenAddress));
  expect(mintPrizeEvent['mint_to'], equals(ownerUser.publicKey));
  expect(mintPrizeEvent['token_type'], equals(TOKEN_TYPE_NON_FUNGIBLE));

  final prizeUuidList = parseTokenUuidList(mintPrizeEvent['token_uuid_list']);
  expect(prizeUuidList, hasLength(mintCount));

  final prizeUuid = prizeUuidList.first;
  expect(prizeUuid.isNotEmpty, isTrue);

  await expectNftBalance(
    c,
    tokenAddress: prizeTokenAddress,
    ownerAddress: ownerUser.publicKey,
    uuid: prizeUuid,
    expectedAmount: '1',
    expectedTokenType: TOKEN_TYPE_NON_FUNGIBLE,
    expectedBurned: false,
  );

  return MintedNftPrize(
    tokenAddress: prizeTokenAddress,
    tokenUuid: prizeUuid,
    symbol: prizeSymbol,
  );
}

Future<void> prepareRaffleParticipantsAndFunding(
  TwoFinanceBlockchain c, {
  required TestUser ownerUser,
  required List<TestUser> players,
  required String paymentTokenAddress,
  required String prizeTokenAddress,
  required String raffleAddress,
  required String fundingAmount,
}) async {
  await c.setPrivateKey(ownerUser.privateKey);

  final allowedUsers = <String, bool>{
    ownerUser.publicKey: true,
    for (final player in players) player.publicKey: true,
    raffleAddress: true,
  };

  final outAllowPayToken = await c.allowUsers(
    paymentTokenAddress,
    allowedUsers,
  );

  expect(outAllowPayToken, isA<ContractOutput>());
  expect(outAllowPayToken.logs, isNotNull);
  expect(outAllowPayToken.logs!, isNotEmpty);
  expect(
    outAllowPayToken.logs!.first.logType,
    equals('Token_AllowedUsersAdded'),
  );
  expect(
    outAllowPayToken.logs!.first.contractAddress,
    equals(paymentTokenAddress),
  );

  final outAllowPrizeToken = await c.allowUsers(
    prizeTokenAddress,
    allowedUsers,
  );

  expect(outAllowPrizeToken, isA<ContractOutput>());
  expect(outAllowPrizeToken.logs, isNotNull);
  expect(outAllowPrizeToken.logs!, isNotEmpty);
  expect(
    outAllowPrizeToken.logs!.first.logType,
    equals('Token_AllowedUsersAdded'),
  );
  expect(
    outAllowPrizeToken.logs!.first.contractAddress,
    equals(prizeTokenAddress),
  );

  for (final player in players) {
    final outTransfer = await c.transferToken(
      tokenAddress: paymentTokenAddress,
      transferTo: player.publicKey,
      amount: fundingAmount,
      decimals: 0,
      tokenType: TOKEN_TYPE_FUNGIBLE,
      uuid: '',
    );

    expect(outTransfer, isA<ContractOutput>());
    expect(outTransfer.logs, isNotNull);
    expect(outTransfer.logs!, isNotEmpty);
    expect(
      outTransfer.logs!.first.contractAddress,
      equals(paymentTokenAddress),
    );

    await expectFtBalance(
      c,
      tokenAddress: paymentTokenAddress,
      ownerAddress: player.publicKey,
      expectedAmount: fundingAmount,
    );
  }
}
