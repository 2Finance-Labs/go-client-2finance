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
)

const (
	walletVersion  = 1
	unlockDuration = 2 * time.Minute
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

	passwordRequiredMethods map[string]bool
}

type IWalletManager interface {
	ImportWallet(privateKey []byte, password string) error
	Lock() error
	Unlock(password string) error
	ForceLock() error
	IsUnlocked() bool
	AddRequiredPasswordMethods(methods ...string) error
	PasswordIsRequired(method string) bool
	RotatePassword(currentPassword string, newPassword string) error

	SignTransaction(
		chainId uint8,
		from, to, method string,
		data utils.JSONB,
		version uint8,
		uuid7 string,
	) (*transaction.Transaction, error)

	SignTransactionWithPassword(
		chainId uint8,
		from, to, method string,
		data utils.JSONB,
		version uint8,
		uuid7, password string,
	) (*transaction.Transaction, error)

	GetPublicKey() string
	GenerateEd25519KeyPairHex() (string, string, error)
}

func NewWalletManager(filePath string) IWalletManager {
	return &WalletManager{
		filePath: filePath,
	}
}

func (w *WalletManager) ImportWallet(privateKey []byte, password string) error {
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

func (w *WalletManager) ForceLock() error {
	return w.Lock()
}

func (w *WalletManager) Unlock(password string) error {
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

	// Se o manager ainda não tem owner, carrega do arquivo.
	// Isso resolve o caso de uma nova instância abrindo uma wallet já existente.
	if w.owner == "" {
		w.owner = walletFile.Owner
	}

	// Se o manager já tinha owner e o arquivo é de outro owner, bloqueia.
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

func (w *WalletManager) isUnlockedLocked() bool {
	return len(w.privateKey) > 0 && time.Now().Before(w.unlockedUntil)
}

func (w *WalletManager) PasswordIsRequired(method string) bool {
	w.mu.RLock()
	defer w.mu.RUnlock()

	if method == "" {
		return false
	}

	if w.passwordRequiredMethods == nil {
		return false
	}

	return w.passwordRequiredMethods[method]
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

func (w *WalletManager) AddRequiredPasswordMethods(methods ...string) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.passwordRequiredMethods == nil {
		w.passwordRequiredMethods = make(map[string]bool)
	}

	for _, method := range methods {
		if method == "" {
			return fmt.Errorf("method name is required")
		}

		w.passwordRequiredMethods[method] = true
	}

	return nil
}

func (w *WalletManager) RotatePassword(currentPassword string, newPassword string) error {
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

func (w *WalletManager) GenerateEd25519KeyPairHex() (string, string, error) {
	publicKey, privateKey, err := keys.GenerateEd25519KeyPair()
	if err != nil {
		return "", "", fmt.Errorf("failed to generate public key: %w", err)
	}

	privateKeyHex := keys.PrivateKeyToHex(privateKey)
	publicKeyHex := keys.PublicKeyToHex(publicKey)

	return publicKeyHex, privateKeyHex, nil
}

func (w *WalletManager) GetPublicKey() string {
	return w.owner
}

func (w *WalletManager) signTransactionWithPrivateKey(
	privateKey []byte,
	chainId uint8,
	from string,
	to string,
	method string,
	data utils.JSONB,
	version uint8,
	uuid7 string,
) (*transaction.Transaction, error) {
	dataRawMessage, err := utils.MapToRawMessage(data)
	if err != nil {
		return nil, fmt.Errorf("failed to convert data to RawMessage: %w", err)
	}

	newTx := transaction.NewTransaction(
		chainId,
		from,
		to,
		method,
		dataRawMessage,
		version,
		uuid7,
	)

	tx := newTx.Get()

	signedTx, err := transaction.SignTransactionHexKey(string(privateKey), tx)
	if err != nil {
		return nil, fmt.Errorf("failed to sign transaction: %w", err)
	}

	return signedTx, nil
}

func (w *WalletManager) SignTransaction(chainId uint8, from, to, method string, data utils.JSONB, version uint8, uuid7 string) (*transaction.Transaction, error) {

	dataRawMessage, err := utils.MapToRawMessage(data)
	if err != nil {
		return nil, fmt.Errorf("failed to convert data to RawMessage: %w", err)
	}

	// 1. create new tx
	newTx := transaction.NewTransaction(chainId, from, to, method, dataRawMessage, version, uuid7)

	// 2. get serialized form (here it's just the object)
	tx := newTx.Get()

	// 3. sign
	signedTx, err := transaction.SignTransactionHexKey(string(w.privateKey), tx)
	if err != nil {
		return nil, fmt.Errorf("failed to sign transaction: %w", err)
	}
	return signedTx, nil
}

func (w *WalletManager) SignTransactionWithPassword(
	chainId uint8,
	from string,
	to string,
	method string,
	data utils.JSONB,
	version uint8,
	uuid7 string,
	password string,
) (*transaction.Transaction, error) {
	if password == "" {
		return nil, errors.New("password is required")
	}

	w.mu.Lock()
	w.lockMemoryLocked()
	w.mu.Unlock()

	if err := w.Unlock(password); err != nil {
		return nil, err
	}

	w.mu.RLock()
	privateKey := cloneBytes(w.privateKey)
	w.mu.RUnlock()

	if len(privateKey) == 0 {
		return nil, errors.New("private key is not loaded")
	}
	defer clearBytes(privateKey)

	return w.signTransactionWithPrivateKey(
		privateKey,
		chainId,
		from,
		to,
		method,
		data,
		version,
		uuid7,
	)
}
