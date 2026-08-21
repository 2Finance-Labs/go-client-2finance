class FeeTier {
  final String min_amount;
  final String max_amount;
  final String min_volume;
  final String max_volume;
  final int fee_bps;

  const FeeTier({
    required this.min_amount,
    required this.max_amount,
    required this.min_volume,
    required this.max_volume,
    required this.fee_bps,
  });

  factory FeeTier.fromJson(Map<String, dynamic> json) {
    return FeeTier(
      min_amount: json['min_amount']?.toString() ?? '',
      max_amount: json['max_amount']?.toString() ?? '',
      min_volume: json['min_volume']?.toString() ?? '',
      max_volume: json['max_volume']?.toString() ?? '',
      fee_bps: json['fee_bps'] is int
          ? json['fee_bps'] as int
          : int.parse(json['fee_bps'].toString()),
    );
  }

  Map<String, dynamic> toJson() => {
    'min_amount': min_amount,
    'max_amount': max_amount,
    'min_volume': min_volume,
    'max_volume': max_volume,
    'fee_bps': fee_bps,
  };

  @override
  bool operator ==(Object other) =>
      identical(this, other) ||
      other is FeeTier &&
          other.min_amount == min_amount &&
          other.max_amount == max_amount &&
          other.min_volume == min_volume &&
          other.max_volume == max_volume &&
          other.fee_bps == fee_bps;

  @override
  int get hashCode =>
      Object.hash(min_amount, max_amount, min_volume, max_volume, fee_bps);

  @override
  String toString() => 'FeeTier(${toJson()})';
}
