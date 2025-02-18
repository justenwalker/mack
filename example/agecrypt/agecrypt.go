package agecrypt

import (
	"bytes"
	"errors"
	"fmt"
	"io"

	"filippo.io/age"

	"github.com/justenwalker/mack/thirdparty/exchange"
)

const (
	defaultMessageType = "age:v1"
)

type Recipient struct {
	KeyID string
	age.Recipient
}

func (i Recipient) String() string {
	if i.KeyID != "" {
		return i.KeyID
	}
	if str, ok := i.Recipient.(fmt.Stringer); ok {
		return str.String()
	}
	return ""
}

func NewEncryptor(r Recipient) (*Encryptor, error) {
	if r.Recipient == nil {
		return nil, errors.New("agecrypt.NewEncoder: r.PublicKey is nil")
	}
	if r.KeyID == "" {
		if str, ok := r.Recipient.(fmt.Stringer); ok {
			r.KeyID = str.String()
		}
	}

	return &Encryptor{
		recipient: r,
	}, nil
}

var (
	_ exchange.Encryptor = (*Encryptor)(nil)
	_ exchange.Decryptor = (*Decryptor)(nil)
)

type Encryptor struct {
	recipient Recipient
}

func (e *Encryptor) EncryptMessage(msg []byte) (em *exchange.EncryptedMessage, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("agecrypt.EncryptMessage: %w", err)
		}
	}()
	var buf bytes.Buffer
	wc, err := age.Encrypt(&buf, e.recipient)
	if err != nil {
		return nil, fmt.Errorf("age.Encrypt for '%s': %w", e.recipient.String(), err)
	}
	if _, err = wc.Write(msg); err != nil {
		return nil, fmt.Errorf("age.Encrypt.Write message: %w", err)
	}
	if err = wc.Close(); err != nil {
		return nil, fmt.Errorf("age.Encrypt.Close message: %w", err)
	}
	return &exchange.EncryptedMessage{
		Type:    defaultMessageType,
		KeyID:   e.recipient.String(),
		Payload: buf.Bytes(),
	}, nil
}

type Identity struct {
	KeyID string
	age.Identity
	age.Recipient
}

func (i Identity) String() string {
	if i.KeyID != "" {
		return i.KeyID
	}
	if str, ok := i.Recipient.(fmt.Stringer); ok {
		return str.String()
	}
	return ""
}

func NewDecryptor(ids []Identity) *Decryptor {
	ageIDs := make([]age.Identity, len(ids))
	for i, id := range ids {
		ageIDs[i] = id
	}
	return &Decryptor{
		ids: ageIDs,
	}
}

type Decryptor struct {
	ids []age.Identity
}

func (d *Decryptor) DecryptMessage(em *exchange.EncryptedMessage) (data []byte, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("agecrypt.DecryptMessage: %w", err)
		}
	}()
	if em.Type != defaultMessageType {
		return nil, fmt.Errorf("type mismatch: EncryptedMessage.Type(%s) != '%s'", em.Type, defaultMessageType)
	}

	// narrow down the list of ids if we can
	ids := filterIdentities(em.KeyID, d.ids)

	r, err := age.Decrypt(bytes.NewReader(em.Payload), ids...)
	if err != nil {
		return nil, fmt.Errorf("age.Decryptor for '%s': %w", em.KeyID, err)
	}
	req, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("age.Decryptor.Read message: %w", err)
	}
	return req, nil
}

func filterIdentities(keyID string, ids []age.Identity) []age.Identity {
	if keyID == "" { // no hint, so try all
		return ids
	}
	for i, id := range ids {
		if str, ok := id.(fmt.Stringer); ok {
			kid := str.String()
			if keyID == kid {
				return ids[i : i+1]
			}
		}
	}
	return ids
}
