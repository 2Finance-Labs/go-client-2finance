import 'dart:io';

/// Loads a CA certificate into a [SecurityContext].
/// Throws if the file is not accessible or invalid.
SecurityContext createSecurityContext(String caCertPath) {
  if (caCertPath.trim().isEmpty) {
    throw Exception('❌ TLS Error: CA certificate path must not be empty');
  }

  final file = File(caCertPath);
  if (!file.existsSync()) {
    throw Exception('❌ TLS Error: CA certificate file not found: $caCertPath');
  }

  final context = SecurityContext();
  try {
    context.setTrustedCertificates(caCertPath);
  } on TlsException catch (e) {
    throw Exception(
      '❌ TLS Error: Failed to load CA certificate from $caCertPath: ${e.message}',
    );
  } catch (e) {
    throw Exception('❌ TLS Error: Unexpected error loading CA cert: $e');
  }

  return context;
}
