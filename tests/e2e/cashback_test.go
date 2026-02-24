package e2e_test

import (
	"log"
	"testing"
	"time"

	"gitlab.com/2finance/2finance-network/blockchain/contract/cashbackV1"
	cashbackV1Domain "gitlab.com/2finance/2finance-network/blockchain/contract/cashbackV1/domain"
	"gitlab.com/2finance/2finance-network/blockchain/contract/contractV1/models"
	cashbackV1Models "gitlab.com/2finance/2finance-network/blockchain/contract/cashbackV1/models"
	tokenV1Domain "gitlab.com/2finance/2finance-network/blockchain/contract/tokenV1/domain"
)

func TestCashbackFlow(t *testing.T) {
	c := setupClient(t)

	owner, ownerPriv := createWallet(t, c)
	c.SetPrivateKey(ownerPriv)

	dec := 1
	stablecoin := false
	tok := createBasicToken(t, c, owner.PublicKey, dec, false, tokenV1Domain.FUNGIBLE, stablecoin)

	// Token (mínimo) validate + log (mantém seu padrão atual)
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

	// Mint (envelope) validate + log
	mintOut, err := c.MintToken(tok.Address, owner.PublicKey, "10000", dec, tok.TokenType)
	if err != nil {
		t.Fatalf("MintToken: %v", err)
	}
	if len(mintOut.States) == 0 {
		t.Fatalf("MintToken returned empty States")
	}
	if mintOut.States[0].Object == nil {
		t.Fatalf("MintToken returned nil state object")
	}

	log.Printf("Mint Output States: %+v", mintOut.States)
	log.Printf("Mint Output Logs: %+v", mintOut.Logs)
	log.Printf("Mint Output Delegated Call: %+v", mintOut.DelegatedCall)

	merchant, _ := createWallet(t, c)
	c.SetPrivateKey(ownerPriv)

	// AllowUsers (token) validate + log (no padrão mint: envelope + campos do state)
	allowMerchantOut, err := c.AllowUsers(tok.Address, map[string]bool{
		merchant.PublicKey: true,
	})
	if err != nil {
		t.Fatalf("AllowUsers(merchant): %v", err)
	}
	if len(allowMerchantOut.States) == 0 {
		t.Fatalf("AllowUsers(merchant) returned empty States")
	}
	if allowMerchantOut.States[0].Object == nil {
		t.Fatalf("AllowUsers(merchant) returned nil state object")
	}

	var ap tokenV1Domain.AccessPolicy
	unmarshalState(t, allowMerchantOut.States[0].Object, &ap)

	if ap.Mode == "" {
		t.Fatalf("AllowUsers(merchant) returned empty mode")
	}
	if ap.Users == nil || !ap.Users[merchant.PublicKey] {
		t.Fatalf("AllowUsers(merchant) missing merchant in users")
	}

	log.Printf("AllowUsers(merchant) Output States: %+v", allowMerchantOut.States)
	log.Printf("AllowUsers(merchant) Output Logs: %+v", allowMerchantOut.Logs)
	log.Printf("AllowUsers(merchant) Output Delegated Call: %+v", allowMerchantOut.DelegatedCall)

	log.Printf("AllowUsers Mode: %s", ap.Mode)
	log.Printf("AllowUsers Users: %+v", ap.Users)

	// Transfer (envelope) validate + log
	trOut, err := c.TransferToken(tok.Address, merchant.PublicKey, "50", dec, tok.TokenType, "")
	if err != nil {
		t.Fatalf("TransferToken: %v", err)
	}
	if len(trOut.States) == 0 {
		t.Fatalf("TransferToken returned empty States")
	}
	if trOut.States[0].Object == nil {
		t.Fatalf("TransferToken returned nil state object")
	}

	var tr tokenV1Domain.Transfer
	unmarshalState(t, trOut.States[0].Object, &tr)

	if tr.TokenAddress != tok.Address {
		t.Fatalf("Transfer token_address mismatch: got %q want %q", tr.TokenAddress, tok.Address)
	}
	if tr.ToAddress != merchant.PublicKey {
		t.Fatalf("Transfer to mismatch: got %q want %q", tr.ToAddress, merchant.PublicKey)
	}
	if tr.TokenType != tok.TokenType {
		t.Fatalf("Transfer token_type mismatch: got %q want %q", tr.TokenType, tok.TokenType)
	}

	log.Printf("Transfer Output States: %+v", trOut.States)
	log.Printf("Transfer Output Logs: %+v", trOut.Logs)
	log.Printf("Transfer Output Delegated Call: %+v", trOut.DelegatedCall)

	log.Printf("Transfer TokenAddress: %s", tr.TokenAddress)
	log.Printf("Transfer FromAddress: %s", tr.FromAddress)
	log.Printf("Transfer ToAddress: %s", tr.ToAddress)
	log.Printf("Transfer Amount: %s", tr.Amount)
	log.Printf("Transfer TokenType: %s", tr.TokenType)
	log.Printf("Transfer UUID: %s", tr.UUID)

	// --------------------------------------------------------------------
	// Deploy Cashback contract (envelope + unmarshal + validate + log)
	// --------------------------------------------------------------------
	start := time.Now().Add(2 * time.Second)
	exp := time.Now().Add(30 * time.Minute)

	contractState := models.ContractStateModel{}
	deployedContract, err := c.DeployContract1(cashbackV1.CASHBACK_CONTRACT_V1)
	if err != nil {
		t.Fatalf("DeployContract: %v", err)
	}
	if len(deployedContract.States) == 0 {
		t.Fatalf("DeployContract returned empty States")
	}
	if deployedContract.States[0].Object == nil {
		t.Fatalf("DeployContract returned nil state object")
	}

	unmarshalState(t, deployedContract.States[0].Object, &contractState)

	if contractState.Address == "" {
		t.Fatalf("DeployContract returned empty contract address")
	}
	address := contractState.Address

	log.Printf("DeployContract(Cashback) Output States: %+v", deployedContract.States)
	log.Printf("DeployContract(Cashback) Output Logs: %+v", deployedContract.Logs)
	log.Printf("DeployContract(Cashback) Output Delegated Call: %+v", deployedContract.DelegatedCall)

	log.Printf("Cashback Contract Address: %s", address)

	// --------------------------------------------------------------------
	// AddCashback (envelope + unmarshal + validate + log)
	// --------------------------------------------------------------------
	addOut, err := c.AddCashback(address, merchant.PublicKey, tok.Address, cashbackV1Domain.PROGRAM_TYPE_FIXED, "250", start, exp, false)
	if err != nil {
		t.Fatalf("AddCashback: %v", err)
	}
	if len(addOut.States) == 0 {
		t.Fatalf("AddCashback returned empty States")
	}
	if addOut.States[0].Object == nil {
		t.Fatalf("AddCashback returned nil state object")
	}

	var cb cashbackV1Models.CashbackStateModel
	unmarshalState(t, addOut.States[0].Object, &cb)

	// Field validation (todos os campos do model, no espírito do Mint)
	if cb.Address == "" {
		t.Fatalf("AddCashback Address empty")
	}
	if cb.Owner != merchant.PublicKey {
		t.Fatalf("AddCashback Owner mismatch: got %q want %q", cb.Owner, merchant.PublicKey)
	}
	if cb.TokenAddress != tok.Address {
		t.Fatalf("AddCashback TokenAddress mismatch: got %q want %q", cb.TokenAddress, tok.Address)
	}
	if cb.ProgramType != cashbackV1Domain.PROGRAM_TYPE_FIXED {
		t.Fatalf("AddCashback ProgramType mismatch: got %q want %q", cb.ProgramType, cashbackV1Domain.PROGRAM_TYPE_FIXED)
	}
	if cb.Percentage != "250" {
		t.Fatalf("AddCashback Percentage mismatch: got %q want %q", cb.Percentage, "250")
	}
	if cb.StartAt == nil {
		t.Fatalf("AddCashback StartAt nil")
	}
	if cb.ExpiredAt == nil {
		t.Fatalf("AddCashback ExpiredAt nil")
	}
	if cb.Paused != false {
		t.Fatalf("AddCashback Paused mismatch: got %v want %v", cb.Paused, false)
	}
	if cb.Hash == "" {
		t.Fatalf("AddCashback Hash empty")
	}
	if cb.CreatedAt.IsZero() {
		t.Fatalf("AddCashback CreatedAt is zero")
	}
	if cb.UpdatedAt.IsZero() {
		t.Fatalf("AddCashback UpdatedAt is zero")
	}

	log.Printf("AddCashback Output States: %+v", addOut.States)
	log.Printf("AddCashback Output Logs: %+v", addOut.Logs)
	log.Printf("AddCashback Output Delegated Call: %+v", addOut.DelegatedCall)

	log.Printf("AddCashback Address: %s", cb.Address)
	log.Printf("AddCashback Owner: %s", cb.Owner)
	log.Printf("AddCashback TokenAddress: %s", cb.TokenAddress)
	log.Printf("AddCashback ProgramType: %s", cb.ProgramType)
	log.Printf("AddCashback Percentage: %s", cb.Percentage)
	log.Printf("AddCashback StartAt: %s", cb.StartAt.String())
	log.Printf("AddCashback ExpiredAt: %s", cb.ExpiredAt.String())
	log.Printf("AddCashback Paused: %v", cb.Paused)
	log.Printf("AddCashback Hash: %s", cb.Hash)
	log.Printf("AddCashback CreatedAt: %s", cb.CreatedAt.String())
	log.Printf("AddCashback UpdatedAt: %s", cb.UpdatedAt.String())

	// --------------------------------------------------------------------
	// AllowUsers(token) for cashback contract address (envelope + validate + log)
	// --------------------------------------------------------------------
	allowCBOut, err := c.AllowUsers(tok.Address, map[string]bool{cb.Address: true})
	if err != nil {
		t.Fatalf("AllowUsers(cashback addr): %v", err)
	}
	if len(allowCBOut.States) == 0 {
		t.Fatalf("AllowUsers(cashback addr) returned empty States")
	}
	if allowCBOut.States[0].Object == nil {
		t.Fatalf("AllowUsers(cashback addr) returned nil state object")
	}

	var ap2 tokenV1Domain.AccessPolicy
	unmarshalState(t, allowCBOut.States[0].Object, &ap2)

	if ap2.Users == nil || !ap2.Users[cb.Address] {
		t.Fatalf("AllowUsers(cashback addr) missing cashback address in users")
	}

	log.Printf("AllowUsers(cashback addr) Output States: %+v", allowCBOut.States)
	log.Printf("AllowUsers(cashback addr) Output Logs: %+v", allowCBOut.Logs)
	log.Printf("AllowUsers(cashback addr) Output Delegated Call: %+v", allowCBOut.DelegatedCall)

	log.Printf("AllowUsers(cashback addr) Mode: %s", ap2.Mode)
	log.Printf("AllowUsers(cashback addr) Users: %+v", ap2.Users)

	// --------------------------------------------------------------------
	// DepositCashbackFunds (envelope + validate + log)
	// --------------------------------------------------------------------
	depOut, err := c.DepositCashbackFunds(cb.Address, tok.Address, amt(1000, dec), tokenV1Domain.FUNGIBLE, "")
	if err != nil {
		t.Fatalf("DepositCashbackFunds: %v", err)
	}
	if len(depOut.States) == 0 {
		t.Fatalf("DepositCashbackFunds returned empty States")
	}
	if depOut.States[0].Object == nil {
		t.Fatalf("DepositCashbackFunds returned nil state object")
	}

	// se Deposit retorna CashbackStateModel atualizado, unmarshal e valida no mesmo padrão
	var cbAfterDeposit cashbackV1Models.CashbackStateModel
	unmarshalState(t, depOut.States[0].Object, &cbAfterDeposit)

	if cbAfterDeposit.Address != cb.Address {
		t.Fatalf("DepositCashbackFunds Address mismatch: got %q want %q", cbAfterDeposit.Address, cb.Address)
	}
	if cbAfterDeposit.TokenAddress != tok.Address {
		t.Fatalf("DepositCashbackFunds TokenAddress mismatch: got %q want %q", cbAfterDeposit.TokenAddress, tok.Address)
	}

	log.Printf("DepositCashbackFunds Output States: %+v", depOut.States)
	log.Printf("DepositCashbackFunds Output Logs: %+v", depOut.Logs)
	log.Printf("DepositCashbackFunds Output Delegated Call: %+v", depOut.DelegatedCall)

	log.Printf("DepositCashbackFunds Address: %s", cbAfterDeposit.Address)
	log.Printf("DepositCashbackFunds Owner: %s", cbAfterDeposit.Owner)
	log.Printf("DepositCashbackFunds TokenAddress: %s", cbAfterDeposit.TokenAddress)
	log.Printf("DepositCashbackFunds ProgramType: %s", cbAfterDeposit.ProgramType)
	log.Printf("DepositCashbackFunds Percentage: %s", cbAfterDeposit.Percentage)
	log.Printf("DepositCashbackFunds StartAt: %s", cbAfterDeposit.StartAt.String())
	log.Printf("DepositCashbackFunds ExpiredAt: %s", cbAfterDeposit.ExpiredAt.String())
	log.Printf("DepositCashbackFunds Paused: %v", cbAfterDeposit.Paused)
	log.Printf("DepositCashbackFunds Hash: %s", cbAfterDeposit.Hash)
	log.Printf("DepositCashbackFunds CreatedAt: %s", cbAfterDeposit.CreatedAt.String())
	log.Printf("DepositCashbackFunds UpdatedAt: %s", cbAfterDeposit.UpdatedAt.String())

	// --------------------------------------------------------------------
	// UpdateCashback (envelope + unmarshal + validate + log)
	// --------------------------------------------------------------------
	updOut, err := c.UpdateCashback(cb.Address, tok.Address, cashbackV1Domain.PROGRAM_TYPE_FIXED, "300", start, exp)
	if err != nil {
		t.Fatalf("UpdateCashback: %v", err)
	}
	if len(updOut.States) == 0 {
		t.Fatalf("UpdateCashback returned empty States")
	}
	if updOut.States[0].Object == nil {
		t.Fatalf("UpdateCashback returned nil state object")
	}

	var cbUpdated cashbackV1Models.CashbackStateModel
	unmarshalState(t, updOut.States[0].Object, &cbUpdated)

	if cbUpdated.Address != cb.Address {
		t.Fatalf("UpdateCashback Address mismatch: got %q want %q", cbUpdated.Address, cb.Address)
	}
	if cbUpdated.TokenAddress != tok.Address {
		t.Fatalf("UpdateCashback TokenAddress mismatch: got %q want %q", cbUpdated.TokenAddress, tok.Address)
	}
	if cbUpdated.ProgramType != cashbackV1Domain.PROGRAM_TYPE_FIXED {
		t.Fatalf("UpdateCashback ProgramType mismatch: got %q want %q", cbUpdated.ProgramType, cashbackV1Domain.PROGRAM_TYPE_FIXED)
	}
	if cbUpdated.Percentage != "300" {
		t.Fatalf("UpdateCashback Percentage mismatch: got %q want %q", cbUpdated.Percentage, "300")
	}
	if cbUpdated.StartAt == nil {
		t.Fatalf("UpdateCashback StartAt nil")
	}
	if cbUpdated.ExpiredAt == nil {
		t.Fatalf("UpdateCashback ExpiredAt nil")
	}

	log.Printf("UpdateCashback Output States: %+v", updOut.States)
	log.Printf("UpdateCashback Output Logs: %+v", updOut.Logs)
	log.Printf("UpdateCashback Output Delegated Call: %+v", updOut.DelegatedCall)

	log.Printf("UpdateCashback Address: %s", cbUpdated.Address)
	log.Printf("UpdateCashback Owner: %s", cbUpdated.Owner)
	log.Printf("UpdateCashback TokenAddress: %s", cbUpdated.TokenAddress)
	log.Printf("UpdateCashback ProgramType: %s", cbUpdated.ProgramType)
	log.Printf("UpdateCashback Percentage: %s", cbUpdated.Percentage)
	log.Printf("UpdateCashback StartAt: %s", cbUpdated.StartAt.String())
	log.Printf("UpdateCashback ExpiredAt: %s", cbUpdated.ExpiredAt.String())
	log.Printf("UpdateCashback Paused: %v", cbUpdated.Paused)
	log.Printf("UpdateCashback Hash: %s", cbUpdated.Hash)
	log.Printf("UpdateCashback CreatedAt: %s", cbUpdated.CreatedAt.String())
	log.Printf("UpdateCashback UpdatedAt: %s", cbUpdated.UpdatedAt.String())

	// --------------------------------------------------------------------
	// Pause / Unpause (envelope + unmarshal + validate + log)
	// --------------------------------------------------------------------
	pauseOut, err := c.PauseCashback(cb.Address, true)
	if err != nil {
		t.Fatalf("PauseCashback: %v", err)
	}
	if len(pauseOut.States) == 0 {
		t.Fatalf("PauseCashback returned empty States")
	}
	if pauseOut.States[0].Object == nil {
		t.Fatalf("PauseCashback returned nil state object")
	}

	var cbPaused cashbackV1Models.CashbackStateModel
	unmarshalState(t, pauseOut.States[0].Object, &cbPaused)

	if cbPaused.Address != cb.Address {
		t.Fatalf("PauseCashback Address mismatch: got %q want %q", cbPaused.Address, cb.Address)
	}
	if !cbPaused.Paused {
		t.Fatalf("PauseCashback expected Paused=true")
	}

	log.Printf("PauseCashback Output States: %+v", pauseOut.States)
	log.Printf("PauseCashback Output Logs: %+v", pauseOut.Logs)
	log.Printf("PauseCashback Output Delegated Call: %+v", pauseOut.DelegatedCall)

	log.Printf("PauseCashback Address: %s", cbPaused.Address)
	log.Printf("PauseCashback Paused: %v", cbPaused.Paused)

	unpauseOut, err := c.UnpauseCashback(cb.Address, false)
	if err != nil {
		t.Fatalf("UnpauseCashback: %v", err)
	}
	if len(unpauseOut.States) == 0 {
		t.Fatalf("UnpauseCashback returned empty States")
	}
	if unpauseOut.States[0].Object == nil {
		t.Fatalf("UnpauseCashback returned nil state object")
	}

	var cbUnpaused cashbackV1Models.CashbackStateModel
	unmarshalState(t, unpauseOut.States[0].Object, &cbUnpaused)

	if cbUnpaused.Address != cb.Address {
		t.Fatalf("UnpauseCashback Address mismatch: got %q want %q", cbUnpaused.Address, cb.Address)
	}
	if cbUnpaused.Paused {
		t.Fatalf("UnpauseCashback expected Paused=false")
	}

	log.Printf("UnpauseCashback Output States: %+v", unpauseOut.States)
	log.Printf("UnpauseCashback Output Logs: %+v", unpauseOut.Logs)
	log.Printf("UnpauseCashback Output Delegated Call: %+v", unpauseOut.DelegatedCall)

	log.Printf("UnpauseCashback Address: %s", cbUnpaused.Address)
	log.Printf("UnpauseCashback Paused: %v", cbUnpaused.Paused)

	// --------------------------------------------------------------------
	// Claim (user) (envelope + validate + log)
	// --------------------------------------------------------------------
	time.Sleep(2 * time.Second)

	user, userPriv := createWallet(t, c)

	c.SetPrivateKey(ownerPriv)
	allowUserOut, err := c.AllowUsers(tok.Address, map[string]bool{user.PublicKey: true})
	if err != nil {
		t.Fatalf("AllowUsers(user): %v", err)
	}
	if len(allowUserOut.States) == 0 {
		t.Fatalf("AllowUsers(user) returned empty States")
	}
	if allowUserOut.States[0].Object == nil {
		t.Fatalf("AllowUsers(user) returned nil state object")
	}

	var ap3 tokenV1Domain.AccessPolicy
	unmarshalState(t, allowUserOut.States[0].Object, &ap3)

	if ap3.Users == nil || !ap3.Users[user.PublicKey] {
		t.Fatalf("AllowUsers(user) missing user in users")
	}

	log.Printf("AllowUsers(user) Output States: %+v", allowUserOut.States)
	log.Printf("AllowUsers(user) Output Logs: %+v", allowUserOut.Logs)
	log.Printf("AllowUsers(user) Output Delegated Call: %+v", allowUserOut.DelegatedCall)

	log.Printf("AllowUsers(user) Mode: %s", ap3.Mode)
	log.Printf("AllowUsers(user) Users: %+v", ap3.Users)

	c.SetPrivateKey(userPriv)

	claimOut, err := c.ClaimCashback(cb.Address, amt(100, dec), tokenV1Domain.FUNGIBLE, "")
	if err != nil {
		t.Fatalf("ClaimCashback: %v", err)
	}
	if len(claimOut.States) == 0 {
		t.Fatalf("ClaimCashback returned empty States")
	}
	if claimOut.States[0].Object == nil {
		t.Fatalf("ClaimCashback returned nil state object")
	}

	var cbAfterClaim cashbackV1Models.CashbackStateModel
	unmarshalState(t, claimOut.States[0].Object, &cbAfterClaim)

	if cbAfterClaim.Address != cb.Address {
		t.Fatalf("ClaimCashback Address mismatch: got %q want %q", cbAfterClaim.Address, cb.Address)
	}

	log.Printf("ClaimCashback Output States: %+v", claimOut.States)
	log.Printf("ClaimCashback Output Logs: %+v", claimOut.Logs)
	log.Printf("ClaimCashback Output Delegated Call: %+v", claimOut.DelegatedCall)

	log.Printf("ClaimCashback Address: %s", cbAfterClaim.Address)
	log.Printf("ClaimCashback Owner: %s", cbAfterClaim.Owner)
	log.Printf("ClaimCashback TokenAddress: %s", cbAfterClaim.TokenAddress)
	log.Printf("ClaimCashback ProgramType: %s", cbAfterClaim.ProgramType)
	log.Printf("ClaimCashback Percentage: %s", cbAfterClaim.Percentage)
	log.Printf("ClaimCashback StartAt: %s", cbAfterClaim.StartAt.String())
	log.Printf("ClaimCashback ExpiredAt: %s", cbAfterClaim.ExpiredAt.String())
	log.Printf("ClaimCashback Paused: %v", cbAfterClaim.Paused)
	log.Printf("ClaimCashback Hash: %s", cbAfterClaim.Hash)
	log.Printf("ClaimCashback CreatedAt: %s", cbAfterClaim.CreatedAt.String())
	log.Printf("ClaimCashback UpdatedAt: %s", cbAfterClaim.UpdatedAt.String())

	// --------------------------------------------------------------------
	// Getters (envelope + unmarshal + validate + log)
	// --------------------------------------------------------------------
	getOut, err := c.GetCashback(cb.Address)
	if err != nil {
		t.Fatalf("GetCashback: %v", err)
	}
	if len(getOut.States) == 0 {
		t.Fatalf("GetCashback returned empty States")
	}
	if getOut.States[0].Object == nil {
		t.Fatalf("GetCashback returned nil state object")
	}

	var cbGet cashbackV1Models.CashbackStateModel
	unmarshalState(t, getOut.States[0].Object, &cbGet)

	if cbGet.Address != cb.Address {
		t.Fatalf("GetCashback Address mismatch: got %q want %q", cbGet.Address, cb.Address)
	}

	log.Printf("GetCashback Output States: %+v", getOut.States)
	log.Printf("GetCashback Output Logs: %+v", getOut.Logs)
	log.Printf("GetCashback Output Delegated Call: %+v", getOut.DelegatedCall)

	log.Printf("GetCashback Address: %s", cbGet.Address)
	log.Printf("GetCashback Owner: %s", cbGet.Owner)
	log.Printf("GetCashback TokenAddress: %s", cbGet.TokenAddress)
	log.Printf("GetCashback ProgramType: %s", cbGet.ProgramType)
	log.Printf("GetCashback Percentage: %s", cbGet.Percentage)
	log.Printf("GetCashback StartAt: %s", cbGet.StartAt.String())
	log.Printf("GetCashback ExpiredAt: %s", cbGet.ExpiredAt.String())
	log.Printf("GetCashback Paused: %v", cbGet.Paused)
	log.Printf("GetCashback Hash: %s", cbGet.Hash)
	log.Printf("GetCashback CreatedAt: %s", cbGet.CreatedAt.String())
	log.Printf("GetCashback UpdatedAt: %s", cbGet.UpdatedAt.String())

	listOut, err := c.ListCashbacks(merchant.PublicKey, tok.Address, "", false, 1, 10, true)
	if err != nil {
		t.Fatalf("ListCashbacks: %v", err)
	}
	if len(listOut.States) == 0 {
		t.Fatalf("ListCashbacks returned empty States")
	}
	if listOut.States[0].Object == nil {
		t.Fatalf("ListCashbacks returned nil state object")
	}

	var list []cashbackV1Models.CashbackStateModel
	unmarshalState(t, listOut.States[0].Object, &list)

	found := false
	for _, it := range list {
		if it.Address == cb.Address {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("ListCashbacks: cashback %s not found", cb.Address)
	}

	log.Printf("ListCashbacks Output States: %+v", listOut.States)
	log.Printf("ListCashbacks Output Logs: %+v", listOut.Logs)
	log.Printf("ListCashbacks Output Delegated Call: %+v", listOut.DelegatedCall)

	log.Printf("ListCashbacks Count: %d", len(list))
}

func TestCashbackFlow_NonFungible(t *testing.T) {
	c := setupClient(t)

	owner, ownerPriv := createWallet(t, c)
	c.SetPrivateKey(ownerPriv)

	dec := 0
	tokenType := tokenV1Domain.NON_FUNGIBLE
	stablecoin := false

	tok := createBasicToken(t, c, owner.PublicKey, dec, false, tokenType, stablecoin)

	// Token (mínimo) validate + log
	if tok.Address == "" {
		t.Fatalf("Token address empty")
	}
	if tok.TokenType != tokenType {
		t.Fatalf("Token type mismatch: got %s want %s", tok.TokenType, tokenType)
	}
	log.Printf("Token Address: %s", tok.Address)
	log.Printf("Token Type: %s", tok.TokenType)

	amount := "1"

	// Mint NFT (envelope + validate + log)
	mintOut, err := c.MintToken(tok.Address, owner.PublicKey, amount, dec, tok.TokenType)
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

	merchant, merchantPriv := createWallet(t, c)

	// AllowUsers(merchant) (envelope + validate + log)
	c.SetPrivateKey(ownerPriv)
	allowMerchantOut, err := c.AllowUsers(tok.Address, map[string]bool{merchant.PublicKey: true})
	if err != nil {
		t.Fatalf("AllowUsers(merchant): %v", err)
	}
	if len(allowMerchantOut.States) == 0 {
		t.Fatalf("AllowUsers(merchant) returned empty States")
	}
	if allowMerchantOut.States[0].Object == nil {
		t.Fatalf("AllowUsers(merchant) returned nil state object")
	}

	var ap tokenV1Domain.AccessPolicy
	unmarshalState(t, allowMerchantOut.States[0].Object, &ap)

	if ap.Users == nil || !ap.Users[merchant.PublicKey] {
		t.Fatalf("AllowUsers(merchant) missing merchant in users")
	}

	log.Printf("AllowUsers(merchant) Output States: %+v", allowMerchantOut.States)
	log.Printf("AllowUsers(merchant) Output Logs: %+v", allowMerchantOut.Logs)
	log.Printf("AllowUsers(merchant) Output Delegated Call: %+v", allowMerchantOut.DelegatedCall)

	log.Printf("AllowUsers Mode: %s", ap.Mode)
	log.Printf("AllowUsers Users: %+v", ap.Users)

	// Transfer NFT (envelope + validate + log)
	trOut, err := c.TransferToken(tok.Address, merchant.PublicKey, amount, dec, tok.TokenType, nftUUID)
	if err != nil {
		t.Fatalf("Transfer NFT: %v", err)
	}
	if len(trOut.States) == 0 {
		t.Fatalf("Transfer NFT returned empty States")
	}
	if trOut.States[0].Object == nil {
		t.Fatalf("Transfer NFT returned nil state object")
	}

	var tr tokenV1Domain.Transfer
	unmarshalState(t, trOut.States[0].Object, &tr)

	if tr.TokenAddress != tok.Address {
		t.Fatalf("Transfer NFT token_address mismatch: got %q want %q", tr.TokenAddress, tok.Address)
	}
	if tr.ToAddress != merchant.PublicKey {
		t.Fatalf("Transfer NFT to mismatch: got %q want %q", tr.ToAddress, merchant.PublicKey)
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
	// Deploy Cashback contract (envelope + unmarshal + validate + log)
	// --------------------------------------------------------------------
	start := time.Now().Add(2 * time.Second)
	exp := time.Now().Add(30 * time.Minute)

	contractState := models.ContractStateModel{}
	deployedContract, err := c.DeployContract1(cashbackV1.CASHBACK_CONTRACT_V1)
	if err != nil {
		t.Fatalf("DeployContract: %v", err)
	}
	if len(deployedContract.States) == 0 {
		t.Fatalf("DeployContract returned empty States")
	}
	if deployedContract.States[0].Object == nil {
		t.Fatalf("DeployContract returned nil state object")
	}

	unmarshalState(t, deployedContract.States[0].Object, &contractState)

	if contractState.Address == "" {
		t.Fatalf("DeployContract returned empty contract address")
	}
	address := contractState.Address

	log.Printf("DeployContract(Cashback) Output States: %+v", deployedContract.States)
	log.Printf("DeployContract(Cashback) Output Logs: %+v", deployedContract.Logs)
	log.Printf("DeployContract(Cashback) Output Delegated Call: %+v", deployedContract.DelegatedCall)

	log.Printf("Cashback Contract Address: %s", address)

	// --------------------------------------------------------------------
	// Merchant cria cashback (AddCashback) (envelope + unmarshal + validate + log)
	// --------------------------------------------------------------------
	c.SetPrivateKey(merchantPriv)

	addOut, err := c.AddCashback(address, merchant.PublicKey, tok.Address, cashbackV1Domain.PROGRAM_TYPE_FIXED, "10000", start, exp, false)
	if err != nil {
		t.Fatalf("AddCashback: %v", err)
	}
	if len(addOut.States) == 0 {
		t.Fatalf("AddCashback returned empty States")
	}
	if addOut.States[0].Object == nil {
		t.Fatalf("AddCashback returned nil state object")
	}

	var cb cashbackV1Models.CashbackStateModel
	unmarshalState(t, addOut.States[0].Object, &cb)

	if cb.Address == "" {
		t.Fatalf("AddCashback Address empty")
	}
	if cb.Owner != merchant.PublicKey {
		t.Fatalf("AddCashback Owner mismatch: got %q want %q", cb.Owner, merchant.PublicKey)
	}
	if cb.TokenAddress != tok.Address {
		t.Fatalf("AddCashback TokenAddress mismatch: got %q want %q", cb.TokenAddress, tok.Address)
	}
	if cb.ProgramType != cashbackV1Domain.PROGRAM_TYPE_FIXED {
		t.Fatalf("AddCashback ProgramType mismatch: got %q want %q", cb.ProgramType, cashbackV1Domain.PROGRAM_TYPE_FIXED)
	}
	if cb.Percentage != "10000" {
		t.Fatalf("AddCashback Percentage mismatch: got %q want %q", cb.Percentage, "10000")
	}
	if cb.StartAt == nil || cb.ExpiredAt == nil {
		t.Fatalf("AddCashback StartAt/ExpiredAt nil: start=%v exp=%v", cb.StartAt, cb.ExpiredAt)
	}
	if cb.Hash == "" {
		t.Fatalf("AddCashback Hash empty")
	}

	log.Printf("AddCashback Output States: %+v", addOut.States)
	log.Printf("AddCashback Output Logs: %+v", addOut.Logs)
	log.Printf("AddCashback Output Delegated Call: %+v", addOut.DelegatedCall)

	log.Printf("AddCashback Address: %s", cb.Address)
	log.Printf("AddCashback Owner: %s", cb.Owner)
	log.Printf("AddCashback TokenAddress: %s", cb.TokenAddress)
	log.Printf("AddCashback ProgramType: %s", cb.ProgramType)
	log.Printf("AddCashback Percentage: %s", cb.Percentage)
	log.Printf("AddCashback StartAt: %s", cb.StartAt.String())
	log.Printf("AddCashback ExpiredAt: %s", cb.ExpiredAt.String())
	log.Printf("AddCashback Paused: %v", cb.Paused)
	log.Printf("AddCashback Hash: %s", cb.Hash)

	// AllowUsers(cashback addr) (envelope + validate + log)
	c.SetPrivateKey(ownerPriv)
	allowCBOut, err := c.AllowUsers(tok.Address, map[string]bool{cb.Address: true})
	if err != nil {
		t.Fatalf("AllowUsers(cashback contract): %v", err)
	}
	if len(allowCBOut.States) == 0 {
		t.Fatalf("AllowUsers(cashback contract) returned empty States")
	}
	if allowCBOut.States[0].Object == nil {
		t.Fatalf("AllowUsers(cashback contract) returned nil state object")
	}

	var ap2 tokenV1Domain.AccessPolicy
	unmarshalState(t, allowCBOut.States[0].Object, &ap2)

	if ap2.Users == nil || !ap2.Users[cb.Address] {
		t.Fatalf("AllowUsers(cashback contract) missing cashback address in users")
	}

	log.Printf("AllowUsers(cashback contract) Output States: %+v", allowCBOut.States)
	log.Printf("AllowUsers(cashback contract) Output Logs: %+v", allowCBOut.Logs)
	log.Printf("AllowUsers(cashback contract) Output Delegated Call: %+v", allowCBOut.DelegatedCall)

	log.Printf("AllowUsers(cashback contract) Mode: %s", ap2.Mode)
	log.Printf("AllowUsers(cashback contract) Users: %+v", ap2.Users)

	// Deposit NFT (MERCHANT assina) (envelope + validate + log)
	c.SetPrivateKey(merchantPriv)
	depOut, err := c.DepositCashbackFunds(cb.Address, tok.Address, amount, tokenV1Domain.NON_FUNGIBLE, nftUUID)
	if err != nil {
		t.Fatalf("DepositCashbackFunds NFT: %v", err)
	}
	if len(depOut.States) == 0 {
		t.Fatalf("DepositCashbackFunds NFT returned empty States")
	}
	if depOut.States[0].Object == nil {
		t.Fatalf("DepositCashbackFunds NFT returned nil state object")
	}

	var cbAfterDeposit cashbackV1Models.CashbackStateModel
	unmarshalState(t, depOut.States[0].Object, &cbAfterDeposit)

	if cbAfterDeposit.Address != cb.Address {
		t.Fatalf("DepositCashbackFunds NFT Address mismatch: got %q want %q", cbAfterDeposit.Address, cb.Address)
	}

	log.Printf("DepositCashbackFunds NFT Output States: %+v", depOut.States)
	log.Printf("DepositCashbackFunds NFT Output Logs: %+v", depOut.Logs)
	log.Printf("DepositCashbackFunds NFT Output Delegated Call: %+v", depOut.DelegatedCall)

	log.Printf("DepositCashbackFunds NFT Address: %s", cbAfterDeposit.Address)
	log.Printf("DepositCashbackFunds NFT Owner: %s", cbAfterDeposit.Owner)
	log.Printf("DepositCashbackFunds NFT TokenAddress: %s", cbAfterDeposit.TokenAddress)
	log.Printf("DepositCashbackFunds NFT ProgramType: %s", cbAfterDeposit.ProgramType)
	log.Printf("DepositCashbackFunds NFT Percentage: %s", cbAfterDeposit.Percentage)
	log.Printf("DepositCashbackFunds NFT StartAt: %s", cbAfterDeposit.StartAt.String())
	log.Printf("DepositCashbackFunds NFT ExpiredAt: %s", cbAfterDeposit.ExpiredAt.String())
	log.Printf("DepositCashbackFunds NFT Paused: %v", cbAfterDeposit.Paused)
	log.Printf("DepositCashbackFunds NFT Hash: %s", cbAfterDeposit.Hash)

	// Update / Pause / Unpause (envelope + unmarshal + validate + log)
	updOut, err := c.UpdateCashback(cb.Address, tok.Address, cashbackV1Domain.PROGRAM_TYPE_FIXED, "10000", start, exp)
	if err != nil {
		t.Fatalf("UpdateCashback: %v", err)
	}
	if len(updOut.States) == 0 {
		t.Fatalf("UpdateCashback returned empty States")
	}
	if updOut.States[0].Object == nil {
		t.Fatalf("UpdateCashback returned nil state object")
	}

	var cbUpdated cashbackV1Models.CashbackStateModel
	unmarshalState(t, updOut.States[0].Object, &cbUpdated)

	if cbUpdated.Address != cb.Address {
		t.Fatalf("UpdateCashback Address mismatch: got %q want %q", cbUpdated.Address, cb.Address)
	}

	log.Printf("UpdateCashback Output States: %+v", updOut.States)
	log.Printf("UpdateCashback Output Logs: %+v", updOut.Logs)
	log.Printf("UpdateCashback Output Delegated Call: %+v", updOut.DelegatedCall)

	log.Printf("UpdateCashback Address: %s", cbUpdated.Address)
	log.Printf("UpdateCashback ProgramType: %s", cbUpdated.ProgramType)
	log.Printf("UpdateCashback Percentage: %s", cbUpdated.Percentage)

	pauseOut, err := c.PauseCashback(cb.Address, true)
	if err != nil {
		t.Fatalf("PauseCashback: %v", err)
	}
	if len(pauseOut.States) == 0 || pauseOut.States[0].Object == nil {
		t.Fatalf("PauseCashback returned empty/nil state")
	}
	var cbPaused cashbackV1Models.CashbackStateModel
	unmarshalState(t, pauseOut.States[0].Object, &cbPaused)
	if !cbPaused.Paused {
		t.Fatalf("PauseCashback expected Paused=true")
	}

	log.Printf("PauseCashback Output States: %+v", pauseOut.States)
	log.Printf("PauseCashback Output Logs: %+v", pauseOut.Logs)
	log.Printf("PauseCashback Output Delegated Call: %+v", pauseOut.DelegatedCall)

	log.Printf("PauseCashback Address: %s", cbPaused.Address)
	log.Printf("PauseCashback Paused: %v", cbPaused.Paused)

	unpauseOut, err := c.UnpauseCashback(cb.Address, false)
	if err != nil {
		t.Fatalf("UnpauseCashback: %v", err)
	}
	if len(unpauseOut.States) == 0 || unpauseOut.States[0].Object == nil {
		t.Fatalf("UnpauseCashback returned empty/nil state")
	}
	var cbUnpaused cashbackV1Models.CashbackStateModel
	unmarshalState(t, unpauseOut.States[0].Object, &cbUnpaused)
	if cbUnpaused.Paused {
		t.Fatalf("UnpauseCashback expected Paused=false")
	}

	log.Printf("UnpauseCashback Output States: %+v", unpauseOut.States)
	log.Printf("UnpauseCashback Output Logs: %+v", unpauseOut.Logs)
	log.Printf("UnpauseCashback Output Delegated Call: %+v", unpauseOut.DelegatedCall)

	log.Printf("UnpauseCashback Address: %s", cbUnpaused.Address)
	log.Printf("UnpauseCashback Paused: %v", cbUnpaused.Paused)

	// Claim as user (envelope + validate + log)
	time.Sleep(2 * time.Second)

	user, userPriv := createWallet(t, c)

	c.SetPrivateKey(ownerPriv)
	allowUserOut, err := c.AllowUsers(tok.Address, map[string]bool{user.PublicKey: true})
	if err != nil {
		t.Fatalf("AllowUsers(user): %v", err)
	}
	if len(allowUserOut.States) == 0 || allowUserOut.States[0].Object == nil {
		t.Fatalf("AllowUsers(user) returned empty/nil state")
	}

	c.SetPrivateKey(userPriv)
	claimOut, err := c.ClaimCashback(cb.Address, amount, tokenV1Domain.NON_FUNGIBLE, nftUUID)
	if err != nil {
		t.Fatalf("ClaimCashback NFT: %v", err)
	}
	if len(claimOut.States) == 0 || claimOut.States[0].Object == nil {
		t.Fatalf("ClaimCashback NFT returned empty/nil state")
	}

	var cbAfterClaim cashbackV1Models.CashbackStateModel
	unmarshalState(t, claimOut.States[0].Object, &cbAfterClaim)

	if cbAfterClaim.Address != cb.Address {
		t.Fatalf("ClaimCashback NFT Address mismatch: got %q want %q", cbAfterClaim.Address, cb.Address)
	}

	log.Printf("ClaimCashback NFT Output States: %+v", claimOut.States)
	log.Printf("ClaimCashback NFT Output Logs: %+v", claimOut.Logs)
	log.Printf("ClaimCashback NFT Output Delegated Call: %+v", claimOut.DelegatedCall)

	log.Printf("ClaimCashback NFT Address: %s", cbAfterClaim.Address)
	log.Printf("ClaimCashback NFT Owner: %s", cbAfterClaim.Owner)
	log.Printf("ClaimCashback NFT TokenAddress: %s", cbAfterClaim.TokenAddress)
	log.Printf("ClaimCashback NFT ProgramType: %s", cbAfterClaim.ProgramType)
	log.Printf("ClaimCashback NFT Percentage: %s", cbAfterClaim.Percentage)
	log.Printf("ClaimCashback NFT StartAt: %s", cbAfterClaim.StartAt.String())
	log.Printf("ClaimCashback NFT ExpiredAt: %s", cbAfterClaim.ExpiredAt.String())
	log.Printf("ClaimCashback NFT Paused: %v", cbAfterClaim.Paused)
	log.Printf("ClaimCashback NFT Hash: %s", cbAfterClaim.Hash)

	// Getters (envelope + unmarshal + validate + log)
	getOut, err := c.GetCashback(cb.Address)
	if err != nil {
		t.Fatalf("GetCashback: %v", err)
	}
	if len(getOut.States) == 0 || getOut.States[0].Object == nil {
		t.Fatalf("GetCashback returned empty/nil state")
	}

	var cbGet cashbackV1Models.CashbackStateModel
	unmarshalState(t, getOut.States[0].Object, &cbGet)

	if cbGet.Address != cb.Address {
		t.Fatalf("GetCashback Address mismatch: got %q want %q", cbGet.Address, cb.Address)
	}

	log.Printf("GetCashback Output States: %+v", getOut.States)
	log.Printf("GetCashback Output Logs: %+v", getOut.Logs)
	log.Printf("GetCashback Output Delegated Call: %+v", getOut.DelegatedCall)

	log.Printf("GetCashback Address: %s", cbGet.Address)
	log.Printf("GetCashback Owner: %s", cbGet.Owner)
	log.Printf("GetCashback TokenAddress: %s", cbGet.TokenAddress)
	log.Printf("GetCashback ProgramType: %s", cbGet.ProgramType)
	log.Printf("GetCashback Percentage: %s", cbGet.Percentage)
	log.Printf("GetCashback StartAt: %s", cbGet.StartAt.String())
	log.Printf("GetCashback ExpiredAt: %s", cbGet.ExpiredAt.String())
	log.Printf("GetCashback Paused: %v", cbGet.Paused)
	log.Printf("GetCashback Hash: %s", cbGet.Hash)

	listOut, err := c.ListCashbacks(merchant.PublicKey, tok.Address, "", false, 1, 10, true)
	if err != nil {
		t.Fatalf("ListCashbacks: %v", err)
	}
	if len(listOut.States) == 0 || listOut.States[0].Object == nil {
		t.Fatalf("ListCashbacks returned empty/nil state")
	}

	var list []cashbackV1Models.CashbackStateModel
	unmarshalState(t, listOut.States[0].Object, &list)

	found := false
	for _, it := range list {
		if it.Address == cb.Address {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("ListCashbacks: cashback %s not found", cb.Address)
	}

	log.Printf("ListCashbacks Output States: %+v", listOut.States)
	log.Printf("ListCashbacks Output Logs: %+v", listOut.Logs)
	log.Printf("ListCashbacks Output Delegated Call: %+v", listOut.DelegatedCall)

	log.Printf("ListCashbacks Count: %d", len(list))
}
