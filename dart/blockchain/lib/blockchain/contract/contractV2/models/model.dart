import 'package:json_annotation/json_annotation.dart';

class ContractState {
  @JsonKey(name: 'address')
  final String address;

  @JsonKey(name: 'contract_version')
  final String contractVersion;

  @JsonKey(name: 'created_at')
  final DateTime createdAt;

  static const String tableName = "contract_state_v2";

  ContractState({
    required this.address,
    required this.contractVersion,
    required this.createdAt,
  });

  /// Cria uma instância a partir de um JSON.
  factory ContractState.fromJson(Map<String, dynamic> json) {
    return ContractState(
      address: json['address'] as String,
      contractVersion: json['contract_version'] as String,
      createdAt: DateTime.parse(json['created_at'] as String),
    );
  }

  /// Converte a instância para JSON.
  Map<String, dynamic> toJson() {
    return {
      'address': address,
      'contract_version': contractVersion,
      'created_at': createdAt.toIso8601String(),
    };
  }

  // Métodos de acesso opcionais (similar ao estilo Go)
  String getAddress() => address;
  String getContractVersion() => contractVersion;
  DateTime getCreatedAt() => createdAt;

  @override
  String toString() {
    return 'ContractState(address: $address, contractVersion: $contractVersion, createdAt: $createdAt)';
  }
}
