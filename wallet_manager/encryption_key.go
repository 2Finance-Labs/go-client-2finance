package wallet_manager

import (
	"fmt"

	"github.com/google/tink/go/aead"
	"github.com/google/tink/go/keyset"
	"github.com/google/tink/go/tink"
	"gitlab.com/2finance/2finance-network/blockchain/encryption/symmetric"
)

type EncryptionKey struct {
	Primitive tink.AEAD
	Owner     string
}

type IEncryptionKey interface {
	NewAEAD() (*keyset.Handle, error)
	LoadAEAD(kh *keyset.Handle) error
	EncryptPrivateKey(privateKey []byte) ([]byte, error)
	DecryptPrivateKey(encryptedPrivateKey []byte) ([]byte, error)
}

func NewEncryption(owner string) IEncryptionKey {
	return &EncryptionKey{
		Owner: owner,
	}
}

func (e *EncryptionKey) NewAEAD() (*keyset.Handle, error) {
	if e.Owner == "" {
		return nil, fmt.Errorf("owner is required")
	}

	primitive, kh, err := symmetric.NewAEAD()
	if err != nil {
		return nil, fmt.Errorf("failed to generate AEAD: %w", err)
	}

	e.Primitive = primitive

	return kh, nil
}

func (e *EncryptionKey) LoadAEAD(kh *keyset.Handle) error {
	if kh == nil {
		return fmt.Errorf("keyset handle is required")
	}

	if e.Owner == "" {
		return fmt.Errorf("owner is required")
	}

	primitive, err := aead.New(kh)
	if err != nil {
		return fmt.Errorf("failed to load AEAD from keyset: %w", err)
	}

	e.Primitive = primitive

	return nil
}

func (e *EncryptionKey) EncryptPrivateKey(privateKey []byte) ([]byte, error) {
	if e.Primitive == nil {
		return nil, fmt.Errorf("AEAD primitive is required")
	}

	if len(privateKey) == 0 {
		return nil, fmt.Errorf("private key is required")
	}

	ciphertext, err := symmetric.EncryptPrivateKey(e.Primitive, privateKey, e.Owner)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt private key: %w", err)
	}

	return ciphertext, nil
}

func (e *EncryptionKey) DecryptPrivateKey(encryptedPrivateKey []byte) ([]byte, error) {
	if e.Primitive == nil {
		return nil, fmt.Errorf("AEAD primitive is required")
	}

	if len(encryptedPrivateKey) == 0 {
		return nil, fmt.Errorf("encrypted private key is required")
	}

	plaintext, err := symmetric.DecryptPrivateKey(e.Primitive, encryptedPrivateKey, e.Owner)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt private key: %w", err)
	}

	return plaintext, nil
}