abstract class FinanceNetworkTransport {
  Future<dynamic> sendRequest(String method, dynamic params, String replyTo);
}

/// Optional capability implemented by transports that speak the commit-first
/// protocol v2 REST API.
abstract class ProtocolV2FinanceNetworkTransport {
  Future<Map<String, dynamic>> submitSignedTransactionV2(
    Map<String, dynamic> signedTransaction,
  );

  Future<Map<String, dynamic>> transactionFinalityV2(String transactionHash);

  Future<Map<String, dynamic>> executionLogsV2(Map<String, String> query);
}
