import 'package:test/test.dart';
import 'package:two_finance_blockchain/blockchain/contract/constants.dart';
import 'package:two_finance_blockchain/two_finance_blockchain.dart';

import 'helpers/fake_mqtt.dart';

String repeatChar(String char, int length) => List.filled(length, char).join();

void main() {
  group('TwoFinanceBlockchain client helpers', () {
    test('setChainID and getChainID update the active chain', () {
      final client = TwoFinanceBlockchain(
        keyManager: KeyManager(),
        mqttClient: FakeMqttClient(),
        chainID: 1,
      );

      expect(client.getChainID(), 1);
      client.setChainID(2);
      expect(client.getChainID(), 2);
      expect(() => client.setChainID(0), throwsException);
    });

    test(
      '[spec:client.list_transactions] listTransactions sends Go-compatible filters and parses response',
      () async {
        final mqtt = FakeMqttClient(
          responseBuilder: (_) => {
            'status': 'success',
            'message': null,
            'data': [
              {
                'chain_id': 1,
                'from': repeatChar('f', 64),
                'to': repeatChar('a', 64),
                'method': 'add_Wallet',
                'data': {'address': repeatChar('a', 64)},
                'version': 1,
                'uuid7': '01931d90-56ec-7b4b-b579-e512831e82c9',
                'hash': repeatChar('b', 64),
                'signature': repeatChar('c', 128),
              },
            ],
          },
        );
        final client = TwoFinanceBlockchain(
          keyManager: KeyManager(),
          mqttClient: mqtt,
          chainID: 1,
        );

        final transactions = await client.listTransactions(
          from: repeatChar('f', 64),
          to: repeatChar('a', 64),
          hash: repeatChar('b', 64),
          dataFilter: {'address': repeatChar('a', 64)},
          version: 1,
          uuid7: '01931d90-56ec-7b4b-b579-e512831e82c9',
          page: 2,
          limit: 25,
          ascending: true,
        );

        expect(transactions, hasLength(1));
        expect(transactions.single.method, 'add_Wallet');
        expect(mqtt.lastRequest['method'], REQUEST_METHOD_GET_TRANSACTIONS);
        expect(mqtt.lastRequest['params'], {
          'from': repeatChar('f', 64),
          'to': repeatChar('a', 64),
          'hash': repeatChar('b', 64),
          'data': {'address': repeatChar('a', 64)},
          'version': 1,
          'uuid7': '01931d90-56ec-7b4b-b579-e512831e82c9',
          'page': 2,
          'limit': 25,
          'ascending': true,
        });
      },
    );

    test(
      '[spec:client.list_logs] listLogs sends Go-compatible filters and parses response',
      () async {
        final mqtt = FakeMqttClient(
          responseBuilder: (_) => {
            'status': 'success',
            'message': null,
            'data': [
              {
                'log_type': 'Wallet_Created',
                'log_index': 1,
                'transaction_hash': repeatChar('b', 64),
                'event': 'eyJvayI6dHJ1ZX0=',
                'contract_version': 'walletV2',
                'contract_address': repeatChar('a', 64),
              },
            ],
          },
        );
        final client = TwoFinanceBlockchain(
          keyManager: KeyManager(),
          mqttClient: mqtt,
          chainID: 1,
        );

        final logs = await client.listLogs(
          logType: const ['Wallet_Created'],
          logIndex: 1,
          transactionHash: repeatChar('b', 64),
          event: const {'address': 'a'},
          contractAddress: repeatChar('a', 64),
          contractVersion: 'walletV2',
          page: 1,
          limit: 10,
          ascending: true,
        );

        expect(logs, hasLength(1));
        expect(logs.single.logType, 'Wallet_Created');
        expect(mqtt.lastRequest['method'], REQUEST_METHOD_GET_LOGS);
        expect(mqtt.lastRequest['params'], {
          'log_type': ['Wallet_Created'],
          'log_index': 1,
          'transaction_hash': repeatChar('b', 64),
          'event': {'address': 'a'},
          'contract_address': repeatChar('a', 64),
          'contract_version': 'walletV2',
          'page': 1,
          'limit': 10,
          'ascending': true,
        });
      },
    );

    test(
      '[spec:client.list_blocks] listBlocks sends Go-compatible filters and parses response',
      () async {
        final timestamp = DateTime.utc(2026, 1, 2, 3, 4, 5);
        final mqtt = FakeMqttClient(
          responseBuilder: (_) => {
            'status': 'success',
            'message': null,
            'data': [
              {
                'number': 7,
                'timestamp': timestamp.toIso8601String(),
                'tx_root': repeatChar('1', 64),
                'log_root': repeatChar('2', 64),
                'state_root': repeatChar('3', 64),
                'timestamp_root': repeatChar('4', 64),
                'tx_count': 2,
                'log_count': 3,
                'state_snapshot_count': 4,
                'hash': repeatChar('5', 64),
                'parent_hash': repeatChar('6', 64),
                'created_at': timestamp.toIso8601String(),
              },
            ],
          },
        );
        final client = TwoFinanceBlockchain(
          keyManager: KeyManager(),
          mqttClient: mqtt,
          chainID: 1,
        );

        final blocks = await client.listBlocks(
          number: 7,
          timestamp: timestamp,
          hash: repeatChar('5', 64),
          previousHash: repeatChar('6', 64),
          transactionsMerkleRoot: repeatChar('1', 64),
          logsMerkleRoot: repeatChar('2', 64),
          statesSnapshotMerkleRoot: repeatChar('3', 64),
          page: 1,
          limit: 5,
          ascending: true,
        );

        expect(blocks, hasLength(1));
        expect(blocks.single.number, 7);
        expect(blocks.single.txRoot, repeatChar('1', 64));
        expect(mqtt.lastRequest['method'], REQUEST_METHOD_GET_BLOCKS);
        expect(mqtt.lastRequest['params'], {
          'number': 7,
          'timestamp': timestamp.toIso8601String(),
          'hash': repeatChar('5', 64),
          'previous_hash': repeatChar('6', 64),
          'transactions_merkle_root': repeatChar('1', 64),
          'logs_merkle_root': repeatChar('2', 64),
          'states_snapshot_merkle_root': repeatChar('3', 64),
          'page': 1,
          'limit': 5,
          'ascending': true,
        });
      },
    );
  });
}
