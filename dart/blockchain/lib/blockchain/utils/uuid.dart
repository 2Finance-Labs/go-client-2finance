import 'dart:typed_data';
import 'package:uuid/uuid.dart';

final _uuid = Uuid();

/// Equivalent to:
/// func NewUUID7() (string, error)
String newUUID7() {
  try {
    return _uuid.v7();
  } catch (e) {
    throw Exception("failed to generate UUIDv7: $e");
  }
}

void validateUUID7(String s) {
  late final Uint8List bytes;

  try {
    bytes = Uint8List.fromList(UuidValue.fromString(s).toBytes());
  } catch (e) {
    throw Exception("invalid UUID format: $e");
  }

  final version = (bytes[6] >> 4) & 0x0f;
  if (version != 7) {
    throw Exception("invalid UUID version: $version");
  }

  final variant = (bytes[8] >> 6) & 0x03;
  if (variant != 2) {
    // RFC4122
    throw Exception("invalid UUID variant");
  }
}

void validateUUID7Strict(String s) {
  late final Uint8List bytes;

  try {
    bytes = Uint8List.fromList(UuidValue.fromString(s).toBytes());
  } catch (e) {
    throw Exception("invalid UUID format: $e");
  }

  // Version
  final version = (bytes[6] >> 4) & 0x0f;
  if (version != 7) {
    throw Exception("invalid UUID version: $version");
  }

  // Variant (RFC4122 = 2)
  final variant = (bytes[8] >> 6) & 0x03;
  if (variant != 2) {
    throw Exception("invalid UUID variant");
  }

  // Timestamp (first 48 bits, ms)
  final ms =
      (bytes[0] << 40) |
      (bytes[1] << 32) |
      (bytes[2] << 24) |
      (bytes[3] << 16) |
      (bytes[4] << 8) |
      (bytes[5]);

  final now = DateTime.now().millisecondsSinceEpoch;
  const maxFutureDriftMs = 5 * 60 * 1000;

  if (ms > now + maxFutureDriftMs) {
    throw Exception("uuid timestamp is in the future");
  }
}
