package e2e_test

import (
	"testing"
	"time"

	client2f "github.com/2Finance-Labs/go-client-2finance/client_2finance"
	"gitlab.com/2finance/2finance-network/blockchain/contract/contractV1/domain"
	"gitlab.com/2finance/2finance-network/blockchain/contract/tokenV1"
	tokenV1Domain "gitlab.com/2finance/2finance-network/blockchain/contract/tokenV1/domain"
	"gitlab.com/2finance/2finance-network/blockchain/contract/walletV1"
	walletV1Domain "gitlab.com/2finance/2finance-network/blockchain/contract/walletV1/domain"
	"gitlab.com/2finance/2finance-network/blockchain/log"
	"gitlab.com/2finance/2finance-network/blockchain/utils"
)

func createWallet(t *testing.T, c client2f.Client2FinanceNetwork) (wallet walletV1Domain.Wallet, walletPrivateKey string) {
	t.Helper()

	pub, priv := genKey(t, c)
	c.SetPrivateKey(priv)

	deployedContract, err := c.DeployContract1(walletV1.WALLET_CONTRACT_V1)
	if err != nil {
		t.Fatalf("DeployContract: %v", err)
	}

	// 1) Unmarshal Log (obj -> Log)
	contractLog, err := utils.UnmarshalLog[log.Log](deployedContract.Logs[0])
	if err != nil {
		t.Fatalf("UnmarshalLog: %v", err)
	}

	// 2) Unmarshal Event (log.Event bytes -> domain.Contract)
	contractDomain, err := utils.UnmarshalEvent[domain.Contract](contractLog.Event)
	if err != nil {
		t.Fatalf("UnmarshalEvent: %v", err)
	}

	// Se você quiser validar:
	if contractDomain.Address == "" {
		t.Fatalf("contract address empty (event=%s)", string(contractLog.Event))
	}

	wOut, err := c.AddWallet(contractDomain.Address, pub)
	if err != nil {
		t.Fatalf("AddWallet: %v", err)
	}

	if len(wOut.Logs) == 0 {
		t.Fatalf("AddWallet returned no logs")
	}

	// tenta achar um log que contenha public_key no event
	var w walletV1Domain.Wallet

	lg, err := utils.UnmarshalLog[log.Log](wOut.Logs[0])
	if err != nil {
		t.Fatalf("UnmarshalLog (AddWallet.Logs[0]): %v", err)
	}

	// tenta decodificar o event para Wallet
	ev, err := utils.UnmarshalEvent[walletV1Domain.Wallet](lg.Event)
	if err != nil {
		t.Fatalf("UnmarshalEvent (AddWallet.Logs[0]): %v", err)
	}

	// valida se é o que queremos
	if ev.PublicKey != "" {
		w = ev
	}

	if w.PublicKey == "" {
		t.Fatalf("wallet event not found in AddWallet logs")
	}

	return w, priv
}

// createBasicToken creates a minimal token owned by ownerPub.
func createBasicToken(
	t *testing.T,
	c client2f.Client2FinanceNetwork,
	ownerPub string,
	decimals int,
	requireFee bool,
	tokenType string,
	stablecoin bool,
) tokenV1Domain.Token {
	t.Helper()

	symbol := "2F" + randSuffix(4)
	name := "2Finance"
	var totalSupply string
	if tokenType == tokenV1Domain.NON_FUNGIBLE {
		totalSupply = "1"
	} else {
		totalSupply = amt(1_000_000, decimals)
	}
	description := "e2e token created by tests"
	image := "https://example.com/image.png"
	website := "https://example.com"
	tagsSocial := map[string]string{"twitter": "https://twitter.com/2finance"}
	tagsCat := map[string]string{"category": "DeFi"}
	tags := map[string]string{"tag1": "DeFi", "tag2": "Blockchain"}
	creator := "2Finance Test"
	creatorWebsite := "https://creator.example"
	accessMode := tokenV1Domain.DENY_ACCESS_MODE
	accessUsers := map[string]bool{}
	frozenAccounts := map[string]bool{}
	feeTiers := []map[string]interface{}{}

	if requireFee {
		feeTiers = []map[string]interface{}{
			{
				"min_amount": "0",
				"max_amount": amt(10_000, decimals),
				"min_volume": "0",
				"max_volume": amt(100_000, decimals),
				"fee_bps":    50,
			},
		}
	}

	feeAddress := ownerPub
	freezeAuthorityRevoked := false
	mintAuthorityRevoked := false
	updateAuthorityRevoked := false
	paused := false
	expiredAt := time.Time{}

	deployedContract, err := c.DeployContract1(tokenV1.TOKEN_CONTRACT_V1)
	if err != nil {
		t.Fatalf("DeployContract: %v", err)
	}
	contractLog, err := utils.UnmarshalLog[log.Log](deployedContract.Logs[0])
	if err != nil {
		t.Fatalf("UnmarshalLog (AddWallet.Logs[0]): %v", err)
	}
	address := contractLog.ContractAddress

	assetGLBUri := "https://example.com/asset.glb"
	transferable := true

	out, err := c.AddToken(
		address,
		symbol,
		name,
		decimals,
		totalSupply,
		description,
		ownerPub,
		image,
		website,
		tagsSocial,
		tagsCat,
		tags,
		creator,
		creatorWebsite,
		accessMode,
		accessUsers,
		frozenAccounts,
		feeTiers,
		feeAddress,
		freezeAuthorityRevoked,
		mintAuthorityRevoked,
		updateAuthorityRevoked,
		paused,
		expiredAt,
		assetGLBUri,
		tokenType,
		transferable,
		stablecoin,
	)
	if err != nil {
		t.Fatalf("AddToken: %v", err)
	}

	var tok tokenV1Domain.Token
	unmarshalLog, err := utils.UnmarshalLog[log.Log](out.Logs[0])
	if err != nil {
		t.Fatalf("UnmarshalLog (AddToken.Logs[0]): %v", err)
	}
	unmarshalEvent, err := utils.UnmarshalEvent[tokenV1Domain.Token](unmarshalLog.Event)
	if err != nil {
		t.Fatalf("UnmarshalEvent (AddToken.Logs[0]): %v", err)
	}
	tok = unmarshalEvent

	if tok.Address == "" {
		t.Fatalf("token address empty (event=%s)", string(unmarshalLog.Event))
	}
	return tok
}

func createMintFT(t *testing.T, c client2f.Client2FinanceNetwork, token tokenV1Domain.Token, to string, amount string, decimals int, tokenType string) tokenV1Domain.MintFT {
	// t.Helper()
	// out, err := c.MintToken(token.Address, to, amount, decimals, tokenType)
	// if err != nil {
	// 	t.Fatalf("MintToken: %v", err)
	// }
	// var m tokenV1Domain.Mint
	// unmarshalState(t, out.States[0].Object, &m)
	// if m.TokenAddress != token.Address {
	// 	t.Fatalf("mint token mismatch: %s != %s", m.TokenAddress, token.Address)
	// }
	// return m
	return tokenV1Domain.MintFT{}
}

func createBurnFT(t *testing.T, c client2f.Client2FinanceNetwork, token tokenV1Domain.Token, amount string, decimals int, tokenType, uuid string) tokenV1Domain.BurnFT {
	// t.Helper()
	// out, err := c.BurnToken(token.Address, amount, decimals, tokenType, uuid)
	// if err != nil {
	// 	t.Fatalf("BurnToken: %v", err)
	// }
	// var b tokenV1Domain.Burn
	// unmarshalState(t, out.States[0].Object, &b)
	// if b.TokenAddress != token.Address {
	// 	t.Fatalf("burn token mismatch: %s != %s", b.TokenAddress, token.Address)
	// }
	// return b
	return tokenV1Domain.BurnFT{}
}

func createTransferFT(t *testing.T, c client2f.Client2FinanceNetwork, token tokenV1Domain.Token, to string, amount string, decimals int, tokenType, uuid string) tokenV1Domain.TransferFT {
	// t.Helper()
	// out, err := c.TransferToken(token.Address, to, amount, decimals, tokenType, uuid)
	// if err != nil {
	// 	t.Fatalf("TransferToken: %v", err)
	// }
	// var tr tokenV1Domain.TransferFT
	// unmarshalState(t, out.States[0].Object, &tr)
	// if tr.ToAddress != to {
	// 	t.Fatalf("transfer to mismatch: %s != %s", tr.ToAddress, to)
	// }
	// return tr
	return tokenV1Domain.TransferFT{}
}
