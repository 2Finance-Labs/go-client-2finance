package wallet_manager

import (
	"fmt"

	"gitlab.com/2finance/2finance-network/blockchain/encryption/encryptor"
)

type EncryptionFile struct {
	FilePath string
	Password string
	Key      []byte
	Env      encryptor.ChainEnvelope
}

type IEncryptionFile interface {
	EncryptFile(encryption EncryptionKey) error
	DecryptFile(encryption EncryptionKey) error
}

func NewEncryptionFile(filePath string, password string) IEncryptionFile {
	return &EncryptionFile{
		FilePath: filePath,
		Password: password, // TODO: Consider using a more secure way to handle passwords, such as environment variables or a secrets manager.
	}
}

func (e *EncryptionFile) EncryptFile(encryption EncryptionKey) error {
	env, cipher, err := encryptor.EncryptFile(e.FilePath, e.Password)
	if err != nil {
		return fmt.Errorf("failed to encrypt file: %w", err)
	}

	e.Env = env
	e.Key = cipher

	return nil
}

func (e *EncryptionFile) DecryptFile(encryption EncryptionKey) error {
	plaintext, err := encryptor.DecryptFile(e.Env, e.Password, e.Key)
	if err != nil {
		return fmt.Errorf("failed to decrypt file: %w", err)
	}

	e.Key = plaintext

	return nil
}
