package agecrypt

import (
	"bytes"
	"math/rand"
	"reflect"
	"testing"
	"testing/quick"

	"filippo.io/age"
)

func TestEncryptorDecryptor(t *testing.T) {
	fn := func(plaintext []byte, r Recipient, ids []Identity) bool {
		enc, err := NewEncryptor(r)
		if err != nil {
			t.Errorf("NewEncryptor: %v", err)
			return false
		}
		dec := NewDecryptor(ids)
		em, err := enc.EncryptMessage(plaintext)
		if err != nil {
			t.Errorf("EncryptMessage: %v", err)
			return false
		}
		dm, err := dec.DecryptMessage(em)
		if err != nil {
			t.Errorf("EncryptMessage: %v", err)
			return false
		}
		return bytes.Equal(plaintext, dm)
	}
	err := quick.Check(fn, &quick.Config{
		Values: func(values []reflect.Value, rand *rand.Rand) {
			sz := rand.Int63n(65536)
			plaintext := make([]byte, sz)
			_, err := rand.Read(plaintext)
			if err != nil {
				panic(err)
			}
			id, err := age.GenerateX25519Identity()
			if err != nil {
				panic(err)
			}
			values[0] = reflect.ValueOf(plaintext)
			values[1] = reflect.ValueOf(Recipient{
				KeyID:     id.Recipient().String(),
				Recipient: id.Recipient(),
			})
			values[2] = reflect.ValueOf([]Identity{
				{
					KeyID:     id.Recipient().String(),
					Identity:  id,
					Recipient: id.Recipient(),
				},
			})
		},
	})
	if err != nil {
		t.Fatalf("EncryptorDecryptor quick check failed: %v", err)
	}
}
