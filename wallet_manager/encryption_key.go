package wallet_manager

import "fmt"

type EncryptionKey struct {
	Primitive tink.AEAD
	Key       []byte
	Owner     string
}

type IEncryptionKey interface {
	NewAEAD() (error)
	EncryptPrivateKey(encryption EncryptionKey) (error)
	DecryptPrivateKey(encryption EncryptionKey) (error)
}

func NewEncryption(key []byte, owner string) IEncryptionKey {
	return &EncryptionKey{
		Key: key,
		Owner: owner,
	}
}

func (e *EncryptionKey) NewAEAD() (error) {
	primitive, _, err := symmetric.NewAEAD()
	if err != nil {
		return fmt.Errorf("failed to generate AEAD: %w", err)
	}

	e.Primitive = primitive

	return nil
}

func (e *EncryptionKey) EncryptPrivateKey(encryption EncryptionKey) (error) {
	ciphertext, err := symmetric.EncryptPrivateKey(e.Primitive, e.Key, e.Owner)
	if err != nil {
		return fmt.Errorf("failed to encrypt private key: %w", err)
	}

	e.Key = ciphertext

	return nil
}

func (e *EncryptionKey) DecryptPrivateKey(encryption EncryptionKey) (error) {
	plaintext, err := symmetric.DecryptPrivateKey(e.Primitive, e.Key, e.Owner)
	if err != nil {
		return fmt.Errorf("failed to decrypt private key: %w", err)
	}

	e.Key = plaintext

	return nil
}
