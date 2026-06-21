package rollup

import (
	"bytes"
	"math/rand"
	"testing"

	cmimc "github.com/consensys/gnark-crypto/ecc/bn254/fr/mimc"
)

func TestAccountSerializeRoundTrip(t *testing.T) {
	r := rand.New(rand.NewSource(1)) //#nosec G404 -- deterministic test
	acc, _, err := NewAccount(3, 100, r)
	if err != nil {
		t.Fatalf("new account: %v", err)
	}
	acc.Nonce = 7

	b := acc.Serialize()
	if len(b) != SizeAccount {
		t.Fatalf("serialized length = %d, want %d", len(b), SizeAccount)
	}

	got, err := Deserialize(b)
	if err != nil {
		t.Fatalf("deserialize: %v", err)
	}
	if got.Index != acc.Index || got.Nonce != acc.Nonce {
		t.Fatalf("index/nonce mismatch: got (%d,%d) want (%d,%d)", got.Index, got.Nonce, acc.Index, acc.Nonce)
	}
	if !got.Balance.Equal(&acc.Balance) {
		t.Fatal("balance mismatch after round trip")
	}
	if !got.PubKey.A.X.Equal(&acc.PubKey.A.X) || !got.PubKey.A.Y.Equal(&acc.PubKey.A.Y) {
		t.Fatal("pubkey mismatch after round trip")
	}
}

func TestDeserializeWrongSize(t *testing.T) {
	if _, err := Deserialize(make([]byte, SizeAccount-1)); err == nil {
		t.Fatal("expected size error, got nil")
	}
}

func TestAccountHashDeterministic(t *testing.T) {
	r := rand.New(rand.NewSource(2)) //#nosec G404 -- deterministic test
	acc, _, err := NewAccount(1, 50, r)
	if err != nil {
		t.Fatalf("new account: %v", err)
	}
	h := cmimc.NewMiMC()
	h1 := acc.Hash(h)
	h2 := acc.Hash(h)
	if !bytes.Equal(h1, h2) {
		t.Fatal("account hash is not deterministic")
	}

	acc.Balance.SetUint64(51)
	h3 := acc.Hash(h)
	if bytes.Equal(h1, h3) {
		t.Fatal("account hash did not change when balance changed")
	}
}
