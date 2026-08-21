import 'dart:convert';

Map<String, dynamic> decodeEvent(String eventBase64) {
  return jsonDecode(utf8.decode(base64Decode(eventBase64)))
      as Map<String, dynamic>;
}

Map<String, dynamic> asMap(dynamic value) {
  if (value == null) return <String, dynamic>{};
  return Map<String, dynamic>.from(value as Map);
}
