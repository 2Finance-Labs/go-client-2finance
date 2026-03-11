package e2e_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gitlab.com/2finance/2finance-network/blockchain/contract/reviewV1"
	reviewV1Domain "gitlab.com/2finance/2finance-network/blockchain/contract/reviewV1/domain"
	"gitlab.com/2finance/2finance-network/blockchain/log"
	"gitlab.com/2finance/2finance-network/blockchain/utils"
	reviewV1Models "gitlab.com/2finance/2finance-network/blockchain/contract/reviewV1/models"
)

func TestReviewFlow(t *testing.T) {
	c := setupClient(t)
	reviewer, reviewerPriv := createWallet(t, c)
	reviewee, _ := createWallet(t, c)
	c.SetPrivateKey(reviewerPriv)

	deployedContract, err := c.DeployContract1(reviewV1.REVIEW_CONTRACT_V1)
	if err != nil {
		t.Fatalf("DeployContract: %v", err)
	}
	contractLog, err := utils.UnmarshalLog[log.Log](deployedContract.Logs[0])
	if err != nil {
		t.Fatalf("UnmarshalLog (AddWallet.Logs[0]): %v", err)
	}

	// ------------------
	//   CREATE REVIEW
	// ------------------
	address := contractLog.ContractAddress
	subjectType := "order"
	subjectId := "order-xyz"
	rating := 5
	comment := "Great experience!"
	tags := map[string]string{"quality": "5"}
	mediaHashes := []string{"bafy1"}
	startAt := time.Now().Add(1 * time.Second)
	expiredAt := time.Now().Add(24 * time.Hour)
	hidden := false

	addReview, err := c.AddReview(address, reviewer.PublicKey, reviewee.PublicKey, subjectType, subjectId, rating, comment, tags, mediaHashes, startAt, expiredAt, hidden)
	if err != nil {
		t.Fatalf("AddReview: %v", err)
	}

	unmarshalLogReview, err := utils.UnmarshalLog[log.Log](addReview.Logs[0])
	if err != nil {
		t.Fatalf("UnmarshalLog (AddReview.Logs[0]): %v", err)
	}
	assert.Equal(t, unmarshalLogReview.LogType, reviewV1Domain.REVIEW_ADDED_LOG)

	review, err := utils.UnmarshalEvent[reviewV1Domain.Review](unmarshalLogReview.Event)
	if err != nil {
		t.Fatalf("UnmarshalEvent (AddReview.Logs[0].Event): %v", err)
	}

	assert.Equal(t, address, review.Address)
	assert.Equal(t, reviewer.PublicKey, review.Reviewer)
	assert.Equal(t, reviewee.PublicKey, review.Reviewee)
	assert.Equal(t, subjectType, review.SubjectType)
	assert.Equal(t, subjectId, review.SubjectID)
	assert.Equal(t, rating, review.Rating)
	assert.Equal(t, comment, review.Comment)
	assert.Equal(t, tags["quality"], review.Tags["quality"])
	assert.Equal(t, mediaHashes[0], review.MediaHashes[0])
	assert.WithinDuration(t, startAt, review.StartAt, time.Second)
	assert.WithinDuration(t, expiredAt, review.ExpiredAt, time.Second)
	assert.Equal(t, hidden, review.Hidden)

	getReview, err := c.GetReview(address)
	if err != nil {
		t.Fatalf("GetReview: %v", err)
	}

	var reviewState reviewV1Models.ReviewStateModel
	err = utils.UnmarshalState[reviewV1Models.ReviewStateModel](getReview.States[0].Object, &reviewState)
	if err != nil {
		t.Fatalf("UnmarshalState: %v", err)
	}

	assert.Equal(t, address, reviewState.Address)
	assert.Equal(t, reviewer.PublicKey, reviewState.Reviewer)
	assert.Equal(t, reviewee.PublicKey, reviewState.Reviewee)
	assert.Equal(t, subjectType, reviewState.SubjectType)
	assert.Equal(t, subjectId, reviewState.SubjectID)
	assert.Equal(t, rating, reviewState.Rating)
	assert.Equal(t, comment, reviewState.Comment)
	assert.Equal(t, tags["quality"], reviewState.Tags["quality"])
	assert.Equal(t, mediaHashes[0], reviewState.MediaHashes[0])
	assert.WithinDuration(t, startAt, reviewState.StartAt, time.Second)
	assert.WithinDuration(t, expiredAt, reviewState.ExpiredAt, time.Second)
	assert.Equal(t, hidden, reviewState.Hidden)

	// ------------------
	//   UPDATE REVIEW
	// ------------------
	newStart := time.Now()
	newExp := time.Now().Add(48 * time.Hour)

	updateReview, err := c.UpdateReview(address, subjectType, subjectId, 4, "Updated comment", map[string]string{"quality": "4"}, []string{"bafy2"}, &newStart, &newExp)
	if err != nil {
		t.Fatalf("UpdateReview: %v", err)
	}

	unmarshalLogUpdate, err := utils.UnmarshalLog[log.Log](updateReview.Logs[0])
	if err != nil {
		t.Fatalf("UnmarshalLog (UpdateReview.Logs[0]): %v", err)
	}
	assert.Equal(t, unmarshalLogUpdate.LogType, reviewV1Domain.REVIEW_UPDATED_LOG)

	updatedReview, err := utils.UnmarshalEvent[reviewV1Domain.Review](unmarshalLogUpdate.Event)
	if err != nil {
		t.Fatalf("UnmarshalEvent (UpdateReview.Logs[0].Event): %v", err)
	}

	assert.Equal(t, address, updatedReview.Address)
	assert.Equal(t, subjectType, updatedReview.SubjectType)
	assert.Equal(t, subjectId, updatedReview.SubjectID)
	assert.Equal(t, 4, updatedReview.Rating)
	assert.Equal(t, "Updated comment", updatedReview.Comment)
	assert.Equal(t, "4", updatedReview.Tags["quality"])
	assert.Equal(t, "bafy2", updatedReview.MediaHashes[0])

	getReview2, err := c.GetReview(address)
	if err != nil {
		t.Fatalf("GetReview (after update): %v", err)
	}

	var reviewState2 reviewV1Models.ReviewStateModel
	err = utils.UnmarshalState[reviewV1Models.ReviewStateModel](getReview2.States[0].Object, &reviewState2)
	if err != nil {
		t.Fatalf("UnmarshalState (after update): %v", err)
	}

	assert.Equal(t, address, reviewState2.Address)
	assert.Equal(t, reviewer.PublicKey, reviewState2.Reviewer)
	assert.Equal(t, reviewee.PublicKey, reviewState2.Reviewee)
	assert.Equal(t, subjectType, reviewState2.SubjectType)
	assert.Equal(t, subjectId, reviewState2.SubjectID)
	assert.Equal(t, 4, reviewState2.Rating)
	assert.Equal(t, "Updated comment", reviewState2.Comment)
	assert.Equal(t, "4", reviewState2.Tags["quality"])
	assert.Equal(t, "bafy2", reviewState2.MediaHashes[0])
	assert.WithinDuration(t, newStart, reviewState2.StartAt, time.Second)
	assert.WithinDuration(t, newExp, reviewState2.ExpiredAt, time.Second)

	// ------------------
	//   HIDE REVIEW
	// ------------------
	hideReview, err := c.HideReview(address, true)
	if err != nil {
		t.Fatalf("HideReview: %v", err)
	}

	unmarshalLogHide, err := utils.UnmarshalLog[log.Log](hideReview.Logs[0])
	if err != nil {
		t.Fatalf("UnmarshalLog (HideReview.Logs[0]): %v", err)
	}
	assert.Equal(t, unmarshalLogHide.LogType, reviewV1Domain.REVIEW_HIDDEN_LOG)

	hiddenReview, err := utils.UnmarshalEvent[reviewV1Domain.Review](unmarshalLogHide.Event)
	if err != nil {
		t.Fatalf("UnmarshalEvent (HideReview.Logs[0].Event): %v", err)
	}

	assert.Equal(t, address, hiddenReview.Address)
	assert.Equal(t, true, hiddenReview.Hidden)

	getReview3, err := c.GetReview(address)
	if err != nil {
		t.Fatalf("GetReview (after hide): %v", err)
	}

	var reviewState3 reviewV1Models.ReviewStateModel
	err = utils.UnmarshalState[reviewV1Models.ReviewStateModel](getReview3.States[0].Object, &reviewState3)
	if err != nil {
		t.Fatalf("UnmarshalState (after hide): %v", err)
	}

	assert.Equal(t, address, reviewState3.Address)
	assert.Equal(t, reviewer.PublicKey, reviewState3.Reviewer)
	assert.Equal(t, reviewee.PublicKey, reviewState3.Reviewee)
	assert.Equal(t, subjectType, reviewState3.SubjectType)
	assert.Equal(t, subjectId, reviewState3.SubjectID)
	assert.Equal(t, 4, reviewState3.Rating)
	assert.Equal(t, "Updated comment", reviewState3.Comment)
	assert.Equal(t, "4", reviewState3.Tags["quality"])
	assert.Equal(t, "bafy2", reviewState3.MediaHashes[0])
	assert.Equal(t, true, reviewState3.Hidden)

	// ------------------
	//   HELPFUL VOTE
	// ------------------
	voter, voterPriv := createWallet(t, c)
	c.SetPrivateKey(voterPriv)

	voteHelpful, err := c.VoteHelpful(address, voter.PublicKey, true)
	if err != nil {
		t.Fatalf("VoteHelpful: %v", err)
	}

	unmarshalLogVote, err := utils.UnmarshalLog[log.Log](voteHelpful.Logs[0])
	if err != nil {
		t.Fatalf("UnmarshalLog (VoteHelpful.Logs[0]): %v", err)
	}
	assert.Equal(t, unmarshalLogVote.LogType, reviewV1Domain.REVIEW_HELPFUL_LOG)

	votedReview, err := utils.UnmarshalEvent[reviewV1Domain.Vote](unmarshalLogVote.Event)
	if err != nil {
		t.Fatalf("UnmarshalEvent (VoteHelpful.Logs[0].Event): %v", err)
	}

	assert.Equal(t, address, votedReview.Address)
	assert.Equal(t, voter.PublicKey, votedReview.Voter)
	assert.Equal(t, true, votedReview.Helpful)

	getReview4, err := c.GetReview(address)
	if err != nil {
		t.Fatalf("GetReview (after helpful vote): %v", err)
	}

	var reviewState4 reviewV1Models.ReviewStateModel
	err = utils.UnmarshalState[reviewV1Models.ReviewStateModel](getReview4.States[0].Object, &reviewState4)
	if err != nil {
		t.Fatalf("UnmarshalState (after helpful vote): %v", err)
	}

	assert.Equal(t, address, reviewState4.Address)
	assert.Equal(t, reviewer.PublicKey, reviewState4.Reviewer)
	assert.Equal(t, reviewee.PublicKey, reviewState4.Reviewee)
	assert.Equal(t, subjectType, reviewState4.SubjectType)
	assert.Equal(t, subjectId, reviewState4.SubjectID)
	assert.Equal(t, 4, reviewState4.Rating)
	assert.Equal(t, "Updated comment", reviewState4.Comment)
	assert.Equal(t, "4", reviewState4.Tags["quality"])
	assert.Equal(t, "bafy2", reviewState4.MediaHashes[0])
	assert.Equal(t, true, reviewState4.Hidden)
	assert.Equal(t, true, reviewState4.HelpfulVotes[voter.PublicKey])

	// ------------------
	//   REPORT REVIEW
	// ------------------
	reporter, reporterPriv := createWallet(t, c)
	c.SetPrivateKey(reporterPriv)

	reportReview, err := c.ReportReview(address, reporter.PublicKey, "spam")
	if err != nil {
		t.Fatalf("ReportReview: %v", err)
	}

	unmarshalLogReport, err := utils.UnmarshalLog[log.Log](reportReview.Logs[0])
	if err != nil {
		t.Fatalf("UnmarshalLog (ReportReview.Logs[0]): %v", err)
	}
	assert.Equal(t, unmarshalLogReport.LogType, reviewV1Domain.REVIEW_REPORTED_LOG)

	reportedReview, err := utils.UnmarshalEvent[reviewV1Domain.Report](unmarshalLogReport.Event)
	if err != nil {
		t.Fatalf("UnmarshalEvent (ReportReview.Logs[0].Event): %v", err)
	}

	assert.Equal(t, address, reportedReview.Address)
	assert.Equal(t, reporter.PublicKey, reportedReview.Reporter)
	assert.Equal(t, "spam", reportedReview.Reason)

	getReview5, err := c.GetReview(address)
	if err != nil {
		t.Fatalf("GetReview (after report): %v", err)
	}

	var reviewState5 reviewV1Models.ReviewStateModel
	err = utils.UnmarshalState[reviewV1Models.ReviewStateModel](getReview5.States[0].Object, &reviewState5)
	if err != nil {
		t.Fatalf("UnmarshalState (after report): %v", err)
	}

	assert.Equal(t, address, reviewState5.Address)
	assert.Equal(t, reviewer.PublicKey, reviewState5.Reviewer)
	assert.Equal(t, reviewee.PublicKey, reviewState5.Reviewee)
	assert.Equal(t, subjectType, reviewState5.SubjectType)
	assert.Equal(t, subjectId, reviewState5.SubjectID)
	assert.Equal(t, 4, reviewState5.Rating)
	assert.Equal(t, "Updated comment", reviewState5.Comment)
	assert.Equal(t, "4", reviewState5.Tags["quality"])
	assert.Equal(t, "bafy2", reviewState5.MediaHashes[0])
	assert.Equal(t, true, reviewState5.Hidden)
	assert.NotEmpty(t, reviewState5.Reports)
	assert.Equal(t, reporter.PublicKey, reviewState5.Reports[0].Reporter)
	assert.Equal(t, "spam", reviewState5.Reports[0].Reason)

	// ------------------
	//   MODERATE REVIEW
	// ------------------
	c.SetPrivateKey(reviewerPriv)

	moderateReview, err := c.ModerateReview(address, reviewV1Domain.MODERATE_STATUS_APPROVED, "ok")
	if err != nil {
		t.Fatalf("ModerateReview: %v", err)
	}

	unmarshalLogModerate, err := utils.UnmarshalLog[log.Log](moderateReview.Logs[0])
	if err != nil {
		t.Fatalf("UnmarshalLog (ModerateReview.Logs[0]): %v", err)
	}
	assert.Equal(t, unmarshalLogModerate.LogType, reviewV1Domain.REVIEW_MODERATED_LOG)

	moderatedReview, err := utils.UnmarshalEvent[reviewV1Domain.Moderation](unmarshalLogModerate.Event)
	if err != nil {
		t.Fatalf("UnmarshalEvent (ModerateReview.Logs[0].Event): %v", err)
	}

	assert.Equal(t, address, moderatedReview.Address)
	assert.Equal(t, reviewV1Domain.MODERATE_STATUS_APPROVED, moderatedReview.Action)
	assert.Equal(t, "ok", moderatedReview.Note)

	getReview6, err := c.GetReview(address)
	if err != nil {
		t.Fatalf("GetReview (after moderate): %v", err)
	}

	var reviewState6 reviewV1Models.ReviewStateModel
	err = utils.UnmarshalState[reviewV1Models.ReviewStateModel](getReview6.States[0].Object, &reviewState6)
	if err != nil {
		t.Fatalf("UnmarshalState (after moderate): %v", err)
	}

	assert.Equal(t, address, reviewState6.Address)
	assert.Equal(t, reviewV1Domain.MODERATE_STATUS_APPROVED, reviewState6.ModerationStatus)
	assert.Equal(t, "ok", reviewState6.ModerationNote)
	assert.Equal(t, 4, reviewState6.Rating)
	assert.Equal(t, "Updated comment", reviewState6.Comment)
	assert.Equal(t, "4", reviewState6.Tags["quality"])
	assert.Equal(t, "bafy2", reviewState6.MediaHashes[0])
	assert.Equal(t, true, reviewState6.Hidden)

	// ------------------
	//   LIST REVIEWS
	// ------------------
	listReviews, err := c.ListReviews(reviewer.PublicKey, reviewee.PublicKey, subjectType, subjectId, nil, 0, 5, 1, 10, true)
	if err != nil {
		t.Fatalf("ListReviews: %v", err)
	}

	var reviews []reviewV1Domain.Review
	for _, state := range listReviews.States {
		var r reviewV1Models.ReviewStateModel
		err = utils.UnmarshalState[reviewV1Models.ReviewStateModel](state.Object, &r)
		if err != nil {
			t.Fatalf("UnmarshalState (ListReviews): %v", err)
		}
		reviews = append(reviews, reviewV1Domain.Review{
			Address:     r.Address,
			Reviewer:    r.Reviewer,
			Reviewee:    r.Reviewee,
			SubjectType: r.SubjectType,
			SubjectID:   r.SubjectID,
			Rating:      r.Rating,
			Comment:     r.Comment,
			Tags:        r.Tags,
			MediaHashes: r.MediaHashes,
			StartAt:     r.StartAt,
			ExpiredAt:   r.ExpiredAt,
			Hidden:      r.Hidden,
		})
	}

	assert.Len(t, reviews, 1)
	assert.Equal(t, address, reviews[0].Address)
	assert.Equal(t, reviewer.PublicKey, reviews[0].Reviewer)
	assert.Equal(t, reviewee.PublicKey, reviews[0].Reviewee)
	assert.Equal(t, subjectType, reviews[0].SubjectType)
	assert.Equal(t, subjectId, reviews[0].SubjectID)
	assert.Equal(t, 4, reviews[0].Rating)
	assert.Equal(t, "Updated comment", reviews[0].Comment)
	assert.Equal(t, "4", reviews[0].Tags["quality"])
	assert.Equal(t, "bafy2", reviews[0].MediaHashes[0])
	assert.WithinDuration(t, newStart, reviews[0].StartAt, time.Second)
	assert.WithinDuration(t, newExp, reviews[0].ExpiredAt, time.Second)
	assert.Equal(t, true, reviews[0].Hidden)
}
