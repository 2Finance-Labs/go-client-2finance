import 'dart:convert';

import 'package:two_finance_blockchain/blockchain/contract/contractV2/models/model.dart'
    as models;
import 'package:two_finance_blockchain/blockchain/contract/contractV2/domain/contract.dart'
    as domain;
import 'package:two_finance_blockchain/blockchain/contract/walletV2/constants.dart';
import 'package:two_finance_blockchain/blockchain/contract/walletV2/models/wallet.dart'
    as walletModels;
import 'package:two_finance_blockchain/blockchain/contract/walletV2/domain/wallet.dart'
    as walletDomain;
import 'package:two_finance_blockchain/blockchain/keys/keys.dart';
import 'package:two_finance_blockchain/blockchain/types/types.dart';
import 'package:two_finance_blockchain/blockchain/utils/json.dart';
import 'package:two_finance_blockchain/two_finance_blockchain.dart';
import 'package:two_finance_blockchain/blockchain/utils/marshal.dart';
import 'package:test/test.dart';
import 'package:web3dart/contracts.dart';
import '../../../helpers/helpers.dart';

void main() async {
  group('TwoFinanceBlockchain E2E - wallet', () {
    late TwoFinanceBlockchain c;

    setUpAll(() async {
      c = await setupClient();
    });

    tearDownAll(() async {
      await teardownClient(c);
    });

    test('generateKeyEd25519 + setPrivateKey deriva publicKeyHex', () async {
      final kp = await validKeyPair();

      // chaves geradas
      expect(kp.privateKey, isNotEmpty);
      expect(kp.publicKey, isNotEmpty);

      await c.setPrivateKey(kp.privateKey);

      // publicKeyHex derivada da private key
      expect(c.publicKeyHex, isNotNull);
      expect(c.publicKeyHex!, isNotEmpty);

      // validação oficial do seu pacote
      KeyManager.validateEDDSAPublicKeyHex(c.publicKeyHex!);

      // e deve bater com a pub gerada pelo helper (se seu validKeyPair for consistente)
      expect(c.publicKeyHex!, equals(kp.publicKey));
    });

    test(
      'E2E: createWallet (deploy + addWallet) + getWallet (roundtrip)',
      () async {
        // 1) signer
        final kp = await validKeyPair();
        await c.setPrivateKey(kp.privateKey);

        // 2) deploy wallet contract
        final deployedContract = await c.deployContract1(WALLET_CONTRACT_V2);

        final contractLogs = deployedContract.logs;
        expect(contractLogs, isNotNull);
        expect(contractLogs!, isNotEmpty);

        // pega o primeiro log
        final firstLog = contractLogs.first;

        // agora decodifica o event (base64 -> bytes) e faz unmarshalEvent
        final contractEvent = unmarshalEvent<domain.Contract>(
          firstLog.event,
          domain.Contract.fromJson,
        );

        expect(contractEvent.address, isNotEmpty);
        expect(contractEvent.contractVersion, equals(WALLET_CONTRACT_V2));
        print('Deployed contract at address: ${contractEvent.address}');
        print('Deployed public key: ${kp.publicKey}');

        final addWalletOutput = await c.addWallet(
          contractEvent.address,
          kp.publicKey,
        );
        expect(addWalletOutput, isNotNull);
        expect(addWalletOutput.logs, isNotNull);
        expect(addWalletOutput.logs!, isNotEmpty);

        final addWalletLog = addWalletOutput.logs!.first;
        expect(addWalletLog.logType, 'Wallet_Created');
        expect(addWalletLog.logIndex, 1);
        expect(addWalletLog.transactionHash, isNotEmpty);
        expect(addWalletLog.contractVersion, WALLET_CONTRACT_V2);
        expect(addWalletLog.contractAddress, contractEvent.address);
        expect(addWalletLog.event, isNotEmpty);

        final createdWallet = unmarshalEvent<walletDomain.Wallet>(
          addWalletLog.event,
          walletDomain.Wallet.fromJson,
        );

        expect(createdWallet.publicKey, equals(kp.publicKey));
        expect(createdWallet.address, equals(contractEvent.address));

        final getWalletOutput = await c.getWalletByAddress(
          contractEvent.address,
        );
        expect(getWalletOutput, isNotNull);
        expect(getWalletOutput.states, isNotNull);
        expect(getWalletOutput.states!, isNotEmpty);

        final getWalletStates = getWalletOutput.states;
        expect(getWalletStates, isNotNull);
        expect(getWalletStates!, isNotEmpty);

        final StateType state = getWalletStates!.first;

        final walletState = unmarshalState<walletModels.WalletState>(
          state.object,
          walletModels.WalletState.fromJson,
        );

        expect(walletState.publicKey, equals(kp.publicKey));
        expect(walletState.address, equals(contractEvent.address));
        expect(walletState.createdAt, isNotNull);
        expect(walletState.createdAt, isA<DateTime>());
        expect(walletState.updatedAt, isNotNull);
        expect(walletState.updatedAt, isA<DateTime>());
        expect(walletState.createdAt.isAfter(DateTime.utc(2000, 1, 1)), isTrue);
        expect(walletState.updatedAt.isAfter(DateTime.utc(2000, 1, 1)), isTrue);
        expect(
          walletState.updatedAt.isAfter(walletState.createdAt) ||
              walletState.updatedAt.isAtSameMomentAs(walletState.createdAt),
          isTrue,
        );
      },
    );
  });
}
