class Wallet {
  final String address;
  final String publicKey;

  Wallet({required this.address, required this.publicKey});

  factory Wallet.fromJson(Map<String, dynamic> json) {
    return Wallet(
      address: json['address'] as String,
      publicKey: json['public_key'] as String,
    );
  }

  Map<String, dynamic> toJson() {
    return {'address': address, 'public_key': publicKey};
  }

  @override
  String toString() => 'Wallet(address: $address, publicKey: $publicKey)';
}
