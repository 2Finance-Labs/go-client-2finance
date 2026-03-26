package e2e_test

import (
	"testing"
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/2finance/2finance-network/blockchain/log"
	tokenV1Domain "gitlab.com/2finance/2finance-network/blockchain/contract/tokenV1/domain"
	"gitlab.com/2finance/2finance-network/blockchain/utils"
	"gitlab.com/2finance/2finance-network/blockchain/types"
	dropV1Domain "gitlab.com/2finance/2finance-network/blockchain/contract/dropV1/domain"
	dropV1Models "gitlab.com/2finance/2finance-network/blockchain/contract/dropV1/models"
	dropV1Inputs "gitlab.com/2finance/2finance-network/blockchain/contract/dropV1/inputs"
	"gitlab.com/2finance/2finance-network/blockchain/contract/dropV1"
	"time"
	client2f "github.com/2Finance-Labs/go-client-2finance/client_2finance"
	

)

type dropCreateFixture struct {
	Input dropV1Inputs.InputNewDrop
}

type dropUpdateFixture struct {
	Input dropV1Inputs.InputUpdateDropMetadata
}

func buildNewDropInput(
	address string,
	programAddress string,
	tokenAddress string,
	owner string,
	startAt time.Time,
	expireAt time.Time,
) dropV1Inputs.InputNewDrop {
	return dropV1Inputs.InputNewDrop{
		Address:              address,
		ProgramAddress:       programAddress,
		TokenAddress:         tokenAddress,
		Owner:                owner,
		Title:                "Airdrop E2E",
		Description:          "E2E description",
		ShortDescription:     "Short desc",
		ImageURL:             "https://img.png",
		BannerURL:            "https://banner.png",
		Category:             "airdrop",
		SocialRequirements:   map[string]bool{"FOLLOW_X": true},
		PostLinks:            map[string]bool{"https://x.com/post/123": true},
		VerificationType:     "ORACLE",
		StartAt:              startAt,
		ExpireAt:             expireAt,
		RequestLimit:         100,
		ClaimAmount:          "1000",
		ClaimIntervalSeconds: 3600,
	}
}

func buildUpdateDropInput(
	address string,
	programAddress string,
	tokenAddress string,
	startAt time.Time,
	expireAt time.Time,
) dropV1Inputs.InputUpdateDropMetadata {
	return dropV1Inputs.InputUpdateDropMetadata{
		Address:              address,
		ProgramAddress:       programAddress,
		TokenAddress:         tokenAddress,
		Title:                "Airdrop E2E (UPDATED)",
		Description:          "Updated description",
		ShortDescription:     "Updated short description",
		ImageURL:             "https://img-updated.png",
		BannerURL:            "https://banner-updated.png",
		Category:             "airdrop",
		SocialRequirements:   map[string]bool{"LIKE_X": true},
		PostLinks:            map[string]bool{"https://x.com/post/456": true},
		StartAt:              startAt,
		ExpireAt:             expireAt,
		VerificationType:     "ORACLE",
		RequestLimit:         10,
		ClaimAmount:          "2",
		ClaimIntervalSeconds: 1800,
	}
}

func assertCreatedDropLog(
	t *testing.T,
	out types.ContractOutput,
	input dropV1Inputs.InputNewDrop,
) dropV1Domain.Drop {
	t.Helper()

	unmarshalLogToken, err := utils.UnmarshalLog[log.Log](out.Logs[0])
	if err != nil {
		t.Fatalf("UnmarshalLog (NewDrop.Logs[0]): %v", err)
	}

	assert.Equal(t, dropV1Domain.DROP_CREATED_LOG, unmarshalLogToken.LogType)

	drop, err := utils.UnmarshalEvent[dropV1Domain.Drop](unmarshalLogToken.Event)
	if err != nil {
		t.Fatalf("UnmarshalEvent (NewDrop.Logs[0]): %v", err)
	}

	assert.Equal(t, input.Address, drop.Address)
	assert.Equal(t, input.ProgramAddress, drop.ProgramAddress)
	assert.Equal(t, input.TokenAddress, drop.TokenAddress)
	assert.Equal(t, input.Owner, drop.Owner)
	assert.Equal(t, input.Title, drop.Title)
	assert.Equal(t, input.Description, drop.Description)
	assert.Equal(t, input.ShortDescription, drop.ShortDescription)
	assert.Equal(t, input.ImageURL, drop.ImageURL)
	assert.Equal(t, input.BannerURL, drop.BannerURL)
	assert.Equal(t, input.Category, drop.Category)
	assert.Equal(t, input.SocialRequirements, drop.SocialRequirements)
	assert.Equal(t, input.PostLinks, drop.PostLinks)
	assert.Equal(t, input.VerificationType, drop.VerificationType)
	assert.WithinDuration(t, input.StartAt, drop.StartAt, time.Second)
	assert.WithinDuration(t, input.ExpireAt, drop.ExpireAt, time.Second)
	assert.Equal(t, input.RequestLimit, drop.RequestLimit)
	assert.Equal(t, input.ClaimAmount, drop.ClaimAmount)
	assert.Equal(t, input.ClaimIntervalSeconds, drop.ClaimIntervalSeconds)

	return drop
}

func assertDropState(
	t *testing.T,
	gotOut types.ContractOutput,
	input dropV1Inputs.InputNewDrop,
) {
	t.Helper()

	var state dropV1Models.DropStateModel
	err := utils.UnmarshalState[dropV1Models.DropStateModel](gotOut.States[0].Object, &state)
	if err != nil {
		t.Fatalf("UnmarshalState (GetDrop.States[0]): %v", err)
	}

	assert.Equal(t, input.Address, state.Address)
	assert.Equal(t, input.ProgramAddress, state.ProgramAddress)
	assert.Equal(t, input.TokenAddress, state.TokenAddress)
	assert.Equal(t, input.Owner, state.Owner)
	assert.Equal(t, input.Title, state.Title)
	assert.Equal(t, input.Description, state.Description)
	assert.Equal(t, input.ShortDescription, state.ShortDescription)
	assert.Equal(t, input.ImageURL, state.ImageURL)
	assert.Equal(t, input.BannerURL, state.BannerURL)
	assert.Equal(t, input.Category, state.Category)
	assert.Equal(t, input.SocialRequirements, state.SocialRequirements)
	assert.Equal(t, input.PostLinks, state.PostLinks)
	assert.Equal(t, input.VerificationType, state.VerificationType)
	assert.WithinDuration(t, input.StartAt, state.StartAt, 5*time.Second)
	assert.WithinDuration(t, input.ExpireAt, state.ExpireAt, 5*time.Second)
	assert.Equal(t, input.RequestLimit, state.RequestLimit)
	assert.Equal(t, input.ClaimAmount, state.ClaimAmount)
	assert.Equal(t, input.ClaimIntervalSeconds, state.ClaimIntervalSeconds)
}

func assertUpdatedDropLog(
	t *testing.T,
	out types.ContractOutput,
	input dropV1Inputs.InputUpdateDropMetadata,
) dropV1Domain.Drop {
	t.Helper()

	unmarshalLogToken, err := utils.UnmarshalLog[log.Log](out.Logs[0])
	if err != nil {
		t.Fatalf("UnmarshalLog (UpdateDropMetadata.Logs[0]): %v", err)
	}

	assert.Equal(t, dropV1Domain.DROP_METADATA_UPDATED_LOG, unmarshalLogToken.LogType)

	dropUpdated, err := utils.UnmarshalEvent[dropV1Domain.Drop](unmarshalLogToken.Event)
	if err != nil {
		t.Fatalf("UnmarshalEvent (UpdateDropMetadata.Logs[0]): %v", err)
	}

	assert.Equal(t, input.Address, dropUpdated.Address)
	assert.Equal(t, input.ProgramAddress, dropUpdated.ProgramAddress)
	assert.Equal(t, input.TokenAddress, dropUpdated.TokenAddress)
	assert.Equal(t, input.Title, dropUpdated.Title)
	assert.Equal(t, input.Description, dropUpdated.Description)
	assert.Equal(t, input.ShortDescription, dropUpdated.ShortDescription)
	assert.Equal(t, input.ImageURL, dropUpdated.ImageURL)
	assert.Equal(t, input.BannerURL, dropUpdated.BannerURL)
	assert.Equal(t, input.Category, dropUpdated.Category)
	assert.Equal(t, input.SocialRequirements, dropUpdated.SocialRequirements)
	assert.Equal(t, input.PostLinks, dropUpdated.PostLinks)
	assert.Equal(t, input.VerificationType, dropUpdated.VerificationType)
	assert.WithinDuration(t, input.StartAt, dropUpdated.StartAt, 10*time.Second)
	assert.WithinDuration(t, input.ExpireAt, dropUpdated.ExpireAt, 10*time.Second)
	assert.Equal(t, input.RequestLimit, dropUpdated.RequestLimit)
	assert.Equal(t, input.ClaimAmount, dropUpdated.ClaimAmount)
	assert.Equal(t, input.ClaimIntervalSeconds, dropUpdated.ClaimIntervalSeconds)

	return dropUpdated
}

func assertUpdatedDropState(
	t *testing.T,
	gotOut types.ContractOutput,
	input dropV1Inputs.InputUpdateDropMetadata,
	expectedSocialRequirements map[string]bool,
	expectedPostLinks map[string]bool,
) {
	t.Helper()

	var state dropV1Models.DropStateModel
	err := utils.UnmarshalState[dropV1Models.DropStateModel](gotOut.States[0].Object, &state)
	if err != nil {
		t.Fatalf("UnmarshalState (GetDrop.States[0]): %v", err)
	}

	assert.Equal(t, input.Address, state.Address)
	assert.Equal(t, input.ProgramAddress, state.ProgramAddress)
	assert.Equal(t, input.TokenAddress, state.TokenAddress)
	assert.Equal(t, input.Title, state.Title)
	assert.Equal(t, input.Description, state.Description)
	assert.Equal(t, input.ShortDescription, state.ShortDescription)
	assert.Equal(t, input.ImageURL, state.ImageURL)
	assert.Equal(t, input.BannerURL, state.BannerURL)
	assert.Equal(t, input.Category, state.Category)
	assert.Equal(t, expectedSocialRequirements, state.SocialRequirements)
	assert.Equal(t, expectedPostLinks, state.PostLinks)
	assert.Equal(t, input.VerificationType, state.VerificationType)
	assert.WithinDuration(t, input.StartAt, state.StartAt, 10*time.Second)
	assert.WithinDuration(t, input.ExpireAt, state.ExpireAt, 10*time.Second)
	assert.Equal(t, input.RequestLimit, state.RequestLimit)
	assert.Equal(t, input.ClaimAmount, state.ClaimAmount)
	assert.Equal(t, input.ClaimIntervalSeconds, state.ClaimIntervalSeconds)
}

type oracleFixture struct {
	Oracle1     string
	Oracle1Priv string
	Oracle2     string
	Oracle3     string
}

func buildOracleFixture(t *testing.T, c client2f.Client2FinanceNetwork) oracleFixture {
	t.Helper()

	oracle1, oracle1Priv := genKey(t, c)
	oracle2, _ := genKey(t, c)
	oracle3, _ := genKey(t, c)

	return oracleFixture{
		Oracle1:     oracle1,
		Oracle1Priv: oracle1Priv,
		Oracle2:     oracle2,
		Oracle3:     oracle3,
	}
}

func mustAllowOracles(
	t *testing.T,
	c client2f.Client2FinanceNetwork,
	address string,
	oracles map[string]bool,
) types.ContractOutput {
	t.Helper()

	out, err := c.AllowOracles(address, oracles)
	if err != nil {
		t.Fatalf("AllowOracles: %v", err)
	}

	return out
}

func assertAllowOraclesLog(
	t *testing.T,
	out types.ContractOutput,
	dropAddress string,
	expected map[string]bool,
) dropV1Domain.Drop {
	t.Helper()

	unmarshalLogAllowOracles, err := utils.UnmarshalLog[log.Log](out.Logs[0])
	if err != nil {
		t.Fatalf("UnmarshalLog (AllowOracles.Logs[0]): %v", err)
	}

	assert.Equal(
		t,
		dropV1Domain.DROP_ORACLES_ALLOWED_LOG,
		unmarshalLogAllowOracles.LogType,
		"allow-oracles log type mismatch",
	)

	allowedOracles, err := utils.UnmarshalEvent[dropV1Domain.Drop](unmarshalLogAllowOracles.Event)
	if err != nil {
		t.Fatalf("UnmarshalEvent (AllowOracles.Event): %v", err)
	}

	assert.Equal(t, dropAddress, allowedOracles.Address, "allowed oracles drop address mismatch")
	assert.Equal(t, expected, allowedOracles.AllowedOracles, "allowed oracles mismatch")

	return allowedOracles
}

func mustDisallowOracles(
	t *testing.T,
	c client2f.Client2FinanceNetwork,
	address string,
	oracles map[string]bool,
) types.ContractOutput {
	t.Helper()

	out, err := c.DisallowOracles(address, oracles)
	if err != nil {
		t.Fatalf("DisallowOracles: %v", err)
	}

	return out
}

func assertDisallowOraclesLog(
	t *testing.T,
	out types.ContractOutput,
	dropAddress string,
	expected map[string]bool,
) dropV1Domain.Drop {
	t.Helper()

	unmarshalLogDisallowOracles, err := utils.UnmarshalLog[log.Log](out.Logs[0])
	if err != nil {
		t.Fatalf("UnmarshalLog (DisallowOracles.Logs[0]): %v", err)
	}

	assert.Equal(
		t,
		dropV1Domain.DROP_ORACLES_DISALLOWED_LOG,
		unmarshalLogDisallowOracles.LogType,
		"disallow-oracles log type mismatch",
	)

	disallowedOracles, err := utils.UnmarshalEvent[dropV1Domain.Drop](unmarshalLogDisallowOracles.Event)
	if err != nil {
		t.Fatalf("UnmarshalEvent (DisallowOracles.Event): %v", err)
	}

	assert.Equal(t, dropAddress, disallowedOracles.Address, "disallowed oracles drop address mismatch")
	assert.Equal(t, expected, disallowedOracles.AllowedOracles, "disallowed oracles mismatch")

	return disallowedOracles
}

func assertDropAllowedOraclesState(
	t *testing.T,
	gotOut types.ContractOutput,
	expected map[string]bool,
) {
	t.Helper()

	var dropStateModelOracles dropV1Models.DropStateModel
	err := utils.UnmarshalState[dropV1Models.DropStateModel](gotOut.States[0].Object, &dropStateModelOracles)
	if err != nil {
		t.Fatalf("UnmarshalState (GetDrop.States[0]): %v", err)
	}

	assert.Equal(t, expected, dropStateModelOracles.AllowedOracles, "GetDrop allowed oracles mismatch")
}

func mustPauseDrop(
	t *testing.T,
	c client2f.Client2FinanceNetwork,
	address string,
) types.ContractOutput {
	t.Helper()

	out, err := c.PauseDrop(address)
	if err != nil {
		t.Fatalf("PauseDrop: %v", err)
	}

	return out
}

func assertPauseDropLog(
	t *testing.T,
	out types.ContractOutput,
	dropAddress string,
) dropV1Domain.Drop {
	t.Helper()

	unmarshalLogPause, err := utils.UnmarshalLog[log.Log](out.Logs[0])
	if err != nil {
		t.Fatalf("UnmarshalLog (PauseDrop.Logs[0]): %v", err)
	}

	assert.Equal(
		t,
		dropV1Domain.DROP_PAUSED_LOG,
		unmarshalLogPause.LogType,
		"pause-drop log type mismatch",
	)

	pausedDrop, err := utils.UnmarshalEvent[dropV1Domain.Drop](unmarshalLogPause.Event)
	if err != nil {
		t.Fatalf("UnmarshalEvent (PauseDrop.Event): %v", err)
	}

	assert.Equal(t, dropAddress, pausedDrop.Address, "paused drop address mismatch")
	assert.Equal(t, true, pausedDrop.Paused, "drop paused state mismatch")

	return pausedDrop
}

func mustUnpauseDrop(
	t *testing.T,
	c client2f.Client2FinanceNetwork,
	address string,
) types.ContractOutput {
	t.Helper()

	out, err := c.UnpauseDrop(address)
	if err != nil {
		t.Fatalf("UnpauseDrop: %v", err)
	}

	return out
}

func assertUnpauseDropLog(
	t *testing.T,
	out types.ContractOutput,
	dropAddress string,
) dropV1Domain.Drop {
	t.Helper()

	unmarshalLogUnpause, err := utils.UnmarshalLog[log.Log](out.Logs[0])
	if err != nil {
		t.Fatalf("UnmarshalLog (UnpauseDrop.Logs[0]): %v", err)
	}

	assert.Equal(
		t,
		dropV1Domain.DROP_UNPAUSED_LOG,
		unmarshalLogUnpause.LogType,
		"unpause-drop log type mismatch",
	)

	unpausedDrop, err := utils.UnmarshalEvent[dropV1Domain.Drop](unmarshalLogUnpause.Event)
	if err != nil {
		t.Fatalf("UnmarshalEvent (UnpauseDrop.Event): %v", err)
	}

	assert.Equal(t, dropAddress, unpausedDrop.Address, "unpaused drop address mismatch")
	assert.Equal(t, false, unpausedDrop.Paused, "drop paused state mismatch")

	return unpausedDrop
}

func assertDropPausedState(
	t *testing.T,
	gotOut types.ContractOutput,
	expected bool,
) {
	t.Helper()

	var dropStateModelPause dropV1Models.DropStateModel
	err := utils.UnmarshalState[dropV1Models.DropStateModel](gotOut.States[0].Object, &dropStateModelPause)
	if err != nil {
		t.Fatalf("UnmarshalState (GetDrop.States[0]): %v", err)
	}

	assert.Equal(t, expected, dropStateModelPause.Paused, "GetDrop paused state mismatch")
}

func mustAttestParticipantEligibility(
	t *testing.T,
	c client2f.Client2FinanceNetwork,
	dropAddress string,
	wallet string,
	eligible bool,
) types.ContractOutput {
	t.Helper()

	out, err := c.AttestParticipantEligibility(dropAddress, wallet, eligible)
	if err != nil {
		t.Fatalf("AttestParticipantEligibility: %v", err)
	}

	return out
}

func assertAttestParticipantEligibilityLog(
	t *testing.T,
	out types.ContractOutput,
	expectedDropAddress string,
	expectedWallet string,
	expectedEligible bool,
	expectedVerificationType string,
) dropV1Domain.EligibilityAttested {
	t.Helper()

	unmarshalLogAttest, err := utils.UnmarshalLog[log.Log](out.Logs[0])
	if err != nil {
		t.Fatalf("UnmarshalLog (AttestParticipantEligibility.Logs[0]): %v", err)
	}

	assert.Equal(
		t,
		dropV1Domain.DROP_ATTESTED_PARTICIPANT_ELIGIBILITY_LOG,
		unmarshalLogAttest.LogType,
		"attest-participant log type mismatch",
	)

	attestedParticipant, err := utils.UnmarshalEvent[dropV1Domain.EligibilityAttested](unmarshalLogAttest.Event)
	if err != nil {
		t.Fatalf("UnmarshalEvent (AttestParticipantEligibility.Event): %v", err)
	}

	assert.Equal(t, expectedDropAddress, attestedParticipant.DropAddress, "attested participant drop address mismatch")
	assert.Equal(t, expectedWallet, attestedParticipant.Wallet, "attested participant address mismatch")
	assert.Equal(t, expectedEligible, attestedParticipant.Eligible, "attested participant eligibility mismatch")
	assert.Equal(t, expectedVerificationType, attestedParticipant.VerificationType, "attested participant verification type mismatch")

	return attestedParticipant
}

func attestParticipantEligibilityAndAssert(
	t *testing.T,
	c client2f.Client2FinanceNetwork,
	signerPriv string,
	dropAddress string,
	wallet string,
	eligible bool,
	expectedVerificationType string,
) dropV1Domain.EligibilityAttested {
	t.Helper()

	c.SetPrivateKey(signerPriv)

	out := mustAttestParticipantEligibility(t, c, dropAddress, wallet, eligible)

	return assertAttestParticipantEligibilityLog(
		t,
		out,
		dropAddress,
		wallet,
		eligible,
		expectedVerificationType,
	)
}

func setSigner(t *testing.T, c client2f.Client2FinanceNetwork, priv string) {
	t.Helper()
	c.SetPrivateKey(priv)
}

func mustDepositDrop(
	t *testing.T,
	c client2f.Client2FinanceNetwork,
	address string,
	programAddress string,
	tokenAddress string,
	amount string,
	tokenUUIDs []string,
) types.ContractOutput {
	t.Helper()

	out, err := c.DepositDrop(
		address,
		programAddress,
		tokenAddress,
		amount,
		tokenUUIDs,
	)
	if err != nil {
		t.Fatalf("DepositDrop: %v", err)
	}

	return out
}

func assertDepositDropLog(
	t *testing.T,
	out types.ContractOutput,
	expectedAddress string,
	expectedProgramAddress string,
	expectedTokenAddress string,
	expectedSenderAddress string,
	expectedAmount string,
) dropV1Domain.DepositFunds {
	t.Helper()

	unmarshalLogDeposit, err := utils.UnmarshalLog[log.Log](out.Logs[0])
	if err != nil {
		t.Fatalf("UnmarshalLog (DepositDrop.Logs[0]): %v", err)
	}

	assert.Equal(
		t,
		dropV1Domain.DROP_DEPOSITED_LOG,
		unmarshalLogDeposit.LogType,
		"deposit-drop-funds log type mismatch",
	)

	depositedFunds, err := utils.UnmarshalEvent[dropV1Domain.DepositFunds](unmarshalLogDeposit.Event)
	if err != nil {
		t.Fatalf("UnmarshalEvent (DepositDrop.Event): %v", err)
	}

	assert.Equal(t, expectedAddress, depositedFunds.Address, "deposited funds drop address mismatch")
	assert.Equal(t, expectedProgramAddress, depositedFunds.ProgramAddress, "deposited funds program address mismatch")
	assert.Equal(t, expectedTokenAddress, depositedFunds.TokenAddress, "deposited funds token address mismatch")
	assert.Equal(t, expectedSenderAddress, depositedFunds.SenderAddress, "deposited funds sender mismatch")
	assert.Equal(t, expectedAmount, depositedFunds.Amount, "deposited funds amount mismatch")

	return depositedFunds
}

func mustWithdrawDrop(
	t *testing.T,
	c client2f.Client2FinanceNetwork,
	address string,
	programAddress string,
	tokenAddress string,
	amount string,
	tokenUUIDs []string,
) types.ContractOutput {
	t.Helper()

	out, err := c.WithdrawDrop(
		address,
		programAddress,
		tokenAddress,
		amount,
		tokenUUIDs,
	)
	if err != nil {
		t.Fatalf("WithdrawDropFunds: %v", err)
	}

	return out
}

func assertWithdrawDropLog(
	t *testing.T,
	out types.ContractOutput,
	expectedAddress string,
	expectedProgramAddress string,
	expectedTokenAddress string,
	expectedReceiverAddress string,
	expectedAmount string,
) dropV1Domain.WithdrawFunds {
	t.Helper()

	unmarshalLogWithdraw, err := utils.UnmarshalLog[log.Log](out.Logs[0])
	if err != nil {
		t.Fatalf("UnmarshalLog (WithdrawDropFunds.Logs[0]): %v", err)
	}

	assert.Equal(
		t,
		dropV1Domain.DROP_WITHDRAWN_LOG,
		unmarshalLogWithdraw.LogType,
		"withdraw-drop-funds log type mismatch",
	)

	withdrawFunds, err := utils.UnmarshalEvent[dropV1Domain.WithdrawFunds](unmarshalLogWithdraw.Event)
	if err != nil {
		t.Fatalf("UnmarshalEvent (WithdrawDropFunds.Event): %v", err)
	}

	assert.Equal(t, expectedAddress, withdrawFunds.Address, "withdrew funds drop address mismatch")
	assert.Equal(t, expectedProgramAddress, withdrawFunds.ProgramAddress, "withdrew funds program address mismatch")
	assert.Equal(t, expectedTokenAddress, withdrawFunds.TokenAddress, "withdrew funds token address mismatch")
	assert.Equal(t, expectedReceiverAddress, withdrawFunds.ReceiverAddress, "withdrew funds receiver address mismatch")
	assert.Equal(t, expectedAmount, withdrawFunds.Amount, "withdrew funds amount mismatch")

	return withdrawFunds
}

func mustClaimDrop(
	t *testing.T,
	c client2f.Client2FinanceNetwork,
	address string,
) types.ContractOutput {
	t.Helper()

	out, err := c.ClaimDrop(address)
	if err != nil {
		t.Fatalf("ClaimDrop: %v", err)
	}

	return out
}

func assertClaimDropError(
	t *testing.T,
	err error,
	expectedMessage string,
) {
	t.Helper()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), expectedMessage)
}

func assertClaimDropLog(
	t *testing.T,
	out types.ContractOutput,
	expectedAddress string,
	expectedWallet string,
	expectedProgramAddress string,
	expectedTokenAddress string,
	expectedClaimAmount string,
) dropV1Domain.Claim {
	t.Helper()

	unmarshalLogClaimEligible, err := utils.UnmarshalLog[log.Log](out.Logs[0])
	if err != nil {
		t.Fatalf("UnmarshalLog (ClaimDrop.Logs[0]): %v", err)
	}

	assert.Equal(
		t,
		dropV1Domain.DROP_CLAIMED_LOG,
		unmarshalLogClaimEligible.LogType,
		"claim-drop log type mismatch",
	)

	claimedDrop, err := utils.UnmarshalEvent[dropV1Domain.Claim](unmarshalLogClaimEligible.Event)
	if err != nil {
		t.Fatalf("UnmarshalEvent (ClaimDrop.Event): %v", err)
	}

	assert.Equal(t, expectedAddress, claimedDrop.Address, "claimed drop address mismatch")
	assert.Equal(t, expectedWallet, claimedDrop.Wallet, "claimed participant address mismatch")
	assert.Equal(t, expectedProgramAddress, claimedDrop.ProgramAddress, "claimed drop program address mismatch")
	assert.Equal(t, expectedTokenAddress, claimedDrop.TokenAddress, "claimed drop token address mismatch")
	assert.Equal(t, expectedClaimAmount, claimedDrop.ClaimAmount, "claimed amount mismatch")

	return claimedDrop
}

func TestDropFlow(t *testing.T) {
	c := setupClient(t)

	owner, ownerPriv := createWallet(t, c)

	// --------------------------------------------------------------------
	// Token setup
	// --------------------------------------------------------------------
	c.SetPrivateKey(ownerPriv)

	dec := 6
	tokenType := tokenV1Domain.FUNGIBLE
	stablecoin := false
	tok := createBasicToken(t, c, owner.PublicKey, dec, true, tokenType, stablecoin)
	amount := "10000"
	if _, err := c.MintToken(tok.Address, owner.PublicKey, amount); err != nil {
		t.Fatalf("MintToken: %v", err)
	}
	// // --------------------------------------------------------------------
	// Deploy Drop contract
	// --------------------------------------------------------------------
	deployedContract, err := c.DeployContract1(dropV1.DROP_CONTRACT_V1)
	if err != nil {
		t.Fatalf("DeployContract: %v", err)
	}
	contractLog, err := utils.UnmarshalLog[log.Log](deployedContract.Logs[0])
	if err != nil {
		t.Fatalf("UnmarshalLog (AddWallet.Logs[0]): %v", err)
	}
	address := contractLog.ContractAddress

	// --------------------------------------------------------------------
	// Create Drop
	// --------------------------------------------------------------------
	programAddress, _ := genKey(t, c)
	tokenAddress, _ := genKey(t, c)

	startAt := time.Now()
	expireAt := time.Now().Add(24 * time.Hour)

	inputDrop := buildNewDropInput(
		address,
		programAddress,
		tokenAddress,
		owner.PublicKey,
		startAt,
		expireAt,
	)

	out, err := c.NewDrop(inputDrop)
	if err != nil {
		t.Fatalf("NewDrop: %v", err)
	}

	drop := assertCreatedDropLog(t, out, inputDrop)

	gotOut, err := c.GetDrop(drop.Address)
	if err != nil {
		t.Fatalf("GetDrop: %v", err)
	}
	assertDropState(t, gotOut, inputDrop)

	startAt = startAt.Add(1 * time.Second)
	time.Sleep(2 * time.Second)
	expireAt = expireAt.Add(50 * time.Hour)

	inputUpdateDrop := buildUpdateDropInput(
		address,
		programAddress,
		tok.Address,
		startAt,
		expireAt,
	)

	// Update Drop Metadata

	outMeta, err := c.UpdateDropMetadata(inputUpdateDrop)
	if err != nil {
		t.Fatalf("UpdateDropMetadata: %v", err)
	}

	assertUpdatedDropLog(t, outMeta, inputUpdateDrop)

	gotUpdatedDrop, err := c.GetDrop(address)
	if err != nil {
		t.Fatalf("GetDrop: %v", err)
	}

	assertUpdatedDropState(
		t,
		gotUpdatedDrop,
		inputUpdateDrop,
		map[string]bool{"FOLLOW_X": true, "LIKE_X": true},
		map[string]bool{"https://x.com/post/123": true, "https://x.com/post/456": true},
	)

	// ALLOW / DISALLOW ORACLES
	oracles := buildOracleFixture(t, c)

	allowedMap := map[string]bool{
		oracles.Oracle1: true,
		oracles.Oracle2: true,
		oracles.Oracle3: true,
	}

	allowOraclesOut := mustAllowOracles(t, c, drop.Address, allowedMap)
	assertAllowOraclesLog(t, allowOraclesOut, drop.Address, allowedMap)

	disallowedMap := map[string]bool{
		oracles.Oracle2: true,
	}

	disallowOraclesOut := mustDisallowOracles(t, c, drop.Address, disallowedMap)
	assertDisallowOraclesLog(t, disallowOraclesOut, drop.Address, disallowedMap)

	gotUpdatedDrop, err = c.GetDrop(address)
	if err != nil {
		t.Fatalf("GetDrop: %v", err)
	}

	resultAllowedOracles := map[string]bool{
		oracles.Oracle1: true,
		oracles.Oracle3: true,
	}

	assertDropAllowedOraclesState(t, gotUpdatedDrop, resultAllowedOracles)

	
	// PAUSE / UNPAUSE
	outPause := mustPauseDrop(t, c, address)
	assertPauseDropLog(t, outPause, drop.Address)

	outUnpause := mustUnpauseDrop(t, c, address)
	assertUnpauseDropLog(t, outUnpause, drop.Address)

	gotUpdatedDrop, err = c.GetDrop(address)
	if err != nil {
		t.Fatalf("GetDrop: %v", err)
	}

	assertDropPausedState(t, gotUpdatedDrop, false)
		

	// ATTEST ELIGIBILITY (ORACLE)

	eligible1, _ := genKey(t, c)
	attestParticipantEligibilityAndAssert(
		t,
		c,
		oracles.Oracle1Priv,
		drop.Address,
		eligible1,
		true,
		dropV1Domain.VERIFICATION_TYPE_ORACLE,
	)

	eligible2, _ := genKey(t, c)
	attestParticipantEligibilityAndAssert(
		t,
		c,
		ownerPriv,
		drop.Address,
		eligible2,
		true,
		dropV1Domain.VERIFICATION_TYPE_ORACLE,
	)
	
	
	// // Deposit Funds
	amountDeposit := "10"
	outDepositFunds := mustDepositDrop(
		t,
		c,
		inputUpdateDrop.Address,
		inputUpdateDrop.ProgramAddress,
		inputUpdateDrop.TokenAddress,
		amountDeposit,
		[]string{},
	)

	assertDepositDropLog(
		t,
		outDepositFunds,
		inputUpdateDrop.Address,
		inputUpdateDrop.ProgramAddress,
		inputUpdateDrop.TokenAddress,
		owner.PublicKey,
		amountDeposit,
	)
	// Withdraw Funds
	amountWithdraw := "3"
	outWithdrawFunds := mustWithdrawDrop(
		t,
		c,
		inputUpdateDrop.Address,
		inputUpdateDrop.ProgramAddress,
		inputUpdateDrop.TokenAddress,
		amountWithdraw,
		[]string{},
	)

	assertWithdrawDropLog(
		t,
		outWithdrawFunds,
		inputUpdateDrop.Address,
		inputUpdateDrop.ProgramAddress,
		inputUpdateDrop.TokenAddress,
		owner.PublicKey,
		amountWithdraw,
	)
	
	// Claim Drop
	userPub, userPriv := genKey(t, c)

	c.SetPrivateKey(userPriv)
	_, err = c.ClaimDrop(drop.Address)
	assertClaimDropError(t, err, "is not eligible for this drop")

	c.SetPrivateKey(oracles.Oracle1Priv)
	_, err = c.AttestParticipantEligibility(drop.Address, userPub, true)
	if err != nil {
		t.Fatalf("AttestParticipantEligibility: %v", err)
	}

	c.SetPrivateKey(userPriv)
	outClaimDropEligible := mustClaimDrop(t, c, drop.Address)

	assertClaimDropLog(
		t,
		outClaimDropEligible,
		drop.Address,
		userPub,
		inputUpdateDrop.ProgramAddress,
		inputUpdateDrop.TokenAddress,
		inputUpdateDrop.ClaimAmount,
	)
}


func mustGetDrop(
	t *testing.T,
	c client2f.Client2FinanceNetwork,
	address string,
) types.ContractOutput {
	t.Helper()

	out, err := c.GetDrop(address)
	if err != nil {
		t.Fatalf("GetDrop: %v", err)
	}

	return out
}

func mustListDrops(
	t *testing.T,
	c client2f.Client2FinanceNetwork,
	owner string,
	page, limit int,
	ascending bool,
) types.ContractOutput {
	t.Helper()

	out, err := c.ListDrops(owner, page, limit, ascending)
	if err != nil {
		t.Fatalf("ListDrops: %v", err)
	}

	return out
}

func assertGetDropState(
	t *testing.T,
	out types.ContractOutput,
	expectedAddress string,
	expectedOwner string,
	expectedProgramAddress string,
	expectedTokenAddress string,
	expectedTitle string,
) dropV1Models.DropStateModel {
	t.Helper()

	require.Len(t, out.States, 1)

	var state dropV1Models.DropStateModel
	err := utils.UnmarshalState[dropV1Models.DropStateModel](out.States[0].Object, &state)
	if err != nil {
		t.Fatalf("UnmarshalState (GetDrop.States[0]): %v", err)
	}

	assert.Equal(t, expectedAddress, state.Address)
	assert.Equal(t, expectedOwner, state.Owner)
	assert.Equal(t, expectedProgramAddress, state.ProgramAddress)
	assert.Equal(t, expectedTokenAddress, state.TokenAddress)
	assert.Equal(t, expectedTitle, state.Title)

	return state
}

func assertListDropsState(
	t *testing.T,
	out types.ContractOutput,
) []dropV1Models.DropStateModel {
	t.Helper()

	require.Len(t, out.States, 1)

	raw, err := json.Marshal(out.States[0].Object)
	if err != nil {
		t.Fatalf("Marshal list state: %v", err)
	}

	var states []dropV1Models.DropStateModel
	if err := json.Unmarshal(raw, &states); err != nil {
		t.Fatalf("Unmarshal list state: %v", err)
	}

	return states
}

func TestClient_GetDrop(t *testing.T) {
	c := setupClient(t)

	owner, ownerPriv := createWallet(t, c)
	c.SetPrivateKey(ownerPriv)

	deployedContract, err := c.DeployContract1(dropV1.DROP_CONTRACT_V1)
	if err != nil {
		t.Fatalf("DeployContract: %v", err)
	}

	deployLog, err := utils.UnmarshalLog[log.Log](deployedContract.Logs[0])
	if err != nil {
		t.Fatalf("UnmarshalLog (DeployContract.Logs[0]): %v", err)
	}

	dropAddress := deployLog.ContractAddress
	programAddress, _ := genKey(t, c)
	tokenAddress, _ := genKey(t, c)

	startAt := time.Now()
	expireAt := time.Now().Add(24 * time.Hour)

	input := dropV1Inputs.InputNewDrop{
		Address:              dropAddress,
		ProgramAddress:       programAddress,
		TokenAddress:         tokenAddress,
		Owner:                owner.PublicKey,
		Title:                "drop get test",
		Description:          "desc",
		ShortDescription:     "short",
		ImageURL:             "https://img.png",
		BannerURL:            "https://banner.png",
		Category:             "airdrop",
		SocialRequirements:   map[string]bool{"follow_x": true},
		PostLinks:            map[string]bool{"https://x.com/post/1": true},
		VerificationType:     dropV1Domain.VERIFICATION_TYPE_ORACLE,
		StartAt:              startAt,
		ExpireAt:             expireAt,
		RequestLimit:         100,
		ClaimAmount:          "10",
		ClaimIntervalSeconds: 3600,
	}

	_, err = c.NewDrop(input)
	if err != nil {
		t.Fatalf("NewDrop: %v", err)
	}

	t.Run("success", func(t *testing.T) {
		out := mustGetDrop(t, c, dropAddress)

		assertGetDropState(
			t,
			out,
			dropAddress,
			owner.PublicKey,
			programAddress,
			tokenAddress,
			"drop get test",
		)
	})

	t.Run("error when address is empty", func(t *testing.T) {
		_, err := c.GetDrop("")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "drop address must be set")
	})

	t.Run("error when address is invalid", func(t *testing.T) {
		_, err := c.GetDrop("invalid-address")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid drop address")
	})
}

func TestClient_ListDrops(t *testing.T) {
	c := setupClient(t)

	owner, ownerPriv := createWallet(t, c)
	c.SetPrivateKey(ownerPriv)

	createDrop := func(title string) string {
		deployedContract, err := c.DeployContract1(dropV1.DROP_CONTRACT_V1)
		if err != nil {
			t.Fatalf("DeployContract: %v", err)
		}

		deployLog, err := utils.UnmarshalLog[log.Log](deployedContract.Logs[0])
		if err != nil {
			t.Fatalf("UnmarshalLog (DeployContract.Logs[0]): %v", err)
		}

		dropAddress := deployLog.ContractAddress
		programAddress, _ := genKey(t, c)
		tokenAddress, _ := genKey(t, c)

		input := dropV1Inputs.InputNewDrop{
			Address:              dropAddress,
			ProgramAddress:       programAddress,
			TokenAddress:         tokenAddress,
			Owner:                owner.PublicKey,
			Title:                title,
			Description:          "desc",
			ShortDescription:     "short",
			ImageURL:             "https://img.png",
			BannerURL:            "https://banner.png",
			Category:             "airdrop",
			SocialRequirements:   map[string]bool{"follow_x": true},
			PostLinks:            map[string]bool{"https://x.com/post/1": true},
			VerificationType:     dropV1Domain.VERIFICATION_TYPE_ORACLE,
			StartAt:              time.Now(),
			ExpireAt:             time.Now().Add(24 * time.Hour),
			RequestLimit:         100,
			ClaimAmount:          "10",
			ClaimIntervalSeconds: 3600,
		}

		_, err = c.NewDrop(input)
		if err != nil {
			t.Fatalf("NewDrop: %v", err)
		}

		return dropAddress
	}

	address1 := createDrop("drop list test 1")
	address2 := createDrop("drop list test 2")

	t.Run("success with owner filter", func(t *testing.T) {
		out := mustListDrops(t, c, owner.PublicKey, 1, 10, true)
		states := assertListDropsState(t, out)

		require.NotEmpty(t, states)

		var found1, found2 bool
		for _, s := range states {
			if s.Address == address1 {
				found1 = true
			}
			if s.Address == address2 {
				found2 = true
			}
		}

		assert.True(t, found1, "address1 not found in ListDrops")
		assert.True(t, found2, "address2 not found in ListDrops")
	})

	t.Run("success with empty owner", func(t *testing.T) {
		out := mustListDrops(t, c, "", 1, 10, true)
		states := assertListDropsState(t, out)
		require.NotEmpty(t, states)
	})

	t.Run("error when owner is invalid", func(t *testing.T) {
		_, err := c.ListDrops("invalid-owner", 1, 10, true)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid owner address")
	})

	t.Run("error when page is invalid", func(t *testing.T) {
		_, err := c.ListDrops(owner.PublicKey, 0, 10, true)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "page must be greater than 0")
	})

	t.Run("error when limit is invalid", func(t *testing.T) {
		_, err := c.ListDrops(owner.PublicKey, 1, 0, true)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "limit must be greater than 0")
	})
}