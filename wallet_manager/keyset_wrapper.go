package wallet_manager

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/google/tink/go/keyset"
)

type WrappedTinkKeyset struct {
	KDF                 KeysetKDFParams `json:"kdf"`
	EncryptedKeysetJSON []byte          `json:"encrypted_keyset_json"`
}

func WrapTinkKeyset(kh *keyset.Handle, password string) (WrappedTinkKeyset, error) {
	if kh == nil {
		return WrappedTinkKeyset{}, fmt.Errorf("keyset handle is required")
	}

	if password == "" {
		return WrappedTinkKeyset{}, errors.New("password is required")
	}

	kdf, err := NewKeysetKDFParams()
	if err != nil {
		return WrappedTinkKeyset{}, err
	}

	wrappingAEAD, err := NewPasswordAEAD(password, kdf)
	if err != nil {
		return WrappedTinkKeyset{}, err
	}

	var buf bytes.Buffer

	if err := kh.Write(keyset.NewJSONWriter(&buf), wrappingAEAD); err != nil {
		return WrappedTinkKeyset{}, fmt.Errorf("failed to write encrypted Tink keyset: %w", err)
	}

	return WrappedTinkKeyset{
		KDF:                 kdf,
		EncryptedKeysetJSON: buf.Bytes(),
	}, nil
}

func UnwrapTinkKeyset(wrapped WrappedTinkKeyset, password string) (*keyset.Handle, error) {
	if password == "" {
		return nil, errors.New("password is required")
	}

	if len(wrapped.EncryptedKeysetJSON) == 0 {
		return nil, fmt.Errorf("encrypted keyset is required")
	}

	wrappingAEAD, err := NewPasswordAEAD(password, wrapped.KDF)
	if err != nil {
		return nil, err
	}

	reader := bytes.NewReader(wrapped.EncryptedKeysetJSON)

	kh, err := keyset.Read(keyset.NewJSONReader(reader), wrappingAEAD)
	if err != nil {
		return nil, fmt.Errorf("failed to read encrypted Tink keyset: %w", err)
	}

	return kh, nil
}
