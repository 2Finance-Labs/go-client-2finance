import 'package:test/test.dart';
import 'package:two_finance_blockchain/blockchain/contract/dropV2/constants.dart';
import 'package:two_finance_blockchain/blockchain/contract/tokenV2/constants.dart';
import 'package:two_finance_blockchain/blockchain/types/types.dart';
import 'package:two_finance_blockchain/blockchain/utils/marshal.dart';
import 'package:two_finance_blockchain/two_finance_blockchain.dart';

import 'drop_test_helpers.dart';
import '../../../helpers/helpers.dart';

Future<(String, List<String>)> _createAndMintDropNftToken(
  TwoFinanceBlockchain c, {
  required TestUser ownerUser,
}) async {
  await c.setPrivateKey(ownerUser.privateKey);

  final deployedToken = await c.deployContract1(TOKEN_CONTRACT_V2);
  expect(deployedToken.logs, isNotNull);
  expect(deployedToken.logs!, isNotEmpty);

  final tokenAddress = deployedToken.logs!.first.contractAddress;
  expect(tokenAddress, isNotEmpty);

  final feeUser = await newTestUser('drop-nft-fee');
  final symbol =
      'DNFT_${DateTime.now().millisecondsSinceEpoch.toString().substring(8)}';

  final outAddToken = await c.addToken(
    address: tokenAddress,
    symbol: symbol,
    name: 'Drop NFT Test Token',
    decimals: 0,
    totalSupply: '1',
    description: 'drop nft token e2e',
    owner: ownerUser.publicKey,
    image: 'https://example.com/drop-nft-token.png',
    website: 'https://example.com',
    tagsSocialMedia: const {'twitter': 'https://twitter.com/2finance'},
    tagsCategory: const {'category': 'Collectibles'},
    tags: const {'tag1': 'NFT', 'tag2': 'Drop'},
    creator: '2Finance Test',
    creatorWebsite: 'https://creator.example',
    allowedUsers: const <String, bool>{},
    blockedUsers: const <String, bool>{},
    frozenAccounts: const <String, dynamic>{},
    feeTiersList: const <Map<String, dynamic>>[],
    feeAddress: feeUser.publicKey,
    freezeAuthorityRevoked: false,
    mintAuthorityRevoked: false,
    updateAuthorityRevoked: false,
    paused: false,
    expiredAt: DateTime.now().toUtc().add(const Duration(days: 365)),
    assetGlbUri: 'https://example.com/drop-nft.glb',
    tokenType: TOKEN_TYPE_NON_FUNGIBLE,
    transferable: true,
    stablecoin: false,
  );
  expect(outAddToken.logs, isNotNull);
  expect(outAddToken.logs!, isNotEmpty);

  final outMintToken = await c.mintToken(
    tokenAddress: tokenAddress,
    mintTo: ownerUser.publicKey,
    amount: '2',
    decimals: 0,
    tokenType: TOKEN_TYPE_NON_FUNGIBLE,
  );
  expect(outMintToken.logs, isNotNull);
  expect(outMintToken.logs!, isNotEmpty);
  expect(outMintToken.logs!.first.logType, equals('Token_Minted_NFT'));

  final mintEvent = unmarshalEvent<JsonMap>(
    outMintToken.logs!.first.event,
    (json) => Map<String, dynamic>.from(json as Map),
  );

  final uuids = parseTokenUuidList(mintEvent['token_uuid_list']);
  expect(uuids, hasLength(2));

  return (tokenAddress, uuids);
}

Future<String> _resolveClaimedUuid(
  TwoFinanceBlockchain c, {
  required String tokenAddress,
  required String claimerAddress,
  required List<String> candidateUuids,
}) async {
  for (final uuid in candidateUuids) {
    try {
      await expectNftBalance(
        c,
        tokenAddress: tokenAddress,
        ownerAddress: claimerAddress,
        uuid: uuid,
        expectedAmount: '1',
        expectedTokenType: TOKEN_TYPE_NON_FUNGIBLE,
      );
      return uuid;
    } catch (_) {
      // try next UUID
    }
  }

  throw StateError('could not determine claimed NFT UUID');
}

void main() {
  final testTimeout = Timeout(Duration(minutes: 4));

  group('DropV2 NFT E2E', () {
    test(
      'E2E: create + deposit nft + attest + claim nft + withdraw nft',
      () async {
        final c = await setupClient();
        addTearDown(() => teardownClient(c));

        final ownerUser = await newTestUser('owner');
        final claimerUser = await newTestUser('claimer');
        final oracleUser = await newTestUser('oracle');

        // ------------------
        // SETUP NFT TOKEN
        // ------------------
        final nftSetup = await _createAndMintDropNftToken(
          c,
          ownerUser: ownerUser,
        );
        final nftTokenAddress = nftSetup.$1;
        final firstUuid = nftSetup.$2.first;
        final secondUuid = nftSetup.$2.last;
        final placeholderTokenUser = await newTestUser(
          'drop-placeholder-token',
        );
        final placeholderTokenAddress = placeholderTokenUser.publicKey;

        // ------------------
        // DEPLOY DROP
        // ------------------
        final deployed = await c.deployContract1(DROP_CONTRACT_V2);
        expect(deployed, isA<ContractOutput>());
        expect(deployed.logs, isNotNull);
        expect(deployed.logs!, isNotEmpty);

        final dropAddress = deployed.logs!.first.contractAddress;
        expect(dropAddress, isNotEmpty);

        // ------------------
        // ALLOW TOKEN USERS
        // ------------------
        final outAllowTokenUsers = await c.allowUsers(nftTokenAddress, {
          ownerUser.publicKey: true,
          claimerUser.publicKey: true,
          dropAddress: true,
        });
        expect(outAllowTokenUsers.logs, isNotNull);
        expect(outAllowTokenUsers.logs!, isNotEmpty);

        final startAt = DateTime.now().toUtc().subtract(
          const Duration(minutes: 5),
        );
        final expireAt = DateTime.now().toUtc().add(const Duration(hours: 2));

        // ------------------
        // CREATE DROP
        // ------------------
        final outNewDrop = await c.newDrop(
          address: dropAddress,
          programAddress: ownerUser.publicKey,
          tokenAddress: placeholderTokenAddress,
          owner: ownerUser.publicKey,
          title: 'Drop NFT Test',
          description: 'Drop NFT description',
          shortDescription: 'Drop NFT short',
          imageUrl: 'https://example.com/drop-nft.png',
          bannerUrl: 'https://example.com/drop-nft-banner.png',
          categories: const {'nft': true, 'collectible': true},
          socialRequirements: const {'twitter': true},
          postLinks: const {'https://example.com/post-nft': true},
          verificationType: 'ORACLE',
          startAt: startAt,
          expireAt: expireAt,
          requestLimit: 2,
          claimAmount: '1000',
          claimIntervalSeconds: 0,
        );
        expect(outNewDrop.logs, isNotNull);
        expect(outNewDrop.logs!, isNotEmpty);
        expect(outNewDrop.logs!.first.logType, equals('Drop_Created'));

        final stateAfterCreate = await getDropState(c, dropAddress);
        expect(
          stateAfterCreate['token_address'],
          equals(placeholderTokenAddress),
        );
        expect(stateAfterCreate['claim_amount'], equals('1000'));

        // ------------------
        // UPDATE DROP METADATA
        // ------------------
        final outUpdateDrop = await c.updateDropMetadata(
          address: dropAddress,
          programAddress: ownerUser.publicKey,
          tokenAddress: nftTokenAddress,
          title: 'Drop NFT Test',
          description: 'Drop NFT description',
          shortDescription: 'Drop NFT short',
          imageUrl: 'https://example.com/drop-nft.png',
          bannerUrl: 'https://example.com/drop-nft-banner.png',
          categories: const {'nft': true, 'collectible': true},
          socialRequirements: const {'twitter': true},
          postLinks: const {'https://example.com/post-nft': true},
          verificationType: 'ORACLE',
          startAt: startAt,
          expireAt: expireAt,
          requestLimit: 2,
          claimAmount: '1',
          claimIntervalSeconds: 0,
        );
        expect(outUpdateDrop.logs, isNotNull);
        expect(outUpdateDrop.logs!, isNotEmpty);
        expect(
          outUpdateDrop.logs!.first.logType,
          equals('Drop_Metadata_Updated'),
        );

        final stateAfterUpdate = await getDropState(c, dropAddress);
        expect(stateAfterUpdate['token_address'], equals(nftTokenAddress));
        expect(stateAfterUpdate['claim_amount'], equals('1'));

        // ------------------
        // ALLOW ORACLE
        // ------------------
        final outAllowOracles = await c.allowOracles(
          address: dropAddress,
          oracles: {oracleUser.publicKey: true},
        );
        expect(outAllowOracles.logs, isNotNull);
        expect(outAllowOracles.logs!, isNotEmpty);
        expect(
          outAllowOracles.logs!.first.logType,
          equals('Drop_Oracles_Allowed'),
        );

        // ------------------
        // DEPOSIT FIRST NFT
        // ------------------
        final outDepositFirst = await c.depositDrop(
          address: dropAddress,
          programAddress: ownerUser.publicKey,
          tokenAddress: nftTokenAddress,
          amount: '1',
          uuids: [firstUuid],
        );
        expect(outDepositFirst.logs, isNotNull);
        expect(outDepositFirst.logs!, isNotEmpty);
        expect(outDepositFirst.logs!.first.logType, equals('Drop_Deposited'));

        await expectNftBalance(
          c,
          tokenAddress: nftTokenAddress,
          ownerAddress: dropAddress,
          uuid: firstUuid,
          expectedAmount: '1',
          expectedTokenType: TOKEN_TYPE_NON_FUNGIBLE,
        );

        // ------------------
        // DEPOSIT SECOND NFT
        // ------------------
        final outDepositSecond = await c.depositDrop(
          address: dropAddress,
          programAddress: ownerUser.publicKey,
          tokenAddress: nftTokenAddress,
          amount: '1',
          uuids: [secondUuid],
        );
        expect(outDepositSecond.logs, isNotNull);
        expect(outDepositSecond.logs!, isNotEmpty);
        expect(outDepositSecond.logs!.first.logType, equals('Drop_Deposited'));

        await expectNftBalance(
          c,
          tokenAddress: nftTokenAddress,
          ownerAddress: dropAddress,
          uuid: secondUuid,
          expectedAmount: '1',
          expectedTokenType: TOKEN_TYPE_NON_FUNGIBLE,
        );

        // ------------------
        // ATTEST ELIGIBILITY
        // ------------------
        await c.setPrivateKey(oracleUser.privateKey);
        final outAttest = await c.attestParticipantEligibility(
          address: dropAddress,
          wallet: claimerUser.publicKey,
          approved: true,
        );
        expect(outAttest.logs, isNotNull);
        expect(outAttest.logs!, isNotEmpty);
        expect(
          outAttest.logs!.first.logType,
          equals('Drop_Attested_Participant_Eligibility'),
        );

        // ------------------
        // CLAIM NFT
        // ------------------
        await c.setPrivateKey(claimerUser.privateKey);
        final outClaim = await c.claimDrop(address: dropAddress);
        expect(outClaim.logs, isNotNull);
        expect(outClaim.logs!, isNotEmpty);
        expect(outClaim.logs!.first.logType, equals('Drop_Claimed'));

        final claimedUuid = await _resolveClaimedUuid(
          c,
          tokenAddress: nftTokenAddress,
          claimerAddress: claimerUser.publicKey,
          candidateUuids: [firstUuid, secondUuid],
        );
        final remainingUuid = claimedUuid == firstUuid ? secondUuid : firstUuid;

        // ------------------
        // WITHDRAW NFT
        // ------------------
        await c.setPrivateKey(ownerUser.privateKey);
        final outWithdraw = await c.withdrawDrop(
          address: dropAddress,
          programAddress: ownerUser.publicKey,
          tokenAddress: nftTokenAddress,
          amount: '1',
          uuids: [remainingUuid],
        );
        expect(outWithdraw.logs, isNotNull);
        expect(outWithdraw.logs!, isNotEmpty);
        expect(outWithdraw.logs!.first.logType, equals('Drop_Withdrawn'));

        await expectNftBalance(
          c,
          tokenAddress: nftTokenAddress,
          ownerAddress: ownerUser.publicKey,
          uuid: remainingUuid,
          expectedAmount: '1',
          expectedTokenType: TOKEN_TYPE_NON_FUNGIBLE,
        );

        final dropState = await getDropState(c, dropAddress);
        expect(
          Map<String, dynamic>.from(
            (dropState['claimed_wallets'] as Map?) ?? const {},
          )[claimerUser.publicKey],
          isTrue,
        );
      },
      timeout: testTimeout,
    );
  });
}
