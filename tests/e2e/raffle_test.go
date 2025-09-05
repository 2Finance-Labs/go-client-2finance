package e2e_test

// import (
// 	"crypto/sha256"
// 	"encoding/hex"
// 	"testing"
// 	"time"

// 	raffleV1Domain "gitlab.com/2finance/2finance-network/blockchain/contract/raffleV1/domain"
// 	raffleV1Inputs "gitlab.com/2finance/2finance-network/blockchain/contract/raffleV1/inputs"
// 	raffleV1Models "gitlab.com/2finance/2finance-network/blockchain/contract/raffleV1/models"
// )
//TODO FIX
// func TestRaffleFlow(t *testing.T) {
// 	c := setupClient(t)
// 	owner, ownerPriv := createWallet(t, c)
// 	c.SetPrivateKey(ownerPriv)
// 	dec := 6
// 	tok := createBasicToken(t, c, owner.PublicKey, dec)

// 	// merchant runs raffle
// 	merchant, merchPriv := createWallet(t, c)
// 	c.SetPrivateKey(merchPriv)

// 	start := time.Now().Add(2 * time.Second)
// 	exp := time.Now().Add(24 * time.Hour)
// 	seed := "e2e-seed"
// 	sum := sha256.Sum256([]byte(seed))
// 	commit := hex.EncodeToString(sum[:])
// 	meta := map[string]string{"campaign":"e2e"}

// 	added, err := c.AddRaffle("", merchant.PublicKey, tok.Address, amt(1, dec), 100, 5, start, exp, false, commit, meta)
// 	if err != nil { t.Fatalf("AddRaffle: %v", err) }
// 	var rf raffleV1Domain.Raffle
// 	unmarshalState(t, added.States[0].Object, &rf)
// 	if rf.Address == "" { t.Fatalf("raffle addr empty") }

// 	// allow raffle to move token if required
// 	_, _ = c.AllowUsers(tok.Address, map[string]bool{rf.Address: true})

// 	// update
// 	newStart := time.Now().Add(1 * time.Hour)
// 	newExp := time.Now().Add(26 * time.Hour)
// 	sum2 := sha256.Sum256([]byte(seed+"2"))
// 	commit2 := hex.EncodeToString(sum2[:])
// 	_, _ = c.UpdateRaffle(rf.Address, rf.TokenAddress, amt(2, dec), 150, 10, &newStart, &newExp, commit2, map[string]string{"k":"v"})

// 	// pause/unpause
// 	_, _ = c.PauseRaffle(rf.Address, true)
// 	_, _ = c.UnpauseRaffle(rf.Address, false)

// 	// wait until original start and enter
// 	waitUntil(t, 15*time.Second, func() bool { return time.Now().After(start) })
// 	user, userPriv := createWallet(t, c)
// 	_ = user
// 	c.SetPrivateKey(userPriv)
// 	if _, err := c.EnterRaffle(rf.Address, 3, tok.Address); err != nil { t.Logf("EnterRaffle warning: %v", err) }

// 	// deposit prize & prize ops
// 	c.SetPrivateKey(merchPriv)
// 	_, _ = c.DepositRaffle(rf.Address, tok.Address, amt(10, dec))
// 	_, _ = c.UpsertRafflePrize(rf.Address, tok.Address, amt(4, dec))
// 	_, _ = c.SetRafflePrizes(rf.Address, []raffleV1Inputs.PrizeInput{{TokenAddress: tok.Address, Amount: amt(2, dec)}})
// 	_, _ = c.RemoveRafflePrize(rf.Address, tok.Address)

// 	// draw & claim best-effort
// 	draw, err := c.DrawRaffle(rf.Address, seed+"2", 2)
// 	if err != nil { t.Logf("DrawRaffle warning: %v", err) }
// 	var d raffleV1Domain.Draw
// 	unmarshalState(t, draw.States[0].Object, &d)
// 	if len(d.Winners) > 0 {
// 		_, _ = c.ClaimRaffle(rf.Address, d.Winners[0])
// 	}

// 	// withdraw leftovers
// 	_, _ = c.WithdrawRaffle(rf.Address, tok.Address, amt(1, dec))

// 	// getters
// 	if _, err := c.GetRaffle(rf.Address); err != nil { t.Fatalf("GetRaffle: %v", err) }
// 	if _, err := c.ListRaffles(merchant.PublicKey, tok.Address, nil, nil, 1, 10, true); err != nil { t.Fatalf("ListRaffles: %v", err) }

// 	// state decode example
// 	stOut, _ := c.GetRaffle(rf.Address)
// 	var st raffleV1Models.RaffleStateModel
// 	unmarshalState(t, stOut.States[0].Object, &st)
// }