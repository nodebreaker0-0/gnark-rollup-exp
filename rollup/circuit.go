package rollup

import (
	tedwards "github.com/consensys/gnark-crypto/ecc/twistededwards"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/std/accumulator/merkle"
	"github.com/consensys/gnark/std/algebra/native/twistededwards"
	"github.com/consensys/gnark/std/hash/mimc"
	"github.com/consensys/gnark/std/signature/eddsa"
	"github.com/nodebreaker0-0/gnark-rollup-exp/gadget"
)

// AccountConstraints is the in-circuit representation of an account. It is an
// alias for gadget.Account, the reusable account-commitment building block.
type AccountConstraints = gadget.Account

// TransferConstraints is the in-circuit representation of a signed transfer.
type TransferConstraints struct {
	Amount    frontend.Variable
	Nonce     frontend.Variable
	Sender    eddsa.PublicKey
	Receiver  eddsa.PublicKey
	Signature eddsa.Signature
}

// Circuit proves that a batch of transfers was applied to the rollup state
// correctly. For each transfer it checks: both accounts are committed in the
// "before" root, both are committed in the "after" root, the transfer signature
// is valid, and balances/nonces are updated according to the rules.
//
// Build one with New(batchSize, pathLen) before compiling, and produce an
// assignment with Assign.
type Circuit struct {
	// public state roots, one pair per transfer in the batch
	RootsBefore []frontend.Variable `gnark:",public"`
	RootsAfter  []frontend.Variable `gnark:",public"`

	// account snapshots before and after, per transfer
	SenderBefore   []AccountConstraints
	SenderAfter    []AccountConstraints
	ReceiverBefore []AccountConstraints
	ReceiverAfter  []AccountConstraints

	// signed transfers
	Transfers []TransferConstraints

	// Merkle inclusion proofs (Path[0] is the account-hash leaf)
	ProofSenderBefore   []merkle.MerkleProof
	ProofSenderAfter    []merkle.MerkleProof
	ProofReceiverBefore []merkle.MerkleProof
	ProofReceiverAfter  []merkle.MerkleProof

	batchSize int
	pathLen   int
}

// New returns a Circuit sized for batchSize transfers with Merkle paths of
// pathLen elements (pathLen = tree depth + 1). Use it both to compile the
// circuit and as the template for assignments.
func New(batchSize, pathLen int) *Circuit {
	c := &Circuit{
		batchSize:           batchSize,
		pathLen:             pathLen,
		RootsBefore:         make([]frontend.Variable, batchSize),
		RootsAfter:          make([]frontend.Variable, batchSize),
		SenderBefore:        make([]AccountConstraints, batchSize),
		SenderAfter:         make([]AccountConstraints, batchSize),
		ReceiverBefore:      make([]AccountConstraints, batchSize),
		ReceiverAfter:       make([]AccountConstraints, batchSize),
		Transfers:           make([]TransferConstraints, batchSize),
		ProofSenderBefore:   make([]merkle.MerkleProof, batchSize),
		ProofSenderAfter:    make([]merkle.MerkleProof, batchSize),
		ProofReceiverBefore: make([]merkle.MerkleProof, batchSize),
		ProofReceiverAfter:  make([]merkle.MerkleProof, batchSize),
	}
	for i := 0; i < batchSize; i++ {
		c.ProofSenderBefore[i].Path = make([]frontend.Variable, pathLen)
		c.ProofSenderAfter[i].Path = make([]frontend.Variable, pathLen)
		c.ProofReceiverBefore[i].Path = make([]frontend.Variable, pathLen)
		c.ProofReceiverAfter[i].Path = make([]frontend.Variable, pathLen)
	}
	return c
}

// Define encodes the rollup constraints.
func (c *Circuit) Define(api frontend.API) error {
	curve, err := twistededwards.NewEdCurve(api, tedwards.BN254)
	if err != nil {
		return err
	}
	h, err := mimc.NewMiMC(api)
	if err != nil {
		return err
	}

	for i := 0; i < c.batchSize; i++ {
		// 1+2. each account is committed at its index in the before/after roots
		// (the gadget binds the account commitment to the proof leaf, the proof
		// to the public root, and checks inclusion).
		gadget.VerifyMembership(api, &h, c.SenderBefore[i], c.ProofSenderBefore[i], c.RootsBefore[i])
		gadget.VerifyMembership(api, &h, c.ReceiverBefore[i], c.ProofReceiverBefore[i], c.RootsBefore[i])
		gadget.VerifyMembership(api, &h, c.SenderAfter[i], c.ProofSenderAfter[i], c.RootsAfter[i])
		gadget.VerifyMembership(api, &h, c.ReceiverAfter[i], c.ProofReceiverAfter[i], c.RootsAfter[i])

		// 3. the transfer is bound to the sender and signed by them
		bindTransfer(api, c.Transfers[i], c.SenderBefore[i], c.ReceiverBefore[i])
		if err := verifySignature(api, curve, &h, c.Transfers[i]); err != nil {
			return err
		}

		// 4. balances and nonce update correctly
		verifyUpdate(api, c.SenderBefore[i], c.ReceiverBefore[i], c.SenderAfter[i], c.ReceiverAfter[i], c.Transfers[i].Amount)
	}
	return nil
}

// bindTransfer asserts the transfer's keys/nonce match the sender/receiver
// accounts, so the prover cannot sign for one account and update another.
func bindTransfer(api frontend.API, t TransferConstraints, sender, receiver AccountConstraints) {
	api.AssertIsEqual(t.Sender.A.X, sender.PubKey.A.X)
	api.AssertIsEqual(t.Sender.A.Y, sender.PubKey.A.Y)
	api.AssertIsEqual(t.Receiver.A.X, receiver.PubKey.A.X)
	api.AssertIsEqual(t.Receiver.A.Y, receiver.PubKey.A.Y)
	api.AssertIsEqual(t.Nonce, sender.Nonce)
}

// verifySignature checks the EdDSA signature over the MiMC hash of the transfer
// fields (matching Transfer.preimage on the native side).
func verifySignature(api frontend.API, curve twistededwards.Curve, h *mimc.MiMC, t TransferConstraints) error {
	h.Reset()
	h.Write(t.Nonce, t.Amount, t.Sender.A.X, t.Sender.A.Y, t.Receiver.A.X, t.Receiver.A.Y)
	msg := h.Sum()

	// eddsa.Verify (via IsValid) hashes H(R,A,msg) without resetting first, so the
	// hasher must be in its initial state here. MiMC.Sum clears buffered data but
	// keeps the chaining value, so an explicit Reset is required.
	h.Reset()
	return eddsa.Verify(curve, t.Signature, msg, t.Sender, h)
}

// verifyUpdate asserts the state transition: nonce+1, amount <= balance, and the
// balances move by exactly amount. Public keys and index are unchanged.
func verifyUpdate(api frontend.API, senderBefore, receiverBefore, senderAfter, receiverAfter AccountConstraints, amount frontend.Variable) {
	api.AssertIsEqual(api.Add(senderBefore.Nonce, 1), senderAfter.Nonce)

	api.AssertIsLessOrEqual(amount, senderBefore.Balance)
	api.AssertIsEqual(api.Sub(senderBefore.Balance, amount), senderAfter.Balance)
	api.AssertIsEqual(api.Add(receiverBefore.Balance, amount), receiverAfter.Balance)

	// identity is preserved
	api.AssertIsEqual(senderBefore.Index, senderAfter.Index)
	api.AssertIsEqual(senderBefore.PubKey.A.X, senderAfter.PubKey.A.X)
	api.AssertIsEqual(senderBefore.PubKey.A.Y, senderAfter.PubKey.A.Y)
	api.AssertIsEqual(receiverBefore.Index, receiverAfter.Index)
	api.AssertIsEqual(receiverBefore.PubKey.A.X, receiverAfter.PubKey.A.X)
	api.AssertIsEqual(receiverBefore.PubKey.A.Y, receiverAfter.PubKey.A.Y)
}
