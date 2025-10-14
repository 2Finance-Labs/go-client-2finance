package e2e_test

import (
	"testing"
	"time"

	reviewV1Domain "gitlab.com/2finance/2finance-network/blockchain/contract/reviewV1/domain"
	reviewV1 "gitlab.com/2finance/2finance-network/blockchain/contract/reviewV1"
	"gitlab.com/2finance/2finance-network/blockchain/contract/contractV1/models"
)



func TestReviewFlow(t *testing.T) {
	c := setupClient(t)
	reviewer, reviewerPriv := createWallet(t, c)
	reviewee, _ := createWallet(t, c)
	c.SetPrivateKey(reviewerPriv)


	start := time.Now().Add(1 * time.Second)
	exp := time.Now().Add(24 * time.Hour)

	contractState := models.ContractStateModel{}
	deployedContract, err := c.DeployContract1(reviewV1.REVIEW_CONTRACT_V1)
	if err != nil { t.Fatalf("DeployContract: %v", err) }
	unmarshalState(t, deployedContract.States[0].Object, &contractState)
	address := contractState.Address


	added, err := c.AddReview(address, reviewer.PublicKey, reviewee.PublicKey, "order", "order-xyz", 5, "Great experience!", map[string]string{"quality":"5"}, []string{"bafy1"}, start, exp, false)
	if err != nil { t.Fatalf("AddReview: %v", err) }
	var r reviewV1Domain.Review
	unmarshalState(t, added.States[0].Object, &r)
	if r.Address == "" { t.Fatalf("review addr empty") }


	// wait until live
	waitUntil(t, 10*time.Second, func() bool { return time.Now().After(start) })


	newStart := time.Now(); newExp := time.Now().Add(48 * time.Hour)
	_, err = c.UpdateReview(r.Address, "order", "order-xyz", 4, "Updated comment", map[string]string{"quality":"4"}, []string{"bafy2"}, &newStart, &newExp)
	if err != nil { t.Fatalf("UpdateReview: %v", err) }
	_, err = c.HideReview(r.Address, true)
	if err != nil { t.Fatalf("HideReview: %v", err) }
	_, err = c.HideReview(r.Address, false)
	if err != nil { t.Fatalf("HideReview: %v", err) }


	// helpful vote by another user
	voter, voterPriv := createWallet(t, c)
	c.SetPrivateKey(voterPriv)
	_, err = c.VoteHelpful(r.Address, voter.PublicKey, true)
	if err != nil { t.Fatalf("VoteHelpful: %v", err) }

	// report & moderate (assume reviewer moderates in this test – adjust to your admin key if needed)
	reporter, reporterPriv := createWallet(t, c)
	c.SetPrivateKey(reporterPriv)
	_, err = c.ReportReview(r.Address, reporter.PublicKey, "spam")
	if err != nil { t.Fatalf("ReportReview: %v", err) }
	c.SetPrivateKey(reviewerPriv)
	_, err = c.ModerateReview(r.Address, reviewV1Domain.MODERATE_STATUS_APPROVED, "ok")
	if err != nil { t.Fatalf("ModerateReview: %v", err) }

	if _, err := c.GetReview(r.Address); err != nil { t.Fatalf("GetReview: %v", err) }
	if _, err := c.ListReviews("", reviewer.PublicKey, reviewee.PublicKey, "order", "order-xyz", nil, 0, 5, 1, 10, true); err != nil { t.Fatalf("ListReviews: %v", err) }
}