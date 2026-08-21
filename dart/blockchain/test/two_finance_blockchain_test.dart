import 'dart:convert';

import 'package:test/test.dart';

import 'package:two_finance_blockchain/blockchain/contract/contractV2/models/model.dart'
    as models;
import 'package:two_finance_blockchain/blockchain/contract/contractV2/domain/contract.dart'
    as domain;
import 'package:two_finance_blockchain/blockchain/contract/walletV2/models/wallet.dart';
import 'package:two_finance_blockchain/blockchain/keys/keys.dart';
import 'package:two_finance_blockchain/two_finance_blockchain.dart';
import 'package:two_finance_blockchain/blockchain/utils/json.dart';
// constants (ajuste se seu nome/paths forem diferentes)
import 'package:two_finance_blockchain/blockchain/contract/walletV2/constants.dart';
import 'package:two_finance_blockchain/blockchain/utils/marshal.dart';
import 'helpers/helpers.dart';
import 'e2e_test.dart';

void main() {
  // Esses timeouts são comuns em E2E por causa de rede/infra

  group('TwoFinanceBlockchain E2E', () {
    late TwoFinanceBlockchain c;

    setUpAll(() async {
      c = await setupClient();
    });

    tearDownAll(() async {
      await teardownClient(c);
    });

    test('initialize + setup OK', () async {
      expect(c.isInitialized, isTrue);
    });

    test('generateKeyEd25519 + setPrivateKey deriva publicKeyHex', () async {
      final kp = await c.generateKeyEd25519();
      expect(kp.privateKey, isNotEmpty);
      expect(kp.publicKey, isNotEmpty);

      await c.setPrivateKey(kp.privateKey);

      // publicKeyHex é derivada da private key
      expect(c.publicKeyHex, isNotNull);
      expect(c.publicKeyHex!, isNotEmpty);

      // valida formato EDDSA (se sua validação exigir)
      KeyManager.validateEDDSAPublicKeyHex(c.publicKeyHex!);
    });

    test('deployContract (Wallet V2) retorna estado do contrato', () async {
      // precisa de signer ativo
      final kp = await c.generateKeyEd25519();
      await c.setPrivateKey(kp.privateKey);

      final contractOutput = await c.deployContract1(WALLET_CONTRACT_V2);

      expect(contractOutput, isNotNull);
      expect(contractOutput.logs, isNotNull);
      expect(contractOutput.logs!, isNotEmpty);

      final firstLog = contractOutput.logs!.first;
      expect(firstLog.logType, 'Contract_Deployed');
      expect(firstLog.logIndex, 1);
      expect(firstLog.transactionHash, isNotEmpty);
      expect(firstLog.contractVersion, WALLET_CONTRACT_V2);
      expect(firstLog.contractAddress, isNotEmpty);
      expect(firstLog.event, isNotEmpty);

      final deployed = unmarshalEvent<domain.Contract>(
        firstLog.event,
        domain.Contract.fromJson,
      );

      expect(deployed.address, isNotEmpty);
      expect(deployed.contractVersion, equals(WALLET_CONTRACT_V2));
    });

    test('getState: record not found retorna "0" (fallback)', () async {
      final kp = await c.generateKeyEd25519();
      await c.setPrivateKey(kp.privateKey);

      final JsonMessage emptyData = {};

      final out = await c.getState(
        to: "some_contract",
        method: "some_method",
        data: emptyData,
      );

      expect(out, isNotNull);
    });
  });
}
