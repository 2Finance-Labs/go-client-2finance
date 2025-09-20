package e2e_test


import (
	client2f "github.com/2Finance-Labs/go-client-2finance/client_2finance"
	walletDomain "gitlab.com/2finance/2finance-network/blockchain/contract/walletV1/domain"
	"gitlab.com/2finance/2finance-network/blockchain/contract/walletV1"
	"gitlab.com/2finance/2finance-network/blockchain/contract/contractV1/models"
	"testing"
)

// createWallet generates a keypair, registers the wallet and returns the parsed state + priv.
func createWallet(t *testing.T, c client2f.Client2FinanceNetwork) (walletDomain.Wallet, string) {
	t.Helper()
	pub, priv := genKey(t, c)
	c.SetPrivateKey(priv)

	contractState := models.ContractStateModel{}
	deployedContract, err := c.DeployContract(walletV1.WALLET_CONTRACT_V1, "")
	if err != nil { t.Fatalf("DeployContract: %v", err) }
	unmarshalState(t, deployedContract.States[0].Object, &contractState)	

	wOut, err := c.AddWallet(contractState.Address, pub)
	if err != nil { t.Fatalf("AddWallet: %v", err) }
	var w walletDomain.Wallet
	unmarshalState(t, wOut.States[0].Object, &w)
	if w.PublicKey == "" { t.Fatalf("wallet public key empty") }
	return w, priv
}


// import "testing"


// func TestWalletTransferOptional(t *testing.T) {
// 	c := setupClient(t)
// 	sender, senderPriv := createWallet(t, c)
// 	rec, _ := createWallet(t, c)
// 	c.SetPrivateKey(senderPriv)

// 	if _, err := c.TransferWallet(rec.PublicKey, "1", 0); err != nil {
// 		// Depending on your implementation this may be unsupported; don't fail the suite.
// 		t.Skipf("TransferWallet not enabled or failed: %v", err)
// 	}
// }