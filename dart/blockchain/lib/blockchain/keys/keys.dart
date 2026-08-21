import 'dart:typed_data';
import 'package:cryptography/cryptography.dart';

/// Representa um par de chaves pública e privada geradas.
class KeyPair2Finance {
  final String publicKey;
  final String privateKey;

  KeyPair2Finance({required this.publicKey, required this.privateKey});

  @override
  String toString() {
    return 'Public Key: $publicKey\nPrivate Key: $privateKey';
  }
}

class KeyManager {
  Future<KeyPair2Finance> generateKeyEd25519() async {
    final algorithm = Ed25519();
    final keyPair = await algorithm.newKeyPair();
    final publicKeyBytes = (await keyPair.extractPublicKey()).bytes;
    final privateKeyBytes = await keyPair.extractPrivateKeyBytes();

    final publicKeyHex = bytesToHex(publicKeyBytes);
    final privateKeyHex = bytesToHex(privateKeyBytes);

    return KeyPair2Finance(publicKey: publicKeyHex, privateKey: privateKeyHex);
  }

  /// Converte uma lista de bytes para uma string hexadecimal.
  static String bytesToHex(List<int> bytes) {
    return bytes.map((byte) => byte.toRadixString(16).padLeft(2, '0')).join();
  }

  /// Converte uma string hexadecimal para uma lista de bytes (Uint8List).
  static Uint8List hexToBytes(String hex) {
    if (hex.length % 2 != 0) {
      throw FormatException(
        'String hexadecimal inválida. O comprimento deve ser par.',
      );
    }
    final bytes = <int>[];
    for (int i = 0; i < hex.length; i += 2) {
      bytes.add(int.parse(hex.substring(i, i + 2), radix: 16));
    }
    return Uint8List.fromList(bytes);
  }

  static void validateEDDSAPublicKeyHex(String publicKeyHex) {
    if (publicKeyHex.isEmpty) {
      throw FormatException('Public key cannot be empty.');
    }
    final bytes = KeyManager.hexToBytes(publicKeyHex);

    // Verifica o tamanho: Ed25519 usa 32 bytes para a chave pública
    if (bytes.length != 32) {
      throw FormatException(
        'Size of public key must be 32 bytes (64 hex characters), received: ${bytes.length} bytes.',
      );
    }

    // Verifica se todos os bytes são zero
    final isAllZero = bytes.every((b) => b == 0);
    if (isAllZero) {
      throw FormatException('Invalid public key: all bytes are zero.');
    }
  }
}
