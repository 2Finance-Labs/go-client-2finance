package wallet_manager

import (
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"
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

	privateKey    []byte
	unlockedUntil time.Time
	lockTimer     *time.Timer

	passwordRequiredMethods map[string]bool
}

type IWalletManager interface {
	SetupWallet(privateKey []byte, password string) error
	Lock() error
	Unlock(password string) error
	ForceLock() error
	IsUnlocked() bool
	RequiresPassword(methodName string) bool
	RotatePassword(currentPassword string, newPassword string) error
	GetPrivateKey(methodName string, password string) ([]byte, error)
}

func NewWalletManager(owner string, filePath string) IWalletManager {
	return &WalletManager{
		owner:    owner,
		filePath: filePath,
		passwordRequiredMethods: map[string]bool{
			"ExportPrivateKey": true,
			"ChangePassword":   true,
			"DeleteWallet":     true,
			"Withdraw":         true,
		},
	}
}

func (w *WalletManager) SetupWallet(privateKey []byte, password string) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if password == "" {
		return errors.New("password is required")
	}

	if len(privateKey) == 0 {
		return fmt.Errorf("private key is required")
	}

	if w.owner == "" {
		return fmt.Errorf("owner is required")
	}

	if w.filePath == "" {
		return fmt.Errorf("wallet file path is required")
	}

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

	walletPayload, err := encryptionFile.Decrypt(*localEncryptedWalletFile, password)
	if err != nil {
		return fmt.Errorf("failed to decrypt wallet file: %w", err)
	}

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

	kh, err := UnwrapTinkKeyset(walletFile.WrappedKeyset, password)
	if err != nil {
		return fmt.Errorf("failed to unwrap keyset: %w", err)
	}

	encryptionKey := NewEncryption(w.owner)

	if err := encryptionKey.LoadAEAD(kh); err != nil {
		return fmt.Errorf("failed to load wallet AEAD: %w", err)
	}

	privateKey, err := encryptionKey.DecryptPrivateKey(walletFile.EncryptedPrivateKey)
	if err != nil {
		return fmt.Errorf("failed to decrypt private key: %w", err)
	}

	w.lockMemoryLocked()

	w.privateKey = privateKey
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

func (w *WalletManager) RequiresPassword(methodName string) bool {
	w.mu.RLock()
	defer w.mu.RUnlock()

	return w.passwordRequiredMethods[methodName]
}

func (w *WalletManager) GetPrivateKey(methodName string, password string) ([]byte, error) {
	if w.RequiresPassword(methodName) {
		if password == "" {
			return nil, errors.New("password is required")
		}

		if err := w.Unlock(password); err != nil {
			return nil, err
		}
	}

	if !w.IsUnlocked() {
		if password == "" {
			return nil, errors.New("wallet is locked")
		}

		if err := w.Unlock(password); err != nil {
			return nil, err
		}
	}

	w.mu.RLock()
	defer w.mu.RUnlock()

	return cloneBytes(w.privateKey), nil
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