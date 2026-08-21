# two_finance_blockchain

SDK Dart para o protocolo commit-first v2 da 2Finance Network.

```dart
final transport = HttpFinanceNetworkTransport(
  baseUrl: 'http://127.0.0.1:19295',
);
final blockchain = TwoFinanceBlockchain(
  keyManager: KeyManager(),
  mqttClient: transport,
  chainID: 2,
  walletManager: walletManager,
);

final prepared = prepareEVMTransactionV2(
  chainID: 2,
  from: walletManager.ownerAddress,
  message: const EVMMessageV2(
    kind: EVMMessageKindV2.create,
    value: evmZeroValueV2,
    gasLimit: 300000,
    calldata: '600a600c600039600a6000f3602a60005260206000f3',
  ),
);
final result = await blockchain.signAndSubmitPreparedTransactionV2(prepared);
final proof = await blockchain.transactionFinalityV2(
  result.transaction.transactionHash,
);
```

Para contratos Native Go, use `prepareNativeGoTransactionV2`. As duas VMs
usam a mesma `SignedTransaction`; somente `data.runtime` e o seletor mudam.


## Tests

Local
```
dart test
```

E2E HTTP v2
```
TWO_FINANCE_NETWORK_URL=http://127.0.0.1:19295 dart test test/blockchain/contract/walletV2/wallet_test.dart
```
