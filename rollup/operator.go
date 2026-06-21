package rollup

import (
	"bytes"
	"errors"
	"hash"

	"github.com/consensys/gnark-crypto/accumulator/merkletree"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
)

// Errors returned while applying a transfer.
var (
	ErrNonExistingAccount = errors.New("rollup: account is not in the state")
	ErrAmountTooHigh      = errors.New("rollup: transfer amount exceeds sender balance")
	ErrNonce              = errors.New("rollup: transfer nonce does not match sender nonce")
)

// MerkleProofData is a native Merkle inclusion proof for one leaf. Path[0] is the
// leaf value (the account hash); Path[1:] are the sibling hashes from leaf to
// root. Index is the leaf position, used in-circuit to order the hashing.
type MerkleProofData struct {
	RootHash []byte
	Path     [][]byte
	Index    uint64
}

// TransferWitness is everything the circuit needs to verify one applied
// transfer: the state roots before and after, both parties' account snapshots
// and inclusion proofs in each state, and the transfer itself (amount, parsed
// public keys, raw signature). The circuit's Assign consumes this directly.
type TransferWitness struct {
	RootBefore []byte
	RootAfter  []byte

	SenderBefore   Account
	SenderAfter    Account
	ReceiverBefore Account
	ReceiverAfter  Account

	SenderProofBefore   MerkleProofData
	SenderProofAfter    MerkleProofData
	ReceiverProofBefore MerkleProofData
	ReceiverProofAfter  MerkleProofData

	Amount            fr.Element
	SenderPubKeyRaw   []byte
	ReceiverPubKeyRaw []byte
	SignatureRaw      []byte
}

// Operator maintains rollup state: the serialized accounts, their hashed leaves
// (the Merkle tree input), and an index from public key to position.
type Operator struct {
	State      []byte            // concatenated serialized accounts
	HashState  []byte            // concatenated account-hash leaves, h.Size() bytes each
	AccountMap map[string]uint64 // pubKey.X bytes -> account index
	nbAccounts int
	h          hash.Hash // MiMC hasher
}

// NewOperator creates an operator managing nbAccounts empty slots, using h as the
// Merkle/leaf hasher (a native MiMC instance).
func NewOperator(nbAccounts int, h hash.Hash) Operator {
	o := Operator{
		State:      make([]byte, SizeAccount*nbAccounts),
		HashState:  make([]byte, h.Size()*nbAccounts),
		AccountMap: make(map[string]uint64),
		nbAccounts: nbAccounts,
		h:          h,
	}
	// initialize leaf hashes for the empty accounts
	for i := 0; i < nbAccounts; i++ {
		h.Reset()
		h.Write(o.State[i*SizeAccount : (i+1)*SizeAccount])
		copy(o.HashState[i*h.Size():(i+1)*h.Size()], h.Sum(nil))
	}
	return o
}

// AddAccount writes acc into the operator's state at acc.Index and indexes it.
func (o *Operator) AddAccount(acc Account) {
	o.AccountMap[string(acc.PubKey.A.X.Marshal())] = acc.Index
	o.writeAccount(acc)
}

// writeAccount serializes acc into State and refreshes its leaf in HashState.
func (o *Operator) writeAccount(acc Account) {
	pos := int(acc.Index)
	copy(o.State[pos*SizeAccount:], acc.Serialize())
	o.h.Reset()
	o.h.Write(acc.Serialize())
	copy(o.HashState[pos*o.h.Size():(pos+1)*o.h.Size()], o.h.Sum(nil))
}

// ReadAccount returns the account stored at index i.
func (o *Operator) ReadAccount(i uint64) (Account, error) {
	if int(i) >= o.nbAccounts {
		return Account{}, ErrNonExistingAccount
	}
	return Deserialize(o.State[int(i)*SizeAccount : int(i)*SizeAccount+SizeAccount])
}

// proof builds a native Merkle inclusion proof for the leaf at index pos against
// the current HashState.
func (o *Operator) proof(pos uint64) (MerkleProofData, error) {
	root, path, _, err := merkletree.BuildReaderProof(bytes.NewReader(o.HashState), o.h, o.h.Size(), pos)
	if err != nil {
		return MerkleProofData{}, err
	}
	return MerkleProofData{RootHash: root, Path: path, Index: pos}, nil
}

// ApplyTransfer validates t against current state, applies it, and returns a
// TransferWitness capturing the before/after state and proofs. The transfer must
// already be signed. It mutates operator state on success.
func (o *Operator) ApplyTransfer(t Transfer) (TransferWitness, error) {
	var w TransferWitness

	posSender, ok := o.AccountMap[string(t.SenderPubKey.A.X.Marshal())]
	if !ok {
		return w, ErrNonExistingAccount
	}
	posReceiver, ok := o.AccountMap[string(t.ReceiverPubKey.A.X.Marshal())]
	if !ok {
		return w, ErrNonExistingAccount
	}

	senderBefore, err := o.ReadAccount(posSender)
	if err != nil {
		return w, err
	}
	receiverBefore, err := o.ReadAccount(posReceiver)
	if err != nil {
		return w, err
	}

	// validate the transfer
	ok, err = t.Verify(o.h)
	if err != nil || !ok {
		return w, ErrWrongSignature
	}
	if t.Amount.Cmp(&senderBefore.Balance) > 0 {
		return w, ErrAmountTooHigh
	}
	if t.Nonce != senderBefore.Nonce {
		return w, ErrNonce
	}

	// capture "before" roots and proofs (pre-update state)
	w.RootBefore = nil
	sproofBefore, err := o.proof(posSender)
	if err != nil {
		return w, err
	}
	rproofBefore, err := o.proof(posReceiver)
	if err != nil {
		return w, err
	}
	w.RootBefore = sproofBefore.RootHash
	w.SenderProofBefore = sproofBefore
	w.ReceiverProofBefore = rproofBefore
	w.SenderBefore = senderBefore
	w.ReceiverBefore = receiverBefore

	// apply the transfer
	senderAfter := senderBefore
	receiverAfter := receiverBefore
	senderAfter.Balance.Sub(&senderBefore.Balance, &t.Amount)
	receiverAfter.Balance.Add(&receiverBefore.Balance, &t.Amount)
	senderAfter.Nonce = senderBefore.Nonce + 1

	o.writeAccount(senderAfter)
	o.writeAccount(receiverAfter)

	// capture "after" roots and proofs (post-update state)
	sproofAfter, err := o.proof(posSender)
	if err != nil {
		return w, err
	}
	rproofAfter, err := o.proof(posReceiver)
	if err != nil {
		return w, err
	}
	w.RootAfter = sproofAfter.RootHash
	w.SenderProofAfter = sproofAfter
	w.ReceiverProofAfter = rproofAfter
	w.SenderAfter = senderAfter
	w.ReceiverAfter = receiverAfter

	w.Amount = t.Amount
	w.SenderPubKeyRaw = t.SenderPubKey.Bytes()
	w.ReceiverPubKeyRaw = t.ReceiverPubKey.Bytes()
	w.SignatureRaw = t.SignatureRaw

	return w, nil
}
