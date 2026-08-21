import 'dart:convert';

import 'package:crypto/crypto.dart';
import 'package:test/test.dart';
import 'package:two_finance_blockchain/blockchain/types/types.dart';
import 'package:two_finance_blockchain/blockchain/contract/raffleV2/constants.dart';
import 'package:two_finance_blockchain/blockchain/utils/marshal.dart';
import 'package:two_finance_blockchain/two_finance_blockchain.dart';
import 'package:two_finance_blockchain/blockchain/contract/tokenV2/constants.dart';
import '../../../helpers/helpers.dart';

typedef JsonMap = Map<String, dynamic>;

Future<JsonMap> _getRaffleState(TwoFinanceBlockchain c, String address) async {
  final out = await c.getRaffle(address);
  expect(out.states, isNotNull);
  expect(out.states!, isNotEmpty);

  return unmarshalState(
    out.states!.first.object,
    (json) => Map<String, dynamic>.from(json as Map),
  );
}

void _expectDateClose(
  String actualIso,
  DateTime expected, {
  int toleranceSeconds = 2,
}) {
  final actual = DateTime.parse(actualIso).toUtc();
  final expectedUtc = expected.toUtc();

  expect(
    actual.difference(expectedUtc).inSeconds.abs(),
    lessThanOrEqualTo(toleranceSeconds),
  );
}

void _expectRaffleSnapshot(
  JsonMap raffle, {
  required String address,
  required String owner,
  required String tokenAddress,
  required String ticketPrice,
  required int maxEntries,
  required int maxEntriesPerUser,
  required bool paused,
  required String seedCommitHex,
  required Map<String, String> metadata,
  required DateTime startAt,
  required DateTime expiredAt,
}) {
  expect(raffle['address'], equals(address));
  expect(raffle['owner'], equals(owner));
  expect(raffle['token_address'], equals(tokenAddress));
  expect(raffle['ticket_price'], equals(ticketPrice));
  expect(raffle['max_entries'], equals(maxEntries));
  expect(raffle['max_entries_per_user'], equals(maxEntriesPerUser));
  expect(raffle['paused'], equals(paused));
  expect(raffle['seed_commit_hex'], equals(seedCommitHex));

  expect(raffle['reveal_seed'], anyOf(isNull, equals('')));

  expect(raffle['hash'], isA<String>());
  expect((raffle['hash'] as String).isNotEmpty, isTrue);

  final raffleMetadata = Map<String, dynamic>.from(
    (raffle['metadata'] as Map?) ?? const {},
  );

  metadata.forEach((key, value) {
    expect(raffleMetadata[key], equals(value));
  });

  _expectDateClose(raffle['start_at'] as String, startAt);
  _expectDateClose(raffle['expired_at'] as String, expiredAt);
}

void main() {
  group('RaffleV2 E2E', () {
    test('E2E: deploy raffle contract + addRaffle + getRaffle', () async {
      final c = await setupClient();

      final ownerUser = await newTestUser('owner');
      final player1User = await newTestUser('player1');
      final player2User = await newTestUser('player2');

      final paymentTokenAddress = await createBasicToken(
        c,
        ownerPrivateKey: ownerUser.privateKey,
        ownerPublicKey: ownerUser.publicKey,
        decimals: 6,
        requireFee: false,
        tokenType: TOKEN_TYPE_FUNGIBLE,
        stablecoin: false,
      );

      final prizeTokenAddress = await createBasicToken(
        c,
        ownerPrivateKey: ownerUser.privateKey,
        ownerPublicKey: ownerUser.publicKey,
        decimals: 6,
        requireFee: false,
        tokenType: TOKEN_TYPE_FUNGIBLE,
        stablecoin: false,
      );

      await c.setPrivateKey(ownerUser.privateKey);

      final deployed = await c.deployContract1(RAFFLE_CONTRACT_V2);

      expect(deployed, isA<ContractOutput>());
      expect(deployed.logs, isNotNull);
      expect(deployed.logs!, isNotEmpty);

      final raffleAddress = deployed.logs!.first.contractAddress;
      expect(raffleAddress, isNotEmpty);

      final outAllowPayToken = await c.allowUsers(paymentTokenAddress, {
        ownerUser.publicKey: true,
        player1User.publicKey: true,
        player2User.publicKey: true,
        raffleAddress: true,
      });

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

      final outAllowPrizeToken = await c.allowUsers(prizeTokenAddress, {
        ownerUser.publicKey: true,
        player1User.publicKey: true,
        player2User.publicKey: true,
        raffleAddress: true,
      });

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

      final outTransferPlayer1 = await c.transferToken(
        tokenAddress: paymentTokenAddress,
        transferTo: player1User.publicKey,
        amount: '3000000',
        decimals: 0,
        tokenType: TOKEN_TYPE_FUNGIBLE,
        uuid: '',
      );

      expect(outTransferPlayer1, isA<ContractOutput>());

      final outTransferPlayer2 = await c.transferToken(
        tokenAddress: paymentTokenAddress,
        transferTo: player2User.publicKey,
        amount: '3000000',
        decimals: 0,
        tokenType: TOKEN_TYPE_FUNGIBLE,
        uuid: '',
      );

      expect(outTransferPlayer2, isA<ContractOutput>());

      final startAt = DateTime.now().toUtc().subtract(
        const Duration(minutes: 5),
      );
      final expiredAt = DateTime.now().toUtc().add(const Duration(hours: 2));

      const ticketPrice = '1000000';
      const maxEntries = 10;
      const maxEntriesPerUser = 3;
      const paused = false;

      const revealSeed = 'raffle-secret-seed-fungible-e2e';
      final seedCommitHex = sha256.convert(utf8.encode(revealSeed)).toString();

      final metadata = <String, String>{
        'name': 'Raffle Fungible E2E',
        'description': 'raffle flow fungible',
        'image': 'https://example.com/raffle.png',
      };

      final outAdd = await c.addRaffle(
        address: raffleAddress,
        owner: ownerUser.publicKey,
        tokenAddress: paymentTokenAddress,
        ticketPrice: ticketPrice,
        maxEntries: maxEntries,
        maxEntriesPerUser: maxEntriesPerUser,
        startAt: startAt,
        expiredAt: expiredAt,
        paused: paused,
        seedCommitHex: seedCommitHex,
        metadata: metadata,
      );

      expect(outAdd, isA<ContractOutput>());
      expect(outAdd.logs, isNotNull);
      expect(outAdd.logs!, isNotEmpty);

      final addLog = outAdd.logs!.first;
      expect(addLog.contractAddress, equals(raffleAddress));
      expect(addLog.logType, equals('Raffle_Added'));

      final addEvent = unmarshalEvent<JsonMap>(
        addLog.event,
        (json) => Map<String, dynamic>.from(json as Map),
      );

      expect(addEvent['address'], equals(raffleAddress));
      expect(addEvent['owner'], equals(ownerUser.publicKey));
      expect(addEvent['token_address'], equals(paymentTokenAddress));
      expect(addEvent['ticket_price'], equals(ticketPrice));
      expect(addEvent['max_entries'], equals(maxEntries));
      expect(addEvent['max_entries_per_user'], equals(maxEntriesPerUser));

      final addMetadata = Map<String, dynamic>.from(
        addEvent['metadata'] as Map,
      );
      expect(addMetadata['name'], equals('Raffle Fungible E2E'));
      expect(addMetadata['description'], equals('raffle flow fungible'));
      expect(addMetadata['image'], equals('https://example.com/raffle.png'));

      expect(addEvent['hash'], isNotNull);
      expect(addEvent['hash'], isA<String>());

      _expectDateClose(addEvent['start_at'] as String, startAt);
      _expectDateClose(addEvent['expired_at'] as String, expiredAt);

      expect(addEvent['seed_commit_hex'], equals(seedCommitHex));

      final raffleState = await _getRaffleState(c, raffleAddress);
      _expectRaffleSnapshot(
        raffleState,
        address: raffleAddress,
        owner: ownerUser.publicKey,
        tokenAddress: paymentTokenAddress,
        ticketPrice: ticketPrice,
        maxEntries: maxEntries,
        maxEntriesPerUser: maxEntriesPerUser,
        paused: paused,
        seedCommitHex: seedCommitHex,
        metadata: metadata,
        startAt: startAt,
        expiredAt: expiredAt,
      );

      // ------------------
      // ADD PRIZE
      // ------------------
      const prizeAmount = '2500000';
      const requestedPrizeUuid = 'prize-ft-001';

      final ownerPrizeBeforeAdd = await getFtBalanceAmount(
        c,
        tokenAddress: prizeTokenAddress,
        ownerAddress: ownerUser.publicKey,
      );

      await c.setPrivateKey(ownerUser.privateKey);

      final outAddPrize = await c.addRafflePrize(
        raffleAddress: raffleAddress,
        tokenAddress: prizeTokenAddress,
        amount: prizeAmount,
        tokenType: TOKEN_TYPE_FUNGIBLE,
        uuid: requestedPrizeUuid,
      );

      expect(outAddPrize, isA<ContractOutput>());
      expect(outAddPrize.logs, isNotNull);
      expect(outAddPrize.logs!, isNotEmpty);

      final addPrizeLog = outAddPrize.logs!.first;
      expect(addPrizeLog.contractAddress, equals(raffleAddress));
      expect(addPrizeLog.logType, equals('Raffle_Prizes_Added'));

      final addPrizeEvent = unmarshalEvent<JsonMap>(
        addPrizeLog.event,
        (json) => Map<String, dynamic>.from(json as Map),
      );

      expect(addPrizeEvent['raffle_address'], equals(raffleAddress));
      expect(addPrizeEvent['sponsor'], equals(ownerUser.publicKey));
      expect(addPrizeEvent['token_address'], equals(prizeTokenAddress));
      expect(addPrizeEvent['amount'], equals(prizeAmount));
      final prizeUuid = addPrizeEvent['uuid'] as String;
      expect(prizeUuid.isNotEmpty, isTrue);

      final rafflePrizeAfterAdd = await getFtBalanceAmount(
        c,
        tokenAddress: prizeTokenAddress,
        ownerAddress: raffleAddress,
      );

      expect(rafflePrizeAfterAdd, equals(prizeAmount));

      final ownerPrizeAfterAdd = await getFtBalanceAmount(
        c,
        tokenAddress: prizeTokenAddress,
        ownerAddress: ownerUser.publicKey,
      );

      expect(
        BigInt.parse(ownerPrizeAfterAdd),
        equals(BigInt.parse(ownerPrizeBeforeAdd) - BigInt.parse(prizeAmount)),
      );

      // ------------------
      // UPDATE RAFFLE
      // ------------------
      final newStart = DateTime.now().toUtc().subtract(
        const Duration(minutes: 10),
      );
      final newExpiredAt = DateTime.now().toUtc().add(const Duration(hours: 3));
      const updatedTicketPrice = '500000';
      const updatedMaxEntries = 20;
      const updatedMaxEntriesPerUser = 5;
      final updatedSeedCommitHex = seedCommitHex;
      final updatedMetadata = <String, String>{
        'name': 'Raffle Fungible E2E Updated',
        'description': 'raffle flow fungible updated',
        'image': 'https://example.com/raffle-updated.png',
      };

      final outUpdate = await c.updateRaffle(
        address: raffleAddress,
        tokenAddress: paymentTokenAddress,
        ticketPrice: updatedTicketPrice,
        maxEntries: updatedMaxEntries,
        maxEntriesPerUser: updatedMaxEntriesPerUser,
        startAt: newStart,
        expiredAt: newExpiredAt,
        seedCommitHex: updatedSeedCommitHex,
        metadata: updatedMetadata,
      );

      expect(outUpdate, isA<ContractOutput>());
      expect(outUpdate.logs, isNotNull);
      expect(outUpdate.logs!, isNotEmpty);

      final updateLog = outUpdate.logs!.first;
      expect(updateLog.contractAddress, equals(raffleAddress));
      expect(updateLog.logType, equals('Raffle_Updated'));

      final updateEvent = unmarshalEvent<JsonMap>(
        updateLog.event,
        (json) => Map<String, dynamic>.from(json as Map),
      );

      expect(updateEvent['address'], equals(raffleAddress));
      expect(updateEvent['token_address'], equals(paymentTokenAddress));
      expect(updateEvent['ticket_price'], equals(updatedTicketPrice));
      expect(updateEvent['max_entries'], equals(updatedMaxEntries));
      expect(
        updateEvent['max_entries_per_user'],
        equals(updatedMaxEntriesPerUser),
      );
      expect(updateEvent['seed_commit_hex'], equals(updatedSeedCommitHex));

      final updateMetadata = Map<String, dynamic>.from(
        updateEvent['metadata'] as Map,
      );
      expect(updateMetadata['name'], equals('Raffle Fungible E2E Updated'));
      expect(
        updateMetadata['description'],
        equals('raffle flow fungible updated'),
      );
      expect(
        updateMetadata['image'],
        equals('https://example.com/raffle-updated.png'),
      );

      expect(updateEvent['hash'], isNotNull);
      expect(updateEvent['hash'], isA<String>());

      _expectDateClose(updateEvent['start_at'] as String, newStart);
      _expectDateClose(updateEvent['expired_at'] as String, newExpiredAt);

      final raffleStateAfterUpdate = await _getRaffleState(c, raffleAddress);

      _expectRaffleSnapshot(
        raffleStateAfterUpdate,
        address: raffleAddress,
        owner: ownerUser.publicKey,
        tokenAddress: paymentTokenAddress,
        ticketPrice: updatedTicketPrice,
        maxEntries: updatedMaxEntries,
        maxEntriesPerUser: updatedMaxEntriesPerUser,
        paused: paused,
        seedCommitHex: updatedSeedCommitHex,
        metadata: updatedMetadata,
        startAt: newStart,
        expiredAt: newExpiredAt,
      );

      // ------------------
      // PAUSE RAFFLE
      // ------------------
      final outPause = await c.pauseRaffle(raffleAddress, true);

      expect(outPause, isA<ContractOutput>());
      expect(outPause.logs, isNotNull);
      expect(outPause.logs!, isNotEmpty);

      final pauseLog = outPause.logs!.first;
      expect(pauseLog.contractAddress, equals(raffleAddress));
      expect(pauseLog.logType, equals('Raffle_Paused'));

      final pauseEvent = unmarshalEvent<JsonMap>(
        pauseLog.event,
        (json) => Map<String, dynamic>.from(json as Map),
      );

      expect(pauseEvent['address'], equals(raffleAddress));
      expect(pauseEvent['paused'], isTrue);

      final raffleStateAfterPause = await _getRaffleState(c, raffleAddress);

      _expectRaffleSnapshot(
        raffleStateAfterPause,
        address: raffleAddress,
        owner: ownerUser.publicKey,
        tokenAddress: paymentTokenAddress,
        ticketPrice: updatedTicketPrice,
        maxEntries: updatedMaxEntries,
        maxEntriesPerUser: updatedMaxEntriesPerUser,
        paused: true,
        seedCommitHex: updatedSeedCommitHex,
        metadata: updatedMetadata,
        startAt: newStart,
        expiredAt: newExpiredAt,
      );

      // ------------------
      // UNPAUSE RAFFLE
      // ------------------
      final outUnpause = await c.unpauseRaffle(raffleAddress, false);

      expect(outUnpause, isA<ContractOutput>());
      expect(outUnpause.logs, isNotNull);
      expect(outUnpause.logs!, isNotEmpty);

      final unpauseLog = outUnpause.logs!.first;
      expect(unpauseLog.contractAddress, equals(raffleAddress));
      expect(unpauseLog.logType, equals('Raffle_Unpaused'));

      final unpauseEvent = unmarshalEvent<JsonMap>(
        unpauseLog.event,
        (json) => Map<String, dynamic>.from(json as Map),
      );

      expect(unpauseEvent['address'], equals(raffleAddress));

      final raffleStateAfterUnpause = await _getRaffleState(c, raffleAddress);

      _expectRaffleSnapshot(
        raffleStateAfterUnpause,
        address: raffleAddress,
        owner: ownerUser.publicKey,
        tokenAddress: paymentTokenAddress,
        ticketPrice: updatedTicketPrice,
        maxEntries: updatedMaxEntries,
        maxEntriesPerUser: updatedMaxEntriesPerUser,
        paused: false,
        seedCommitHex: updatedSeedCommitHex,
        metadata: updatedMetadata,
        startAt: newStart,
        expiredAt: newExpiredAt,
      );

      // ------------------
      // ENTER RAFFLE FT
      // ------------------

      final ticketPriceInt = int.parse(updatedTicketPrice);

      final player1PayBeforeEnter = await getFtBalanceAmount(
        c,
        tokenAddress: paymentTokenAddress,
        ownerAddress: player1User.publicKey,
      );

      final player2PayBeforeEnter = await getFtBalanceAmount(
        c,
        tokenAddress: paymentTokenAddress,
        ownerAddress: player2User.publicKey,
      );

      await enterRaffleFtAndExpect(
        c,
        user: player1User,
        raffleAddress: raffleAddress,
        payTokenAddress: paymentTokenAddress,
        tickets: 2,
        expectedPaid: (2 * ticketPriceInt).toString(),
        requestUuid: 'enter-ft-001',
      );

      await enterRaffleFtAndExpect(
        c,
        user: player2User,
        raffleAddress: raffleAddress,
        payTokenAddress: paymentTokenAddress,
        tickets: 1,
        expectedPaid: (1 * ticketPriceInt).toString(),
        requestUuid: 'enter-ft-002',
      );

      final player1PayAfterEnter = await getFtBalanceAmount(
        c,
        tokenAddress: paymentTokenAddress,
        ownerAddress: player1User.publicKey,
      );

      final player2PayAfterEnter = await getFtBalanceAmount(
        c,
        tokenAddress: paymentTokenAddress,
        ownerAddress: player2User.publicKey,
      );

      final rafflePayAfterEnter = await getFtBalanceAmount(
        c,
        tokenAddress: paymentTokenAddress,
        ownerAddress: raffleAddress,
      );

      expect(
        BigInt.parse(player1PayAfterEnter),
        equals(
          BigInt.parse(player1PayBeforeEnter) - BigInt.from(2 * ticketPriceInt),
        ),
      );

      expect(
        BigInt.parse(player2PayAfterEnter),
        equals(
          BigInt.parse(player2PayBeforeEnter) - BigInt.from(1 * ticketPriceInt),
        ),
      );

      expect(
        BigInt.parse(rafflePayAfterEnter),
        equals(BigInt.from(3 * ticketPriceInt)),
      );

      final raffleStateAfterEntries = await _getRaffleState(c, raffleAddress);

      _expectRaffleSnapshot(
        raffleStateAfterEntries,
        address: raffleAddress,
        owner: ownerUser.publicKey,
        tokenAddress: paymentTokenAddress,
        ticketPrice: updatedTicketPrice,
        maxEntries: updatedMaxEntries,
        maxEntriesPerUser: updatedMaxEntriesPerUser,
        paused: false,
        seedCommitHex: updatedSeedCommitHex,
        metadata: updatedMetadata,
        startAt: newStart,
        expiredAt: newExpiredAt,
      );

      // ------------------
      // DRAW
      // ------------------
      await c.setPrivateKey(ownerUser.privateKey);

      final outDraw = await c.drawRaffle(raffleAddress, revealSeed);

      expect(outDraw, isA<ContractOutput>());
      expect(outDraw.logs, isNotNull);
      expect(outDraw.logs!, isNotEmpty);

      final drawLog = outDraw.logs!.first;
      expect(drawLog.contractAddress, equals(raffleAddress));
      expect(drawLog.logType, equals('Raffle_Drawn'));

      final drawEvent = unmarshalEvent<JsonMap>(
        drawLog.event,
        (json) => Map<String, dynamic>.from(json as Map),
      );

      expect(drawEvent['address'], equals(raffleAddress));
      expect(drawEvent['reveal_seed'], equals(revealSeed));
      expect(drawEvent['seed_commit_hex'], equals(updatedSeedCommitHex));
      expect(drawEvent['winner_count'], equals(1));

      final winners = List<dynamic>.from(drawEvent['winners'] as List).map((
        item,
      ) {
        if (item is String) return item;

        final map = Map<String, dynamic>.from(item as Map);
        final value = map['winner'] ?? map['address'] ?? map['public_key'];

        if (value == null) {
          fail('Formato inesperado em winners: $map');
        }

        return value.toString();
      }).toList();

      expect(winners, hasLength(1));

      final winner = winners.first;
      expect(
        winner,
        anyOf(equals(player1User.publicKey), equals(player2User.publicKey)),
      );

      final raffleStateAfterDraw = await _getRaffleState(c, raffleAddress);

      expect(raffleStateAfterDraw['reveal_seed'], anyOf(isNull, equals('')));
      expect(
        raffleStateAfterDraw['seed_commit_hex'],
        equals(updatedSeedCommitHex),
      );

      final drawMetadata = Map<String, dynamic>.from(
        (raffleStateAfterDraw['metadata'] as Map?) ?? const {},
      );
      expect(
        drawMetadata['description'],
        equals(updatedMetadata['description']),
      );

      // ------------------
      // CLAIM
      // ------------------
      final outWinnerPrizeBefore = await c.getTokenBalance(
        tokenAddress: prizeTokenAddress,
        ownerAddress: winner,
      );

      final winnerPrizeBefore =
          (outWinnerPrizeBefore.states == null ||
              outWinnerPrizeBefore.states!.isEmpty)
          ? '0'
          : unmarshalState(
                  outWinnerPrizeBefore.states!.first.object,
                  (json) => Map<String, dynamic>.from(json as Map),
                )['amount']
                as String;

      final rafflePrizeBeforeClaim = await getFtBalanceAmount(
        c,
        tokenAddress: prizeTokenAddress,
        ownerAddress: raffleAddress,
      );

      expect(rafflePrizeBeforeClaim, equals(prizeAmount));

      if (winner == player1User.publicKey) {
        await c.setPrivateKey(player1User.privateKey);
      } else if (winner == player2User.publicKey) {
        await c.setPrivateKey(player2User.privateKey);
      } else {
        fail('unexpected winner: $winner');
      }

      final outClaim = await c.claimRaffle(
        address: raffleAddress,
        winner: winner,
        tokenType: TOKEN_TYPE_FUNGIBLE,
        uuid: prizeUuid,
      );

      expect(outClaim, isA<ContractOutput>());
      expect(outClaim.logs, isNotNull);
      expect(outClaim.logs!, isNotEmpty);

      final claimLog = outClaim.logs!.first;
      expect(claimLog.contractAddress, equals(raffleAddress));
      expect(claimLog.logType, equals('Raffle_Claimed'));

      final claimEvent = unmarshalEvent<JsonMap>(
        claimLog.event,
        (json) => Map<String, dynamic>.from(json as Map),
      );

      expect(claimEvent['address'], equals(raffleAddress));
      expect(claimEvent['winner'], equals(winner));

      final winnerPrizeAfter = await getFtBalanceAmount(
        c,
        tokenAddress: prizeTokenAddress,
        ownerAddress: winner,
      );

      expect(
        BigInt.parse(winnerPrizeAfter),
        equals(BigInt.parse(winnerPrizeBefore) + BigInt.parse(prizeAmount)),
      );

      // ------------------
      // WITHDRAW
      // ------------------
      const withdrawAmount = '500000';
      const withdrawUuid = 'withdraw-ft-001';

      await c.setPrivateKey(ownerUser.privateKey);

      final ownerPayBeforeWithdraw = await getFtBalanceAmount(
        c,
        tokenAddress: paymentTokenAddress,
        ownerAddress: ownerUser.publicKey,
      );

      final rafflePayBeforeWithdraw = await getFtBalanceAmount(
        c,
        tokenAddress: paymentTokenAddress,
        ownerAddress: raffleAddress,
      );

      expect(rafflePayBeforeWithdraw, equals('1500000'));

      final outWithdraw = await c.withdrawRaffle(
        address: raffleAddress,
        tokenAddress: paymentTokenAddress,
        amount: withdrawAmount,
        tokenType: TOKEN_TYPE_FUNGIBLE,
        uuid: withdrawUuid,
      );

      expect(outWithdraw, isA<ContractOutput>());
      expect(outWithdraw.logs, isNotNull);
      expect(outWithdraw.logs!, isNotEmpty);

      final withdrawLog = outWithdraw.logs!.first;
      expect(withdrawLog.contractAddress, equals(raffleAddress));
      expect(withdrawLog.logType, equals('Raffle_Withdrawn'));

      final withdrawEvent = unmarshalEvent<JsonMap>(
        withdrawLog.event,
        (json) => Map<String, dynamic>.from(json as Map),
      );

      expect(withdrawEvent['address'], equals(raffleAddress));
      expect(withdrawEvent['token_address'], equals(paymentTokenAddress));
      expect(withdrawEvent['amount'], equals(withdrawAmount));

      final ownerPayAfterWithdraw = await getFtBalanceAmount(
        c,
        tokenAddress: paymentTokenAddress,
        ownerAddress: ownerUser.publicKey,
      );

      final rafflePayAfterWithdraw = await getFtBalanceAmount(
        c,
        tokenAddress: paymentTokenAddress,
        ownerAddress: raffleAddress,
      );

      expect(
        BigInt.parse(ownerPayAfterWithdraw),
        equals(
          BigInt.parse(ownerPayBeforeWithdraw) + BigInt.parse(withdrawAmount),
        ),
      );

      expect(
        BigInt.parse(rafflePayAfterWithdraw),
        equals(
          BigInt.parse(rafflePayBeforeWithdraw) - BigInt.parse(withdrawAmount),
        ),
      );

      expect(rafflePayAfterWithdraw, equals('1000000'));

      // ------------------
      // LIST PRIZES
      // ------------------
      await c.setPrivateKey(ownerUser.privateKey);

      final outListPrizes = await c.listPrizes(
        raffleAddress: raffleAddress,
        page: 1,
        limit: 10,
        ascending: true,
      );

      expect(outListPrizes, isA<ContractOutput>());
      expect(outListPrizes.states, isNotNull);
      expect(outListPrizes.states!, isNotEmpty);

      final rawPrizes = outListPrizes.states!.first.object;
      expect(rawPrizes, isA<List>());

      final prizes = List<Map<String, dynamic>>.from(
        (rawPrizes as List).map((e) => Map<String, dynamic>.from(e as Map)),
      );

      expect(prizes, isNotEmpty);

      Map<String, dynamic>? foundPrize;

      for (final p in prizes) {
        if (p['raffle_address'] == raffleAddress &&
            p['token_address'] == prizeTokenAddress &&
            p['amount'] == prizeAmount) {
          foundPrize = p;
          break;
        }
      }

      expect(foundPrize, isNotNull);

      expect(foundPrize!['uuid'], equals(prizeUuid));
      expect(foundPrize['sponsor'], equals(ownerUser.publicKey));
      expect(foundPrize['winner'], equals(winner));
      expect(foundPrize['claimed'], isTrue);
      expect(foundPrize['created_at'], isNotNull);
      expect(foundPrize['updated_at'] ?? foundPrize['deleted_at'], isNotNull);
    });
  });
}
