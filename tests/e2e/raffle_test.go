package e2e_test

import (
	"log"
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

	// --------------------------------------------------------------------
	// Owner + base token
	// --------------------------------------------------------------------
	owner, ownerPriv := createWallet(t, c)
	c.SetPrivateKey(ownerPriv)

	dec := 6
	stablecoin := false
	tok := createBasicToken(t, c, owner.PublicKey, dec, false, tokenV1Domain.FUNGIBLE, stablecoin)

	// Token (mínimo) validate + log
	if tok.Address == "" {
		t.Fatalf("Token address empty")
	}
	if tok.Creator == "" {
		t.Fatalf("Token creator empty")
	}
	if tok.Decimals != dec {
		t.Fatalf("Token decimals mismatch: got %d want %d", tok.Decimals, dec)
	}
	if tok.TokenType != tokenV1Domain.FUNGIBLE {
		t.Fatalf("Token type mismatch: got %s want %s", tok.TokenType, tokenV1Domain.FUNGIBLE)
	}
	if tok.Stablecoin != stablecoin {
		t.Fatalf("Token stablecoin mismatch: got %v want %v", tok.Stablecoin, stablecoin)
	}

	log.Printf("Token Address: %s", tok.Address)
	log.Printf("Token Symbol: %s", tok.Symbol)
	log.Printf("Token Name: %s", tok.Name)
	log.Printf("Token Decimals: %d", tok.Decimals)
	log.Printf("Token Total Supply: %s", tok.TotalSupply)
	log.Printf("Token Type: %s", tok.TokenType)
	log.Printf("Token Stablecoin: %v", tok.Stablecoin)

	// --------------------------------------------------------------------
	// Players (pub/priv map)
	// --------------------------------------------------------------------
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

	// --------------------------------------------------------------------
	// Allow players on base token (owner)
	// --------------------------------------------------------------------
	c.SetPrivateKey(ownerPriv)
	allowPlayersOut, err := c.AllowUsers(tok.Address, map[string]bool{
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
	})
	if err != nil {
		t.Fatalf("AllowUsers(participants): %v", err)
	}
	if len(allowPlayersOut.States) == 0 || allowPlayersOut.States[0].Object == nil {
		t.Fatalf("AllowUsers(participants) returned empty/nil state")
	}

	var ap tokenV1Domain.AccessPolicy
	unmarshalState(t, allowPlayersOut.States[0].Object, &ap)
	if ap.Users == nil || !ap.Users[bob.PublicKey] {
		t.Fatalf("AllowUsers(participants) missing bob in allowlist")
	}

	log.Printf("AllowUsers(participants) Output States: %+v", allowPlayersOut.States)
	log.Printf("AllowUsers(participants) Output Logs: %+v", allowPlayersOut.Logs)
	log.Printf("AllowUsers(participants) Output Delegated Call: %+v", allowPlayersOut.DelegatedCall)

	// --------------------------------------------------------------------
	// Mint base tokens to players - padrão mint
	// --------------------------------------------------------------------
	mintAmt := amt(100, dec)

	type mintCase struct {
		name string
		to   string
	}
	mints := []mintCase{
		{"bob", bob.PublicKey},
		{"alice", alice.PublicKey},
		{"robert", robert.PublicKey},
		{"alfred", alfred.PublicKey},
		{"luiz", luiz.PublicKey},
		{"jorge", jorge.PublicKey},
		{"luigui", luigui.PublicKey},
		{"superman", superman.PublicKey},
		{"spiderman", spiderman.PublicKey},
		{"batman", batman.PublicKey},
		{"wonderwoman", wonderwoman.PublicKey},
	}

	for _, it := range mints {
		out, err := c.MintToken(tok.Address, it.to, mintAmt, dec, tok.TokenType)
		if err != nil {
			t.Fatalf("MintToken(%s): %v", it.name, err)
		}
		if len(out.States) == 0 || out.States[0].Object == nil {
			t.Fatalf("MintToken(%s) returned empty/nil state", it.name)
		}

		var mint tokenV1Domain.Mint
		unmarshalState(t, out.States[0].Object, &mint)

		if mint.TokenAddress != tok.Address {
			t.Fatalf("Mint(%s) TokenAddress mismatch: got %s want %s", it.name, mint.TokenAddress, tok.Address)
		}
		if mint.MintTo != it.to {
			t.Fatalf("Mint(%s) ToAddress mismatch: got %s want %s", it.name, mint.MintTo, it.to)
		}
		if mint.Amount != mintAmt {
			t.Fatalf("Mint(%s) Amount mismatch: got %s want %s", it.name, mint.Amount, mintAmt)
		}
		if mint.TokenType != tok.TokenType {
			t.Fatalf("Mint(%s) TokenType mismatch: got %s want %s", it.name, mint.TokenType, tok.TokenType)
		}
		if tok.TokenType == tokenV1Domain.FUNGIBLE && len(mint.TokenUUIDList) != 0 {
			t.Fatalf("Mint(%s) fungible should not generate UUIDs", it.name)
		}

		log.Printf("Mint(%s) Output States: %+v", it.name, out.States)
		log.Printf("Mint(%s) Output Logs: %+v", it.name, out.Logs)
		log.Printf("Mint(%s) Output Delegated Call: %+v", it.name, out.DelegatedCall)

		log.Printf("Mint(%s) TokenAddress: %s", it.name, mint.TokenAddress)
		log.Printf("Mint(%s) ToAddress: %s", it.name, mint.MintTo)
		log.Printf("Mint(%s) Amount: %s", it.name, mint.Amount)
		log.Printf("Mint(%s) TokenType: %s", it.name, mint.TokenType)
		log.Printf("Mint(%s) TokenUUIDList: %+v", it.name, mint.TokenUUIDList)
	}

	// --------------------------------------------------------------------
	// Merchant runs raffle
	// --------------------------------------------------------------------
	merchant, merchPriv := createWallet(t, c)

	c.SetPrivateKey(ownerPriv)
	allowMerchantOut, err := c.AllowUsers(tok.Address, map[string]bool{merchant.PublicKey: true})
	if err != nil {
		t.Fatalf("AllowUsers(merchant): %v", err)
	}
	if len(allowMerchantOut.States) == 0 || allowMerchantOut.States[0].Object == nil {
		t.Fatalf("AllowUsers(merchant) returned empty/nil state")
	}

	log.Printf("AllowUsers(merchant) Output States: %+v", allowMerchantOut.States)
	log.Printf("AllowUsers(merchant) Output Logs: %+v", allowMerchantOut.Logs)
	log.Printf("AllowUsers(merchant) Output Delegated Call: %+v", allowMerchantOut.DelegatedCall)

	// --------------------------------------------------------------------
	// Deploy Raffle contract
	// --------------------------------------------------------------------
	c.SetPrivateKey(merchPriv)

	start := time.Now().Add(2 * time.Second)
	exp := time.Now().Add(24 * time.Hour)
	seedPass := "e2e-seed"
	commit := seed.CommitSeed(seedPass)
	meta := map[string]string{"campaign": "e2e"}

	contractState := models.ContractStateModel{}
	deployedContract, err := c.DeployContract1(raffleV1.RAFFLE_CONTRACT_V1)
	if err != nil {
		t.Fatalf("DeployContract(Raffle): %v", err)
	}
	if len(deployedContract.States) == 0 || deployedContract.States[0].Object == nil {
		t.Fatalf("DeployContract(Raffle) returned empty/nil state")
	}
	unmarshalState(t, deployedContract.States[0].Object, &contractState)
	address := contractState.Address
	if address == "" {
		t.Fatalf("DeployContract(Raffle) returned empty contract address")
	}

	log.Printf("DeployContract(Raffle) Output States: %+v", deployedContract.States)
	log.Printf("DeployContract(Raffle) Output Logs: %+v", deployedContract.Logs)
	log.Printf("DeployContract(Raffle) Output Delegated Call: %+v", deployedContract.DelegatedCall)
	log.Printf("Raffle Contract Address: %s", address)

	// --------------------------------------------------------------------
	// AddRaffle (merchant) - valida campos determinísticos + log
	// --------------------------------------------------------------------
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
	if len(added.States) == 0 || added.States[0].Object == nil {
		t.Fatalf("AddRaffle returned empty/nil state")
	}

	var rf raffleV1Domain.Raffle
	unmarshalState(t, added.States[0].Object, &rf)

	if rf.Address == "" {
		t.Fatalf("raffle addr empty")
	}
	if rf.Owner != merchant.PublicKey {
		t.Fatalf("AddRaffle Owner mismatch: got %q want %q", rf.Owner, merchant.PublicKey)
	}
	if rf.TokenAddress != tok.Address {
		t.Fatalf("AddRaffle TokenAddress mismatch: got %q want %q", rf.TokenAddress, tok.Address)
	}
	if rf.TicketPrice != amt(1, dec) {
		t.Fatalf("AddRaffle TicketPrice mismatch: got %q want %q", rf.TicketPrice, amt(1, dec))
	}
	if rf.MaxEntries != 100 {
		t.Fatalf("AddRaffle MaxEntries mismatch: got %d want %d", rf.MaxEntries, 100)
	}
	if rf.MaxEntriesPerUser != 5 {
		t.Fatalf("AddRaffle MaxEntriesPerUser mismatch: got %d want %d", rf.MaxEntriesPerUser, 5)
	}
	if rf.Hash == "" {
		t.Fatalf("AddRaffle Hash empty")
	}

	log.Printf("AddRaffle Output States: %+v", added.States)
	log.Printf("AddRaffle Output Logs: %+v", added.Logs)
	log.Printf("AddRaffle Output Delegated Call: %+v", added.DelegatedCall)

	log.Printf("AddRaffle Address: %s", rf.Address)
	log.Printf("AddRaffle Owner: %s", rf.Owner)
	log.Printf("AddRaffle TokenAddress: %s", rf.TokenAddress)
	log.Printf("AddRaffle TicketPrice: %s", rf.TicketPrice)
	log.Printf("AddRaffle MaxEntries: %d", rf.MaxEntries)
	log.Printf("AddRaffle MaxEntriesPerUser: %d", rf.MaxEntriesPerUser)
	log.Printf("AddRaffle StartAt: %s", rf.StartAt.String())
	log.Printf("AddRaffle ExpiredAt: %s", rf.ExpiredAt.String())
	log.Printf("AddRaffle Paused: %v", rf.Paused)
	log.Printf("AddRaffle Hash: %s", rf.Hash)
	log.Printf("AddRaffle SeedCommitHex: %s", rf.SeedCommitHex)

	// --------------------------------------------------------------------
	// Allow raffle contract to move base token (owner)
	// --------------------------------------------------------------------
	c.SetPrivateKey(ownerPriv)
	allowRaffleOut, err := c.AllowUsers(tok.Address, map[string]bool{rf.Address: true})
	if err != nil {
		t.Fatalf("AllowUsers(raffle -> tok): %v", err)
	}
	if len(allowRaffleOut.States) == 0 || allowRaffleOut.States[0].Object == nil {
		t.Fatalf("AllowUsers(raffle -> tok) returned empty/nil state")
	}

	log.Printf("AllowUsers(raffle -> tok) Output States: %+v", allowRaffleOut.States)
	log.Printf("AllowUsers(raffle -> tok) Output Logs: %+v", allowRaffleOut.Logs)
	log.Printf("AllowUsers(raffle -> tok) Output Delegated Call: %+v", allowRaffleOut.DelegatedCall)

	// --------------------------------------------------------------------
	// Update raffle (merchant) - valida campos determinísticos + log
	// --------------------------------------------------------------------
	c.SetPrivateKey(merchPriv)

	newStart := time.Now().Add(1 * time.Hour)
	newExp := time.Now().Add(26 * time.Hour)
	commit2 := seed.CommitSeed(seedPass + "2")

	upOut, err := c.UpdateRaffle(
		rf.Address,
		rf.TokenAddress,
		amt(2, dec),
		150,
		10,
		&newStart,
		&newExp,
		commit2,
		map[string]string{"k": "v"},
	)
	if err != nil {
		t.Fatalf("UpdateRaffle: %v", err)
	}
	if len(upOut.States) == 0 || upOut.States[0].Object == nil {
		t.Fatalf("UpdateRaffle returned empty/nil state")
	}

	var rfUp raffleV1Domain.Raffle
	unmarshalState(t, upOut.States[0].Object, &rfUp)

	if rfUp.Address != rf.Address {
		t.Fatalf("UpdateRaffle Address mismatch: got %q want %q", rfUp.Address, rf.Address)
	}
	if rfUp.TokenAddress != rf.TokenAddress {
		t.Fatalf("UpdateRaffle TokenAddress mismatch: got %q want %q", rfUp.TokenAddress, rf.TokenAddress)
	}
	if rfUp.TicketPrice != amt(2, dec) {
		t.Fatalf("UpdateRaffle TicketPrice mismatch: got %q want %q", rfUp.TicketPrice, amt(2, dec))
	}
	if rfUp.MaxEntries != 150 {
		t.Fatalf("UpdateRaffle MaxEntries mismatch: got %d want %d", rfUp.MaxEntries, 150)
	}
	if rfUp.MaxEntriesPerUser != 10 {
		t.Fatalf("UpdateRaffle MaxEntriesPerUser mismatch: got %d want %d", rfUp.MaxEntriesPerUser, 10)
	}
	if rfUp.SeedCommitHex != commit2 {
		t.Fatalf("UpdateRaffle SeedCommitHex mismatch: got %q want %q", rfUp.SeedCommitHex, commit2)
	}

	log.Printf("UpdateRaffle Output States: %+v", upOut.States)
	log.Printf("UpdateRaffle Output Logs: %+v", upOut.Logs)
	log.Printf("UpdateRaffle Output Delegated Call: %+v", upOut.DelegatedCall)

	// --------------------------------------------------------------------
	// Pause / Unpause (merchant) - valida + log
	// --------------------------------------------------------------------
	pauseOut, err := c.PauseRaffle(rf.Address, true)
	if err != nil {
		t.Fatalf("PauseRaffle: %v", err)
	}
	if len(pauseOut.States) == 0 || pauseOut.States[0].Object == nil {
		t.Fatalf("PauseRaffle returned empty/nil state")
	}
	var paused raffleV1Domain.Raffle
	unmarshalState(t, pauseOut.States[0].Object, &paused)
	if paused.Address != rf.Address {
		t.Fatalf("PauseRaffle Address mismatch: got %q want %q", paused.Address, rf.Address)
	}
	if !paused.Paused {
		t.Fatalf("PauseRaffle expected Paused=true")
	}

	log.Printf("PauseRaffle Output States: %+v", pauseOut.States)
	log.Printf("PauseRaffle Output Logs: %+v", pauseOut.Logs)
	log.Printf("PauseRaffle Output Delegated Call: %+v", pauseOut.DelegatedCall)

	unpauseOut, err := c.UnpauseRaffle(rf.Address, false)
	if err != nil {
		t.Fatalf("UnpauseRaffle: %v", err)
	}
	if len(unpauseOut.States) == 0 || unpauseOut.States[0].Object == nil {
		t.Fatalf("UnpauseRaffle returned empty/nil state")
	}
	var unpaused raffleV1Domain.Raffle
	unmarshalState(t, unpauseOut.States[0].Object, &unpaused)
	if unpaused.Address != rf.Address {
		t.Fatalf("UnpauseRaffle Address mismatch: got %q want %q", unpaused.Address, rf.Address)
	}
	if unpaused.Paused {
		t.Fatalf("UnpauseRaffle expected Paused=false")
	}

	log.Printf("UnpauseRaffle Output States: %+v", unpauseOut.States)
	log.Printf("UnpauseRaffle Output Logs: %+v", unpauseOut.Logs)
	log.Printf("UnpauseRaffle Output Delegated Call: %+v", unpauseOut.DelegatedCall)

	// --------------------------------------------------------------------
	// Wait start and enter (players)
	// --------------------------------------------------------------------
	waitUntil(t, 10*time.Second, func() bool { return time.Now().After(start) })

	enter := func(name, priv string, tickets int) {
		c.SetPrivateKey(priv)
		out, err := c.EnterRaffle(rf.Address, tickets, tok.Address, tokenV1Domain.FUNGIBLE, "")
		if err != nil {
			t.Fatalf("EnterRaffle(%s): %v", name, err)
		}
		if len(out.States) == 0 || out.States[0].Object == nil {
			t.Fatalf("EnterRaffle(%s) returned empty/nil state", name)
		}

		var et raffleV1Models.EnterTicketsModel
		unmarshalState(t, out.States[0].Object, &et)

		if et.RaffleAddress != rf.Address {
			t.Fatalf("EnterRaffle(%s) raffle_address mismatch: got %q want %q", name, et.RaffleAddress, rf.Address)
		}
		// entrant no model = "entrant"
		if et.Entrant == "" {
			t.Fatalf("EnterRaffle(%s) entrant empty", name)
		}
		if et.Tickets != tickets {
			t.Fatalf("EnterRaffle(%s) tickets mismatch: got %d want %d", name, et.Tickets, tickets)
		}
		if et.PayTokenAddress != tok.Address {
			t.Fatalf("EnterRaffle(%s) pay_token_address mismatch: got %q want %q", name, et.PayTokenAddress, tok.Address)
		}
		if et.AmountPaid == "" {
			t.Fatalf("EnterRaffle(%s) amount_paid empty", name)
		}
		if et.UUID == "" {
			t.Fatalf("EnterRaffle(%s) uuid empty", name)
		}

		log.Printf("EnterRaffle(%s) Output States: %+v", name, out.States)
		log.Printf("EnterRaffle(%s) Output Logs: %+v", name, out.Logs)
		log.Printf("EnterRaffle(%s) Output Delegated Call: %+v", name, out.DelegatedCall)

		log.Printf("EnterRaffle(%s) RaffleAddress: %s", name, et.RaffleAddress)
		log.Printf("EnterRaffle(%s) UUID: %s", name, et.UUID)
		log.Printf("EnterRaffle(%s) Entrant: %s", name, et.Entrant)
		log.Printf("EnterRaffle(%s) Tickets: %d", name, et.Tickets)
		log.Printf("EnterRaffle(%s) PayTokenAddress: %s", name, et.PayTokenAddress)
		log.Printf("EnterRaffle(%s) AmountPaid: %s", name, et.AmountPaid)
	}

	enter("bob", bobPriv, 2)
	enter("bob", bobPriv, 7)
	enter("bob", bobPriv, 3)
	enter("bob", bobPriv, 5)
	enter("alice", alicePriv, 5)
	enter("robert", robertPriv, 11)
	enter("alfred", alfredPriv, 13)
	enter("luiz", luizPriv, 17)
	enter("jorge", jorgePriv, 19)
	enter("luigui", luiguiPriv, 23)
	enter("superman", supermanPriv, 29)
	enter("superman", supermanPriv, 31)
	enter("spiderman", spidermanPriv, 37)
	enter("batman", batmanPriv, 41)
	enter("wonderwoman", wonderwomanPriv, 43)

	// ------------------------------------------------------------
	// PRIZES: prize owner creates tokens and deposits in raffle
	// ------------------------------------------------------------
	ownerPrize, ownerPrizePriv := createWallet(t, c)
	mapOfPubPriv[ownerPrize.PublicKey] = ownerPrizePriv

	c.SetPrivateKey(ownerPrizePriv)
	tok1 := createBasicToken(t, c, ownerPrize.PublicKey, 6, false, tokenV1Domain.FUNGIBLE, stablecoin)
	tok2 := createBasicToken(t, c, ownerPrize.PublicKey, 6, false, tokenV1Domain.FUNGIBLE, stablecoin)
	tok3 := createBasicToken(t, c, ownerPrize.PublicKey, 6, false, tokenV1Domain.FUNGIBLE, stablecoin)
	tok4 := createBasicToken(t, c, ownerPrize.PublicKey, 6, false, tokenV1Domain.FUNGIBLE, stablecoin)

	// allow raffle to move prize tokens
	for _, pt := range []struct {
		name string
		tok  tokenV1Domain.Token
	}{
		{"tok1", tok1},
		{"tok2", tok2},
		{"tok3", tok3},
		{"tok4", tok4},
	} {
		out, err := c.AllowUsers(pt.tok.Address, map[string]bool{rf.Address: true})
		if err != nil {
			t.Fatalf("AllowUsers(%s -> raffle): %v", pt.name, err)
		}
		if len(out.States) == 0 || out.States[0].Object == nil {
			t.Fatalf("AllowUsers(%s -> raffle) returned empty/nil state", pt.name)
		}
		log.Printf("AllowUsers(%s -> raffle) Output States: %+v", pt.name, out.States)
		log.Printf("AllowUsers(%s -> raffle) Output Logs: %+v", pt.name, out.Logs)
		log.Printf("AllowUsers(%s -> raffle) Output Delegated Call: %+v", pt.name, out.DelegatedCall)
	}

	// deposit prizes (ownerPrize signs) - valida + log
	addPrize := func(name, prizeTokenAddr, amount string) raffleV1Domain.RafflePrize {
		out, err := c.AddRafflePrize(rf.Address, prizeTokenAddr, amount, tokenV1Domain.FUNGIBLE, "")
		if err != nil {
			t.Fatalf("AddRafflePrize(%s): %v", name, err)
		}
		if len(out.States) == 0 || out.States[0].Object == nil {
			t.Fatalf("AddRafflePrize(%s) returned empty/nil state", name)
		}

		var rp raffleV1Domain.RafflePrize
		unmarshalState(t, out.States[0].Object, &rp)

		if rp.RaffleAddress != rf.Address {
			t.Fatalf("AddRafflePrize(%s) raffle_address mismatch: got %q want %q", name, rp.RaffleAddress, rf.Address)
		}
		if rp.TokenAddress != prizeTokenAddr {
			t.Fatalf("AddRafflePrize(%s) token_address mismatch: got %q want %q", name, rp.TokenAddress, prizeTokenAddr)
		}
		if rp.Amount != amount {
			t.Fatalf("AddRafflePrize(%s) amount mismatch: got %q want %q", name, rp.Amount, amount)
		}
		if rp.UUID == "" {
			t.Fatalf("AddRafflePrize(%s) uuid empty", name)
		}

		log.Printf("AddRafflePrize(%s) Output States: %+v", name, out.States)
		log.Printf("AddRafflePrize(%s) Output Logs: %+v", name, out.Logs)
		log.Printf("AddRafflePrize(%s) Output Delegated Call: %+v", name, out.DelegatedCall)

		log.Printf("AddRafflePrize(%s) RaffleAddress: %s", name, rp.RaffleAddress)
		log.Printf("AddRafflePrize(%s) UUID: %s", name, rp.UUID)
		log.Printf("AddRafflePrize(%s) Sponsor: %s", name, rp.Sponsor)
		log.Printf("AddRafflePrize(%s) TokenAddress: %s", name, rp.TokenAddress)
		log.Printf("AddRafflePrize(%s) Amount: %s", name, rp.Amount)

		return rp
	}

	rp1 := addPrize("tok1", tok1.Address, amt(2, dec))
	_ = addPrize("tok2", tok2.Address, amt(3, dec))
	_ = addPrize("tok3", tok3.Address, amt(4, dec))
	_ = addPrize("tok4", tok4.Address, amt(5, dec))

	// remove last prize (ownerPrize signs) - valida + log (usa rp1 só pra garantir UUID válido de algum)
	remOut, err := c.RemoveRafflePrize(rf.Address, tokenV1Domain.FUNGIBLE, rp1.UUID)
	if err != nil {
		t.Fatalf("RemoveRafflePrize: %v", err)
	}
	if len(remOut.States) == 0 || remOut.States[0].Object == nil {
		t.Fatalf("RemoveRafflePrize returned empty/nil state")
	}
	var removed raffleV1Domain.RafflePrize
	unmarshalState(t, remOut.States[0].Object, &removed)
	if removed.UUID != rp1.UUID {
		t.Fatalf("RemoveRafflePrize UUID mismatch: got %q want %q", removed.UUID, rp1.UUID)
	}

	log.Printf("RemoveRafflePrize Output States: %+v", remOut.States)
	log.Printf("RemoveRafflePrize Output Logs: %+v", remOut.Logs)
	log.Printf("RemoveRafflePrize Output Delegated Call: %+v", remOut.DelegatedCall)

	// --------------------------------------------------------------------
	// Draw (merchant) - valida + log
	// --------------------------------------------------------------------
	c.SetPrivateKey(merchPriv)
	drawOut, err := c.DrawRaffle(rf.Address, seedPass+"2")
	if err != nil {
		t.Fatalf("DrawRaffle: %v", err)
	}
	if len(drawOut.States) == 0 || drawOut.States[0].Object == nil {
		t.Fatalf("DrawRaffle returned empty/nil state")
	}

	var drawn []raffleV1Models.RafflePrizeModel
	unmarshalState(t, drawOut.States[0].Object, &drawn)

	log.Printf("DrawRaffle Output States: %+v", drawOut.States)
	log.Printf("DrawRaffle Output Logs: %+v", drawOut.Logs)
	log.Printf("DrawRaffle Output Delegated Call: %+v", drawOut.DelegatedCall)

	// --------------------------------------------------------------------
	// ListPrizes - valida + log
	// --------------------------------------------------------------------
	listPrizesOut, err := c.ListPrizes(rf.Address, 1, 50, true)
	if err != nil {
		t.Fatalf("ListPrizes: %v", err)
	}
	if len(listPrizesOut.States) == 0 || listPrizesOut.States[0].Object == nil {
		t.Fatalf("ListPrizes returned empty/nil state")
	}

	var prizes []raffleV1Models.RafflePrizeModel
	unmarshalState(t, listPrizesOut.States[0].Object, &prizes)

	foundAny := false
	for _, pz := range prizes {
		if pz.UUID != "" {
			foundAny = true
			break
		}
	}
	if !foundAny {
		t.Fatalf("ListPrizes returned empty prize list (no UUIDs)")
	}

	log.Printf("ListPrizes Output States: %+v", listPrizesOut.States)
	log.Printf("ListPrizes Output Logs: %+v", listPrizesOut.Logs)
	log.Printf("ListPrizes Output Delegated Call: %+v", listPrizesOut.DelegatedCall)

	// --------------------------------------------------------------------
	// Claim prizes - valida + log
	// --------------------------------------------------------------------
	for idx, prize := range prizes {
		if prize.Winner == "" {
			continue
		}

		// 1) owner do token do prêmio libera o WINNER no token do prêmio
		c.SetPrivateKey(ownerPrizePriv)
		allowWinnerOut, err := c.AllowUsers(prize.TokenAddress, map[string]bool{prize.Winner: true})
		if err != nil {
			t.Fatalf("AllowUsers(prizeToken -> winner): token=%s winner=%s err=%v", prize.TokenAddress, prize.Winner, err)
		}
		if len(allowWinnerOut.States) == 0 || allowWinnerOut.States[0].Object == nil {
			t.Fatalf("AllowUsers(prizeToken -> winner) returned empty/nil state")
		}

		// 2) winner assina o claim
		priv, ok := mapOfPubPriv[prize.Winner]
		if !ok {
			t.Fatalf("missing private key for winner %s", prize.Winner)
		}
		c.SetPrivateKey(priv)

		claimOut, err := c.ClaimRaffle(rf.Address, prize.Winner, tokenV1Domain.FUNGIBLE, "")
		if err != nil {
			t.Fatalf("ClaimRaffle: %v", err)
		}
		if len(claimOut.States) == 0 || claimOut.States[0].Object == nil {
			t.Fatalf("ClaimRaffle returned empty/nil state")
		}

		var claimed raffleV1Models.RafflePrizeModel
		unmarshalState(t, claimOut.States[0].Object, &claimed)

		if claimed.UUID == "" {
			t.Fatalf("ClaimRaffle returned empty prize uuid")
		}
		if claimed.Winner != prize.Winner {
			t.Fatalf("ClaimRaffle winner mismatch: got %q want %q", claimed.Winner, prize.Winner)
		}

		log.Printf("ClaimRaffle Output States: %+v", claimOut.States)
		log.Printf("ClaimRaffle Output Logs: %+v", claimOut.Logs)
		log.Printf("ClaimRaffle Output Delegated Call: %+v", claimOut.DelegatedCall)

		// última iteração: claim duplicado (espera erro)
		if len(prizes) == idx+1 {
			c.SetPrivateKey(priv)
			_, err := c.ClaimRaffle(rf.Address, prize.Winner, tokenV1Domain.FUNGIBLE, "")
			if err == nil {
				t.Fatalf("expected error on duplicate claim, got nil")
			}
		}
	}

	// --------------------------------------------------------------------
	// List prizes claimed - valida todos claimed
	// --------------------------------------------------------------------
	listClaimedOut, err := c.ListPrizes(rf.Address, 1, 50, true)
	if err != nil {
		t.Fatalf("ListPrizes(claimed): %v", err)
	}
	if len(listClaimedOut.States) == 0 || listClaimedOut.States[0].Object == nil {
		t.Fatalf("ListPrizes(claimed) returned empty/nil state")
	}

	var claimedList []raffleV1Models.RafflePrizeModel
	unmarshalState(t, listClaimedOut.States[0].Object, &claimedList)

	for _, prize := range claimedList {
		// se seu backend mantiver prizes sem winner (ex: removidos), você pode flexibilizar aqui.
		if prize.Winner != "" && !prize.Claimed {
			t.Fatalf("prize must be claimed: %+v", prize)
		}
	}

	// --------------------------------------------------------------------
	// Withdraw leftovers (merchant)
	// --------------------------------------------------------------------
	c.SetPrivateKey(merchPriv)
	withdrawOut, err := c.WithdrawRaffle(rf.Address, tok.Address, amt(1, dec), tokenV1Domain.FUNGIBLE, "")
	if err != nil {
		t.Fatalf("WithdrawRaffle: %v", err)
	}
	if len(withdrawOut.States) == 0 || withdrawOut.States[0].Object == nil {
		t.Fatalf("WithdrawRaffle returned empty/nil state")
	}

	log.Printf("WithdrawRaffle Output States: %+v", withdrawOut.States)
	log.Printf("WithdrawRaffle Output Logs: %+v", withdrawOut.Logs)
	log.Printf("WithdrawRaffle Output Delegated Call: %+v", withdrawOut.DelegatedCall)
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

	if tok.Address == "" {
		t.Fatalf("Token address empty")
	}
	if tok.TokenType != tokenType {
		t.Fatalf("Token type mismatch: got %s want %s", tok.TokenType, tokenType)
	}

	log.Printf("Token Address: %s", tok.Address)
	log.Printf("Token Type: %s", tok.TokenType)

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

	// allow players
	c.SetPrivateKey(ownerPriv)
	allowPlayersOut, err := c.AllowUsers(tok.Address, map[string]bool{
		bob.PublicKey:   true,
		alice.PublicKey: true,
	})
	if err != nil {
		t.Fatalf("AllowUsers(players): %v", err)
	}
	if len(allowPlayersOut.States) == 0 || allowPlayersOut.States[0].Object == nil {
		t.Fatalf("AllowUsers(players) returned empty/nil state")
	}

	log.Printf("AllowUsers(players) Output States: %+v", allowPlayersOut.States)
	log.Printf("AllowUsers(players) Output Logs: %+v", allowPlayersOut.Logs)
	log.Printf("AllowUsers(players) Output Delegated Call: %+v", allowPlayersOut.DelegatedCall)

	// --------------------------------------------------------------------
	// Mint 1 NFT ticket per player - padrão mint
	// --------------------------------------------------------------------
	for i := range players {
		mintOut, err := c.MintToken(tok.Address, players[i].pub, "1", dec, tokenType)
		if err != nil {
			t.Fatalf("MintToken NFT: %v", err)
		}
		if len(mintOut.States) == 0 || mintOut.States[0].Object == nil {
			t.Fatalf("MintToken NFT returned empty/nil state")
		}

		var mint tokenV1Domain.Mint
		unmarshalState(t, mintOut.States[0].Object, &mint)

		if mint.TokenAddress != tok.Address {
			t.Fatalf("Mint TokenAddress mismatch: got %s want %s", mint.TokenAddress, tok.Address)
		}
		if mint.MintTo != players[i].pub {
			t.Fatalf("Mint ToAddress mismatch: got %s want %s", mint.MintTo, players[i].pub)
		}
		if mint.TokenType != tok.TokenType {
			t.Fatalf("Mint TokenType mismatch: got %s want %s", mint.TokenType, tok.TokenType)
		}
		if len(mint.TokenUUIDList) != 1 {
			t.Fatalf("expected 1 uuid, got %d", len(mint.TokenUUIDList))
		}

		players[i].uuid = mint.TokenUUIDList[0]
		if players[i].uuid == "" {
			t.Fatalf("minted uuid empty")
		}

		log.Printf("Mint Output States: %+v", mintOut.States)
		log.Printf("Mint Output Logs: %+v", mintOut.Logs)
		log.Printf("Mint Output Delegated Call: %+v", mintOut.DelegatedCall)

		log.Printf("Mint TokenAddress: %s", mint.TokenAddress)
		log.Printf("Mint ToAddress: %s", mint.MintTo)
		log.Printf("Mint Amount: %s", mint.Amount)
		log.Printf("Mint TokenType: %s", mint.TokenType)
		log.Printf("Mint TokenUUIDList: %+v", mint.TokenUUIDList)
	}

	// --------------------------------------------------------------------
	// Merchant runs raffle
	// --------------------------------------------------------------------
	merchant, merchPriv := createWallet(t, c)

	c.SetPrivateKey(ownerPriv)
	allowMerchantOut, err := c.AllowUsers(tok.Address, map[string]bool{merchant.PublicKey: true})
	if err != nil {
		t.Fatalf("AllowUsers(merchant): %v", err)
	}
	if len(allowMerchantOut.States) == 0 || allowMerchantOut.States[0].Object == nil {
		t.Fatalf("AllowUsers(merchant) returned empty/nil state")
	}

	// --------------------------------------------------------------------
	// Deploy raffle contract
	// --------------------------------------------------------------------
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
	if len(deployedContract.States) == 0 || deployedContract.States[0].Object == nil {
		t.Fatalf("DeployContract returned empty/nil state")
	}
	unmarshalState(t, deployedContract.States[0].Object, &contractState)

	// --------------------------------------------------------------------
	// AddRaffle (merchant) - valida + log
	// --------------------------------------------------------------------
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
	if len(added.States) == 0 || added.States[0].Object == nil {
		t.Fatalf("AddRaffle NFT returned empty/nil state")
	}

	var rf raffleV1Domain.Raffle
	unmarshalState(t, added.States[0].Object, &rf)

	if rf.Address == "" {
		t.Fatalf("raffle addr empty")
	}
	if rf.Owner != merchant.PublicKey {
		t.Fatalf("AddRaffle Owner mismatch: got %q want %q", rf.Owner, merchant.PublicKey)
	}
	if rf.TokenAddress != tok.Address {
		t.Fatalf("AddRaffle TokenAddress mismatch: got %q want %q", rf.TokenAddress, tok.Address)
	}
	if rf.TicketPrice != "1" {
		t.Fatalf("AddRaffle TicketPrice mismatch: got %q want %q", rf.TicketPrice, "1")
	}
	if rf.MaxEntries != 10 {
		t.Fatalf("AddRaffle MaxEntries mismatch: got %d want %d", rf.MaxEntries, 10)
	}
	if rf.MaxEntriesPerUser != 1 {
		t.Fatalf("AddRaffle MaxEntriesPerUser mismatch: got %d want %d", rf.MaxEntriesPerUser, 1)
	}
	if rf.SeedCommitHex != commit {
		t.Fatalf("AddRaffle SeedCommitHex mismatch: got %q want %q", rf.SeedCommitHex, commit)
	}
	if rf.Hash == "" {
		t.Fatalf("AddRaffle Hash empty")
	}

	log.Printf("AddRaffle NFT Output States: %+v", added.States)
	log.Printf("AddRaffle NFT Output Logs: %+v", added.Logs)
	log.Printf("AddRaffle NFT Output Delegated Call: %+v", added.DelegatedCall)

	// allow raffle to move ticket token
	c.SetPrivateKey(ownerPriv)
	allowRaffleOut, err := c.AllowUsers(tok.Address, map[string]bool{rf.Address: true})
	if err != nil {
		t.Fatalf("AllowUsers(raffle -> ticketToken): %v", err)
	}
	if len(allowRaffleOut.States) == 0 || allowRaffleOut.States[0].Object == nil {
		t.Fatalf("AllowUsers(raffle -> ticketToken) returned empty/nil state")
	}

	// --------------------------------------------------------------------
	// Players enter raffle (UUID obrigatório) - valida + log
	// --------------------------------------------------------------------
	waitUntil(t, 10*time.Second, func() bool { return time.Now().After(start) })

	for _, p := range players {
		c.SetPrivateKey(p.priv)

		out, err := c.EnterRaffle(rf.Address, 1, tok.Address, tokenType, p.uuid)
		if err != nil {
			t.Fatalf("EnterRaffle NFT: %v", err)
		}
		if len(out.States) == 0 || out.States[0].Object == nil {
			t.Fatalf("EnterRaffle NFT returned empty/nil state")
		}

		var et raffleV1Models.EnterTicketsModel
		unmarshalState(t, out.States[0].Object, &et)

		if et.RaffleAddress != rf.Address {
			t.Fatalf("EnterRaffle NFT raffle_address mismatch: got %q want %q", et.RaffleAddress, rf.Address)
		}
		if et.Entrant != p.pub {
			t.Fatalf("EnterRaffle NFT entrant mismatch: got %q want %q", et.Entrant, p.pub)
		}
		if et.Tickets != 1 {
			t.Fatalf("EnterRaffle NFT tickets mismatch: got %d want %d", et.Tickets, 1)
		}
		if et.PayTokenAddress != tok.Address {
			t.Fatalf("EnterRaffle NFT pay_token_address mismatch: got %q want %q", et.PayTokenAddress, tok.Address)
		}
		if et.AmountPaid == "" {
			t.Fatalf("EnterRaffle NFT amount_paid empty")
		}
		if et.UUID == "" {
			t.Fatalf("EnterRaffle NFT uuid empty")
		}

		log.Printf("EnterRaffle NFT Output States: %+v", out.States)
		log.Printf("EnterRaffle NFT Output Logs: %+v", out.Logs)
		log.Printf("EnterRaffle NFT Output Delegated Call: %+v", out.DelegatedCall)
	}

	// --------------------------------------------------------------------
	// CREATE PRIZES (NFT)
	// --------------------------------------------------------------------
	prizeOwner, prizePriv := createWallet(t, c)
	c.SetPrivateKey(prizePriv)

	prizeToken := createBasicToken(t, c, prizeOwner.PublicKey, 0, false, tokenType, stablecoin)

	mintPrizeOut, err := c.MintToken(prizeToken.Address, prizeOwner.PublicKey, "1", 0, tokenType)
	if err != nil {
		t.Fatalf("Mint prize NFT: %v", err)
	}
	if len(mintPrizeOut.States) == 0 || mintPrizeOut.States[0].Object == nil {
		t.Fatalf("Mint prize NFT returned empty/nil state")
	}

	var prizeMint tokenV1Domain.Mint
	unmarshalState(t, mintPrizeOut.States[0].Object, &prizeMint)
	if len(prizeMint.TokenUUIDList) != 1 {
		t.Fatalf("prize mint expected 1 uuid, got %d", len(prizeMint.TokenUUIDList))
	}
	prizeUUID := prizeMint.TokenUUIDList[0]
	if prizeUUID == "" {
		t.Fatalf("prize uuid empty")
	}

	allowPrizeOut, err := c.AllowUsers(prizeToken.Address, map[string]bool{rf.Address: true})
	if err != nil {
		t.Fatalf("AllowUsers(prizeToken -> raffle): %v", err)
	}
	if len(allowPrizeOut.States) == 0 || allowPrizeOut.States[0].Object == nil {
		t.Fatalf("AllowUsers(prizeToken -> raffle) returned empty/nil state")
	}

	// deposit prize (prizeOwner signs) - valida + log
	addPrizeOut, err := c.AddRafflePrize(rf.Address, prizeToken.Address, "1", tokenType, prizeUUID)
	if err != nil {
		t.Fatalf("AddRafflePrize NFT: %v", err)
	}
	if len(addPrizeOut.States) == 0 || addPrizeOut.States[0].Object == nil {
		t.Fatalf("AddRafflePrize NFT returned empty/nil state")
	}

	var rp raffleV1Domain.RafflePrize
	unmarshalState(t, addPrizeOut.States[0].Object, &rp)

	if rp.RaffleAddress != rf.Address {
		t.Fatalf("AddRafflePrize NFT raffle_address mismatch: got %q want %q", rp.RaffleAddress, rf.Address)
	}
	if rp.TokenAddress != prizeToken.Address {
		t.Fatalf("AddRafflePrize NFT token_address mismatch: got %q want %q", rp.TokenAddress, prizeToken.Address)
	}
	if rp.Amount != "1" {
		t.Fatalf("AddRafflePrize NFT amount mismatch: got %q want %q", rp.Amount, "1")
	}
	if rp.UUID == "" {
		t.Fatalf("AddRafflePrize NFT uuid empty")
	}

	log.Printf("AddRafflePrize NFT Output States: %+v", addPrizeOut.States)
	log.Printf("AddRafflePrize NFT Output Logs: %+v", addPrizeOut.Logs)
	log.Printf("AddRafflePrize NFT Output Delegated Call: %+v", addPrizeOut.DelegatedCall)

	// --------------------------------------------------------------------
	// DRAW
	// --------------------------------------------------------------------
	c.SetPrivateKey(merchPriv)
	drawOut, err := c.DrawRaffle(rf.Address, seedPass)
	if err != nil {
		t.Fatalf("DrawRaffle NFT: %v", err)
	}
	if len(drawOut.States) == 0 || drawOut.States[0].Object == nil {
		t.Fatalf("DrawRaffle NFT returned empty/nil state")
	}

	var prizes []raffleV1Models.RafflePrizeModel
	unmarshalState(t, drawOut.States[0].Object, &prizes)

	log.Printf("DrawRaffle NFT Output States: %+v", drawOut.States)
	log.Printf("DrawRaffle NFT Output Logs: %+v", drawOut.Logs)
	log.Printf("DrawRaffle NFT Output Delegated Call: %+v", drawOut.DelegatedCall)

	// --------------------------------------------------------------------
	// Claim prize (winner)
	// --------------------------------------------------------------------
	for _, pz := range prizes {
		if pz.Winner == "" {
			continue
		}

		// prize owner libera winner no prize token
		c.SetPrivateKey(prizePriv)
		allowWinnerOut, err := c.AllowUsers(prizeToken.Address, map[string]bool{pz.Winner: true})
		if err != nil {
			t.Fatalf("AllowUsers(prizeToken -> winner): %v", err)
		}
		if len(allowWinnerOut.States) == 0 || allowWinnerOut.States[0].Object == nil {
			t.Fatalf("AllowUsers(prizeToken -> winner) returned empty/nil state")
		}

		// winner signs claim
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
		claimOut, err := c.ClaimRaffle(rf.Address, pz.Winner, tokenType, pz.UUID)
		if err != nil {
			t.Fatalf("ClaimRaffle NFT: %v", err)
		}
		if len(claimOut.States) == 0 || claimOut.States[0].Object == nil {
			t.Fatalf("ClaimRaffle NFT returned empty/nil state")
		}

		var claimed raffleV1Models.RafflePrizeModel
		unmarshalState(t, claimOut.States[0].Object, &claimed)
		if claimed.UUID == "" {
			t.Fatalf("ClaimRaffle NFT returned empty uuid")
		}
		if claimed.Winner != pz.Winner {
			t.Fatalf("ClaimRaffle NFT winner mismatch: got %q want %q", claimed.Winner, pz.Winner)
		}

		log.Printf("ClaimRaffle NFT Output States: %+v", claimOut.States)
		log.Printf("ClaimRaffle NFT Output Logs: %+v", claimOut.Logs)
		log.Printf("ClaimRaffle NFT Output Delegated Call: %+v", claimOut.DelegatedCall)
	}
}
