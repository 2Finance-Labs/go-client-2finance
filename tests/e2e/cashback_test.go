package e2e_test


import (
	"testing"
	"time"
	cashbackV1Domain "gitlab.com/2finance/2finance-network/blockchain/contract/cashbackV1/domain"
	"gitlab.com/2finance/2finance-network/blockchain/contract/contractV1/models"
	"gitlab.com/2finance/2finance-network/blockchain/contract/cashbackV1"
)


func TestCashbackFlow(t *testing.T) {
	c := setupClient(t)
	
	owner, ownerPriv := createWallet(t, c)
	c.SetPrivateKey(ownerPriv)


	dec := 1
	tok := createBasicToken(t, c, owner.PublicKey, dec, false)
	_ = createMint(t, c, tok, owner.PublicKey, "10000", dec)

	merchant, _ := createWallet(t, c)

	c.SetPrivateKey(ownerPriv)
	_ = createTransfer(t, c, tok, merchant.PublicKey, "50", dec)

	start := time.Now().Add(2 * time.Second)
	exp := time.Now().Add(30 * time.Minute)

	contractState := models.ContractStateModel{}
	deployedContract, err := c.DeployContract1(cashbackV1.CASHBACK_CONTRACT_V1)
	if err != nil { t.Fatalf("DeployContract: %v", err) }
	unmarshalState(t, deployedContract.States[0].Object, &contractState)
	address := contractState.Address

	out, err := c.AddCashback(address, merchant.PublicKey, tok.Address, cashbackV1Domain.PROGRAM_TYPE_FIXED, "250", start, exp, false)
	if err != nil { t.Fatalf("AddCashback: %v", err) }
	var cb cashbackV1Domain.Cashback
	unmarshalState(t, out.States[0].Object, &cb)
	if cb.Address == "" { t.Fatalf("cashback addr empty") }


	_, _ = c.AllowUsers(tok.Address, map[string]bool{cb.Address: true})
	if _, err := c.DepositCashbackFunds(cb.Address, tok.Address, amt(1000, dec)); err != nil { t.Fatalf("DepositCashbackFunds: %v", err) }
	if _, err := c.UpdateCashback(cb.Address, tok.Address, cashbackV1Domain.PROGRAM_TYPE_FIXED, "300", start, exp); err != nil { t.Fatalf("UpdateCashback: %v", err) }
	_, _ = c.PauseCashback(cb.Address, true)
	_, _ = c.UnpauseCashback(cb.Address, false)

	time.Sleep(2 * time.Second)
	// claim as user (best-effort)
	user, userPriv := createWallet(t, c)
	c.SetPrivateKey(ownerPriv)
	_, err = c.AllowUsers(tok.Address, map[string]bool{user.PublicKey: true})
	if err != nil { t.Fatalf("AllowUsers: %v", err) }

	c.SetPrivateKey(userPriv)
	if _, err := c.ClaimCashback(cb.Address, amt(100, dec)); err != nil { t.Fatalf("ClaimCashback warning: %v", err) }


	// getters
	if _, err := c.GetCashback(cb.Address); err != nil { t.Fatalf("GetCashback: %v", err) }
	if _, err := c.ListCashbacks(merchant.PublicKey, tok.Address, "", false, 1, 10, true); err != nil { t.Fatalf("ListCashbacks: %v", err) }
}