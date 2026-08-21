import 'package:two_finance_blockchain/blockchain/keys/keys.dart';

class Raffle {
  final String address;
  final String owner;
  final String tokenAddress;
  final String ticketPrice;
  final int maxEntries;
  final int maxEntriesPerUser;
  final DateTime startAt;
  final DateTime expiredAt;
  final bool paused;
  final String seedCommitHex;
  final String revealSeed;
  final String hash;
  final Map<String, String> metadata;

  Raffle({
    required this.address,
    required this.owner,
    required this.tokenAddress,
    required this.ticketPrice,
    required this.maxEntries,
    required this.maxEntriesPerUser,
    required this.startAt,
    required this.expiredAt,
    this.paused = false,
    this.seedCommitHex = '',
    this.revealSeed = '',
    this.hash = '',
    this.metadata = const <String, String>{},
  }) {
    _validateAddress(address, 'address');
    _validateAddress(owner, 'owner');
    _validateAddress(tokenAddress, 'token_address');
  }

  static void _validateAddress(String addr, String label) {
    try {
      KeyManager.validateEDDSAPublicKeyHex(addr.trim());
    } catch (e) {
      throw ArgumentError("Invalid $label '$addr': $e");
    }
  }

  static int _parseInt(dynamic value, String label) {
    if (value is int) return value;
    if (value is num) return value.toInt();
    if (value is String && value.isNotEmpty) return int.parse(value);

    throw ArgumentError("Invalid $label '$value'");
  }

  static DateTime _parseDateTime(dynamic value, String label) {
    if (value is DateTime) return value;
    if (value is String && value.isNotEmpty) return DateTime.parse(value);

    throw ArgumentError("Invalid $label '$value'");
  }

  static Map<String, String> _parseMetadata(dynamic value) {
    if (value == null) return const <String, String>{};

    if (value is Map) {
      return value.map((key, val) => MapEntry(key.toString(), val.toString()));
    }

    throw ArgumentError("Invalid metadata '$value'");
  }

  factory Raffle.fromJson(Map<String, dynamic> json) {
    return Raffle(
      address: (json['address'] ?? '') as String,
      owner: (json['owner'] ?? '') as String,
      tokenAddress: (json['token_address'] ?? '') as String,
      ticketPrice: (json['ticket_price'] ?? '') as String,
      maxEntries: _parseInt(json['max_entries'], 'max_entries'),
      maxEntriesPerUser: _parseInt(
        json['max_entries_per_user'],
        'max_entries_per_user',
      ),
      startAt: _parseDateTime(json['start_at'], 'start_at'),
      expiredAt: _parseDateTime(json['expired_at'], 'expired_at'),
      paused: json['paused'] as bool? ?? false,
      seedCommitHex: (json['seed_commit_hex'] ?? '') as String,
      revealSeed: (json['reveal_seed'] ?? '') as String,
      hash: (json['hash'] ?? '') as String,
      metadata: _parseMetadata(json['metadata']),
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'address': address,
      'owner': owner,
      'token_address': tokenAddress,
      'ticket_price': ticketPrice,
      'max_entries': maxEntries,
      'max_entries_per_user': maxEntriesPerUser,
      'start_at': startAt.toIso8601String(),
      'expired_at': expiredAt.toIso8601String(),
      'paused': paused,
      'seed_commit_hex': seedCommitHex,
      'reveal_seed': revealSeed,
      'hash': hash,
      'metadata': metadata,
    };
  }

  @override
  String toString() {
    return 'Raffle(address: $address, owner: $owner, tokenAddress: $tokenAddress, paused: $paused)';
  }
}
