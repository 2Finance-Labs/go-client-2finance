import 'package:two_finance_blockchain/blockchain/keys/keys.dart';

class CashBack {
  final String owner;
  final String tokenAddress;
  final String programType;
  final String percentage;
  final DateTime startAt;
  final DateTime expiredAt;
  final bool paused;

  CashBack({
    required this.owner,
    required this.tokenAddress,
    required this.programType,
    required this.percentage,
    required this.startAt,
    required this.expiredAt,
    required this.paused,
  }) {
    _validateAddress(owner, "owner");
    _validateAddress(tokenAddress, "token_address");
  }

  /// Valida se o endereço é uma chave EdDSA válida
  static void _validateAddress(String addr, String label) {
    try {
      KeyManager.validateEDDSAPublicKeyHex(addr.trim());
    } catch (e) {
      throw ArgumentError("Invalid $label '$addr': $e");
    }
  }

  factory CashBack.fromJson(Map<String, dynamic> json) {
    final cb = CashBack(
      owner: json['owner'] as String,
      tokenAddress: json['token_address'] as String,
      programType: json['program_type'] as String,
      percentage: json['percentage'] as String,
      startAt: DateTime.parse(json['start_at'] as String),
      expiredAt: DateTime.parse(json['expired_at'] as String),
      paused: json['paused'] as bool,
    );

    // Revalida endereços quando vem da API
    _validateAddress(cb.owner, "owner");
    _validateAddress(cb.tokenAddress, "token_address");

    return cb;
  }

  Map<String, dynamic> toJson() {
    return {
      'owner': owner,
      'token_address': tokenAddress,
      'program_type': programType,
      'percentage': percentage,
      'start_at': startAt.toIso8601String(),
      'expired_at': expiredAt.toIso8601String(),
      'paused': paused,
    };
  }

  @override
  String toString() =>
      'CashBack(type: $programType, token: $tokenAddress, percentage: $percentage, paused: $paused)';
}
