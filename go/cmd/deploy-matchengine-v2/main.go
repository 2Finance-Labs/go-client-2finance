package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	client2f "github.com/2Finance-Labs/2finance-sdk-client/client_2finance"
	"github.com/2Finance-Labs/2finance-sdk-client/wallet_manager"
	"gitlab.com/2finance/2finance-network/blockchain/contract/contractV2/domain"
	"gitlab.com/2finance/2finance-network/blockchain/contract/matchEngineV2"
	"gitlab.com/2finance/2finance-network/blockchain/utils"
)

func main() {
	privateKey := strings.TrimSpace(os.Getenv("MATCH_ENGINE_ORACLE_PRIVATE_KEY"))
	if privateKey == "" {
		fatalf("MATCH_ENGINE_ORACLE_PRIVATE_KEY is required")
	}
	networkURL := strings.TrimSpace(os.Getenv("TWO_FINANCE_NETWORK_URL"))
	if networkURL == "" {
		networkURL = "http://127.0.0.1:19295"
	}
	chainID := uint8(1)
	if value := strings.TrimSpace(os.Getenv("CHAIN_ID")); value != "" {
		parsed, err := strconv.ParseUint(value, 10, 8)
		if err != nil {
			fatalf("invalid CHAIN_ID: %v", err)
		}
		chainID = uint8(parsed)
	}
	allowSequenceGaps := false
	if value := strings.TrimSpace(os.Getenv("MATCH_ENGINE_ALLOW_SEQUENCE_GAPS")); value != "" {
		parsed, err := strconv.ParseBool(value)
		if err != nil {
			fatalf("invalid MATCH_ENGINE_ALLOW_SEQUENCE_GAPS: %v", err)
		}
		allowSequenceGaps = parsed
	}

	walletPath := filepath.Join(os.TempDir(), fmt.Sprintf("2finance-matchengine-oracle-%d.wallet", time.Now().UnixNano()))
	defer os.Remove(walletPath)
	wallet := wallet_manager.NewWalletManager(walletPath)
	const password = "2finance-local-matchengine-oracle"
	if err := wallet.ImportPrivateKey([]byte(privateKey), password); err != nil {
		fatalf("import oracle key: %v", err)
	}
	if err := wallet.UnlockWithPassword(password); err != nil {
		fatalf("unlock oracle key: %v", err)
	}

	client := client2f.NewV2(networkURL, http.DefaultClient, wallet)
	client.SetChainID(chainID)
	contractAddress := strings.TrimSpace(os.Getenv("MATCH_ENGINE_CONTRACT_ADDRESS"))
	deployHash := ""
	if contractAddress == "" {
		output, err := client.DeployContract1(matchEngineV2.MATCH_ENGINE_CONTRACT_V2)
		if err != nil {
			fatalf("deploy %s: %v", matchEngineV2.MATCH_ENGINE_CONTRACT_V2, err)
		}
		if len(output.Logs) == 0 {
			fatalf("deployment returned no logs")
		}
		var deployed domain.Contract
		if err := json.Unmarshal(output.Logs[0].Event, &deployed); err != nil {
			fatalf("decode deployment event: %v", err)
		}
		if deployed.Address == "" || deployed.ContractVersion != matchEngineV2.MATCH_ENGINE_CONTRACT_V2 {
			fatalf("invalid deployment event: address=%q version=%q", deployed.Address, deployed.ContractVersion)
		}
		contractAddress = deployed.Address
		deployHash = output.Logs[0].TransactionHash
	}

	owner := wallet.OwnerAddress()
	registerHash := sendContractTransaction(client, chainID, owner, contractAddress, matchEngineV2.METHOD_REGISTER_MATCH_ENGINE, map[string]interface{}{
		"owner":               owner,
		"engine_id":           "matchengine-local-v2",
		"oracle":              owner,
		"mqtt_client_id":      "matchengine_v2",
		"allow_sequence_gaps": allowSequenceGaps,
		"metadata": map[string]interface{}{
			"environment":           "local",
			"events_json_retention": "full",
		},
	})
	symbolHash := sendContractTransaction(client, chainID, owner, contractAddress, matchEngineV2.METHOD_UPSERT_SYMBOL, map[string]interface{}{
		"owner":          owner,
		"engine_id":      "matchengine-local-v2",
		"symbol_id":      uint64(1),
		"market":         "BTC/USDT",
		"base_asset":     "BTC",
		"quote_asset":    "USDT",
		"base_asset_id":  uint64(2),
		"quote_asset_id": uint64(1),
		"status":         "active",
	})

	result := map[string]string{
		"address":             contractAddress,
		"contract_version":    matchEngineV2.MATCH_ENGINE_CONTRACT_V2,
		"deploy_tx_hash":      deployHash,
		"register_tx_hash":    registerHash,
		"symbol_tx_hash":      symbolHash,
		"engine_id":           "matchengine-local-v2",
		"oracle":              owner,
		"allow_sequence_gaps": strconv.FormatBool(allowSequenceGaps),
	}
	encoded, err := json.Marshal(result)
	if err != nil {
		fatalf("encode result: %v", err)
	}
	fmt.Println(string(encoded))
}

func sendContractTransaction(client client2f.Client2FinanceNetwork, chainID uint8, from, to, method string, data map[string]interface{}) string {
	uuid7, err := utils.NewUUID7()
	if err != nil {
		fatalf("create UUID for %s: %v", method, err)
	}
	output, err := client.SignAndSendTransaction(chainID, from, to, method, data, 1, uuid7)
	if err != nil {
		fatalf("%s: %v", method, err)
	}
	if len(output.Logs) == 0 {
		fatalf("%s returned no logs", method)
	}
	return output.Logs[0].TransactionHash
}

func fatalf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}
