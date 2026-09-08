library two_finance_blockchain;

import 'dart:async';
import 'dart:convert';
import 'dart:typed_data';

import 'package:cryptography/cryptography.dart';
import 'package:uuid/uuid.dart';

import 'package:two_finance_blockchain/blockchain/contract/constants.dart';
import 'package:two_finance_blockchain/blockchain/contract/tokenV2/domain/token.dart';
import 'package:two_finance_blockchain/blockchain/contract/walletV2/constants.dart';
import 'package:two_finance_blockchain/blockchain/block/block.dart';
import 'package:two_finance_blockchain/blockchain/keys/keys.dart';
import 'package:two_finance_blockchain/blockchain/transaction/transaction.dart';
import 'package:two_finance_blockchain/blockchain/types/types.dart';
import 'package:two_finance_blockchain/blockchain/log/log.dart';

import 'package:two_finance_blockchain/blockchain/utils/decimals.dart';
import 'package:two_finance_blockchain/blockchain/utils/json.dart';
import 'package:two_finance_blockchain/blockchain/utils/uuid.dart';
import 'package:two_finance_blockchain/infra/transport/transport.dart';
import 'package:two_finance_blockchain/protocol_v2.dart';

import 'package:two_finance_blockchain/blockchain/contract/tokenV2/constants.dart';
import 'blockchain/contract/raffleV2/constants.dart';
import 'blockchain/contract/reviewV2/constants.dart';
import 'blockchain/contract/cashbackV2/constants.dart';
import 'blockchain/contract/paymentV2/constants.dart';
import 'blockchain/contract/couponsV2/constants.dart';
import 'blockchain/contract/dropV2/constants.dart';
import 'blockchain/contract/contractV2/constants.dart';
export 'blockchain/block/block.dart';
export 'blockchain/keys/keys.dart';
export 'blockchain/log/log.dart';
export 'blockchain/transaction/transaction.dart';
export 'blockchain/types/types.dart';
export 'infra/http/http_transport.dart';
export 'infra/mqtt/mqtt_stub.dart' if (dart.library.io) 'infra/mqtt/mqtt.dart';
export 'infra/transport/transport.dart';
export 'protocol_v2.dart';
export 'trading_authorization.dart';
export 'wallet_connect_protocol.dart';
export 'wallet_manager.dart';
import 'wallet_manager.dart';

part 'review.dart';
part 'token.dart';
part 'lifecycle.dart';
part 'wallet.dart';
part 'raffle.dart';
part 'cashback.dart';
part 'payment.dart';
part 'faucet.dart';
part 'coupons.dart';
part 'member_get_member.dart';
part 'drop.dart';

class TwoFinanceBlockchain {
  String? _privateKeyHex;
  String? _publicKeyHex;
  IWalletManager? _walletManager;
  String? get publicKeyHex => _walletManager?.ownerAddress ?? _publicKeyHex;

  bool _isInitialized = false;
  bool get isInitialized => _isInitialized;
  late String _replyTo;

  Future<void> initialize() async {
    if (_isInitialized) return;
    // initState();
    _isInitialized = true;
  }

  void _initState() {
    //super.initState();
    final uuid = Uuid();
    _replyTo = uuid.v4();
  }

  final KeyManager _keyManager;
  final FinanceNetworkTransport _mqttClient;
  int _chainID;

  TwoFinanceBlockchain({
    required KeyManager keyManager,
    required FinanceNetworkTransport mqttClient,
    required int chainID,
    IWalletManager? walletManager,
  }) : _keyManager = keyManager,
       _mqttClient = mqttClient,
       _chainID = chainID,
       _walletManager = walletManager {
    _initState();
  }

  void setWalletManager(IWalletManager walletManager) {
    _walletManager = walletManager;
  }

  void setChainID(int chainID) {
    if (chainID <= 0) {
      throw Exception('chain ID must be greater than zero');
    }
    _chainID = chainID;
  }

  int getChainID() => _chainID;

  Future<void> setPrivateKey(String privateKeyHex) async {
    try {
      final algorithm = Ed25519();

      // Usa a função utilitária da própria classe
      final Uint8List privateKeyBytes = KeyManager.hexToBytes(privateKeyHex);

      if (privateKeyBytes.length < 32) {
        throw Exception(
          "Chave privada muito curta para derivar a semente (precisa de pelo menos 32 bytes = 64 hex). "
          "Recebido: ${privateKeyBytes.length} bytes.",
        );
      }

      // Usa apenas os primeiros 32 bytes como semente
      final seedBytes = Uint8List.fromList(privateKeyBytes.sublist(0, 32));

      final keyPair = await algorithm.newKeyPairFromSeed(seedBytes);
      final publicKeyBytes = (await keyPair.extractPublicKey()).bytes;

      // Armazena as chaves ativas usando os métodos auxiliares
      _privateKeyHex = privateKeyHex;
      _publicKeyHex = KeyManager.bytesToHex(publicKeyBytes);

      print('✅ Chave privada definida e chave pública derivada com sucesso!');
      print('🔑 Chave Pública Ativa: $_publicKeyHex');
    } on FormatException catch (e) {
      throw FormatException('Erro de formato na chave privada: ${e.message}');
    } catch (e) {
      throw Exception(
        'Erro ao definir chave privada e derivar chave pública: $e',
      );
    }
  }

  Future<KeyPair2Finance> generateKeyEd25519() async {
    return await _keyManager.generateKeyEd25519();
  }

  Future<Uint8List> sendTransaction(
    String method,
    dynamic tx,
    String replyTo,
  ) async {
    final data = await _mqttClient.sendRequest(method, tx, replyTo);
    final encodedData = json.encode(data);
    return Uint8List.fromList(utf8.encode(encodedData));
  }

  Future<ContractOutput> signAndSendTransaction({
    required int chainID,
    required String from,
    required String to,
    required String method,
    required JsonMessage data,
    required int version,
    required String uuid7,
  }) async {
    KeyManager.validateEDDSAPublicKeyHex(from);

    final transactionData =
        _mqttClient is ProtocolV2FinanceNetworkTransport &&
            data['runtime'] == null
        ? <String, dynamic>{
            'runtime': runtimeNativeGoV2,
            'version': nativeGoRuntimeVersionV2,
            'payload': Map<String, dynamic>.from(data),
          }
        : data;

    final newTx = Transaction.create(
      chainID: chainID,
      from: from,
      to: to,
      method: method,
      data: transactionData,
      version: version,
      uuid7: uuid7,
    );

    final walletManager = _walletManager;
    Transaction txSigned;
    if (walletManager != null) {
      txSigned = await walletManager.signTransaction(
        SignTransactionInput(
          chainID: chainID,
          from: from,
          to: to,
          method: method,
          data: transactionData,
          version: version,
          uuid7: uuid7,
        ),
      );
    } else {
      final privateKey = _privateKeyHex;
      if (privateKey == null) {
        throw Exception("Active private key is not initialized");
      }
      final tx = newTx.get();
      txSigned = await signTransaction(privateKey, tx);
    }
    if (_mqttClient is ProtocolV2FinanceNetworkTransport) {
      final result = await submitSignedTransactionV2(
        SignedTransaction.fromTransaction(txSigned),
      );
      return _contractOutput(result.executionOutput);
    }

    throw StateError(
      'protocol V2 is unavailable; configure a protocol V2 transport',
    );
  }

  Future<SignedTransaction> signPreparedTransaction(
    PreparedTransaction tx,
  ) async {
    final walletManager = _walletManager;
    if (walletManager == null) {
      throw Exception('wallet manager is required');
    }

    return walletManager.signPreparedTransaction(tx);
  }

  Future<ContractOutput> submitSignedTransaction(SignedTransaction tx) async {
    if (_mqttClient is ProtocolV2FinanceNetworkTransport) {
      final result = await submitSignedTransactionV2(tx);
      return _contractOutput(result.executionOutput);
    }
    throw StateError(
      'protocol V2 is unavailable; configure a protocol V2 transport',
    );
  }

  Future<ContractOutput> signAndSendPreparedTransaction(
    PreparedTransaction tx,
  ) async {
    if (_mqttClient is! ProtocolV2FinanceNetworkTransport) {
      throw StateError(
        'protocol V2 is discontinued; configure a protocol V2 transport',
      );
    }
    var prepared = tx;
    if (_mqttClient is ProtocolV2FinanceNetworkTransport &&
        prepared.data['runtime'] == null) {
      prepared = PreparedTransaction(
        chainID: tx.chainID,
        from: tx.from,
        to: tx.to,
        method: tx.method,
        data: <String, dynamic>{
          'runtime': runtimeNativeGoV2,
          'version': nativeGoRuntimeVersionV2,
          'payload': Map<String, dynamic>.from(tx.data),
        },
        version: tx.version,
        uuid7: tx.uuid7,
        authorization: tx.authorization,
      );
    }
    final signed = await signPreparedTransaction(prepared);
    return submitSignedTransaction(signed);
  }

  /// Submits an already signed transaction through the commit-first v2 API.
  Future<ProtocolV2SubmitResult> submitSignedTransactionV2(
    SignedTransaction tx,
  ) async {
    final transport = _protocolV2Transport();
    final response = await transport.submitSignedTransactionV2(tx.toJson());
    return ProtocolV2SubmitResult.fromJson(response);
  }

  Future<ProtocolV2SubmitResult> signAndSubmitPreparedTransactionV2(
    PreparedTransaction tx,
  ) async {
    final signed = await signPreparedTransaction(tx);
    return submitSignedTransactionV2(signed);
  }

  /// Returns the signed body, early commit, execution artifact and block proof.
  Future<ProtocolV2TransactionStatus> transactionFinalityV2(
    String transactionHash,
  ) async {
    final transport = _protocolV2Transport();
    final response = await transport.transactionFinalityV2(transactionHash);
    return ProtocolV2TransactionStatus.fromJson(response);
  }

  Future<ProtocolV2ExecutionLogsPage> executionLogsV2({
    String transactionHash = '',
    String runtime = '',
    String contractAddress = '',
    String eventSignature = '',
    int page = 1,
    int limit = 100,
    bool ascending = false,
  }) async {
    if (page < 1 || limit < 1 || limit > 1000) {
      throw ArgumentError('page must be positive and limit must be 1..1000');
    }
    final response = await _protocolV2Transport().executionLogsV2({
      'transaction_hash': transactionHash,
      'runtime': runtime,
      'contract_address': contractAddress,
      'event_signature': eventSignature,
      'page': '$page',
      'limit': '$limit',
      'ascending': '$ascending',
    });
    return ProtocolV2ExecutionLogsPage.fromJson(response);
  }

  ProtocolV2FinanceNetworkTransport _protocolV2Transport() {
    final transport = _mqttClient;
    if (transport is! ProtocolV2FinanceNetworkTransport) {
      throw StateError(
        'the configured transport does not support the protocol v2 REST API',
      );
    }
    return transport as ProtocolV2FinanceNetworkTransport;
  }

  ContractOutput _contractOutput(dynamic decoded) {
    if (decoded is Map) {
      return ContractOutput.fromJson(Map<String, dynamic>.from(decoded));
    }
    if (decoded is int && decoded == 0 || decoded == null) {
      return ContractOutput(states: const <StateType>[], logs: const <Log>[]);
    }
    throw Exception(
      'unexpected transaction response type: ${decoded.runtimeType}',
    );
  }

  Future<ContractOutput> getState({
    required String to,
    required String method,
    required JsonMessage data,
  }) async {
    try {
      final txInput = {'to': to, 'method': method, 'data': data};

      final responseBytes = await sendTransaction(
        REQUEST_METHOD_GET_STATE,
        txInput,
        _replyTo,
      );

      final dynamic decoded = json.decode(utf8.decode(responseBytes));
      // ✅ caso normal: veio um objeto JSON (ContractOutput)
      if (decoded is Map<String, dynamic>) {
        return ContractOutput.fromJson(decoded);
      }

      // ✅ caso "not found": veio 0 (fallback)
      if (decoded is int && decoded == 0) {
        // escolha 1: retornar vazio com listas vazias (mais fácil pra teste)
        return ContractOutput(states: const <StateType>[], logs: const <Log>[]);
        // escolha 2 (se você preferir zero-value mesmo): return ContractOutput();
      }

      // ✅ caso "null" (se o Go realmente mandar null em algum cenário)
      if (decoded == null) {
        return ContractOutput();
      }

      throw Exception(
        'unexpected getState response type: ${decoded.runtimeType}',
      );
    } catch (e) {
      throw Exception('failed to get state: from getState function $e');
    }
  }

  Future<List<Transaction>> listTransactions({
    String from = '',
    String to = '',
    String hash = '',
    JsonMessage dataFilter = const {},
    int version = 0,
    String uuid7 = '',
    int page = 0,
    int limit = 0,
    bool ascending = false,
  }) async {
    final txInput = {
      'from': from,
      'to': to,
      'hash': hash,
      'data': dataFilter,
      'version': version,
      'uuid7': uuid7,
      'page': page,
      'limit': limit,
      'ascending': ascending,
    };

    final responseBytes = await sendTransaction(
      REQUEST_METHOD_GET_TRANSACTIONS,
      txInput,
      _replyTo,
    );
    final decoded = json.decode(utf8.decode(responseBytes));
    if (decoded is int && decoded == 0) return const <Transaction>[];
    if (decoded is! List) {
      throw Exception(
        'unexpected transactions response type: ${decoded.runtimeType}',
      );
    }
    return decoded
        .map((item) => Transaction.fromJson(Map<String, dynamic>.from(item)))
        .toList();
  }

  Future<List<Log>> listLogs({
    List<String> logType = const <String>[],
    int logIndex = 0,
    String transactionHash = '',
    JsonMessage event = const {},
    String contractAddress = '',
    String contractVersion = '',
    int page = 0,
    int limit = 0,
    bool ascending = false,
  }) async {
    final logInput = {
      'log_type': logType,
      'log_index': logIndex,
      'transaction_hash': transactionHash,
      'event': event,
      'contract_address': contractAddress,
      'contract_version': contractVersion,
      'page': page,
      'limit': limit,
      'ascending': ascending,
    };

    final responseBytes = await sendTransaction(
      REQUEST_METHOD_GET_LOGS,
      logInput,
      _replyTo,
    );
    final decoded = json.decode(utf8.decode(responseBytes));
    if (decoded is int && decoded == 0) return const <Log>[];
    if (decoded is! List) {
      throw Exception('unexpected logs response type: ${decoded.runtimeType}');
    }
    return decoded
        .map((item) => Log.fromJson(Map<String, dynamic>.from(item)))
        .toList();
  }

  Future<List<Block>> listBlocks({
    int number = 0,
    DateTime? timestamp,
    String hash = '',
    String previousHash = '',
    String transactionsMerkleRoot = '',
    String logsMerkleRoot = '',
    String statesSnapshotMerkleRoot = '',
    int page = 0,
    int limit = 0,
    bool ascending = false,
  }) async {
    final blockInput = {
      'number': number,
      if (timestamp != null) 'timestamp': timestamp.toUtc().toIso8601String(),
      'hash': hash,
      'previous_hash': previousHash,
      'transactions_merkle_root': transactionsMerkleRoot,
      'logs_merkle_root': logsMerkleRoot,
      'states_snapshot_merkle_root': statesSnapshotMerkleRoot,
      'page': page,
      'limit': limit,
      'ascending': ascending,
    };

    final responseBytes = await sendTransaction(
      REQUEST_METHOD_GET_BLOCKS,
      blockInput,
      _replyTo,
    );
    final decoded = json.decode(utf8.decode(responseBytes));
    if (decoded is int && decoded == 0) return const <Block>[];
    if (decoded is! List) {
      throw Exception(
        'unexpected blocks response type: ${decoded.runtimeType}',
      );
    }
    return decoded
        .map((item) => Block.fromJson(Map<String, dynamic>.from(item)))
        .toList();
  }

  Future<ContractOutput> deployContract1(String contractVersion) async {
    contractVersion = _publicContractVersionV2(contractVersion);
    print("Deploying contract version: $contractVersion");

    final chainID = _chainID;
    final from = _publicKeyHex!;
    if (from.isEmpty) {
      throw Exception('from address is required');
    }

    KeyManager.validateEDDSAPublicKeyHex(from);
    if (contractVersion.isEmpty) {
      throw Exception('contract version is required');
    }

    String to = DEPLOY_CONTRACT_ADDRESS;

    final method = METHOD_DEPLOY_CONTRACT;
    final JsonMessage data = {'contract_version': contractVersion};

    final version = 1;
    final uuid7 = newUUID7();

    try {
      final contractOutput = await signAndSendTransaction(
        chainID: chainID,
        from: from,
        to: to,
        method: method,
        data: data,
        version: version,
        uuid7: uuid7,
      );
      return contractOutput;
    } catch (e) {
      throw Exception('failed to deploy contract: $e');
    }
  }

  Future<ContractOutput> deployContract2(
    String contractAddress,
    String contractVersion,
  ) async {
    contractVersion = _publicContractVersionV2(contractVersion);
    print(
      "Deploying contract version: $contractVersion to address: $contractAddress",
    );

    final chainID = _chainID;
    final from = _publicKeyHex!;
    if (from.isEmpty) {
      throw Exception('from address is required');
    }

    KeyManager.validateEDDSAPublicKeyHex(from);
    if (contractVersion.isEmpty) {
      throw Exception('contract version is required');
    }

    String to = "";
    if (contractAddress.isNotEmpty) {
      to = contractAddress;
    }

    final method = METHOD_DEPLOY_CONTRACT;
    final JsonMessage data = {'contract_version': contractVersion};

    final version = 1;
    final uuid7 = newUUID7();

    try {
      final contractOutput = await signAndSendTransaction(
        chainID: chainID,
        from: from,
        to: to,
        method: method,
        data: data,
        version: version,
        uuid7: uuid7,
      );
      return contractOutput;
    } catch (e) {
      throw Exception('failed to deploy contract: $e');
    }
  }

  String _publicContractVersionV2(String version) {
    final normalized = version.trim();
    if (normalized.endsWith('V2')) return normalized;
    if (normalized.endsWith('V2')) {
      return '${normalized.substring(0, normalized.length - 2)}V2';
    }
    throw ArgumentError.value(
      version,
      'contractVersion',
      'only V2 native contract versions are supported',
    );
  }
}
