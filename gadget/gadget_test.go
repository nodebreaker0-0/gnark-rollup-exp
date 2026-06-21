package gadget

import (
	"bytes"
	stdhash "hash"
	"math/rand"
	"testing"

	"github.com/consensys/gnark-crypto/accumulator/merkletree"
	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	cmimc "github.com/consensys/gnark-crypto/ecc/bn254/fr/mimc"
	ceddsa "github.com/consensys/gnark-crypto/ecc/bn254/twistededwards/eddsa"
	tedwards "github.com/consensys/gnark-crypto/ecc/twistededwards"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/std/accumulator/merkle"
	"github.com/consensys/gnark/std/hash/mimc"
	"github.com/consensys/gnark/test"
)

// membershipCircuit exercises the gadget in isolation.
type membershipCircuit struct {
	Account Account
	Proof   merkle.MerkleProof
	Root    frontend.Variable `gnark:",public"`
}

func (c *membershipCircuit) Define(api frontend.API) error {
	h, err := mimc.NewMiMC(api)
	if err != nil {
		return err
	}
	VerifyMembership(api, &h, c.Account, c.Proof, c.Root)
	return nil
}

// nativeAccount mirrors gadget.Account off-circuit, committing the same way.
type nativeAccount struct {
	index, nonce uint64
	balance      fr.Element
	pub          ceddsa.PublicKey
}

func (a nativeAccount) commit(h stdhash.Hash) []byte {
	h.Reset()
	var idx, non fr.Element
	idx.SetUint64(a.index)
	non.SetUint64(a.nonce)
	for _, e := range []fr.Element{idx, non, a.balance, a.pub.A.X, a.pub.A.Y} {
		b := e.Bytes()
		h.Write(b[:])
	}
	return h.Sum(nil)
}

func TestVerifyMembership(t *testing.T) {
	const nbLeaves = 16
	const target = 3
	r := rand.New(rand.NewSource(5)) //#nosec G404 -- deterministic test
	h := cmimc.NewMiMC()

	accounts := make([]nativeAccount, nbLeaves)
	var buf bytes.Buffer
	for i := 0; i < nbLeaves; i++ {
		priv, err := ceddsa.GenerateKey(r)
		if err != nil {
			t.Fatalf("key %d: %v", i, err)
		}
		a := nativeAccount{index: uint64(i), nonce: uint64(i), pub: priv.PublicKey}
		a.balance.SetUint64(uint64(100 + i))
		accounts[i] = a
		buf.Write(a.commit(h)) // 32-byte leaf per account
	}

	root, proofPath, numLeaves, err := merkletree.BuildReaderProof(&buf, h, h.Size(), target)
	if err != nil {
		t.Fatalf("build proof: %v", err)
	}
	if !merkletree.VerifyProof(h, root, proofPath, target, numLeaves) {
		t.Fatal("native proof must verify")
	}

	circuit := &membershipCircuit{Proof: merkle.MerkleProof{Path: make([]frontend.Variable, len(proofPath))}}

	assignment := &membershipCircuit{Root: root}
	assignment.Account.Index = accounts[target].index
	assignment.Account.Nonce = accounts[target].nonce
	assignment.Account.Balance = accounts[target].balance
	assignment.Account.PubKey.Assign(tedwards.BN254, accounts[target].pub.Bytes())
	assignment.Proof.RootHash = root
	assignment.Proof.Path = make([]frontend.Variable, len(proofPath))
	for i := range proofPath {
		assignment.Proof.Path[i] = proofPath[i]
	}

	if err := test.IsSolved(circuit, assignment, ecc.BN254.ScalarField()); err != nil {
		t.Fatalf("gadget should verify a valid membership proof: %v", err)
	}

	// Tampering with the balance breaks the commitment -> leaf mismatch.
	bad := *assignment
	bad.Account.Balance = fr.NewElement(999999)
	if err := test.IsSolved(circuit, &bad, ecc.BN254.ScalarField()); err == nil {
		t.Fatal("gadget accepted a tampered account, but should not")
	}
}
