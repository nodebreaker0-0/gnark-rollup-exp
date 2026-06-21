package rollup

import (
	"math/rand"
	"testing"

	cmimc "github.com/consensys/gnark-crypto/ecc/bn254/fr/mimc"
)

func TestTransferSignVerify(t *testing.T) {
	r := rand.New(rand.NewSource(42)) //#nosec G404 -- deterministic test
	sender, senderPriv, err := NewAccount(0, 100, r)
	if err != nil {
		t.Fatalf("sender: %v", err)
	}
	receiver, _, err := NewAccount(1, 100, r)
	if err != nil {
		t.Fatalf("receiver: %v", err)
	}

	h := cmimc.NewMiMC()
	transfer := NewTransfer(10, sender.PubKey, receiver.PubKey, sender.Nonce)
	if _, err := transfer.Sign(senderPriv, h); err != nil {
		t.Fatalf("sign: %v", err)
	}

	ok, err := transfer.Verify(h)
	if err != nil || !ok {
		t.Fatalf("expected valid signature, got ok=%v err=%v", ok, err)
	}
}

func TestTransferTamperFails(t *testing.T) {
	r := rand.New(rand.NewSource(43)) //#nosec G404 -- deterministic test
	sender, senderPriv, _ := NewAccount(0, 100, r)
	receiver, _, _ := NewAccount(1, 100, r)

	h := cmimc.NewMiMC()
	transfer := NewTransfer(10, sender.PubKey, receiver.PubKey, sender.Nonce)
	if _, err := transfer.Sign(senderPriv, h); err != nil {
		t.Fatalf("sign: %v", err)
	}

	// Tamper with the amount after signing.
	transfer.Amount.SetUint64(1_000_000)

	ok, err := transfer.Verify(h)
	if ok {
		t.Fatal("tampered transfer should not verify")
	}
	if err == nil {
		t.Fatal("expected an error for tampered transfer")
	}
}
