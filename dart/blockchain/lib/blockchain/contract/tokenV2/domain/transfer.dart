class Transfer {
  final String tokenAddress;
  final String from;
  final String to;
  final String amount;

  Transfer({
    required this.tokenAddress,
    required this.from,
    required this.to,
    required this.amount,
  });

  factory Transfer.fromJson(Map<String, dynamic> json) {
    return Transfer(
      tokenAddress: json['token_address'] as String,
      from: json['from'] as String,
      to: json['to'] as String,
      amount: json['amount'] as String,
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'token_address': tokenAddress,
      'from': from,
      'to': to,
      'amount': amount,
    };
  }

  @override
  String toString() =>
      'Transfer(tokenAddress: $tokenAddress, from: $from, to: $to, amount: $amount)';
}
