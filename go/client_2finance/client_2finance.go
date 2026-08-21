package client_2finance

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"gitlab.com/2finance/2finance-network/blockchain/block"
	"gitlab.com/2finance/2finance-network/blockchain/contract/contractV2"
	inputsDropV2 "gitlab.com/2finance/2finance-network/blockchain/contract/dropV2/inputs"
	inputsFXLifecycleV2 "gitlab.com/2finance/2finance-network/blockchain/contract/fxLifecycleV2/inputs"
	"gitlab.com/2finance/2finance-network/blockchain/contract/lifecycleCommonV2"
	inputsPaymentV2 "gitlab.com/2finance/2finance-network/blockchain/contract/paymentV2/inputs"
	"gitlab.com/2finance/2finance-network/blockchain/contractversionv2"
	"gitlab.com/2finance/2finance-network/blockchain/encryption/keys"
	blockchainLog "gitlab.com/2finance/2finance-network/blockchain/log"
	"gitlab.com/2finance/2finance-network/blockchain/transaction"
	"gitlab.com/2finance/2finance-network/blockchain/types"
	"gitlab.com/2finance/2finance-network/blockchain/utils"
	"gitlab.com/2finance/2finance-network/blockchain/virtualmachine"

	sdknetwork "github.com/2Finance-Labs/2finance-sdk-client/network"
	"github.com/2Finance-Labs/2finance-sdk-client/protocol"
	"github.com/2Finance-Labs/2finance-sdk-client/wallet_manager"
	"github.com/google/uuid"
	"gitlab.com/2finance/2finance-network/infra/event"
)

// Interface exposes the client behavior
type Client2FinanceNetwork interface {
	// Client
	SetChainID(chainId uint8)
	SetWalletManager(wallet wallet_manager.IWalletManager)

	SendTransaction(method string, tx interface{}, replyTo string) (outputBytes []byte, err error)
	SignPreparedTransaction(tx protocol.PreparedTransaction) (protocol.SignedTransaction, error)
	SubmitSignedTransaction(tx protocol.SignedTransaction) (types.ContractOutput, error)
	SignAndSendPreparedTransaction(tx protocol.PreparedTransaction) (types.ContractOutput, error)

	// CHAIN
	ListTransactions(from, to, hash string, dataFilter map[string]interface{}, version uint8,
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
	SignAndSendTransaction(
		chainId uint8,
		from string,
		to string,
		method string,
		data map[string]interface{},
		version uint8,
		uuid7 string,
	) (types.ContractOutput, error)
	GetState(
		to string,
		method string,
		data map[string]interface{}) (types.ContractOutput, error)
	ListBlocks(blockNumber uint64, blockTimestamp time.Time, hash string, previousHash string, transactionMerkleRoot string,
		page, limit int,
		ascending bool) ([]block.Block, error)

	// WALLET
	AddWallet(address, pubKey string) (types.ContractOutput, error)
	GetWalletByPublicKey(pubKey string) (types.ContractOutput, error)
	GetWalletByAddress(address string) (types.ContractOutput, error)

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
		allowedUsers map[string]bool,
		blockedUsers map[string]bool,
		frozenAccounts map[string]bool,
		feeTiersList []map[string]interface{},
		feeAddress string,
		freezeAuthorityRevoked bool,
		mintAuthorityRevoked bool,
		updateAuthorityRevoked bool,
		paused bool,
		expired_at time.Time,
		assetGLBUri string,
		tokenType string,
		transferable bool, assetType string) (types.ContractOutput, error)
	MintToken(to, mintTo, amount string) (types.ContractOutput, error)
	BurnToken(to, amount string, tokenUUIDList []string) (types.ContractOutput, error)
	TransferToken(tokenAddress, transferTo, amount string, tokenUUIDList []string) (types.ContractOutput, error)
	AddAllowedUsers(tokenAddress string, allowedUsers map[string]bool) (types.ContractOutput, error)
	RemoveAllowedUsers(tokenAddress string, allowedUsers map[string]bool) (types.ContractOutput, error)
	AddBlockedUsers(tokenAddress string, blockedUsers map[string]bool) (types.ContractOutput, error)
	RemoveBlockedUsers(tokenAddress string, blockedUsers map[string]bool) (types.ContractOutput, error)
	RevokeFreezeAuthority(tokenAddress string, revoke bool) (types.ContractOutput, error)
	RevokeMintAuthority(tokenAddress string, revoke bool) (types.ContractOutput, error)
	RevokeUpdateAuthority(tokenAddress string, revoke bool) (types.ContractOutput, error)
	UpdateMetadata(tokenAddress, symbol, name string, decimals int, description, image, website string,
		tagsSocialMedia, tagsCategory, tags map[string]string,
		creator, creatorWebsite string, expired_at time.Time) (types.ContractOutput, error)

	FreezeWallet(tokenAddress string, wallet string) (types.ContractOutput, error)
	UnfreezeWallet(tokenAddress string, wallet string) (types.ContractOutput, error)

	PauseToken(tokenAddress string, pause bool) (types.ContractOutput, error)
	UnpauseToken(tokenAddress string, unpause bool) (types.ContractOutput, error)

	UpdateFeeTiers(tokenAddress string, feeTierList []map[string]interface{}) (types.ContractOutput, error)
	UpdateFeeAddress(tokenAddress, feeAddress string) (types.ContractOutput, error)
	UpdateGlbFile(tokenAddress string, newAssetGLBUri string) (types.ContractOutput, error)
	TransferableToken(tokenAddress string, transferable bool) (types.ContractOutput, error)
	UntransferableToken(tokenAddress string, transferable bool) (types.ContractOutput, error)

	GetToken(tokenAddress string, symbol string, name string) (types.ContractOutput, error)
	ListTokens(ownerAddress, symbol, name, tokenType string, page, limit int, ascending bool) (types.ContractOutput, error)

	GetTokenBalance(tokenAddress, ownerAddress string) (types.ContractOutput, error)
	GetTokenBalanceNFT(tokenAddress, ownerAddress, tokenUUID string) (types.ContractOutput, error)
	ListTokenBalances(tokenAddress, ownerAddress, tokenType string, page, limit int, ascending bool) (types.ContractOutput, error)

	NewDrop(
		in inputsDropV2.InputNewDrop,
	) (types.ContractOutput, error)

	UpdateDropMetadata(
		in inputsDropV2.InputUpdateDropMetadata,
	) (types.ContractOutput, error)

	AllowOracles(
		address string,
		oracles map[string]bool,
	) (types.ContractOutput, error)

	DisallowOracles(
		address string,
		oracles map[string]bool,
	) (types.ContractOutput, error)

	DepositDrop(
		address string,
		programAddress string,
		tokenAddress string,
		amount string,
		uuid []string,
	) (types.ContractOutput, error)

	ClaimDrop(
		address string,
	) (types.ContractOutput, error)

	LastClaimed(
		address string,
		wallet string,
	) (types.ContractOutput, error)

	WithdrawDrop(
		address string,
		programAddress string,
		tokenAddress string,
		amount string,
		uuid []string,
	) (types.ContractOutput, error)

	PauseDrop(
		address string,
	) (types.ContractOutput, error)

	UnpauseDrop(
		address string,
	) (types.ContractOutput, error)

	AttestParticipantEligibility(
		address string,
		wallet string,
		approved bool,
	) (types.ContractOutput, error)

	GetDrop(address string) (types.ContractOutput, error)
	ListDrops(
		owner string,
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
		tokenType string,
		uuid string,
	) (types.ContractOutput, error)

	WithdrawCashbackFunds(
		address string,
		tokenAddress string,
		amount string,
		tokenType string,
		uuid string,
	) (types.ContractOutput, error)

	PauseCashback(address string, paused bool) (types.ContractOutput, error)
	UnpauseCashback(address string, paused bool) (types.ContractOutput, error)
	ClaimCashback(address, amount, tokenType, uuid string) (types.ContractOutput, error)
	// getters
	GetCashback(address string) (types.ContractOutput, error)
	//TODO fix to ListCashbacks
	ListCashbacks(owner string, tokenAddress string, programType string, paused bool, page int, limit int, ascending bool) (types.ContractOutput, error)

	AddCoupon(
		address string, // optional, depends on your infra
		discountType string, // "percentage" | "fixed-amount"
		percentageBPS string, // required if percentage
		fixedAmount string, // required if fixed-amount
		minOrder string, // optional, "" means none
		startAt time.Time,
		expiredAt time.Time,
		paused bool,
		stackable bool,
		maxRedemptions int,
		perUserLimit int,
		passcodeHash string, // sha256(preimage)
		voucherOwner string,
		symbol string,
		name string,
		amount string,
		description string,
		image string,
		website string,
		tagsSocialMedia map[string]string,
		tagsCategory map[string]string,
		tags map[string]string,
		creator string,
		creatorWebsite string,
		assetGLBUri string,
	) (types.ContractOutput, error)

	UpdateCoupon(
		address string,
		tokenAddress string,
		discountType string,
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

	IssueVoucher(
		address string, // coupon address
		toAddress string,
		amount string, // integer string
	) (types.ContractOutput, error)

	RedeemVoucher(
		address string, // coupon address
		orderAmount string, // integer string
		passcode string,
		voucherUUID string,
	) (types.ContractOutput, error)

	// getters
	GetCoupon(address string) (types.ContractOutput, error)
	ListCoupons(owner, tokenAddress, discountType string, paused *bool, page, limit int, ascending bool) (types.ContractOutput, error)

	// Payment
	CreatePayment(in inputsPaymentV2.InputCreate) (types.ContractOutput, error)
	DirectPay(in inputsPaymentV2.InputDirectPay) (types.ContractOutput, error)

	AuthorizePayment(in inputsPaymentV2.InputAuthorize) (types.ContractOutput, error)
	CapturePayment(in inputsPaymentV2.InputCapture) (types.ContractOutput, error)
	VoidPayment(in inputsPaymentV2.InputVoidPayment) (types.ContractOutput, error)
	RefundPayment(in inputsPaymentV2.InputRefund) (types.ContractOutput, error)

	UnpausePayment(in inputsPaymentV2.InputPause) (types.ContractOutput, error)
	PausePayment(in inputsPaymentV2.InputPause) (types.ContractOutput, error)

	GetPayment(address string) (types.ContractOutput, error)
	ListPayments(in inputsPaymentV2.InputList) (types.ContractOutput, error)
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
		tokenType string,
		uuid string,
	) (types.ContractOutput, error)
	WithdrawMgM(
		mgmAddress string,
		amount string,
		tokenType string,
		uuid string,
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
	ListReviews(
		reviewer, reviewee, subjectType, subjectID string,
		includeHidden *bool,
		minRating, maxRating, page, limit int,
		asc bool,
	) (types.ContractOutput, error)

	AddRaffle(address, owner, tokenAddress, ticketPrice string, maxEntries, maxEntriesPerUser int, startAt, expiredAt time.Time, paused bool, seedCommitHex string, metadata map[string]string) (types.ContractOutput, error)
	UpdateRaffle(address, tokenAddress, ticketPrice string, maxEntries, maxEntriesPerUser int, startAt, expiredAt *time.Time, seedCommitHex string, metadata map[string]string) (types.ContractOutput, error)
	PauseRaffle(address string, paused bool) (types.ContractOutput, error)
	UnpauseRaffle(address string, paused bool) (types.ContractOutput, error)
	EnterRaffle(address string, tickets int, payTokenAddress, tokenType, uuid string) (types.ContractOutput, error)
	DrawRaffle(address, revealSeed string) (types.ContractOutput, error)
	ClaimRaffle(address, prizeUUID string) (types.ContractOutput, error)
	WithdrawRaffle(address, tokenAddress, amount, tokenType, uuid string) (types.ContractOutput, error)
	AddRafflePrize(raffleAddress string, tokenAddress string, amount string, uuidNFTs []string) (types.ContractOutput, error)
	RemoveRafflePrize(raffleAddress string, uuid string) (types.ContractOutput, error)

	GetRaffle(address string) (types.ContractOutput, error)
	ListRaffles(owner, tokenAddress string, paused *bool, activeOnly *bool, page, limit int, asc bool) (types.ContractOutput, error)
	GetPrize(address string, prizeUUID string) (types.ContractOutput, error)
	ListPrizes(raffleAddress string, page, limit int, asc bool) (types.ContractOutput, error)

	// LIFECYCLES
	StartFX(in inputsFXLifecycleV2.InputStartFX) (types.ContractOutput, error)
	AdvanceFX(in inputsFXLifecycleV2.InputAdvanceFX) (types.ContractOutput, error)
	FailFX(in inputsFXLifecycleV2.InputFailFX) (types.ContractOutput, error)
	GetFX(address, requestID string) (types.ContractOutput, error)

	StartOnboarding(in lifecycleCommonV2.StartInput) (types.ContractOutput, error)
	AdvanceOnboarding(in lifecycleCommonV2.AdvanceInput) (types.ContractOutput, error)
	FailOnboarding(in lifecycleCommonV2.FailInput) (types.ContractOutput, error)
	GetOnboarding(address, requestID string) (types.ContractOutput, error)

	StartReceiving(in lifecycleCommonV2.StartInput) (types.ContractOutput, error)
	AdvanceReceiving(in lifecycleCommonV2.AdvanceInput) (types.ContractOutput, error)
	FailReceiving(in lifecycleCommonV2.FailInput) (types.ContractOutput, error)
	GetReceiving(address, requestID string) (types.ContractOutput, error)

	StartSending(in lifecycleCommonV2.StartInput) (types.ContractOutput, error)
	AdvanceSending(in lifecycleCommonV2.AdvanceInput) (types.ContractOutput, error)
	FailSending(in lifecycleCommonV2.FailInput) (types.ContractOutput, error)
	GetSending(address, requestID string) (types.ContractOutput, error)

	StartMultiCurrency(in lifecycleCommonV2.StartInput) (types.ContractOutput, error)
	AdvanceMultiCurrency(in lifecycleCommonV2.AdvanceInput) (types.ContractOutput, error)
	FailMultiCurrency(in lifecycleCommonV2.FailInput) (types.ContractOutput, error)
	GetMultiCurrency(address, requestID string) (types.ContractOutput, error)
}

type NetworkClient struct {
	protocolV2    *sdknetwork.Client
	replyTo       string
	chainId       uint8
	walletManager wallet_manager.IWalletManager
}

// NewV2 creates the contract-oriented client on top of the commit-first HTTP
// API. Contract reads continue through the HTTP protocol-V2 query endpoint;
// every signed write is submitted through /v2/2finance-network/transactions.
func NewV2(
	baseURL string,
	httpClient *http.Client,
	walletManager wallet_manager.IWalletManager,
) Client2FinanceNetwork {
	return &NetworkClient{
		protocolV2:    sdknetwork.New(baseURL, httpClient),
		replyTo:       uuid.NewString(),
		walletManager: walletManager,
	}
}

func (c *NetworkClient) SetChainID(chainId uint8) {
	if chainId < 1 || chainId > 2 {
		log.Fatalf("invalid chainId: %d, available values are 2 testnet or 1 mainnet", chainId)
	}
	c.chainId = chainId
}

func (c *NetworkClient) SetWalletManager(wallet wallet_manager.IWalletManager) {
	if wallet == nil {
		log.Fatalf("wallet manager cannot be nil")
	}
	c.walletManager = wallet
}

func (c *NetworkClient) GetChainID() uint8 {
	return c.chainId
}

func (c *NetworkClient) ListTransactions(from, to, hash string, dataFilter map[string]interface{}, version uint8,
	page, limit int,
	ascending bool) ([]transaction.Transaction, error) {

	if from == "" && to == "" && hash == "" {
		return nil, fmt.Errorf("at least one of from, to or hash must be set")
	}

	if from != "" {
		if err := keys.ValidateEDDSAPublicKeyHex(from); err != nil {
			return nil, fmt.Errorf("invalid from address: %w", err)
		}
	}

	if to != "" {
		if err := keys.ValidateEDDSAPublicKeyHex(to); err != nil {
			return nil, fmt.Errorf("invalid to address: %w", err)
		}
	}

	dataFilterRawMessage, err := utils.MapToRawMessage(dataFilter)
	if err != nil {
		return nil, fmt.Errorf("failed to convert dataFilter to RawMessage: %w", err)
	}

	transactionInput := transaction.TransactionInput{
		From:      from,
		To:        to,
		Hash:      hash,
		Data:      dataFilterRawMessage,
		Version:   version,
		Page:      page,
		Limit:     limit,
		Ascending: ascending,
	}
	transactionBytes, err := c.SendTransaction(virtualmachine.REQUEST_METHOD_GET_TRANSACTIONS, transactionInput, c.replyTo)
	if err != nil {
		return nil, fmt.Errorf("failed to send transaction - List Transactions: %w", err)
	}

	var transactions []transaction.Transaction
	if err := json.Unmarshal(transactionBytes, &transactions); err != nil {
		return nil, fmt.Errorf("failed to unmarshal transactions: %w", err)
	}

	return transactions, nil
}

func (c *NetworkClient) ListLogs(logType []string, logIndex uint, transactionHash string, event map[string]interface{}, contractAddress string,
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

	logsBytes, err := c.SendTransaction(virtualmachine.REQUEST_METHOD_GET_LOGS, logInput, c.replyTo)
	if err != nil {
		return nil, fmt.Errorf("failed to send transaction: - List Logs %w", err)
	}

	var logs []blockchainLog.Log
	if err := json.Unmarshal(logsBytes, &logs); err != nil {
		return nil, fmt.Errorf("failed to unmarshal logs: %w", err)
	}

	return logs, nil
}

func (c *NetworkClient) SendTransaction(method string, tx interface{}, replyTo string) (outputBytes []byte, err error) {
	if c.protocolV2 == nil {
		return nil, fmt.Errorf("protocol V2 is unavailable: configure the protocol V2 HTTP client")
	}
	return c.sendHTTPV2(method, tx)
}

func (c *NetworkClient) sendHTTPV2(method string, params interface{}) ([]byte, error) {
	if method == virtualmachine.REQUEST_METHOD_SEND {
		networkTransaction, err := networkTransactionFrom(params)
		if err != nil {
			return nil, err
		}
		signed, err := protocol.SignedTransactionFromNetwork(networkTransaction)
		if err != nil {
			return nil, err
		}
		result, err := c.protocolV2.SubmitSignedTransactionV2(context.Background(), signed)
		if err != nil {
			return nil, fmt.Errorf("submit commit-first transaction: %w", err)
		}
		return append([]byte(nil), result.ExecutionOutput...), nil
	}

	var response struct {
		Code int             `json:"code"`
		Msg  string          `json:"msg"`
		Data json.RawMessage `json:"data"`
	}
	normalizedParams, err := normalizePublicContractVersionsV2(params)
	if err != nil {
		return nil, fmt.Errorf("normalize protocol V2 query: %w", err)
	}
	err = c.protocolV2.Post(
		context.Background(),
		"/v2/2finance-network/query",
		event.RequestPayload{Method: method, Params: normalizedParams},
		&response,
	)
	if err != nil {
		return nil, fmt.Errorf("query finance network over HTTP: %w", err)
	}
	if response.Code != http.StatusOK {
		return nil, fmt.Errorf("finance network query returned code %d: %s", response.Code, response.Msg)
	}
	return append([]byte(nil), response.Data...), nil
}

func networkTransactionFrom(value interface{}) (*transaction.Transaction, error) {
	switch typed := value.(type) {
	case *transaction.Transaction:
		if typed == nil {
			return nil, fmt.Errorf("signed transaction is nil")
		}
		return typed, nil
	case transaction.Transaction:
		copy := typed
		return &copy, nil
	default:
		encoded, err := json.Marshal(value)
		if err != nil {
			return nil, fmt.Errorf("encode signed transaction: %w", err)
		}
		var decoded transaction.Transaction
		if err := json.Unmarshal(encoded, &decoded); err != nil {
			return nil, fmt.Errorf("decode signed transaction: %w", err)
		}
		return &decoded, nil
	}
}

func (c *NetworkClient) SignAndSendTransaction(
	chainId uint8,
	from string,
	to string,
	method string,
	data map[string]interface{},
	version uint8,
	uuid7 string,
) (types.ContractOutput, error) {
	if err := keys.ValidateEDDSAPublicKeyHex(from); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid from address: %w", err)
	}

	if c.walletManager == nil {
		return types.ContractOutput{}, fmt.Errorf("wallet manager is required")
	}

	if !c.walletManager.IsUnlocked() {
		return types.ContractOutput{}, fmt.Errorf("wallet is locked")
	}

	if c.protocolV2 == nil {
		return types.ContractOutput{}, fmt.Errorf("protocol V2 is unavailable: configure the protocol V2 HTTP client")
	}
	normalizedValue, normalizeErr := normalizePublicContractVersionsV2(data)
	if normalizeErr != nil {
		return types.ContractOutput{}, normalizeErr
	}
	normalizedData, ok := normalizedValue.(map[string]interface{})
	if !ok {
		return types.ContractOutput{}, fmt.Errorf("protocol V2 transaction data must be an object")
	}
	prepared, prepareErr := protocol.PrepareNativeGoTransactionV2(
		protocol.NativeGoTransactionV2Input{
			ChainID: chainId,
			From:    from,
			To:      to,
			Method:  method,
			Payload: normalizedData,
			UUID7:   uuid7,
		},
	)
	if prepareErr != nil {
		return types.ContractOutput{}, fmt.Errorf("prepare native Go v2 transaction: %w", prepareErr)
	}
	signed, err := c.walletManager.SignPreparedTransaction(prepared)
	if err != nil {
		return types.ContractOutput{}, fmt.Errorf("failed to sign transaction: %w", err)
	}
	txSigned, err := signed.ToNetworkTransaction()
	if err != nil {
		return types.ContractOutput{}, fmt.Errorf("convert signed V2 transaction: %w", err)
	}

	contractOutputBytes, err := c.SendTransaction(
		virtualmachine.REQUEST_METHOD_SEND,
		txSigned,
		c.replyTo,
	)
	if err != nil {
		return types.ContractOutput{}, fmt.Errorf("failed to send transaction: %w", err)
	}

	var contractOutput types.ContractOutput
	if err := json.Unmarshal(contractOutputBytes, &contractOutput); err != nil {
		return types.ContractOutput{}, fmt.Errorf("failed to unmarshal contract output: %w", err)
	}

	return contractOutput, nil
}

func (c *NetworkClient) GetState(
	to string,
	method string,
	data map[string]interface{},
) (types.ContractOutput, error) {
	// Convert data map to JSONB
	jsonData, err := utils.MapToRawMessage(data)
	if err != nil {
		return types.ContractOutput{}, fmt.Errorf("failed to marshal data to JSONB: %w", err)
	}

	// Build a transaction input without signature and hash for query
	txInput := transaction.TransactionInput{
		To:     to,
		Method: method,
		Data:   jsonData,
	}

	// Use a unique reply topic
	contractOutputBytes, err := c.SendTransaction(virtualmachine.REQUEST_METHOD_GET_STATE, txInput, c.replyTo)
	if err != nil {
		return types.ContractOutput{}, err
	}

	var contractOutput types.ContractOutput
	if err := json.Unmarshal(contractOutputBytes, &contractOutput); err != nil {
		return types.ContractOutput{}, fmt.Errorf("failed to unmarshal contract output: %w", err)
	}
	return contractOutput, nil
}

func (c *NetworkClient) ListBlocks(blockNumber uint64, blockTimestamp time.Time, hash string, previousHash string,
	transactionMerkleRoot string,
	page, limit int,
	ascending bool) ([]block.Block, error) {

	blockParams := block.BlockParams{
		Number:                 blockNumber,
		Timestamp:              blockTimestamp,
		Hash:                   hash,
		PreviousHash:           previousHash,
		TransactionsMerkleRoot: transactionMerkleRoot,
		Page:                   page,
		Limit:                  limit,
		Ascending:              ascending,
	}

	blockBytes, err := c.SendTransaction(virtualmachine.REQUEST_METHOD_GET_BLOCKS, blockParams, c.replyTo)
	if err != nil {
		return nil, fmt.Errorf("failed to send transaction - List Blocks %w", err)
	}

	var blocks []block.Block
	if err := json.Unmarshal(blockBytes, &blocks); err != nil {
		return nil, fmt.Errorf("failed to unmarshal blocks: %w", err)
	}

	return blocks, nil
}

func (c *NetworkClient) DeployContract1(contractVersion string) (types.ContractOutput, error) {
	if c.walletManager == nil {
		return types.ContractOutput{}, fmt.Errorf("wallet manager is required")
	}

	from := c.walletManager.OwnerAddress()
	if from == "" {
		return types.ContractOutput{}, fmt.Errorf("from address is required")
	}

	if err := keys.ValidateEDDSAPublicKeyHex(from); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid from address: %w", err)
	}

	if contractVersion == "" {
		return types.ContractOutput{}, fmt.Errorf("contract version is required")
	}
	contractVersion, err := publicContractVersionV2(contractVersion)
	if err != nil {
		return types.ContractOutput{}, err
	}

	to := types.DEPLOY_CONTRACT_ADDRESS
	method := contractV2.METHOD_DEPLOY_CONTRACT

	data := map[string]interface{}{
		"contract_version": contractVersion,
	}

	version := uint8(1)

	uuid7, err := utils.NewUUID7()
	if err != nil {
		return types.ContractOutput{}, fmt.Errorf("failed to generate UUIDv7: %w", err)
	}

	contractOutput, err := c.SignAndSendTransaction(
		c.chainId,
		from,
		to,
		method,
		data,
		version,
		uuid7,
	)
	if err != nil {
		return types.ContractOutput{}, fmt.Errorf("failed to deploy contract: %w", err)
	}

	return contractOutput, nil
}

func (c *NetworkClient) DeployContract2(contractVersion, contractAddress string) (types.ContractOutput, error) {
	if c.walletManager == nil {
		return types.ContractOutput{}, fmt.Errorf("wallet manager is required")
	}

	from := c.walletManager.OwnerAddress()
	if from == "" {
		return types.ContractOutput{}, fmt.Errorf("from address is required")
	}

	if err := keys.ValidateEDDSAPublicKeyHex(from); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid from address: %w", err)
	}

	if contractVersion == "" {
		return types.ContractOutput{}, fmt.Errorf("contract version is required")
	}
	contractVersion, err := publicContractVersionV2(contractVersion)
	if err != nil {
		return types.ContractOutput{}, err
	}

	if contractAddress == "" {
		return types.ContractOutput{}, fmt.Errorf("contract address is required")
	}

	if err := keys.ValidateEDDSAPublicKeyHex(contractAddress); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid contract address: %w", err)
	}

	to := contractAddress
	method := contractV2.METHOD_DEPLOY_CONTRACT2

	data := map[string]interface{}{
		"contract_version": contractVersion,
	}

	version := uint8(1)

	uuid7, err := utils.NewUUID7()
	if err != nil {
		return types.ContractOutput{}, fmt.Errorf("failed to generate UUIDv7: %w", err)
	}

	contractOutput, err := c.SignAndSendTransaction(
		c.chainId,
		from,
		to,
		method,
		data,
		version,
		uuid7,
	)
	if err != nil {
		return types.ContractOutput{}, fmt.Errorf("failed to deploy contract: %w", err)
	}

	return contractOutput, nil
}

func publicContractVersionV2(version string) (string, error) {
	if _, ok := contractversionv2.Implementation(version); ok {
		return version, nil
	}
	return "", fmt.Errorf("unsupported native V2 contract version: %s", version)
}

func normalizePublicContractVersionsV2(value interface{}) (interface{}, error) {
	encoded, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	decoder := json.NewDecoder(bytes.NewReader(encoded))
	decoder.UseNumber()
	var decoded interface{}
	if err := decoder.Decode(&decoded); err != nil {
		return nil, err
	}
	return normalizeContractVersionValueV2(decoded)
}

func normalizeContractVersionValueV2(value interface{}) (interface{}, error) {
	switch typed := value.(type) {
	case map[string]interface{}:
		for key, child := range typed {
			if key == "contract_version" {
				version, ok := child.(string)
				if !ok {
					return nil, fmt.Errorf("contract_version must be a string")
				}
				public, err := publicContractVersionV2(version)
				if err != nil {
					return nil, err
				}
				typed[key] = public
				continue
			}
			normalized, err := normalizeContractVersionValueV2(child)
			if err != nil {
				return nil, err
			}
			typed[key] = normalized
		}
		return typed, nil
	case []interface{}:
		for index, child := range typed {
			normalized, err := normalizeContractVersionValueV2(child)
			if err != nil {
				return nil, err
			}
			typed[index] = normalized
		}
	}
	return value, nil
}
