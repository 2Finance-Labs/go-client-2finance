package wallet_manager

import (
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"gitlab.com/2finance/2finance-network/blockchain/encryption/keys"
	"gitlab.com/2finance/2finance-network/blockchain/transaction"
	"gitlab.com/2finance/2finance-network/blockchain/utils"

	"github.com/2Finance-Labs/2finance-sdk-client/protocol"
)

const (
	walletVersion  = 1
	unlockDuration = 2 * time.Minute
)

type WalletAction string

const (
	ActionExportPrivateKey WalletAction = "export_private_key"
	ActionChangePassword   WalletAction = "change_password"
	ActionDeleteWallet     WalletAction = "delete_wallet"
	ActionWithdraw         WalletAction = "withdraw"
)

type WalletFile struct {
	Version             int               `json:"version"`
	Owner               string            `json:"owner"`
	EncryptedPrivateKey []byte            `json:"encrypted_private_key"`
	WrappedKeyset       WrappedTinkKeyset `json:"wrapped_keyset"`
	CreatedAt           time.Time         `json:"created_at"`
	UpdatedAt           time.Time         `json:"updated_at"`
}

type WalletManager struct {
	mu sync.RWMutex

	filePath string
	owner    string

	privateKey []byte

	unlockedUntil time.Time
	lockTimer     *time.Timer
}

type SignTransactionInput struct {
	ChainID       uint8
	From          string
	To            string
	Method        string
	Data          map[string]interface{}
	Version       uint8
	UUID7         string
	Authorization *transaction.AuthorizationEnvelope
}

type IWalletManager interface {
	ImportPrivateKey(privateKey []byte, password string) error

	Lock() error
	UnlockWithPassword(password string) error
	IsUnlocked() bool

	ChangePassword(currentPassword string, newPassword string) error

	OwnerAddress() string

	SignTransaction(input SignTransactionInput) (*transaction.Transaction, error)
	SignPreparedTransaction(input protocol.PreparedTransaction) (protocol.SignedTransaction, error)
}

func NewWalletManager(filePath string) IWalletManager {
	return &WalletManager{
		filePath: filePath,
	}
}

func (w *WalletManager) ImportPrivateKey(privateKey []byte, password string) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if password == "" {
		return errors.New("password is required")
	}

	if len(privateKey) == 0 {
		return fmt.Errorf("private key is required")
	}

	if w.filePath == "" {
		return fmt.Errorf("wallet file path is required")
	}

	privateKeyHex := string(privateKey)

	publicKey, err := keys.PublicKeyFromEd25519PrivateHex(privateKeyHex)
	if err != nil {
		return fmt.Errorf("failed to derive public key from private key: %w", err)
	}

	owner := keys.PublicKeyToHex(publicKey)
	if owner == "" {
		return fmt.Errorf("owner is required")
	}

	w.owner = owner

	encryptionKey := NewEncryption(w.owner)

	kh, err := encryptionKey.NewAEAD()
	if err != nil {
		return fmt.Errorf("failed to create wallet AEAD: %w", err)
	}

	encryptedPrivateKey, err := encryptionKey.EncryptPrivateKey(privateKey)
	if err != nil {
		return fmt.Errorf("failed to encrypt private key: %w", err)
	}

	wrappedKeyset, err := WrapTinkKeyset(kh, password)
	if err != nil {
		return fmt.Errorf("failed to wrap keyset: %w", err)
	}

	now := time.Now()

	walletFile := WalletFile{
		Version:             walletVersion,
		Owner:               w.owner,
		EncryptedPrivateKey: encryptedPrivateKey,
		WrappedKeyset:       wrappedKeyset,
		CreatedAt:           now,
		UpdatedAt:           now,
	}

	walletPayload, err := json.Marshal(walletFile)
	if err != nil {
		return fmt.Errorf("failed to marshal wallet file: %w", err)
	}
	defer clearBytes(walletPayload)

	encryptionFile := NewEncryptionFile(w.filePath)

	localEncryptedWalletFile, err := encryptionFile.Encrypt(walletPayload, password)
	if err != nil {
		return fmt.Errorf("failed to encrypt wallet file: %w", err)
	}

	if err := encryptionFile.Write(*localEncryptedWalletFile); err != nil {
		return fmt.Errorf("failed to write wallet file: %w", err)
	}

	clearBytes(privateKey)
	w.lockMemoryLocked()

	return nil
}

func (w *WalletManager) Lock() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.lockMemoryLocked()

	return nil
}

func (w *WalletManager) UnlockWithPassword(password string) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if password == "" {
		return errors.New("password is required")
	}

	if w.filePath == "" {
		return fmt.Errorf("wallet file path is required")
	}

	encryptionFile := NewEncryptionFile(w.filePath)

	localEncryptedWalletFile, err := encryptionFile.Read()
	if err != nil {
		return fmt.Errorf("failed to read wallet file: %w", err)
	}

	walletPayload, err := encryptionFile.Decrypt(*localEncryptedWalletFile, password)
	if err != nil {
		return fmt.Errorf("failed to decrypt wallet file: %w", err)
	}
	defer clearBytes(walletPayload)

	var walletFile WalletFile
	if err := json.Unmarshal(walletPayload, &walletFile); err != nil {
		return fmt.Errorf("failed to unmarshal wallet file: %w", err)
	}

	if walletFile.Version != walletVersion {
		return fmt.Errorf("unsupported wallet version: %d", walletFile.Version)
	}

	if walletFile.Owner == "" {
		return fmt.Errorf("wallet owner is required")
	}

	if len(walletFile.EncryptedPrivateKey) == 0 {
		return fmt.Errorf("encrypted private key is required")
	}

	if w.owner == "" {
		w.owner = walletFile.Owner
	}

	if walletFile.Owner != w.owner {
		return fmt.Errorf("wallet owner mismatch")
	}

	kh, err := UnwrapTinkKeyset(walletFile.WrappedKeyset, password)
	if err != nil {
		return fmt.Errorf("failed to unwrap keyset: %w", err)
	}

	encryptionKey := NewEncryption(walletFile.Owner)

	if err := encryptionKey.LoadAEAD(kh); err != nil {
		return fmt.Errorf("failed to load wallet AEAD: %w", err)
	}

	privateKey, err := encryptionKey.DecryptPrivateKey(walletFile.EncryptedPrivateKey)
	if err != nil {
		return fmt.Errorf("failed to decrypt private key: %w", err)
	}

	w.lockMemoryLocked()

	w.privateKey = privateKey
	w.owner = walletFile.Owner
	w.unlockedUntil = time.Now().Add(unlockDuration)

	w.lockTimer = time.AfterFunc(unlockDuration, func() {
		_ = w.Lock()
	})

	return nil
}

func (w *WalletManager) IsUnlocked() bool {
	w.mu.RLock()
	defer w.mu.RUnlock()

	return w.isUnlockedLocked()
}

func (w *WalletManager) ChangePassword(currentPassword string, newPassword string) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if currentPassword == "" {
		return errors.New("current password is required")
	}

	if newPassword == "" {
		return errors.New("new password is required")
	}

	if currentPassword == newPassword {
		return errors.New("new password must be different from current password")
	}

	if w.owner == "" {
		return fmt.Errorf("owner is required")
	}

	if w.filePath == "" {
		return fmt.Errorf("wallet file path is required")
	}

	encryptionFile := NewEncryptionFile(w.filePath)

	localEncryptedWalletFile, err := encryptionFile.Read()
	if err != nil {
		return fmt.Errorf("failed to read wallet file: %w", err)
	}

	walletPayload, err := encryptionFile.Decrypt(*localEncryptedWalletFile, currentPassword)
	if err != nil {
		return fmt.Errorf("failed to decrypt wallet file with current password: %w", err)
	}
	defer clearBytes(walletPayload)

	var walletFile WalletFile
	if err := json.Unmarshal(walletPayload, &walletFile); err != nil {
		return fmt.Errorf("failed to unmarshal wallet file: %w", err)
	}

	if walletFile.Version != walletVersion {
		return fmt.Errorf("unsupported wallet version: %d", walletFile.Version)
	}

	if walletFile.Owner != w.owner {
		return fmt.Errorf("wallet owner mismatch")
	}

	kh, err := UnwrapTinkKeyset(walletFile.WrappedKeyset, currentPassword)
	if err != nil {
		return fmt.Errorf("failed to unwrap keyset with current password: %w", err)
	}

	newWrappedKeyset, err := WrapTinkKeyset(kh, newPassword)
	if err != nil {
		return fmt.Errorf("failed to wrap keyset with new password: %w", err)
	}

	walletFile.WrappedKeyset = newWrappedKeyset
	walletFile.UpdatedAt = time.Now()

	updatedWalletPayload, err := json.Marshal(walletFile)
	if err != nil {
		return fmt.Errorf("failed to marshal updated wallet file: %w", err)
	}
	defer clearBytes(updatedWalletPayload)

	newLocalEncryptedWalletFile, err := encryptionFile.Encrypt(updatedWalletPayload, newPassword)
	if err != nil {
		return fmt.Errorf("failed to encrypt wallet file with new password: %w", err)
	}

	if err := encryptionFile.Write(*newLocalEncryptedWalletFile); err != nil {
		return fmt.Errorf("failed to write rotated wallet file: %w", err)
	}

	w.lockMemoryLocked()

	return nil
}

func (w *WalletManager) OwnerAddress() string {
	w.mu.RLock()
	defer w.mu.RUnlock()

	return w.owner
}

func (w *WalletManager) SignTransaction(input SignTransactionInput) (*transaction.Transaction, error) {
	if input.Method == "" {
		return nil, errors.New("method is required")
	}
	if len(input.Data) == 0 {
		return nil, errors.New("data cannot be empty")
	}

	w.mu.RLock()

	if !w.isUnlockedLocked() {
		w.mu.RUnlock()
		return nil, errors.New("wallet is locked")
	}

	privateKey := cloneBytes(w.privateKey)

	w.mu.RUnlock()
	defer clearBytes(privateKey)

	return w.signTransactionWithPrivateKey(
		privateKey,
		input.ChainID,
		input.From,
		input.To,
		input.Method,
		input.Data,
		input.Version,
		input.UUID7,
		input.Authorization,
	)
}

func (w *WalletManager) SignPreparedTransaction(input protocol.PreparedTransaction) (protocol.SignedTransaction, error) {
	signed, err := w.SignTransaction(SignTransactionInput{
		ChainID:       input.ChainID,
		From:          input.From,
		To:            input.To,
		Method:        input.Method,
		Data:          input.Data,
		Version:       input.Version,
		UUID7:         input.UUID7,
		Authorization: input.Authorization,
	})
	if err != nil {
		return protocol.SignedTransaction{}, err
	}

	return protocol.SignedTransactionFromNetwork(signed)
}

func GenerateEd25519KeyPairHex() (string, string, error) {
	publicKey, privateKey, err := keys.GenerateEd25519KeyPair()
	if err != nil {
		return "", "", fmt.Errorf("failed to generate public key: %w", err)
	}

	privateKeyHex := keys.PrivateKeyToHex(privateKey)
	publicKeyHex := keys.PublicKeyToHex(publicKey)

	return publicKeyHex, privateKeyHex, nil
}

func (w *WalletManager) exportPrivateKey(password string) ([]byte, error) {
	return w.privateKeyForAction(ActionExportPrivateKey, password)
}

func (w *WalletManager) privateKeyForAction(action WalletAction, password string) ([]byte, error) {
	if requiresPassword(action) {
		if password == "" {
			return nil, errors.New("password is required")
		}

		if err := w.UnlockWithPassword(password); err != nil {
			return nil, err
		}
	}

	if !w.IsUnlocked() {
		if password == "" {
			return nil, errors.New("wallet is locked")
		}

		if err := w.UnlockWithPassword(password); err != nil {
			return nil, err
		}
	}

	w.mu.RLock()
	defer w.mu.RUnlock()

	return cloneBytes(w.privateKey), nil
}

func requiresPassword(action WalletAction) bool {
	switch action {
	case ActionExportPrivateKey,
		ActionChangePassword,
		ActionDeleteWallet,
		ActionWithdraw:
		return true
	default:
		return false
	}
}

func (w *WalletManager) isUnlockedLocked() bool {
	return len(w.privateKey) > 0 && time.Now().Before(w.unlockedUntil)
}

func (w *WalletManager) lockMemoryLocked() {
	if w.lockTimer != nil {
		w.lockTimer.Stop()
		w.lockTimer = nil
	}

	if len(w.privateKey) > 0 {
		clearBytes(w.privateKey)
	}

	w.privateKey = nil
	w.unlockedUntil = time.Time{}
}

func (w *WalletManager) signTransactionWithPrivateKey(
	privateKey []byte,
	chainID uint8,
	from string,
	to string,
	method string,
	data utils.JSONB,
	version uint8,
	uuid7 string,
	authorization *transaction.AuthorizationEnvelope,
) (*transaction.Transaction, error) {
	if len(privateKey) == 0 {
		return nil, errors.New("private key is required")
	}

	dataRawMessage, err := utils.MapToRawMessage(data)
	if err != nil {
		return nil, fmt.Errorf("failed to convert data to RawMessage: %w", err)
	}

	newTx := transaction.NewTransaction(
		chainID,
		from,
		to,
		method,
		dataRawMessage,
		version,
		uuid7,
	)

	tx := newTx.Get()
	tx.Authorization = authorization

	signedTx, err := transaction.SignTransactionHexKey(string(privateKey), tx)
	if err != nil {
		return nil, fmt.Errorf("failed to sign transaction: %w", err)
	}

	return signedTx, nil
}

func (w *WalletManager) loadPrivateKeyOnce(password string) ([]byte, error) {
	w.mu.RLock()
	filePath := w.filePath
	owner := w.owner
	w.mu.RUnlock()

	if password == "" {
		return nil, errors.New("password is required")
	}

	if filePath == "" {
		return nil, fmt.Errorf("wallet file path is required")
	}

	encryptionFile := NewEncryptionFile(filePath)

	localEncryptedWalletFile, err := encryptionFile.Read()
	if err != nil {
		return nil, fmt.Errorf("failed to read wallet file: %w", err)
	}

	walletPayload, err := encryptionFile.Decrypt(*localEncryptedWalletFile, password)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt wallet file: %w", err)
	}
	defer clearBytes(walletPayload)

	var walletFile WalletFile
	if err := json.Unmarshal(walletPayload, &walletFile); err != nil {
		return nil, fmt.Errorf("failed to unmarshal wallet file: %w", err)
	}

	if walletFile.Version != walletVersion {
		return nil, fmt.Errorf("unsupported wallet version: %d", walletFile.Version)
	}

	if walletFile.Owner == "" {
		return nil, fmt.Errorf("wallet owner is required")
	}

	if len(walletFile.EncryptedPrivateKey) == 0 {
		return nil, fmt.Errorf("encrypted private key is required")
	}

	if owner != "" && walletFile.Owner != owner {
		return nil, fmt.Errorf("wallet owner mismatch")
	}

	kh, err := UnwrapTinkKeyset(walletFile.WrappedKeyset, password)
	if err != nil {
		return nil, fmt.Errorf("failed to unwrap keyset: %w", err)
	}

	encryptionKey := NewEncryption(walletFile.Owner)

	if err := encryptionKey.LoadAEAD(kh); err != nil {
		return nil, fmt.Errorf("failed to load wallet AEAD: %w", err)
	}

	privateKey, err := encryptionKey.DecryptPrivateKey(walletFile.EncryptedPrivateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt private key: %w", err)
	}

	return privateKey, nil
}
