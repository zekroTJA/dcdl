package hasher

import (
	"errors"

	"github.com/speps/go-hashids"
)

var (
	hasher, _ = hashids.New()

	ErrEmptyData = errors.New("empty data")
)

func Hash(v ...int) (hash string, err error) {
	if len(v) == 0 {
		err = ErrEmptyData
		return
	}

	hash, err = hasher.Encode(v)
	return
}
