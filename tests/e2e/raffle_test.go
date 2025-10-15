package e2e_test

import (
	"testing"
	"time"
	"fmt"
	"gitlab.com/2finance/2finance-network/blockchain/encryption/seed"

	raffleV1Domain "gitlab.com/2finance/2finance-network/blockchain/contract/raffleV1/domain"
	raffleV1Models "gitlab.com/2finance/2finance-network/blockchain/contract/raffleV1/models"
	raffleV1 "gitlab.com/2finance/2finance-network/blockchain/contract/raffleV1"
	"gitlab.com/2finance/2finance-network/blockchain/contract/contractV1/models"
)

func TestRaffleFlow(t *testing.T) {
	c := setupClient(t)
	owner, ownerPriv := createWallet(t, c)

	c.SetPrivateKey(ownerPriv)
	dec := 6
	tok := createBasicToken(t, c, owner.PublicKey, dec, false)

	bob, bobPriv := createWallet(t, c)
	alice, alicePriv := createWallet(t, c)
	robert, robertPriv := createWallet(t, c)
	alfred, alfredPriv := createWallet(t, c)
	luiz, luizPriv := createWallet(t, c)
	jorge, jorgePriv := createWallet(t, c)
	luigui, luiguiPriv := createWallet(t, c)
	superman, supermanPriv := createWallet(t, c)
	spiderman, spidermanPriv := createWallet(t, c)
	batman, batmanPriv := createWallet(t, c)
	wonderwoman, wonderwomanPriv := createWallet(t, c)

	mapOfPubPriv := map[string]string{
		bob.PublicKey: bobPriv,
		alice.PublicKey: alicePriv,
		robert.PublicKey: robertPriv,
		alfred.PublicKey: alfredPriv,
		luiz.PublicKey: luizPriv,
		jorge.PublicKey: jorgePriv,
		luigui.PublicKey: luiguiPriv,
		superman.PublicKey: supermanPriv,
		spiderman.PublicKey: spidermanPriv,
		batman.PublicKey: batmanPriv,
		wonderwoman.PublicKey: wonderwomanPriv,
	}
	
	mintAmt := amt(100, dec)
	c.SetPrivateKey(ownerPriv)
	if _, err := c.MintToken(tok.Address, bob.PublicKey, mintAmt, dec); err != nil { t.Fatalf("MintToken: %v", err) }
	if _, err := c.MintToken(tok.Address, alice.PublicKey, mintAmt, dec); err != nil { t.Fatalf("MintToken: %v", err) }
	if _, err := c.MintToken(tok.Address, robert.PublicKey, mintAmt, dec); err != nil { t.Fatalf("MintToken: %v", err) }
	if _, err := c.MintToken(tok.Address, alfred.PublicKey, mintAmt, dec); err != nil { t.Fatalf("MintToken: %v", err) }
	if _, err := c.MintToken(tok.Address, luiz.PublicKey, mintAmt, dec); err != nil { t.Fatalf("MintToken: %v", err) }
	if _, err := c.MintToken(tok.Address, jorge.PublicKey, mintAmt, dec); err != nil { t.Fatalf("MintToken: %v", err) }
	if _, err := c.MintToken(tok.Address, luigui.PublicKey, mintAmt, dec); err != nil { t.Fatalf("MintToken: %v", err) }
	if _, err := c.MintToken(tok.Address, superman.PublicKey, mintAmt, dec); err != nil { t.Fatalf("MintToken: %v", err) }
	if _, err := c.MintToken(tok.Address, spiderman.PublicKey, mintAmt, dec); err != nil { t.Fatalf("MintToken: %v", err) }
	if _, err := c.MintToken(tok.Address, batman.PublicKey, mintAmt, dec); err != nil { t.Fatalf("MintToken: %v", err) }
	if _, err := c.MintToken(tok.Address, wonderwoman.PublicKey, mintAmt, dec); err != nil { t.Fatalf("MintToken: %v", err) }


	// merchant runs raffle
	merchant, merchPriv := createWallet(t, c)
	c.SetPrivateKey(merchPriv)

	start := time.Now().Add(2 * time.Second)
	exp := time.Now().Add(24 * time.Hour)
	seedPass := "e2e-seed"
	commit := seed.CommitSeed(seedPass) // store locally for later reveal
	meta := map[string]string{"campaign":"e2e"}

	contractState := models.ContractStateModel{}
	deployedContract, err := c.DeployContract1(raffleV1.RAFFLE_CONTRACT_V1)
	if err != nil { t.Fatalf("DeployContract: %v", err) }
	unmarshalState(t, deployedContract.States[0].Object, &contractState)
	address := contractState.Address

	added, err := c.AddRaffle(address, merchant.PublicKey, tok.Address, amt(1, dec), 100, 5, start, exp, false, commit, meta)
	if err != nil { t.Fatalf("AddRaffle: %v", err) }
	var rf raffleV1Domain.Raffle
	unmarshalState(t, added.States[0].Object, &rf)
	if rf.Address == "" { t.Fatalf("raffle addr empty") }

	// allow raffle to move token if required
	_, _ = c.AllowUsers(tok.Address, map[string]bool{rf.Address: true})

	// update
	newStart := time.Now().Add(1 * time.Hour)
	newExp := time.Now().Add(26 * time.Hour)
	commit2 := seed.CommitSeed(seedPass + "2")
	_, err = c.UpdateRaffle(rf.Address, rf.TokenAddress, amt(2, dec), 150, 10, &newStart, &newExp, commit2, map[string]string{"k":"v"})
	if err != nil { t.Fatalf("UpdateRaffle: %v", err) }

	// pause/unpause
	_, err = c.PauseRaffle(rf.Address, true)
	if err != nil { t.Fatalf("PauseRaffle: %v", err) }
	_, err = c.UnpauseRaffle(rf.Address, false)
	if err != nil { t.Fatalf("UnpauseRaffle: %v", err) }

	// wait until original start and enter
	waitUntil(t, 10*time.Second, func() bool { return time.Now().After(start) })
	
	c.SetPrivateKey(bobPriv)
	if _, err := c.EnterRaffle(rf.Address, 2, tok.Address); err != nil { t.Fatalf("EnterRaffle: %v", err) }
	if _, err := c.EnterRaffle(rf.Address, 7, tok.Address); err != nil { t.Fatalf("EnterRaffle: %v", err) }
	if _, err := c.EnterRaffle(rf.Address, 3, tok.Address); err != nil { t.Fatalf("EnterRaffle: %v", err) }
	if _, err := c.EnterRaffle(rf.Address, 5, tok.Address); err != nil { t.Fatalf("EnterRaffle: %v", err) }
	c.SetPrivateKey(alicePriv)
	if _, err := c.EnterRaffle(rf.Address, 5, tok.Address); err != nil { t.Fatalf("EnterRaffle warning: %v", err) }
	c.SetPrivateKey(robertPriv)
	if _, err := c.EnterRaffle(rf.Address, 11, tok.Address); err != nil { t.Fatalf("EnterRaffle: %v", err) }
	c.SetPrivateKey(alfredPriv)
	if _, err := c.EnterRaffle(rf.Address, 13, tok.Address); err != nil { t.Fatalf("EnterRaffle: %v", err) }
	c.SetPrivateKey(luizPriv)
	if _, err := c.EnterRaffle(rf.Address, 17, tok.Address); err != nil { t.Fatalf("EnterRaffle: %v", err) }
	c.SetPrivateKey(jorgePriv)
	if _, err := c.EnterRaffle(rf.Address, 19, tok.Address); err != nil { t.Fatalf("EnterRaffle: %v", err) }
	c.SetPrivateKey(luiguiPriv)
	if _, err := c.EnterRaffle(rf.Address, 23, tok.Address); err != nil { t.Fatalf("EnterRaffle: %v", err) }
	c.SetPrivateKey(supermanPriv)
	if _, err := c.EnterRaffle(rf.Address, 29, tok.Address); err != nil { t.Fatalf("EnterRaffle: %v", err) }
	if _, err := c.EnterRaffle(rf.Address, 31, tok.Address); err != nil { t.Fatalf("EnterRaffle: %v", err) }
	c.SetPrivateKey(spidermanPriv)
	if _, err := c.EnterRaffle(rf.Address, 37, tok.Address); err != nil { t.Fatalf("EnterRaffle: %v", err) }
	c.SetPrivateKey(batmanPriv)
	if _, err := c.EnterRaffle(rf.Address, 41, tok.Address); err != nil { t.Fatalf("EnterRaffle: %v", err) }
	c.SetPrivateKey(wonderwomanPriv)
	if _, err := c.EnterRaffle(rf.Address, 43, tok.Address); err != nil { t.Fatalf("EnterRaffle: %v", err) }

	ownerPrize, ownerPrizePriv := createWallet(t, c)
	c.SetPrivateKey(ownerPrizePriv)
	dec = 6
	tok1 := createBasicToken(t, c, ownerPrize.PublicKey, dec, false)
	dec = 6
	tok2 := createBasicToken(t, c, ownerPrize.PublicKey, dec, false)
	dec = 6
	tok3 := createBasicToken(t, c, ownerPrize.PublicKey, dec, false)
	dec = 6
	tok4 := createBasicToken(t, c, ownerPrize.PublicKey, dec, false)
	// deposit prize & prize ops
	c.SetPrivateKey(ownerPrizePriv)
	output, err := c.AddRafflePrize(rf.Address, tok1.Address, amt(2, dec))
	if err != nil { t.Fatalf("AddRafflePrize: %v", err) }

	output, err = c.AddRafflePrize(rf.Address, tok2.Address, amt(3, dec))
	if err != nil { t.Fatalf("AddRafflePrize: %v", err) }

	output, err = c.AddRafflePrize(rf.Address, tok3.Address, amt(4, dec))
	if err != nil { t.Fatalf("AddRafflePrize: %v", err) }

	output, err = c.AddRafflePrize(rf.Address, tok4.Address, amt(5, dec))
	if err != nil { t.Fatalf("AddRafflePrize: %v", err) }

	var r raffleV1Domain.RafflePrize
	unmarshalState(t, output.States[0].Object, &r)

	fmt.Println("r.UUID:", r.UUID)
	_, err = c.RemoveRafflePrize(rf.Address, r.UUID)
	if err != nil { t.Fatalf("RemoveRafflePrize: %v", err) }

	c.SetPrivateKey(merchPriv)
	//TODO NEEDS TO MAKE THIS DETERMINISTIC
	draw, err := c.DrawRaffle(rf.Address, seedPass+"2")
	if err != nil { t.Fatalf("DrawRaffle warning: %v", err) }
	// fmt.Printf("DrawRaffle output: %+v\n", draw)
	var d []raffleV1Models.RafflePrizeModel
	unmarshalState(t, draw.States[0].Object, &d)

	listPrizes, err := c.ListPrizes(rf.Address, 1, 10, true)
	if err != nil { t.Fatalf("ListPrizes: %v", err) }

	unmarshalState(t, listPrizes.States[0].Object, &d)
	for index, prize := range d {
		fmt.Printf("Prize: %+v\n", prize)
		if prize.Winner != "" {
			c.SetPrivateKey(mapOfPubPriv[prize.Winner])
			claim, err := c.ClaimRaffle(rf.Address, prize.Winner)
			if err != nil { t.Fatalf("ClaimRaffle warning: %v", err) }
			fmt.Printf("ClaimRaffle output: %+v\n", claim)
		}
		if len(d) == index + 1 {
			c.SetPrivateKey(mapOfPubPriv[prize.Winner])
			_, err := c.ClaimRaffle(rf.Address, prize.Winner)
			if err == nil { t.Fatalf("A error must not be nil: %v", err) }
		}
	}

	listPrizesClaimed, err := c.ListPrizes(rf.Address, 1, 10, true)
	if err != nil { t.Fatalf("ListPrizes: %v", err) }
	
	unmarshalState(t, listPrizesClaimed.States[0].Object, &d)
	for _, prize := range d {
		fmt.Printf("Prize: %+v\n", prize)
		if !prize.Claimed {
			t.Fatalf("Prize must be already claimed: %+v\n", prize)
		}
	}

	// // withdraw leftovers
	c.SetPrivateKey(merchPriv)
	_, err = c.WithdrawRaffle(rf.Address, tok.Address, amt(1, dec))
	if err != nil { t.Fatalf("WithdrawRaffle: %v", err) }

	// // getters
	// if _, err := c.GetRaffle(rf.Address); err != nil { t.Fatalf("GetRaffle: %v", err) }
	// if _, err := c.ListRaffles(merchant.PublicKey, tok.Address, nil, nil, 1, 10, true); err != nil { t.Fatalf("ListRaffles: %v", err) }
}