class Burn {
  final String tokenAddress;
  final String burnFrom;
  final String amount;
  final String tokenType;
  final String uuid;

  Burn({
    required this.tokenAddress,
    required this.burnFrom,
    required this.amount,
    required this.tokenType,
    required this.uuid,
  });

  factory Burn.fromJson(Map<String, dynamic> json) {
    return Burn(
      tokenAddress: json['token_address'] as String,
      burnFrom: json['burn_from'] as String,
      amount: json['amount'] as String,
      tokenType: json['token_type'] as String,
      uuid: json['uuid'] as String,
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'token_address': tokenAddress,
      'burn_from': burnFrom,
      'amount': amount,
      'token_type': tokenType,
      'uuid': uuid,
    };
  }

  @override
  String toString() =>
      'Burn(tokenAddress: $tokenAddress, burnFrom: $burnFrom, amount: $amount, tokenType: $tokenType, uuid: $uuid)';
}
