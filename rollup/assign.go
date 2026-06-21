package rollup

import (
	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	tedwards "github.com/consensys/gnark-crypto/ecc/twistededwards"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/std/accumulator/merkle"
)

// toElem converts a uint64 into a field element for witness assignment.
func toElem(v uint64) fr.Element {
	var e fr.Element
	e.SetUint64(v)
	return e
}

// Assign builds a fully populated Circuit assignment from a batch of applied
// transfers (as produced by Operator.ApplyTransfer). pathLen must match the
// Merkle path length of the proofs (tree depth + 1). The returned value is ready
// to pass to the prover alongside a circuit built with New(len(witnesses), pathLen).
func Assign(witnesses []TransferWitness, pathLen int) *Circuit {
	c := New(len(witnesses), pathLen)

	for i := range witnesses {
		w := witnesses[i]

		c.RootsBefore[i] = w.RootBefore
		c.RootsAfter[i] = w.RootAfter

		assignAccount(&c.SenderBefore[i], w.SenderBefore)
		assignAccount(&c.SenderAfter[i], w.SenderAfter)
		assignAccount(&c.ReceiverBefore[i], w.ReceiverBefore)
		assignAccount(&c.ReceiverAfter[i], w.ReceiverAfter)

		assignProof(&c.ProofSenderBefore[i], w.SenderProofBefore)
		assignProof(&c.ProofSenderAfter[i], w.SenderProofAfter)
		assignProof(&c.ProofReceiverBefore[i], w.ReceiverProofBefore)
		assignProof(&c.ProofReceiverAfter[i], w.ReceiverProofAfter)

		c.Transfers[i].Amount = w.Amount
		c.Transfers[i].Nonce = toElem(w.SenderBefore.Nonce)
		c.Transfers[i].Sender.Assign(tedwards.BN254, w.SenderPubKeyRaw)
		c.Transfers[i].Receiver.Assign(tedwards.BN254, w.ReceiverPubKeyRaw)
		c.Transfers[i].Signature.Assign(tedwards.BN254, w.SignatureRaw)
	}
	return c
}

func assignAccount(dst *AccountConstraints, acc Account) {
	dst.Index = toElem(acc.Index)
	dst.Nonce = toElem(acc.Nonce)
	dst.Balance = acc.Balance
	dst.PubKey.Assign(tedwards.BN254, acc.PubKey.Bytes())
}

func assignProof(dst *merkle.MerkleProof, p MerkleProofData) {
	dst.RootHash = p.RootHash
	dst.Path = make([]frontend.Variable, len(p.Path))
	for j := range p.Path {
		dst.Path[j] = p.Path[j]
	}
}
