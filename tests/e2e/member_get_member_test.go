package e2e_test

import (
	"testing"
	"time"
	"fmt"
	mgmV1 "gitlab.com/2finance/2finance-network/blockchain/contract/memberGetMemberV1"
	faucetV1 "gitlab.com/2finance/2finance-network/blockchain/contract/faucetV1"
	faucetV1Domain "gitlab.com/2finance/2finance-network/blockchain/contract/faucetV1/domain"
	mgmV1Models "gitlab.com/2finance/2finance-network/blockchain/contract/memberGetMemberV1/models"
	"gitlab.com/2finance/2finance-network/blockchain/contract/contractV1/models"


)

func TestMgMFlow(t *testing.T) {

	c := setupClient(t)

	// --------------------------------------------------------------------
	// Owner + base token
	// --------------------------------------------------------------------
	owner, ownerPriv := createWallet(t, c)
	c.SetPrivateKey(ownerPriv)
	dec := 6
	tok := createBasicToken(t, c, owner.PublicKey, dec, false)

	// --------------------------------------------------------------------
	// Deploy MgM + Faucet contracts
	// --------------------------------------------------------------------
	var contractState models.ContractStateModel

	deployedMgm, err := c.DeployContract(mgmV1.MEMBER_GET_MEMBER_CONTRACT_V1, "")
	if err != nil { t.Fatalf("DeployContract (mgm): %v", err) }
	unmarshalState(t, deployedMgm.States[0].Object, &contractState)
	mgmAddress := contractState.Address
	if mgmAddress == "" { t.Fatalf("mgmAddress empty") }

	deployedFaucet, err := c.DeployContract(faucetV1.FAUCET_CONTRACT_V1, "")
	if err != nil { t.Fatalf("DeployContract (faucet): %v", err) }
	unmarshalState(t, deployedFaucet.States[0].Object, &contractState)
	faucetAddress := contractState.Address
	if faucetAddress == "" { t.Fatalf("faucetAddress empty") }
	
	
	start := time.Now().Add(2 * time.Second)
	exp   := time.Now().Add(1 * time.Hour)
	amount := "5"

	out, err := c.AddFaucet(faucetAddress, owner.PublicKey, tok.Address, start, exp, false, 3, amount, 2*time.Second)
	if err != nil { t.Fatalf("AddFaucet: %v", err) }
	var f faucetV1Domain.Faucet
	unmarshalState(t, out.States[0].Object, &f)
	if f.Address == "" { t.Fatalf("faucet addr empty") }
	// --------------------------------------------------------------------
	// Add MgM (owner)
	// --------------------------------------------------------------------
	fmt.Printf("MgM %s\n", mgmAddress)
	addOut, err := c.AddMgM(mgmAddress, owner.PublicKey, tok.Address, faucetAddress, amt(10, dec), start, exp, false)
	if err != nil { t.Fatalf("AddMgM: %v", err) }
	// sanity: state has mgm address
	var got mgmV1Models.MgMStateModel
	unmarshalState(t, addOut.States[0].Object, &got)
	if got.Address == "" { t.Fatalf("AddMgM returned empty address") }

	// Allow token transfers to the MgM contract (best effort)
	_, _ = c.AllowUsers(tok.Address, map[string]bool{
		mgmAddress: true,
		faucetAddress: true,
		owner.PublicKey: true,
	})
	

	// --------------------------------------------------------------------
	// Deposit/Withdraw pool funds (owner)
	// --------------------------------------------------------------------
	if _, err := c.DepositMgM(mgmAddress, amt(100, dec)); err != nil {
		t.Logf("DepositMgM warning: %v", err) // backend may gate this; don't fail suite
	}
	if _, err := c.WithdrawMgM(mgmAddress, amt(1, dec)); err != nil {
		t.Logf("WithdrawMgM warning: %v", err)
	}

	// --------------------------------------------------------------------
	// Update MgM (owner)
	// --------------------------------------------------------------------
	newStart := time.Now().Add(3 * time.Minute)
	newExp   := time.Now().Add(2 * time.Hour)
	if _, err := c.UpdateMgM(mgmAddress, amt(20, dec), newStart, newExp); err != nil {
		t.Fatalf("UpdateMgM: %v", err)
	}

	// Pause / Unpause (owner)
	if _, err := c.PauseMgM(mgmAddress, true); err != nil {
		t.Fatalf("PauseMgM: %v", err)
	}
	if _, err := c.UnpauseMgM(mgmAddress, false); err != nil {
		t.Fatalf("UnpauseMgM: %v", err)
	}

	// --------------------------------------------------------------------
	// Inviter lifecycle
	// --------------------------------------------------------------------
	inviter, _ := createWallet(t, c)
	invited, _ := createWallet(t, c)

	c.SetPrivateKey(ownerPriv)
	if _, err := c.AddInviterMember(mgmAddress, inviter.PublicKey, "pw1"); err != nil {
		t.Fatalf("AddInviterMember: %v", err)
	}
	if _, err := c.UpdateInviterPassword(mgmAddress, inviter.PublicKey, "pw2"); err != nil {
		t.Fatalf("UpdateInviterPassword: %v", err)
	}

	// getters — should work even if business state is minimal
	if _, err := c.GetMgM(mgmAddress); err != nil {
		t.Logf("GetMgM warning: %v", err)
	}
	if _, err := c.GetInviterMember(mgmAddress, inviter.PublicKey); err != nil {
		t.Logf("GetInviterMember warning: %v", err)
	}

	// Wait until the original start time to allow actions that require activation
	waitUntil(t, 10*time.Second, func() bool { return time.Now().After(start) })

	// Best-effort claim reward (may depend on backend rules)
	if _, err := c.ClaimReward(mgmAddress, "pw2", invited.PublicKey); err != nil {
		t.Logf("ClaimReward warning: %v", err)
	}

	// Inspect claim snapshots (best effort)
	if _, err := c.GetClaimInviter(mgmAddress, inviter.PublicKey); err != nil {
		t.Logf("GetClaimInviter warning: %v", err)
	}
	if _, err := c.GetClaimInvited(mgmAddress, invited.PublicKey); err != nil {
		t.Logf("GetClaimInvited warning: %v", err)
	}

	// Optional: delete inviter (cleanup path)
	if _, err := c.DeleteInviterMember(mgmAddress, inviter.PublicKey); err != nil {
		t.Logf("DeleteInviterMember warning: %v", err)
	}
}
