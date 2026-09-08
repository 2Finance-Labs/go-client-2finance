import 'package:test/test.dart';
import 'package:two_finance_blockchain/two_finance_blockchain.dart';

class RecordingProtocolV2Transport
    implements FinanceNetworkTransport, ProtocolV2FinanceNetworkTransport {
  Map<String, dynamic>? submitted;

  @override
  Future<dynamic> sendRequest(String method, dynamic params, String replyTo) {
    throw StateError('legacy request path must not be used for v2 writes');
  }

  @override
  Future<Map<String, dynamic>> submitSignedTransactionV2(
    Map<String, dynamic> signedTransaction,
  ) async {
    submitted = signedTransaction;
    return {
      'execution_output': {'states': <dynamic>[], 'logs': <dynamic>[]},
      'transaction': _status(signedTransaction),
    };
  }

  @override
  Future<Map<String, dynamic>> transactionFinalityV2(
    String transactionHash,
  ) async {
    final signed = submitted;
    if (signed == null) throw StateError('no transaction submitted');
    return _status(signed);
  }

  @override
  Future<Map<String, dynamic>> executionLogsV2(
    Map<String, String> query,
  ) async => {'logs': <dynamic>[], 'page': 1, 'limit': 100};

  Map<String, dynamic> _status(Map<String, dynamic> signed) => {
    'protocol_version': 2,
    'chain_id': '2finance-dev-v2',
    'transaction_hash': signed['hash'],
    'signed_transaction': signed,
    'commit': {
      'sequence': 1,
      'admission_height': 2,
      'status': 'committed',
      'committed_at': '2026-08-01T01:02:03Z',
    },
    'artifact': {
      'runtime': 'native_go',
      'runtime_version': 1,
      'execution_spec_hash': '1' * 64,
      'receipt_hash': '2' * 64,
      'artifact_hash': '3' * 64,
      'log_root': '4' * 64,
      'write_set_root': '5' * 64,
      'state_root_before': '6' * 64,
      'state_root_after': '7' * 64,
      'timestamp_hash': '8' * 64,
    },
  };
}

void main() {
  test(
    'contract write is signed with Native Go envelope and sent over v2',
    () async {
      final transport = RecordingProtocolV2Transport();
      final client = TwoFinanceBlockchain(
        keyManager: KeyManager(),
        mqttClient: transport,
        chainID: 2,
      );
      await client.setPrivateKey('60' * 32);
      final from = client.publicKeyHex!;

      final output = await client.signAndSendTransaction(
        chainID: 2,
        from: from,
        to: DEPLOY_CONTRACT_ADDRESS,
        method: 'deploy_contract',
        data: {'contract_version': 'walletV2'},
        version: 1,
        uuid7: '019fba91-223d-7c2e-9825-65eb7c3607e0',
      );

      expect(output.states, isEmpty);
      final submitted = transport.submitted!;
      expect(submitted['hash'], hasLength(64));
      expect(submitted['signature'], hasLength(128));
      final data = Map<String, dynamic>.from(submitted['data'] as Map);
      expect(data['runtime'], runtimeNativeGoV2);
      expect(data['version'], nativeGoRuntimeVersionV2);
      expect(data['payload'], {'contract_version': 'walletV2'});

      final status = await client.transactionFinalityV2(
        submitted['hash'] as String,
      );
      expect(status.signedTransaction.signature, submitted['signature']);
      expect(status.commit.sequence, 1);

      final logs = await client.executionLogsV2(runtime: runtimeNativeGoV2);
      expect(logs.logs, isEmpty);
    },
  );
}
