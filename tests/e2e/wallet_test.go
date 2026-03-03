package e2e_test


import (
	walletDomain "gitlab.com/2finance/2finance-network/blockchain/contract/walletV1/domain"
	walletModels "gitlab.com/2finance/2finance-network/blockchain/contract/walletV1/models"
	"gitlab.com/2finance/2finance-network/blockchain/contract/walletV1"
	"gitlab.com/2finance/2finance-network/blockchain/contract/contractV1/domain"
	"gitlab.com/2finance/2finance-network/blockchain/log"
	"gitlab.com/2finance/2finance-network/blockchain/utils"
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestWalletWorkflow(t *testing.T) {
	t.Helper()
	c := setupClient(t)

	pub, priv := genKey(t, c)
	c.SetPrivateKey(priv)

	deployedContract, err := c.DeployContract1(walletV1.WALLET_CONTRACT_V1)
	if err != nil {
		t.Fatalf("DeployContract: %v", err)
	}
	if len(deployedContract.Logs) == 0 {
		t.Fatalf("DeployContract returned no logs")
	}

	// 1) Unmarshal first deploy log
	contractLog, err := utils.UnmarshalLog[log.Log](deployedContract.Logs[0])
	if err != nil {
		t.Fatalf("UnmarshalLog (DeployContract.Logs[0]): %v", err)
	}
	assert.Equal(t, contractLog.LogType, domain.DEPLOYED_CONTRACT_LOG, "deploy log type mismatch")
	// 2) Decode deploy event -> Contract
	contractDomain, err := utils.UnmarshalEvent[domain.Contract](contractLog.Event)
	if err != nil {
		t.Fatalf("UnmarshalEvent (DeployContract.Logs[0]): %v", err)
	}
	if contractDomain.Address == "" {
		t.Fatalf("contract address empty (event=%s)", string(contractLog.Event))
	}

	// 3) AddWallet
	wOut, err := c.AddWallet(contractDomain.Address, pub)
	if err != nil {
		t.Fatalf("AddWallet: %v", err)
	}
	if len(wOut.Logs) == 0 {
		t.Fatalf("AddWallet returned no logs")
	}

	// 4) Unmarshal first add-wallet log
	walletLog, err := utils.UnmarshalLog[log.Log](wOut.Logs[0])
	if err != nil {
		t.Fatalf("UnmarshalLog (AddWallet.Logs[0]): %v", err)
	}
	assert.Equal(t, walletLog.LogType, walletDomain.WALLET_CREATED_LOG, "add-wallet log type mismatch")
	// 5) Decode add-wallet event -> Wallet
	wallet, err := utils.UnmarshalEvent[walletDomain.Wallet](walletLog.Event)
	if err != nil {
		t.Fatalf("UnmarshalEvent (AddWallet.Logs[0]): %v", err)
	}
	assert.Equal(t, wallet.PublicKey, pub, "wallet public key mismatch")
	assert.Equal(t, wallet.Address, contractDomain.Address, "wallet address mismatch with contract address")
	
	// 6) GetWallet by address
	wState, err := c.GetWalletByAddress(contractDomain.Address)
	if err != nil {
		t.Fatalf("GetWallet: %v", err)
	}
	if len(wState.Logs) != 0 {
		t.Fatalf("GetWallet returned logs")
	}

	if wState.States == nil || len(wState.States) == 0 {
		t.Fatalf("GetWallet returned no states")
	}

	walletState := walletDomain.Wallet{}
	err = utils.UnmarshalState[walletDomain.Wallet](wState.States[0].Object, &walletState)
	if err != nil {
		t.Fatalf("UnmarshalState (GetWallet.States[0]): %v", err)
	}
	if walletState.PublicKey == "" {
		t.Fatalf("wallet state public key empty (state=%s)", wState.States[0].Object)
	}
	if walletState.PublicKey != wallet.PublicKey {
		t.Fatalf("wallet state public key mismatch: expected %s, got %s", wallet.PublicKey, walletState.PublicKey)
	}
	

	wStateByPub, err := c.GetWalletByPublicKey(wallet.PublicKey)
	if err != nil {
		t.Fatalf("GetWalletByPublicKey: %v", err)
	}
	if len(wStateByPub.Logs) != 0 {
		t.Fatalf("GetWalletByPublicKey returned logs")
	}
	if wStateByPub.States == nil || len(wStateByPub.States) == 0 {
		t.Fatalf("GetWalletByPublicKey returned no states")
	}

	walletStateByPub := walletModels.WalletStateModel{}
	err = utils.UnmarshalState[walletModels.WalletStateModel](wStateByPub.States[0].Object, &walletStateByPub)
	if err != nil {
		t.Fatalf("UnmarshalState (GetWalletByPublicKey.States[0]): %v", err)
	}
	assert.Equal(t, walletStateByPub.PublicKey, walletState.PublicKey, "wallet state public key mismatch between GetWallet and GetWalletByPublicKey")
	assert.Equal(t, walletStateByPub.Address, walletState.Address, "wallet state address mismatch between GetWallet and GetWalletByPublicKey")
	assert.NotEmpty(t, walletStateByPub.CreatedAt, "wallet state CreatedAt should not be empty")
	assert.NotEmpty(t, walletStateByPub.UpdatedAt, "wallet state UpdatedAt should not be empty")
}