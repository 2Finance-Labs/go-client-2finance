import 'dart:convert';

import 'package:test/test.dart';

import 'package:two_finance_blockchain/blockchain/types/types.dart';
import 'package:two_finance_blockchain/blockchain/contract/reviewV2/constants.dart';
import 'package:two_finance_blockchain/blockchain/utils/marshal.dart';
import 'package:two_finance_blockchain/two_finance_blockchain.dart';
import '../../../helpers/helpers.dart' hide unmarshalEvent;

typedef JsonMap = Map<String, dynamic>;

Future<JsonMap> _getReviewState(TwoFinanceBlockchain c, String address) async {
  final out = await c.getReview(address: address);
  expect(out.states, isNotNull);
  expect(out.states!, isNotEmpty);

  return unmarshalState(
    out.states!.first.object,
    (json) => Map<String, dynamic>.from(json as Map),
  );
}

void _expectDateClose(
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

void _expectReviewSnapshot(
  JsonMap review, {
  required String address,
  required String reviewer,
  required String reviewee,
  required String subjectType,
  required String subjectID,
  required int rating,
  required String comment,
  required String tagQuality,
  required String mediaHash,
  required bool hidden,
  required String moderationStatus,
  required String moderationNote,
  DateTime? startAt,
  DateTime? expiredAt,
}) {
  expect(review['address'], equals(address));
  expect(review['reviewer'], equals(reviewer));
  expect(review['reviewee'], equals(reviewee));
  expect(review['subject_type'], equals(subjectType));
  expect(review['subject_id'], equals(subjectID));
  expect(review['rating'], equals(rating));
  expect(review['comment'], equals(comment));
  expect(
    Map<String, dynamic>.from(review['tags'] as Map)['quality'],
    equals(tagQuality),
  );
  expect((review['media_hashes'] as List).first, equals(mediaHash));
  expect(review['hidden'], equals(hidden));
  expect(review['moderation_status'], equals(moderationStatus));
  expect(review['moderation_note'], equals(moderationNote));

  if (startAt != null) {
    _expectDateClose(review['start_at'] as String, startAt);
  }

  if (expiredAt != null) {
    _expectDateClose(review['expired_at'] as String, expiredAt);
  }
}

void main() {
  group('ReviewV2 E2E', () {
    test('E2E: deploy review contract + addReview + getReview', () async {
      final c = await setupClient();

      final reviewerKp = await validKeyPair();
      final reviewer = reviewerKp.publicKey;
      final revieweeKp = await validKeyPair();
      final reviewee = revieweeKp.publicKey;

      await c.setPrivateKey(reviewerKp.privateKey);

      final deployed = await c.deployContract1(REVIEW_CONTRACT_V2);
      expect(deployed.logs, isNotNull);
      expect(deployed.logs!, isNotEmpty);

      final reviewAddress = deployed.logs!.first.contractAddress;
      expect(reviewAddress, isNotEmpty);

      final subjectType = 'order';
      final subjectID = 'order-xyz';
      final rating = 5;
      final comment = 'Great experience!';
      final tags = <String, String>{'quality': '5'};
      final mediaHashes = <String>['bafy1'];
      final startAt = DateTime.now().toUtc().add(const Duration(seconds: 1));
      final expiredAt = DateTime.now().toUtc().add(const Duration(hours: 24));
      const hidden = false;

      final outAdd = await c.addReview(
        address: reviewAddress,
        reviewer: reviewer,
        reviewee: reviewee,
        subjectType: subjectType,
        subjectID: subjectID,
        rating: rating,
        comment: comment,
        tags: tags,
        mediaHashes: mediaHashes,
        startAt: startAt,
        expiredAt: expiredAt,
        hidden: hidden,
      );

      expect(outAdd, isA<ContractOutput>());
      expect(outAdd.logs, isNotNull);
      expect(outAdd.logs!, isNotEmpty);

      final addLog = outAdd.logs!.first;
      expect(addLog.contractAddress, equals(reviewAddress));
      expect(addLog.logType, equals('Review_Added'));

      final addEvent = unmarshalEvent<JsonMap>(
        addLog.event,
        (json) => Map<String, dynamic>.from(json as Map),
      );

      expect(addEvent['address'], equals(reviewAddress));
      expect(addEvent['reviewer'], equals(reviewer));
      expect(addEvent['reviewee'], equals(reviewee));
      expect(addEvent['subject_type'], equals(subjectType));
      expect(addEvent['subject_id'], equals(subjectID));
      expect(addEvent['rating'], equals(rating));
      expect(addEvent['comment'], equals(comment));
      expect(
        Map<String, dynamic>.from(addEvent['tags'] as Map)['quality'],
        equals('5'),
      );
      expect((addEvent['media_hashes'] as List).first, equals('bafy1'));

      expect(addEvent.containsKey('hidden'), isFalse);
      expect(addEvent['moderation_status'], equals('pending'));

      _expectDateClose(addEvent['start_at'] as String, startAt);
      _expectDateClose(addEvent['expired_at'] as String, expiredAt);

      final reviewState = await _getReviewState(c, reviewAddress);

      _expectReviewSnapshot(
        reviewState,
        address: reviewAddress,
        reviewer: reviewer,
        reviewee: reviewee,
        subjectType: subjectType,
        subjectID: subjectID,
        rating: rating,
        comment: comment,
        tagQuality: '5',
        mediaHash: 'bafy1',
        hidden: false,
        moderationStatus: 'pending',
        moderationNote: '',
        startAt: startAt,
        expiredAt: expiredAt,
      );

      expect(reviewState['helpful_votes'], isNull);
      expect(reviewState['reports'], isNull);

      // ------------------
      // UPDATE REVIEW
      // ------------------
      final newStart = DateTime.now().toUtc();
      final newExpiredAt = DateTime.now().toUtc().add(
        const Duration(hours: 48),
      );

      final outUpdate = await c.updateReview(
        address: reviewAddress,
        subjectType: subjectType,
        subjectID: subjectID,
        rating: 4,
        comment: 'Updated comment',
        tags: const <String, String>{'quality': '4'},
        mediaHashes: const <String>['bafy2'],
        startAt: newStart,
        expiredAt: newExpiredAt,
      );

      expect(outUpdate, isA<ContractOutput>());
      expect(outUpdate.logs, isNotNull);
      expect(outUpdate.logs!, isNotEmpty);

      final updateLog = outUpdate.logs!.first;
      expect(updateLog.contractAddress, equals(reviewAddress));
      expect(updateLog.logType, equals('Review_Updated'));

      final updateEvent = unmarshalEvent<JsonMap>(
        updateLog.event,
        (json) => Map<String, dynamic>.from(json as Map),
      );

      expect(updateEvent['address'], equals(reviewAddress));
      expect(updateEvent['subject_type'], equals(subjectType));
      expect(updateEvent['subject_id'], equals(subjectID));
      expect(updateEvent['rating'], equals(4));
      expect(updateEvent['comment'], equals('Updated comment'));
      expect(
        Map<String, dynamic>.from(updateEvent['tags'] as Map)['quality'],
        equals('4'),
      );
      expect((updateEvent['media_hashes'] as List).first, equals('bafy2'));

      final reviewStateAfterUpdate = await _getReviewState(c, reviewAddress);

      _expectReviewSnapshot(
        reviewStateAfterUpdate,
        address: reviewAddress,
        reviewer: reviewer,
        reviewee: reviewee,
        subjectType: subjectType,
        subjectID: subjectID,
        rating: 4,
        comment: 'Updated comment',
        tagQuality: '4',
        mediaHash: 'bafy2',
        hidden: false,
        moderationStatus: 'pending',
        moderationNote: '',
        startAt: newStart,
        expiredAt: newExpiredAt,
      );

      expect(reviewStateAfterUpdate['helpful_votes'], isNull);
      expect(reviewStateAfterUpdate['reports'], isNull);

      // ------------------
      // HIDE REVIEW
      // ------------------
      final outHide = await c.hideReview(reviewAddress, true);

      expect(outHide, isA<ContractOutput>());
      expect(outHide.logs, isNotNull);
      expect(outHide.logs!, isNotEmpty);

      final hideLog = outHide.logs!.first;
      expect(hideLog.contractAddress, equals(reviewAddress));
      expect(hideLog.logType, equals('Review_Hidden'));

      final hideEvent = unmarshalEvent<JsonMap>(
        hideLog.event,
        (json) => Map<String, dynamic>.from(json as Map),
      );

      expect(hideEvent['address'], equals(reviewAddress));
      expect(hideEvent['hidden'], isTrue);

      final reviewStateAfterHide = await _getReviewState(c, reviewAddress);

      _expectReviewSnapshot(
        reviewStateAfterHide,
        address: reviewAddress,
        reviewer: reviewer,
        reviewee: reviewee,
        subjectType: subjectType,
        subjectID: subjectID,
        rating: 4,
        comment: 'Updated comment',
        tagQuality: '4',
        mediaHash: 'bafy2',
        hidden: true,
        moderationStatus: 'pending',
        moderationNote: '',
      );

      // ------------------
      // HELPFUL VOTE
      // ------------------
      final voterKp = await validKeyPair();
      final voter = voterKp.publicKey;

      await c.setPrivateKey(voterKp.privateKey);

      final outVoteHelpful = await c.voteHelpful(reviewAddress, voter, true);

      expect(outVoteHelpful, isA<ContractOutput>());
      expect(outVoteHelpful.logs, isNotNull);
      expect(outVoteHelpful.logs!, isNotEmpty);

      final voteLog = outVoteHelpful.logs!.first;
      expect(voteLog.contractAddress, equals(reviewAddress));
      expect(voteLog.logType, equals('Review_Helpful_Voted'));

      final voteEvent = unmarshalEvent<JsonMap>(
        voteLog.event,
        (json) => Map<String, dynamic>.from(json as Map),
      );

      expect(voteEvent['address'], equals(reviewAddress));
      expect(voteEvent['voter'], equals(voter));
      expect(voteEvent['helpful'], isTrue);

      final reviewStateAfterVote = await _getReviewState(c, reviewAddress);

      _expectReviewSnapshot(
        reviewStateAfterVote,
        address: reviewAddress,
        reviewer: reviewer,
        reviewee: reviewee,
        subjectType: subjectType,
        subjectID: subjectID,
        rating: 4,
        comment: 'Updated comment',
        tagQuality: '4',
        mediaHash: 'bafy2',
        hidden: true,
        moderationStatus: 'pending',
        moderationNote: '',
      );

      final helpfulVotes = reviewStateAfterVote['helpful_votes'];
      expect(helpfulVotes, isNotNull);
      expect(Map<String, dynamic>.from(helpfulVotes as Map)[voter], isTrue);

      // ------------------
      // REPORT REVIEW
      // ------------------
      final reporterKp = await validKeyPair();
      final reporter = reporterKp.publicKey;

      await c.setPrivateKey(reporterKp.privateKey);

      final outReport = await c.reportReview(reviewAddress, reporter, 'spam');

      expect(outReport, isA<ContractOutput>());
      expect(outReport.logs, isNotNull);
      expect(outReport.logs!, isNotEmpty);

      final reportLog = outReport.logs!.first;
      expect(reportLog.contractAddress, equals(reviewAddress));
      expect(reportLog.logType, equals('Review_Reported'));

      final reportEvent = unmarshalEvent<JsonMap>(
        reportLog.event,
        (json) => Map<String, dynamic>.from(json as Map),
      );

      expect(reportEvent['address'], equals(reviewAddress));
      expect(reportEvent['reporter'], equals(reporter));
      expect(reportEvent['reason'], equals('spam'));

      final reviewStateAfterReport = await _getReviewState(c, reviewAddress);

      _expectReviewSnapshot(
        reviewStateAfterReport,
        address: reviewAddress,
        reviewer: reviewer,
        reviewee: reviewee,
        subjectType: subjectType,
        subjectID: subjectID,
        rating: 4,
        comment: 'Updated comment',
        tagQuality: '4',
        mediaHash: 'bafy2',
        hidden: true,
        moderationStatus: 'pending',
        moderationNote: '',
      );

      final reports = reviewStateAfterReport['reports'];
      expect(reports, isNotNull);
      expect(reports, isA<List>());
      expect((reports as List), isNotEmpty);

      final firstReport = Map<String, dynamic>.from(reports.first as Map);
      expect(firstReport['reporter'], equals(reporter));
      expect(firstReport['reason'], equals('spam'));

      // ------------------
      // MODERATE REVIEW
      // ------------------
      await c.setPrivateKey(reviewerKp.privateKey);

      const approvedStatus = 'approved';

      final outModerate = await c.moderateReview(
        reviewAddress,
        approvedStatus,
        'ok',
      );

      expect(outModerate, isA<ContractOutput>());
      expect(outModerate.logs, isNotNull);
      expect(outModerate.logs!, isNotEmpty);

      final moderateLog = outModerate.logs!.first;
      expect(moderateLog.contractAddress, equals(reviewAddress));
      expect(moderateLog.logType, equals('Review_Moderated'));

      final moderateEvent = unmarshalEvent<JsonMap>(
        moderateLog.event,
        (json) => Map<String, dynamic>.from(json as Map),
      );

      expect(moderateEvent['address'], equals(reviewAddress));
      expect(moderateEvent['action'], equals(approvedStatus));
      expect(moderateEvent['note'], equals('ok'));

      final reviewStateAfterModerate = await _getReviewState(c, reviewAddress);

      _expectReviewSnapshot(
        reviewStateAfterModerate,
        address: reviewAddress,
        reviewer: reviewer,
        reviewee: reviewee,
        subjectType: subjectType,
        subjectID: subjectID,
        rating: 4,
        comment: 'Updated comment',
        tagQuality: '4',
        mediaHash: 'bafy2',
        hidden: true,
        moderationStatus: approvedStatus,
        moderationNote: 'ok',
      );

      // ------------------
      // LIST REVIEWS
      // ------------------
      final outListReviews = await c.listReviews(
        reviewer: reviewer,
        reviewee: reviewee,
        subjectType: subjectType,
        subjectID: subjectID,
        minRating: 0,
        maxRating: 5,
        page: 1,
        limit: 10,
        asc: true,
      );

      expect(outListReviews.states, isNotNull);
      expect(outListReviews.states!, isNotEmpty);

      final listedReviews = outListReviews.states!
          .map(
            (state) => unmarshalState(
              state.object,
              (json) => Map<String, dynamic>.from(json as Map),
            ),
          )
          .toList();

      expect(listedReviews.length, equals(1));

      final listedReview = listedReviews.first;

      _expectReviewSnapshot(
        listedReview,
        address: reviewAddress,
        reviewer: reviewer,
        reviewee: reviewee,
        subjectType: subjectType,
        subjectID: subjectID,
        rating: 4,
        comment: 'Updated comment',
        tagQuality: '4',
        mediaHash: 'bafy2',
        hidden: true,
        moderationStatus: approvedStatus,
        moderationNote: 'ok',
        startAt: newStart,
        expiredAt: newExpiredAt,
      );
    });
  });
}
