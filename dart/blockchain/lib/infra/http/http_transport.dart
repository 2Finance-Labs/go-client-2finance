import 'dart:convert';

import 'package:http/http.dart' as http;
import 'package:two_finance_blockchain/infra/event/request_response.dart';
import 'package:two_finance_blockchain/infra/transport/transport.dart';

typedef TokenProvider = Future<String?> Function();

class FinanceNetworkHttpException implements Exception {
  const FinanceNetworkHttpException({
    required this.statusCode,
    required this.responseBody,
  });

  final int statusCode;
  final String responseBody;

  @override
  String toString() => '2finance-network HTTP $statusCode: $responseBody';
}

class HttpFinanceNetworkTransport
    implements FinanceNetworkTransport, ProtocolV2FinanceNetworkTransport {
  HttpFinanceNetworkTransport({
    String? baseUrl,
    http.Client? httpClient,
    TokenProvider? tokenProvider,
  }) : baseUrl = (baseUrl ?? 'http://127.0.0.1:9095').replaceFirst(
         RegExp(r'/$'),
         '',
       ),
       _httpClient = httpClient ?? http.Client(),
       _tokenProvider = tokenProvider;

  final String baseUrl;
  final http.Client _httpClient;
  final TokenProvider? _tokenProvider;

  @override
  Future<dynamic> sendRequest(
    String method,
    dynamic params,
    String replyTo,
  ) async {
    return _request(
      httpMethod: 'POST',
      uri: Uri.parse('$baseUrl/v2/2finance-network/query'),
      body: RequestPayload(method: method, params: params).toJson(),
    );
  }

  @override
  Future<Map<String, dynamic>> submitSignedTransactionV2(
    Map<String, dynamic> signedTransaction,
  ) async {
    const attempts = 6;
    for (var attempt = 1; attempt <= attempts; attempt++) {
      try {
        final data = await _request(
          httpMethod: 'POST',
          uri: Uri.parse('$baseUrl/v2/2finance-network/transactions'),
          body: signedTransaction,
        );
        return _object(data, 'protocol v2 submission data');
      } on FinanceNetworkHttpException catch (error) {
        if (!_isTransientProtocolV2Conflict(error) || attempt == attempts) {
          rethrow;
        }
        await Future<void>.delayed(
          Duration(milliseconds: 50 * (1 << (attempt - 1))),
        );
      }
    }
    throw StateError('protocol v2 transaction submission attempts exhausted');
  }

  @override
  Future<Map<String, dynamic>> transactionFinalityV2(
    String transactionHash,
  ) async {
    final hash = transactionHash.trim().toLowerCase();
    if (!RegExp(r'^[0-9a-f]{64}$').hasMatch(hash)) {
      throw ArgumentError(
        'transaction hash must be 32-byte hexadecimal',
        'transactionHash',
      );
    }
    final data = await _request(
      httpMethod: 'GET',
      uri: Uri.parse(
        '$baseUrl/v2/2finance-network/transactions/$hash/finality',
      ),
    );
    return _object(data, 'protocol v2 finality data');
  }

  @override
  Future<Map<String, dynamic>> executionLogsV2(
    Map<String, String> query,
  ) async {
    final filtered = Map<String, String>.fromEntries(
      query.entries.where((entry) => entry.value.trim().isNotEmpty),
    );
    final data = await _request(
      httpMethod: 'GET',
      uri: Uri.parse(
        '$baseUrl/v2/2finance-network/logs',
      ).replace(queryParameters: filtered.isEmpty ? null : filtered),
    );
    return _object(data, 'protocol v2 execution logs');
  }

  Future<dynamic> _request({
    required String httpMethod,
    required Uri uri,
    Object? body,
  }) async {
    final headers = <String, String>{
      'Accept': 'application/json',
      if (body != null) 'Content-Type': 'application/json',
    };
    final accessToken = await _tokenProvider?.call();
    if (accessToken != null && accessToken.trim().isNotEmpty) {
      headers['Authorization'] = _bearer(accessToken);
    }

    late final http.Response response;
    switch (httpMethod) {
      case 'POST':
        response = await _httpClient.post(
          uri,
          headers: headers,
          body: body == null ? null : jsonEncode(body),
        );
        break;
      case 'GET':
        response = await _httpClient.get(uri, headers: headers);
        break;
      default:
        throw ArgumentError.value(httpMethod, 'httpMethod');
    }

    final responseBody = response.body;
    if (response.statusCode < 200 || response.statusCode >= 300) {
      throw FinanceNetworkHttpException(
        statusCode: response.statusCode,
        responseBody: responseBody,
      );
    }

    final decoded = jsonDecode(responseBody);
    if (decoded is! Map<String, dynamic>) {
      throw const FormatException(
        '2finance-network response is not an object.',
      );
    }

    final code = decoded['code'] ?? decoded['Code'];
    if (code is num && code != 200) {
      final msg = decoded['msg'] ?? decoded['Msg'] ?? 'request failed';
      final data = decoded['data'] ?? decoded['Data'];
      throw Exception('2finance-network error $code: $msg $data');
    }

    return decoded['data'] ?? decoded['Data'];
  }

  Map<String, dynamic> _object(dynamic value, String label) {
    if (value is! Map) {
      throw FormatException('$label is not an object.');
    }
    return Map<String, dynamic>.from(value);
  }

  String _bearer(String accessToken) {
    final trimmed = accessToken.trim();
    if (trimmed.toLowerCase().startsWith('bearer ')) {
      return trimmed;
    }
    return 'Bearer $trimmed';
  }

  bool _isTransientProtocolV2Conflict(FinanceNetworkHttpException error) {
    if (error.statusCode != 422) {
      return false;
    }
    final message = error.responseBody.toLowerCase();
    return message.contains("can't serialize access") ||
        message.contains('try restarting transaction') ||
        message.contains('transaction is killed') ||
        message.contains('lock wait timeout') ||
        message.contains('context deadline exceeded');
  }
}
