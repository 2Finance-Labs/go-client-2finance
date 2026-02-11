package e2e_test

import (
	"testing"
	"time"

	"gitlab.com/2finance/2finance-network/blockchain/encryption/seed"

	"gitlab.com/2finance/2finance-network/blockchain/contract/contractV1/models"
	raffleV1 "gitlab.com/2finance/2finance-network/blockchain/contract/raffleV1"
	raffleV1Domain "gitlab.com/2finance/2finance-network/blockchain/contract/raffleV1/domain"
	raffleV1Models "gitlab.com/2finance/2finance-network/blockchain/contract/raffleV1/models"
	tokenV1Domain "gitlab.com/2finance/2finance-network/blockchain/contract/tokenV1/domain"
)

func TestRaffleFlow(t *testing.T) {
	c := setupClient(t)
	owner, ownerPriv := createWallet(t, c)

	c.SetPrivateKey(ownerPriv)
	dec := 6
	stablecoin := false
	tok := createBasicToken(t, c, owner.PublicKey, dec, false, tokenV1Domain.FUNGIBLE, stablecoin)

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

	// owner permite participantes no token base
	c.SetPrivateKey(ownerPriv)
	if _, err := c.AllowUsers(tok.Address, map[string]bool{
		bob.PublicKey:         true,
		alice.PublicKey:       true,
		robert.PublicKey:      true,
		alfred.PublicKey:      true,
		luiz.PublicKey:        true,
		jorge.PublicKey:       true,
		luigui.PublicKey:      true,
		superman.PublicKey:    true,
		spiderman.PublicKey:   true,
		batman.PublicKey:      true,
		wonderwoman.PublicKey: true,
	}); err != nil {
		t.Fatalf("AllowUsers(participants): %v", err)
	}

	mintAmt := amt(100, dec)
	if _, err := c.MintToken(tok.Address, bob.PublicKey, mintAmt, dec, tok.TokenType); err != nil {
		t.Fatalf("MintToken(bob): %v", err)
	}
	if _, err := c.MintToken(tok.Address, alice.PublicKey, mintAmt, dec, tok.TokenType); err != nil {
		t.Fatalf("MintToken(alice): %v", err)
	}
	if _, err := c.MintToken(tok.Address, robert.PublicKey, mintAmt, dec, tok.TokenType); err != nil {
		t.Fatalf("MintToken(robert): %v", err)
	}
	if _, err := c.MintToken(tok.Address, alfred.PublicKey, mintAmt, dec, tok.TokenType); err != nil {
		t.Fatalf("MintToken(alfred): %v", err)
	}
	if _, err := c.MintToken(tok.Address, luiz.PublicKey, mintAmt, dec, tok.TokenType); err != nil {
		t.Fatalf("MintToken(luiz): %v", err)
	}
	if _, err := c.MintToken(tok.Address, jorge.PublicKey, mintAmt, dec, tok.TokenType); err != nil {
		t.Fatalf("MintToken(jorge): %v", err)
	}
	if _, err := c.MintToken(tok.Address, luigui.PublicKey, mintAmt, dec, tok.TokenType); err != nil {
		t.Fatalf("MintToken(luigui): %v", err)
	}
	if _, err := c.MintToken(tok.Address, superman.PublicKey, mintAmt, dec, tok.TokenType); err != nil {
		t.Fatalf("MintToken(superman): %v", err)
	}
	if _, err := c.MintToken(tok.Address, spiderman.PublicKey, mintAmt, dec, tok.TokenType); err != nil {
		t.Fatalf("MintToken(spiderman): %v", err)
	}
	if _, err := c.MintToken(tok.Address, batman.PublicKey, mintAmt, dec, tok.TokenType); err != nil {
		t.Fatalf("MintToken(batman): %v", err)
	}
	if _, err := c.MintToken(tok.Address, wonderwoman.PublicKey, mintAmt, dec, tok.TokenType); err != nil {
		t.Fatalf("MintToken(wonderwoman): %v", err)
	}

	// merchant cria/gerencia raffle
	merchant, merchPriv := createWallet(t, c)

	// owner do token base permite merchant operar (se aplicável)
	c.SetPrivateKey(ownerPriv)
	if _, err := c.AllowUsers(tok.Address, map[string]bool{merchant.PublicKey: true}); err != nil {
		t.Fatalf("AllowUsers(merchant): %v", err)
	}

	c.SetPrivateKey(merchPriv)
	start := time.Now().Add(2 * time.Second)
	exp := time.Now().Add(24 * time.Hour)
	seedPass := "e2e-seed"
	commit := seed.CommitSeed(seedPass)
	meta := map[string]string{"campaign": "e2e"}

	contractState := models.ContractStateModel{}
	deployedContract, err := c.DeployContract1(raffleV1.RAFFLE_CONTRACT_V1)
	if err != nil {
		t.Fatalf("DeployContract: %v", err)
	}
	unmarshalState(t, deployedContract.States[0].Object, &contractState)
	address := contractState.Address

	added, err := c.AddRaffle(
		address,
		merchant.PublicKey,
		tok.Address,
		amt(1, dec),
		100,
		5,
		start,
		exp,
		false,
		commit,
		meta,
	)
	if err != nil {
		t.Fatalf("AddRaffle: %v", err)
	}
	var rf raffleV1Domain.Raffle
	unmarshalState(t, added.States[0].Object, &rf)
	if rf.Address == "" {
		t.Fatalf("raffle addr empty")
	}

	// owner do token base permite o contrato do raffle movimentar o token base (entrada/saída)
	c.SetPrivateKey(ownerPriv)
	if _, err := c.AllowUsers(tok.Address, map[string]bool{rf.Address: true}); err != nil {
		t.Fatalf("AllowUsers(raffle -> tok): %v", err)
	}

	// update raffle (merchant)
	c.SetPrivateKey(merchPriv)
	newStart := time.Now().Add(1 * time.Hour)
	newExp := time.Now().Add(26 * time.Hour)
	commit2 := seed.CommitSeed(seedPass + "2")
	if _, err := c.UpdateRaffle(
		rf.Address,
		rf.TokenAddress,
		amt(2, dec),
		150,
		10,
		&newStart,
		&newExp,
		commit2,
		map[string]string{"k": "v"},
	); err != nil {
		t.Fatalf("UpdateRaffle: %v", err)
	}

	if _, err := c.PauseRaffle(rf.Address, true); err != nil {
		t.Fatalf("PauseRaffle: %v", err)
	}
	if _, err := c.UnpauseRaffle(rf.Address, false); err != nil {
		t.Fatalf("UnpauseRaffle: %v", err)
	}

	// esperar start e entrar
	waitUntil(t, 10*time.Second, func() bool { return time.Now().After(start) })

	c.SetPrivateKey(bobPriv)
	if _, err := c.EnterRaffle(rf.Address, 2, tok.Address, tokenV1Domain.FUNGIBLE, ""); err != nil {
		t.Fatalf("EnterRaffle(bob): %v", err)
	}
	if _, err := c.EnterRaffle(rf.Address, 7, tok.Address, tokenV1Domain.FUNGIBLE, ""); err != nil {
		t.Fatalf("EnterRaffle(bob): %v", err)
	}
	if _, err := c.EnterRaffle(rf.Address, 3, tok.Address, tokenV1Domain.FUNGIBLE, ""); err != nil {
		t.Fatalf("EnterRaffle(bob): %v", err)
	}
	if _, err := c.EnterRaffle(rf.Address, 5, tok.Address, tokenV1Domain.FUNGIBLE, ""); err != nil {
		t.Fatalf("EnterRaffle(bob): %v", err)
	}

	c.SetPrivateKey(alicePriv)
	if _, err := c.EnterRaffle(rf.Address, 5, tok.Address, tokenV1Domain.FUNGIBLE, ""); err != nil {
		t.Fatalf("EnterRaffle(alice): %v", err)
	}

	c.SetPrivateKey(robertPriv)
	if _, err := c.EnterRaffle(rf.Address, 11, tok.Address, tokenV1Domain.FUNGIBLE, ""); err != nil {
		t.Fatalf("EnterRaffle(robert): %v", err)
	}

	c.SetPrivateKey(alfredPriv)
	if _, err := c.EnterRaffle(rf.Address, 13, tok.Address, tokenV1Domain.FUNGIBLE, ""); err != nil {
		t.Fatalf("EnterRaffle(alfred): %v", err)
	}

	c.SetPrivateKey(luizPriv)
	if _, err := c.EnterRaffle(rf.Address, 17, tok.Address, tokenV1Domain.FUNGIBLE, ""); err != nil {
		t.Fatalf("EnterRaffle(luiz): %v", err)
	}

	c.SetPrivateKey(jorgePriv)
	if _, err := c.EnterRaffle(rf.Address, 19, tok.Address, tokenV1Domain.FUNGIBLE, ""); err != nil {
		t.Fatalf("EnterRaffle(jorge): %v", err)
	}

	c.SetPrivateKey(luiguiPriv)
	if _, err := c.EnterRaffle(rf.Address, 23, tok.Address, tokenV1Domain.FUNGIBLE, ""); err != nil {
		t.Fatalf("EnterRaffle(luigui): %v", err)
	}

	c.SetPrivateKey(supermanPriv)
	if _, err := c.EnterRaffle(rf.Address, 29, tok.Address, tokenV1Domain.FUNGIBLE, ""); err != nil {
		t.Fatalf("EnterRaffle(superman): %v", err)
	}
	if _, err := c.EnterRaffle(rf.Address, 31, tok.Address, tokenV1Domain.FUNGIBLE, ""); err != nil {
		t.Fatalf("EnterRaffle(superman): %v", err)
	}

	c.SetPrivateKey(spidermanPriv)
	if _, err := c.EnterRaffle(rf.Address, 37, tok.Address, tokenV1Domain.FUNGIBLE, ""); err != nil {
		t.Fatalf("EnterRaffle(spiderman): %v", err)
	}

	c.SetPrivateKey(batmanPriv)
	if _, err := c.EnterRaffle(rf.Address, 41, tok.Address, tokenV1Domain.FUNGIBLE, ""); err != nil {
		t.Fatalf("EnterRaffle(batman): %v", err)
	}

	c.SetPrivateKey(wonderwomanPriv)
	if _, err := c.EnterRaffle(rf.Address, 43, tok.Address, tokenV1Domain.FUNGIBLE, ""); err != nil {
		t.Fatalf("EnterRaffle(wonderwoman): %v", err)
	}

	// ------------------------------------------------------------
	// PRÊMIOS: ownerPrize cria tokens separados e deposita no raffle
	// ------------------------------------------------------------
	ownerPrize, ownerPrizePriv := createWallet(t, c)
	mapOfPubPriv[ownerPrize.PublicKey] = ownerPrizePriv

	c.SetPrivateKey(ownerPrizePriv)
	tok1 := createBasicToken(t, c, ownerPrize.PublicKey, 6, false, tokenV1Domain.FUNGIBLE, stablecoin)
	tok2 := createBasicToken(t, c, ownerPrize.PublicKey, 6, false, tokenV1Domain.FUNGIBLE, stablecoin)
	tok3 := createBasicToken(t, c, ownerPrize.PublicKey, 6, false, tokenV1Domain.FUNGIBLE, stablecoin)
	tok4 := createBasicToken(t, c, ownerPrize.PublicKey, 6, false, tokenV1Domain.FUNGIBLE, stablecoin)

	if _, err := c.AllowUsers(tok1.Address, map[string]bool{rf.Address: true}); err != nil {
		t.Fatalf("AllowUsers(tok1 -> raffle): %v", err)
	}
	if _, err := c.AllowUsers(tok2.Address, map[string]bool{rf.Address: true}); err != nil {
		t.Fatalf("AllowUsers(tok2 -> raffle): %v", err)
	}
	if _, err := c.AllowUsers(tok3.Address, map[string]bool{rf.Address: true}); err != nil {
		t.Fatalf("AllowUsers(tok3 -> raffle): %v", err)
	}
	if _, err := c.AllowUsers(tok4.Address, map[string]bool{rf.Address: true}); err != nil {
		t.Fatalf("AllowUsers(tok4 -> raffle): %v", err)
	}

	// deposit prizes (ownerPrize assina)
	output, err := c.AddRafflePrize(rf.Address, tok1.Address, amt(2, dec), tokenV1Domain.FUNGIBLE, "")
	if err != nil {
		t.Fatalf("AddRafflePrize(tok1): %v", err)
	}
	output, err = c.AddRafflePrize(rf.Address, tok2.Address, amt(3, dec), tokenV1Domain.FUNGIBLE, "")
	if err != nil {
		t.Fatalf("AddRafflePrize(tok2): %v", err)
	}
	output, err = c.AddRafflePrize(rf.Address, tok3.Address, amt(4, dec), tokenV1Domain.FUNGIBLE, "")
	if err != nil {
		t.Fatalf("AddRafflePrize(tok3): %v", err)
	}
	output, err = c.AddRafflePrize(rf.Address, tok4.Address, amt(5, dec), tokenV1Domain.FUNGIBLE, "")
	if err != nil {
		t.Fatalf("AddRafflePrize(tok4): %v", err)
	}

	var rp raffleV1Domain.RafflePrize
	unmarshalState(t, output.States[0].Object, &rp)

	// remove last prize (ownerPrize assina)
	if _, err := c.RemoveRafflePrize(rf.Address, tokenV1Domain.FUNGIBLE, rp.UUID); err != nil {
		t.Fatalf("RemoveRafflePrize: %v", err)
	}

	// draw (merchant assina)
	c.SetPrivateKey(merchPriv)
	draw, err := c.DrawRaffle(rf.Address, seedPass+"2")
	if err != nil {
		t.Fatalf("DrawRaffle: %v", err)
	}

	var d []raffleV1Models.RafflePrizeModel
	unmarshalState(t, draw.States[0].Object, &d)

	listPrizes, err := c.ListPrizes(rf.Address, 1, 10, true)
	if err != nil {
		t.Fatalf("ListPrizes: %v", err)
	}
	unmarshalState(t, listPrizes.States[0].Object, &d)

	// ------------------------------------------------------------
	// IMPORTANTÍSSIMO (corrige seu erro atual):
	// Winner (assinante do ClaimRaffle) também precisa TER ACESSO ao token do prêmio.
	// Então: antes de cada Claim, o ownerPrize (dono dos tokens de prêmio) faz AllowUsers(prize.TokenAddress, winner).
	// ------------------------------------------------------------
	for index, prize := range d {
		if prize.Winner != "" {
			// 1) owner do token do prêmio (ownerPrize) libera o WINNER no token do prêmio
			c.SetPrivateKey(ownerPrizePriv)
			if _, err := c.AllowUsers(prize.TokenAddress, map[string]bool{prize.Winner: true}); err != nil {
				t.Fatalf("AllowUsers(prizeToken -> winner): token=%s winner=%s err=%v", prize.TokenAddress, prize.Winner, err)
			}

			// 2) winner assina o claim
			priv, ok := mapOfPubPriv[prize.Winner]
			if !ok {
				t.Fatalf("missing private key for winner %s", prize.Winner)
			}
			c.SetPrivateKey(priv)

			if _, err := c.ClaimRaffle(rf.Address, prize.Winner, tokenV1Domain.FUNGIBLE, ""); err != nil {
				t.Fatalf("ClaimRaffle: %v", err)
			}
		}

		// última iteração: claim duplicado (espera erro)
		if len(d) == index+1 && prize.Winner != "" {
			priv, ok := mapOfPubPriv[prize.Winner]
			if !ok {
				t.Fatalf("missing private key for winner %s", prize.Winner)
			}
			c.SetPrivateKey(priv)

			_, err := c.ClaimRaffle(rf.Address, prize.Winner, tokenV1Domain.FUNGIBLE, "")
			if err == nil {
				t.Fatalf("expected error on duplicate claim, got nil")
			}
		}
	}

	listPrizesClaimed, err := c.ListPrizes(rf.Address, 1, 10, true)
	if err != nil {
		t.Fatalf("ListPrizes(claimed): %v", err)
	}

	unmarshalState(t, listPrizesClaimed.States[0].Object, &d)
	for _, prize := range d {
		if !prize.Claimed {
			t.Fatalf("prize must be claimed: %+v", prize)
		}
	}

	// withdraw leftovers (merchant)
	c.SetPrivateKey(merchPriv)
	if _, err := c.WithdrawRaffle(rf.Address, tok.Address, amt(1, dec), tokenV1Domain.FUNGIBLE, ""); err != nil {
		t.Fatalf("WithdrawRaffle: %v", err)
	}
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
	stablecoin := false
	tok := createBasicToken(t, c, owner.PublicKey, dec, false, tokenType, stablecoin)

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

	c.SetPrivateKey(ownerPriv)
	if _, err := c.AllowUsers(tok.Address, map[string]bool{
		bob.PublicKey:   true,
		alice.PublicKey: true,
	}); err != nil {
		t.Fatalf("AllowUsers(players): %v", err)
	}

	// --------------------------------------------------------------------
	// Mint 1 NFT ticket per player
	// --------------------------------------------------------------------
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
		if len(mint.TokenUUIDList) == 0 {
			t.Fatalf("MintToken NFT returned empty uuid list")
		}
		players[i].uuid = mint.TokenUUIDList[0]
	}

	// --------------------------------------------------------------------
	// Merchant runs raffle
	// --------------------------------------------------------------------
	merchant, merchPriv := createWallet(t, c)

	c.SetPrivateKey(ownerPriv)
	if _, err := c.AllowUsers(tok.Address, map[string]bool{merchant.PublicKey: true}); err != nil {
		t.Fatalf("AllowUsers(merchant): %v", err)
	}

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
	if rf.Address == "" {
		t.Fatalf("raffle addr empty")
	}

	c.SetPrivateKey(ownerPriv)
	if _, err := c.AllowUsers(tok.Address, map[string]bool{rf.Address: true}); err != nil {
		t.Fatalf("AllowUsers(raffle -> ticketToken): %v", err)
	}

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
	// CREATE PRIZES (NFT)
	// --------------------------------------------------------------------
	prizeOwner, prizePriv := createWallet(t, c)
	c.SetPrivateKey(prizePriv)

	prizeToken := createBasicToken(t, c, prizeOwner.PublicKey, 0, false, tokenType, stablecoin)

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
	if len(prizeMint.TokenUUIDList) == 0 {
		t.Fatalf("prize mint returned empty uuid list")
	}
	prizeUUID := prizeMint.TokenUUIDList[0]

	if _, err := c.AllowUsers(prizeToken.Address, map[string]bool{rf.Address: true}); err != nil {
		t.Fatalf("AllowUsers(prizeToken -> raffle): %v", err)
	}

	// deposit prize (prizeOwner assina)
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
	// DRAW
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
		if pz.Winner == "" {
			continue
		}

		c.SetPrivateKey(prizePriv)
		if _, err := c.AllowUsers(prizeToken.Address, map[string]bool{pz.Winner: true}); err != nil {
			t.Fatalf("AllowUsers(prizeToken -> winner): %v", err)
		}

		// winner assina o claim
		var winnerPriv string
		for _, p := range players {
			if p.pub == pz.Winner {
				winnerPriv = p.priv
				break
			}
		}
		if winnerPriv == "" {
			t.Fatalf("winner priv not found for %s", pz.Winner)
		}

		c.SetPrivateKey(winnerPriv)
		if _, err := c.ClaimRaffle(
			rf.Address,
			pz.Winner,
			tokenType,
			pz.UUID, // UUID do prize (seu modelo usa UUID aqui)
		); err != nil {
			t.Fatalf("ClaimRaffle NFT: %v", err)
		}
	}
}

