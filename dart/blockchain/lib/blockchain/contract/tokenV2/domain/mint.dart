class Mint {
  final String tokenAddress;
  final String mintTo;
  final String amount;
  final String tokenType;
  final List<String> tokenUUIDList;

  Mint({
    required this.tokenAddress,
    required this.mintTo,
    required this.amount,
    required this.tokenType,
    required this.tokenUUIDList,
  });

  factory Mint.fromJson(Map<String, dynamic> json) {
    return Mint(
      tokenAddress: json['token_address'] as String,
      mintTo: json['mint_to'] as String,
      amount: json['amount'] as String,
      tokenType: json['token_type'] as String,
      tokenUUIDList: (json['token_uuid_list'] as List<dynamic>)
          .map((e) => e as String)
          .toList(),
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'token_address': tokenAddress,
      'mint_to': mintTo,
      'amount': amount,
      'token_type': tokenType,
      'token_uuid_list': tokenUUIDList,
    };
  }

  @override
  String toString() =>
      'Mint(tokenAddress: $tokenAddress, mintTo: $mintTo, amount: $amount, tokenType: $tokenType, tokenUUIDList: $tokenUUIDList)';
}
