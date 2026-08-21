import 'dart:convert';

import 'package:mqtt_client/mqtt_client.dart';
import 'package:two_finance_blockchain/blockchain/contract/constants.dart';
import 'package:two_finance_blockchain/infra/event/request_response.dart';
import 'package:two_finance_blockchain/infra/mqtt/mqtt.dart';
import 'package:two_finance_blockchain/infra/transport/transport.dart';

typedef FakeResponseBuilder =
    Map<String, dynamic> Function(Map<String, dynamic> request);

class FakeMqttClient
    implements MqttClientInterface, ProtocolV2FinanceNetworkTransport {
  FakeMqttClient({FakeResponseBuilder? responseBuilder})
    : _responseBuilder =
          responseBuilder ??
          ((_) => {
            'status': RESPONSE_STATUS_SUCCESS,
            'message': null,
            'data': {'states': <dynamic>[], 'logs': <dynamic>[]},
          });

  final FakeResponseBuilder _responseBuilder;
  final MqttClient _client = MqttClient('localhost', 'fake-client');
  final List<Map<String, dynamic>> publishedRequests = <Map<String, dynamic>>[];
  final Map<String, MessageHandler?> _handlers = <String, MessageHandler?>{};
  Map<String, dynamic>? _submittedV2;

  @override
  MqttClient? get client => _client;

  Map<String, dynamic> get lastRequest => publishedRequests.last;

  @override
  Future<void> connect() async {}

  @override
  Future<void> disconnect() async {}

  @override
  Future<void> publish(String topic, String payload) async {
    final request = json.decode(payload) as Map<String, dynamic>;
    publishedRequests.add(request);

    final replyTo = topic.split('/').last;
    final responseTopic = '$TRANSACTIONS_RESPONSE_TOPIC/$replyTo';
    final handler = _handlers[responseTopic];
    if (handler == null) return;

    final response = _responseBuilder(request);
    final message = MqttPublishMessage()
      ..payload.message.addAll(utf8.encode(json.encode(response)));
    final received = MqttReceivedMessage<MqttMessage>(responseTopic, message);
    handler(_client, received);
  }

  @override
  Future<void> subscribe(String topic, {MessageHandler? handler}) async {
    _handlers[topic] = handler;
  }

  @override
  Future<void> unsubscribe(String topic) async {
    _handlers.remove(topic);
  }

  @override
  Future<dynamic> sendRequest(
    String method,
    dynamic params,
    String replyTo,
  ) async {
    final request =
        json.decode(
              json.encode(
                RequestPayload(method: method, params: params).toJson(),
              ),
            )
            as Map<String, dynamic>;
    publishedRequests.add(request);
    final response = _responseBuilder(request);
    if (response['status'] == RESPONSE_STATUS_ERROR) {
      if (response['message']?.toString().contains('record not found') ==
          true) {
        return 0;
      }
      throw Exception('error in response: ${response['message']}');
    }
    return response['data'];
  }

  @override
  Future<Map<String, dynamic>> submitSignedTransactionV2(
    Map<String, dynamic> signedTransaction,
  ) async {
    _submittedV2 = Map<String, dynamic>.from(signedTransaction);
    final envelope = Map<String, dynamic>.from(
      signedTransaction['data'] as Map,
    );
    final params = Map<String, dynamic>.from(signedTransaction)
      ..['data'] = Map<String, dynamic>.from(envelope['payload'] as Map);
    final request = <String, dynamic>{
      'method': REQUEST_METHOD_SEND_TRANSACTION,
      'params': params,
    };
    publishedRequests.add(request);
    final response = _responseBuilder(request);
    if (response['status'] == RESPONSE_STATUS_ERROR) {
      throw Exception('error in response: ${response['message']}');
    }
    return {
      'execution_output': response['data'],
      'transaction': _v2Status(signedTransaction),
    };
  }

  @override
  Future<Map<String, dynamic>> transactionFinalityV2(
    String transactionHash,
  ) async {
    final signed = _submittedV2;
    if (signed == null || signed['hash'] != transactionHash) {
      throw StateError('transaction was not submitted by this test transport');
    }
    return _v2Status(signed);
  }

  @override
  Future<Map<String, dynamic>> executionLogsV2(
    Map<String, String> query,
  ) async => {'logs': <dynamic>[], 'page': 1, 'limit': 100};

  Map<String, dynamic> _v2Status(Map<String, dynamic> signed) => {
    'protocol_version': 2,
    'chain_id': '2finance-test-v2',
    'transaction_hash': signed['hash'],
    'signed_transaction': signed,
    'commit': {
      'sequence': 1,
      'admission_height': 1,
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
