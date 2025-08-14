package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/2Finance-Labs/go-client-2finance/client_2finance"
	faucetV1Domain "gitlab.com/2finance/2finance-network/blockchain/contract/faucetV1/domain"
	tokenV1Domain "gitlab.com/2finance/2finance-network/blockchain/contract/tokenV1/domain"
	"gitlab.com/2finance/2finance-network/blockchain/contract/walletV1/domain"
	"gitlab.com/2finance/2finance-network/config"
)

func generateRandomSuffix(length int) (string, error) {
	bytes := make([]byte, length)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes)[:length], nil
}

func execute(client client_2finance.Client2FinanceNetwork) {
	pubKeyRecentRegistered, privateKeyRecentRegistered, err := client.GenerateKeyEd25519()
	if err != nil {
		log.Fatalf("Error generating keys: %v", err)
	}

	// Set the private key for signing transactions
	client.SetPrivateKey(privateKeyRecentRegistered)

	log.Printf("Public Key: %s\n", pubKeyRecentRegistered)
	//log.Printf("Private Key: %s\n", privKey)

	walletContract, err := client.AddWallet(pubKeyRecentRegistered)
	if err != nil {
		log.Fatalf("Error registering wallet: %v", err)
	}

	// Step 1: Extract the Object map
	rawWallet := walletContract.States[0].Object

	// Step 2: Marshal it back into JSON
	walletBytes, err := json.Marshal(rawWallet)
	if err != nil {
		log.Fatalf("Error marshaling wallet object: %v", err)
	}

	// Step 3: Unmarshal into your domain.Wallet struct
	var wallet domain.Wallet
	err = json.Unmarshal(walletBytes, &wallet)
	if err != nil {
		log.Fatalf("Error unmarshalling into domain.Wallet: %v", err)
	}
	log.Printf("Wallet PublicKey: %s\n", wallet.PublicKey)
	log.Printf("Wallet Amount: %s\n", wallet.Amount)
	log.Printf("Wallet CreatedAt: %s\n", wallet.CreatedAt)
	log.Printf("Wallet UpdatedAt: %s\n", wallet.UpdatedAt)

	walletContract, err = client.GetWallet(wallet.PublicKey)
	if err != nil {
		log.Fatalf("Error getting wallet: %v", err)
	}

	// Step 1: Extract the Object map
	rawWallet = walletContract.States[0].Object

	// Step 2: Marshal it back into JSON
	walletBytes, err = json.Marshal(rawWallet)
	if err != nil {
		log.Fatalf("Error marshaling wallet object: %v", err)
	}

	// Step 3: Unmarshal into your domain.Wallet struct
	err = json.Unmarshal(walletBytes, &wallet)
	if err != nil {
		log.Fatalf("Error unmarshalling into domain.Wallet: %v", err)
	}
	log.Printf("Wallet PublicKey: %s\n", wallet.PublicKey)
	log.Printf("Wallet Amount: %s\n", wallet.Amount)
	log.Printf("Wallet CreatedAt: %s\n", wallet.CreatedAt)
	log.Printf("Wallet UpdatedAt: %s\n", wallet.UpdatedAt)

	getNonce, err := client.GetNonce(wallet.PublicKey)
	if err != nil {
		log.Fatalf("Error getting nonce: %v", err)
	}
	log.Printf("Nonce: %d\n", getNonce)

	listTransactions, err := client.ListTransactions(wallet.PublicKey, "", "", nil, 0, 1, 10, true)
	if err != nil {
		log.Fatalf("Error listing transactions: %v", err)
	}
	log.Printf("List Transactions: %+v\n", listTransactions)

	transaction := listTransactions[0]
	listLogs, err := client.ListLogs("wallet_created", 1, transaction.Hash, nil, "", 1, 10, true)
	if err != nil {
		log.Fatalf("Error listing logs: %v", err)
	}
	log.Printf("List Logs: %+v\n", listLogs)

	// blocks, err := client.ListBlocks(0, time.Time{}, "", "", "", 1, 10, true)
	// if err != nil {
	// 	log.Fatalf("Error listing blocks: %v", err)
	// }
	// log.Printf("List Blocks: %+v\n", blocks)

	// //TODOOOOOOOOOOOOOOOOOOOOOO
	// //client.SetPrivateKey(pvtKeyDefault)
	// // transfer, err := client.TransferWallet(pubKeyDefault, "1", 18)
	// // if err != nil {
	// // 	log.Fatalf("Error transferring wallet: %v", err)
	// // }
	// // log.Printf("Transfer Response: %+v\n", transfer)
	// // log.Printf("Transfer EVENT: %+v\n", transfer.Event)
	// // log.Printf("Transfer: %+v\n", transfer.Wallet)

	symbol := "2F"
	suffix, err := generateRandomSuffix(4)
	if err != nil {
		panic(err)
	}
	symbol = symbol + suffix
	name := "2Finance"
	decimals := 3
	totalSupply := "10"
	description := "2Finance is a decentralized finance platform that offers a range of financial services, including lending, borrowing, and trading."
	owner := wallet.PublicKey
	image := "https://example.com/image.png"
	website := "https://example.com"
	tagsSocialMedia := map[string]string{
		"twitter": "https://twitter.com/2finance",
	}

	tagsCategory := map[string]string{
		"category": "DeFi",
	}
	tags := map[string]string{
		"tag1": "DeFi",
		"tag2": "Blockchain",
	}
	creator := "2Finance Creator"
	creatorWebsite := "https://creator.com"
	allowUsers := map[string]bool{
		"43b23ffdd134ff73eda6cad0a5bd0d97877dd63ab8ba21ffe49d80fe51fd5dec": true,
	}
	blockUsers := map[string]bool{
		"e8ef1e9a97c08ce9ba388b5df7f43964ce19317c3a77338d39d80898cbe22914": true,
	}
	feeTiersList := []map[string]interface{}{
		{
			"min_amount": "0",
			"max_amount": "1000000000000000000",
			"min_volume": "0",
			"max_volume": "10000000000000000000",
			"fee_bps":    50,
		},
		{
			"min_amount": "1000000000000000001",
			"max_amount": "10000000000000000000",
			"min_volume": "10000000000000000001",
			"max_volume": "50000000000000000000",
			"fee_bps":    25,
		},
		{
			"min_amount": "10000000000000000001",
			"max_amount": "100000000000000000000",
			"min_volume": "50000000000000000001",
			"max_volume": "5000000000000000000100",
			"fee_bps":    10,
		},
	}

	feeAddress := "fe1b01a9861bb265b141c00517d7697c8a0d8286492a14d776ca33ffdded43c1"

	freezeAuthorityRevoked := false
	mintAuthorityRevoked := false
	updateAuthorityRevoked := false
	paused := false
	expired_at := time.Time{} // 30 days from now
	tokenContract, err := client.AddToken(symbol, name, decimals, totalSupply, description, owner, image, website, tagsSocialMedia, tagsCategory, tags, creator, creatorWebsite, allowUsers, blockUsers, feeTiersList, feeAddress, freezeAuthorityRevoked, mintAuthorityRevoked, updateAuthorityRevoked, paused, expired_at)
	if err != nil {
		log.Fatalf("Error adding token: %v", err)
	}

	fmt.Printf("Token Contract: %+v\n", tokenContract)

	rawToken := tokenContract.States[0].Object
	tokenBytes, err := json.Marshal(rawToken)
	if err != nil {
		log.Fatalf("Error marshaling token object: %v", err)
	}
	var token tokenV1Domain.Token
	err = json.Unmarshal(tokenBytes, &token)
	if err != nil {
		log.Fatalf("Error unmarshalling into domain.Token: %v", err)
	}
	log.Printf("Token Symbol: %s\n", token.Symbol)
	log.Printf("Token Name: %s\n", token.Name)
	log.Printf("Token Decimals: %d\n", token.Decimals)
	log.Printf("Token TotalSupply: %s\n", token.TotalSupply)
	log.Printf("Token Description: %s\n", token.Description)
	log.Printf("Token Address: %s\n", token.Address)
	log.Printf("Token Owner: %s\n", token.Owner)
	log.Printf("Token Image: %s\n", token.Image)
	log.Printf("Token Website: %s\n", token.Website)
	log.Printf("Token TagsSocialMedia: %+v\n", token.TagsSocialMedia)
	log.Printf("Token TagsCategory: %+v\n", token.TagsCategory)
	log.Printf("Token Tags: %+v\n", token.Tags)
	log.Printf("Token Creator: %s\n", token.Creator)
	log.Printf("Token CreatorWebsite: %s\n", token.CreatorWebsite)
	log.Printf("Token AllowUsers: %+v\n", token.AllowUsersMap)
	log.Printf("Token BlockUsers: %+v\n", token.BlockUsersMap)
	log.Printf("Token FeeTiersList: %+v\n", token.FeeTiersList)
	log.Printf("Token FeeAddress: %s\n", token.FeeAddress)
	log.Printf("Token FreezeAuthorityRevoked: %t\n", token.FreezeAuthorityRevoked)
	log.Printf("Token MintAuthorityRevoked: %t\n", token.MintAuthorityRevoked)
	log.Printf("Token UpdateAuthorityRevoked: %t\n", token.UpdateAuthorityRevoked)
	log.Printf("Token Paused: %t\n", token.Paused)

	rawMint := tokenContract.States[1].Object
	mintBytes, err := json.Marshal(rawMint)
	if err != nil {
		log.Fatalf("Error marshaling mint object: %v", err)
	}

	var mint tokenV1Domain.Mint
	err = json.Unmarshal(mintBytes, &mint)
	if err != nil {
		log.Fatalf("Error unmarshalling into domain.Mint: %v", err)
	}
	log.Printf("Mint TokenAddress: %s\n", mint.TokenAddress)
	log.Printf("Mint MintTo: %s\n", mint.MintTo)
	log.Printf("Mint Amount: %s\n", mint.Amount)

	rawBalance := tokenContract.States[2].Object
	balanceBytes, err := json.Marshal(rawBalance)
	if err != nil {
		log.Fatalf("Error marshaling balance object: %v", err)
	}
	var balance tokenV1Domain.Balance
	err = json.Unmarshal(balanceBytes, &balance)
	if err != nil {
		log.Fatalf("Error unmarshalling into domain.Balance: %v", err)
	}
	log.Printf("Balance TokenAddress: %s\n", balance.TokenAddress)
	log.Printf("Balance OwnerAddress: %s\n", balance.OwnerAddress)
	log.Printf("Balance Amount: %s\n", balance.Amount)

	mintContract, err := client.MintToken(token.Address, wallet.PublicKey, "35", decimals)
	if err != nil {
		log.Fatalf("Error minting token: %v", err)
	}

	log.Printf("Mint Contract: %+v\n", mintContract)
	rawMint = mintContract.States[0].Object
	mintBytes, err = json.Marshal(rawMint)
	if err != nil {
		log.Fatalf("Error marshaling mint object: %v", err)
	}
	err = json.Unmarshal(mintBytes, &mint)
	if err != nil {
		log.Fatalf("Error unmarshalling into domain.Mint: %v", err)
	}
	log.Printf("Mint TokenAddress: %s\n", mint.TokenAddress)
	log.Printf("Mint MintTo: %s\n", mint.MintTo)
	log.Printf("Mint Amount: %s\n", mint.Amount)

	burnContract, err := client.BurnToken(token.Address, "12", decimals)
	if err != nil {
		log.Fatalf("Error burning token: %v", err)
	}
	log.Printf("Burn Contract: %+v\n", burnContract)
	rawBurn := burnContract.States[0].Object
	burnBytes, err := json.Marshal(rawBurn)
	if err != nil {
		log.Fatalf("Error marshaling burn object: %v", err)
	}
	var burn tokenV1Domain.Burn
	err = json.Unmarshal(burnBytes, &burn)
	if err != nil {
		log.Fatalf("Error unmarshalling into domain.Burn: %v", err)
	}
	log.Printf("Burn TokenAddress: %s\n", burn.TokenAddress)
	log.Printf("Burn BurnFrom: %s\n", burn.BurnFrom)
	log.Printf("Burn Amount: %s\n", burn.Amount)

	feeTiersList = []map[string]interface{}{
		{
			"min_amount": "99",
			"max_amount": "1000000000000000000",
			"min_volume": "0",
			"max_volume": "10000000000000000000",
			"fee_bps":    50,
		},
		{
			"min_amount": "1000000000000000001",
			"max_amount": "10000000000000000000",
			"min_volume": "10000000000000000001",
			"max_volume": "50000000000000000000",
			"fee_bps":    25,
		},
	}

	feeTiersListContract, err := client.UpdateFeeTiers(token.Address, feeTiersList)
	if err != nil {
		log.Fatalf("Error updating fee tiers list: %v", err)
	}
	rawFeeTiersList := feeTiersListContract.States[0].Object
	feeTiersListBytes, err := json.Marshal(rawFeeTiersList)
	if err != nil {
		log.Fatalf("Error marshaling fee tiers list object: %v", err)
	}
	var tokenFeeTiersList []tokenV1Domain.FeeTier
	err = json.Unmarshal(feeTiersListBytes, &tokenFeeTiersList)
	if err != nil {
		log.Fatalf("Error unmarshalling into domain.FeeTiersList: %v", err)
	}
	//log.Printf("Update Fee Tiers List Token Address: %s\n", tokenFeeTiersList.Address)
	log.Printf("Update Fee Tiers List Token FeeTiersList: %+v\n", tokenFeeTiersList)

	pubKeyNew, _, err := client.GenerateKeyEd25519()
	if err != nil {
		log.Fatalf("Error generating keys: %v", err)
	}

	allowUsers = map[string]bool{
		pubKeyNew: true,
	}

	//COMMENTED OUT TO MAKE THE FLOW WORK, IS WORKING
	// blockUsers = map[string]bool{
	// 	pubKeyNew: true,
	// }
	// blockUsersContract, err := client.BlockUsers(token.Address, blockUsers)
	// if err != nil {
	// 	log.Fatalf("Error blocking users: %v", err)
	// }
	// log.Printf("Block Users Contract: %+v\n", blockUsersContract)
	// rawBlockUsers := blockUsersContract.States[0].Object
	// blockUsersBytes, err := json.Marshal(rawBlockUsers)
	// if err != nil {
	// 	log.Fatalf("Error marshaling block users object: %v", err)
	// }
	// var tokenBlockUsers tokenV1Domain.Token
	// err = json.Unmarshal(blockUsersBytes, &tokenBlockUsers)
	// if err != nil {
	// 	log.Fatalf("Error unmarshalling into domain.BlockUsers: %v", err)
	// }
	// log.Printf("Block Users Token Address: %s\n", tokenBlockUsers.Address)
	// log.Printf("Block Users Token BlockUsers: %s\n", tokenBlockUsers.BlockUsersMap)

	allowUsersContract, err := client.AllowUsers(token.Address, allowUsers)
	if err != nil {
		log.Fatalf("Error allowing users: %v", err)
	}
	log.Printf("Allow Users Contract: %+v\n", allowUsersContract)
	rawAllowUsers := allowUsersContract.States[0].Object
	allowUsersBytes, err := json.Marshal(rawAllowUsers)
	if err != nil {
		log.Fatalf("Error marshaling allow users object: %v", err)
	}
	var tokenAllowUsers tokenV1Domain.Token
	err = json.Unmarshal(allowUsersBytes, &tokenAllowUsers)
	if err != nil {
		log.Fatalf("Error unmarshalling into domain.AllowUsers: %v", err)
	}
	log.Printf("Allow Users Token Address: %s\n", tokenAllowUsers.Address)
	log.Printf("Allow Users Token AllowUsers: %+v\n", tokenAllowUsers.AllowUsersMap)

	transferContract, err := client.TransferToken(token.Address, pubKeyNew, "1", decimals)
	if err != nil {
		log.Fatalf("Error transferring token: %v", err)
	}
	log.Printf("Transfer Contract: %+v\n", transferContract)
	rawTransfer := transferContract.States[0].Object
	transferBytes, err := json.Marshal(rawTransfer)
	if err != nil {
		log.Fatalf("Error marshaling transfer object: %v", err)
	}
	var transfer tokenV1Domain.Transfer
	err = json.Unmarshal(transferBytes, &transfer)
	if err != nil {
		log.Fatalf("Error unmarshalling into domain.Transfer: %v", err)
	}
	log.Printf("Transfer TokenAddress: %s\n", transfer.TokenAddress)
	log.Printf("Transfer ToAddress: %s\n", transfer.ToAddress)
	log.Printf("Transfer Amount: %s\n", transfer.Amount)

	allowUsers = map[string]bool{
		"43b23ffdd134ff73eda6cad0a5bd0d97877dd63ab8ba21ffe49d80fe51fd5dec": true,
		"e8ef1e9a97c08ce9ba388b5df7f43964ce19317c3a77338d39d80898cbe22914": true,
		"fe1b01a9861bb265b141c00517d7697c8a0d8286492a14d776ca33ffdded43c1": true,
	}

	addAllowUsersContract, err := client.AllowUsers(token.Address, allowUsers)
	if err != nil {
		log.Fatalf("Error adding allow users: %v", err)
	}
	log.Printf("Add Allow Users Contract: %+v\n", addAllowUsersContract)
	rawAddAllowUsers := addAllowUsersContract.States[0].Object
	addAllowUsersBytes, err := json.Marshal(rawAddAllowUsers)
	if err != nil {
		log.Fatalf("Error marshaling add allow users object: %v", err)
	}
	var tokenAllowUsers2 tokenV1Domain.Token
	err = json.Unmarshal(addAllowUsersBytes, &tokenAllowUsers2)
	if err != nil {
		log.Fatalf("Error unmarshalling into domain.AddAllowUsers: %v", err)
	}
	log.Printf("Add Allow Users Token Address: %+v\n", tokenAllowUsers2)
	log.Printf("Add Allow Users Token Owner: %+v\n", tokenAllowUsers2.AllowUsersMap)

	DeleteAllowUsersContract, err := client.DisallowUsers(token.Address, allowUsers)
	if err != nil {
		log.Fatalf("Error deleting allow users: %v", err)
	}
	log.Printf("Delete Allow Users Contract: %+v\n", DeleteAllowUsersContract)
	rawDeleteAllowUsers := DeleteAllowUsersContract.States[0].Object
	deleteAllowUsersBytes, err := json.Marshal(rawDeleteAllowUsers)
	if err != nil {
		log.Fatalf("Error marshaling delete allow users object: %v", err)
	}
	var tokenDeleteAllowUsers tokenV1Domain.Token
	err = json.Unmarshal(deleteAllowUsersBytes, &tokenDeleteAllowUsers)
	if err != nil {
		log.Fatalf("Error unmarshalling into domain.DeleteAllowUsers: %v", err)
	}
	log.Printf("Delete Allow Users Token Address: %s\n", tokenDeleteAllowUsers.Address)
	log.Printf("Delete Allow Users Token AllowUsers: %+v\n", tokenDeleteAllowUsers.AllowUsersMap)

	blockUsers = map[string]bool{
		"43b23ffdd134ff73eda6cad0a5bd0d97877dd63ab8ba21ffe49d80fe51fd5dec": true,
		"e8ef1e9a97c08ce9ba388b5df7f43964ce19317c3a77338d39d80898cbe22914": true,
		"fe1b01a9861bb265b141c00517d7697c8a0d8286492a14d776ca33ffdded43c1": true,
	}
	addBlockUsersContract, err := client.BlockUsers(token.Address, blockUsers)
	if err != nil {
		log.Fatalf("Error adding block users: %v", err)
	}
	log.Printf("Add Block Users Contract: %+v\n", addBlockUsersContract)
	rawAddBlockUsers := addBlockUsersContract.States[0].Object
	addBlockUsersBytes, err := json.Marshal(rawAddBlockUsers)
	if err != nil {
		log.Fatalf("Error marshaling add block users object: %v", err)
	}
	var tokenAddBlockUsers tokenV1Domain.Token
	err = json.Unmarshal(addBlockUsersBytes, &tokenAddBlockUsers)
	if err != nil {
		log.Fatalf("Error unmarshalling into domain.AddBlockUsers: %v", err)
	}
	log.Printf("Add Block Users Token Address: %s\n", tokenAddBlockUsers.Address)
	log.Printf("Add Block Users Token BlockUsers: %+v\n", tokenAddBlockUsers.BlockUsersMap)

	DeleteBlockUsersContract, err := client.UnblockUsers(token.Address, blockUsers)
	if err != nil {
		log.Fatalf("Error deleting block users: %v", err)
	}
	log.Printf("Delete Block Users Contract: %+v\n", DeleteBlockUsersContract)
	rawDeleteBlockUsers := DeleteBlockUsersContract.States[0].Object
	deleteBlockUsersBytes, err := json.Marshal(rawDeleteBlockUsers)
	if err != nil {
		log.Fatalf("Error marshaling delete block users object: %v", err)
	}
	var tokenDeleteBlockUsers tokenV1Domain.Token
	err = json.Unmarshal(deleteBlockUsersBytes, &tokenDeleteBlockUsers)
	if err != nil {
		log.Fatalf("Error unmarshalling into domain.DeleteBlockUsers: %v", err)
	}
	log.Printf("Delete Block Users Token Address: %s\n", tokenDeleteBlockUsers.Address)
	log.Printf("Delete Block Users Token BlockUsers: %+v\n", tokenDeleteBlockUsers.BlockUsersMap)

	//COMMENTED OUT TO MAKE THE FLOW WORK, IS WORKING
	// freezeAuthorityRevoked = true
	// FreezeAuthorityRevokedContract, err := client.RevokeFreezeAuthority(token.Address, freezeAuthorityRevoked)
	// if err != nil {
	// 	log.Fatalf("Error revoking freeze authority: %v", err)
	// }
	// log.Printf("Freeze Authority Revoked Contract: %+v\n", FreezeAuthorityRevokedContract)
	// rawFreezeAuthorityRevoked := FreezeAuthorityRevokedContract.States[0].Object
	// freezeAuthorityRevokedBytes, err := json.Marshal(rawFreezeAuthorityRevoked)
	// if err != nil {
	// 	log.Fatalf("Error marshaling freeze authority revoked object: %v", err)
	// }
	// var tokenFreezeAuthorityRevoked tokenV1Domain.Token
	// err = json.Unmarshal(freezeAuthorityRevokedBytes, &tokenFreezeAuthorityRevoked)
	// if err != nil {
	// 	log.Fatalf("Error unmarshalling into domain.FreezeAuthorityRevoked: %v", err)
	// }
	// log.Printf("Freeze Authority Revoked Token Address: %s\n", tokenFreezeAuthorityRevoked.Address)
	// log.Printf("Freeze Authority Revoked Token FreezeAuthorityRevoked: %t\n", tokenFreezeAuthorityRevoked.FreezeAuthorityRevoked)

	mintAuthorityRevoked = true
	MintAuthorityRevokedContract, err := client.RevokeMintAuthority(token.Address, mintAuthorityRevoked)
	if err != nil {
		log.Fatalf("Error revoking mint authority: %v", err)
	}
	log.Printf("Mint Authority Revoked Contract: %+v\n", MintAuthorityRevokedContract)
	rawMintAuthorityRevoked := MintAuthorityRevokedContract.States[0].Object
	mintAuthorityRevokedBytes, err := json.Marshal(rawMintAuthorityRevoked)
	if err != nil {
		log.Fatalf("Error marshaling mint authority revoked object: %v", err)
	}
	var tokenMintAuthorityRevoked tokenV1Domain.Token
	err = json.Unmarshal(mintAuthorityRevokedBytes, &tokenMintAuthorityRevoked)
	if err != nil {
		log.Fatalf("Error unmarshalling into domain.MintAuthorityRevoked: %v", err)
	}
	log.Printf("Mint Authority Revoked Token Address: %s\n", tokenMintAuthorityRevoked.Address)
	log.Printf("Mint Authority Revoked Token MintAuthorityRevoked: %t\n", tokenMintAuthorityRevoked.MintAuthorityRevoked)

	//UPDATE TOKEN METADATA
	symbolNew := "2F-NEW"
	suffixNew, err := generateRandomSuffix(4)
	if err != nil {
		log.Fatalf("Error generating random suffix: %v", err)
	}
	symbolNew = symbolNew + suffixNew
	nameNew := "2Finance New"
	decimalsNew := 4
	descriptionNew := "2Finance New is an upgraded version of the 2Finance platform, offering enhanced features and services."
	imageNew := "https://example.com/image-new.png"
	websiteNew := "https://example-new.com"
	tagsSocialMediaNew := map[string]string{
		"twitter": "https://twitter.com/2finance-new",
	}
	tagsCategoryNew := map[string]string{
		"category": "DeFi New",
	}
	tagsNew := map[string]string{
		"tag1": "DeFi New",
		"tag2": "Blockchain New",
	}
	creatorNew := "2Finance Creator New"
	creatorWebsiteNew := "https://creator-new.com"
	expired_atNew := time.Now().AddDate(0, 0, 30) // 30 days from now
	updateMetadataContract, err := client.UpdateMetadata(token.Address, symbolNew, nameNew, decimalsNew, descriptionNew, imageNew, websiteNew, tagsSocialMediaNew, tagsCategoryNew, tagsNew, creatorNew, creatorWebsiteNew, expired_atNew)
	if err != nil {
		log.Fatalf("Error updating token metadata: %v", err)
	}
	log.Printf("Update Metadata Contract: %+v\n", updateMetadataContract)
	rawUpdateMetadata := updateMetadataContract.States[0].Object
	updateMetadataBytes, err := json.Marshal(rawUpdateMetadata)
	if err != nil {
		log.Fatalf("Error marshaling update metadata object: %v", err)
	}
	var tokenUpdateMetadata tokenV1Domain.Token
	err = json.Unmarshal(updateMetadataBytes, &tokenUpdateMetadata)
	if err != nil {
		log.Fatalf("Error unmarshalling into domain.UpdateMetadata: %v", err)
	}
	log.Printf("Update Metadata Token Symbol: %s\n", tokenUpdateMetadata.Symbol)
	log.Printf("Update Metadata Token Name: %s\n", tokenUpdateMetadata.Name)
	log.Printf("Update Metadata Token Decimals: %d\n", tokenUpdateMetadata.Decimals)
	log.Printf("Update Metadata Token TotalSupply: %s\n", tokenUpdateMetadata.TotalSupply)
	log.Printf("Update Metadata Token Description: %s\n", tokenUpdateMetadata.Description)
	log.Printf("Update Metadata Token Address: %s\n", tokenUpdateMetadata.Address)
	log.Printf("Update Metadata Token Image: %s\n", tokenUpdateMetadata.Image)
	log.Printf("Update Metadata Token Website: %s\n", tokenUpdateMetadata.Website)
	log.Printf("Update Metadata Token TagsSocialMedia: %+v\n", tokenUpdateMetadata.TagsSocialMedia)
	log.Printf("Update Metadata Token TagsCategory: %+v\n", tokenUpdateMetadata.TagsCategory)
	log.Printf("Update Metadata Token Tags: %+v\n", tokenUpdateMetadata.Tags)
	log.Printf("Update Metadata Token Creator: %s\n", tokenUpdateMetadata.Creator)
	log.Printf("Update Metadata Token CreatorWebsite: %s\n", tokenUpdateMetadata.CreatorWebsite)

	updateAuthorityRevoked = true
	UpdateAuthorityRevokedContract, err := client.RevokeUpdateAuthority(token.Address, updateAuthorityRevoked)
	if err != nil {
		log.Fatalf("Error revoking update authority: %v", err)
	}
	log.Printf("Update Authority Revoked Contract: %+v\n", UpdateAuthorityRevokedContract)
	rawUpdateAuthorityRevoked := UpdateAuthorityRevokedContract.States[0].Object
	updateAuthorityRevokedBytes, err := json.Marshal(rawUpdateAuthorityRevoked)
	if err != nil {
		log.Fatalf("Error marshaling update authority revoked object: %v", err)
	}
	var tokenUpdateAuthorityRevoked tokenV1Domain.Token
	err = json.Unmarshal(updateAuthorityRevokedBytes, &tokenUpdateAuthorityRevoked)
	if err != nil {
		log.Fatalf("Error unmarshalling into domain.UpdateAuthorityRevoked: %v", err)
	}
	log.Printf("Update Authority Revoked Token Address: %s\n", tokenUpdateAuthorityRevoked.Address)
	log.Printf("Update Authority Revoked Token UpdateAuthorityRevoked: %t\n", tokenUpdateAuthorityRevoked.UpdateAuthorityRevoked)

	pause := true
	pauseContract, err := client.PauseToken(token.Address, pause)
	if err != nil {
		log.Fatalf("Error pausing token: %v", err)
	}
	log.Printf("Pause Contract: %+v\n", pauseContract)
	rawPause := pauseContract.States[0].Object
	pauseBytes, err := json.Marshal(rawPause)
	if err != nil {
		log.Fatalf("Error marshaling pause object: %v", err)
	}
	var tokenPause tokenV1Domain.Token
	err = json.Unmarshal(pauseBytes, &tokenPause)
	if err != nil {
		log.Fatalf("Error unmarshalling into domain.Pause: %v", err)
	}
	log.Printf("Pause Token Address: %s\n", tokenPause.Address)
	log.Printf("Pause Token Paused: %t\n", tokenPause.Paused)

	unpause := false
	unpauseContract, err := client.UnpauseToken(token.Address, unpause)
	if err != nil {
		log.Fatalf("Error unpausing token: %v", err)
	}
	log.Printf("Unpause Contract: %+v\n", unpauseContract)
	rawUnpause := unpauseContract.States[0].Object
	unpauseBytes, err := json.Marshal(rawUnpause)
	if err != nil {
		log.Fatalf("Error marshaling unpause object: %v", err)
	}
	var tokenUnpause tokenV1Domain.Token
	err = json.Unmarshal(unpauseBytes, &tokenUnpause)
	if err != nil {
		log.Fatalf("Error unmarshalling into domain.Unpause: %v", err)
	}
	log.Printf("Unpause Token Address: %s\n", tokenUnpause.Address)
	log.Printf("Unpause Token Paused: %t\n", tokenUnpause.Paused)

	symbol = "2F-E"
	suffix, err = generateRandomSuffix(4)
	symbol = symbol + suffix
	expired_at = time.Now().AddDate(0, 0, 30) // 30 days from now
	mintAuthorityRevoked = false
	tokenContract, err = client.AddToken(symbol, name, decimals, totalSupply, description, owner, image, website, tagsSocialMedia, tagsCategory, tags, creator, creatorWebsite, allowUsers, blockUsers, feeTiersList, feeAddress, freezeAuthorityRevoked, mintAuthorityRevoked, updateAuthorityRevoked, paused, expired_at)
	if err != nil {
		log.Fatalf("Error adding token with expiration: %v", err)
	}
	log.Printf("Token Contract with Expiration: %+v\n", tokenContract)
	rawToken = tokenContract.States[0].Object
	tokenBytes, err = json.Marshal(rawToken)
	if err != nil {
		log.Fatalf("Error marshaling token object with expiration: %v", err)
	}
	err = json.Unmarshal(tokenBytes, &token)
	if err != nil {
		log.Fatalf("Error unmarshalling into domain.Token with expiration: %v", err)
	}
	log.Printf("Token Symbol with Expiration: %s\n", token.Symbol)
	log.Printf("Token Name with Expiration: %s\n", token.Name)
	log.Printf("Token Decimals with Expiration: %d\n", token.Decimals)
	log.Printf("Token TotalSupply with Expiration: %s\n", token.TotalSupply)
	log.Printf("Token Description with Expiration: %s\n", token.Description)
	log.Printf("Token Address with Expiration: %s\n", token.Address)
	log.Printf("Token Owner with Expiration: %s\n", token.Owner)
	log.Printf("Token Image with Expiration: %s\n", token.Image)
	log.Printf("Token Website with Expiration: %s\n", token.Website)
	log.Printf("Token TagsSocialMedia with Expiration: %+v\n", token.TagsSocialMedia)
	log.Printf("Token TagsCategory with Expiration: %+v\n", token.TagsCategory)
	log.Printf("Token Tags with Expiration: %+v\n", token.Tags)
	log.Printf("Token Creator with Expiration: %s\n", token.Creator)
	log.Printf("Token CreatorWebsite with Expiration: %s\n", token.CreatorWebsite)
	log.Printf("Token AllowUsers with Expiration: %+v\n", token.AllowUsersMap)
	log.Printf("Token BlockUsers with Expiration: %+v\n", token.BlockUsersMap)
	log.Printf("Token FeeTiersList with Expiration: %+v\n", token.FeeTiersList)
	log.Printf("Token FreezeAuthorityRevoked with Expiration: %t\n", token.FreezeAuthorityRevoked)
	log.Printf("Token MintAuthorityRevoked with Expiration: %t\n", token.MintAuthorityRevoked)
	log.Printf("Token UpdateAuthorityRevoked with Expiration: %t\n", token.UpdateAuthorityRevoked)
	log.Printf("Token Paused with Expiration: %t\n", token.Paused)
	log.Printf("Token ExpiredAt with Expiration: %s\n", token.ExpiredAt)
	//time.Sleep(5 * time.Second) // Wait for the transaction to be processed
	// Example of adding mint with expiration
	mintContract, err = client.MintToken(token.Address, wallet.PublicKey, "35", decimals)
	if err != nil {
		log.Fatalf("Error minting token with expiration: %v", err)
	}
	log.Printf("Mint Contract with Expiration: %+v\n", mintContract)
	rawMint = mintContract.States[0].Object
	mintBytes, err = json.Marshal(rawMint)
	if err != nil {
		log.Fatalf("Error marshaling mint object with expiration: %v", err)
	}
	err = json.Unmarshal(mintBytes, &mint)
	if err != nil {
		log.Fatalf("Error unmarshalling into domain.Mint with expiration: %v", err)
	}
	log.Printf("Mint TokenAddress with Expiration: %s\n", mint.TokenAddress)
	log.Printf("Mint MintTo with Expiration: %s\n", mint.MintTo)
	log.Printf("Mint Amount with Expiration: %s\n", mint.Amount)

	log.Printf("All operations completed successfully.")

	tokenContract, err = client.GetToken(token.Address, "", "")
	if err != nil {
		log.Fatalf("Error getting token: %v", err)
	}
	log.Printf("Token Contract: %+v\n", tokenContract)
	rawToken = tokenContract.States[0].Object
	tokenBytes, err = json.Marshal(rawToken)
	if err != nil {
		log.Fatalf("Error marshaling token object: %v", err)
	}

	listTokens, err := client.ListTokens("", "", "", 1, 10, true)
	if err != nil {
		log.Fatalf("Error listing tokens: %v", err)
	}
	log.Printf("List Tokens: %+v\n", listTokens)

	getTokenBalances, err := client.GetTokenBalance(token.Address, wallet.PublicKey)
	if err != nil {
		log.Fatalf("Error getting token balance: %v", err)
	}
	log.Printf("Get Token Balance: %+v\n", getTokenBalances)

	listTokenBalances, err := client.ListTokenBalances(token.Address, "", 1, 10, true)
	if err != nil {
		log.Fatalf("Error listing token balances: %v", err)
	}
	log.Printf("List Token Balances: %+v\n", listTokenBalances)

	listTokenBalances2, err := client.ListTokenBalances("", "", 1, 10, true)
	if err != nil {
		log.Fatalf("Error listing token transactions: %v", err)
	}
	log.Printf("List Token Balances: %+v\n", listTokenBalances2)

	// GERAÇÃO DE INFORMAÇÕES PARA A FAUCET

	pubKey, privKey, err := client.GenerateKeyEd25519()
	if err != nil {
		log.Fatalf("Erro ao gerar chave: %v", err)
	}

	pubKeyFaucetClaimer, _, err := client.GenerateKeyEd25519()
	if err != nil {
		log.Fatalf("Erro ao gerar chave: %v", err)
	}

	fmt.Println("Public Key:", pubKey)
	fmt.Println("Public Key Faucet Claimer:", pubKeyFaucetClaimer)

	client.SetPrivateKey(privKey)

	log.Printf("Public Key: %s\n", pubKey)
	//log.Printf("Private Key: %s\n", privKey)

	walletContract2, err := client.AddWallet(pubKey)
	if err != nil {
		log.Fatalf("Error registering wallet: %v", err)
	}

	// Step 1: Extract the Object map
	rawWallet2 := walletContract2.States[0].Object

	// Step 2: Marshal it back into JSON
	walletBytes2, err := json.Marshal(rawWallet2)
	if err != nil {
		log.Fatalf("Error marshaling wallet object: %v", err)
	}

	// Step 3: Unmarshal into your domain.Wallet struct
	var wallet2 domain.Wallet
	err = json.Unmarshal(walletBytes2, &wallet2)
	if err != nil {
		log.Fatalf("Error unmarshalling into domain.Wallet: %v", err)
	}
	log.Printf("Wallet PublicKey: %s\n", wallet2.PublicKey)
	log.Printf("Wallet Amount: %s\n", wallet2.Amount)
	log.Printf("Wallet CreatedAt: %s\n", wallet2.CreatedAt)
	log.Printf("Wallet UpdatedAt: %s\n", wallet2.UpdatedAt)

	symbol = "2F"
	suffix, err = generateRandomSuffix(4)
	if err != nil {
		panic(err)
	}
	symbol = symbol + suffix
	name = "2Finance"
	decimals = 3
	totalSupply = "10000"
	description = "2Finance is a decentralized finance platform that offers a range of financial services, including lending, borrowing, and trading."
	owner = wallet2.PublicKey
	image = "https://example.com/image.png"
	website = "https://example.com"
	tagsSocialMedia = map[string]string{
		"twitter": "https://twitter.com/2finance",
	}

	tagsCategory = map[string]string{
		"category": "DeFi",
	}
	tags = map[string]string{
		"tag1": "DeFi",
		"tag2": "Blockchain",
	}
	creator = "2Finance Creator"
	creatorWebsite = "https://creator.com"
	allowUsers = map[string]bool{
		"43b23ffdd134ff73eda6cad0a5bd0d97877dd63ab8ba21ffe49d80fe51fd5dec": true,
		pubKey: true,
	}
	blockUsers = map[string]bool{
		"e8ef1e9a97c08ce9ba388b5df7f43964ce19317c3a77338d39d80898cbe22914": true,
	}
	feeTiersList = []map[string]interface{}{
		{
			"min_amount": "0",
			"max_amount": "1000000000000000000",
			"min_volume": "0",
			"max_volume": "10000000000000000000",
			"fee_bps":    50,
		},
		{
			"min_amount": "1000000000000000001",
			"max_amount": "10000000000000000000",
			"min_volume": "10000000000000000001",
			"max_volume": "50000000000000000000",
			"fee_bps":    25,
		},
		{
			"min_amount": "10000000000000000001",
			"max_amount": "100000000000000000000",
			"min_volume": "50000000000000000001",
			"max_volume": "5000000000000000000100",
			"fee_bps":    10,
		},
	}

	feeAddress = "fe1b01a9861bb265b141c00517d7697c8a0d8286492a14d776ca33ffdded43c1"

	freezeAuthorityRevoked = false
	mintAuthorityRevoked = false
	updateAuthorityRevoked = false
	paused = false
	expired_at = time.Time{} // 30 days from now
	tokenContract, err = client.AddToken(symbol, name, decimals, totalSupply, description, owner, image, website, tagsSocialMedia, tagsCategory, tags, creator, creatorWebsite, allowUsers, blockUsers, feeTiersList, feeAddress, freezeAuthorityRevoked, mintAuthorityRevoked, updateAuthorityRevoked, paused, expired_at)
	if err != nil {
		log.Fatalf("Error adding token: %v", err)
	}

	fmt.Printf("Token Contract Wallet2: %+v\n", tokenContract)

	rawToken = tokenContract.States[0].Object
	tokenBytes, err = json.Marshal(rawToken)
	if err != nil {
		log.Fatalf("Error marshaling token object: %v", err)
	}
	var token2 tokenV1Domain.Token
	err = json.Unmarshal(tokenBytes, &token2)
	if err != nil {
		log.Fatalf("Error unmarshalling into domain.Token: %v", err)
	}

	tokenAddr := token2.Address
	startAt := time.Now().Add(5 * time.Second)
	fmt.Print("Tempo no contract:", startAt)
	expireAt := time.Now().Add(10 * time.Minute)
	requestLimit := 2
	requestsByUser := map[string]int{
		token.Owner: 1,
	}
	amountState := "10"
	claimIntervalDuration := time.Duration(1 * time.Second)

	getTokenBalances, err = client.GetTokenBalance(tokenAddr, wallet2.PublicKey)
	if err != nil {
		log.Fatalf("Error getting token balance: %v", err)
	}
	log.Printf("Get Token Balance Wallet2 Before Mint: %+v\n", getTokenBalances)

	mintContract, err = client.MintToken(tokenAddr, wallet2.PublicKey, "35", decimals)
	if err != nil {
		log.Fatalf("Error minting token with expiration: %v", err)
	}


	getTokenBalances, err = client.GetTokenBalance(tokenAddr, wallet2.PublicKey)
	if err != nil {
		log.Fatalf("Error getting token balance: %v", err)
	}
	log.Printf("Get Token Balance Wallet2 After Mint: %+v\n", getTokenBalances)

	log.Printf("Owner %s\n", owner)
	log.Printf("Token Address %s", tokenAddr)

	// ✅ ADD FAUCET
	faucetAdd, err := client.AddFaucet(
		owner,
		tokenAddr,
		startAt,
		expireAt,
		paused,
		requestLimit,
		amountState,
		claimIntervalDuration,
	)
	if err != nil {
		log.Fatalf("Error adding faucet: %v", err)
	}
	log.Printf("Faucet Added Successfully:\n%+v\n", faucetAdd)

	rawFaucet := faucetAdd.States[0].Object
	faucetBytes, err := json.Marshal(rawFaucet)
	if err != nil {
		log.Fatalf("Error marshaling faucet object: %v", err)
	}
	var faucet faucetV1Domain.Faucet
	err = json.Unmarshal(faucetBytes, &faucet)
	if err != nil {
		log.Fatalf("Error unmarshalling into domain.Faucet: %v", err)
	}

	// ✅ GET FAUCET
	getFaucet, err := client.GetFaucet(faucet.Address)
	if err != nil {
		log.Fatalf("Error geting faucet: %v", err)
	}
	log.Printf("Faucet Geted Successfully:\n%+v\n", getFaucet)

	// ✅ UPDATE FAUCET
	lastClaimByUser := map[string]time.Time{
		faucet.Address: time.Now().Add(8 * time.Second).UTC().Truncate(time.Second),
	}
	requestLimit = 10

	amountState = "7"
	faucetUpdate, err := client.UpdateFaucet(
		faucet.Address,
		startAt,
		expireAt,
		requestLimit,
		requestsByUser,
		amountState,
		claimIntervalDuration,
		lastClaimByUser,
	)
	if err != nil {
		log.Fatalf("Error updating faucet: %v", err)
	}
	log.Printf("Faucet Updated Successfully:\n%+v\n", faucetUpdate)

	log.Printf("Token BlockUsers: %+v\n", token2.BlockUsersMap)
	log.Printf("Token AllowUsers Before Added: %+v\n", token2.AllowUsersMap)
	

	// ✅ DEPOSIT FUNDS FAUCET
	allowUsers[faucet.Address] = true
	_, err = client.AllowUsers(token2.Address, allowUsers)
	if err != nil {
		log.Fatalf("Error adding allow list: %v", err)
	}
	log.Printf("Token AllowUsers After Added: %+v\n", token2.AllowUsersMap)

	// ✅ GET FAUCET
	amount := "200"
	depositFunds, err := client.DepositFunds(faucet.Address, tokenAddr, amount)
	if err != nil {
		log.Fatalf("Error depositing funds in faucet: %v", err)
	}
	log.Printf("Faucet Deposit Funds Successfully:\n%+v\n", depositFunds)

	getTokenBalances, err = client.GetTokenBalance(tokenAddr, faucet.Address)
	if err != nil {
		log.Fatalf("Error getting token balance: %v", err)
	}
	log.Printf("Get Token Balance Faucet Address After Deposit: %+v\n", getTokenBalances)

	// ✅ WITHDRAW FUNDS FAUCET
	amount = "119"
	withdrawFunds, err := client.WithdrawFunds(faucet.Address, tokenAddr, amount)
	if err != nil {
		log.Fatalf("Error withdrawing funds in faucet: %v", err)
	}
	log.Printf("Faucet Withdraw Funds Successfully:\n%+v\n", withdrawFunds)

	getTokenBalances, err = client.GetTokenBalance(tokenAddr, faucet.Address)
	if err != nil {
		log.Fatalf("Error getting token balance: %v", err)
	}
	log.Printf("Get Token Balance Faucet Address After Deposit: %+v\n", getTokenBalances)

	getTokenBalances, err = client.GetTokenBalance(tokenAddr, wallet2.PublicKey)
	if err != nil {
		log.Fatalf("Error getting token balance: %v", err)
	}
	log.Printf("Get Token Balance Wallet2 After Withdraw: %+v\n", getTokenBalances)


	// ✅ PAUSE FAUCET
	paused = true
	faucetPause, err := client.PauseFaucet(faucet.Address, paused)
	if err != nil {
		log.Fatalf("Error pausing faucet: %v", err)
	}
	log.Printf("Faucet Paused Successfully:\n%+v\n", faucetPause)

	// ✅ UNPAUSE FAUCET
	paused = false
	faucetUnpause, err := client.UnpauseFaucet(
		faucet.Address,
		paused,
	)
	if err != nil {
		log.Fatalf("Error unpausing faucet: %v", err)
	}
	log.Printf("Faucet Unpaused Successfully:\n%+v\n", faucetUnpause)

	//✅ REQUEST LIMIT FAUCETS
	updateRequestLimit, err := client.UpdateRequestLimitPerUser(faucet.Address, 2)
	if err != nil {
		log.Fatalf("Error udating request limit: %v", err)
	}
	log.Printf("Faucet Updating Request Limit Successfully:\n%+v\n", updateRequestLimit)

	// ✅ GET FAUCET
	getFaucet, err = client.GetFaucet(faucet.Address)
	if err != nil {
		log.Fatalf("Error geting faucet: %v", err)
	}
	log.Printf("Faucet Geted Successfully:\n%+v\n", getFaucet)


	claimFunds, err := client.ClaimFunds(faucet.Address)
	if err != nil {
		log.Fatalf("Error claim funds: %v", err)
	}
	log.Printf("Faucet Claim Funds Successfully:\n%+v\n", claimFunds)
	
	//Comment the line below to wait for the periodicity to take effect
	time.Sleep(2 * time.Second)

	//✅ CLAIM FUNDS FAUCETS - Periodicity
	claimFunds2, err := client.ClaimFunds(faucet.Address)
	if err != nil {
		log.Fatalf("Error claim funds with periodicity: %v", err)
	}
	log.Printf("Faucet Claim Funds Successfully with periodicity:\n%+v\n", claimFunds2)
	

	//✅ CLAIM FUNDS FAUCETS - Periodicity
	claimFunds3, err := client.ClaimFunds(faucet.Address)
	if err != nil {
		log.Fatalf("Error claim funds with periodicity: %v", err)
	}
	log.Printf("Faucet Claim Funds Successfully with periodicity:\n%+v\n", claimFunds3)


		// ✅ LIST FAUCETS
	listFaucets, err := client.ListFaucets(owner, tokenAddr, requestLimit, 1, 10, true)
	if err != nil {
		log.Fatalf("Error listing faucets: %v", err)
	}
	log.Printf("Faucet Listed Successfully:\n%+v\n", listFaucets)
}

func main() {
	arg := os.Args[1]
	config.Load_config(arg, "./.env")
	emqxHost := fmt.Sprintf("%s://%s:%s", config.EMQX_SCHEME, config.EMQX_HOST, config.EMQX_PORT)
	//pvtKeyDefault := "a5dcc28b2d34572af024273ac0ae3637f071eb36a7056d10385ec6bef5c92191fe1b01a9861bb265b141c00517d7697c8a0d8286492a14d776ca33ffdded43c1"
	//pubKeyDefault := "fe1b01a9861bb265b141c00517d7697c8a0d8286492a14d776ca33ffdded43c1"
	var wg sync.WaitGroup
	client := client_2finance.New(emqxHost, config.EMQX_CLIENT_ID, false)
	for i := 0; i < 15; i++ {
		wg.Add(1)
		go func(client client_2finance.Client2FinanceNetwork) {
			defer wg.Done()
			execute(client)
		}(client)
		wg.Wait()
	}

}
