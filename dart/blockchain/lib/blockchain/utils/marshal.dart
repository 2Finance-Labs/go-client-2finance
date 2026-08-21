import 'dart:convert';
import 'dart:typed_data';
import 'package:two_finance_blockchain/blockchain/log/log.dart';

T unmarshalState<T>(Object? obj, T Function(Map<String, dynamic>) fromJson) {
  if (obj == null) {
    throw Exception('unmarshalState: object is null');
  }
  // Map<String, dynamic>
  if (obj is Map<String, dynamic>) {
    return fromJson(obj);
  }

  // Map<dynamic, dynamic> -> Map<String, dynamic>
  if (obj is Map) {
    final mapped = <String, dynamic>{};
    obj.forEach((k, v) {
      mapped[k.toString()] = v;
    });
    return fromJson(mapped);
  }

  // JSON string
  if (obj is String) {
    final s = obj.trim();
    if (s.isEmpty) {
      throw Exception('unmarshalState: empty JSON string');
    }
    final decoded = jsonDecode(s);
    if (decoded is Map<String, dynamic>) return fromJson(decoded);
    if (decoded is Map) {
      final mapped = <String, dynamic>{};
      decoded.forEach((k, v) => mapped[k.toString()] = v);
      return fromJson(mapped);
    }
    throw Exception(
      'unmarshalState: JSON did not decode to object: ${decoded.runtimeType}',
    );
  }

  // UTF-8 JSON bytes
  if (obj is List<int>) {
    final decoded = jsonDecode(utf8.decode(obj));
    if (decoded is Map<String, dynamic>) return fromJson(decoded);
    if (decoded is Map) {
      final mapped = <String, dynamic>{};
      decoded.forEach((k, v) => mapped[k.toString()] = v);
      return fromJson(mapped);
    }
    throw Exception(
      'unmarshalState: bytes JSON did not decode to object: ${decoded.runtimeType}',
    );
  }

  throw Exception(
    'unmarshalState: unsupported state object type: ${obj.runtimeType}',
  );
}

T unmarshalLog<T>(dynamic obj, T Function(Map<String, dynamic>) fromJson) {
  try {
    if (obj is Map<String, dynamic>) {
      return fromJson(obj);
    }

    if (obj is Map) {
      final mapped = <String, dynamic>{};
      obj.forEach((k, v) {
        mapped[k.toString()] = v;
      });
      return fromJson(mapped);
    }

    final String jsonString = jsonEncode(obj);
    final Map<String, dynamic> map = jsonDecode(jsonString);

    return fromJson(map);
  } catch (e) {
    throw Exception("marshal/unmarshal log: $e");
  }
}

T unmarshalEvent<T>(String event, T Function(Map<String, dynamic>) fromJson) {
  if (event.isEmpty) {
    throw Exception("empty event");
  }

  try {
    final bytes = base64Decode(event);
    final jsonString = utf8.decode(bytes);
    final Map<String, dynamic> map = jsonDecode(jsonString);

    return fromJson(map);
  } catch (e) {
    throw Exception("unmarshal event: $e");
  }
}
