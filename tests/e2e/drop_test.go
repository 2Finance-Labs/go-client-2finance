package e2e_test

import (
	"testing"
	"github.com/stretchr/testify/assert"
	"gitlab.com/2finance/2finance-network/blockchain/log"
	tokenV1Domain "gitlab.com/2finance/2finance-network/blockchain/contract/tokenV1/domain"
	"gitlab.com/2finance/2finance-network/blockchain/utils"
	dropV1Domain "gitlab.com/2finance/2finance-network/blockchain/contract/dropV1/domain"
	dropV1Models "gitlab.com/2finance/2finance-network/blockchain/contract/dropV1/models"
	"gitlab.com/2finance/2finance-network/blockchain/contract/dropV1"
	"time"
)

func TestDropFlow(t *testing.T) {
	c := setupClient(t)

	owner, ownerPriv := createWallet(t, c)
	// user, userPriv := createWallet(t, c)
	// verifier, _ := createWallet(t, c)

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


	tokenType = tokenV1Domain.FUNGIBLE
	stablecoin = true
	tok2 := createBasicToken(t, c, owner.PublicKey, dec, true, tokenType, stablecoin)
	amount = "10000"
	if _, err := c.MintToken(tok2.Address, owner.PublicKey, amount); err != nil {
		t.Fatalf("MintToken: %v", err)
	}

	// --------------------------------------------------------------------
	// Token setup NONFUNGIBLE (NFT)
	// --------------------------------------------------------------------
	nonFungibleTokenType := tokenV1Domain.NON_FUNGIBLE
	nftAmount := "50"
	nft := createBasicToken(t, c, owner.PublicKey, 0, false, nonFungibleTokenType, false)
	if _, err := c.MintToken(nft.Address, owner.PublicKey, nftAmount); err != nil {
		t.Fatalf("MintToken (NFT): %v", err)
	}

	nftAmount2 := "150"
	nft2 := createBasicToken(t, c, owner.PublicKey, 0, false, nonFungibleTokenType, false)
	if _, err := c.MintToken(nft2.Address, owner.PublicKey, nftAmount2); err != nil {
		t.Fatalf("MintToken (NFT 2): %v", err)
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
	title := "Airdrop E2E"
	description := "E2E description"
	shortDescription := "Short desc"
	image := "https://img.png"
	banner := "https://banner.png"
	category := "airdrop"
	socialRequirements := map[string]bool{"FOLLOW_X": true}
	postLinks := map[string]bool{"https://x.com/post/123": true}
	verificationType := "MANUAL"
	manualReviewRequired := true
	out, err := c.NewDrop(
		address,
		owner.PublicKey,
		title,
		description,
		shortDescription,
		image,
		banner,
		category,
		socialRequirements,
		postLinks,
		verificationType,
		manualReviewRequired,
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
	
	assert.Equal(t, address, drop.Address, "drop address empty")
	assert.Equal(t, title, drop.Title, "drop title mismatch")
	assert.Equal(t, description, drop.Description, "drop description mismatch")
	assert.Equal(t, shortDescription, drop.ShortDescription, "drop short description mismatch")
	assert.Equal(t, image, drop.ImageURL, "drop image mismatch")
	assert.Equal(t, banner, drop.BannerURL, "drop banner mismatch")
	assert.Equal(t, category, drop.Category, "drop category mismatch")
	assert.Equal(t, socialRequirements, drop.SocialRequirements, "drop social requirements mismatch")
	assert.Equal(t, postLinks, drop.PostLinks, "drop post links mismatch")
	assert.Equal(t, verificationType, drop.VerificationType, "drop verification type mismatch")
	assert.Equal(t, manualReviewRequired, drop.ManualReviewRequired, "drop manual review required mismatch")

	gotOut, err := c.GetDrop(drop.Address)
	if err != nil {
		t.Fatalf("GetDrop: %v", err)
	}

	var dropStateModel dropV1Models.DropStateModel
	err = utils.UnmarshalState[dropV1Models.DropStateModel](gotOut.States[0].Object, &dropStateModel)
	if err != nil {
		t.Fatalf("UnmarshalState (GetDrop.States[0]): %v", err)
	}
	assert.Equal(t, address, dropStateModel.Address, "GetDrop address mismatch")
	assert.Equal(t, title, dropStateModel.Title, "GetDrop title mismatch")
	assert.Equal(t, description, dropStateModel.Description, "GetDrop description mismatch")
	assert.Equal(t, shortDescription, dropStateModel.ShortDescription, "GetDrop short description mismatch")
	assert.Equal(t, image, dropStateModel.ImageURL, "GetDrop image mismatch")
	assert.Equal(t, banner, dropStateModel.BannerURL, "GetDrop banner mismatch")
	assert.Equal(t, category, dropStateModel.Category, "GetDrop category mismatch")
	assert.Equal(t, socialRequirements, dropStateModel.SocialRequirements, "GetDrop social requirements mismatch")
	assert.Equal(t, postLinks, dropStateModel.PostLinks, "GetDrop post links mismatch")
	assert.Equal(t, verificationType, dropStateModel.VerificationType, "GetDrop verification type mismatch")
	assert.Equal(t, manualReviewRequired, dropStateModel.ManualReviewRequired, "GetDrop manual review required mismatch")

	updatedTitle := "Airdrop E2E (UPDATED)"
	updatedDescription := "Updated description"
	updatedShortDescription := "Updated short description"
	updatedImage := "https://img-updated.png"
	updatedBanner := "https://banner-updated.png"
	updatedCategory := "airdrop"
	updatedSocialRequirements := map[string]bool{"LIKE_X": true}
	updatedPostLinks := map[string]bool{"https://x.com/post/456": true}
	updatedVerificationType := "ORACLE"
	updatedManualReviewRequired := false

	outMeta, err := c.UpdateDropMetadata(
		address,
		updatedTitle,
		updatedDescription,
		updatedShortDescription,
		updatedImage,
		updatedBanner,
		updatedCategory,
		updatedSocialRequirements,
		updatedPostLinks,
		updatedVerificationType,
		updatedManualReviewRequired,
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
	
	assert.Equal(t, address, dropUpdated.Address, "updated drop address empty")
	assert.Equal(t, updatedTitle, dropUpdated.Title, "updated drop title mismatch")
	assert.Equal(t, updatedDescription, dropUpdated.Description, "updated drop description mismatch")
	assert.Equal(t, updatedShortDescription, dropUpdated.ShortDescription, "updated drop short description mismatch")
	assert.Equal(t, updatedImage, dropUpdated.ImageURL, "updated drop image mismatch")
	assert.Equal(t, updatedBanner, dropUpdated.BannerURL, "updated drop banner mismatch")
	assert.Equal(t, updatedCategory, dropUpdated.Category, "updated drop category mismatch")
	assert.Equal(t, updatedSocialRequirements, dropUpdated.SocialRequirements, "updated drop social requirements mismatch")
	assert.Equal(t, updatedPostLinks, dropUpdated.PostLinks, "updated drop post links mismatch")
	assert.Equal(t, updatedVerificationType, dropUpdated.VerificationType, "updated drop verification type mismatch")
	assert.Equal(t, updatedManualReviewRequired, dropUpdated.ManualReviewRequired, "updated drop manual review required mismatch")

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
	assert.Equal(t, updatedTitle, dropStateModelUpdated.Title, "GetDrop title mismatch")
	assert.Equal(t, updatedDescription, dropStateModelUpdated.Description, "GetDrop description mismatch")
	assert.Equal(t, updatedShortDescription, dropStateModelUpdated.ShortDescription, "GetDrop short description mismatch")
	assert.Equal(t, updatedImage, dropStateModelUpdated.ImageURL, "GetDrop image mismatch")
	assert.Equal(t, updatedBanner, dropStateModelUpdated.BannerURL, "GetDrop banner mismatch")
	assert.Equal(t, updatedCategory, dropStateModelUpdated.Category, "GetDrop category mismatch")
	assert.Equal(t, resultSocialRequirements, dropStateModelUpdated.SocialRequirements, "GetDrop social requirements mismatch")
	assert.Equal(t, resultPostLinks, dropStateModelUpdated.PostLinks, "GetDrop post links mismatch")
	assert.Equal(t, updatedVerificationType, dropStateModelUpdated.VerificationType, "GetDrop verification type mismatch")
	assert.Equal(t, updatedManualReviewRequired, dropStateModelUpdated.ManualReviewRequired, "GetDrop manual review required mismatch")

	startAt := time.Now()
	expireAt := time.Now().Add(24 * time.Hour)
	
	programAddresses := map[string]string{"program1": "program1", "token2": "program2"}
	outUpdatedDropSettings, err := c.UpdateDropSettings(
		address,
		programAddresses,
		startAt,
		expireAt,
		100,
		map[string]string{"token1": "100", "token2": "200"},
		3600,
	)
	if err != nil {
		t.Fatalf("UpdateDropSettings: %v", err)
	}

	unmarshalSettingsLog, err := utils.UnmarshalLog[log.Log](outUpdatedDropSettings.Logs[0])
	if err != nil {
		t.Fatalf("UnmarshalLog (UpdateDropSettings.Logs[0]): %v", err)
	}
	assert.Equal(t, dropV1Domain.DROP_SETTINGS_UPDATED_LOG, unmarshalSettingsLog.LogType, "update-drop-settings log type mismatch")

	dropSettingsUpdated, err := utils.UnmarshalEvent[dropV1Domain.Settings](unmarshalSettingsLog.Event)
	if err != nil {
		t.Fatalf("UnmarshalEvent (UpdateDropSettings.Logs[0]): %v", err)
	}

	assert.Equal(t, address, dropSettingsUpdated.Address, "updated drop settings address empty")
	assert.Equal(t, programAddresses, dropSettingsUpdated.ProgramAddresses, "updated drop settings program addresses mismatch")
	assert.Equal(t, startAt.Format(time.RFC3339), dropSettingsUpdated.StartAt.Format(time.RFC3339), "updated drop settings startAt mismatch")
	assert.Equal(t, expireAt.Format(time.RFC3339), dropSettingsUpdated.ExpireAt.Format(time.RFC3339), "updated drop settings expireAt mismatch")

	
	// --------------------------------------------------------------------
	// GET Drop (reader) - cobre METHOD_GET_DROP
	// --------------------------------------------------------------------
	
	

	// // --------------------------------------------------------------------
	// // LIST Airdrops (reader) - cobre METHOD_LIST_AIRDROPS
	// // --------------------------------------------------------------------
	// listOut, err := c.ListAirdrops(owner.PublicKey, 1, 50, false)
	// if err != nil {
	// 	t.Fatalf("ListAirdrops: %v", err)
	// }

	// var list []airdropModels.AirdropStateModel
	// unmarshalState(t, listOut.States[0].Object, &list)

	// found := false
	// for _, it := range list {
	// 	if it.Address == ad.Address {
	// 		found = true
	// 		break
	// 	}
	// }
	// if !found {
	// 	t.Fatalf("ListAirdrops: created airdrop %s not found in list", ad.Address)
	// }

	// // --------------------------------------------------------------------
	// // Allowlist token: owner + faucet + user
	// // --------------------------------------------------------------------
	// if _, err := c.AllowUsers(tok.Address, map[string]bool{
	// 	owner.PublicKey:  true,
	// 	ad.FaucetAddress: true,
	// 	user.PublicKey:   true,
	// }); err != nil {
	// 	t.Fatalf("AllowUsers(token): %v", err)
	// }

	// // --------------------------------------------------------------------
	// // Pause / Unpause (owner)
	// // --------------------------------------------------------------------
	// c.SetPrivateKey(ownerPriv)
	// if _, err := c.PauseAirdrop(ad.Address); err != nil {
	// 	t.Fatalf("PauseAirdrop: %v", err)
	// }
	// if _, err := c.UnpauseAirdrop(ad.Address); err != nil {
	// 	t.Fatalf("UnpauseAirdrop: %v", err)
	// }

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
