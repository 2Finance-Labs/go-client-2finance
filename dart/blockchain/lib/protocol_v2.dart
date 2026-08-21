import 'package:two_finance_blockchain/blockchain/transaction/transaction.dart';
import 'package:two_finance_blockchain/blockchain/utils/json.dart';
import 'package:two_finance_blockchain/blockchain/utils/uuid.dart';
import 'package:two_finance_blockchain/wallet_manager.dart';

const int protocolVersionV2 = 2;
const int nativeGoRuntimeVersionV2 = 1;
const int evmRuntimeVersionV2 = 2;
const int transactionVersionV2 = 1;
const String runtimeNativeGoV2 = 'native_go';
const String runtimeEVMV2 = 'evm';
const String evmReservedAddressV2 =
    '2b6bf3044c330548317fb564120b29a0246eaec4a255145dd350ec288e63264e';
const String evmExecuteMethodV2 = 'evm_execute';
const String evmZeroValueV2 =
    '0000000000000000000000000000000000000000000000000000000000000000';

enum EVMMessageKindV2 {
  call,
  create;

  String get wireName => name;
}

class EVMAccessTupleV2 {
  const EVMAccessTupleV2({
    required this.address,
    this.storageKeys = const <String>[],
  });

  final String address;
  final List<String> storageKeys;

  Map<String, dynamic> toJson() => {
    'address': address,
    'storage_keys': storageKeys,
  };
}

class EVMMessageV2 {
  const EVMMessageV2({
    required this.kind,
    this.to,
    required this.value,
    required this.gasLimit,
    required this.calldata,
    this.accessList = const <EVMAccessTupleV2>[],
  });

  final EVMMessageKindV2 kind;
  final String? to;
  final String value;
  final int gasLimit;
  final String calldata;
  final List<EVMAccessTupleV2> accessList;

  Map<String, dynamic> toJson() => {
    'kind': kind.wireName,
    if (to != null) 'to': to,
    'value': value,
    'gas_limit': gasLimit,
    'calldata': calldata,
    if (accessList.isNotEmpty)
      'access_list': accessList.map((entry) => entry.toJson()).toList(),
  };
}

PreparedTransaction prepareNativeGoTransactionV2({
  required int chainID,
  required String from,
  required String to,
  required String method,
  required JsonMessage payload,
  String? uuid7,
  AuthorizationEnvelope? authorization,
}) {
  if (to.trim().isEmpty) {
    throw ArgumentError.value(to, 'to', 'native destination is required');
  }
  if (method.trim().isEmpty) {
    throw ArgumentError.value(method, 'method', 'native method is required');
  }
  return _prepareTransactionV2(
    chainID: chainID,
    from: from,
    to: to,
    method: method,
    payload: payload,
    runtime: runtimeNativeGoV2,
    runtimeVersion: nativeGoRuntimeVersionV2,
    uuid7: uuid7,
    authorization: authorization,
  );
}

PreparedTransaction prepareEVMTransactionV2({
  required int chainID,
  required String from,
  required EVMMessageV2 message,
  String? uuid7,
  AuthorizationEnvelope? authorization,
}) {
  if (message.kind == EVMMessageKindV2.call &&
      (message.to == null || message.to!.isEmpty)) {
    throw ArgumentError('EVM call destination is required');
  }
  if (message.kind == EVMMessageKindV2.create && message.to != null) {
    throw ArgumentError('EVM create must not contain a destination');
  }
  return _prepareTransactionV2(
    chainID: chainID,
    from: from,
    to: evmReservedAddressV2,
    method: evmExecuteMethodV2,
    payload: message.toJson(),
    runtime: runtimeEVMV2,
    runtimeVersion: evmRuntimeVersionV2,
    uuid7: uuid7,
    authorization: authorization,
  );
}

PreparedTransaction _prepareTransactionV2({
  required int chainID,
  required String from,
  required String to,
  required String method,
  required JsonMessage payload,
  required String runtime,
  required int runtimeVersion,
  String? uuid7,
  AuthorizationEnvelope? authorization,
}) {
  if (chainID < 1 || chainID > 2) {
    throw ArgumentError.value(chainID, 'chainID', 'must be 1 or 2');
  }
  if (from.trim().isEmpty) {
    throw ArgumentError.value(from, 'from', 'sender is required');
  }
  if (payload.isEmpty) {
    throw ArgumentError.value(payload, 'payload', 'must not be empty');
  }
  final id = uuid7 ?? newUUID7();
  validateUUID7(id);
  return PreparedTransaction(
    chainID: chainID,
    from: from,
    to: to,
    method: method,
    data: <String, dynamic>{
      'runtime': runtime,
      'version': runtimeVersion,
      'payload': Map<String, dynamic>.from(payload),
    },
    version: transactionVersionV2,
    uuid7: id,
    authorization: authorization,
  );
}

class ProtocolV2Commit {
  const ProtocolV2Commit({
    required this.sequence,
    required this.admissionHeight,
    required this.status,
    this.blockHeight,
    this.blockHash,
    this.membershipPosition,
    required this.committedAt,
    this.includedAt,
  });

  final int sequence;
  final int admissionHeight;
  final String status;
  final int? blockHeight;
  final String? blockHash;
  final int? membershipPosition;
  final DateTime committedAt;
  final DateTime? includedAt;

  factory ProtocolV2Commit.fromJson(Map<String, dynamic> json) =>
      ProtocolV2Commit(
        sequence: (json['sequence'] as num).toInt(),
        admissionHeight: (json['admission_height'] as num).toInt(),
        status: json['status'] as String,
        blockHeight: (json['block_height'] as num?)?.toInt(),
        blockHash: json['block_hash'] as String?,
        membershipPosition: (json['membership_position'] as num?)?.toInt(),
        committedAt: DateTime.parse(json['committed_at'] as String),
        includedAt: json['included_at'] == null
            ? null
            : DateTime.parse(json['included_at'] as String),
      );
}

class ProtocolV2Artifact {
  const ProtocolV2Artifact({
    required this.runtime,
    required this.runtimeVersion,
    required this.executionSpecHash,
    required this.receiptHash,
    required this.artifactHash,
    required this.logRoot,
    required this.writeSetRoot,
    required this.stateRootBefore,
    required this.stateRootAfter,
    required this.timestampHash,
  });

  final String runtime;
  final int runtimeVersion;
  final String executionSpecHash;
  final String receiptHash;
  final String artifactHash;
  final String logRoot;
  final String writeSetRoot;
  final String stateRootBefore;
  final String stateRootAfter;
  final String timestampHash;

  factory ProtocolV2Artifact.fromJson(Map<String, dynamic> json) =>
      ProtocolV2Artifact(
        runtime: json['runtime'] as String,
        runtimeVersion: (json['runtime_version'] as num).toInt(),
        executionSpecHash: json['execution_spec_hash'] as String,
        receiptHash: json['receipt_hash'] as String,
        artifactHash: json['artifact_hash'] as String,
        logRoot: json['log_root'] as String,
        writeSetRoot: json['write_set_root'] as String,
        stateRootBefore: json['state_root_before'] as String,
        stateRootAfter: json['state_root_after'] as String,
        timestampHash: json['timestamp_hash'] as String,
      );
}

class ProtocolV2Membership {
  const ProtocolV2Membership({
    required this.position,
    required this.transactionHash,
    required this.commitSequence,
  });

  final int position;
  final String transactionHash;
  final int commitSequence;

  factory ProtocolV2Membership.fromJson(Map<String, dynamic> json) =>
      ProtocolV2Membership(
        position: (json['position'] as num).toInt(),
        transactionHash: json['transaction_hash'] as String,
        commitSequence: (json['commit_sequence'] as num).toInt(),
      );
}

class ProtocolV2Finality {
  const ProtocolV2Finality({
    required this.blockEnvelope,
    required this.commitCertificate,
    required this.membership,
  });

  final Map<String, dynamic> blockEnvelope;
  final Map<String, dynamic> commitCertificate;
  final List<ProtocolV2Membership> membership;

  factory ProtocolV2Finality.fromJson(
    Map<String, dynamic> json,
  ) => ProtocolV2Finality(
    blockEnvelope: _map(json['block_envelope'], 'block_envelope'),
    commitCertificate: _map(json['commit_certificate'], 'commit_certificate'),
    membership: ((json['membership'] as List?) ?? const <dynamic>[])
        .map(
          (entry) =>
              ProtocolV2Membership.fromJson(_map(entry, 'membership entry')),
        )
        .toList(),
  );
}

class ProtocolV2TransactionStatus {
  const ProtocolV2TransactionStatus({
    required this.protocolVersion,
    required this.chainID,
    required this.transactionHash,
    required this.signedTransactionBody,
    required this.commit,
    required this.artifact,
    this.finality,
  });

  final int protocolVersion;
  final String chainID;
  final String transactionHash;
  final Map<String, dynamic> signedTransactionBody;
  final ProtocolV2Commit commit;
  final ProtocolV2Artifact artifact;
  final ProtocolV2Finality? finality;

  SignedTransaction get signedTransaction =>
      SignedTransaction.fromJson(signedTransactionBody);

  factory ProtocolV2TransactionStatus.fromJson(Map<String, dynamic> json) =>
      ProtocolV2TransactionStatus(
        protocolVersion: (json['protocol_version'] as num).toInt(),
        chainID: json['chain_id'] as String,
        transactionHash: json['transaction_hash'] as String,
        signedTransactionBody: _map(
          json['signed_transaction'],
          'signed_transaction',
        ),
        commit: ProtocolV2Commit.fromJson(_map(json['commit'], 'commit')),
        artifact: ProtocolV2Artifact.fromJson(
          _map(json['artifact'], 'artifact'),
        ),
        finality: json['finality'] == null
            ? null
            : ProtocolV2Finality.fromJson(_map(json['finality'], 'finality')),
      );
}

class ProtocolV2SubmitResult {
  const ProtocolV2SubmitResult({
    required this.executionOutput,
    required this.transaction,
  });

  final dynamic executionOutput;
  final ProtocolV2TransactionStatus transaction;

  factory ProtocolV2SubmitResult.fromJson(Map<String, dynamic> json) =>
      ProtocolV2SubmitResult(
        executionOutput: json['execution_output'],
        transaction: ProtocolV2TransactionStatus.fromJson(
          _map(json['transaction'], 'transaction'),
        ),
      );
}

class ProtocolV2ExecutionLog {
  const ProtocolV2ExecutionLog({
    required this.chainID,
    required this.transactionHash,
    required this.logIndex,
    required this.sourceLogIndex,
    required this.runtime,
    required this.contractAddress,
    required this.contractVersion,
    required this.eventSignature,
    required this.topics,
    required this.data,
    required this.rawData,
    required this.leafHash,
    required this.createdAt,
  });

  final String chainID;
  final String transactionHash;
  final int logIndex;
  final int sourceLogIndex;
  final String runtime;
  final String contractAddress;
  final String contractVersion;
  final String eventSignature;
  final List<String> topics;
  final dynamic data;
  final String rawData;
  final String leafHash;
  final DateTime createdAt;

  factory ProtocolV2ExecutionLog.fromJson(Map<String, dynamic> json) =>
      ProtocolV2ExecutionLog(
        chainID: json['chain_id'] as String,
        transactionHash: json['transaction_hash'] as String,
        logIndex: (json['log_index'] as num).toInt(),
        sourceLogIndex: (json['source_log_index'] as num).toInt(),
        runtime: json['runtime'] as String,
        contractAddress: json['contract_address'] as String,
        contractVersion: json['contract_version'] as String,
        eventSignature: json['event_signature'] as String,
        topics: ((json['topics'] as List?) ?? const <dynamic>[])
            .map((value) => value as String)
            .toList(),
        data: json['data'],
        rawData: json['raw_data'] as String,
        leafHash: json['leaf_hash'] as String,
        createdAt: DateTime.parse(json['created_at'] as String),
      );
}

class ProtocolV2ExecutionLogsPage {
  const ProtocolV2ExecutionLogsPage({
    required this.logs,
    required this.page,
    required this.limit,
  });

  final List<ProtocolV2ExecutionLog> logs;
  final int page;
  final int limit;

  factory ProtocolV2ExecutionLogsPage.fromJson(Map<String, dynamic> json) =>
      ProtocolV2ExecutionLogsPage(
        logs: ((json['logs'] as List?) ?? const <dynamic>[])
            .map(
              (entry) =>
                  ProtocolV2ExecutionLog.fromJson(_map(entry, 'execution log')),
            )
            .toList(),
        page: (json['page'] as num).toInt(),
        limit: (json['limit'] as num).toInt(),
      );
}

Map<String, dynamic> _map(dynamic value, String field) {
  if (value is! Map) {
    throw FormatException('$field is not an object');
  }
  return Map<String, dynamic>.from(value);
}
