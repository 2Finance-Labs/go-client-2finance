package e2e_test

import (
	"testing"
	"github.com/stretchr/testify/assert"
	"gitlab.com/2finance/2finance-network/blockchain/log"
	tokenV1Domain "gitlab.com/2finance/2finance-network/blockchain/contract/tokenV1/domain"
	"gitlab.com/2finance/2finance-network/blockchain/utils"
	dropV1Domain "gitlab.com/2finance/2finance-network/blockchain/contract/dropV1/domain"
	dropV1Models "gitlab.com/2finance/2finance-network/blockchain/contract/dropV1/models"
	dropV1Inputs "gitlab.com/2finance/2finance-network/blockchain/contract/dropV1/inputs"
	"gitlab.com/2finance/2finance-network/blockchain/contract/dropV1"
	"time"
)

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

	// coupon1, err := createBasicCoupon(t, c, owner.PublicKey, couponV1Domain.DISCOUNT_TYPE_PERCENTAGE)
	// if err != nil {
	// 	t.Fatalf("CreateBasicCoupon: %v", err)
	// }

	// coupon2, err := createBasicCoupon(t, c, owner.PublicKey, couponV1Domain.DISCOUNT_TYPE_FIXED)
	// if err != nil {
	// 	t.Fatalf("CreateBasicCoupon: %v", err)
	// }
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
	// Create Drop (agora passa faucetAddress e NÃO tem nonce)
	// --------------------------------------------------------------------
	programAddress, _ := genKey(t, c)
	tokenAddress, _ := genKey(t, c)
	title := "Airdrop E2E"
	description := "E2E description"
	shortDescription := "Short desc"
	image := "https://img.png"
	banner := "https://banner.png"
	category := "airdrop"
	socialRequirements := map[string]bool{"FOLLOW_X": true}
	postLinks := map[string]bool{"https://x.com/post/123": true}
	verificationType := "ORACLE"
	startAt := time.Now()
	expireAt := time.Now().Add(24 * time.Hour)
	requestLimit := 100
	claimAmount := "1000"
	claimIntervalSeconds := 3600

	inputDrop := dropV1Inputs.InputNewDrop{
		Address: address,
		ProgramAddress: programAddress,
		TokenAddress: tokenAddress,
		Owner: owner.PublicKey,
		Title: title,
		Description: description,
		ShortDescription: shortDescription,
		ImageURL: image,
		BannerURL: banner,
		Category: category,
		SocialRequirements: socialRequirements,
		PostLinks: postLinks,
		VerificationType: verificationType,
		StartAt: startAt,
		ExpireAt: expireAt,
		RequestLimit: requestLimit,
		ClaimAmount: claimAmount,
		ClaimIntervalSeconds: claimIntervalSeconds,
	}

	out, err := c.NewDrop(
		inputDrop,
	)
	if err != nil {
		t.Fatalf("NewDrop: %v", err)
	}

	unmarshalLogToken, err := utils.UnmarshalLog[log.Log](out.Logs[0])
	if err != nil {
		t.Fatalf("UnmarshalLog (AddToken.Logs[0]): %v", err)
	}
	assert.Equal(t, dropV1Domain.DROP_CREATED_LOG, unmarshalLogToken.LogType, "add-token log type mismatch")

	drop, err := utils.UnmarshalEvent[dropV1Domain.Drop](unmarshalLogToken.Event)
	if err != nil {
		t.Fatalf("UnmarshalEvent (AddToken.Logs[0]): %v", err)
	}
	
	assert.Equal(t, inputDrop.Address, drop.Address, "drop address empty")
	assert.Equal(t, inputDrop.ProgramAddress, drop.ProgramAddress, "drop program address mismatch")
	assert.Equal(t, inputDrop.TokenAddress, drop.TokenAddress, "drop token address mismatch")
	assert.Equal(t, inputDrop.Owner, drop.Owner, "drop owner mismatch")
	assert.Equal(t, inputDrop.Title, drop.Title, "drop title mismatch")
	assert.Equal(t, inputDrop.Description, drop.Description, "drop description mismatch")
	assert.Equal(t, inputDrop.ShortDescription, drop.ShortDescription, "drop short description mismatch")
	assert.Equal(t, inputDrop.ImageURL, drop.ImageURL, "drop image mismatch")
	assert.Equal(t, inputDrop.BannerURL, drop.BannerURL, "drop banner mismatch")
	assert.Equal(t, inputDrop.Category, drop.Category, "drop category mismatch")
	assert.Equal(t, inputDrop.SocialRequirements, drop.SocialRequirements, "drop social requirements mismatch")
	assert.Equal(t, inputDrop.PostLinks, drop.PostLinks, "drop post links mismatch")
	assert.Equal(t, inputDrop.VerificationType, drop.VerificationType, "drop verification type mismatch")
	assert.WithinDuration(t, inputDrop.StartAt, drop.StartAt, time.Second, "drop startAt mismatch")
	assert.WithinDuration(t, inputDrop.ExpireAt, drop.ExpireAt, time.Second, "drop expireAt mismatch")
	assert.Equal(t, inputDrop.RequestLimit, drop.RequestLimit, "drop request limit mismatch")
	assert.Equal(t, inputDrop.ClaimAmount, drop.ClaimAmount, "drop claim amount mismatch")
	assert.Equal(t, inputDrop.ClaimIntervalSeconds, drop.ClaimIntervalSeconds, "drop claim interval seconds mismatch")

	gotOut, err := c.GetDrop(drop.Address)
	if err != nil {
		t.Fatalf("GetDrop: %v", err)
	}

	var dropStateModel dropV1Models.DropStateModel
	err = utils.UnmarshalState[dropV1Models.DropStateModel](gotOut.States[0].Object, &dropStateModel)
	if err != nil {
		t.Fatalf("UnmarshalState (GetDrop.States[0]): %v", err)
	}
	assert.Equal(t, inputDrop.Address, dropStateModel.Address, "GetDrop address mismatch")
	assert.Equal(t, inputDrop.ProgramAddress, dropStateModel.ProgramAddress, "GetDrop program address mismatch")
	assert.Equal(t, inputDrop.TokenAddress, dropStateModel.TokenAddress, "GetDrop token address mismatch")
	assert.Equal(t, inputDrop.Owner, dropStateModel.Owner, "GetDrop owner mismatch")
	assert.Equal(t, inputDrop.Title, dropStateModel.Title, "GetDrop title mismatch")
	assert.Equal(t, inputDrop.Description, dropStateModel.Description, "GetDrop description mismatch")
	assert.Equal(t, inputDrop.ShortDescription, dropStateModel.ShortDescription, "GetDrop short description mismatch")
	assert.Equal(t, inputDrop.ImageURL, dropStateModel.ImageURL, "GetDrop image mismatch")
	assert.Equal(t, inputDrop.BannerURL, dropStateModel.BannerURL, "GetDrop banner mismatch")
	assert.Equal(t, inputDrop.Category, dropStateModel.Category, "GetDrop category mismatch")
	assert.Equal(t, inputDrop.SocialRequirements, dropStateModel.SocialRequirements, "GetDrop social requirements mismatch")
	assert.Equal(t, inputDrop.PostLinks, dropStateModel.PostLinks, "GetDrop post links mismatch")
	assert.Equal(t, inputDrop.VerificationType, dropStateModel.VerificationType, "GetDrop verification type mismatch")
	assert.WithinDuration(t, inputDrop.StartAt, dropStateModel.StartAt, time.Second*5, "GetDrop startAt mismatch")
	assert.WithinDuration(t, inputDrop.ExpireAt, dropStateModel.ExpireAt, time.Second*5, "GetDrop expireAt mismatch")
	assert.Equal(t, inputDrop.RequestLimit, dropStateModel.RequestLimit, "GetDrop request limit mismatch")
	assert.Equal(t, inputDrop.ClaimAmount, dropStateModel.ClaimAmount, "GetDrop claim amount mismatch")
	assert.Equal(t, inputDrop.ClaimIntervalSeconds, dropStateModel.ClaimIntervalSeconds, "GetDrop claim interval seconds mismatch")

	startAt = startAt.Add(10 * time.Hour)
	expireAt = expireAt.Add(50 * time.Hour)
	inputUpdateDrop := dropV1Inputs.InputUpdateDropMetadata{
		Address: address,
		ProgramAddress: programAddress,
		TokenAddress: tok.Address,
		Title: "Airdrop E2E (UPDATED)",
		Description: "Updated description",
		ShortDescription: "Updated short description",
		ImageURL: "https://img-updated.png",
		BannerURL: "https://banner-updated.png",
		Category: "airdrop",
		SocialRequirements: map[string]bool{"LIKE_X": true},
		PostLinks: map[string]bool{"https://x.com/post/456": true},
		StartAt: startAt,
		ExpireAt: expireAt,
		VerificationType: "ORACLE",
		RequestLimit: 10,
		ClaimAmount: "500",
		ClaimIntervalSeconds: 1800,
	}
	// UPDATE DROP METADATA
	outMeta, err := c.UpdateDropMetadata(
		inputUpdateDrop,
	)
	if err != nil {
		t.Fatalf("UpdateDropMetadata: %v", err)
	}

	unmarshalLogToken, err = utils.UnmarshalLog[log.Log](outMeta.Logs[0])
	if err != nil {
		t.Fatalf("UnmarshalLog (UpdateDropMetadata.Logs[0]): %v", err)
	}
	assert.Equal(t, dropV1Domain.DROP_METADATA_UPDATED_LOG, unmarshalLogToken.LogType, "update-drop-metadata log type mismatch")

	dropUpdated, err := utils.UnmarshalEvent[dropV1Domain.Drop](unmarshalLogToken.Event)
	if err != nil {
		t.Fatalf("UnmarshalEvent (UpdateDropMetadata.Logs[0]): %v", err)
	}
	
	assert.Equal(t, inputUpdateDrop.Address, dropUpdated.Address, "updated drop address empty")
	assert.Equal(t, inputUpdateDrop.ProgramAddress, dropUpdated.ProgramAddress, "updated drop program address mismatch")
	assert.Equal(t, inputUpdateDrop.TokenAddress, dropUpdated.TokenAddress, "updated drop token address mismatch")
	assert.Equal(t, inputUpdateDrop.Title, dropUpdated.Title, "updated drop title mismatch")
	assert.Equal(t, inputUpdateDrop.Description, dropUpdated.Description, "updated drop description mismatch")
	assert.Equal(t, inputUpdateDrop.ShortDescription, dropUpdated.ShortDescription, "updated drop short description mismatch")
	assert.Equal(t, inputUpdateDrop.ImageURL, dropUpdated.ImageURL, "updated drop image mismatch")
	assert.Equal(t, inputUpdateDrop.BannerURL, dropUpdated.BannerURL, "updated drop banner mismatch")
	assert.Equal(t, inputUpdateDrop.Category, dropUpdated.Category, "updated drop category mismatch")
	assert.Equal(t, inputUpdateDrop.SocialRequirements, dropUpdated.SocialRequirements, "updated drop social requirements mismatch")
	assert.Equal(t, inputUpdateDrop.PostLinks, dropUpdated.PostLinks, "updated drop post links mismatch")
	assert.Equal(t, inputUpdateDrop.VerificationType, dropUpdated.VerificationType, "updated drop verification type mismatch")
	assert.WithinDuration(t, inputUpdateDrop.StartAt, dropUpdated.StartAt, time.Second*10, "updated drop startAt mismatch")
	assert.WithinDuration(t, inputUpdateDrop.ExpireAt, dropUpdated.ExpireAt, time.Second*10, "updated drop expireAt mismatch")
	assert.Equal(t, inputUpdateDrop.RequestLimit, dropUpdated.RequestLimit, "updated drop request limit mismatch")
	assert.Equal(t, inputUpdateDrop.ClaimAmount, dropUpdated.ClaimAmount, "updated drop claim amount mismatch")
	assert.Equal(t, inputUpdateDrop.ClaimIntervalSeconds, dropUpdated.ClaimIntervalSeconds, "updated drop claim interval seconds mismatch")

	gotUpdatedDrop, err := c.GetDrop(address)
	if err != nil {
		t.Fatalf("GetDrop: %v", err)
	}

	var dropStateModelUpdated dropV1Models.DropStateModel
	err = utils.UnmarshalState[dropV1Models.DropStateModel](gotUpdatedDrop.States[0].Object, &dropStateModelUpdated)
	if err != nil {
		t.Fatalf("UnmarshalState (GetDrop.States[0]): %v", err)
	}

	resultSocialRequirements := map[string]bool{"FOLLOW_X": true, "LIKE_X": true}
	resultPostLinks := map[string]bool{"https://x.com/post/123": true, "https://x.com/post/456": true}
	assert.Equal(t, inputUpdateDrop.Title, dropStateModelUpdated.Title, "GetDrop title mismatch")
	assert.Equal(t, inputUpdateDrop.Description, dropStateModelUpdated.Description, "GetDrop description mismatch")
	assert.Equal(t, inputUpdateDrop.ShortDescription, dropStateModelUpdated.ShortDescription, "GetDrop short description mismatch")
	assert.Equal(t, inputUpdateDrop.ImageURL, dropStateModelUpdated.ImageURL, "GetDrop image mismatch")
	assert.Equal(t, inputUpdateDrop.BannerURL, dropStateModelUpdated.BannerURL, "GetDrop banner mismatch")
	assert.Equal(t, inputUpdateDrop.Category, dropStateModelUpdated.Category, "GetDrop category mismatch")
	assert.Equal(t, resultSocialRequirements, dropStateModelUpdated.SocialRequirements, "GetDrop social requirements mismatch")
	assert.Equal(t, resultPostLinks, dropStateModelUpdated.PostLinks, "GetDrop post links mismatch")
	assert.Equal(t, inputUpdateDrop.VerificationType, dropStateModelUpdated.VerificationType, "GetDrop verification type mismatch")
	assert.WithinDuration(t, startAt, dropStateModelUpdated.StartAt, time.Second * 10, "GetDrop startAt mismatch")
	assert.WithinDuration(t, expireAt, dropStateModelUpdated.ExpireAt, time.Second * 10, "GetDrop expireAt mismatch")
	assert.Equal(t, inputUpdateDrop.RequestLimit, dropStateModelUpdated.RequestLimit, "GetDrop request limit mismatch")
	assert.Equal(t, inputUpdateDrop.ClaimAmount, dropStateModelUpdated.ClaimAmount, "GetDrop claim amount mismatch")
	assert.Equal(t, inputUpdateDrop.ClaimIntervalSeconds, dropStateModelUpdated.ClaimIntervalSeconds, "GetDrop claim interval seconds mismatch")

	// ALLOW / DISALLOW ORACLES
	oracle1, oracle1Priv := genKey(t, c)
	oracle2, _ := genKey(t, c)
	oracle3, _ := genKey(t, c)
	allowOraclesOut, err := c.AllowOracles(
		drop.Address,
		map[string]bool{
			oracle1: true,
			oracle2: true,
			oracle3: true,
		},
	)
	if err != nil {
		t.Fatalf("AllowOracles: %v", err)
	}
	
	unmarshalLogAllowOracles, err := utils.UnmarshalLog[log.Log](allowOraclesOut.Logs[0])
	if err != nil {
		t.Fatalf("UnmarshalLog (AllowOracles.Logs[0]): %v", err)
	}
	assert.Equal(t, dropV1Domain.DROP_ORACLES_ALLOWED_LOG, unmarshalLogAllowOracles.LogType, "allow-oracles log type mismatch")

	allowedOracles, err := utils.UnmarshalEvent[dropV1Domain.Drop](unmarshalLogAllowOracles.Event)
	if err != nil {
		t.Fatalf("UnmarshalEvent (AllowOracles.Event): %v", err)
	}

	assert.Equal(t, drop.Address, allowedOracles.Address, "allowed oracles drop address mismatch")
	assert.Equal(t, map[string]bool{oracle1: true, oracle2: true, oracle3: true}, allowedOracles.AllowedOracles, "allowed oracles mismatch")

	disallowOraclesOut, err := c.DisallowOracles(
		drop.Address,
		map[string]bool{
			oracle2: true,
		},
	)
	if err != nil {
		t.Fatalf("DisallowOracles: %v", err)
	}

	unmarshalLogDisallowOracles, err := utils.UnmarshalLog[log.Log](disallowOraclesOut.Logs[0])
	if err != nil {
		t.Fatalf("UnmarshalLog (DisallowOracles.Logs[0]): %v", err)
	}
	assert.Equal(t, dropV1Domain.DROP_ORACLES_DISALLOWED_LOG, unmarshalLogDisallowOracles.LogType, "disallow-oracles log type mismatch")

	disallowedOracles, err := utils.UnmarshalEvent[dropV1Domain.Drop](unmarshalLogDisallowOracles.Event)
	if err != nil {
		t.Fatalf("UnmarshalEvent (DisallowOracles.Event): %v", err)
	}

	assert.Equal(t, drop.Address, disallowedOracles.Address, "disallowed oracles drop address mismatch")
	assert.Equal(t, map[string]bool{oracle2: true}, disallowedOracles.AllowedOracles, "disallowed oracles mismatch")

	gotUpdatedDrop, err = c.GetDrop(address)
	if err != nil {
		t.Fatalf("GetDrop: %v", err)
	}
	var dropStateModelOracles dropV1Models.DropStateModel
	err = utils.UnmarshalState[dropV1Models.DropStateModel](gotUpdatedDrop.States[0].Object, &dropStateModelOracles)
	if err != nil {
		t.Fatalf("UnmarshalState (GetDrop.States[0]): %v", err)
	}
	resultAllowedOracles := map[string]bool{oracle1: true, oracle3: true}
	assert.Equal(t, resultAllowedOracles, dropStateModelOracles.AllowedOracles, "GetDrop allowed oracles mismatch")

	
	// PAUSE / UNPAUSE
	outPause, err := c.PauseDrop(address)
	if err != nil {
		t.Fatalf("PauseDrop: %v", err)
	}
	unmarshalLogPause, err := utils.UnmarshalLog[log.Log](outPause.Logs[0])
	if err != nil {
		t.Fatalf("UnmarshalLog (PauseDrop.Logs[0]): %v", err)
	}
	assert.Equal(t, dropV1Domain.DROP_PAUSED_LOG, unmarshalLogPause.LogType, "pause-drop log type mismatch")

	pausedDrop, err := utils.UnmarshalEvent[dropV1Domain.Drop](unmarshalLogPause.Event)
	if err != nil {
		t.Fatalf("UnmarshalEvent (PauseDrop.Event): %v", err)
	}
	assert.Equal(t, drop.Address, pausedDrop.Address, "paused drop address mismatch")
	assert.Equal(t, true, pausedDrop.Paused, "drop paused state mismatch")
	
	outUnpause, err := c.UnpauseDrop(address)
	if err != nil {
		t.Fatalf("UnpauseDrop: %v", err)
	}
	
	unmarshalLogUnpause, err := utils.UnmarshalLog[log.Log](outUnpause.Logs[0])
	if err != nil {
		t.Fatalf("UnmarshalLog (UnpauseDrop.Logs[0]): %v", err)
	}
	assert.Equal(t, dropV1Domain.DROP_UNPAUSED_LOG, unmarshalLogUnpause.LogType, "unpause-drop log type mismatch")

	unpausedDrop, err := utils.UnmarshalEvent[dropV1Domain.Drop](unmarshalLogUnpause.Event)
	if err != nil {
		t.Fatalf("UnmarshalEvent (UnpauseDrop.Event): %v", err)
	}
	assert.Equal(t, drop.Address, unpausedDrop.Address, "unpaused drop address mismatch")
	assert.Equal(t, false, unpausedDrop.Paused, "drop paused state mismatch")

	gotUpdatedDrop, err = c.GetDrop(address)
	if err != nil {
		t.Fatalf("GetDrop: %v", err)
	}

	var dropStateModelPause dropV1Models.DropStateModel
	err = utils.UnmarshalState[dropV1Models.DropStateModel](gotUpdatedDrop.States[0].Object, &dropStateModelPause)
	if err != nil {
		t.Fatalf("UnmarshalState (GetDrop.States[0]): %v", err)
	}
	assert.Equal(t, false, dropStateModelPause.Paused, "GetDrop paused state mismatch")
	

	// ATTEST ELIGIBILITY (ORACLE)
	eligible1, _ := genKey(t, c)
	c.SetPrivateKey(oracle1Priv)
	outAttestParticipant, err := c.AttestParticipantEligibility(
		drop.Address,
		eligible1,
		true,
	)
	if err != nil {
		t.Fatalf("AttestParticipantEligibility: %v", err)
	}
	unmarshalLogAttest, err := utils.UnmarshalLog[log.Log](outAttestParticipant.Logs[0])
	if err != nil {
		t.Fatalf("UnmarshalLog (AttestParticipantEligibility.Logs[0]): %v", err)
	}
	assert.Equal(t, dropV1Domain.DROP_ATTESTED_PARTICIPANT_ELIGIBILITY_LOG, unmarshalLogAttest.LogType, "attest-participant log type mismatch")

	attestedParticipant, err := utils.UnmarshalEvent[dropV1Domain.EligibilityAttested](unmarshalLogAttest.Event)
	if err != nil {
		t.Fatalf("UnmarshalEvent (AttestParticipantEligibility.Event): %v", err)
	}
	assert.Equal(t, drop.Address, attestedParticipant.DropAddress, "attested participant drop address mismatch")
	assert.Equal(t, eligible1, attestedParticipant.Wallet, "attested participant address mismatch")
	assert.Equal(t, true, attestedParticipant.Eligible, "attested participant eligibility mismatch")
	assert.Equal(t, dropV1Domain.VERIFICATION_TYPE_ORACLE, attestedParticipant.VerificationType, "attested participant verification type mismatch")

	// ATTEST ELIGIBILITY OWNER
	c.SetPrivateKey(ownerPriv)
	eligible2, _ := genKey(t, c)
	outAttestParticipantOwner, err := c.AttestParticipantEligibility(
		drop.Address,
		eligible2,
		true,
	)
	if err != nil {
		t.Fatalf("AttestParticipantEligibility: %v", err)
	}
	unmarshalLogAttestOwner, err := utils.UnmarshalLog[log.Log](outAttestParticipantOwner.Logs[0])
	if err != nil {
		t.Fatalf("UnmarshalLog (AttestParticipantEligibility.Logs[0]): %v", err)
	}
	assert.Equal(t, dropV1Domain.DROP_ATTESTED_PARTICIPANT_ELIGIBILITY_LOG, unmarshalLogAttestOwner.LogType, "attest-participant log type mismatch")

	attestedParticipantOwner, err := utils.UnmarshalEvent[dropV1Domain.EligibilityAttested](unmarshalLogAttestOwner.Event)
	if err != nil {
		t.Fatalf("UnmarshalEvent (AttestParticipantEligibility.Event): %v", err)
	}
	assert.Equal(t, drop.Address, attestedParticipantOwner.DropAddress, "attested participant drop address mismatch")
	assert.Equal(t, eligible2, attestedParticipantOwner.Wallet, "attested participant address mismatch")
	assert.Equal(t, true, attestedParticipantOwner.Eligible, "attested participant eligibility mismatch")
	assert.Equal(t, dropV1Domain.VERIFICATION_TYPE_ORACLE, attestedParticipantOwner.VerificationType, "attested participant verification type mismatch")
	
	// // Deposit Funds
	// addressDrop := drop.Address
	// programAddressDrop := drop.ProgramAddress
	// tokenAddressDrop := drop.TokenAddress
	// amountDeposit := "10"
	// outDepositFunds, err := c.DepositDrop(
	// 	addressDrop,
	// 	programAddressDrop,
	// 	tokenAddressDrop,
	// 	amountDeposit,
	// 	[]string{},
	// )
	// if err != nil {
	// 	t.Fatalf("DepositDrop: %v", err)
	// }
	// unmarshalLogDeposit, err := utils.UnmarshalLog[log.Log](outDepositFunds.Logs[0])
	// if err != nil {
	// 	t.Fatalf("UnmarshalLog (DepositDrop.Logs[0]): %v", err)
	// }
	// assert.Equal(t, dropV1Domain.DROP_DEPOSITED_LOG, unmarshalLogDeposit.LogType, "deposit-drop-funds log type mismatch")

	// depositedFunds, err := utils.UnmarshalEvent[dropV1Domain.DepositFunds](unmarshalLogDeposit.Event)
	// if err != nil {
	// 	t.Fatalf("UnmarshalEvent (DepositDrop.Event): %v", err)
	// }
	// assert.Equal(t, drop.Address, depositedFunds.Address, "deposited funds drop address mismatch")
	// assert.Equal(t, drop.ProgramAddress, depositedFunds.ProgramAddress, "deposited funds program address mismatch")
	// assert.Equal(t, drop.TokenAddress, depositedFunds.TokenAddress, "deposited funds token address mismatch")
	// assert.Equal(t, owner, depositedFunds.SenderAddress, "deposited funds sender mismatch")
	// assert.Equal(t, amountDeposit, depositedFunds.Amount, "deposited funds amount mismatch")
	// // --------------------------------------------------------------------
	// // Deposit funds (owner)
	// // --------------------------------------------------------------------
	// if _, err := c.DepositAirdrop(ad.Address, amt(200, dec), tokenType, ""); err != nil {
	// 	t.Fatalf("DepositAirdrop: %v", err)
	// }

	// // --------------------------------------------------------------------
	// // Manual attest (owner)
	// // --------------------------------------------------------------------
	// if _, err := c.ManuallyAttestParticipantEligibility(ad.Address, user.PublicKey, true); err != nil {
	// 	t.Fatalf("ManuallyAttestParticipantEligibility: %v", err)
	// }

	// // --------------------------------------------------------------------
	// // Wait start
	// // --------------------------------------------------------------------
	// waitUntil(t, 15*time.Second, func() bool {
	// 	return time.Now().After(start)
	// })

	// // --------------------------------------------------------------------
	// // Claim (user)
	// // --------------------------------------------------------------------
	// c.SetPrivateKey(userPriv)

	// if _, err := c.ClaimAirdrop(ad.Address, tok.TokenType); err != nil {
	// 	t.Fatalf("ClaimAirdrop: %v", err)
	// }

	// // Double-claim deve falhar
	// if _, err := c.ClaimAirdrop(ad.Address, tok.TokenType); err == nil {
	// 	t.Fatalf("ClaimAirdrop: expected error on double-claim, got nil")
	// }

	// // --------------------------------------------------------------------
	// // Withdraw remaining funds (owner)
	// // --------------------------------------------------------------------
	// time.Sleep(2 * time.Second)

	// c.SetPrivateKey(ownerPriv)

	// if _, err := c.WithdrawAirdropFunds(ad.Address, amt(50, dec), tokenType, ""); err != nil {
	// 	t.Fatalf("WithdrawAirdropFunds: %v", err)
	// }

	// // --------------------------------------------------------------------
	// // Update metadata (owner) - cobre METHOD_UPDATE_AIRDROP_METADATA
	// // --------------------------------------------------------------------
	// c.SetPrivateKey(ownerPriv)

	// newTitle := "Airdrop E2E (UPDATED)"
	// newDesc := "Updated description"
	// newShort := "Updated short"
	// newImg := "https://img-updated.png"
	// newBanner := "https://banner-updated.png"
	// newCategory := "airdrop"

	// newReq := map[string]bool{"FOLLOW_X": true, "LIKE_X": true}
	// newLinks := []string{"https://x.com/post/updated"}

	// newVerificationType := "ORACLE"

	// newManualReviewRequired := true
	// newVerifier := verifier.PublicKey

	// outMeta, err := c.UpdateAirdropMetadata(
	// 	ad.Address,
	// 	newTitle,
	// 	newDesc,
	// 	newShort,
	// 	newImg,
	// 	newBanner,
	// 	newCategory,
	// 	newReq,
	// 	newLinks,
	// 	newVerificationType,
	// 	newVerifier,
	// 	newManualReviewRequired,
	// )
	// if err != nil {
	// 	t.Fatalf("UpdateAirdropMetadata: %v", err)
	// }

	// var adUpdated airdropModels.AirdropStateModel
	// unmarshalState(t, outMeta.States[0].Object, &adUpdated)

	// // --------------------------------------------------------------------
	// // GET pós-update - garante consistência de leitura
	// // --------------------------------------------------------------------
	// gotOut2, err := c.GetAirdrop(ad.Address)
	// if err != nil {
	// 	t.Fatalf("GetAirdrop(post-update): %v", err)
	// }

	// var adGet2 airdropModels.AirdropStateModel
	// unmarshalState(t, gotOut2.States[0].Object, &adGet2)

	// if adGet2.Title != newTitle {
	// 	t.Fatalf("GetAirdrop(post-update) mismatch: title=%q want=%q", adGet2.Title, newTitle)
	// }
	// if adGet2.ShortDescription != newShort {
	// 	t.Fatalf("GetAirdrop(post-update) mismatch: short=%q want=%q", adGet2.ShortDescription, newShort)
	// }
	// if adGet2.VerificationType != newVerificationType {
	// 	t.Fatalf("GetAirdrop(post-update) mismatch: verification_type=%q want=%q", adGet2.VerificationType, newVerificationType)
	// }

	// // --------------------------------------------------------------------
	// // Allow oracle (owner)
	// // --------------------------------------------------------------------
	// oracle, oraclePriv := createWallet(t, c)
	// userOracle, userOraclePriv := createWallet(t, c)

	// c.SetPrivateKey(ownerPriv)
	// if _, err := c.AllowOracles(ad.Address, map[string]bool{
	// 	oracle.PublicKey: true,
	// }); err != nil {
	// 	t.Fatalf("AllowOracles: %v", err)
	// }

	// if _, err := c.AllowUsers(tok.Address, map[string]bool{
	// 	userOracle.PublicKey: true,
	// }); err != nil {
	// 	t.Fatalf("AllowUsers(token userOracle): %v", err)
	// }

	// // --------------------------------------------------------------------
	// // Attest eligibility (oracle)
	// // --------------------------------------------------------------------
	// c.SetPrivateKey(oraclePriv)
	// if _, err := c.AttestParticipantEligibility(ad.Address, userOracle.PublicKey, true); err != nil {
	// 	t.Fatalf("AttestParticipantEligibility: %v", err)
	// }

	// // --------------------------------------------------------------------
	// // Claim (user)
	// // --------------------------------------------------------------------
	// c.SetPrivateKey(userOraclePriv)

	// if _, err := c.ClaimAirdrop(ad.Address, tok.TokenType); err != nil {
	// 	t.Fatalf("ClaimAirdrop: %v", err)
	// }

	// // Double-claim deve falhar
	// if _, err := c.ClaimAirdrop(ad.Address, tok.TokenType); err == nil {
	// 	t.Fatalf("ClaimAirdrop: expected error on double-claim, got nil")
	// }

	// // --------------------------------------------------------------------
	// // Withdraw remaining funds (owner)
	// // --------------------------------------------------------------------
	// time.Sleep(2 * time.Second)

	// c.SetPrivateKey(ownerPriv)

	// if _, err := c.WithdrawAirdropFunds(ad.Address, amt(50, dec), tokenType, ""); err != nil {
	// 	t.Fatalf("WithdrawAirdropFunds: %v", err)
	// }
}
