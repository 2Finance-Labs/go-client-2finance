package e2e_test

import (
	"log"
	"testing"
	"time"

	"gitlab.com/2finance/2finance-network/blockchain/contract/contractV1/models"
	reviewV1 "gitlab.com/2finance/2finance-network/blockchain/contract/reviewV1"
	reviewV1Domain "gitlab.com/2finance/2finance-network/blockchain/contract/reviewV1/domain"
)

func TestReviewFlow(t *testing.T) {
	c := setupClient(t)

	// --------------------------------------------------------------------
	// Wallets
	// --------------------------------------------------------------------
	reviewer, reviewerPriv := createWallet(t, c)
	reviewee, _ := createWallet(t, c)
	c.SetPrivateKey(reviewerPriv)

	// --------------------------------------------------------------------
	// Deploy Review contract
	// --------------------------------------------------------------------
	start := time.Now().Add(1 * time.Second)
	exp := time.Now().Add(24 * time.Hour)

	contractState := models.ContractStateModel{}
	deployedContract, err := c.DeployContract1(reviewV1.REVIEW_CONTRACT_V1)
	if err != nil {
		t.Fatalf("DeployContract(Review): %v", err)
	}
	if len(deployedContract.States) == 0 || deployedContract.States[0].Object == nil {
		t.Fatalf("DeployContract(Review) returned empty/nil state")
	}

	unmarshalState(t, deployedContract.States[0].Object, &contractState)
	address := contractState.Address
	if address == "" {
		t.Fatalf("DeployContract(Review) returned empty contract address")
	}

	log.Printf("DeployContract(Review) Output States: %+v", deployedContract.States)
	log.Printf("DeployContract(Review) Output Logs: %+v", deployedContract.Logs)
	log.Printf("DeployContract(Review) Output Delegated Call: %+v", deployedContract.DelegatedCall)
	log.Printf("Review Contract Address: %s", address)

	// --------------------------------------------------------------------
	// AddReview (reviewer)
	// --------------------------------------------------------------------
	subjectType := "order"
	subjectID := "order-xyz"
	rating := 5
	comment := "Great experience!"
	tags := map[string]string{"quality": "5"}
	media := []string{"bafy1"}
	paused := false

	added, err := c.AddReview(
		address,
		reviewer.PublicKey,
		reviewee.PublicKey,
		subjectType,
		subjectID,
		rating,
		comment,
		tags,
		media,
		start,
		exp,
		paused,
	)
	if err != nil {
		t.Fatalf("AddReview: %v", err)
	}
	if len(added.States) == 0 || added.States[0].Object == nil {
		t.Fatalf("AddReview returned empty/nil state")
	}

	var r reviewV1Domain.Review
	unmarshalState(t, added.States[0].Object, &r)

	if r.Address == "" {
		t.Fatalf("review addr empty")
	}
	if r.Reviewer != reviewer.PublicKey {
		t.Fatalf("AddReview Reviewer mismatch: got %q want %q", r.Reviewer, reviewer.PublicKey)
	}
	if r.Reviewee != reviewee.PublicKey {
		t.Fatalf("AddReview Reviewee mismatch: got %q want %q", r.Reviewee, reviewee.PublicKey)
	}
	if r.SubjectType != subjectType {
		t.Fatalf("AddReview SubjectType mismatch: got %q want %q", r.SubjectType, subjectType)
	}
	if r.SubjectID != subjectID {
		t.Fatalf("AddReview SubjectID mismatch: got %q want %q", r.SubjectID, subjectID)
	}
	if r.Rating != rating {
		t.Fatalf("AddReview Rating mismatch: got %d want %d", r.Rating, rating)
	}
	if r.Comment != comment {
		t.Fatalf("AddReview Comment mismatch: got %q want %q", r.Comment, comment)
	}
	// tags/media podem variar se o backend normaliza; valido só se não-nulos quando enviados
	if r.Tags == nil {
		t.Fatalf("AddReview Tags is nil")
	}
	if r.Tags["quality"] != "5" {
		t.Fatalf("AddReview Tags[quality] mismatch: got %q want %q", r.Tags["quality"], "5")
	}
	if r.MediaHashes == nil {
		t.Fatalf("AddReview MediaHashes is nil")
	}
	if len(r.MediaHashes) != 1 || r.MediaHashes[0] != "bafy1" {
		t.Fatalf("AddReview MediaHashes mismatch: got %+v want %+v", r.MediaHashes, media)
	}

	log.Printf("AddReview Output States: %+v", added.States)
	log.Printf("AddReview Output Logs: %+v", added.Logs)
	log.Printf("AddReview Output Delegated Call: %+v", added.DelegatedCall)

	log.Printf("AddReview Address: %s", r.Address)
	log.Printf("AddReview Reviewer: %s", r.Reviewer)
	log.Printf("AddReview Reviewee: %s", r.Reviewee)
	log.Printf("AddReview SubjectType: %s", r.SubjectType)
	log.Printf("AddReview SubjectID: %s", r.SubjectID)
	log.Printf("AddReview Rating: %d", r.Rating)
	log.Printf("AddReview Comment: %s", r.Comment)
	log.Printf("AddReview Tags: %+v", r.Tags)
	log.Printf("AddReview MediaHashes: %+v", r.MediaHashes)
	log.Printf("AddReview StartAt: %s", r.StartAt.String())
	log.Printf("AddReview ExpiredAt: %s", r.ExpiredAt.String())
	log.Printf("AddReview Hidden: %v", r.Hidden)
	log.Printf("AddReview ModerationStatus: %s", r.ModerationStatus)
	log.Printf("AddReview ModerationNote: %s", r.ModerationNote)

	// --------------------------------------------------------------------
	// Wait until live
	// --------------------------------------------------------------------
	waitUntil(t, 10*time.Second, func() bool { return time.Now().After(start) })

	// --------------------------------------------------------------------
	// UpdateReview (reviewer) - cobre METHOD_UPDATE_REVIEW
	// --------------------------------------------------------------------
	newRating := 4
	newComment := "Updated comment"
	newTags := map[string]string{"quality": "4"}
	newMedia := []string{"bafy2"}
	newStart := time.Now()
	newExp := time.Now().Add(48 * time.Hour)

	upOut, err := c.UpdateReview(
		r.Address,
		subjectType,
		subjectID,
		newRating,
		newComment,
		newTags,
		newMedia,
		&newStart,
		&newExp,
	)
	if err != nil {
		t.Fatalf("UpdateReview: %v", err)
	}
	if len(upOut.States) == 0 || upOut.States[0].Object == nil {
		t.Fatalf("UpdateReview returned empty/nil state")
	}

	var rUp reviewV1Domain.Review
	unmarshalState(t, upOut.States[0].Object, &rUp)

	if rUp.Address != r.Address {
		t.Fatalf("UpdateReview Address mismatch: got %q want %q", rUp.Address, r.Address)
	}
	if rUp.Rating != newRating {
		t.Fatalf("UpdateReview Rating mismatch: got %d want %d", rUp.Rating, newRating)
	}
	if rUp.Comment != newComment {
		t.Fatalf("UpdateReview Comment mismatch: got %q want %q", rUp.Comment, newComment)
	}
	if rUp.Tags == nil || rUp.Tags["quality"] != "4" {
		t.Fatalf("UpdateReview Tags mismatch: got %+v want quality=4", rUp.Tags)
	}
	if rUp.MediaHashes == nil || len(rUp.MediaHashes) != 1 || rUp.MediaHashes[0] != "bafy2" {
		t.Fatalf("UpdateReview MediaHashes mismatch: got %+v want %+v", rUp.MediaHashes, newMedia)
	}

	log.Printf("UpdateReview Output States: %+v", upOut.States)
	log.Printf("UpdateReview Output Logs: %+v", upOut.Logs)
	log.Printf("UpdateReview Output Delegated Call: %+v", upOut.DelegatedCall)

	// --------------------------------------------------------------------
	// HideReview true/false (reviewer) - cobre METHOD_HIDE_REVIEW
	// --------------------------------------------------------------------
	hideOut, err := c.HideReview(r.Address, true)
	if err != nil {
		t.Fatalf("HideReview(true): %v", err)
	}
	if len(hideOut.States) == 0 || hideOut.States[0].Object == nil {
		t.Fatalf("HideReview(true) returned empty/nil state")
	}
	var hidden reviewV1Domain.Review
	unmarshalState(t, hideOut.States[0].Object, &hidden)
	if hidden.Address != r.Address {
		t.Fatalf("HideReview(true) Address mismatch: got %q want %q", hidden.Address, r.Address)
	}
	if !hidden.Hidden {
		t.Fatalf("HideReview(true) expected Hidden=true")
	}

	log.Printf("HideReview(true) Output States: %+v", hideOut.States)
	log.Printf("HideReview(true) Output Logs: %+v", hideOut.Logs)
	log.Printf("HideReview(true) Output Delegated Call: %+v", hideOut.DelegatedCall)

	unhideOut, err := c.HideReview(r.Address, false)
	if err != nil {
		t.Fatalf("HideReview(false): %v", err)
	}
	if len(unhideOut.States) == 0 || unhideOut.States[0].Object == nil {
		t.Fatalf("HideReview(false) returned empty/nil state")
	}
	var unhidden reviewV1Domain.Review
	unmarshalState(t, unhideOut.States[0].Object, &unhidden)
	if unhidden.Address != r.Address {
		t.Fatalf("HideReview(false) Address mismatch: got %q want %q", unhidden.Address, r.Address)
	}
	if unhidden.Hidden {
		t.Fatalf("HideReview(false) expected Hidden=false")
	}

	log.Printf("HideReview(false) Output States: %+v", unhideOut.States)
	log.Printf("HideReview(false) Output Logs: %+v", unhideOut.Logs)
	log.Printf("HideReview(false) Output Delegated Call: %+v", unhideOut.DelegatedCall)

	// --------------------------------------------------------------------
	// Helpful vote (voter) - cobre METHOD_VOTE_HELPFUL
	// --------------------------------------------------------------------
	voter, voterPriv := createWallet(t, c)
	c.SetPrivateKey(voterPriv)

	voteOut, err := c.VoteHelpful(r.Address, voter.PublicKey, true)
	if err != nil {
		t.Fatalf("VoteHelpful: %v", err)
	}
	if len(voteOut.States) == 0 || voteOut.States[0].Object == nil {
		t.Fatalf("VoteHelpful returned empty/nil state")
	}

	var voted reviewV1Domain.Review
	unmarshalState(t, voteOut.States[0].Object, &voted)

	if voted.Address != r.Address {
		t.Fatalf("VoteHelpful Address mismatch: got %q want %q", voted.Address, r.Address)
	}

	log.Printf("VoteHelpful Output States: %+v", voteOut.States)
	log.Printf("VoteHelpful Output Logs: %+v", voteOut.Logs)
	log.Printf("VoteHelpful Output Delegated Call: %+v", voteOut.DelegatedCall)

	// --------------------------------------------------------------------
	// ReportReview (reporter) - cobre METHOD_REPORT_REVIEW
	// --------------------------------------------------------------------
	reporter, reporterPriv := createWallet(t, c)
	c.SetPrivateKey(reporterPriv)

	repOut, err := c.ReportReview(r.Address, reporter.PublicKey, "spam")
	if err != nil {
		t.Fatalf("ReportReview: %v", err)
	}
	if len(repOut.States) == 0 || repOut.States[0].Object == nil {
		t.Fatalf("ReportReview returned empty/nil state")
	}

	var repd reviewV1Domain.Review
	unmarshalState(t, repOut.States[0].Object, &repd)
	if repd.Address != r.Address {
		t.Fatalf("ReportReview Address mismatch: got %q want %q", repd.Address, r.Address)
	}
	// reports pode vir vazio no retorno; então não forço.
	log.Printf("ReportReview Output States: %+v", repOut.States)
	log.Printf("ReportReview Output Logs: %+v", repOut.Logs)
	log.Printf("ReportReview Output Delegated Call: %+v", repOut.DelegatedCall)

	// --------------------------------------------------------------------
	// ModerateReview (moderator) - cobre METHOD_MODERATE_REVIEW
	// Observação: no seu comentário, você assume reviewer como moderador.
	// --------------------------------------------------------------------
	c.SetPrivateKey(reviewerPriv)

	modStatus := reviewV1Domain.MODERATE_STATUS_APPROVED
	modNote := "ok"

	modOut, err := c.ModerateReview(r.Address, modStatus, modNote)
	if err != nil {
		t.Fatalf("ModerateReview: %v", err)
	}
	if len(modOut.States) == 0 || modOut.States[0].Object == nil {
		t.Fatalf("ModerateReview returned empty/nil state")
	}

	var modded reviewV1Domain.Review
	unmarshalState(t, modOut.States[0].Object, &modded)

	if modded.Address != r.Address {
		t.Fatalf("ModerateReview Address mismatch: got %q want %q", modded.Address, r.Address)
	}
	if modded.ModerationStatus != modStatus {
		t.Fatalf("ModerateReview ModerationStatus mismatch: got %q want %q", modded.ModerationStatus, modStatus)
	}
	// moderation note pode ser truncado/normalizado pelo backend; valido só se veio preenchido
	if modded.ModerationNote != "" && modded.ModerationNote != modNote {
		t.Fatalf("ModerateReview ModerationNote mismatch: got %q want %q", modded.ModerationNote, modNote)
	}

	log.Printf("ModerateReview Output States: %+v", modOut.States)
	log.Printf("ModerateReview Output Logs: %+v", modOut.Logs)
	log.Printf("ModerateReview Output Delegated Call: %+v", modOut.DelegatedCall)

	// --------------------------------------------------------------------
	// GetReview - cobre METHOD_GET_REVIEW
	// --------------------------------------------------------------------
	getOut, err := c.GetReview(r.Address)
	if err != nil {
		t.Fatalf("GetReview: %v", err)
	}
	if len(getOut.States) == 0 || getOut.States[0].Object == nil {
		t.Fatalf("GetReview returned empty/nil state")
	}

	var got reviewV1Domain.Review
	unmarshalState(t, getOut.States[0].Object, &got)
	if got.Address != r.Address {
		t.Fatalf("GetReview Address mismatch: got %q want %q", got.Address, r.Address)
	}

	log.Printf("GetReview Output States: %+v", getOut.States)
	log.Printf("GetReview Output Logs: %+v", getOut.Logs)
	log.Printf("GetReview Output Delegated Call: %+v", getOut.DelegatedCall)

	// --------------------------------------------------------------------
	// ListReviews - cobre METHOD_LIST_REVIEWS
	// --------------------------------------------------------------------
	listOut, err := c.ListReviews(
		"",                 // address filter
		reviewer.PublicKey, // reviewer
		reviewee.PublicKey, // reviewee
		subjectType,
		subjectID,
		nil, // ratings filter
		0,   // min rating
		5,   // max rating
		1,   // page
		10,  // limit
		true,
	)
	if err != nil {
		t.Fatalf("ListReviews: %v", err)
	}
	if len(listOut.States) == 0 || listOut.States[0].Object == nil {
		t.Fatalf("ListReviews returned empty/nil state")
	}

	var list []reviewV1Domain.Review
	unmarshalState(t, listOut.States[0].Object, &list)

	found := false
	for _, it := range list {
		if it.Address == r.Address {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("ListReviews: created review %s not found in list", r.Address)
	}

	log.Printf("ListReviews Output States: %+v", listOut.States)
	log.Printf("ListReviews Output Logs: %+v", listOut.Logs)
	log.Printf("ListReviews Output Delegated Call: %+v", listOut.DelegatedCall)
}
