package wallet_manager

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

const walletFileAssociatedData = "wallet-manager-file:v1"

type LocalEncryptedWalletFile struct {
	KDF    KeysetKDFParams `json:"kdf"`
	Cipher []byte          `json:"cipher"`
}

type EncryptionFile struct {
	FilePath string
}

type IEncryptionFile interface {
	Encrypt(data []byte, password string) (*LocalEncryptedWalletFile, error)
	Decrypt(localFile LocalEncryptedWalletFile, password string) ([]byte, error)
	Write(localFile LocalEncryptedWalletFile) error
	Read() (*LocalEncryptedWalletFile, error)
}

func NewEncryptionFile(filePath string) IEncryptionFile {
	return &EncryptionFile{
		FilePath: filePath,
	}
}

func (e *EncryptionFile) Encrypt(data []byte, password string) (*LocalEncryptedWalletFile, error) {
	if password == "" {
		return nil, errors.New("password is required")
	}

	if len(data) == 0 {
		return nil, fmt.Errorf("wallet data is required")
	}

	kdf, err := NewKeysetKDFParams()
	if err != nil {
		return nil, fmt.Errorf("failed to create wallet file KDF params: %w", err)
	}

	passwordAEAD, err := NewPasswordAEAD(password, kdf)
	if err != nil {
		return nil, fmt.Errorf("failed to create wallet file AEAD: %w", err)
	}

	cipherBytes, err := passwordAEAD.Encrypt(data, []byte(walletFileAssociatedData))
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt wallet file data: %w", err)
	}

	return &LocalEncryptedWalletFile{
		KDF:    kdf,
		Cipher: cipherBytes,
	}, nil
}

func (e *EncryptionFile) Write(localFile LocalEncryptedWalletFile) error {
	if e.FilePath == "" {
		return fmt.Errorf("wallet file path is required")
	}

	if len(localFile.Cipher) == 0 {
		return fmt.Errorf("wallet encrypted data is required")
	}

	dir := filepath.Dir(e.FilePath)

	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("failed to create wallet directory: %w", err)
	}

	finalBytes, err := json.Marshal(localFile)
	if err != nil {
		return fmt.Errorf("failed to marshal encrypted wallet file: %w", err)
	}

	if err := os.WriteFile(e.FilePath, finalBytes, 0600); err != nil {
		return fmt.Errorf("failed to write encrypted wallet file: %w", err)
	}

	return nil
}

func (e *EncryptionFile) Read() (*LocalEncryptedWalletFile, error) {
	if e.FilePath == "" {
		return nil, fmt.Errorf("wallet file path is required")
	}

	finalBytes, err := os.ReadFile(e.FilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read encrypted wallet file: %w", err)
	}

	var localFile LocalEncryptedWalletFile
	if err := json.Unmarshal(finalBytes, &localFile); err != nil {
		return nil, fmt.Errorf("failed to unmarshal encrypted wallet file: %w", err)
	}

	if len(localFile.Cipher) == 0 {
		return nil, fmt.Errorf("wallet encrypted data is required")
	}

	return &localFile, nil
}

func (e *EncryptionFile) Decrypt(localFile LocalEncryptedWalletFile, password string) ([]byte, error) {
	if password == "" {
		return nil, errors.New("password is required")
	}

	if len(localFile.Cipher) == 0 {
		return nil, fmt.Errorf("wallet encrypted data is required")
	}

	passwordAEAD, err := NewPasswordAEAD(password, localFile.KDF)
	if err != nil {
		return nil, fmt.Errorf("failed to create wallet file AEAD: %w", err)
	}

	plaintext, err := passwordAEAD.Decrypt(localFile.Cipher, []byte(walletFileAssociatedData))
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt wallet file data: %w", err)
	}

	return plaintext, nil
}