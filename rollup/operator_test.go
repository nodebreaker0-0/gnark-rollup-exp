package rollup

import (
	"math/rand"
	"testing"

	"github.com/consensys/gnark-crypto/accumulator/merkletree"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	cmimc "github.com/consensys/gnark-crypto/ecc/bn254/fr/mimc"
	"github.com/consensys/gnark-crypto/ecc/bn254/twistededwards/eddsa"
)

func newElem(v uint64) fr.Element {
	var e fr.Element
	e.SetUint64(v)
	return e
}

// newTestOperator builds an operator with n deterministic accounts, account i
// having balance 20+i. Returns the operator and the private keys.
func newTestOperator(t *testing.T, n int) (Operator, []eddsa.PrivateKey) {
	t.Helper()
	r := rand.New(rand.NewSource(7)) //#nosec G404 -- deterministic test
	op := NewOperator(n, cmimc.NewMiMC())
	privs := make([]eddsa.PrivateKey, n)
	for i := 0; i < n; i++ {
		acc, priv, err := NewAccount(i, uint64(20+i), r)
		if err != nil {
			t.Fatalf("account %d: %v", i, err)
		}
		op.AddAccount(acc)
		privs[i] = priv
	}
	return op, privs
}

func TestApplyTransferUpdatesBalances(t *testing.T) {
	op, privs := newTestOperator(t, 16)

	sender, _ := op.ReadAccount(0)
	receiver, _ := op.ReadAccount(1)
	wantSender := sender.Balance
	wantReceiver := receiver.Balance

	transfer := NewTransfer(5, sender.PubKey, receiver.PubKey, sender.Nonce)
	if _, err := transfer.Sign(privs[0], cmimc.NewMiMC()); err != nil {
		t.Fatalf("sign: %v", err)
	}

	w, err := op.ApplyTransfer(transfer)
	if err != nil {
		t.Fatalf("apply transfer: %v", err)
	}

	// sender balance decreased by 5, nonce incremented
	five, expSender, expReceiver := newElem(5), wantSender, wantReceiver
	expSender.Sub(&expSender, &five)
	expReceiver.Add(&expReceiver, &five)
	if !w.SenderAfter.Balance.Equal(&expSender) {
		t.Fatal("sender balance not decremented correctly")
	}
	if !w.ReceiverAfter.Balance.Equal(&expReceiver) {
		t.Fatal("receiver balance not incremented correctly")
	}
	if w.SenderAfter.Nonce != w.SenderBefore.Nonce+1 {
		t.Fatal("sender nonce not incremented")
	}
}

func TestApplyTransferProofsVerifyNatively(t *testing.T) {
	const n = 16
	op, privs := newTestOperator(t, n)
	h := cmimc.NewMiMC()

	sender, _ := op.ReadAccount(0)
	receiver, _ := op.ReadAccount(1)
	transfer := NewTransfer(3, sender.PubKey, receiver.PubKey, sender.Nonce)
	if _, err := transfer.Sign(privs[0], h); err != nil {
		t.Fatalf("sign: %v", err)
	}

	w, err := op.ApplyTransfer(transfer)
	if err != nil {
		t.Fatalf("apply: %v", err)
	}

	check := func(name string, p MerkleProofData) {
		if !merkletree.VerifyProof(h, p.RootHash, p.Path, p.Index, n) {
			t.Fatalf("%s merkle proof failed native verification", name)
		}
	}
	check("sender-before", w.SenderProofBefore)
	check("receiver-before", w.ReceiverProofBefore)
	check("sender-after", w.SenderProofAfter)
	check("receiver-after", w.ReceiverProofAfter)
}

func TestApplyTransferRejectsBadNonce(t *testing.T) {
	op, privs := newTestOperator(t, 16)
	sender, _ := op.ReadAccount(0)
	receiver, _ := op.ReadAccount(1)

	transfer := NewTransfer(1, sender.PubKey, receiver.PubKey, sender.Nonce+9)
	if _, err := transfer.Sign(privs[0], cmimc.NewMiMC()); err != nil {
		t.Fatalf("sign: %v", err)
	}
	if _, err := op.ApplyTransfer(transfer); err != ErrNonce {
		t.Fatalf("expected ErrNonce, got %v", err)
	}
}

func TestApplyTransferRejectsOverdraft(t *testing.T) {
	op, privs := newTestOperator(t, 16)
	sender, _ := op.ReadAccount(0) // balance 20
	receiver, _ := op.ReadAccount(1)

	transfer := NewTransfer(1_000, sender.PubKey, receiver.PubKey, sender.Nonce)
	if _, err := transfer.Sign(privs[0], cmimc.NewMiMC()); err != nil {
		t.Fatalf("sign: %v", err)
	}
	if _, err := op.ApplyTransfer(transfer); err != ErrAmountTooHigh {
		t.Fatalf("expected ErrAmountTooHigh, got %v", err)
	}
}
