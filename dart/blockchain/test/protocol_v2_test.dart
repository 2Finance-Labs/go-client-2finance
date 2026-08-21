import 'package:test/test.dart';
import 'package:two_finance_blockchain/blockchain/types/types.dart';
import 'package:two_finance_blockchain/blockchain/utils/uuid.dart';
import 'package:two_finance_blockchain/protocol_v2.dart';

void main() {
  test('one prepared transaction shape selects Native Go or EVM', () {
    final from = '1' * 64;
    final native = prepareNativeGoTransactionV2(
      chainID: 2,
      from: from,
      to: DEPLOY_CONTRACT_ADDRESS,
      method: 'deploy_contract',
      payload: {'contract_version': 'walletV2'},
    );
    final evm = prepareEVMTransactionV2(
      chainID: 2,
      from: from,
      message: const EVMMessageV2(
        kind: EVMMessageKindV2.create,
        value: evmZeroValueV2,
        gasLimit: 300000,
        calldata: '600a600c600039600a6000f3602a60005260206000f3',
      ),
    );

    expect(native.data['runtime'], runtimeNativeGoV2);
    expect(native.data['version'], nativeGoRuntimeVersionV2);
    expect(native.data['payload'], {'contract_version': 'walletV2'});
    expect(evm.data['runtime'], runtimeEVMV2);
    expect(evm.data['version'], evmRuntimeVersionV2);
    expect(evm.to, evmReservedAddressV2);
    expect(evm.method, evmExecuteMethodV2);
    expect((evm.data['payload'] as Map)['kind'], 'create');
    expect(native.version, transactionVersionV2);
    expect(evm.version, transactionVersionV2);
    expect(() => validateUUID7(native.uuid7), returnsNormally);
    expect(() => validateUUID7(evm.uuid7), returnsNormally);
  });

  test(
    'transaction status exposes signed body, commit, artifact and proof',
    () {
      final hash = 'a' * 64;
      final status = ProtocolV2TransactionStatus.fromJson({
        'protocol_version': 2,
        'chain_id': '2finance-dev-v2',
        'transaction_hash': hash,
        'signed_transaction': {
          'chain_id': 2,
          'from': '1' * 64,
          'to': DEPLOY_CONTRACT_ADDRESS,
          'method': 'deploy_contract',
          'data': {
            'runtime': 'native_go',
            'version': 1,
            'payload': {'contract_version': 'walletV2'},
          },
          'version': 1,
          'uuid7': '019fba91-223d-7c2e-9825-65eb7c3607e0',
          'hash': hash,
          'signature': 'b' * 128,
        },
        'commit': {
          'sequence': 1,
          'admission_height': 2,
          'status': 'included',
          'block_height': 2,
          'block_hash': 'c' * 64,
          'membership_position': 0,
          'committed_at': '2026-08-01T01:02:03Z',
          'included_at': '2026-08-01T01:02:04Z',
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
        'finality': {
          'block_envelope': {'protocol_version': 2, 'height': 2},
          'commit_certificate': {
            'certificate_version': 2,
            'signed_voting_power': 3,
            'total_voting_power': 3,
          },
          'membership': [
            {'position': 0, 'transaction_hash': hash, 'commit_sequence': 1},
          ],
        },
      });

      expect(status.protocolVersion, 2);
      expect(status.signedTransaction.hash, hash);
      expect(status.commit.sequence, 1);
      expect(status.artifact.runtime, runtimeNativeGoV2);
      expect(status.finality?.membership.single.transactionHash, hash);
      expect(status.finality?.commitCertificate['signed_voting_power'], 3);
    },
  );
}
