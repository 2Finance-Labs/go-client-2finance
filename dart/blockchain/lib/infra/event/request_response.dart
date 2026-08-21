const String TRANSACTIONS_REQUEST_TOPIC = "request/+";
const String TRANSACTIONS_RESPONSE_TOPIC = "response";

const String RESPONSE_STATUS_SUCCESS = "success";
const String RESPONSE_STATUS_ERROR = "error";

class RequestPayload {
  final String method;
  final dynamic params;

  RequestPayload({required this.method, required this.params});

  Map<String, dynamic> toJson() => {'method': method, 'params': params};
}

class ResponsePayload {
  final String status;
  final String? message;
  final dynamic data;

  ResponsePayload({
    required this.status,
    required this.message,
    required this.data,
  });

  factory ResponsePayload.fromJson(Map<String, dynamic> json) {
    return ResponsePayload(
      status: json['status'],
      message: json['message'],
      data: json['data'],
    );
  }
}
