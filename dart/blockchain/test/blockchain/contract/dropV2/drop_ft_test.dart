import 'package:test/test.dart';
import 'package:two_finance_blockchain/blockchain/contract/dropV2/constants.dart';
import 'package:two_finance_blockchain/blockchain/contract/tokenV2/constants.dart';
import 'package:two_finance_blockchain/blockchain/types/types.dart';
import 'package:two_finance_blockchain/blockchain/utils/marshal.dart';
import 'package:two_finance_blockchain/two_finance_blockchain.dart';

import 'drop_test_helpers.dart';
import '../../../helpers/helpers.dart';

void main() {
  final testTimeout = Timeout(Duration(minutes: 4));

  group('DropV2 FT E2E', () {
    test(
      'E2E: create + update + allow oracle + attest + deposit + claim + withdraw + pause + list',
      () async {
        final c = await setupClient();
        addTearDown(() => teardownClient(c));

        final ownerUser = await newTestUser('owner');
        final claimerUser = await newTestUser('claimer');
        final oracleUser = await newTestUser('oracle');

        // ------------------
        // SETUP TOKEN
        // ------------------
        final tokenAddress = await createBasicToken(
          c,
          ownerPrivateKey: ownerUser.privateKey,
          ownerPublicKey: ownerUser.publicKey,
          decimals: 0,
          requireFee: false,
          tokenType: TOKEN_TYPE_FUNGIBLE,
          stablecoin: false,
        );

        await c.setPrivateKey(ownerUser.privateKey);
        final deployed = await c.deployContract1(DROP_CONTRACT_V2);
        expect(deployed, isA<ContractOutput>());
        expect(deployed.logs, isNotNull);
        expect(deployed.logs!, isNotEmpty);

        final dropAddress = deployed.logs!.first.contractAddress;
        expect(dropAddress, isNotEmpty);

        // ------------------
        // ALLOW TOKEN USERS
        // ------------------
        final outAllowTokenUsers = await c.allowUsers(tokenAddress, {
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

        const categories = <String, bool>{'airdrop': true, 'campaign': true};
        const socialRequirements = <String, bool>{'twitter': true};
        const postLinks = <String, bool>{'https://example.com/post-ft': true};

        // ------------------
        // CREATE DROP
        // ------------------
        final outNewDrop = await c.newDrop(
          address: dropAddress,
          programAddress: ownerUser.publicKey,
          tokenAddress: tokenAddress,
          owner: ownerUser.publicKey,
          title: 'Drop FT Test',
          description: 'Drop FT description',
          shortDescription: 'Drop FT short',
          imageUrl: 'https://example.com/drop-ft.png',
          bannerUrl: 'https://example.com/drop-ft-banner.png',
          categories: categories,
          socialRequirements: socialRequirements,
          postLinks: postLinks,
          verificationType: 'ORACLE',
          startAt: startAt,
          expireAt: expireAt,
          requestLimit: 2,
          claimAmount: '100',
          claimIntervalSeconds: 60,
        );

        expect(outNewDrop.logs, isNotNull);
        expect(outNewDrop.logs!, isNotEmpty);
        expect(outNewDrop.logs!.first.logType, equals('Drop_Created'));

        final createdEvent = unmarshalEvent<JsonMap>(
          outNewDrop.logs!.first.event,
          (json) => Map<String, dynamic>.from(json as Map),
        );
        expect(createdEvent['address'], equals(dropAddress));
        expect(createdEvent['token_address'], equals(tokenAddress));
        expect(createdEvent['program_address'], equals(ownerUser.publicKey));

        final stateAfterCreate = await getDropState(c, dropAddress);
        expectDropSnapshot(
          stateAfterCreate,
          address: dropAddress,
          owner: ownerUser.publicKey,
          programAddress: ownerUser.publicKey,
          tokenAddress: tokenAddress,
          title: 'Drop FT Test',
          description: 'Drop FT description',
          shortDescription: 'Drop FT short',
          imageUrl: 'https://example.com/drop-ft.png',
          bannerUrl: 'https://example.com/drop-ft-banner.png',
          categories: categories,
          socialRequirements: socialRequirements,
          postLinks: postLinks,
          verificationType: 'ORACLE',
          startAt: startAt,
          expireAt: expireAt,
          requestLimit: 2,
          claimAmount: '100',
          claimIntervalSeconds: 60,
          paused: false,
        );

        // ------------------
        // UPDATE DROP
        // ------------------
        final updatedStartAt = startAt.subtract(const Duration(minutes: 10));
        final updatedExpireAt = expireAt.add(const Duration(hours: 1));
        const updatedCategories = <String, bool>{
          'airdrop': true,
          'campaign': true,
          'vip': true,
        };
        const updatedSocial = <String, bool>{'twitter': true, 'discord': true};
        const updatedPosts = <String, bool>{
          'https://example.com/post-ft-updated': true,
        };

        final outUpdate = await c.updateDropMetadata(
          address: dropAddress,
          programAddress: ownerUser.publicKey,
          tokenAddress: tokenAddress,
          title: 'Drop FT Updated',
          description: 'Drop FT description updated',
          shortDescription: 'Drop FT short updated',
          imageUrl: 'https://example.com/drop-ft-updated.png',
          bannerUrl: 'https://example.com/drop-ft-banner-updated.png',
          categories: updatedCategories,
          socialRequirements: updatedSocial,
          postLinks: updatedPosts,
          verificationType: 'ORACLE',
          startAt: updatedStartAt,
          expireAt: updatedExpireAt,
          requestLimit: 3,
          claimAmount: '100',
          claimIntervalSeconds: 120,
        );

        expect(outUpdate.logs, isNotNull);
        expect(outUpdate.logs!, isNotEmpty);
        expect(outUpdate.logs!.first.logType, equals('Drop_Metadata_Updated'));

        final stateAfterUpdate = await getDropState(c, dropAddress);
        expectDropSnapshot(
          stateAfterUpdate,
          address: dropAddress,
          owner: ownerUser.publicKey,
          programAddress: ownerUser.publicKey,
          tokenAddress: tokenAddress,
          title: 'Drop FT Updated',
          description: 'Drop FT description updated',
          shortDescription: 'Drop FT short updated',
          imageUrl: 'https://example.com/drop-ft-updated.png',
          bannerUrl: 'https://example.com/drop-ft-banner-updated.png',
          categories: updatedCategories,
          socialRequirements: updatedSocial,
          postLinks: updatedPosts,
          verificationType: 'ORACLE',
          startAt: updatedStartAt,
          expireAt: updatedExpireAt,
          requestLimit: 3,
          claimAmount: '100',
          claimIntervalSeconds: 120,
          paused: false,
        );

        final actualPostLinks = Map<String, dynamic>.from(
          (stateAfterUpdate['post_links'] as Map?) ?? const {},
        );

        expect(actualPostLinks['https://example.com/post-ft'], isTrue);
        expect(actualPostLinks['https://example.com/post-ft-updated'], isTrue);

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

        final stateAfterOracleAllow = await getDropState(c, dropAddress);
        expect(
          Map<String, dynamic>.from(
            (stateAfterOracleAllow['allowed_oracles'] as Map?) ?? const {},
          )[oracleUser.publicKey],
          isTrue,
        );

        // ------------------
        // DEPOSIT FT
        // ------------------
        final outDeposit = await c.depositDrop(
          address: dropAddress,
          programAddress: ownerUser.publicKey,
          tokenAddress: tokenAddress,
          amount: '300',
        );
        expect(outDeposit.logs, isNotNull);
        expect(outDeposit.logs!, isNotEmpty);
        expect(outDeposit.logs!.first.logType, equals('Drop_Deposited'));

        await expectFtBalance(
          c,
          tokenAddress: tokenAddress,
          ownerAddress: dropAddress,
          expectedAmount: '300',
        );

        await c.setPrivateKey(claimerUser.privateKey);

        await expectLater(
          c.claimDrop(address: dropAddress),
          throwsA(
            predicate(
              (e) => e.toString().contains('is not eligible for this drop'),
            ),
          ),
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

        final stateAfterAttest = await getDropState(c, dropAddress);
        expect(
          Map<String, dynamic>.from(
            (stateAfterAttest['eligible_wallets'] as Map?) ?? const {},
          )[claimerUser.publicKey],
          isTrue,
        );

        // ------------------
        // CLAIM DROP
        // ------------------
        await c.setPrivateKey(claimerUser.privateKey);
        final outClaim = await c.claimDrop(address: dropAddress);
        expect(outClaim.logs, isNotNull);
        expect(outClaim.logs!, isNotEmpty);
        expect(outClaim.logs!.first.logType, equals('Drop_Claimed'));

        await expectFtBalance(
          c,
          tokenAddress: tokenAddress,
          ownerAddress: claimerUser.publicKey,
          expectedAmount: '100',
        );
        await expectFtBalance(
          c,
          tokenAddress: tokenAddress,
          ownerAddress: dropAddress,
          expectedAmount: '200',
        );

        final stateAfterClaim = await getDropState(c, dropAddress);
        expect(
          Map<String, dynamic>.from(
            (stateAfterClaim['claimed_wallets'] as Map?) ?? const {},
          )[claimerUser.publicKey],
          isTrue,
        );

        // ------------------
        // DISALLOW ORACLE
        // ------------------
        await c.setPrivateKey(ownerUser.privateKey);
        final outDisallowOracles = await c.disallowOracles(
          address: dropAddress,
          oracles: {oracleUser.publicKey: true},
        );
        expect(outDisallowOracles.logs, isNotNull);
        expect(outDisallowOracles.logs!, isNotEmpty);
        expect(
          outDisallowOracles.logs!.first.logType,
          equals('Drop_Oracles_Disallowed'),
        );

        final stateAfterOracleDisallow = await getDropState(c, dropAddress);
        expect(
          Map<String, dynamic>.from(
            (stateAfterOracleDisallow['allowed_oracles'] as Map?) ?? const {},
          )[oracleUser.publicKey],
          isNull,
        );

        // ------------------
        // PAUSE / UNPAUSE
        // ------------------
        await c.setPrivateKey(ownerUser.privateKey);
        final outPause = await c.pauseDrop(dropAddress);
        expect(outPause.logs, isNotNull);
        expect(outPause.logs!, isNotEmpty);
        expect(outPause.logs!.first.logType, equals('Drop_Paused'));

        final stateAfterPause = await getDropState(c, dropAddress);
        expect(stateAfterPause['paused'], isTrue);

        final outUnpause = await c.unpauseDrop(dropAddress);
        expect(outUnpause.logs, isNotNull);
        expect(outUnpause.logs!, isNotEmpty);
        expect(outUnpause.logs!.first.logType, equals('Drop_Unpaused'));

        final stateAfterUnpause = await getDropState(c, dropAddress);
        expect(stateAfterUnpause['paused'], isFalse);

        // ------------------
        // WITHDRAW FT
        // ------------------
        final ownerBalanceBeforeWithdraw = await getFtBalanceAmount(
          c,
          tokenAddress: tokenAddress,
          ownerAddress: ownerUser.publicKey,
        );

        final outWithdraw = await c.withdrawDrop(
          address: dropAddress,
          programAddress: ownerUser.publicKey,
          tokenAddress: tokenAddress,
          amount: '200',
        );
        expect(outWithdraw.logs, isNotNull);
        expect(outWithdraw.logs!, isNotEmpty);
        expect(outWithdraw.logs!.first.logType, equals('Drop_Withdrawn'));

        final ownerBalanceAfterWithdraw = await getFtBalanceAmount(
          c,
          tokenAddress: tokenAddress,
          ownerAddress: ownerUser.publicKey,
        );
        expect(
          ownerBalanceAfterWithdraw,
          equals(
            (BigInt.parse(ownerBalanceBeforeWithdraw) + BigInt.from(200))
                .toString(),
          ),
        );

        await expectFtBalance(
          c,
          tokenAddress: tokenAddress,
          ownerAddress: dropAddress,
          expectedAmount: '0',
        );

        // ------------------
        // LIST DROPS
        // ------------------
        final outList = await c.listDrops(
          owner: ownerUser.publicKey,
          page: 1,
          limit: 10,
          ascending: true,
        );
        expect(outList.states, isNotNull);
        expect(outList.states!, isNotEmpty);

        final items = parseDropListState(outList.states!.first.object);
        expect(items.any((item) => item['address'] == dropAddress), isTrue);
      },
      timeout: testTimeout,
    );

    // --------------
    // TEST GET DROP
    // --------------
    test('GetDrop: success and errors', () async {
      final c = await setupClient();
      addTearDown(() => teardownClient(c));

      final ownerUser = await newTestUser('owner');
      final startAt = DateTime.now().toUtc();
      final expireAt = startAt.add(const Duration(hours: 24));
      const categories = <String, bool>{'airdrop': true};
      const socialRequirements = <String, bool>{'follow_x': true};
      const postLinks = <String, bool>{'https://x.com/post/1': true};

      final tokenAddress = await createBasicToken(
        c,
        ownerPrivateKey: ownerUser.privateKey,
        ownerPublicKey: ownerUser.publicKey,
        decimals: 0,
        requireFee: false,
        tokenType: TOKEN_TYPE_FUNGIBLE,
        stablecoin: false,
      );

      await c.setPrivateKey(ownerUser.privateKey);
      final deployed = await c.deployContract1(DROP_CONTRACT_V2);
      final dropAddress = deployed.logs!.first.contractAddress;

      await c.newDrop(
        address: dropAddress,
        programAddress: ownerUser.publicKey,
        tokenAddress: tokenAddress,
        owner: ownerUser.publicKey,
        title: 'drop get test',
        description: 'desc',
        shortDescription: 'short',
        imageUrl: 'https://img.png',
        bannerUrl: 'https://banner.png',
        categories: const {'airdrop': true},
        socialRequirements: socialRequirements,
        postLinks: postLinks,
        verificationType: 'ORACLE',
        startAt: startAt,
        expireAt: expireAt,
        requestLimit: 100,
        claimAmount: '10',
        claimIntervalSeconds: 3600,
      );

      final out = await c.getDrop(address: dropAddress);
      expect(out.states, isNotNull);
      expect(out.states!, isNotEmpty);
      expect(out.states!, hasLength(1));

      final state = unmarshalState(
        out.states!.first.object,
        (json) => Map<String, dynamic>.from(json as Map),
      );

      expectDropSnapshot(
        state,
        address: dropAddress,
        owner: ownerUser.publicKey,
        programAddress: ownerUser.publicKey,
        tokenAddress: tokenAddress,
        title: 'drop get test',
        description: 'desc',
        shortDescription: 'short',
        imageUrl: 'https://img.png',
        bannerUrl: 'https://banner.png',
        categories: categories,
        socialRequirements: socialRequirements,
        postLinks: postLinks,
        verificationType: 'ORACLE',
        startAt: startAt,
        expireAt: expireAt,
        requestLimit: 100,
        claimAmount: '10',
        claimIntervalSeconds: 3600,
        paused: false,
      );

      expect(
        () => c.getDrop(address: ''),
        throwsA(
          isA<ArgumentError>().having(
            (e) => e.message,
            'message',
            contains('must be set'),
          ),
        ),
      );

      expect(
        () => c.getDrop(address: 'invalid-address'),
        throwsA(
          isA<ArgumentError>().having(
            (e) => e.message,
            'message',
            contains('invalid'),
          ),
        ),
      );
    }, timeout: testTimeout);

    test('LastClaimed', () async {
      final c = await setupClient();
      addTearDown(() => teardownClient(c));

      final ownerUser = await newTestUser('owner');
      final claimerUser = await newTestUser('claimer');
      final oracleUser = await newTestUser('oracle');

      final tokenAddress = await createBasicToken(
        c,
        ownerPrivateKey: ownerUser.privateKey,
        ownerPublicKey: ownerUser.publicKey,
        decimals: 0,
        requireFee: false,
        tokenType: TOKEN_TYPE_FUNGIBLE,
        stablecoin: false,
      );

      await c.setPrivateKey(ownerUser.privateKey);
      final deployed = await c.deployContract1(DROP_CONTRACT_V2);
      final dropAddress = deployed.logs!.first.contractAddress;
      expect(dropAddress, isNotEmpty);

      final startAt = DateTime.now().toUtc().subtract(
        const Duration(minutes: 5),
      );
      final expireAt = DateTime.now().toUtc().add(const Duration(hours: 2));

      await c.newDrop(
        address: dropAddress,
        programAddress: ownerUser.publicKey,
        tokenAddress: tokenAddress,
        owner: ownerUser.publicKey,
        title: 'drop last claimed test',
        description: 'desc',
        shortDescription: 'short',
        imageUrl: 'https://img.png',
        bannerUrl: 'https://banner.png',
        categories: const {'airdrop': true},
        socialRequirements: const {'follow_x': true},
        postLinks: const {'https://x.com/post/1': true},
        verificationType: 'ORACLE',
        startAt: startAt,
        expireAt: expireAt,
        requestLimit: 100,
        claimAmount: '10',
        claimIntervalSeconds: 3600,
      );

      final outAllowOracles = await c.allowOracles(
        address: dropAddress,
        oracles: {oracleUser.publicKey: true},
      );
      expect(outAllowOracles.logs, isNotNull);
      expect(outAllowOracles.logs!, isNotEmpty);

      await c.setPrivateKey(ownerUser.privateKey);

      final outDeposit = await c.depositDrop(
        address: dropAddress,
        programAddress: ownerUser.publicKey,
        tokenAddress: tokenAddress,
        amount: '100',
      );
      expect(outDeposit.logs, isNotNull);
      expect(outDeposit.logs!, isNotEmpty);
      expect(outDeposit.logs!.first.logType, equals('Drop_Deposited'));

      final outAttest = await c.attestParticipantEligibility(
        address: dropAddress,
        wallet: claimerUser.publicKey,
        approved: true,
      );
      expect(outAttest.logs, isNotNull);
      expect(outAttest.logs!, isNotEmpty);

      await c.setPrivateKey(claimerUser.privateKey);
      final outClaim = await c.claimDrop(address: dropAddress);
      expect(outClaim.logs, isNotNull);
      expect(outClaim.logs!, isNotEmpty);
      expect(outClaim.logs!.first.logType, equals('Drop_Claimed'));

      final outLastClaimed = await c.lastClaimed(
        address: dropAddress,
        wallet: claimerUser.publicKey,
      );

      expect(outLastClaimed, isA<ContractOutput>());
    }, timeout: testTimeout);

    // -----------------
    // TEST LIST DROPS
    // -----------------
    test('ListDrops', () async {
      final c = await setupClient();
      addTearDown(() => teardownClient(c));

      final ownerUser = await newTestUser('owner');
      await c.setPrivateKey(ownerUser.privateKey);

      Future<String> createQuickDrop(String title) async {
        final tokenAddress = await createBasicToken(
          c,
          ownerPrivateKey: ownerUser.privateKey,
          ownerPublicKey: ownerUser.publicKey,
          decimals: 0,
          requireFee: false,
          tokenType: TOKEN_TYPE_FUNGIBLE,
          stablecoin: false,
        );

        final deployed = await c.deployContract1(DROP_CONTRACT_V2);
        final dropAddress = deployed.logs!.first.contractAddress;

        await c.newDrop(
          address: dropAddress,
          programAddress: ownerUser.publicKey,
          tokenAddress: tokenAddress,
          owner: ownerUser.publicKey,
          title: title,
          description: 'desc',
          shortDescription: 'short',
          imageUrl: 'https://img.png',
          bannerUrl: 'https://banner.png',
          categories: const {'airdrop': true},
          socialRequirements: const {'follow_x': true},
          postLinks: const {'https://x.com/post/1': true},
          verificationType: 'ORACLE',
          startAt: DateTime.now().toUtc(),
          expireAt: DateTime.now().toUtc().add(const Duration(hours: 24)),
          requestLimit: 100,
          claimAmount: '10',
          claimIntervalSeconds: 3600,
        );

        return dropAddress;
      }

      final address1 = await createQuickDrop('drop list test 1');
      final address2 = await createQuickDrop('drop list test 2');

      final outOwner = await c.listDrops(
        owner: ownerUser.publicKey,
        page: 1,
        limit: 10,
        ascending: true,
      );
      expect(outOwner.states, isNotNull);
      final itemsOwner = parseDropListState(outOwner.states!.first.object);
      expect(itemsOwner.any((item) => item['address'] == address1), isTrue);
      expect(itemsOwner.any((item) => item['address'] == address2), isTrue);

      final outEmptyOwner = await c.listDrops(
        owner: '',
        page: 1,
        limit: 10,
        ascending: true,
      );
      expect(outEmptyOwner.states, isNotNull);
      expect(outEmptyOwner.states, isNotEmpty);

      expect(
        () => c.listDrops(owner: 'invalid-owner', page: 1, limit: 10),
        throwsA(
          isA<ArgumentError>().having(
            (e) => e.message,
            'message',
            contains('invalid'),
          ),
        ),
      );

      expect(
        () => c.listDrops(owner: ownerUser.publicKey, page: 0, limit: 10),
        throwsA(
          isA<ArgumentError>().having(
            (e) => e.message,
            'message',
            contains('page must be greater than 0'),
          ),
        ),
      );

      expect(
        () => c.listDrops(owner: ownerUser.publicKey, page: 1, limit: 0),
        throwsA(
          isA<ArgumentError>().having(
            (e) => e.message,
            'message',
            contains('limit must be greater than 0'),
          ),
        ),
      );
    }, timeout: testTimeout);
  });
}
