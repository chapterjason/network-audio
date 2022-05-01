package ulidx

import (
	rcrypto "crypto/rand"

	"github.com/oklog/ulid/v2"
)

func New() (ulid.ULID, error) {
	return ulid.New(ulid.Now(), rcrypto.Reader)
}

func MustNew() ulid.ULID {
	id, err := New()

	if err != nil {
		panic(err)
	}

	return id
}
