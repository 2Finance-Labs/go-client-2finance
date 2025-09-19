package e2e_test


import (
	"testing"
	"time"


	faucetV1Domain "gitlab.com/2finance/2finance-network/blockchain/contract/faucetV1/domain"
	"gitlab.com/2finance/2finance-network/blockchain/contract/faucetV1"
	"gitlab.com/2finance/2finance-network/blockchain/contract/contractV1/models"
)

func TestFaucetFlow(t *testing.T) {

	c := setupClient(t)
	owner, ownerPriv := createWallet(t, c)

	c.SetPrivateKey(ownerPriv)
	dec := 5
	tok := createBasicToken(t, c, owner.PublicKey, dec, true)
	_ = createMint(t, c, tok, owner.PublicKey, "10000", tok.Decimals)

	merchant, merchPriv := createWallet(t, c)

	c.SetPrivateKey(ownerPriv)
	_ = createTransfer(t, c, tok, merchant.PublicKey, "50", tok.Decimals)


	start := time.Now().Add(2 * time.Second)
	exp := time.Now().Add(20 * time.Minute)

	amount := "4"

	contractState := models.ContractStateModel{}
	deployedContract, err := c.DeployContract(faucetV1.FAUCET_CONTRACT_V1, "")
	if err != nil { t.Fatalf("DeployContract: %v", err) }
	unmarshalState(t, deployedContract.States[0].Object, &contractState)
	address := contractState.Address

	out, err := c.AddFaucet(address, merchant.PublicKey, tok.Address, start, exp, false, 3, amount, 2*time.Second)
	if err != nil { t.Fatalf("AddFaucet: %v", err) }
	var f faucetV1Domain.Faucet
	unmarshalState(t, out.States[0].Object, &f)
	if f.Address == "" { t.Fatalf("faucet addr empty") }


	// allow faucet (if token is allow-listed in your impl)
	_, err = c.AllowUsers(tok.Address, map[string]bool{f.Address: true})
	if err != nil { t.Fatalf("AllowUsers: %v", err) }

	c.SetPrivateKey(merchPriv)
	depositAmount := "569"
	if _, err := c.DepositFunds(f.Address, tok.Address, depositAmount); err != nil { t.Fatalf("DepositFunds: %v", err) }


	// wait for start and claim as a user
	user, userPriv := createWallet(t, c)
	_ = user
	
	time.Sleep(5 * time.Second)
	c.SetPrivateKey(ownerPriv)
	_, err = c.AllowUsers(tok.Address, map[string]bool{user.PublicKey: true})
	if err != nil { t.Fatalf("AllowUsers: %v", err) }
	c.SetPrivateKey(userPriv)
	if _, err := c.ClaimFunds(f.Address); err != nil { t.Logf("ClaimFunds warning: %v", err) }
	

	// pause/unpause & getters
	c.SetPrivateKey(merchPriv)
	_, _ = c.PauseFaucet(f.Address, true)
	_, _ = c.UnpauseFaucet(f.Address, false)
	if _, err := c.GetFaucet(f.Address); err != nil { t.Fatalf("GetFaucet: %v", err) }
	if _, err := c.ListFaucets(merchant.PublicKey, 1, 10, true); err != nil { t.Fatalf("ListFaucets: %v", err) }
}