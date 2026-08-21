// Token contract version
const String TOKEN_CONTRACT_V2 = "tokenV2";

// Methods (espelhe seus nomes reais do Go tokenV2.METHOD_*)
const String METHOD_ADD_TOKEN = "add_Token";
const String METHOD_MINT_TOKEN = "mint_Token";
const String METHOD_BURN_TOKEN = "burn_Token";
const String METHOD_TRANSFER_TOKEN = "transfer_Token";

//const String METHOD_CHANGE_ACCESS_MODE = "change_Access_Mode";

const String METHOD_ADD_ALLOWED_USERS = "add_Allowed_Users";
const String METHOD_REMOVE_ALLOWED_USERS = "remove_Allowed_Users";
const String METHOD_ADD_BLOCKED_USERS = "add_Blocked_Users";
const String METHOD_REMOVE_BLOCKED_USERS = "remove_Blocked_Users";

const String METHOD_REVOKE_FREEZE_AUTHORITY = "revoke_FreezeAuthority";
const String METHOD_REVOKE_MINT_AUTHORITY = "revoke_MintAuthority";
const String METHOD_REVOKE_UPDATE_AUTHORITY = "revoke_UpdateAuthority";

const String METHOD_UPDATE_METADATA = "update_Metadata";
const String METHOD_PAUSE_TOKEN = "pause_Token";
const String METHOD_UNPAUSE_TOKEN = "unpause_Token";
const String METHOD_UPDATE_FEE_TIERS = "update_FeeTiers";
const String METHOD_UPDATE_FEE_ADDRESS = "update_FeeAddress";
const String METHOD_UPDATE_GLB_FILE = "update_GLB_File";
const String METHOD_TRANSFERABLE_TOKEN = "transferable_Token";
const String METHOD_UNTRANSFERABLE_TOKEN = "untransferable_Token";
const String METHOD_FREEZE_WALLET = "freeze_wallet";
const String METHOD_UNFREEZE_WALLET = "unfreeze_wallet";

const String METHOD_GET_TOKEN = "get_Token";
const String METHOD_LIST_TOKENS = "list_Tokens";
const String METHOD_GET_TOKEN_BALANCE = "get_TokenBalance";
const String METHOD_GET_TOKEN_BALANCE_NFT = "get_TokenBalanceNFT";

const String METHOD_LIST_TOKEN_BALANCES = "list_TokenBalances";
const String METHOD_LIST_TRANSFERS = "list_Transfers";

// Token types (ajuste para seus values reais)
const String TOKEN_TYPE_FUNGIBLE = "FUNGIBLE";
const String TOKEN_TYPE_NON_FUNGIBLE = "NON_FUNGIBLE";
