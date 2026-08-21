import 'dart:convert';
import 'dart:io';
import 'dart:typed_data';

import 'package:test/test.dart';
import 'package:two_finance_blockchain/blockchain/types/types.dart';
import 'package:two_finance_blockchain/blockchain/utils/uuid.dart';
import 'package:two_finance_blockchain/wallet_manager.dart';

void main() {
  group('WalletManager', () {
    test(
      'imports, unlocks, signs prepared transactions, and changes password',
      () async {
        final dir = await Directory.systemTemp.createTemp(
          'wallet-manager-test-',
        );
        addTearDown(() async {
          if (await dir.exists()) {
            await dir.delete(recursive: true);
          }
        });

        final walletPath = '${dir.path}/wallet.json';
        final keyPair = await generateEd25519KeyPairHex();
        final manager = WalletManager(walletPath);

        await manager.importPrivateKey(
          Uint8List.fromList(utf8.encode(keyPair.privateKey)),
          'old-password',
        );

        expect(manager.ownerAddress, keyPair.publicKey);
        expect(manager.isUnlocked(), isFalse);
        expect(File(walletPath).existsSync(), isTrue);

        final prepared = PreparedTransaction(
          chainID: 1,
          from: manager.ownerAddress,
          to: DEPLOY_CONTRACT_ADDRESS,
          method: 'deploy_contract',
          data: {'contract_version': 'walletV2'},
          version: 1,
          uuid7: newUUID7(),
        );

        await expectLater(
          manager.unlockWithPassword('wrong-password'),
          throwsA(isA<StateError>()),
        );
        expect(manager.isUnlocked(), isFalse);
        await expectLater(
          manager.signPreparedTransaction(prepared),
          throwsA(isA<StateError>()),
        );

        await manager.unlockWithPassword('old-password');
        expect(manager.isUnlocked(), isTrue);

        await expectLater(
          manager.signPreparedTransaction(
            PreparedTransaction(
              chainID: 1,
              from: manager.ownerAddress,
              to: DEPLOY_CONTRACT_ADDRESS,
              method: 'deploy_contract',
              data: const <String, dynamic>{},
              version: 1,
              uuid7: newUUID7(),
            ),
          ),
          throwsA(isA<Exception>()),
        );

        final signed = await manager.signPreparedTransaction(prepared);
        expect(signed.hash, hasLength(64));
        expect(signed.signature, hasLength(128));

        await manager.changePassword('old-password', 'new-password');
        expect(manager.isUnlocked(), isFalse);

        await manager.unlockWithPassword('new-password');
        expect(manager.isUnlocked(), isTrue);

        await manager.lock();
        expect(manager.isUnlocked(), isFalse);
      },
    );
  });
}
