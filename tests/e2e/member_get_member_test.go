package e2e_test

import (
	"log"
	"testing"
	"time"

	"gitlab.com/2finance/2finance-network/blockchain/contract/contractV1/models"
	faucetV1 "gitlab.com/2finance/2finance-network/blockchain/contract/faucetV1"
	faucetV1Models "gitlab.com/2finance/2finance-network/blockchain/contract/faucetV1/models"
	mgmV1 "gitlab.com/2finance/2finance-network/blockchain/contract/memberGetMemberV1"
	memberGetMemberV1Models "gitlab.com/2finance/2finance-network/blockchain/contract/memberGetMemberV1/models"
	tokenV1Domain "gitlab.com/2finance/2finance-network/blockchain/contract/tokenV1/domain"
)

func TestMgMFlow(t *testing.T) {
	c := setupClient(t)

	// --------------------------------------------------------------------
	// Owner + base token
	// --------------------------------------------------------------------
	owner, ownerPriv := createWallet(t, c)
	c.SetPrivateKey(ownerPriv)

	dec := 6
	stablecoin := true
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
	// Deploy MgM + Faucet contracts
	// --------------------------------------------------------------------
	var contractState models.ContractStateModel

	deployedMgm, err := c.DeployContract1(mgmV1.MEMBER_GET_MEMBER_CONTRACT_V1)
	if err != nil {
		t.Fatalf("DeployContract (mgm): %v", err)
	}
	if len(deployedMgm.States) == 0 {
		t.Fatalf("DeployContract (mgm) returned empty States")
	}
	if deployedMgm.States[0].Object == nil {
		t.Fatalf("DeployContract (mgm) returned nil state object")
	}
	unmarshalState(t, deployedMgm.States[0].Object, &contractState)
	mgmAddress := contractState.Address
	if mgmAddress == "" {
		t.Fatalf("mgmAddress empty")
	}

	log.Printf("DeployContract(mgm) Output States: %+v", deployedMgm.States)
	log.Printf("DeployContract(mgm) Output Logs: %+v", deployedMgm.Logs)
	log.Printf("DeployContract(mgm) Output Delegated Call: %+v", deployedMgm.DelegatedCall)
	log.Printf("MgM Contract Address: %s", mgmAddress)

	deployedFaucet, err := c.DeployContract1(faucetV1.FAUCET_CONTRACT_V1)
	if err != nil {
		t.Fatalf("DeployContract (faucet): %v", err)
	}
	if len(deployedFaucet.States) == 0 {
		t.Fatalf("DeployContract (faucet) returned empty States")
	}
	if deployedFaucet.States[0].Object == nil {
		t.Fatalf("DeployContract (faucet) returned nil state object")
	}
	unmarshalState(t, deployedFaucet.States[0].Object, &contractState)
	faucetAddress := contractState.Address
	if faucetAddress == "" {
		t.Fatalf("faucetAddress empty")
	}

	log.Printf("DeployContract(faucet) Output States: %+v", deployedFaucet.States)
	log.Printf("DeployContract(faucet) Output Logs: %+v", deployedFaucet.Logs)
	log.Printf("DeployContract(faucet) Output Delegated Call: %+v", deployedFaucet.DelegatedCall)
	log.Printf("Faucet Contract Address: %s", faucetAddress)

	// --------------------------------------------------------------------
	// Create Faucet (owner)
	// --------------------------------------------------------------------
	start := time.Now().Add(2 * time.Second)
	exp := time.Now().Add(1 * time.Hour)
	amount := "5"

	faucetOut, err := c.AddFaucet(
		faucetAddress,
		owner.PublicKey,
		tok.Address,
		start,
		exp,
		false,
		3,
		amount,
		2*time.Second,
	)
	if err != nil {
		t.Fatalf("AddFaucet: %v", err)
	}
	if len(faucetOut.States) == 0 {
		t.Fatalf("AddFaucet returned empty States")
	}
	if faucetOut.States[0].Object == nil {
		t.Fatalf("AddFaucet returned nil state object")
	}

	var f faucetV1Models.FaucetStateModel
	unmarshalState(t, faucetOut.States[0].Object, &f)

	if f.Address == "" {
		t.Fatalf("AddFaucet returned empty faucet address")
	}
	if f.Owner != owner.PublicKey {
		t.Fatalf("AddFaucet Owner mismatch: got %q want %q", f.Owner, owner.PublicKey)
	}
	if f.TokenAddress != tok.Address {
		t.Fatalf("AddFaucet TokenAddress mismatch: got %q want %q", f.TokenAddress, tok.Address)
	}
	if f.StartTime == nil || f.ExpireTime == nil {
		t.Fatalf("AddFaucet start/expire nil: start=%v expire=%v", f.StartTime, f.ExpireTime)
	}
	if f.Paused != false {
		t.Fatalf("AddFaucet Paused mismatch: got %v want %v", f.Paused, false)
	}
	if f.RequestLimit != 3 {
		t.Fatalf("AddFaucet RequestLimit mismatch: got %d want %d", f.RequestLimit, 3)
	}
	if f.Hash == "" {
		t.Fatalf("AddFaucet Hash empty")
	}
	if f.ClaimAmount != amount {
		t.Fatalf("AddFaucet ClaimAmount mismatch: got %q want %q", f.ClaimAmount, amount)
	}
	if f.ClaimIntervalDuration == nil {
		t.Fatalf("AddFaucet ClaimIntervalDuration nil")
	}
	if *f.ClaimIntervalDuration != 2*time.Second {
		t.Fatalf("AddFaucet ClaimIntervalDuration mismatch: got %v want %v", *f.ClaimIntervalDuration, 2*time.Second)
	}

	log.Printf("AddFaucet Output States: %+v", faucetOut.States)
	log.Printf("AddFaucet Output Logs: %+v", faucetOut.Logs)
	log.Printf("AddFaucet Output Delegated Call: %+v", faucetOut.DelegatedCall)

	log.Printf("AddFaucet Address: %s", f.Address)
	log.Printf("AddFaucet Owner: %s", f.Owner)
	log.Printf("AddFaucet TokenAddress: %s", f.TokenAddress)
	log.Printf("AddFaucet StartTime: %s", f.StartTime.String())
	log.Printf("AddFaucet ExpireTime: %s", f.ExpireTime.String())
	log.Printf("AddFaucet Paused: %v", f.Paused)
	log.Printf("AddFaucet RequestLimit: %d", f.RequestLimit)
	log.Printf("AddFaucet ClaimAmount: %s", f.ClaimAmount)
	log.Printf("AddFaucet ClaimIntervalDuration: %v", *f.ClaimIntervalDuration)
	log.Printf("AddFaucet RequestsByUserJSONB: %+v", f.RequestsByUserJSONB)
	log.Printf("AddFaucet LastClaimByUserJSONB: %+v", f.LastClaimByUserJSONB)
	log.Printf("AddFaucet Hash: %s", f.Hash)

	// --------------------------------------------------------------------
	// Add MgM (owner)
	// --------------------------------------------------------------------
	log.Printf("MgM %s", mgmAddress)

	rewardAmount := amt(10, dec)

	addOut, err := c.AddMgM(
		mgmAddress,
		owner.PublicKey,
		tok.Address,
		faucetAddress,
		rewardAmount,
		start,
		exp,
		false,
	)
	if err != nil {
		t.Fatalf("AddMgM: %v", err)
	}
	if len(addOut.States) == 0 {
		t.Fatalf("AddMgM returned empty States")
	}
	if addOut.States[0].Object == nil {
		t.Fatalf("AddMgM returned nil state object")
	}

	var mgm memberGetMemberV1Models.MgMStateModel
	unmarshalState(t, addOut.States[0].Object, &mgm)

	if mgm.Address == "" {
		t.Fatalf("AddMgM returned empty address")
	}
	if mgm.Address != mgmAddress {
		t.Fatalf("AddMgM address mismatch: got %q want %q", mgm.Address, mgmAddress)
	}
	if mgm.Owner != owner.PublicKey {
		t.Fatalf("AddMgM owner mismatch: got %q want %q", mgm.Owner, owner.PublicKey)
	}
	if mgm.TokenAddress != tok.Address {
		t.Fatalf("AddMgM token_address mismatch: got %q want %q", mgm.TokenAddress, tok.Address)
	}
	if mgm.FaucetAddress != faucetAddress {
		t.Fatalf("AddMgM faucet_address mismatch: got %q want %q", mgm.FaucetAddress, faucetAddress)
	}
	if mgm.StartAt == nil || mgm.ExpireAt == nil {
		t.Fatalf("AddMgM start/expire nil: start=%v expire=%v", mgm.StartAt, mgm.ExpireAt)
	}
	if mgm.Paused != false {
		t.Fatalf("AddMgM paused mismatch: got %v want %v", mgm.Paused, false)
	}
	if mgm.Hash == "" {
		t.Fatalf("AddMgM hash empty")
	}

	log.Printf("AddMgM Output States: %+v", addOut.States)
	log.Printf("AddMgM Output Logs: %+v", addOut.Logs)
	log.Printf("AddMgM Output Delegated Call: %+v", addOut.DelegatedCall)

	log.Printf("AddMgM Address: %s", mgm.Address)
	log.Printf("AddMgM Owner: %s", mgm.Owner)
	log.Printf("AddMgM TokenAddress: %s", mgm.TokenAddress)
	log.Printf("AddMgM FaucetAddress: %s", mgm.FaucetAddress)
	log.Printf("AddMgM StartAt: %s", mgm.StartAt.String())
	log.Printf("AddMgM ExpireAt: %s", mgm.ExpireAt.String())
	log.Printf("AddMgM Paused: %v", mgm.Paused)
	log.Printf("AddMgM Hash: %s", mgm.Hash)

	// --------------------------------------------------------------------
	// Allow token transfers to MgM + Faucet + Owner (best effort, mas com validate+log se vier state)
	// --------------------------------------------------------------------
	allowOut, err := c.AllowUsers(tok.Address, map[string]bool{
		mgmAddress:      true,
		faucetAddress:   true,
		owner.PublicKey: true,
	})
	if err != nil {
		t.Fatalf("AllowUsers(token): %v", err)
	}
	if len(allowOut.States) == 0 || allowOut.States[0].Object == nil {
		t.Fatalf("AllowUsers(token) returned empty/nil state")
	}

	var ap tokenV1Domain.AccessPolicy
	unmarshalState(t, allowOut.States[0].Object, &ap)
	if ap.Users == nil || !ap.Users[mgmAddress] || !ap.Users[faucetAddress] || !ap.Users[owner.PublicKey] {
		t.Fatalf("AllowUsers(token) missing expected users")
	}

	log.Printf("AllowUsers(token) Output States: %+v", allowOut.States)
	log.Printf("AllowUsers(token) Output Logs: %+v", allowOut.Logs)
	log.Printf("AllowUsers(token) Output Delegated Call: %+v", allowOut.DelegatedCall)

	log.Printf("AllowUsers Mode: %s", ap.Mode)
	log.Printf("AllowUsers Users: %+v", ap.Users)

	// --------------------------------------------------------------------
	// Deposit/Withdraw pool funds (se o backend gatear, mantém warning)
	// --------------------------------------------------------------------
	depOut, err := c.DepositMgM(mgmAddress, amt(100, dec), tokenV1Domain.FUNGIBLE, "")
	if err != nil {
		t.Logf("DepositMgM warning: %v", err)
	} else {
		if len(depOut.States) == 0 || depOut.States[0].Object == nil {
			t.Fatalf("DepositMgM returned empty/nil state")
		}

		var mgmAfterDep memberGetMemberV1Models.MgMStateModel
		unmarshalState(t, depOut.States[0].Object, &mgmAfterDep)
		if mgmAfterDep.Address != mgmAddress {
			t.Fatalf("DepositMgM address mismatch: got %q want %q", mgmAfterDep.Address, mgmAddress)
		}

		log.Printf("DepositMgM Output States: %+v", depOut.States)
		log.Printf("DepositMgM Output Logs: %+v", depOut.Logs)
		log.Printf("DepositMgM Output Delegated Call: %+v", depOut.DelegatedCall)

		log.Printf("DepositMgM Address: %s", mgmAfterDep.Address)
		log.Printf("DepositMgM Hash: %s", mgmAfterDep.Hash)
	}

	withOut, err := c.WithdrawMgM(mgmAddress, amt(1, dec), tokenV1Domain.FUNGIBLE, "")
	if err != nil {
		t.Logf("WithdrawMgM warning: %v", err)
	} else {
		if len(withOut.States) == 0 || withOut.States[0].Object == nil {
			t.Fatalf("WithdrawMgM returned empty/nil state")
		}

		var mgmAfterWith memberGetMemberV1Models.MgMStateModel
		unmarshalState(t, withOut.States[0].Object, &mgmAfterWith)
		if mgmAfterWith.Address != mgmAddress {
			t.Fatalf("WithdrawMgM address mismatch: got %q want %q", mgmAfterWith.Address, mgmAddress)
		}

		log.Printf("WithdrawMgM Output States: %+v", withOut.States)
		log.Printf("WithdrawMgM Output Logs: %+v", withOut.Logs)
		log.Printf("WithdrawMgM Output Delegated Call: %+v", withOut.DelegatedCall)

		log.Printf("WithdrawMgM Address: %s", mgmAfterWith.Address)
		log.Printf("WithdrawMgM Hash: %s", mgmAfterWith.Hash)
	}

	// --------------------------------------------------------------------
	// Update MgM (owner)
	// --------------------------------------------------------------------
	newStart := time.Now().Add(3 * time.Minute)
	newExp := time.Now().Add(2 * time.Hour)

	updOut, err := c.UpdateMgM(mgmAddress, amt(20, dec), newStart, newExp)
	if err != nil {
		t.Fatalf("UpdateMgM: %v", err)
	}
	if len(updOut.States) == 0 || updOut.States[0].Object == nil {
		t.Fatalf("UpdateMgM returned empty/nil state")
	}

	var mgmUpd memberGetMemberV1Models.MgMStateModel
	unmarshalState(t, updOut.States[0].Object, &mgmUpd)
	if mgmUpd.Address != mgmAddress {
		t.Fatalf("UpdateMgM address mismatch: got %q want %q", mgmUpd.Address, mgmAddress)
	}
	if mgmUpd.StartAt == nil || mgmUpd.ExpireAt == nil {
		t.Fatalf("UpdateMgM start/expire nil")
	}

	log.Printf("UpdateMgM Output States: %+v", updOut.States)
	log.Printf("UpdateMgM Output Logs: %+v", updOut.Logs)
	log.Printf("UpdateMgM Output Delegated Call: %+v", updOut.DelegatedCall)

	log.Printf("UpdateMgM Address: %s", mgmUpd.Address)
	log.Printf("UpdateMgM Owner: %s", mgmUpd.Owner)
	log.Printf("UpdateMgM TokenAddress: %s", mgmUpd.TokenAddress)
	log.Printf("UpdateMgM FaucetAddress: %s", mgmUpd.FaucetAddress)
	log.Printf("UpdateMgM StartAt: %s", mgmUpd.StartAt.String())
	log.Printf("UpdateMgM ExpireAt: %s", mgmUpd.ExpireAt.String())
	log.Printf("UpdateMgM Paused: %v", mgmUpd.Paused)
	log.Printf("UpdateMgM Hash: %s", mgmUpd.Hash)

	// Pause / Unpause (owner)
	pauseOut, err := c.PauseMgM(mgmAddress, true)
	if err != nil {
		t.Fatalf("PauseMgM: %v", err)
	}
	if len(pauseOut.States) == 0 || pauseOut.States[0].Object == nil {
		t.Fatalf("PauseMgM returned empty/nil state")
	}
	var mgmPaused memberGetMemberV1Models.MgMStateModel
	unmarshalState(t, pauseOut.States[0].Object, &mgmPaused)
	if mgmPaused.Address != mgmAddress {
		t.Fatalf("PauseMgM address mismatch: got %q want %q", mgmPaused.Address, mgmAddress)
	}
	if !mgmPaused.Paused {
		t.Fatalf("PauseMgM expected Paused=true")
	}

	log.Printf("PauseMgM Output States: %+v", pauseOut.States)
	log.Printf("PauseMgM Output Logs: %+v", pauseOut.Logs)
	log.Printf("PauseMgM Output Delegated Call: %+v", pauseOut.DelegatedCall)
	log.Printf("PauseMgM Address: %s", mgmPaused.Address)
	log.Printf("PauseMgM Paused: %v", mgmPaused.Paused)

	unpauseOut, err := c.UnpauseMgM(mgmAddress, false)
	if err != nil {
		t.Fatalf("UnpauseMgM: %v", err)
	}
	if len(unpauseOut.States) == 0 || unpauseOut.States[0].Object == nil {
		t.Fatalf("UnpauseMgM returned empty/nil state")
	}
	var mgmUnpaused memberGetMemberV1Models.MgMStateModel
	unmarshalState(t, unpauseOut.States[0].Object, &mgmUnpaused)
	if mgmUnpaused.Address != mgmAddress {
		t.Fatalf("UnpauseMgM address mismatch: got %q want %q", mgmUnpaused.Address, mgmAddress)
	}
	if mgmUnpaused.Paused {
		t.Fatalf("UnpauseMgM expected Paused=false")
	}

	log.Printf("UnpauseMgM Output States: %+v", unpauseOut.States)
	log.Printf("UnpauseMgM Output Logs: %+v", unpauseOut.Logs)
	log.Printf("UnpauseMgM Output Delegated Call: %+v", unpauseOut.DelegatedCall)
	log.Printf("UnpauseMgM Address: %s", mgmUnpaused.Address)
	log.Printf("UnpauseMgM Paused: %v", mgmUnpaused.Paused)

	// --------------------------------------------------------------------
	// Inviter lifecycle
	// --------------------------------------------------------------------
	inviter, inviterPriv := createWallet(t, c)
	invited, _ := createWallet(t, c)

	c.SetPrivateKey(ownerPriv)

	addInvOut, err := c.AddInviterMember(mgmAddress, inviter.PublicKey, "pw1")
	if err != nil {
		t.Fatalf("AddInviterMember: %v", err)
	}
	if len(addInvOut.States) == 0 || addInvOut.States[0].Object == nil {
		t.Fatalf("AddInviterMember returned empty/nil state")
	}

	var inv memberGetMemberV1Models.InviterMemberStateModel
	unmarshalState(t, addInvOut.States[0].Object, &inv)

	if inv.MgmAddress != mgmAddress {
		t.Fatalf("AddInviterMember mgm_address mismatch: got %q want %q", inv.MgmAddress, mgmAddress)
	}
	if inv.InviterAddress != inviter.PublicKey {
		t.Fatalf("AddInviterMember inviter_address mismatch: got %q want %q", inv.InviterAddress, inviter.PublicKey)
	}
	if inv.PasswordHash == "" {
		t.Fatalf("AddInviterMember password_hash empty")
	}

	log.Printf("AddInviterMember Output States: %+v", addInvOut.States)
	log.Printf("AddInviterMember Output Logs: %+v", addInvOut.Logs)
	log.Printf("AddInviterMember Output Delegated Call: %+v", addInvOut.DelegatedCall)

	log.Printf("AddInviterMember MgmAddress: %s", inv.MgmAddress)
	log.Printf("AddInviterMember InviterAddress: %s", inv.InviterAddress)
	log.Printf("AddInviterMember PasswordHash: %s", inv.PasswordHash)

	updPwOut, err := c.UpdateInviterPassword(mgmAddress, inviter.PublicKey, "pw2")
	if err != nil {
		t.Fatalf("UpdateInviterPassword: %v", err)
	}
	if len(updPwOut.States) == 0 || updPwOut.States[0].Object == nil {
		t.Fatalf("UpdateInviterPassword returned empty/nil state")
	}

	var invUpd memberGetMemberV1Models.InviterMemberStateModel
	unmarshalState(t, updPwOut.States[0].Object, &invUpd)

	if invUpd.MgmAddress != mgmAddress {
		t.Fatalf("UpdateInviterPassword mgm_address mismatch: got %q want %q", invUpd.MgmAddress, mgmAddress)
	}
	if invUpd.InviterAddress != inviter.PublicKey {
		t.Fatalf("UpdateInviterPassword inviter_address mismatch: got %q want %q", invUpd.InviterAddress, inviter.PublicKey)
	}
	if invUpd.PasswordHash == "" {
		t.Fatalf("UpdateInviterPassword password_hash empty")
	}
	if invUpd.PasswordHash == inv.PasswordHash {
		t.Fatalf("UpdateInviterPassword expected password hash to change")
	}

	log.Printf("UpdateInviterPassword Output States: %+v", updPwOut.States)
	log.Printf("UpdateInviterPassword Output Logs: %+v", updPwOut.Logs)
	log.Printf("UpdateInviterPassword Output Delegated Call: %+v", updPwOut.DelegatedCall)

	log.Printf("UpdateInviterPassword MgmAddress: %s", invUpd.MgmAddress)
	log.Printf("UpdateInviterPassword InviterAddress: %s", invUpd.InviterAddress)
	log.Printf("UpdateInviterPassword PasswordHash: %s", invUpd.PasswordHash)

	// --------------------------------------------------------------------
	// Getters (com unmarshal + validate + log)
	// --------------------------------------------------------------------
	getMgmOut, err := c.GetMgM(mgmAddress)
	if err != nil {
		t.Fatalf("GetMgM: %v", err)
	}
	if len(getMgmOut.States) == 0 || getMgmOut.States[0].Object == nil {
		t.Fatalf("GetMgM returned empty/nil state")
	}

	var mgmGet memberGetMemberV1Models.MgMStateModel
	unmarshalState(t, getMgmOut.States[0].Object, &mgmGet)
	if mgmGet.Address != mgmAddress {
		t.Fatalf("GetMgM address mismatch: got %q want %q", mgmGet.Address, mgmAddress)
	}

	log.Printf("GetMgM Output States: %+v", getMgmOut.States)
	log.Printf("GetMgM Output Logs: %+v", getMgmOut.Logs)
	log.Printf("GetMgM Output Delegated Call: %+v", getMgmOut.DelegatedCall)

	log.Printf("GetMgM Address: %s", mgmGet.Address)
	log.Printf("GetMgM Owner: %s", mgmGet.Owner)
	log.Printf("GetMgM TokenAddress: %s", mgmGet.TokenAddress)
	log.Printf("GetMgM FaucetAddress: %s", mgmGet.FaucetAddress)
	log.Printf("GetMgM Paused: %v", mgmGet.Paused)
	log.Printf("GetMgM Hash: %s", mgmGet.Hash)

	getInvOut, err := c.GetInviterMember(mgmAddress, inviter.PublicKey)
	if err != nil {
		t.Fatalf("GetInviterMember: %v", err)
	}
	if len(getInvOut.States) == 0 || getInvOut.States[0].Object == nil {
		t.Fatalf("GetInviterMember returned empty/nil state")
	}

	var invGet memberGetMemberV1Models.InviterMemberStateModel
	unmarshalState(t, getInvOut.States[0].Object, &invGet)
	if invGet.MgmAddress != mgmAddress {
		t.Fatalf("GetInviterMember mgm_address mismatch: got %q want %q", invGet.MgmAddress, mgmAddress)
	}
	if invGet.InviterAddress != inviter.PublicKey {
		t.Fatalf("GetInviterMember inviter_address mismatch: got %q want %q", invGet.InviterAddress, inviter.PublicKey)
	}
	if invGet.PasswordHash == "" {
		t.Fatalf("GetInviterMember password_hash empty")
	}

	log.Printf("GetInviterMember Output States: %+v", getInvOut.States)
	log.Printf("GetInviterMember Output Logs: %+v", getInvOut.Logs)
	log.Printf("GetInviterMember Output Delegated Call: %+v", getInvOut.DelegatedCall)

	log.Printf("GetInviterMember MgmAddress: %s", invGet.MgmAddress)
	log.Printf("GetInviterMember InviterAddress: %s", invGet.InviterAddress)
	log.Printf("GetInviterMember PasswordHash: %s", invGet.PasswordHash)

	// --------------------------------------------------------------------
	// Wait until start (para ações que exigem ativação)
	// --------------------------------------------------------------------
	waitUntil(t, 10*time.Second, func() bool { return time.Now().After(start) })

	// --------------------------------------------------------------------
	// Claim reward (inviter)
	// --------------------------------------------------------------------
	c.SetPrivateKey(inviterPriv)

	claimOut, err := c.ClaimReward(mgmAddress, invited.PublicKey, "pw2")
	if err != nil {
		t.Fatalf("ClaimReward: %v", err)
	}
	// (não assumo shape do state aqui; apenas valido envelope e log)
	if len(claimOut.States) == 0 || claimOut.States[0].Object == nil {
		t.Fatalf("ClaimReward returned empty/nil state")
	}

	log.Printf("ClaimReward Output States: %+v", claimOut.States)
	log.Printf("ClaimReward Output Logs: %+v", claimOut.Logs)
	log.Printf("ClaimReward Output Delegated Call: %+v", claimOut.DelegatedCall)

	// --------------------------------------------------------------------
	// Claim snapshots (inviter + invited)
	// --------------------------------------------------------------------
	getClaimInvOut, err := c.GetClaimInviter(mgmAddress, inviter.PublicKey)
	if err != nil {
		t.Fatalf("GetClaimInviter: %v", err)
	}
	if len(getClaimInvOut.States) == 0 || getClaimInvOut.States[0].Object == nil {
		t.Fatalf("GetClaimInviter returned empty/nil state")
	}

	var claimInv memberGetMemberV1Models.ClaimStateModel
	unmarshalState(t, getClaimInvOut.States[0].Object, &claimInv)

	if claimInv.MgmAddress != mgmAddress {
		t.Fatalf("GetClaimInviter mgm_address mismatch: got %q want %q", claimInv.MgmAddress, mgmAddress)
	}
	if claimInv.InviterAddress != inviter.PublicKey {
		t.Fatalf("GetClaimInviter inviter_address mismatch: got %q want %q", claimInv.InviterAddress, inviter.PublicKey)
	}

	log.Printf("GetClaimInviter Output States: %+v", getClaimInvOut.States)
	log.Printf("GetClaimInviter Output Logs: %+v", getClaimInvOut.Logs)
	log.Printf("GetClaimInviter Output Delegated Call: %+v", getClaimInvOut.DelegatedCall)

	log.Printf("GetClaimInviter MgmAddress: %s", claimInv.MgmAddress)
	log.Printf("GetClaimInviter InviterAddress: %s", claimInv.InviterAddress)
	log.Printf("GetClaimInviter InvitedAddress: %s", claimInv.InvitedAddress)

	getClaimInvitedOut, err := c.GetClaimInvited(mgmAddress, invited.PublicKey)
	if err != nil {
		t.Fatalf("GetClaimInvited: %v", err)
	}
	if len(getClaimInvitedOut.States) == 0 || getClaimInvitedOut.States[0].Object == nil {
		t.Fatalf("GetClaimInvited returned empty/nil state")
	}

	var claimInvited memberGetMemberV1Models.ClaimStateModel
	unmarshalState(t, getClaimInvitedOut.States[0].Object, &claimInvited)

	if claimInvited.MgmAddress != mgmAddress {
		t.Fatalf("GetClaimInvited mgm_address mismatch: got %q want %q", claimInvited.MgmAddress, mgmAddress)
	}
	if claimInvited.InvitedAddress != invited.PublicKey {
		t.Fatalf("GetClaimInvited invited_address mismatch: got %q want %q", claimInvited.InvitedAddress, invited.PublicKey)
	}

	log.Printf("GetClaimInvited Output States: %+v", getClaimInvitedOut.States)
	log.Printf("GetClaimInvited Output Logs: %+v", getClaimInvitedOut.Logs)
	log.Printf("GetClaimInvited Output Delegated Call: %+v", getClaimInvitedOut.DelegatedCall)

	log.Printf("GetClaimInvited MgmAddress: %s", claimInvited.MgmAddress)
	log.Printf("GetClaimInvited InviterAddress: %s", claimInvited.InviterAddress)
	log.Printf("GetClaimInvited InvitedAddress: %s", claimInvited.InvitedAddress)

	// --------------------------------------------------------------------
	// Delete inviter (cleanup)
	// --------------------------------------------------------------------
	delOut, err := c.DeleteInviterMember(mgmAddress, inviter.PublicKey)
	if err != nil {
		t.Fatalf("DeleteInviterMember: %v", err)
	}
	if len(delOut.States) == 0 || delOut.States[0].Object == nil {
		t.Fatalf("DeleteInviterMember returned empty/nil state")
	}

	log.Printf("DeleteInviterMember Output States: %+v", delOut.States)
	log.Printf("DeleteInviterMember Output Logs: %+v", delOut.Logs)
	log.Printf("DeleteInviterMember Output Delegated Call: %+v", delOut.DelegatedCall)
}

func TestMgMFlow_NonFungible(t *testing.T) {
	c := setupClient(t)

	// --------------------------------------------------------------------
	// Owner + base token (NFT)
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
	if tok.TokenType != tokenV1Domain.NON_FUNGIBLE {
		t.Fatalf("Token type mismatch: got %s want %s", tok.TokenType, tokenV1Domain.NON_FUNGIBLE)
	}

	log.Printf("Token Address: %s", tok.Address)
	log.Printf("Token Type: %s", tok.TokenType)

	// --------------------------------------------------------------------
	// Mint NFT (owner) - padrão mint
	// --------------------------------------------------------------------
	mintOut, err := c.MintToken(tok.Address, owner.PublicKey, "1", dec, tok.TokenType)
	if err != nil {
		t.Fatalf("MintToken NFT: %v", err)
	}
	if len(mintOut.States) == 0 {
		t.Fatalf("MintToken NFT returned empty States")
	}
	if mintOut.States[0].Object == nil {
		t.Fatalf("MintToken NFT returned nil state object")
	}

	var mint tokenV1Domain.Mint
	unmarshalState(t, mintOut.States[0].Object, &mint)

	if mint.TokenAddress != tok.Address {
		t.Fatalf("Mint TokenAddress mismatch: got %s want %s", mint.TokenAddress, tok.Address)
	}
	if mint.MintTo != owner.PublicKey {
		t.Fatalf("Mint ToAddress mismatch: got %s want %s", mint.MintTo, owner.PublicKey)
	}
	if mint.TokenType != tok.TokenType {
		t.Fatalf("Mint TokenType mismatch: got %s want %s", mint.TokenType, tok.TokenType)
	}
	if len(mint.TokenUUIDList) != 1 {
		t.Fatalf("expected 1 NFT uuid, got %d", len(mint.TokenUUIDList))
	}
	nftUUID := mint.TokenUUIDList[0]
	if nftUUID == "" {
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

	// --------------------------------------------------------------------
	// Deploy MgM + Faucet contracts
	// --------------------------------------------------------------------
	var contractState models.ContractStateModel

	deployedMgm, err := c.DeployContract1(mgmV1.MEMBER_GET_MEMBER_CONTRACT_V1)
	if err != nil {
		t.Fatalf("DeployContract (mgm): %v", err)
	}
	if len(deployedMgm.States) == 0 || deployedMgm.States[0].Object == nil {
		t.Fatalf("DeployContract (mgm) returned empty/nil state")
	}
	unmarshalState(t, deployedMgm.States[0].Object, &contractState)
	mgmAddress := contractState.Address
	if mgmAddress == "" {
		t.Fatalf("mgmAddress empty")
	}

	log.Printf("DeployContract(mgm) Output States: %+v", deployedMgm.States)
	log.Printf("DeployContract(mgm) Output Logs: %+v", deployedMgm.Logs)
	log.Printf("DeployContract(mgm) Output Delegated Call: %+v", deployedMgm.DelegatedCall)
	log.Printf("MgM Contract Address: %s", mgmAddress)

	deployedFaucet, err := c.DeployContract1(faucetV1.FAUCET_CONTRACT_V1)
	if err != nil {
		t.Fatalf("DeployContract (faucet): %v", err)
	}
	if len(deployedFaucet.States) == 0 || deployedFaucet.States[0].Object == nil {
		t.Fatalf("DeployContract (faucet) returned empty/nil state")
	}
	unmarshalState(t, deployedFaucet.States[0].Object, &contractState)
	faucetAddress := contractState.Address
	if faucetAddress == "" {
		t.Fatalf("faucetAddress empty")
	}

	log.Printf("DeployContract(faucet) Output States: %+v", deployedFaucet.States)
	log.Printf("DeployContract(faucet) Output Logs: %+v", deployedFaucet.Logs)
	log.Printf("DeployContract(faucet) Output Delegated Call: %+v", deployedFaucet.DelegatedCall)
	log.Printf("Faucet Contract Address: %s", faucetAddress)

	// --------------------------------------------------------------------
	// Faucet setup (owner)
	// --------------------------------------------------------------------
	start := time.Now().Add(2 * time.Second)
	exp := time.Now().Add(1 * time.Hour)

	faucetOut, err := c.AddFaucet(
		faucetAddress,
		owner.PublicKey,
		tok.Address,
		start,
		exp,
		false,
		1,   // max claims (NFT)
		"1", // amount (NFT sempre 1)
		2*time.Second,
	)
	if err != nil {
		t.Fatalf("AddFaucet NFT: %v", err)
	}
	if len(faucetOut.States) == 0 || faucetOut.States[0].Object == nil {
		t.Fatalf("AddFaucet NFT returned empty/nil state")
	}

	var f faucetV1Models.FaucetStateModel
	unmarshalState(t, faucetOut.States[0].Object, &f)

	if f.Address == "" {
		t.Fatalf("AddFaucet NFT returned empty address")
	}
	if f.Owner != owner.PublicKey {
		t.Fatalf("AddFaucet NFT Owner mismatch: got %q want %q", f.Owner, owner.PublicKey)
	}
	if f.TokenAddress != tok.Address {
		t.Fatalf("AddFaucet NFT TokenAddress mismatch: got %q want %q", f.TokenAddress, tok.Address)
	}
	if f.ClaimAmount != "1" {
		t.Fatalf("AddFaucet NFT ClaimAmount mismatch: got %q want %q", f.ClaimAmount, "1")
	}

	log.Printf("AddFaucet NFT Output States: %+v", faucetOut.States)
	log.Printf("AddFaucet NFT Output Logs: %+v", faucetOut.Logs)
	log.Printf("AddFaucet NFT Output Delegated Call: %+v", faucetOut.DelegatedCall)

	log.Printf("AddFaucet Address: %s", f.Address)
	log.Printf("AddFaucet Owner: %s", f.Owner)
	log.Printf("AddFaucet TokenAddress: %s", f.TokenAddress)
	log.Printf("AddFaucet ClaimAmount: %s", f.ClaimAmount)

	// --------------------------------------------------------------------
	// Add MgM (owner)
	// --------------------------------------------------------------------
	addOut, err := c.AddMgM(
		mgmAddress,
		owner.PublicKey,
		tok.Address,
		faucetAddress,
		"1", // reward amount (NFT)
		start,
		exp,
		false,
	)
	if err != nil {
		t.Fatalf("AddMgM NFT: %v", err)
	}
	if len(addOut.States) == 0 || addOut.States[0].Object == nil {
		t.Fatalf("AddMgM NFT returned empty/nil state")
	}

	var mgm memberGetMemberV1Models.MgMStateModel
	unmarshalState(t, addOut.States[0].Object, &mgm)

	if mgm.Address != mgmAddress {
		t.Fatalf("AddMgM NFT address mismatch: got %q want %q", mgm.Address, mgmAddress)
	}
	if mgm.Owner != owner.PublicKey {
		t.Fatalf("AddMgM NFT owner mismatch: got %q want %q", mgm.Owner, owner.PublicKey)
	}
	if mgm.TokenAddress != tok.Address {
		t.Fatalf("AddMgM NFT token_address mismatch: got %q want %q", mgm.TokenAddress, tok.Address)
	}
	if mgm.FaucetAddress != faucetAddress {
		t.Fatalf("AddMgM NFT faucet_address mismatch: got %q want %q", mgm.FaucetAddress, faucetAddress)
	}
	if mgm.Hash == "" {
		t.Fatalf("AddMgM NFT hash empty")
	}

	log.Printf("AddMgM NFT Output States: %+v", addOut.States)
	log.Printf("AddMgM NFT Output Logs: %+v", addOut.Logs)
	log.Printf("AddMgM NFT Output Delegated Call: %+v", addOut.DelegatedCall)

	log.Printf("AddMgM Address: %s", mgm.Address)
	log.Printf("AddMgM Owner: %s", mgm.Owner)
	log.Printf("AddMgM TokenAddress: %s", mgm.TokenAddress)
	log.Printf("AddMgM FaucetAddress: %s", mgm.FaucetAddress)
	log.Printf("AddMgM Hash: %s", mgm.Hash)

	// --------------------------------------------------------------------
	// Allowlist (token) (validate + log)
	// --------------------------------------------------------------------
	allowOut, err := c.AllowUsers(tok.Address, map[string]bool{
		mgmAddress:      true,
		faucetAddress:   true,
		owner.PublicKey: true,
	})
	if err != nil {
		t.Fatalf("AllowUsers(token): %v", err)
	}
	if len(allowOut.States) == 0 || allowOut.States[0].Object == nil {
		t.Fatalf("AllowUsers(token) returned empty/nil state")
	}

	var ap tokenV1Domain.AccessPolicy
	unmarshalState(t, allowOut.States[0].Object, &ap)
	if ap.Users == nil || !ap.Users[mgmAddress] || !ap.Users[faucetAddress] || !ap.Users[owner.PublicKey] {
		t.Fatalf("AllowUsers(token) missing expected users")
	}

	log.Printf("AllowUsers(token) Output States: %+v", allowOut.States)
	log.Printf("AllowUsers(token) Output Logs: %+v", allowOut.Logs)
	log.Printf("AllowUsers(token) Output Delegated Call: %+v", allowOut.DelegatedCall)

	// --------------------------------------------------------------------
	// Deposit NFT into Faucet (owner)
	// --------------------------------------------------------------------
	depOut, err := c.DepositFunds(f.Address, tok.Address, "1", tokenType, nftUUID)
	if err != nil {
		t.Fatalf("DepositFunds NFT: %v", err)
	}
	if len(depOut.States) == 0 || depOut.States[0].Object == nil {
		t.Fatalf("DepositFunds NFT returned empty/nil state")
	}

	var fAfterDep faucetV1Models.FaucetStateModel
	unmarshalState(t, depOut.States[0].Object, &fAfterDep)
	if fAfterDep.Address != f.Address {
		t.Fatalf("DepositFunds NFT faucet address mismatch: got %q want %q", fAfterDep.Address, f.Address)
	}

	log.Printf("DepositFunds NFT Output States: %+v", depOut.States)
	log.Printf("DepositFunds NFT Output Logs: %+v", depOut.Logs)
	log.Printf("DepositFunds NFT Output Delegated Call: %+v", depOut.DelegatedCall)

	// --------------------------------------------------------------------
	// Update MgM (owner)
	// --------------------------------------------------------------------
	newStart := time.Now().Add(3 * time.Minute)
	newExp := time.Now().Add(2 * time.Hour)

	updOut, err := c.UpdateMgM(mgmAddress, "1", newStart, newExp)
	if err != nil {
		t.Fatalf("UpdateMgM NFT: %v", err)
	}
	if len(updOut.States) == 0 || updOut.States[0].Object == nil {
		t.Fatalf("UpdateMgM NFT returned empty/nil state")
	}

	var mgmUpd memberGetMemberV1Models.MgMStateModel
	unmarshalState(t, updOut.States[0].Object, &mgmUpd)
	if mgmUpd.Address != mgmAddress {
		t.Fatalf("UpdateMgM NFT address mismatch: got %q want %q", mgmUpd.Address, mgmAddress)
	}

	log.Printf("UpdateMgM NFT Output States: %+v", updOut.States)
	log.Printf("UpdateMgM NFT Output Logs: %+v", updOut.Logs)
	log.Printf("UpdateMgM NFT Output Delegated Call: %+v", updOut.DelegatedCall)

	// Pause / Unpause
	pauseOut, err := c.PauseMgM(mgmAddress, true)
	if err != nil {
		t.Fatalf("PauseMgM: %v", err)
	}
	if len(pauseOut.States) == 0 || pauseOut.States[0].Object == nil {
		t.Fatalf("PauseMgM returned empty/nil state")
	}
	var mgmPaused memberGetMemberV1Models.MgMStateModel
	unmarshalState(t, pauseOut.States[0].Object, &mgmPaused)
	if !mgmPaused.Paused {
		t.Fatalf("PauseMgM expected Paused=true")
	}

	log.Printf("PauseMgM Output States: %+v", pauseOut.States)
	log.Printf("PauseMgM Output Logs: %+v", pauseOut.Logs)
	log.Printf("PauseMgM Output Delegated Call: %+v", pauseOut.DelegatedCall)

	unpauseOut, err := c.UnpauseMgM(mgmAddress, false)
	if err != nil {
		t.Fatalf("UnpauseMgM: %v", err)
	}
	if len(unpauseOut.States) == 0 || unpauseOut.States[0].Object == nil {
		t.Fatalf("UnpauseMgM returned empty/nil state")
	}
	var mgmUnpaused memberGetMemberV1Models.MgMStateModel
	unmarshalState(t, unpauseOut.States[0].Object, &mgmUnpaused)
	if mgmUnpaused.Paused {
		t.Fatalf("UnpauseMgM expected Paused=false")
	}

	log.Printf("UnpauseMgM Output States: %+v", unpauseOut.States)
	log.Printf("UnpauseMgM Output Logs: %+v", unpauseOut.Logs)
	log.Printf("UnpauseMgM Output Delegated Call: %+v", unpauseOut.DelegatedCall)

	// --------------------------------------------------------------------
	// Inviter lifecycle + getters + claim + snapshots + delete
	// --------------------------------------------------------------------
	inviter, inviterPriv := createWallet(t, c)
	invited, _ := createWallet(t, c)

	c.SetPrivateKey(ownerPriv)

	addInvOut, err := c.AddInviterMember(mgmAddress, inviter.PublicKey, "pw1")
	if err != nil {
		t.Fatalf("AddInviterMember: %v", err)
	}
	if len(addInvOut.States) == 0 || addInvOut.States[0].Object == nil {
		t.Fatalf("AddInviterMember returned empty/nil state")
	}

	updPwOut, err := c.UpdateInviterPassword(mgmAddress, inviter.PublicKey, "pw2")
	if err != nil {
		t.Fatalf("UpdateInviterPassword: %v", err)
	}
	if len(updPwOut.States) == 0 || updPwOut.States[0].Object == nil {
		t.Fatalf("UpdateInviterPassword returned empty/nil state")
	}

	getMgmOut, err := c.GetMgM(mgmAddress)
	if err != nil {
		t.Fatalf("GetMgM: %v", err)
	}
	if len(getMgmOut.States) == 0 || getMgmOut.States[0].Object == nil {
		t.Fatalf("GetMgM returned empty/nil state")
	}

	getInvOut, err := c.GetInviterMember(mgmAddress, inviter.PublicKey)
	if err != nil {
		t.Fatalf("GetInviterMember: %v", err)
	}
	if len(getInvOut.States) == 0 || getInvOut.States[0].Object == nil {
		t.Fatalf("GetInviterMember returned empty/nil state")
	}

	waitUntil(t, 10*time.Second, func() bool { return time.Now().After(start) })

	c.SetPrivateKey(inviterPriv)
	claimOut, err := c.ClaimReward(mgmAddress, invited.PublicKey, "pw2")
	if err != nil {
		t.Fatalf("ClaimReward NFT: %v", err)
	}
	if len(claimOut.States) == 0 || claimOut.States[0].Object == nil {
		t.Fatalf("ClaimReward NFT returned empty/nil state")
	}

	log.Printf("ClaimReward NFT Output States: %+v", claimOut.States)
	log.Printf("ClaimReward NFT Output Logs: %+v", claimOut.Logs)
	log.Printf("ClaimReward NFT Output Delegated Call: %+v", claimOut.DelegatedCall)

	getClaimInvOut, err := c.GetClaimInviter(mgmAddress, inviter.PublicKey)
	if err != nil {
		t.Fatalf("GetClaimInviter: %v", err)
	}
	if len(getClaimInvOut.States) == 0 || getClaimInvOut.States[0].Object == nil {
		t.Fatalf("GetClaimInviter returned empty/nil state")
	}

	getClaimInvitedOut, err := c.GetClaimInvited(mgmAddress, invited.PublicKey)
	if err != nil {
		t.Fatalf("GetClaimInvited: %v", err)
	}
	if len(getClaimInvitedOut.States) == 0 || getClaimInvitedOut.States[0].Object == nil {
		t.Fatalf("GetClaimInvited returned empty/nil state")
	}

	c.SetPrivateKey(ownerPriv)
	delOut, err := c.DeleteInviterMember(mgmAddress, inviter.PublicKey)
	if err != nil {
		t.Fatalf("DeleteInviterMember: %v", err)
	}
	if len(delOut.States) == 0 || delOut.States[0].Object == nil {
		t.Fatalf("DeleteInviterMember returned empty/nil state")
	}

	log.Printf("DeleteInviterMember Output States: %+v", delOut.States)
	log.Printf("DeleteInviterMember Output Logs: %+v", delOut.Logs)
	log.Printf("DeleteInviterMember Output Delegated Call: %+v", delOut.DelegatedCall)
}
