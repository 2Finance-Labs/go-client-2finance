part of 'two_finance_blockchain.dart';

extension Review on TwoFinanceBlockchain {
  /// Mock dos métodos sendTransaction e getState
  /// Substitua com sua implementação real

  Future<ContractOutput> addReview({
    required String address,
    required String reviewer,
    required String reviewee,
    required String subjectType,
    required String subjectID,
    required int rating,
    required String comment,
    Map<String, String>? tags,
    List<String>? mediaHashes,
    required DateTime startAt,
    required DateTime expiredAt,
    required bool hidden,
  }) async {
    final from = _publicKeyHex!;
    if (from.isEmpty) {
      throw Exception("from address not set");
    }
    KeyManager.validateEDDSAPublicKeyHex(from);

    if (address.isEmpty) {
      throw Exception("address not set");
    }
    KeyManager.validateEDDSAPublicKeyHex(address);

    if (reviewer.isEmpty) {
      throw Exception("reviewer not set");
    }
    KeyManager.validateEDDSAPublicKeyHex(reviewer);

    if (reviewee.isEmpty) {
      throw Exception("reviewee not set");
    }
    KeyManager.validateEDDSAPublicKeyHex(reviewee);

    if (subjectType.isEmpty) {
      throw Exception("subject_type not set");
    }

    if (subjectID.isEmpty) {
      throw Exception("subject_id not set");
    }

    if (rating < 1 || rating > 5) {
      throw Exception("rating must be between 1 and 5");
    }

    if (startAt.millisecondsSinceEpoch == 0) {
      throw Exception("start_at not set");
    }

    if (expiredAt.millisecondsSinceEpoch == 0) {
      throw Exception("expired_at not set");
    }

    final to = address;
    const method = METHOD_ADD_REVIEW;
    final uuid7 = newUUID7();
    const int version = 1;

    final data = <String, dynamic>{
      "address": address,
      "reviewer": reviewer,
      "reviewee": reviewee,
      "subject_type": subjectType,
      "subject_id": subjectID,
      "rating": rating,
      "comment": comment,
      "tags": tags ?? <String, String>{},
      "media_hashes": mediaHashes ?? <String>[],
      "start_at": startAt.toIso8601String(),
      "expired_at": expiredAt.toIso8601String(),
      "hidden": hidden,
    };

    return signAndSendTransaction(
      chainID: _chainID,
      from: from,
      to: to,
      method: method,
      data: data,
      version: version,
      uuid7: uuid7,
    );
  }

  Future<ContractOutput> getReview({required String address}) async {
    final from = _publicKeyHex!;
    if (from.isEmpty) {
      throw Exception("from address not set");
    }
    KeyManager.validateEDDSAPublicKeyHex(from);

    if (address.isEmpty) {
      throw Exception("review address must be set");
    }
    KeyManager.validateEDDSAPublicKeyHex(address);

    const method = METHOD_GET_REVIEW;

    return getState(to: address, method: method, data: {"address": address});
  }

  Future<ContractOutput> updateReview({
    required String address,
    required String subjectType,
    required String subjectID,
    int rating = 0,
    String? comment,
    Map<String, String>? tags,
    List<String>? mediaHashes,
    DateTime? startAt,
    DateTime? expiredAt,
  }) async {
    if (address.isEmpty) {
      throw Exception("address not set");
    }
    KeyManager.validateEDDSAPublicKeyHex(address);

    final from = _publicKeyHex!;
    if (from.isEmpty) {
      throw Exception("from address not set");
    }
    KeyManager.validateEDDSAPublicKeyHex(from);

    if (subjectType.isEmpty) {
      throw Exception("subject_type not set");
    }

    if (subjectID.isEmpty) {
      throw Exception("subject_id not set");
    }

    if (rating != 0 && (rating < 1 || rating > 5)) {
      throw Exception("rating must be between 1 and 5");
    }

    final to = address;
    const method = METHOD_UPDATE_REVIEW;
    final uuid7 = newUUID7();
    const int version = 1;

    final data = <String, dynamic>{
      "address": address,
      "subject_type": subjectType,
      "subject_id": subjectID,
      "rating": rating,
      "comment": comment ?? "",
      "tags": tags ?? <String, String>{},
      "media_hashes": mediaHashes ?? <String>[],
    };

    if (startAt != null) {
      data["start_at"] = startAt.toIso8601String();
    }

    if (expiredAt != null) {
      data["expired_at"] = expiredAt.toIso8601String();
    }

    return signAndSendTransaction(
      chainID: _chainID,
      from: from,
      to: to,
      method: method,
      data: data,
      version: version,
      uuid7: uuid7,
    );
  }

  Future<ContractOutput> hideReview(String address, bool hidden) async {
    if (address.isEmpty) {
      throw Exception("address not set");
    }
    KeyManager.validateEDDSAPublicKeyHex(address);

    final from = _publicKeyHex!;
    if (from.isEmpty) {
      throw Exception("from address not set");
    }
    KeyManager.validateEDDSAPublicKeyHex(from);

    final to = address;
    const method = METHOD_HIDE_REVIEW;
    final uuid7 = newUUID7();
    const int version = 1;

    final data = <String, dynamic>{"address": address, "hidden": hidden};

    return signAndSendTransaction(
      chainID: _chainID,
      from: from,
      to: to,
      method: method,
      data: data,
      version: version,
      uuid7: uuid7,
    );
  }

  Future<ContractOutput> voteHelpful(
    String address,
    String voter,
    bool isHelpful,
  ) async {
    if (address.isEmpty) {
      throw Exception("address not set");
    }
    KeyManager.validateEDDSAPublicKeyHex(address);

    if (voter.isEmpty) {
      throw Exception("voter not set");
    }
    KeyManager.validateEDDSAPublicKeyHex(voter);

    final from = _publicKeyHex!;
    if (from.isEmpty) {
      throw Exception("from address not set");
    }
    KeyManager.validateEDDSAPublicKeyHex(from);

    final to = address;
    const method = METHOD_VOTE_HELPFUL;
    final uuid7 = newUUID7();
    const int version = 1;

    final data = <String, dynamic>{
      "address": address,
      "voter": voter,
      "is_helpful": isHelpful,
    };

    return signAndSendTransaction(
      chainID: _chainID,
      from: from,
      to: to,
      method: method,
      data: data,
      version: version,
      uuid7: uuid7,
    );
  }

  Future<ContractOutput> reportReview(
    String address,
    String reporter,
    String reason,
  ) async {
    if (address.isEmpty) {
      throw Exception("address not set");
    }
    KeyManager.validateEDDSAPublicKeyHex(address);

    if (reporter.isEmpty) {
      throw Exception("reporter not set");
    }
    KeyManager.validateEDDSAPublicKeyHex(reporter);

    if (reason.isEmpty) {
      throw Exception("reason not set");
    }

    final from = _publicKeyHex!;
    if (from.isEmpty) {
      throw Exception("from address not set");
    }
    KeyManager.validateEDDSAPublicKeyHex(from);

    final to = address;
    const method = METHOD_REPORT_REVIEW;
    final uuid7 = newUUID7();
    const int version = 1;

    final data = <String, dynamic>{
      "address": address,
      "reporter": reporter,
      "reason": reason,
    };

    return signAndSendTransaction(
      chainID: _chainID,
      from: from,
      to: to,
      method: method,
      data: data,
      version: version,
      uuid7: uuid7,
    );
  }

  Future<ContractOutput> moderateReview(
    String address,
    String action,
    String note,
  ) async {
    if (address.isEmpty) {
      throw Exception("address not set");
    }
    KeyManager.validateEDDSAPublicKeyHex(address);

    if (action.isEmpty) {
      throw Exception("action not set");
    }

    final from = _publicKeyHex!;
    if (from.isEmpty) {
      throw Exception("from address not set");
    }
    KeyManager.validateEDDSAPublicKeyHex(from);

    final to = address;
    const method = METHOD_MODERATE_REVIEW;
    final uuid7 = newUUID7();
    const int version = 1;

    final data = <String, dynamic>{
      "address": address,
      "action": action,
      "note": note,
    };

    return signAndSendTransaction(
      chainID: _chainID,
      from: from,
      to: to,
      method: method,
      data: data,
      version: version,
      uuid7: uuid7,
    );
  }

  Future<ContractOutput> listReviews({
    String reviewer = '',
    String reviewee = '',
    String subjectType = '',
    String subjectID = '',
    bool? includeHidden,
    int minRating = 0,
    int maxRating = 0,
    int page = 1,
    int limit = 10,
    bool asc = true,
  }) async {
    final from = _publicKeyHex!;
    if (from.isEmpty) {
      throw Exception('from address not set');
    }

    KeyManager.validateEDDSAPublicKeyHex(from);

    if (reviewer.isNotEmpty) {
      KeyManager.validateEDDSAPublicKeyHex(reviewer);
    }

    if (reviewee.isNotEmpty) {
      KeyManager.validateEDDSAPublicKeyHex(reviewee);
    }

    if (page < 1) {
      throw Exception('page must be greater than 0');
    }

    if (limit < 1) {
      throw Exception('limit must be greater than 0');
    }

    if (minRating < 0 || minRating > 5) {
      throw Exception('min_rating must be between 0 and 5');
    }

    if (maxRating < 0 || maxRating > 5) {
      throw Exception('max_rating must be between 0 and 5');
    }

    if (maxRating != 0 && minRating > maxRating) {
      throw Exception('min_rating cannot be greater than max_rating');
    }

    const method = METHOD_LIST_REVIEWS;
    final data = <String, dynamic>{
      "reviewer": reviewer,
      "reviewee": reviewee,
      "subject_type": subjectType,
      "subject_id": subjectID,
      "min_rating": minRating,
      "max_rating": maxRating,
      "page": page,
      "limit": limit,
      "ascending": asc,
      "contract_version": REVIEW_CONTRACT_V2,
    };

    if (includeHidden != null) {
      data["include_hidden"] = includeHidden;
    }
    return getState(to: '', method: method, data: data);
  }
}
