package wallet_manager

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"gitlab.com/2finance/2finance-network/blockchain/encryption/encryptor"
)

type LocalEncryptedWalletFile struct {
	Env    encryptor.ChainEnvelope `json:"env"`
	Cipher []byte                  `json:"cipher"`
}

type EncryptionFile struct {
	FilePath string
}

type IEncryptionFile interface {
	EncryptAndWrite(data []byte, password string) error
	ReadAndDecrypt(password string) ([]byte, error)
}

func NewEncryptionFile(filePath string) IEncryptionFile {
	return &EncryptionFile{
		FilePath: filePath,
	}
}

func (e *EncryptionFile) EncryptAndWrite(data []byte, password string) error {
	if password == "" {
		return errors.New("password is required")
	}

	if len(data) == 0 {
		return fmt.Errorf("wallet data is required")
	}

	if e.FilePath == "" {
		return fmt.Errorf("wallet file path is required")
	}

	dir := filepath.Dir(e.FilePath)

	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("failed to create wallet directory: %w", err)
	}

	tmpFile, err := os.CreateTemp(dir, ".wallet-plain-*")
	if err != nil {
		return fmt.Errorf("failed to create temporary wallet file: %w", err)
	}

	tmpPath := tmpFile.Name()

	defer func() {
		_ = os.Remove(tmpPath)
	}()

	if err := tmpFile.Chmod(0600); err != nil {
		_ = tmpFile.Close()
		return fmt.Errorf("failed to chmod temporary wallet file: %w", err)
	}

	if _, err := tmpFile.Write(data); err != nil {
		_ = tmpFile.Close()
		return fmt.Errorf("failed to write temporary wallet file: %w", err)
	}

	if err := tmpFile.Close(); err != nil {
		return fmt.Errorf("failed to close temporary wallet file: %w", err)
	}

	env, cipherBytes, err := encryptor.EncryptFile(tmpPath, password)
	if err != nil {
		return fmt.Errorf("failed to encrypt wallet file: %w", err)
	}

	env = encryptor.WithLocation(env, e.FilePath)

	localFile := LocalEncryptedWalletFile{
		Env:    env,
		Cipher: cipherBytes,
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

func (e *EncryptionFile) ReadAndDecrypt(password string) ([]byte, error) {
	if password == "" {
		return nil, errors.New("password is required")
	}

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

	if err := encryptor.VerifyCipher(localFile.Env, localFile.Cipher); err != nil {
		return nil, fmt.Errorf("wallet file integrity check failed: %w", err)
	}

	plaintext, err := encryptor.DecryptFile(localFile.Env, password, localFile.Cipher)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt wallet file: %w", err)
	}

	return plaintext, nil
}
