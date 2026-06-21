// Package gadget holds reusable in-circuit building blocks for account-based
// state machines: an account commitment and a Merkle-membership check that binds
// an account to a tree root. The rollup circuit is built from these; they are
// exported so other circuits can reuse them directly.
package gadget

import (
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/std/accumulator/merkle"
	"github.com/consensys/gnark/std/hash"
	"github.com/consensys/gnark/std/signature/eddsa"
)

// Account is the in-circuit representation of an account. Its commitment is the
// value stored at the account's Merkle leaf.
type Account struct {
	Index   frontend.Variable
	Nonce   frontend.Variable
	Balance frontend.Variable
	PubKey  eddsa.PublicKey
}

// Commit returns the MiMC commitment H(index || nonce || balance || pubKeyX ||
// pubKeyY). The hasher is reset before use.
func (a Account) Commit(h hash.FieldHasher) frontend.Variable {
	h.Reset()
	h.Write(a.Index, a.Nonce, a.Balance, a.PubKey.A.X, a.PubKey.A.Y)
	return h.Sum()
}

// VerifyMembership asserts that a is committed at index a.Index in the Merkle
// tree whose root is root, using proof. It (1) binds the account commitment to
// the proof's leaf (proof.Path[0]), (2) binds the proof to the public root, and
// (3) checks the inclusion proof.
func VerifyMembership(api frontend.API, h hash.FieldHasher, a Account, proof merkle.MerkleProof, root frontend.Variable) {
	api.AssertIsEqual(a.Commit(h), proof.Path[0])
	api.AssertIsEqual(proof.RootHash, root)
	proof.VerifyProof(api, h, a.Index)
}
