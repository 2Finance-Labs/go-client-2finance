class BalanceState {
  final int? id;
  final String? tokenAddress;
  final String? ownerAddress;
  final String? amount;

  final String? tokenUuid;
  final String? tokenType;
  final bool? burned;

  final DateTime? createdAt;
  final DateTime? updatedAt;

  const BalanceState({
    this.id,
    this.tokenAddress,
    this.ownerAddress,
    this.amount,
    this.tokenUuid,
    this.tokenType,
    this.burned,
    this.createdAt,
    this.updatedAt,
  });

  factory BalanceState.fromJson(Map<String, dynamic> json) {
    int? _parseInt(dynamic v) {
      if (v is int) return v;
      if (v is String) return int.tryParse(v);
      return null;
    }

    DateTime? _parseDate(dynamic v) {
      if (v is String) return DateTime.tryParse(v);
      return null;
    }

    bool? _parseBool(dynamic v) {
      if (v is bool) return v;
      if (v is String) {
        if (v.toLowerCase() == 'true') return true;
        if (v.toLowerCase() == 'false') return false;
      }
      return null;
    }

    return BalanceState(
      id: _parseInt(json['id']),
      tokenAddress: json['token_address'] as String?,
      ownerAddress: json['owner_address'] as String?,
      amount: json['amount'] as String?,
      tokenUuid: json['token_uuid'] as String?,
      tokenType: json['token_type'] as String?,
      burned: _parseBool(json['burned']),
      createdAt: _parseDate(json['created_at']),
      updatedAt: _parseDate(json['updated_at']),
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'id': id,
      'token_address': tokenAddress,
      'owner_address': ownerAddress,
      'amount': amount,
      'token_uuid': tokenUuid,
      'token_type': tokenType,
      'burned': burned,
      'created_at': createdAt?.toIso8601String(),
      'updated_at': updatedAt?.toIso8601String(),
    };
  }

  @override
  String toString() =>
      'Balance(tokenAddress: $tokenAddress, ownerAddress: $ownerAddress, amount: $amount, tokenUuid: $tokenUuid, tokenType: $tokenType, burned: $burned)';
}
