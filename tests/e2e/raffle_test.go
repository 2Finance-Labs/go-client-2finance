package e2e_test

import (
	"fmt"
	"testing"
	"time"

	"gitlab.com/2finance/2finance-network/blockchain/encryption/seed"

	"gitlab.com/2finance/2finance-network/blockchain/contract/contractV1/models"
	raffleV1 "gitlab.com/2finance/2finance-network/blockchain/contract/raffleV1"
	raffleV1Domain "gitlab.com/2finance/2finance-network/blockchain/contract/raffleV1/domain"
	raffleV1Models "gitlab.com/2finance/2finance-network/blockchain/contract/raffleV1/models"
	tokenV1Domain "gitlab.com/2finance/2finance-network/blockchain/contract/tokenV1/domain"
)

// FAILING TESTS
func TestRaffleFlow(t *testing.T) {
	c := setupClient(t)
	owner, ownerPriv := createWallet(t, c)

	c.SetPrivateKey(ownerPriv)
	dec := 6
	tok := createBasicToken(t, c, owner.PublicKey, dec, false, tokenV1Domain.FUNGIBLE)

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
		bob.PublicKey:         bobPriv,
		alice.PublicKey:       alicePriv,
		robert.PublicKey:      robertPriv,
		alfred.PublicKey:      alfredPriv,
		luiz.PublicKey:        luizPriv,
		jorge.PublicKey:       jorgePriv,
		luigui.PublicKey:      luiguiPriv,
		superman.PublicKey:    supermanPriv,
		spiderman.PublicKey:   spidermanPriv,
		batman.PublicKey:      batmanPriv,
		wonderwoman.PublicKey: wonderwomanPriv,
	}

	mintAmt := amt(100, dec)
	c.SetPrivateKey(ownerPriv)
	if _, err := c.MintToken(tok.Address, bob.PublicKey, mintAmt, dec, tok.TokenType); err != nil {
		t.Fatalf("MintToken: %v", err)
	}
	if _, err := c.MintToken(tok.Address, alice.PublicKey, mintAmt, dec, tok.TokenType); err != nil {
		t.Fatalf("MintToken: %v", err)
	}
	if _, err := c.MintToken(tok.Address, robert.PublicKey, mintAmt, dec, tok.TokenType); err != nil {
		t.Fatalf("MintToken: %v", err)
	}
	if _, err := c.MintToken(tok.Address, alfred.PublicKey, mintAmt, dec, tok.TokenType); err != nil {
		t.Fatalf("MintToken: %v", err)
	}
	if _, err := c.MintToken(tok.Address, luiz.PublicKey, mintAmt, dec, tok.TokenType); err != nil {
		t.Fatalf("MintToken: %v", err)
	}
	if _, err := c.MintToken(tok.Address, jorge.PublicKey, mintAmt, dec, tok.TokenType); err != nil {
		t.Fatalf("MintToken: %v", err)
	}
	if _, err := c.MintToken(tok.Address, luigui.PublicKey, mintAmt, dec, tok.TokenType); err != nil {
		t.Fatalf("MintToken: %v", err)
	}
	if _, err := c.MintToken(tok.Address, superman.PublicKey, mintAmt, dec, tok.TokenType); err != nil {
		t.Fatalf("MintToken: %v", err)
	}
	if _, err := c.MintToken(tok.Address, spiderman.PublicKey, mintAmt, dec, tok.TokenType); err != nil {
		t.Fatalf("MintToken: %v", err)
	}
	if _, err := c.MintToken(tok.Address, batman.PublicKey, mintAmt, dec, tok.TokenType); err != nil {
		t.Fatalf("MintToken: %v", err)
	}
	if _, err := c.MintToken(tok.Address, wonderwoman.PublicKey, mintAmt, dec, tok.TokenType); err != nil {
		t.Fatalf("MintToken: %v", err)
	}

	// merchant runs raffle
	merchant, merchPriv := createWallet(t, c)
	c.SetPrivateKey(merchPriv)

	start := time.Now().Add(2 * time.Second)
	exp := time.Now().Add(24 * time.Hour)
	seedPass := "e2e-seed"
	commit := seed.CommitSeed(seedPass) // store locally for later reveal
	meta := map[string]string{"campaign": "e2e"}

	contractState := models.ContractStateModel{}
	deployedContract, err := c.DeployContract1(raffleV1.RAFFLE_CONTRACT_V1)
	if err != nil {
		t.Fatalf("DeployContract: %v", err)
	}
	unmarshalState(t, deployedContract.States[0].Object, &contractState)
	address := contractState.Address

	added, err := c.AddRaffle(address, merchant.PublicKey, tok.Address, amt(1, dec), 100, 5, start, exp, false, commit, meta)
	if err != nil {
		t.Fatalf("AddRaffle: %v", err)
	}
	var rf raffleV1Domain.Raffle
	unmarshalState(t, added.States[0].Object, &rf)
	if rf.Address == "" {
		t.Fatalf("raffle addr empty")
	}

	// allow raffle to move token if required
	_, _ = c.AllowUsers(tok.Address, map[string]bool{rf.Address: true})

	// update
	newStart := time.Now().Add(1 * time.Hour)
	newExp := time.Now().Add(26 * time.Hour)
	commit2 := seed.CommitSeed(seedPass + "2")
	_, err = c.UpdateRaffle(rf.Address, rf.TokenAddress, amt(2, dec), 150, 10, &newStart, &newExp, commit2, map[string]string{"k": "v"})
	if err != nil {
		t.Fatalf("UpdateRaffle: %v", err)
	}

	// pause/unpause
	_, err = c.PauseRaffle(rf.Address, true)
	if err != nil {
		t.Fatalf("PauseRaffle: %v", err)
	}
	_, err = c.UnpauseRaffle(rf.Address, false)
	if err != nil {
		t.Fatalf("UnpauseRaffle: %v", err)
	}

	// wait until original start and enter
	waitUntil(t, 10*time.Second, func() bool { return time.Now().After(start) })

	c.SetPrivateKey(bobPriv)
	if _, err := c.EnterRaffle(rf.Address, 2, tok.Address, tokenV1Domain.FUNGIBLE, ""); err != nil {
		t.Fatalf("EnterRaffle: %v", err)
	}
	if _, err := c.EnterRaffle(rf.Address, 7, tok.Address, tokenV1Domain.FUNGIBLE, ""); err != nil {
		t.Fatalf("EnterRaffle: %v", err)
	}
	if _, err := c.EnterRaffle(rf.Address, 3, tok.Address, tokenV1Domain.FUNGIBLE, ""); err != nil {
		t.Fatalf("EnterRaffle: %v", err)
	}
	if _, err := c.EnterRaffle(rf.Address, 5, tok.Address, tokenV1Domain.FUNGIBLE, ""); err != nil {
		t.Fatalf("EnterRaffle: %v", err)
	}
	c.SetPrivateKey(alicePriv)
	if _, err := c.EnterRaffle(rf.Address, 5, tok.Address, tokenV1Domain.FUNGIBLE, ""); err != nil {
		t.Fatalf("EnterRaffle warning: %v", err)
	}
	c.SetPrivateKey(robertPriv)
	if _, err := c.EnterRaffle(rf.Address, 11, tok.Address, tokenV1Domain.FUNGIBLE, ""); err != nil {
		t.Fatalf("EnterRaffle: %v", err)
	}
	c.SetPrivateKey(alfredPriv)
	if _, err := c.EnterRaffle(rf.Address, 13, tok.Address, tokenV1Domain.FUNGIBLE, ""); err != nil {
		t.Fatalf("EnterRaffle: %v", err)
	}
	c.SetPrivateKey(luizPriv)
	if _, err := c.EnterRaffle(rf.Address, 17, tok.Address, tokenV1Domain.FUNGIBLE, ""); err != nil {
		t.Fatalf("EnterRaffle: %v", err)
	}
	c.SetPrivateKey(jorgePriv)
	if _, err := c.EnterRaffle(rf.Address, 19, tok.Address, tokenV1Domain.FUNGIBLE, ""); err != nil {
		t.Fatalf("EnterRaffle: %v", err)
	}
	c.SetPrivateKey(luiguiPriv)
	if _, err := c.EnterRaffle(rf.Address, 23, tok.Address, tokenV1Domain.FUNGIBLE, ""); err != nil {
		t.Fatalf("EnterRaffle: %v", err)
	}
	c.SetPrivateKey(supermanPriv)
	if _, err := c.EnterRaffle(rf.Address, 29, tok.Address, tokenV1Domain.FUNGIBLE, ""); err != nil {
		t.Fatalf("EnterRaffle: %v", err)
	}
	if _, err := c.EnterRaffle(rf.Address, 31, tok.Address, tokenV1Domain.FUNGIBLE, ""); err != nil {
		t.Fatalf("EnterRaffle: %v", err)
	}
	c.SetPrivateKey(spidermanPriv)
	if _, err := c.EnterRaffle(rf.Address, 37, tok.Address, tokenV1Domain.FUNGIBLE, ""); err != nil {
		t.Fatalf("EnterRaffle: %v", err)
	}
	c.SetPrivateKey(batmanPriv)
	if _, err := c.EnterRaffle(rf.Address, 41, tok.Address, tokenV1Domain.FUNGIBLE, ""); err != nil {
		t.Fatalf("EnterRaffle: %v", err)
	}
	c.SetPrivateKey(wonderwomanPriv)
	if _, err := c.EnterRaffle(rf.Address, 43, tok.Address, tokenV1Domain.FUNGIBLE, ""); err != nil {
		t.Fatalf("EnterRaffle: %v", err)
	}

	ownerPrize, ownerPrizePriv := createWallet(t, c)
	c.SetPrivateKey(ownerPrizePriv)

	tok1 := createBasicToken(t, c, ownerPrize.PublicKey, 6, false, tokenV1Domain.FUNGIBLE)
	tok2 := createBasicToken(t, c, ownerPrize.PublicKey, 6, false, tokenV1Domain.FUNGIBLE)
	tok3 := createBasicToken(t, c, ownerPrize.PublicKey, 6, false, tokenV1Domain.FUNGIBLE)
	tok4 := createBasicToken(t, c, ownerPrize.PublicKey, 6, false, tokenV1Domain.FUNGIBLE)

	// deposit prize & prize ops
	c.SetPrivateKey(ownerPrizePriv)
	output, err := c.AddRafflePrize(rf.Address, tok1.Address, amt(2, dec), tokenV1Domain.FUNGIBLE, "")
	if err != nil {
		t.Fatalf("AddRafflePrize: %v", err)
	}

	output, err = c.AddRafflePrize(rf.Address, tok2.Address, amt(3, dec), tokenV1Domain.FUNGIBLE, "")
	if err != nil {
		t.Fatalf("AddRafflePrize: %v", err)
	}

	output, err = c.AddRafflePrize(rf.Address, tok3.Address, amt(4, dec), tokenV1Domain.FUNGIBLE, "")
	if err != nil {
		t.Fatalf("AddRafflePrize: %v", err)
	}

	output, err = c.AddRafflePrize(rf.Address, tok4.Address, amt(5, dec), tokenV1Domain.FUNGIBLE, "")
	if err != nil {
		t.Fatalf("AddRafflePrize: %v", err)
	}

	var r raffleV1Domain.RafflePrize
	unmarshalState(t, output.States[0].Object, &r)

	fmt.Println("r.UUID:", r.UUID)
	_, err = c.RemoveRafflePrize(rf.Address, tokenV1Domain.FUNGIBLE,r.UUID)
	if err != nil {
		t.Fatalf("RemoveRafflePrize: %v", err)
	}

	c.SetPrivateKey(merchPriv)
	//TODO NEEDS TO MAKE THIS DETERMINISTIC
	draw, err := c.DrawRaffle(rf.Address, seedPass+"2")
	if err != nil {
		t.Fatalf("DrawRaffle warning: %v", err)
	}
	// fmt.Printf("DrawRaffle output: %+v\n", draw)
	var d []raffleV1Models.RafflePrizeModel
	unmarshalState(t, draw.States[0].Object, &d)

	listPrizes, err := c.ListPrizes(rf.Address, 1, 10, true)
	if err != nil {
		t.Fatalf("ListPrizes: %v", err)
	}

	unmarshalState(t, listPrizes.States[0].Object, &d)
	for index, prize := range d {
		fmt.Printf("Prize: %+v\n", prize)
		if prize.Winner != "" {
			c.SetPrivateKey(mapOfPubPriv[prize.Winner])
			claim, err := c.ClaimRaffle(rf.Address, prize.Winner, tokenV1Domain.FUNGIBLE, "")
			if err != nil {
				t.Fatalf("ClaimRaffle warning: %v", err)
			}
			fmt.Printf("ClaimRaffle output: %+v\n", claim)
		}
		if len(d) == index+1 {
			c.SetPrivateKey(mapOfPubPriv[prize.Winner])
			_, err := c.ClaimRaffle(rf.Address, prize.Winner, tokenV1Domain.FUNGIBLE, "")
			if err == nil {
				t.Fatalf("A error must not be nil: %v", err)
			}
		}
	}

	listPrizesClaimed, err := c.ListPrizes(rf.Address, 1, 10, true)
	if err != nil {
		t.Fatalf("ListPrizes: %v", err)
	}

	unmarshalState(t, listPrizesClaimed.States[0].Object, &d)
	for _, prize := range d {
		fmt.Printf("Prize: %+v\n", prize)
		if !prize.Claimed {
			t.Fatalf("Prize must be already claimed: %+v\n", prize)
		}
	}

	// // withdraw leftovers
	c.SetPrivateKey(merchPriv)
	_, err = c.WithdrawRaffle(rf.Address, tok.Address, amt(1, dec), tokenV1Domain.FUNGIBLE, "")
	if err != nil {
		t.Fatalf("WithdrawRaffle: %v", err)
	}

	// // getters
	// if _, err := c.GetRaffle(rf.Address); err != nil { t.Fatalf("GetRaffle: %v", err) }
	// if _, err := c.ListRaffles(merchant.PublicKey, tok.Address, nil, nil, 1, 10, true); err != nil { t.Fatalf("ListRaffles: %v", err) }
}

func TestRaffleFlow_NonFungible(t *testing.T) {
	c := setupClient(t)

	// --------------------------------------------------------------------
	// Owner + NFT base token (ticket)
	// --------------------------------------------------------------------
	owner, ownerPriv := createWallet(t, c)
	c.SetPrivateKey(ownerPriv)

	dec := 0
	tokenType := tokenV1Domain.NON_FUNGIBLE
	tok := createBasicToken(t, c, owner.PublicKey, dec, false, tokenType)

	// --------------------------------------------------------------------
	// Players
	// --------------------------------------------------------------------
	type player struct {
		pub  string
		priv string
		uuid string
	}

	bob, bobPriv := createWallet(t, c)
	alice, alicePriv := createWallet(t, c)

	players := []player{
		{bob.PublicKey, bobPriv, ""},
		{alice.PublicKey, alicePriv, ""},
	}

	// --------------------------------------------------------------------
	// Mint 1 NFT ticket per player
	// --------------------------------------------------------------------
	c.SetPrivateKey(ownerPriv)
	for i := range players {
		mintOut, err := c.MintToken(
			tok.Address,
			players[i].pub,
			"1",
			dec,
			tokenType,
		)
		if err != nil {
			t.Fatalf("MintToken NFT: %v", err)
		}

		var mint tokenV1Domain.Mint
		unmarshalState(t, mintOut.States[0].Object, &mint)
		players[i].uuid = mint.TokenUUIDList[0]
	}

	// --------------------------------------------------------------------
	// Merchant runs raffle
	// --------------------------------------------------------------------
	merchant, merchPriv := createWallet(t, c)
	c.SetPrivateKey(merchPriv)

	start := time.Now().Add(2 * time.Second)
	exp := time.Now().Add(24 * time.Hour)

	seedPass := "e2e-seed-nft"
	commit := seed.CommitSeed(seedPass)
	meta := map[string]string{"campaign": "e2e-nft"}

	var contractState models.ContractStateModel
	deployedContract, err := c.DeployContract1(raffleV1.RAFFLE_CONTRACT_V1)
	if err != nil {
		t.Fatalf("DeployContract: %v", err)
	}
	unmarshalState(t, deployedContract.States[0].Object, &contractState)

	added, err := c.AddRaffle(
		contractState.Address,
		merchant.PublicKey,
		tok.Address,
		"1", // ticket price = 1 NFT
		10,
		1,
		start,
		exp,
		false,
		commit,
		meta,
	)
	if err != nil {
		t.Fatalf("AddRaffle NFT: %v", err)
	}

	var rf raffleV1Domain.Raffle
	unmarshalState(t, added.States[0].Object, &rf)

	// allow raffle
	_, _ = c.AllowUsers(tok.Address, map[string]bool{rf.Address: true})

	// --------------------------------------------------------------------
	// Players enter raffle (UUID obrigatório)
	// --------------------------------------------------------------------
	waitUntil(t, 10*time.Second, func() bool { return time.Now().After(start) })

	for _, p := range players {
		c.SetPrivateKey(p.priv)
		if _, err := c.EnterRaffle(
			rf.Address,
			1,
			tok.Address,
			tokenType,
			p.uuid,
		); err != nil {
			t.Fatalf("EnterRaffle NFT: %v", err)
		}
	}

	// --------------------------------------------------------------------
	// CREATE PRIZES
	// --------------------------------------------------------------------
	prizeOwner, prizePriv := createWallet(t, c)
	c.SetPrivateKey(prizePriv)

	prizeToken := createBasicToken(t, c, prizeOwner.PublicKey, 0, false, tokenType)

	mintPrize, err := c.MintToken(
		prizeToken.Address,
		prizeOwner.PublicKey,
		"1",
		0,
		tokenType,
	)
	if err != nil {
		t.Fatalf("Mint prize NFT: %v", err)
	}

	var prizeMint tokenV1Domain.Mint
	unmarshalState(t, mintPrize.States[0].Object, &prizeMint)
	prizeUUID := prizeMint.TokenUUIDList[0]

	if _, err := c.AddRafflePrize(
		rf.Address,
		prizeToken.Address,
		"1",
		tokenType,
		prizeUUID,
	); err != nil {
		t.Fatalf("AddRafflePrize NFT: %v", err)
	}

	// --------------------------------------------------------------------
	// DRAW (agora winnerCount > 0)
	// --------------------------------------------------------------------
	c.SetPrivateKey(merchPriv)
	draw, err := c.DrawRaffle(rf.Address, seedPass)
	if err != nil {
		t.Fatalf("DrawRaffle NFT: %v", err)
	}

	var prizes []raffleV1Models.RafflePrizeModel
	unmarshalState(t, draw.States[0].Object, &prizes)

	// --------------------------------------------------------------------
	// Claim prize
	// --------------------------------------------------------------------
	for _, pz := range prizes {
		if pz.Winner != "" {
			for _, p := range players {
				if p.pub == pz.Winner {
					c.SetPrivateKey(p.priv)
					if _, err := c.ClaimRaffle(
						rf.Address,
						pz.Winner,
						tokenType,
						pz.UUID,
					); err != nil {
						t.Fatalf("ClaimRaffle NFT: %v", err)
					}
				}
			}
		}
	}
}
