class PaymentState {
  final int? id;
  final String? address;
  final String? owner;
  final String? tokenAddress;
  final String? orderId;
  final String? payer;
  final String? payee;
  final String? amount;
  final String? capturedAmount;
  final String? refundedAmount;
  final String? status;
  final bool? paused;
  final String? hash;
  final DateTime? expiredAt;
  final DateTime? createdAt;
  final DateTime? updatedAt;

  const PaymentState({
    this.id,
    this.address,
    this.owner,
    this.tokenAddress,
    this.orderId,
    this.payer,
    this.payee,
    this.amount,
    this.capturedAmount,
    this.refundedAmount,
    this.status,
    this.paused,
    this.hash,
    this.expiredAt,
    this.createdAt,
    this.updatedAt,
  });

  factory PaymentState.fromJson(Map<String, dynamic> json) {
    int? parseInt(dynamic v) {
      if (v is int) return v;
      if (v is String) return int.tryParse(v);
      return null;
    }

    bool? parseBool(dynamic v) {
      if (v is bool) return v;
      if (v is String) {
        if (v.toLowerCase() == 'true') return true;
        if (v.toLowerCase() == 'false') return false;
      }
      return null;
    }

    DateTime? parseDate(dynamic v) {
      if (v is String) return DateTime.tryParse(v);
      return null;
    }

    return PaymentState(
      id: parseInt(json['id']),
      address: json['address'] as String?,
      owner: json['owner'] as String?,
      tokenAddress: json['token_address'] as String?,
      orderId: json['order_id'] as String?,
      payer: json['payer'] as String?,
      payee: json['payee'] as String?,
      amount: json['amount'] as String?,
      capturedAmount: json['captured_amount'] as String?,
      refundedAmount: json['refunded_amount'] as String?,
      status: json['status'] as String?,
      paused: parseBool(json['paused']),
      hash: json['hash'] as String?,
      expiredAt: parseDate(json['expired_at']),
      createdAt: parseDate(json['created_at']),
      updatedAt: parseDate(json['updated_at']),
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'id': id,
      'address': address,
      'owner': owner,
      'token_address': tokenAddress,
      'order_id': orderId,
      'payer': payer,
      'payee': payee,
      'amount': amount,
      'captured_amount': capturedAmount,
      'refunded_amount': refundedAmount,
      'status': status,
      'paused': paused,
      'hash': hash,
      'expired_at': expiredAt?.toIso8601String(),
      'created_at': createdAt?.toIso8601String(),
      'updated_at': updatedAt?.toIso8601String(),
    };
  }
}
