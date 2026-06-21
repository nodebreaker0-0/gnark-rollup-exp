package rollup

import (
	"errors"
	"hash"

	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"github.com/consensys/gnark-crypto/ecc/bn254/twistededwards/eddsa"
)

// ErrWrongSignature is returned when a transfer's signature does not verify.
var ErrWrongSignature = errors.New("rollup: invalid transfer signature")

// Transfer is a signed value transfer between two accounts. The signature is
// over the MiMC hash of (nonce || amount || senderPubKey || receiverPubKey).
type Transfer struct {
	Nonce          uint64
	Amount         fr.Element
	SenderPubKey   eddsa.PublicKey
	ReceiverPubKey eddsa.PublicKey

	// Signature holds the parsed signature and its raw serialized bytes. The raw
	// bytes are kept for assigning the in-circuit witness.
	Signature    eddsa.Signature
	SignatureRaw []byte
}

// NewTransfer creates an unsigned transfer.
func NewTransfer(amount uint64, from, to eddsa.PublicKey, nonce uint64) Transfer {
	var t Transfer
	t.Nonce = nonce
	t.Amount.SetUint64(amount)
	t.SenderPubKey = from
	t.ReceiverPubKey = to
	return t
}

// preimage returns the message that is signed/verified: the MiMC hash of
// nonce || amount || senderX || senderY || receiverX || receiverY.
func (t *Transfer) preimage(h hash.Hash) []byte {
	h.Reset()

	var n fr.Element
	n.SetUint64(t.Nonce)
	nb := n.Bytes()
	h.Write(nb[:])

	ab := t.Amount.Bytes()
	h.Write(ab[:])

	sx := t.SenderPubKey.A.X.Bytes()
	h.Write(sx[:])
	sy := t.SenderPubKey.A.Y.Bytes()
	h.Write(sy[:])

	rx := t.ReceiverPubKey.A.X.Bytes()
	h.Write(rx[:])
	ry := t.ReceiverPubKey.A.Y.Bytes()
	h.Write(ry[:])

	return h.Sum(nil)
}

// Sign signs the transfer with priv and stores the signature on the transfer. It
// returns the raw signature bytes. h must be the MiMC hasher used everywhere in
// the rollup.
func (t *Transfer) Sign(priv eddsa.PrivateKey, h hash.Hash) ([]byte, error) {
	msg := t.preimage(h)

	sigBytes, err := priv.Sign(msg, h)
	if err != nil {
		return nil, err
	}
	if _, err := t.Signature.SetBytes(sigBytes); err != nil {
		return nil, err
	}
	t.SignatureRaw = sigBytes
	return sigBytes, nil
}

// Verify checks the transfer's signature against its sender public key.
func (t *Transfer) Verify(h hash.Hash) (bool, error) {
	msg := t.preimage(h)

	ok, err := t.SenderPubKey.Verify(t.SignatureRaw, msg, h)
	if err != nil {
		return false, err
	}
	if !ok {
		return false, ErrWrongSignature
	}
	return true, nil
}
