import 'package:two_finance_blockchain/blockchain/keys/keys.dart';

class Review {
  final String? address;
  final String reviewer;
  final String reviewee;
  final String subjectType;
  final String subjectID;
  final int rating;
  final String comment;
  final Map<String, String>? tags;
  final List<String>? mediaHashes;
  final DateTime startAt;
  final DateTime expiredAt;
  final bool hidden;

  Review({
    this.address,
    required this.reviewer,
    required this.reviewee,
    required this.subjectType,
    required this.subjectID,
    required this.rating,
    required this.comment,
    this.tags,
    this.mediaHashes,
    required this.startAt,
    required this.expiredAt,
    required this.hidden,
  }) {
    _validateAddress(reviewer, "reviewer");
    _validateAddress(reviewee, "reviewee");
    if (address != null) {
      _validateAddress(address!, "address");
    }
  }

  /// Valida se o endereço é uma chave EdDSA válida
  static void _validateAddress(String addr, String label) {
    try {
      KeyManager.validateEDDSAPublicKeyHex(addr.trim());
    } catch (e) {
      throw ArgumentError("Invalid $label '$addr': $e");
    }
  }

  factory Review.fromJson(Map<String, dynamic> json) {
    final review = Review(
      address: json['address'] as String?,
      reviewer: json['reviewer'] as String,
      reviewee: json['reviewee'] as String,
      subjectType: json['subject_type'] as String,
      subjectID: json['subject_id'] as String,
      rating: json['rating'] as int,
      comment: json['comment'] as String,
      tags: (json['tags'] as Map<String, dynamic>?)?.map(
        (k, v) => MapEntry(k, v as String),
      ),
      mediaHashes: (json['media_hashes'] as List<dynamic>?)
          ?.map((e) => e as String)
          .toList(),
      startAt: DateTime.parse(json['start_at'] as String),
      expiredAt: DateTime.parse(json['expired_at'] as String),
      hidden: json['hidden'] as bool,
    );

    // Revalida os endereços ao criar via fromJson
    if (review.address != null) {
      _validateAddress(review.address!, "address");
    }
    _validateAddress(review.reviewer, "reviewer");
    _validateAddress(review.reviewee, "reviewee");

    return review;
  }

  Map<String, dynamic> toJson() {
    return {
      'address': address,
      'reviewer': reviewer,
      'reviewee': reviewee,
      'subject_type': subjectType,
      'subject_id': subjectID,
      'rating': rating,
      'comment': comment,
      'tags': tags,
      'media_hashes': mediaHashes,
      'start_at': startAt.toIso8601String(),
      'expired_at': expiredAt.toIso8601String(),
      'hidden': hidden,
    };
  }

  @override
  String toString() =>
      'Review(subject: $subjectType/$subjectID, reviewer: $reviewer, rating: $rating, hidden: $hidden)';
}
