//wallet.go
class WalletState {
  final String address;
  final String publicKey;
  final DateTime createdAt;
  final DateTime updatedAt;

  WalletState({
    required this.address,
    required this.publicKey,
    required this.createdAt,
    required this.updatedAt,
  });

  factory WalletState.fromJson(Map<String, dynamic> json) {
    return WalletState(
      address: json['address'] as String,
      publicKey: json['public_key'] as String,
      createdAt: DateTime.parse(json['created_at'] as String),
      updatedAt: DateTime.parse(json['updated_at'] as String),
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'address': address,
      'public_key': publicKey,
      'created_at': createdAt.toIso8601String(),
      'updated_at': updatedAt.toIso8601String(),
    };
  }

  @override
  String toString() =>
      'Wallet(address: $address, publicKey: $publicKey, createdAt: $createdAt, updatedAt: $updatedAt)';
}
