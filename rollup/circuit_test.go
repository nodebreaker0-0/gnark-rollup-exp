package rollup

import (
	"math/rand"
	"testing"

	"github.com/consensys/gnark-crypto/ecc"
	cmimc "github.com/consensys/gnark-crypto/ecc/bn254/fr/mimc"
	"github.com/consensys/gnark-crypto/ecc/bn254/twistededwards/eddsa"
	"github.com/consensys/gnark/test"
	"github.com/nodebreaker0-0/gnark-rollup-exp/prove"
)

// buildBatch applies `count` sequential transfers from account 0 to account 1
// and returns the witnesses plus the Merkle path length.
func buildBatch(t testing.TB, nbAccounts, count int) ([]TransferWitness, int) {
	t.Helper()
	r := rand.New(rand.NewSource(99)) //#nosec G404 -- deterministic test
	op := NewOperator(nbAccounts, cmimc.NewMiMC())

	privs := make([]eddsa.PrivateKey, nbAccounts)
	for i := 0; i < nbAccounts; i++ {
		acc, priv, err := NewAccount(i, uint64(100+i), r)
		if err != nil {
			t.Fatalf("account %d: %v", i, err)
		}
		op.AddAccount(acc)
		privs[i] = priv
	}

	witnesses := make([]TransferWitness, count)
	for k := 0; k < count; k++ {
		sender, _ := op.ReadAccount(0)
		receiver, _ := op.ReadAccount(1)
		transfer := NewTransfer(1, sender.PubKey, receiver.PubKey, sender.Nonce)
		if _, err := transfer.Sign(privs[0], cmimc.NewMiMC()); err != nil {
			t.Fatalf("sign %d: %v", k, err)
		}
		w, err := op.ApplyTransfer(transfer)
		if err != nil {
			t.Fatalf("apply %d: %v", k, err)
		}
		witnesses[k] = w
	}
	return witnesses, len(witnesses[0].SenderProofBefore.Path)
}

func TestCircuitSolvesBatch1(t *testing.T) {
	witnesses, pathLen := buildBatch(t, 16, 1)
	circuit := New(len(witnesses), pathLen)
	assignment := Assign(witnesses, pathLen)

	if err := test.IsSolved(circuit, assignment, ecc.BN254.ScalarField()); err != nil {
		t.Fatalf("circuit should solve for a valid batch: %v", err)
	}
}

func TestCircuitProveVerifyBatch1(t *testing.T) {
	witnesses, pathLen := buildBatch(t, 16, 1)
	circuit := New(len(witnesses), pathLen)
	assignment := Assign(witnesses, pathLen)

	if err := prove.Run(circuit, assignment); err != nil {
		t.Fatalf("expected end-to-end proof to verify: %v", err)
	}
}

func TestCircuitRejectsTamperedRoot(t *testing.T) {
	witnesses, pathLen := buildBatch(t, 16, 1)
	circuit := New(len(witnesses), pathLen)
	assignment := Assign(witnesses, pathLen)

	// Corrupt the public "after" root: the proof should no longer solve.
	assignment.RootsAfter[0] = toElem(123456789)

	if err := test.IsSolved(circuit, assignment, ecc.BN254.ScalarField()); err == nil {
		t.Fatal("circuit solved with a tampered root, but should not")
	}
}

func TestCircuitProveVerifyBatch3(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping multi-transfer proving in -short mode")
	}
	witnesses, pathLen := buildBatch(t, 16, 3)
	circuit := New(len(witnesses), pathLen)
	assignment := Assign(witnesses, pathLen)

	if err := prove.Run(circuit, assignment); err != nil {
		t.Fatalf("expected end-to-end proof to verify for batch of 3: %v", err)
	}
}

// TestCircuitProveVerifyPLONK cross-checks that the same rollup circuit proves
// and verifies under the PLONK backend, not just Groth16.
func TestCircuitProveVerifyPLONK(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping PLONK proving in -short mode")
	}
	witnesses, pathLen := buildBatch(t, 16, 1)
	circuit := New(len(witnesses), pathLen)
	assignment := Assign(witnesses, pathLen)

	if err := prove.RunPLONK(circuit, assignment); err != nil {
		t.Fatalf("expected end-to-end PLONK proof to verify: %v", err)
	}
}

// TestCircuitProveVerifyMultiPair proves a batch with two distinct sender/receiver
// pairs (0->1 and 2->3), showing the operator and circuit are not specialized to
// a single pair.
func TestCircuitProveVerifyMultiPair(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping multi-pair proving in -short mode")
	}
	const n = 16
	r := rand.New(rand.NewSource(123)) //#nosec G404 -- deterministic test
	op := NewOperator(n, cmimc.NewMiMC())
	privs := make([]eddsa.PrivateKey, n)
	for i := 0; i < n; i++ {
		acc, priv, err := NewAccount(i, uint64(200+i), r)
		if err != nil {
			t.Fatalf("account %d: %v", i, err)
		}
		op.AddAccount(acc)
		privs[i] = priv
	}

	pairs := [][2]int{{0, 1}, {2, 3}}
	witnesses := make([]TransferWitness, 0, len(pairs))
	for _, p := range pairs {
		sender, _ := op.ReadAccount(uint64(p[0]))
		receiver, _ := op.ReadAccount(uint64(p[1]))
		transfer := NewTransfer(2, sender.PubKey, receiver.PubKey, sender.Nonce)
		if _, err := transfer.Sign(privs[p[0]], cmimc.NewMiMC()); err != nil {
			t.Fatalf("sign %v: %v", p, err)
		}
		w, err := op.ApplyTransfer(transfer)
		if err != nil {
			t.Fatalf("apply %v: %v", p, err)
		}
		witnesses = append(witnesses, w)
	}

	pathLen := len(witnesses[0].SenderProofBefore.Path)
	circuit := New(len(witnesses), pathLen)
	assignment := Assign(witnesses, pathLen)
	if err := prove.Run(circuit, assignment); err != nil {
		t.Fatalf("expected multi-pair batch proof to verify: %v", err)
	}
}
