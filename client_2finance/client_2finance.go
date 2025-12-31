package client_2finance

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"gitlab.com/2finance/2finance-network/blockchain/block"
	"gitlab.com/2finance/2finance-network/blockchain/handler"
	"gitlab.com/2finance/2finance-network/blockchain/contract/contractV1"
	"gitlab.com/2finance/2finance-network/blockchain/encryption/keys"
	blockchainLog "gitlab.com/2finance/2finance-network/blockchain/log"
	"gitlab.com/2finance/2finance-network/blockchain/transaction"
	"gitlab.com/2finance/2finance-network/blockchain/types"
	"gitlab.com/2finance/2finance-network/blockchain/utils"
	"gitlab.com/2finance/2finance-network/infra/mqtt"

	"strings"

	"github.com/google/uuid"
	"gitlab.com/2finance/2finance-network/infra/event"
)

// Interface exposes the client behavior
type Client2FinanceNetwork interface {
	// Client
	SetPrivateKey(privateKey string)
	GetPrivateKey() string
	GetPublicKey() string
	GenerateKeyEd25519() (string, string, error)

	SendTransaction(method string, tx interface{}, replyTo string) (outputBytes []byte, err error)

	// CHAIN
	GetNonce(publicKey string) (uint64, error)
	ListTransactions(from, to, hash string, dataFilter map[string]interface{}, nonce uint64,
		page, limit int,
		ascending bool) ([]transaction.Transaction, error)
	ListLogs(logType []string, logIndex uint, transactionHash string, event map[string]interface{}, contractAddress string,
		page, limit int,
		ascending bool) ([]blockchainLog.Log, error)
	DeployContract1(
		contractVersion string,
	) (types.ContractOutput, error)
	DeployContract2(
		contractVersion string, 
		contractAddress string,
		) (types.ContractOutput, error)
	SignTransaction(from, to, method string, data utils.JSONB, nonce uint64) (*transaction.Transaction, error)
	SignAndSendTransaction(
		from string,
		to string,
		method string,
		data map[string]interface{}) (types.ContractOutput, error)
	GetState(
		to string,
		method string,
		data map[string]interface{}) (types.ContractOutput, error)
	ListBlocks(blockNumber uint64, blockTimestamp time.Time, hash string, previousHash string,
		merkleRoot string,
		page, limit int,
		ascending bool) ([]block.Block, error)

	// WALLET
	AddWallet(address, pubKey string) (types.ContractOutput, error)
	GetWallet(pubKey string) (types.ContractOutput, error)
	TransferWallet(to, amount string, decimals int) (types.ContractOutput, error)

	// TOKEN
	AddToken(
		address string,
		symbol string,
		name string,
		decimals int,
		totalSupply string,
		description string,
		owner string,
		image string,
		website string,
		tagsSocialMedia map[string]string,
		tagsCategory map[string]string,
		tags map[string]string,
		creator string,
		creatorWebsite string,
		allowUsers map[string]bool,
		blockUsers map[string]bool,
		feeTiersList []map[string]interface{},
		feeAddress string,
		freezeAuthorityRevoked bool,
		mintAuthorityRevoked bool,
		updateAuthorityRevoked bool,
		paused bool,
		expired_at time.Time,
		assetGLBUri string,
		tokenType string) (types.ContractOutput, error)
	MintToken(to, mintTo, amount string, decimals int, tokenType string) (types.ContractOutput, error)
	BurnToken(to, amount string, decimals int, tokenType string, uuid string) (types.ContractOutput, error)
	TransferToken(tokenAddress, transferTo, amount string, decimals int, tokenType string, uuid string) (types.ContractOutput, error)
	AllowUsers(tokenAddress string, users map[string]bool) (types.ContractOutput, error)
	DisallowUsers(tokenAddress string, users map[string]bool) (types.ContractOutput, error)
	BlockUsers(tokenAddress string, users map[string]bool) (types.ContractOutput, error)
	UnblockUsers(tokenAddress string, users map[string]bool) (types.ContractOutput, error)
	RevokeFreezeAuthority(tokenAddress string, revoke bool) (types.ContractOutput, error)
	RevokeMintAuthority(tokenAddress string, revoke bool) (types.ContractOutput, error)
	RevokeUpdateAuthority(tokenAddress string, revoke bool) (types.ContractOutput, error)
	UpdateMetadata(tokenAddress, symbol, name string, decimals int, description, image, website string,
		tagsSocialMedia, tagsCategory, tags map[string]string,
		creator, creatorWebsite string, expired_at time.Time) (types.ContractOutput, error)
	PauseToken(tokenAddress string, pause bool) (types.ContractOutput, error)
	
	UnpauseToken(tokenAddress string, unpause bool) (types.ContractOutput, error)
	UpdateFeeTiers(tokenAddress string, feeTierList []map[string]interface{}) (types.ContractOutput, error)
	UpdateFeeAddress(tokenAddress, feeAddress string) (types.ContractOutput, error)
	GetToken(tokenAddress string, symbol string, name string) (types.ContractOutput, error)
	ListTokens(ownerAddress, symbol, name string, page, limit int, ascending bool) (types.ContractOutput, error)

	GetTokenBalance(tokenAddress, ownerAddress string) (types.ContractOutput, error)
	ListTokenBalances(tokenAddress, ownerAddress string, page, limit int, ascending bool) (types.ContractOutput, error)

	// FAUCET
	AddFaucet(
		address string,
		owner string,
		tokenAddress string,
		startTime time.Time,
		expireTime time.Time,
		paused bool,
		requestLimit int,
		claimAmount string,
		claimIntervalDuration time.Duration,
	) (types.ContractOutput, error)
	UpdateFaucet(
		address string,
		startTime time.Time,
		expireTime time.Time,
		requestLimit int,
		requestsByUser map[string]int,
		claimAmount string,
		claimIntervalDuration time.Duration,
		lastClaimByUser map[string]time.Time,
	) (types.ContractOutput, error)
	DepositFunds(address, tokenAddress, amount string) (types.ContractOutput, error)
	WithdrawFunds(address, tokenAddress, amount string) (types.ContractOutput, error)
	PauseFaucet(address string, pause bool) (types.ContractOutput, error)
	UnpauseFaucet(address string, pause bool) (types.ContractOutput, error)
	UpdateRequestLimitPerUser(address string, requestLimit int) (types.ContractOutput, error)
	ClaimFunds(address string) (types.ContractOutput, error)

	GetFaucet(faucetAddress string) (types.ContractOutput, error)
	ListFaucets(
		ownerAddress string,
		page, limit int,
		ascending bool,
	) (types.ContractOutput, error)

	// CASHBACK
	AddCashback(
		address string,
        owner string,
        tokenAddress string,
        programType string,
        percentage string, // basis points, e.g. "250" = 2.50%
        startAt time.Time,
        expiredAt time.Time,
		paused bool,
    ) (types.ContractOutput, error)

    UpdateCashback(
        address string,
        tokenAddress string,
        programType string,
        percentage string,
        startAt time.Time,
        expiredAt time.Time,
    ) (types.ContractOutput, error)

    DepositCashbackFunds(
        address string,
		tokenAddress string,
        amount string,
    ) (types.ContractOutput, error)

    WithdrawCashbackFunds(
        address string,
        tokenAddress string,
        amount string,
    ) (types.ContractOutput, error)

    PauseCashback(address string, paused bool) (types.ContractOutput, error)
    UnpauseCashback(address string, paused bool) (types.ContractOutput, error)
	ClaimCashback(address, amount string) (types.ContractOutput, error)
	// getters
	GetCashback(address string) (types.ContractOutput, error)
	//TODO fix to ListCashbacks
	ListCashbacks(owner string, tokenAddress string, programType string, paused bool, page int, limit int, ascending bool) (types.ContractOutput, error)


	AddCoupon(
		address string, // optional, depends on your infra
		tokenAddress string,
		programType string,   // "percentage" | "fixed-amount"
		percentageBPS string, // required if percentage
		fixedAmount string,   // required if fixed-amount
		minOrder string,      // optional, "" means none
		startAt time.Time,
		expiredAt time.Time,
		paused bool,
		stackable bool,
		maxRedemptions int,
		perUserLimit int,
		passcodeHash string, // sha256(preimage)
	) (types.ContractOutput, error)

	UpdateCoupon(
		address string,
		tokenAddress string,
		programType string,
		percentageBPS string,
		fixedAmount string,
		minOrder string,
		startAt time.Time,
		expiredAt time.Time,
		stackable bool,
		maxRedemptions int,
		perUserLimit int,
		passcodeHash string, // optional; pass "" to keep
	) (types.ContractOutput, error)

	PauseCoupon(address string, paused bool) (types.ContractOutput, error)
	UnpauseCoupon(address string, paused bool) (types.ContractOutput, error)
	
	// Redeem coupon, 
	//TODO change this to Redeem Manual, because is possible to add orderAmount
	RedeemCoupon(
		address string,     // coupon address
		orderAmount string, // integer string
		passcode string,
	) (types.ContractOutput, error)

	// getters
	GetCoupon(address string) (types.ContractOutput, error)
	ListCoupons(owner, tokenAddress, programType string, paused *bool, page, limit int, ascending bool) (types.ContractOutput, error)

	CreatePayment(
		address string,
		tokenAddress string, // ERC-20-like token on your chain
		orderId string,
		payer string,
		payee string,
		amount string, // integer string
		expiredAt time.Time,
	) (types.ContractOutput, error)

	DirectPay(
		address string,
		tokenAddress string,
		orderId string,
		payer string,
		payee string,
		amount string,
	) (types.ContractOutput, error)

	AuthorizePayment(address string)(types.ContractOutput, error)

	CapturePayment(
		address string) (types.ContractOutput, error)

	VoidPayment(
		address string) (types.ContractOutput, error)

	RefundPayment(
		address string,
		amount string) (types.ContractOutput, error)

	UnpausePayment(address string, paused bool) (types.ContractOutput, error)
	PausePayment(address string, paused bool) (types.ContractOutput, error)

	GetPayment(address string) (types.ContractOutput, error)
	ListPayments(payer, payee, orderId, tokenAddress string, status []string, page, limit int, ascending bool) (types.ContractOutput, error)
	//MEMBER GET MEMBER
	AddMgM(
		address string,
		owner string,
		tokenAddress string,
		faucetAddress string,
		amount string,
		startAt time.Time,
		expireAt time.Time,
		paused bool,
	) (types.ContractOutput, error)
	UpdateMgM(
		mgmAddress string,
		amount string,
		startAt time.Time,
		expireAt time.Time,
	) (types.ContractOutput, error)
	PauseMgM(mgmAddress string, pause bool) (types.ContractOutput, error)
	UnpauseMgM(mgmAddress string, pause bool) (types.ContractOutput, error)
	DepositMgM(
		mgmAddress string,
		amount string,
	) (types.ContractOutput, error)
	WithdrawMgM(
		mgmAddress string,
		amount string,
	) (types.ContractOutput, error)

	AddInviterMember(mgmAddress, inviterAddress, password string) (types.ContractOutput, error)
	UpdateInviterPassword(mgmAddress, inviterAddress, newPassword string) (types.ContractOutput, error)
	DeleteInviterMember(mgmAddress, inviterAddress string) (types.ContractOutput, error)
	ClaimReward(mgmAddress, invitedAddress, password string) (types.ContractOutput, error)

	GetMgM(mgmAddress string) (types.ContractOutput, error)
	GetInviterMember(mgmAddress string, inviterAddress string) (types.ContractOutput, error)
	GetClaimInviter(mgmAddress string, inviterAddress string) (types.ContractOutput, error)
	GetClaimInvited(mgmAddress string, invitedAddress string) (types.ContractOutput, error)


	AddReview(address, reviewer, reviewee, subjectType, subjectID string, rating int, comment string,
		tags map[string]string, mediaHashes []string, startAt, expiredAt time.Time, hidden bool,
	) (types.ContractOutput, error)

	UpdateReview(address, subjectType, subjectID string, rating int, comment string,
		tags map[string]string, mediaHashes []string, startAt, expiredAt *time.Time,
	) (types.ContractOutput, error)

	HideReview(address string, hidden bool) (types.ContractOutput, error)
	
	VoteHelpful(address, voter string, isHelpful bool) (types.ContractOutput, error)
	ReportReview(address, reporter, reason string) (types.ContractOutput, error)
	ModerateReview(address, action, note string) (types.ContractOutput, error)

	GetReview(address string) (types.ContractOutput, error)
	ListReviews(owner, reviewer, reviewee, subjectType, subjectID string, includeHidden *bool, minRating, maxRating, page, limit int, asc bool) (types.ContractOutput, error)

	AddRaffle(address, owner, tokenAddress, ticketPrice string, maxEntries, maxEntriesPerUser int, startAt, expiredAt time.Time, paused bool, seedCommitHex string, metadata map[string]string) (types.ContractOutput, error)
	UpdateRaffle(address, tokenAddress, ticketPrice string, maxEntries, maxEntriesPerUser int, startAt, expiredAt *time.Time, seedCommitHex string, metadata map[string]string) (types.ContractOutput, error)
	PauseRaffle(address string, paused bool) (types.ContractOutput, error)
	UnpauseRaffle(address string, paused bool) (types.ContractOutput, error)
	EnterRaffle(address string, tickets int, payTokenAddress string) (types.ContractOutput, error)
	DrawRaffle(address, revealSeed string) (types.ContractOutput, error)
	ClaimRaffle(address, winner string) (types.ContractOutput, error)
	WithdrawRaffle(address, tokenAddress, amount string) (types.ContractOutput, error)
	AddRafflePrize(raffleAddress string, tokenAddress string, amount string) (types.ContractOutput, error)
	RemoveRafflePrize(raffleAddress string, uuid string) (types.ContractOutput, error)

	// GetRaffle(address string) (types.ContractOutput, error)
	// ListRaffles(owner, tokenAddress string, paused *bool, activeOnly *bool, page, limit int, asc bool) (types.ContractOutput, error)

	ListPrizes(raffleAddress string, page, limit int, asc bool) (types.ContractOutput, error)


}

type networkClient struct {
	mqttClient mqtt.MQTT
	privateKey string
	publicKey  string
	replyTo    string
}

// New creates a new client
func New(broker, clientID string, debug bool) Client2FinanceNetwork {

	mqttClient := mqtt.New(broker, clientID, debug)
	mqttClient.Connect()
	replyTo := uuid.NewString()
	return &networkClient{
		mqttClient: mqttClient,
		replyTo:    replyTo,
	}
}

func (c *networkClient) SetPrivateKey(privateKey string) {
	c.privateKey = privateKey
	pubKey, err := keys.PublicKeyFromEd25519PrivateHex(privateKey)
	if err != nil {
		log.Fatalf("Error getting public key from private key: %v", err)
	}
	hex := keys.PublicKeyToHex(pubKey)
	if err != nil {
		log.Fatalf("Error converting public key to hex: %v", err)
	}
	c.publicKey = hex
}

func (c *networkClient) GetPrivateKey() string {
	return c.privateKey
}

func (c *networkClient) GetPublicKey() string {
	return c.publicKey
}

// SendRequest publishes the payload to the MQTT broker
func (c *networkClient) sendRequest(topic string, payload interface{}) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal payload error: %w", err)
	}

	if err := c.mqttClient.Publish(topic, data); err != nil {
		return fmt.Errorf("publish error: %w", err)
	}

	return nil
}

func (c *networkClient) GenerateKeyEd25519() (string, string, error) {
	publicKey, privateKey, err := keys.GenerateKeyEd25519()
	if err != nil {
		return "", "", fmt.Errorf("failed to generate public key: %w", err)
	}

	return publicKey, privateKey, nil
}

func (c *networkClient) GetNonce(publicKey string) (uint64, error) {
	if publicKey == "" {
		return 0, fmt.Errorf("public key not set")
	}
	if err := keys.ValidateEDDSAPublicKey(publicKey); err != nil {
		return 0, fmt.Errorf("invalid public key: %w", err)
	}

	transactionInput := transaction.TransactionInput{
		From: publicKey,
	}
	nonceBytes, err := c.SendTransaction(handler.REQUEST_METHOD_GET_NONCE, transactionInput, c.replyTo)
	if err != nil {
		return 0, fmt.Errorf("failed to get nonce: %w", err)
	}

	var nonce uint64
	if err := json.Unmarshal(nonceBytes, &nonce); err != nil {
		return 0, fmt.Errorf("failed to unmarshal nonce: %w", err)
	}

	return nonce, nil
}

func (c *networkClient) ListTransactions(from, to, hash string, dataFilter map[string]interface{}, nonce uint64,
	page, limit int,
	ascending bool) ([]transaction.Transaction, error) {

	if from == "" && to == "" && hash == "" {
		return nil, fmt.Errorf("at least one of from, to or hash must be set")
	}

	if from != "" {
		if err := keys.ValidateEDDSAPublicKey(from); err != nil {
			return nil, fmt.Errorf("invalid from address: %w", err)
		}
	}

	if to != "" {
		if err := keys.ValidateEDDSAPublicKey(to); err != nil {
			return nil, fmt.Errorf("invalid to address: %w", err)
		}
	}

	transactionInput := transaction.TransactionInput{
		From:      from,
		To:        to,
		Hash:      hash,
		Data:      dataFilter,
		Nonce:     nonce,
		Page:      page,
		Limit:     limit,
		Ascending: ascending,
	}
	transactionBytes, err := c.SendTransaction(handler.REQUEST_METHOD_GET_TRANSACTIONS, transactionInput, c.replyTo)
	if err != nil {
		return nil, fmt.Errorf("failed to send transaction - List Transactions: %w", err)
	}

	var transactions []transaction.Transaction
	if err := json.Unmarshal(transactionBytes, &transactions); err != nil {
		return nil, fmt.Errorf("failed to unmarshal transactions: %w", err)
	}

	return transactions, nil
}

func (c *networkClient) ListLogs(logType []string, logIndex uint, transactionHash string, event map[string]interface{}, contractAddress string,
	page, limit int,
	ascending bool) ([]blockchainLog.Log, error) {
	if len(logType) == 0 && transactionHash == "" && contractAddress == "" {
		return nil, fmt.Errorf("at least one of logType, transactionHash or contractAddress must be set")
	}

	logInput := blockchainLog.LogParams{
		LogType:         logType,
		LogIndex:        logIndex,
		TransactionHash: transactionHash,
		Event:           event,
		ContractAddress: contractAddress,
		Page:            page,
		Limit:           limit,
		Ascending:       ascending,
	}

	logsBytes, err := c.SendTransaction(handler.REQUEST_METHOD_GET_LOGS, logInput, c.replyTo)
	if err != nil {
		return nil, fmt.Errorf("failed to send transaction: - List Logs %w", err)
	}

	var logs []blockchainLog.Log
	if err := json.Unmarshal(logsBytes, &logs); err != nil {
		return nil, fmt.Errorf("failed to unmarshal logs: %w", err)
	}

	return logs, nil
}

func (c *networkClient) sendAndWaitResponse(method string, params interface{}, replyTo string) ([]byte, error) {
	replyTopic := fmt.Sprintf("%s/%s", event.TRANSACTIONS_RESPONSE_TOPIC, replyTo)
	responseChan := make(chan []byte, 1)
	if err := c.receiveResponse(replyTopic, func(data []byte) {
		responseChan <- data
	}); err != nil {
		return nil, fmt.Errorf("failed to subscribe to reply topic: %w", err)
	}

	payload := event.RequestPayload{
		Method: method,
		Params: params,
	}
	// Use the original topic and append the replyTo
	orig := event.TRANSACTIONS_REQUEST_TOPIC
	base := strings.TrimSuffix(orig, "/+")
	newTopic := fmt.Sprintf("%s/%s", base, replyTo)
	if err := c.sendRequest(newTopic, payload); err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	select {
	case resp := <-responseChan:
		return resp, nil
	case <-time.After(10 * time.Second):
		return nil, fmt.Errorf("timeout waiting for response on topic %s", replyTo)
	}
}

// ReceiveResponse subscribes to a topic and calls the handler with raw payload
func (c *networkClient) receiveResponse(topic string, handler func([]byte)) error {
	return c.mqttClient.SubscribeWithHandler(topic, func(_ mqtt.Client, msg mqtt.Message) {
		handler(msg.Payload())
	})
}

func (c *networkClient) SendTransaction(method string, tx interface{}, replyTo string) (outputBytes []byte, err error) {

	// Send the transaction to the network
	bytes, err := c.sendAndWaitResponse(method, tx, replyTo)
	if err != nil {
		return nil, fmt.Errorf("failed to send transaction: - Handler Request %w", err)
	}

	// Decode the response envelope
	var resp event.ResponsePayload
	if err := json.Unmarshal(bytes, &resp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	// Re-encode the inner Data to raw JSON bytes
	outputBytes, err = json.Marshal(resp.Data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal response data: %w", err)
	}

	if resp.Status == event.RESPONSE_STATUS_ERROR {
		return nil, fmt.Errorf("error in response: %s", resp.Message)
	}

	return outputBytes, nil
}

// SendTransaction builds, signs, and sends a transaction to the blockchain.
func (c *networkClient) SignAndSendTransaction(
	from string,
	to string,
	method string,
	data map[string]interface{},
) (types.ContractOutput, error) {
	// Validate public key (from address)
	if err := keys.ValidateEDDSAPublicKey(from); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid from address: %w", err)
	}

	// Get current nonce and increment
	nonce, err := c.GetNonce(from)
	if err != nil {
		if strings.Contains(err.Error(), "record not found") {
			// If nonce not found, start from 0
			nonce = 0
		} else {
			// If any other error, return it
			return types.ContractOutput{}, fmt.Errorf("failed to get nonce: %w", err)
		}
	}

	nonce++

	txSigned, err := c.SignTransaction(from, to, method, data, nonce)
	if err != nil {
		return types.ContractOutput{}, fmt.Errorf("failed to sign transaction: %w", err)
	}
	contractOutputBytes, err := c.SendTransaction(handler.REQUEST_METHOD_SEND, txSigned, c.replyTo)
	if err != nil {
		return types.ContractOutput{}, fmt.Errorf("failed to send transaction: %w", err)
	}

	var contractOutput types.ContractOutput
	if err := json.Unmarshal(contractOutputBytes, &contractOutput); err != nil {
		return types.ContractOutput{}, fmt.Errorf("failed to unmarshal contract output: %w", err)
	}

	return contractOutput, nil
}

func (c *networkClient) GetState(
	to string,
	method string,
	data map[string]interface{},
) (types.ContractOutput, error) {
	// Convert data map to JSONB
	jsonData, err := utils.MapToJSONB(data)
	if err != nil {
		return types.ContractOutput{}, fmt.Errorf("failed to marshal data to JSONB: %w", err)
	}

	// Build a transaction input without signature and hash for query
	txInput := transaction.TransactionInput{
		To:              to,
		Method:          method,
		Data:            jsonData,
	}

	// Use a unique reply topic
	contractOutputBytes, err := c.SendTransaction(handler.REQUEST_METHOD_GET_STATE, txInput, c.replyTo)
	if err != nil {
		return types.ContractOutput{}, err
	}

	var contractOutput types.ContractOutput
	if err := json.Unmarshal(contractOutputBytes, &contractOutput); err != nil {
		return types.ContractOutput{}, fmt.Errorf("failed to unmarshal contract output: %w", err)
	}
	return contractOutput, nil
}

func (c *networkClient) ListBlocks(blockNumber uint64, blockTimestamp time.Time, hash string, previousHash string,
	merkleRoot string,
	page, limit int,
	ascending bool) ([]block.Block, error) {

	blockParams := block.BlockParams{
		BlockNumber:    blockNumber,
		BlockTimestamp: blockTimestamp,
		Hash:           hash,
		PreviousHash:   previousHash,
		MerkleRoot:     merkleRoot,
		Page:           page,
		Limit:          limit,
		Ascending:      ascending,
	}

	blockBytes, err := c.SendTransaction(handler.REQUEST_METHOD_GET_BLOCKS, blockParams, c.replyTo)
	if err != nil {
		return nil, fmt.Errorf("failed to send transaction - List Blocks %w", err)
	}

	var blocks []block.Block
	if err := json.Unmarshal(blockBytes, &blocks); err != nil {
		return nil, fmt.Errorf("failed to unmarshal blocks: %w", err)
	}

	return blocks, nil
}


func (c *networkClient) SignTransaction(from, to, method string, data utils.JSONB, nonce uint64) (*transaction.Transaction, error) {
	// 1. create new tx
	newTx := transaction.NewTransaction(from, to, method, data, nonce)

	// 2. get serialized form (here it's just the object)
	tx := newTx.Get()

	// 3. sign
	signedTx, err := transaction.SignTransactionHexKey(c.privateKey, tx)
	if err != nil {
		return nil, fmt.Errorf("failed to sign transaction: %w", err)
	}
	return signedTx, nil
}

func (c *networkClient) DeployContract1(contractVersion string) (types.ContractOutput, error) {
	if c.publicKey == "" {
		return types.ContractOutput{}, fmt.Errorf("from address is required")
	}
	from := c.publicKey
	
	if err := keys.ValidateEDDSAPublicKey(from); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid from address: %w", err)
	}

	if contractVersion == "" {
		return types.ContractOutput{}, fmt.Errorf("contract version is required")
	}

	to := types.DEPLOY_CONTRACT_ADDRESS
	
	method := contractV1.METHOD_DEPLOY_CONTRACT
	data := map[string]interface{}{
		"contract_version": contractVersion,
	}
	contractOutput, err := c.SignAndSendTransaction(from, to, method, data)
	if err != nil {
		return types.ContractOutput{}, fmt.Errorf("failed to deploy contract: %w", err)
	}
	return contractOutput, nil
}

func (c *networkClient) DeployContract2(contractVersion, contractAddress string) (types.ContractOutput, error) {
	if c.publicKey == "" {
		return types.ContractOutput{}, fmt.Errorf("from address is required")
	}
	from := c.publicKey
	
	if err := keys.ValidateEDDSAPublicKey(from); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid from address: %w", err)
	}

	if contractVersion == "" {
		return types.ContractOutput{}, fmt.Errorf("contract version is required")
	}
	to := types.DEPLOY_CONTRACT_ADDRESS
	if contractAddress != "" {
		to = contractAddress
	}
	method := contractV1.METHOD_DEPLOY_CONTRACT2
	data := map[string]interface{}{
		"contract_version": contractVersion,
	}
	contractOutput, err := c.SignAndSendTransaction(from, to, method, data)
	if err != nil {
		return types.ContractOutput{}, fmt.Errorf("failed to deploy contract: %w", err)
	}
	return contractOutput, nil
}