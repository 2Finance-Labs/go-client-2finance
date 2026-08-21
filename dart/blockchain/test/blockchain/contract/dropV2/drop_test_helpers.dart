import 'dart:convert';

import 'package:test/test.dart';
import 'package:two_finance_blockchain/blockchain/utils/marshal.dart';
import 'package:two_finance_blockchain/two_finance_blockchain.dart';

Future<Map<String, dynamic>> getDropState(
  TwoFinanceBlockchain c,
  String address,
) async {
  final out = await c.getDrop(address: address);
  expect(out.states, isNotNull);
  expect(out.states!, isNotEmpty);

  return unmarshalState(
    out.states!.first.object,
    (json) => Map<String, dynamic>.from(json as Map),
  );
}

List<Map<String, dynamic>> parseDropListState(Object? obj) {
  if (obj == null) {
    throw StateError('drop list state is null');
  }

  dynamic decoded = obj;
  if (decoded is String) {
    decoded = jsonDecode(decoded);
  }

  if (decoded is! List) {
    throw StateError('unexpected drop list state type: ${decoded.runtimeType}');
  }

  return List<dynamic>.from(
    decoded,
  ).map((item) => Map<String, dynamic>.from(item as Map)).toList();
}

void expectDateClose(
  String actualIso,
  DateTime expected, {
  int toleranceSeconds = 2,
}) {
  final actual = DateTime.parse(actualIso).toUtc();
  final expectedUtc = expected.toUtc();

  expect(
    actual.difference(expectedUtc).inSeconds.abs(),
    lessThanOrEqualTo(toleranceSeconds),
  );
}

void expectDropSnapshot(
  Map<String, dynamic> drop, {
  required String address,
  required String owner,
  required String programAddress,
  required String tokenAddress,
  required String title,
  required String description,
  required String shortDescription,
  required String imageUrl,
  required String bannerUrl,
  required Map<String, bool> categories,
  required Map<String, bool> socialRequirements,
  required Map<String, bool> postLinks,
  required String verificationType,
  required DateTime startAt,
  required DateTime expireAt,
  required int requestLimit,
  required String claimAmount,
  required int claimIntervalSeconds,
  required bool paused,
}) {
  expect(drop['address'], equals(address));
  expect(drop['owner'], equals(owner));
  expect(drop['program_address'], equals(programAddress));
  expect(drop['token_address'], equals(tokenAddress));
  expect(drop['title'], equals(title));
  expect(drop['description'], equals(description));
  expect(drop['short_description'], equals(shortDescription));
  expect(drop['image_url'], equals(imageUrl));
  expect(drop['banner_url'], equals(bannerUrl));
  expect(drop['verification_type'], equals(verificationType));
  expect(drop['request_limit'], equals(requestLimit));
  expect(drop['claim_amount'], equals(claimAmount));
  expect(drop['claim_interval_seconds'], equals(claimIntervalSeconds));
  expect(drop['paused'], equals(paused));

  final actualCategories = Map<String, dynamic>.from(
    (drop['categories'] as Map?) ?? const {},
  );
  final actualSocialRequirements = Map<String, dynamic>.from(
    (drop['social_requirements'] as Map?) ?? const {},
  );
  final actualPostLinks = Map<String, dynamic>.from(
    (drop['post_links'] as Map?) ?? const {},
  );

  categories.forEach((key, value) {
    expect(actualCategories[key], equals(value));
  });
  socialRequirements.forEach((key, value) {
    expect(actualSocialRequirements[key], equals(value));
  });
  postLinks.forEach((key, value) {
    expect(actualPostLinks[key], equals(value));
  });

  expectDateClose(drop['start_at'] as String, startAt);
  expectDateClose(drop['expire_at'] as String, expireAt);
}
