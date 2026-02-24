package e2e_test

import (
	"log"
	"testing"
	"time"

	"gitlab.com/2finance/2finance-network/blockchain/contract/contractV1/models"
	faucetV1Models "gitlab.com/2finance/2finance-network/blockchain/contract/faucetV1/models"
	"gitlab.com/2finance/2finance-network/blockchain/contract/faucetV1"
	tokenV1Domain "gitlab.com/2finance/2finance-network/blockchain/contract/tokenV1/domain"
)

func TestFaucetFlow(t *testing.T) {
	c := setupClient(t)

	owner, ownerPriv := createWallet(t, c)
	c.SetPrivateKey(ownerPriv)

	dec := 5
	stablecoin := false
	tok := createBasicToken(t, c, owner.PublicKey, dec, true, tokenV1Domain.FUNGIBLE, stablecoin)
	_ = createMint(t, c, tok, owner.PublicKey, "10000", tok.Decimals, tok.TokenType)

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

	log.Printf("Token Address: %s", tok.Address)
	log.Printf("Token Symbol: %s", tok.Symbol)
	log.Printf("Token Name: %s", tok.Name)
	log.Printf("Token Decimals: %d", tok.Decimals)
	log.Printf("Token Total Supply: %s", tok.TotalSupply)
	log.Printf("Token Type: %s", tok.TokenType)
	log.Printf("Token Stablecoin: %v", tok.Stablecoin)

	merchant, merchPriv := createWallet(t, c)

	// Allow merchant
	c.SetPrivateKey(ownerPriv)
	allowMerchOut, err := c.AllowUsers(tok.Address, map[string]bool{
		merchant.PublicKey: true,
	})
	if err != nil {
		t.Fatalf("AllowUsers(merchant): %v", err)
	}
	if len(allowMerchOut.States) == 0 {
		t.Fatalf("AllowUsers(merchant) returned empty States")
	}
	if allowMerchOut.States[0].Object == nil {
		t.Fatalf("AllowUsers(merchant) returned nil state object")
	}

	var apMerch tokenV1Domain.AccessPolicy
	unmarshalState(t, allowMerchOut.States[0].Object, &apMerch)
	if apMerch.Users == nil || !apMerch.Users[merchant.PublicKey] {
		t.Fatalf("AllowUsers(merchant) missing merchant in users")
	}

	log.Printf("AllowUsers(merchant) Output States: %+v", allowMerchOut.States)
	log.Printf("AllowUsers(merchant) Output Logs: %+v", allowMerchOut.Logs)
	log.Printf("AllowUsers(merchant) Output Delegated Call: %+v", allowMerchOut.DelegatedCall)

	log.Printf("AllowUsers Mode: %s", apMerch.Mode)
	log.Printf("AllowUsers Users: %+v", apMerch.Users)

	// Transfer owner -> merchant (helper)
	c.SetPrivateKey(ownerPriv)
	_ = createTransfer(t, c, tok, merchant.PublicKey, "50", tok.Decimals, tok.TokenType, "")

	start := time.Now().Add(2 * time.Second)
	exp := time.Now().Add(20 * time.Minute)

	amount := "4"

	// --------------------------------------------------------------------
	// Deploy Faucet contract (envelope + unmarshal + validate + log)
	// --------------------------------------------------------------------
	contractState := models.ContractStateModel{}
	deployedContract, err := c.DeployContract1(faucetV1.FAUCET_CONTRACT_V1)
	if err != nil {
		t.Fatalf("DeployContract(Faucet): %v", err)
	}
	if len(deployedContract.States) == 0 {
		t.Fatalf("DeployContract(Faucet) returned empty States")
	}
	if deployedContract.States[0].Object == nil {
		t.Fatalf("DeployContract(Faucet) returned nil state object")
	}

	unmarshalState(t, deployedContract.States[0].Object, &contractState)

	if contractState.Address == "" {
		t.Fatalf("DeployContract(Faucet) returned empty contract address")
	}
	address := contractState.Address

	log.Printf("DeployContract(Faucet) Output States: %+v", deployedContract.States)
	log.Printf("DeployContract(Faucet) Output Logs: %+v", deployedContract.Logs)
	log.Printf("DeployContract(Faucet) Output Delegated Call: %+v", deployedContract.DelegatedCall)

	log.Printf("Faucet Contract Address: %s", address)

	// --------------------------------------------------------------------
	// AddFaucet (envelope + unmarshal + validate + log)
	// --------------------------------------------------------------------
	addOut, err := c.AddFaucet(
		address,
		merchant.PublicKey,
		tok.Address,
		start,
		exp,
		false, // paused
		3,     // request limit
		amount,
		2*time.Second,
	)
	if err != nil {
		t.Fatalf("AddFaucet: %v", err)
	}
	if len(addOut.States) == 0 {
		t.Fatalf("AddFaucet returned empty States")
	}
	if addOut.States[0].Object == nil {
		t.Fatalf("AddFaucet returned nil state object")
	}

	var f faucetV1Models.FaucetStateModel
	unmarshalState(t, addOut.States[0].Object, &f)

	// Field validation (state model)
	if f.Address == "" {
		t.Fatalf("AddFaucet Address empty")
	}
	if f.Owner != merchant.PublicKey {
		t.Fatalf("AddFaucet Owner mismatch: got %q want %q", f.Owner, merchant.PublicKey)
	}
	if f.TokenAddress != tok.Address {
		t.Fatalf("AddFaucet TokenAddress mismatch: got %q want %q", f.TokenAddress, tok.Address)
	}
	if f.StartTime == nil {
		t.Fatalf("AddFaucet StartTime nil")
	}
	if f.ExpireTime == nil {
		t.Fatalf("AddFaucet ExpireTime nil")
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
	if f.CreatedAt.IsZero() {
		t.Fatalf("AddFaucet CreatedAt is zero")
	}
	if f.UpdatedAt.IsZero() {
		t.Fatalf("AddFaucet UpdatedAt is zero")
	}

	// campos de faucet específicos
	if f.ClaimAmount != amount {
		t.Fatalf("AddFaucet ClaimAmount mismatch: got %q want %q", f.ClaimAmount, amount)
	}
	if f.ClaimIntervalDuration == nil {
		t.Fatalf("AddFaucet ClaimIntervalDuration nil")
	}
	if *f.ClaimIntervalDuration != 2*time.Second {
		t.Fatalf("AddFaucet ClaimIntervalDuration mismatch: got %v want %v", *f.ClaimIntervalDuration, 2*time.Second)
	}

	log.Printf("AddFaucet Output States: %+v", addOut.States)
	log.Printf("AddFaucet Output Logs: %+v", addOut.Logs)
	log.Printf("AddFaucet Output Delegated Call: %+v", addOut.DelegatedCall)

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
	log.Printf("AddFaucet CreatedAt: %s", f.CreatedAt.String())
	log.Printf("AddFaucet UpdatedAt: %s", f.UpdatedAt.String())

	// --------------------------------------------------------------------
	// Allow faucet address (token) (envelope + validate + log)
	// --------------------------------------------------------------------
	allowFaucetOut, err := c.AllowUsers(tok.Address, map[string]bool{f.Address: true})
	if err != nil {
		t.Fatalf("AllowUsers(faucet): %v", err)
	}
	if len(allowFaucetOut.States) == 0 {
		t.Fatalf("AllowUsers(faucet) returned empty States")
	}
	if allowFaucetOut.States[0].Object == nil {
		t.Fatalf("AllowUsers(faucet) returned nil state object")
	}

	var apFaucet tokenV1Domain.AccessPolicy
	unmarshalState(t, allowFaucetOut.States[0].Object, &apFaucet)
	if apFaucet.Users == nil || !apFaucet.Users[f.Address] {
		t.Fatalf("AllowUsers(faucet) missing faucet in users")
	}

	log.Printf("AllowUsers(faucet) Output States: %+v", allowFaucetOut.States)
	log.Printf("AllowUsers(faucet) Output Logs: %+v", allowFaucetOut.Logs)
	log.Printf("AllowUsers(faucet) Output Delegated Call: %+v", allowFaucetOut.DelegatedCall)

	log.Printf("AllowUsers Mode: %s", apFaucet.Mode)
	log.Printf("AllowUsers Users: %+v", apFaucet.Users)

	// --------------------------------------------------------------------
	// DepositFunds (merchant signs) (envelope + unmarshal + validate + log)
	// --------------------------------------------------------------------
	c.SetPrivateKey(merchPriv)

	depositAmount := "569"
	depositOut, err := c.DepositFunds(f.Address, tok.Address, depositAmount, tokenV1Domain.FUNGIBLE, "")
	if err != nil {
		t.Fatalf("DepositFunds: %v", err)
	}
	if len(depositOut.States) == 0 {
		t.Fatalf("DepositFunds returned empty States")
	}
	if depositOut.States[0].Object == nil {
		t.Fatalf("DepositFunds returned nil state object")
	}

	var fAfterDeposit faucetV1Models.FaucetStateModel
	unmarshalState(t, depositOut.States[0].Object, &fAfterDeposit)

	if fAfterDeposit.Address != f.Address {
		t.Fatalf("DepositFunds Faucet address mismatch: got %q want %q", fAfterDeposit.Address, f.Address)
	}
	if fAfterDeposit.TokenAddress != tok.Address {
		t.Fatalf("DepositFunds TokenAddress mismatch: got %q want %q", fAfterDeposit.TokenAddress, tok.Address)
	}
	if fAfterDeposit.Owner != merchant.PublicKey {
		t.Fatalf("DepositFunds Owner mismatch: got %q want %q", fAfterDeposit.Owner, merchant.PublicKey)
	}

	log.Printf("DepositFunds Output States: %+v", depositOut.States)
	log.Printf("DepositFunds Output Logs: %+v", depositOut.Logs)
	log.Printf("DepositFunds Output Delegated Call: %+v", depositOut.DelegatedCall)

	log.Printf("DepositFunds FaucetAddress: %s", fAfterDeposit.Address)

	// --------------------------------------------------------------------
	// Wait start and claim as user (envelope + validate + log)
	// --------------------------------------------------------------------
	user, userPriv := createWallet(t, c)

	time.Sleep(5 * time.Second)

	// allow user (owner signs)
	c.SetPrivateKey(ownerPriv)
	allowUserOut, err := c.AllowUsers(tok.Address, map[string]bool{user.PublicKey: true})
	if err != nil {
		t.Fatalf("AllowUsers(user): %v", err)
	}
	if len(allowUserOut.States) == 0 || allowUserOut.States[0].Object == nil {
		t.Fatalf("AllowUsers(user) returned empty/nil state")
	}

	// claim (user signs)
	c.SetPrivateKey(userPriv)

	claimOut, err := c.ClaimFunds(f.Address, tokenV1Domain.FUNGIBLE, "")
	if err != nil {
		t.Fatalf("ClaimFunds: %v", err)
	}
	if len(claimOut.States) == 0 {
		t.Fatalf("ClaimFunds returned empty States")
	}
	if claimOut.States[0].Object == nil {
		t.Fatalf("ClaimFunds returned nil state object")
	}

	var fAfterClaim faucetV1Models.FaucetStateModel
	unmarshalState(t, claimOut.States[0].Object, &fAfterClaim)

	if fAfterClaim.Address != f.Address {
		t.Fatalf("ClaimFunds Faucet address mismatch: got %q want %q", fAfterClaim.Address, f.Address)
	}

	log.Printf("ClaimFunds Output States: %+v", claimOut.States)
	log.Printf("ClaimFunds Output Logs: %+v", claimOut.Logs)
	log.Printf("ClaimFunds Output Delegated Call: %+v", claimOut.DelegatedCall)

	log.Printf("ClaimFunds FaucetAddress: %s", fAfterClaim.Address)
	log.Printf("ClaimFunds RequestsByUserJSONB: %+v", fAfterClaim.RequestsByUserJSONB)
	log.Printf("ClaimFunds LastClaimByUserJSONB: %+v", fAfterClaim.LastClaimByUserJSONB)

	// --------------------------------------------------------------------
	// Pause / Unpause (merchant) (envelope + unmarshal + validate + log)
	// --------------------------------------------------------------------
	c.SetPrivateKey(merchPriv)

	pauseOut, err := c.PauseFaucet(f.Address, true)
	if err != nil {
		t.Fatalf("PauseFaucet: %v", err)
	}
	if len(pauseOut.States) == 0 || pauseOut.States[0].Object == nil {
		t.Fatalf("PauseFaucet returned empty/nil state")
	}

	var fPaused faucetV1Models.FaucetStateModel
	unmarshalState(t, pauseOut.States[0].Object, &fPaused)

	if fPaused.Address != f.Address {
		t.Fatalf("PauseFaucet address mismatch: got %q want %q", fPaused.Address, f.Address)
	}
	if !fPaused.Paused {
		t.Fatalf("PauseFaucet expected Paused=true")
	}

	log.Printf("PauseFaucet Output States: %+v", pauseOut.States)
	log.Printf("PauseFaucet Output Logs: %+v", pauseOut.Logs)
	log.Printf("PauseFaucet Output Delegated Call: %+v", pauseOut.DelegatedCall)

	log.Printf("PauseFaucet Address: %s", fPaused.Address)
	log.Printf("PauseFaucet Paused: %v", fPaused.Paused)

	unpauseOut, err := c.UnpauseFaucet(f.Address, false)
	if err != nil {
		t.Fatalf("UnpauseFaucet: %v", err)
	}
	if len(unpauseOut.States) == 0 || unpauseOut.States[0].Object == nil {
		t.Fatalf("UnpauseFaucet returned empty/nil state")
	}

	var fUnpaused faucetV1Models.FaucetStateModel
	unmarshalState(t, unpauseOut.States[0].Object, &fUnpaused)

	if fUnpaused.Address != f.Address {
		t.Fatalf("UnpauseFaucet address mismatch: got %q want %q", fUnpaused.Address, f.Address)
	}
	if fUnpaused.Paused {
		t.Fatalf("UnpauseFaucet expected Paused=false")
	}

	log.Printf("UnpauseFaucet Output States: %+v", unpauseOut.States)
	log.Printf("UnpauseFaucet Output Logs: %+v", unpauseOut.Logs)
	log.Printf("UnpauseFaucet Output Delegated Call: %+v", unpauseOut.DelegatedCall)

	log.Printf("UnpauseFaucet Address: %s", fUnpaused.Address)
	log.Printf("UnpauseFaucet Paused: %v", fUnpaused.Paused)

	// --------------------------------------------------------------------
	// Getters (envelope + unmarshal + validate + log)
	// --------------------------------------------------------------------
	getOut, err := c.GetFaucet(f.Address)
	if err != nil {
		t.Fatalf("GetFaucet: %v", err)
	}
	if len(getOut.States) == 0 || getOut.States[0].Object == nil {
		t.Fatalf("GetFaucet returned empty/nil state")
	}

	var fGet faucetV1Models.FaucetStateModel
	unmarshalState(t, getOut.States[0].Object, &fGet)

	if fGet.Address != f.Address {
		t.Fatalf("GetFaucet address mismatch: got %q want %q", fGet.Address, f.Address)
	}

	log.Printf("GetFaucet Output States: %+v", getOut.States)
	log.Printf("GetFaucet Output Logs: %+v", getOut.Logs)
	log.Printf("GetFaucet Output Delegated Call: %+v", getOut.DelegatedCall)

	log.Printf("GetFaucet Address: %s", fGet.Address)
	log.Printf("GetFaucet Owner: %s", fGet.Owner)
	log.Printf("GetFaucet TokenAddress: %s", fGet.TokenAddress)
	log.Printf("GetFaucet StartTime: %s", fGet.StartTime.String())
	log.Printf("GetFaucet ExpireTime: %s", fGet.ExpireTime.String())
	log.Printf("GetFaucet Paused: %v", fGet.Paused)
	log.Printf("GetFaucet RequestLimit: %d", fGet.RequestLimit)
	log.Printf("GetFaucet ClaimAmount: %s", fGet.ClaimAmount)
	if fGet.ClaimIntervalDuration != nil {
		log.Printf("GetFaucet ClaimIntervalDuration: %v", *fGet.ClaimIntervalDuration)
	}
	log.Printf("GetFaucet RequestsByUserJSONB: %+v", fGet.RequestsByUserJSONB)
	log.Printf("GetFaucet LastClaimByUserJSONB: %+v", fGet.LastClaimByUserJSONB)
	log.Printf("GetFaucet Hash: %s", fGet.Hash)
	log.Printf("GetFaucet CreatedAt: %s", fGet.CreatedAt.String())
	log.Printf("GetFaucet UpdatedAt: %s", fGet.UpdatedAt.String())

	listOut, err := c.ListFaucets(merchant.PublicKey, 1, 10, true)
	if err != nil {
		t.Fatalf("ListFaucets: %v", err)
	}
	if len(listOut.States) == 0 || listOut.States[0].Object == nil {
		t.Fatalf("ListFaucets returned empty/nil state")
	}

	var list []faucetV1Models.FaucetStateModel
	unmarshalState(t, listOut.States[0].Object, &list)

	found := false
	for _, it := range list {
		if it.Address == f.Address {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("ListFaucets: faucet %s not found", f.Address)
	}

	log.Printf("ListFaucets Output States: %+v", listOut.States)
	log.Printf("ListFaucets Output Logs: %+v", listOut.Logs)
	log.Printf("ListFaucets Output Delegated Call: %+v", listOut.DelegatedCall)

	log.Printf("ListFaucets Count: %d", len(list))
}

func TestFaucetFlow_NonFungible(t *testing.T) {
	c := setupClient(t)

	owner, ownerPriv := createWallet(t, c)
	c.SetPrivateKey(ownerPriv)

	dec := 0
	tokenType := tokenV1Domain.NON_FUNGIBLE
	stablecoin := false

	tok := createBasicToken(t, c, owner.PublicKey, dec, true, tokenType, stablecoin)

	// Token (mínimo) validate + log
	if tok.Address == "" {
		t.Fatalf("Token address empty")
	}
	if tok.TokenType != tokenV1Domain.NON_FUNGIBLE {
		t.Fatalf("Token type mismatch: got %s want %s", tok.TokenType, tokenV1Domain.NON_FUNGIBLE)
	}

	log.Printf("Token Address: %s", tok.Address)
	log.Printf("Token Type: %s", tok.TokenType)

	// --------------------------------------------------------------------
	// Mint NFT (envelope + validate + log)
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
		t.Fatalf("Mint NFT TokenAddress mismatch: got %q want %q", mint.TokenAddress, tok.Address)
	}
	if mint.MintTo != owner.PublicKey {
		t.Fatalf("Mint NFT MintTo mismatch: got %q want %q", mint.MintTo, owner.PublicKey)
	}
	if mint.TokenType != tok.TokenType {
		t.Fatalf("Mint NFT TokenType mismatch: got %q want %q", mint.TokenType, tok.TokenType)
	}
	if len(mint.TokenUUIDList) != 1 {
		t.Fatalf("expected 1 NFT uuid, got %d", len(mint.TokenUUIDList))
	}
	nftUUID := mint.TokenUUIDList[0]
	if nftUUID == "" {
		t.Fatalf("minted nft uuid empty")
	}

	log.Printf("Mint Output States: %+v", mintOut.States)
	log.Printf("Mint Output Logs: %+v", mintOut.Logs)
	log.Printf("Mint Output Delegated Call: %+v", mintOut.DelegatedCall)

	log.Printf("Mint TokenAddress: %s", mint.TokenAddress)
	log.Printf("Mint ToAddress: %s", mint.MintTo)
	log.Printf("Mint Amount: %s", mint.Amount)
	log.Printf("Mint TokenType: %s", mint.TokenType)
	log.Printf("Mint TokenUUIDList: %+v", mint.TokenUUIDList)

	merchant, merchPriv := createWallet(t, c)

	// --------------------------------------------------------------------
	// Allow merchant + Transfer NFT owner -> merchant (envelope + validate + log)
	// --------------------------------------------------------------------
	c.SetPrivateKey(ownerPriv)

	allowMerchOut, err := c.AllowUsers(tok.Address, map[string]bool{merchant.PublicKey: true})
	if err != nil {
		t.Fatalf("AllowUsers(merchant): %v", err)
	}
	if len(allowMerchOut.States) == 0 || allowMerchOut.States[0].Object == nil {
		t.Fatalf("AllowUsers(merchant) returned empty/nil state")
	}

	var apMerch tokenV1Domain.AccessPolicy
	unmarshalState(t, allowMerchOut.States[0].Object, &apMerch)
	if apMerch.Users == nil || !apMerch.Users[merchant.PublicKey] {
		t.Fatalf("AllowUsers(merchant) missing merchant in users")
	}

	log.Printf("AllowUsers(merchant) Output States: %+v", allowMerchOut.States)
	log.Printf("AllowUsers(merchant) Output Logs: %+v", allowMerchOut.Logs)
	log.Printf("AllowUsers(merchant) Output Delegated Call: %+v", allowMerchOut.DelegatedCall)

	trOut, err := c.TransferToken(tok.Address, merchant.PublicKey, "1", dec, tokenType, nftUUID)
	if err != nil {
		t.Fatalf("Transfer NFT to merchant: %v", err)
	}
	if len(trOut.States) == 0 || trOut.States[0].Object == nil {
		t.Fatalf("Transfer NFT returned empty/nil state")
	}

	var tr tokenV1Domain.Transfer
	unmarshalState(t, trOut.States[0].Object, &tr)

	if tr.ToAddress != merchant.PublicKey {
		t.Fatalf("Transfer NFT to merchant mismatch: got %q want %q", tr.ToAddress, merchant.PublicKey)
	}
	if tr.UUID != nftUUID {
		t.Fatalf("Transfer NFT uuid mismatch: got %q want %q", tr.UUID, nftUUID)
	}

	log.Printf("Transfer NFT Output States: %+v", trOut.States)
	log.Printf("Transfer NFT Output Logs: %+v", trOut.Logs)
	log.Printf("Transfer NFT Output Delegated Call: %+v", trOut.DelegatedCall)

	log.Printf("Transfer TokenAddress: %s", tr.TokenAddress)
	log.Printf("Transfer FromAddress: %s", tr.FromAddress)
	log.Printf("Transfer ToAddress: %s", tr.ToAddress)
	log.Printf("Transfer Amount: %s", tr.Amount)
	log.Printf("Transfer TokenType: %s", tr.TokenType)
	log.Printf("Transfer UUID: %s", tr.UUID)

	// --------------------------------------------------------------------
	// Deploy Faucet Contract (envelope + unmarshal + validate + log)
	// --------------------------------------------------------------------
	start := time.Now().Add(2 * time.Second)
	exp := time.Now().Add(20 * time.Minute)

	contractState := models.ContractStateModel{}
	deployedContract, err := c.DeployContract1(faucetV1.FAUCET_CONTRACT_V1)
	if err != nil {
		t.Fatalf("DeployContract(Faucet): %v", err)
	}
	if len(deployedContract.States) == 0 || deployedContract.States[0].Object == nil {
		t.Fatalf("DeployContract(Faucet) returned empty/nil state")
	}
	unmarshalState(t, deployedContract.States[0].Object, &contractState)
	if contractState.Address == "" {
		t.Fatalf("DeployContract(Faucet) returned empty contract address")
	}

	log.Printf("DeployContract(Faucet) Output States: %+v", deployedContract.States)
	log.Printf("DeployContract(Faucet) Output Logs: %+v", deployedContract.Logs)
	log.Printf("DeployContract(Faucet) Output Delegated Call: %+v", deployedContract.DelegatedCall)

	log.Printf("Faucet Contract Address: %s", contractState.Address)

	// --------------------------------------------------------------------
	// Add Faucet (merchant signs) (envelope + unmarshal + validate + log)
	// --------------------------------------------------------------------
	c.SetPrivateKey(merchPriv)

	addOut, err := c.AddFaucet(
		contractState.Address,
		merchant.PublicKey,
		tok.Address,
		start,
		exp,
		false,
		1, // request limit
		"1",
		2*time.Second,
	)
	if err != nil {
		t.Fatalf("AddFaucet NFT: %v", err)
	}
	if len(addOut.States) == 0 || addOut.States[0].Object == nil {
		t.Fatalf("AddFaucet NFT returned empty/nil state")
	}

	var f faucetV1Models.FaucetStateModel
	unmarshalState(t, addOut.States[0].Object, &f)

	if f.Address == "" {
		t.Fatalf("AddFaucet NFT Address empty")
	}
	if f.Owner != merchant.PublicKey {
		t.Fatalf("AddFaucet NFT Owner mismatch: got %q want %q", f.Owner, merchant.PublicKey)
	}
	if f.TokenAddress != tok.Address {
		t.Fatalf("AddFaucet NFT TokenAddress mismatch: got %q want %q", f.TokenAddress, tok.Address)
	}
	if f.RequestLimit != 1 {
		t.Fatalf("AddFaucet NFT RequestLimit mismatch: got %d want %d", f.RequestLimit, 1)
	}
	if f.ClaimAmount != "1" {
		t.Fatalf("AddFaucet NFT ClaimAmount mismatch: got %q want %q", f.ClaimAmount, "1")
	}
	if f.ClaimIntervalDuration == nil {
		t.Fatalf("AddFaucet NFT ClaimIntervalDuration nil")
	}
	if *f.ClaimIntervalDuration != 2*time.Second {
		t.Fatalf("AddFaucet NFT ClaimIntervalDuration mismatch: got %v want %v", *f.ClaimIntervalDuration, 2*time.Second)
	}
	if f.Hash == "" {
		t.Fatalf("AddFaucet NFT Hash empty")
	}

	log.Printf("AddFaucet NFT Output States: %+v", addOut.States)
	log.Printf("AddFaucet NFT Output Logs: %+v", addOut.Logs)
	log.Printf("AddFaucet NFT Output Delegated Call: %+v", addOut.DelegatedCall)

	log.Printf("AddFaucet Address: %s", f.Address)
	log.Printf("AddFaucet Owner: %s", f.Owner)
	log.Printf("AddFaucet TokenAddress: %s", f.TokenAddress)
	log.Printf("AddFaucet StartTime: %s", f.StartTime.String())
	log.Printf("AddFaucet ExpireTime: %s", f.ExpireTime.String())
	log.Printf("AddFaucet Paused: %v", f.Paused)
	log.Printf("AddFaucet RequestLimit: %d", f.RequestLimit)
	log.Printf("AddFaucet ClaimAmount: %s", f.ClaimAmount)
	if f.ClaimIntervalDuration != nil {
		log.Printf("AddFaucet ClaimIntervalDuration: %v", *f.ClaimIntervalDuration)
	}
	log.Printf("AddFaucet RequestsByUserJSONB: %+v", f.RequestsByUserJSONB)
	log.Printf("AddFaucet LastClaimByUserJSONB: %+v", f.LastClaimByUserJSONB)
	log.Printf("AddFaucet Hash: %s", f.Hash)

	// --------------------------------------------------------------------
	// Allow faucet to interact with token (owner signs) (envelope + validate + log)
	// --------------------------------------------------------------------
	c.SetPrivateKey(ownerPriv)
	allowFaucetOut, err := c.AllowUsers(tok.Address, map[string]bool{f.Address: true})
	if err != nil {
		t.Fatalf("AllowUsers(faucet): %v", err)
	}
	if len(allowFaucetOut.States) == 0 || allowFaucetOut.States[0].Object == nil {
		t.Fatalf("AllowUsers(faucet) returned empty/nil state")
	}

	var apFaucet tokenV1Domain.AccessPolicy
	unmarshalState(t, allowFaucetOut.States[0].Object, &apFaucet)
	if apFaucet.Users == nil || !apFaucet.Users[f.Address] {
		t.Fatalf("AllowUsers(faucet) missing faucet in users")
	}

	log.Printf("AllowUsers(faucet) Output States: %+v", allowFaucetOut.States)
	log.Printf("AllowUsers(faucet) Output Logs: %+v", allowFaucetOut.Logs)
	log.Printf("AllowUsers(faucet) Output Delegated Call: %+v", allowFaucetOut.DelegatedCall)

	// --------------------------------------------------------------------
	// Deposit NFT into Faucet (merchant signs) (envelope + validate + log)
	// --------------------------------------------------------------------
	c.SetPrivateKey(merchPriv)

	depositOut, err := c.DepositFunds(f.Address, tok.Address, "1", tokenType, nftUUID)
	if err != nil {
		t.Fatalf("DepositFunds NFT: %v", err)
	}
	if len(depositOut.States) == 0 || depositOut.States[0].Object == nil {
		t.Fatalf("DepositFunds NFT returned empty/nil state")
	}

	var fAfterDeposit faucetV1Models.FaucetStateModel
	unmarshalState(t, depositOut.States[0].Object, &fAfterDeposit)

	if fAfterDeposit.Address != f.Address {
		t.Fatalf("DepositFunds NFT faucet address mismatch: got %q want %q", fAfterDeposit.Address, f.Address)
	}

	log.Printf("DepositFunds NFT Output States: %+v", depositOut.States)
	log.Printf("DepositFunds NFT Output Logs: %+v", depositOut.Logs)
	log.Printf("DepositFunds NFT Output Delegated Call: %+v", depositOut.DelegatedCall)

	log.Printf("DepositFunds FaucetAddress: %s", fAfterDeposit.Address)

	// --------------------------------------------------------------------
	// Wait start + create user + allow user + claim (envelope + validate + log)
	// --------------------------------------------------------------------
	time.Sleep(5 * time.Second)

	user, userPriv := createWallet(t, c)

	c.SetPrivateKey(ownerPriv)
	if _, err := c.AllowUsers(tok.Address, map[string]bool{user.PublicKey: true}); err != nil {
		t.Fatalf("AllowUsers(user): %v", err)
	}

	c.SetPrivateKey(userPriv)
	claimOut, err := c.ClaimFunds(f.Address, tokenType, nftUUID)
	if err != nil {
		t.Fatalf("ClaimFunds NFT: %v", err)
	}
	if len(claimOut.States) == 0 || claimOut.States[0].Object == nil {
		t.Fatalf("ClaimFunds NFT returned empty/nil state")
	}

	var fAfterClaim faucetV1Models.FaucetStateModel
	unmarshalState(t, claimOut.States[0].Object, &fAfterClaim)

	if fAfterClaim.Address != f.Address {
		t.Fatalf("ClaimFunds NFT faucet address mismatch: got %q want %q", fAfterClaim.Address, f.Address)
	}

	log.Printf("ClaimFunds NFT Output States: %+v", claimOut.States)
	log.Printf("ClaimFunds NFT Output Logs: %+v", claimOut.Logs)
	log.Printf("ClaimFunds NFT Output Delegated Call: %+v", claimOut.DelegatedCall)

	log.Printf("ClaimFunds FaucetAddress: %s", fAfterClaim.Address)
	log.Printf("ClaimFunds RequestsByUserJSONB: %+v", fAfterClaim.RequestsByUserJSONB)
	log.Printf("ClaimFunds LastClaimByUserJSONB: %+v", fAfterClaim.LastClaimByUserJSONB)

	// --------------------------------------------------------------------
	// Pause / Unpause (merchant) (envelope + unmarshal + validate + log)
	// --------------------------------------------------------------------
	c.SetPrivateKey(merchPriv)

	pauseOut, err := c.PauseFaucet(f.Address, true)
	if err != nil {
		t.Fatalf("PauseFaucet: %v", err)
	}
	if len(pauseOut.States) == 0 || pauseOut.States[0].Object == nil {
		t.Fatalf("PauseFaucet returned empty/nil state")
	}

	var fPaused faucetV1Models.FaucetStateModel
	unmarshalState(t, pauseOut.States[0].Object, &fPaused)

	if fPaused.Address != f.Address {
		t.Fatalf("PauseFaucet address mismatch: got %q want %q", fPaused.Address, f.Address)
	}
	if !fPaused.Paused {
		t.Fatalf("PauseFaucet expected Paused=true")
	}

	log.Printf("PauseFaucet Output States: %+v", pauseOut.States)
	log.Printf("PauseFaucet Output Logs: %+v", pauseOut.Logs)
	log.Printf("PauseFaucet Output Delegated Call: %+v", pauseOut.DelegatedCall)

	unpauseOut, err := c.UnpauseFaucet(f.Address, false)
	if err != nil {
		t.Fatalf("UnpauseFaucet: %v", err)
	}
	if len(unpauseOut.States) == 0 || unpauseOut.States[0].Object == nil {
		t.Fatalf("UnpauseFaucet returned empty/nil state")
	}

	var fUnpaused faucetV1Models.FaucetStateModel
	unmarshalState(t, unpauseOut.States[0].Object, &fUnpaused)

	if fUnpaused.Address != f.Address {
		t.Fatalf("UnpauseFaucet address mismatch: got %q want %q", fUnpaused.Address, f.Address)
	}
	if fUnpaused.Paused {
		t.Fatalf("UnpauseFaucet expected Paused=false")
	}

	log.Printf("UnpauseFaucet Output States: %+v", unpauseOut.States)
	log.Printf("UnpauseFaucet Output Logs: %+v", unpauseOut.Logs)
	log.Printf("UnpauseFaucet Output Delegated Call: %+v", unpauseOut.DelegatedCall)

	// --------------------------------------------------------------------
	// Getters (envelope + unmarshal + validate + log)
	// --------------------------------------------------------------------
	getOut, err := c.GetFaucet(f.Address)
	if err != nil {
		t.Fatalf("GetFaucet: %v", err)
	}
	if len(getOut.States) == 0 || getOut.States[0].Object == nil {
		t.Fatalf("GetFaucet returned empty/nil state")
	}

	var fGet faucetV1Models.FaucetStateModel
	unmarshalState(t, getOut.States[0].Object, &fGet)

	if fGet.Address != f.Address {
		t.Fatalf("GetFaucet address mismatch: got %q want %q", fGet.Address, f.Address)
	}

	log.Printf("GetFaucet Output States: %+v", getOut.States)
	log.Printf("GetFaucet Output Logs: %+v", getOut.Logs)
	log.Printf("GetFaucet Output Delegated Call: %+v", getOut.DelegatedCall)

	listOut, err := c.ListFaucets(merchant.PublicKey, 1, 10, true)
	if err != nil {
		t.Fatalf("ListFaucets: %v", err)
	}
	if len(listOut.States) == 0 || listOut.States[0].Object == nil {
		t.Fatalf("ListFaucets returned empty/nil state")
	}

	var list []faucetV1Models.FaucetStateModel
	unmarshalState(t, listOut.States[0].Object, &list)

	found := false
	for _, it := range list {
		if it.Address == f.Address {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("ListFaucets: faucet %s not found", f.Address)
	}

	log.Printf("ListFaucets Output States: %+v", listOut.States)
	log.Printf("ListFaucets Output Logs: %+v", listOut.Logs)
	log.Printf("ListFaucets Output Delegated Call: %+v", listOut.DelegatedCall)
}
