import 'package:test/test.dart';
import 'package:two_finance_blockchain/blockchain/contract/constants.dart';
import 'package:two_finance_blockchain/blockchain/contract/walletV2/constants.dart';
import 'package:two_finance_blockchain/two_finance_blockchain.dart';

import '../../../helpers/fake_mqtt.dart';
import '../../../helpers/helpers.dart';

void main() {
  group('Wallet client payloads', () {
    test(
      '[spec:wallet.add] addWallet signs to the wallet address with Go payload fields',
      () async {
        final signer = await validKeyPair();
        final walletAddress = await validPublicKeyHex();
        final walletPublicKey = await validPublicKeyHex();
        final mqtt = FakeMqttClient();
        final client = TwoFinanceBlockchain(
          keyManager: KeyManager(),
          mqttClient: mqtt,
          chainID: 1,
        );
        await client.setPrivateKey(signer.privateKey);

        await client.addWallet(walletAddress, walletPublicKey);

        final request = mqtt.lastRequest;
        final params = Map<String, dynamic>.from(request['params'] as Map);
        expect(request['method'], REQUEST_METHOD_SEND_TRANSACTION);
        expect(params['chain_id'], 1);
        expect(params['from'], signer.publicKey);
        expect(params['to'], walletAddress);
        expect(params['method'], METHOD_ADD_WALLET);
        expect(params['data'], {
          'address': walletAddress,
          'public_key': walletPublicKey,
        });
        expect(params['version'], 1);
        expect(params['hash'], isA<String>());
        expect(params['signature'], isA<String>());
      },
    );

    test(
      '[spec:wallet.get_by_address] getWalletByAddress queries state by address',
      () async {
        final walletAddress = await validPublicKeyHex();
        final mqtt = FakeMqttClient();
        final client = TwoFinanceBlockchain(
          keyManager: KeyManager(),
          mqttClient: mqtt,
          chainID: 1,
        );

        await client.getWalletByAddress(walletAddress);

        expect(mqtt.lastRequest['method'], REQUEST_METHOD_GET_STATE);
        expect(mqtt.lastRequest['params'], {
          'to': walletAddress,
          'method': METHOD_GET_WALLET_BY_ADDRESS,
          'data': {'address': walletAddress},
        });
      },
    );

    test(
      '[spec:wallet.get_by_public_key] getWalletByPublicKey preserves global public-key lookup',
      () async {
        final publicKey = await validPublicKeyHex();
        final mqtt = FakeMqttClient();
        final client = TwoFinanceBlockchain(
          keyManager: KeyManager(),
          mqttClient: mqtt,
          chainID: 1,
        );

        await client.getWalletByPublicKey(publicKey);

        expect(mqtt.lastRequest['method'], REQUEST_METHOD_GET_STATE);
        expect(mqtt.lastRequest['params'], {
          'to': '',
          'method': METHOD_GET_WALLET_BY_PUBLIC_KEY,
          'data': {
            'public_key': publicKey,
            'contract_version': WALLET_CONTRACT_V2,
          },
        });
      },
    );
  });
}
