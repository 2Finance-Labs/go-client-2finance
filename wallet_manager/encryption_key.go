package wallet_manager

import (
	"fmt"

	"github.com/google/tink/go/tink"
	"gitlab.com/2finance/2finance-network/blockchain/encryption/symmetric"
)

type EncryptionKey struct {
	Primitive  tink.AEAD
	PrivateKey []byte
	Owner      string
}

type IEncryptionKey interface {
	NewAEAD() error
	EncryptPrivateKey(encryption EncryptionKey) error
	DecryptPrivateKey(encryption EncryptionKey) error
}

func NewEncryption(privateKey []byte, owner string) IEncryptionKey {
	return &EncryptionKey{
		PrivateKey: privateKey,
		Owner:      owner,
	}
}

func (e *EncryptionKey) NewAEAD() error {
	primitive, _, err := symmetric.NewAEAD()
	if err != nil {
		return fmt.Errorf("failed to generate AEAD: %w", err)
	}

	e.Primitive = primitive

	return nil
}

func (e *EncryptionKey) EncryptPrivateKey(encryption EncryptionKey) error {
	ciphertext, err := symmetric.EncryptPrivateKey(e.Primitive, e.PrivateKey, e.Owner)
	if err != nil {
		return fmt.Errorf("failed to encrypt private key: %w", err)
	}

	e.PrivateKey = ciphertext

	return nil
}

func (e *EncryptionKey) DecryptPrivateKey(encryption EncryptionKey) error {
	plaintext, err := symmetric.DecryptPrivateKey(e.Primitive, e.PrivateKey, e.Owner)
	if err != nil {
		return fmt.Errorf("failed to decrypt private key: %w", err)
	}

	e.PrivateKey = plaintext

	return nil
}
