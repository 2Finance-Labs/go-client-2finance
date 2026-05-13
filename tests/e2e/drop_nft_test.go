package e2e_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/2finance/2finance-network/blockchain/contract/dropV1"
	"gitlab.com/2finance/2finance-network/blockchain/contract/tokenV1"
	tokenV1Domain "gitlab.com/2finance/2finance-network/blockchain/contract/tokenV1/domain"
	tokenV1Models "gitlab.com/2finance/2finance-network/blockchain/contract/tokenV1/models"
	"gitlab.com/2finance/2finance-network/blockchain/log"
	"gitlab.com/2finance/2finance-network/blockchain/utils"
)

func TestDropFlowNFT(t *testing.T) {
	// --------------------------------------------------------------------
	// Owner local wallet
	// --------------------------------------------------------------------
	ownerSigner := setupSignerWallet(t)

	c := setupClient(t, ownerSigner.Wallet)

	useWallet(t, c, ownerSigner.Wallet)
	owner := createWallet(t, c, ownerSigner.PublicKey)

	// --------------------------------------------------------------------
	// Token setup (NFT - SEM createBasicToken)
	// --------------------------------------------------------------------
	useWallet(t, c, ownerSigner.Wallet)

	deployedContract, err := c.DeployContract1(tokenV1.TOKEN_CONTRACT_V1)
	require.NoError(t, err)

	contractLog, err := utils.UnmarshalLog[log.Log](deployedContract.Logs[0])
	require.NoError(t, err)

	nftTokenAddress := contractLog.ContractAddress
	require.NotEmpty(t, nftTokenAddress)

	addTokenOut, err := c.AddToken(
		nftTokenAddress,
		"DNFT"+randSuffix(4),
		"Drop NFT",
		0,
		"1",
		"drop nft e2e",
		owner.PublicKey,
		"https://example.com/image.png",
		"https://example.com",
		map[string]string{},
		map[string]string{},
		map[string]string{},
		"creator",
		"https://creator.com",
		map[string]bool{},
		map[string]bool{},
		map[string]bool{},
		[]map[string]interface{}{},
		owner.PublicKey,
		false,
		false,
		false,
		false,
		time.Time{},
		"https://example.com/asset.glb",
		tokenV1Domain.NON_FUNGIBLE,
		true,
		tokenV1Domain.TOKEN_ASSET_TYPE,
	)
	require.NoError(t, err)

	addTokenLog, err := utils.UnmarshalLog[log.Log](addTokenOut.Logs[0])
	require.NoError(t, err)
	assert.Equal(t, tokenV1Domain.TOKEN_CREATED_LOG, addTokenLog.LogType)

	// --------------------------------------------------------------------
	// Mint NFTs
	// --------------------------------------------------------------------
	useWallet(t, c, ownerSigner.Wallet)

	mintOut, err := c.MintToken(nftTokenAddress, owner.PublicKey, "3")
	require.NoError(t, err)

	mintLog, err := utils.UnmarshalLog[log.Log](mintOut.Logs[0])
	require.NoError(t, err)

	mintEvent, err := utils.UnmarshalEvent[tokenV1Domain.MintNFT](mintLog.Event)
	require.NoError(t, err)

	require.Len(t, mintEvent.TokenUUIDList, 3)

	uuid1 := mintEvent.TokenUUIDList[0]
	uuid2 := mintEvent.TokenUUIDList[1]

	// --------------------------------------------------------------------
	// Deploy Drop
	// --------------------------------------------------------------------
	useWallet(t, c, ownerSigner.Wallet)

	deployedDrop, err := c.DeployContract1(dropV1.DROP_CONTRACT_V1)
	require.NoError(t, err)

	dropLog, err := utils.UnmarshalLog[log.Log](deployedDrop.Logs[0])
	require.NoError(t, err)

	dropAddress := dropLog.ContractAddress
	require.NotEmpty(t, dropAddress)

	// --------------------------------------------------------------------
	// Create Drop
	// --------------------------------------------------------------------
	useWallet(t, c, ownerSigner.Wallet)

	tmpWM := setupWalletManager(t)

	programAddress, _ := genKey(t, tmpWM)
	tokenAddress, _ := genKey(t, tmpWM)

	startAt := time.Now()
	expireAt := time.Now().Add(24 * time.Hour)

	inputDrop := buildNewDropInput(
		dropAddress,
		programAddress,
		tokenAddress,
		owner.PublicKey,
		startAt,
		expireAt,
	)

	out, err := c.NewDrop(inputDrop)
	require.NoError(t, err)

	drop := assertCreatedDropLog(t, out, inputDrop)

	startAt = startAt.Add(1 * time.Second)
	time.Sleep(2 * time.Second)
	expireAt = expireAt.Add(50 * time.Hour)

	inputUpdate := buildUpdateDropInput(
		dropAddress,
		programAddress,
		nftTokenAddress,
		startAt,
		expireAt,
	)

	inputUpdate.ClaimAmount = "1"

	useWallet(t, c, ownerSigner.Wallet)

	_, err = c.UpdateDropMetadata(inputUpdate)
	require.NoError(t, err)

	// --------------------------------------------------------------------
	// Oracles
	// --------------------------------------------------------------------
	useWallet(t, c, ownerSigner.Wallet)

	oracles := buildOracleFixture(t)

	mustAllowOracles(t, c, drop.Address, map[string]bool{
		oracles.Oracle1: true,
	})

	// --------------------------------------------------------------------
	// Eligibility before attestation
	// --------------------------------------------------------------------
	userSigner := setupSignerWallet(t)

	useWallet(t, c, userSigner.Wallet)

	_, err = c.ClaimDrop(drop.Address)
	assertClaimDropError(t, err, "is not eligible")

	// --------------------------------------------------------------------
	// Oracle attests user eligibility
	// --------------------------------------------------------------------
	useWallet(t, c, oracles.Oracle1WM)

	_, err = c.AttestParticipantEligibility(drop.Address, userSigner.PublicKey, true)
	require.NoError(t, err)

	// --------------------------------------------------------------------
	// Deposit NFTs
	// --------------------------------------------------------------------
	useWallet(t, c, ownerSigner.Wallet)

	depositUUIDs := []string{uuid1, uuid2}
	fmt.Printf("Depositing NFTs with UUIDs: %v\n", depositUUIDs)

	_, err = c.DepositDrop(
		drop.Address,
		programAddress,
		nftTokenAddress,
		"2",
		depositUUIDs,
	)
	require.NoError(t, err)

	// --------------------------------------------------------------------
	// Withdraw 1 NFT
	// --------------------------------------------------------------------
	useWallet(t, c, ownerSigner.Wallet)

	_, err = c.WithdrawDrop(
		drop.Address,
		programAddress,
		nftTokenAddress,
		"1",
		[]string{uuid2},
	)
	require.NoError(t, err)

	getBalance, err := c.ListTokenBalances(
		nftTokenAddress,
		drop.Address,
		tokenV1Domain.NON_FUNGIBLE,
		1,
		10,
		true,
	)
	require.NoError(t, err)

	var balanceList []tokenV1Models.BalanceStateModel
	err = utils.UnmarshalState[[]tokenV1Models.BalanceStateModel](getBalance.States[0].Object, &balanceList)
	require.NoError(t, err)

	assert.Len(t, getBalance.States, 1)
	assert.Len(t, balanceList, 1)
	assert.Equal(t, tokenV1Domain.NON_FUNGIBLE, balanceList[0].TokenType)

	// --------------------------------------------------------------------
	// Claim
	// --------------------------------------------------------------------
	useWallet(t, c, userSigner.Wallet)

	outClaim, err := c.ClaimDrop(drop.Address)
	require.NoError(t, err)

	assertClaimDropLog(
		t,
		outClaim,
		drop.Address,
		userSigner.PublicKey,
		programAddress,
		nftTokenAddress,
		"1",
	)

	// --------------------------------------------------------------------
	// Validate user received NFT
	// --------------------------------------------------------------------
	userBalanceOut, err := c.GetTokenBalanceNFT(nftTokenAddress, userSigner.PublicKey, uuid1)
	require.NoError(t, err)

	var userBalance tokenV1Models.BalanceStateModel
	err = utils.UnmarshalState[tokenV1Models.BalanceStateModel](userBalanceOut.States[0].Object, &userBalance)
	require.NoError(t, err)

	assert.Equal(t, userSigner.PublicKey, userBalance.OwnerAddress)
	assert.Equal(t, uuid1, userBalance.TokenUUID)
	assert.Equal(t, "1", userBalance.Amount)
}