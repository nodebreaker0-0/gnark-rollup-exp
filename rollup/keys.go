package rollup

import (
	"io"

	"github.com/consensys/gnark-crypto/ecc/bn254/twistededwards/eddsa"
)

// NewAccount creates an account at the given index with the given balance and a
// freshly generated EdDSA key pair, returning the account and its private key.
// Randomness is taken from r; pass a seeded source for deterministic tests.
func NewAccount(index int, balance uint64, r io.Reader) (Account, eddsa.PrivateKey, error) {
	priv, err := eddsa.GenerateKey(r)
	if err != nil {
		return Account{}, eddsa.PrivateKey{}, err
	}

	var a Account
	a.Index = uint64(index)
	a.Nonce = 0
	a.Balance.SetUint64(balance)
	a.PubKey = priv.PublicKey

	return a, *priv, nil
}
