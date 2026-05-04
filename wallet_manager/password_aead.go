package wallet_manager

import (
	"crypto/cipher"
	"errors"
	"fmt"

	"golang.org/x/crypto/argon2"
	"golang.org/x/crypto/chacha20poly1305"
)

const (
	keysetSaltSize = 16

	argonTime    uint32 = 3
	argonMemory  uint32 = 64 * 1024
	argonThreads uint8  = 1
	argonKeyLen  uint32 = 32
)

type KeysetKDFParams struct {
	Alg      string `json:"alg"`
	Time     uint32 `json:"time"`
	MemoryKB uint32 `json:"memory_kb"`
	Parallel uint8  `json:"parallel"`
	KeyLen   uint32 `json:"key_len"`
	Salt     []byte `json:"salt"`
}

type PasswordAEAD struct {
	aead cipher.AEAD
}

func NewKeysetKDFParams() (KeysetKDFParams, error) {
	salt, err := randomBytes(keysetSaltSize)
	if err != nil {
		return KeysetKDFParams{}, fmt.Errorf("failed to generate salt: %w", err)
	}

	return KeysetKDFParams{
		Alg:      "argon2id",
		Time:     argonTime,
		MemoryKB: argonMemory,
		Parallel: argonThreads,
		KeyLen:   argonKeyLen,
		Salt:     salt,
	}, nil
}

func NewPasswordAEAD(password string, params KeysetKDFParams) (*PasswordAEAD, error) {
	if password == "" {
		return nil, errors.New("password is required")
	}

	if params.Alg != "argon2id" {
		return nil, fmt.Errorf("unsupported KDF algorithm: %s", params.Alg)
	}

	if len(params.Salt) == 0 {
		return nil, fmt.Errorf("salt is required")
	}

	key := argon2.IDKey(
		[]byte(password),
		params.Salt,
		params.Time,
		params.MemoryKB,
		params.Parallel,
		params.KeyLen,
	)
	defer clearBytes(key)

	aeadPrimitive, err := chacha20poly1305.NewX(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create password AEAD: %w", err)
	}

	return &PasswordAEAD{
		aead: aeadPrimitive,
	}, nil
}

func (p *PasswordAEAD) Encrypt(plaintext, associatedData []byte) ([]byte, error) {
	nonce, err := randomBytes(chacha20poly1305.NonceSizeX)
	if err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	ciphertext := p.aead.Seal(nil, nonce, plaintext, associatedData)

	output := make([]byte, 0, len(nonce)+len(ciphertext))
	output = append(output, nonce...)
	output = append(output, ciphertext...)

	return output, nil
}

func (p *PasswordAEAD) Decrypt(ciphertext, associatedData []byte) ([]byte, error) {
	if len(ciphertext) <= chacha20poly1305.NonceSizeX {
		return nil, fmt.Errorf("invalid ciphertext")
	}

	nonce := ciphertext[:chacha20poly1305.NonceSizeX]
	encrypted := ciphertext[chacha20poly1305.NonceSizeX:]

	plaintext, err := p.aead.Open(nil, nonce, encrypted, associatedData)
	if err != nil {
		return nil, fmt.Errorf("invalid password or corrupted keyset")
	}

	return plaintext, nil
}
