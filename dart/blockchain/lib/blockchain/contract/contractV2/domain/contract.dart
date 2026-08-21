class Contract {
  final String address;
  final String contractVersion;

  Contract({required this.address, required this.contractVersion});

  factory Contract.fromJson(Map<String, dynamic> json) {
    return Contract(
      address: json['address'] as String,
      contractVersion: json['contract_version'] as String,
    );
  }
}
