package e2e_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gitlab.com/2finance/2finance-network/blockchain/contract/contractV1/domain"
	"gitlab.com/2finance/2finance-network/blockchain/contract/walletV1"
	walletDomain "gitlab.com/2finance/2finance-network/blockchain/contract/walletV1/domain"
	walletModels "gitlab.com/2finance/2finance-network/blockchain/contract/walletV1/models"
	"gitlab.com/2finance/2finance-network/blockchain/log"
	"gitlab.com/2finance/2finance-network/blockchain/utils"
)

func TestWalletWorkflow(t *testing.T) {
	t.Helper()

	// ------------------
	//   LOCAL WALLET
	// ------------------
	wm := setupWalletManager(t)

	pub, priv := genKey(t, wm)
	importAndUnlockWallet(t, wm, pub, priv)

	c := setupClient(t, wm)

	// ------------------
	//   DEPLOY WALLET CONTRACT
	// ------------------
	deployedContract, err := c.DeployContract1(walletV1.WALLET_CONTRACT_V1)
	if err != nil {
		t.Fatalf("DeployContract: %v", err)
	}

	if len(deployedContract.Logs) == 0 {
		t.Fatalf("DeployContract returned no logs")
	}

	contractLog, err := utils.UnmarshalLog[log.Log](deployedContract.Logs[0])
	if err != nil {
		t.Fatalf("UnmarshalLog (DeployContract.Logs[0]): %v", err)
	}

	assert.Equal(t, domain.DEPLOYED_CONTRACT_LOG, contractLog.LogType, "deploy log type mismatch")

	contractDomain, err := utils.UnmarshalEvent[domain.Contract](contractLog.Event)
	if err != nil {
		t.Fatalf("UnmarshalEvent (DeployContract.Logs[0]): %v", err)
	}

	if contractDomain.Address == "" {
		t.Fatalf("contract address empty (event=%s)", string(contractLog.Event))
	}

	// ------------------
	//   ADD WALLET
	// ------------------
	wOut, err := c.AddWallet(contractDomain.Address, pub)
	if err != nil {
		t.Fatalf("AddWallet: %v", err)
	}

	if len(wOut.Logs) == 0 {
		t.Fatalf("AddWallet returned no logs")
	}

	walletLog, err := utils.UnmarshalLog[log.Log](wOut.Logs[0])
	if err != nil {
		t.Fatalf("UnmarshalLog (AddWallet.Logs[0]): %v", err)
	}

	assert.Equal(t, walletDomain.WALLET_CREATED_LOG, walletLog.LogType, "add-wallet log type mismatch")

	wallet, err := utils.UnmarshalEvent[walletDomain.Wallet](walletLog.Event)
	if err != nil {
		t.Fatalf("UnmarshalEvent (AddWallet.Logs[0]): %v", err)
	}

	assert.Equal(t, pub, wallet.PublicKey, "wallet public key mismatch")
	assert.Equal(t, contractDomain.Address, wallet.Address, "wallet address mismatch with contract address")

	// ------------------
	//   GET WALLET BY ADDRESS
	// ------------------
	wState, err := c.GetWalletByAddress(contractDomain.Address)
	if err != nil {
		t.Fatalf("GetWalletByAddress: %v", err)
	}

	if len(wState.Logs) != 0 {
		t.Fatalf("GetWalletByAddress returned logs")
	}

	if wState.States == nil || len(wState.States) == 0 {
		t.Fatalf("GetWalletByAddress returned no states")
	}

	walletState := walletModels.WalletStateModel{}

	err = utils.UnmarshalState[walletModels.WalletStateModel](wState.States[0].Object, &walletState)
	if err != nil {
		t.Fatalf("UnmarshalState (GetWalletByAddress.States[0]): %v", err)
	}

	assert.Equal(t, wallet.PublicKey, walletState.PublicKey, "wallet state public key mismatch between GetWalletByAddress and AddWallet event")
	assert.Equal(t, wallet.Address, walletState.Address, "wallet state address mismatch between GetWalletByAddress and AddWallet event")
	assert.NotEmpty(t, walletState.CreatedAt, "wallet state CreatedAt should not be empty")
	assert.NotEmpty(t, walletState.UpdatedAt, "wallet state UpdatedAt should not be empty")

	// ------------------
	//   GET WALLET BY PUBLIC KEY
	// ------------------
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

	assert.Equal(t, walletState.PublicKey, walletStateByPub.PublicKey, "wallet state public key mismatch between GetWalletByAddress and GetWalletByPublicKey")
	assert.Equal(t, walletState.Address, walletStateByPub.Address, "wallet state address mismatch between GetWalletByAddress and GetWalletByPublicKey")
	assert.NotEmpty(t, walletStateByPub.CreatedAt, "wallet state CreatedAt should not be empty")
	assert.NotEmpty(t, walletStateByPub.UpdatedAt, "wallet state UpdatedAt should not be empty")
}