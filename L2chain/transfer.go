/*
Copyright © 2020 ConsenSys

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"hash"

	eddsa "github.com/consensys/gnark/crypto/signature/eddsa/bn256"
	"github.com/consensys/gurvy/bn256/fr"
)

// Transfer describe a rollup transfer
type Transfer struct {
	nonce          uint64
	amount         fr.Element
	senderPubKey   eddsa.PublicKey
	receiverPubKey eddsa.PublicKey
	signature      eddsa.Signature // signature of the sender's account
}

// NewTransfer creates a new transfer (to be signed)
func NewTransfer(amount uint64, from, to eddsa.PublicKey, nonce uint64) Transfer {

	var res Transfer

	res.nonce = nonce
	res.amount.SetUint64(amount)
	res.senderPubKey = from
	res.receiverPubKey = to

	return res
}

// Sign signs a transaction
func (t *Transfer) Sign(priv eddsa.PrivateKey, h hash.Hash) (eddsa.Signature, error) {

	h.Reset()
	var frNonce, msg fr.Element

	// serializing transfer. The signature is on h(nonce || amount || senderpubKey (x&y) || receiverPubkey(x&y))
	// (each pubkey consist of 2 chunks of 256bits)
	frNonce.SetUint64(t.nonce)
	h.Write(frNonce.Bytes())
	h.Write(t.amount.Bytes())
	h.Write(t.senderPubKey.A.X.Bytes())
	h.Write(t.senderPubKey.A.Y.Bytes())
	h.Write(t.receiverPubKey.A.X.Bytes())
	h.Write(t.receiverPubKey.A.Y.Bytes())
	bmsg := h.Sum([]byte{})
	msg.SetBytes(bmsg)

	sig, err := eddsa.Sign(msg, t.senderPubKey, priv)
	if err != nil {
		return sig, err
	}
	t.signature = sig
	return sig, nil
}

// Verify verifies the signature of the transfer.
// The message to sign is the hash (o.h) of the account.
func (t *Transfer) Verify(h hash.Hash) (bool, error) {

	h.Reset()
	var frNonce, msg fr.Element

	// serializing transfer. The msg to sign is
	// nonce || amount || senderpubKey(x&y) || receiverPubkey(x&y)
	// (each pubkey consist of 2 chunks of 256bits)
	frNonce.SetUint64(t.nonce)
	h.Write(frNonce.Bytes())
	h.Write(t.amount.Bytes())
	h.Write(t.senderPubKey.A.X.Bytes())
	h.Write(t.senderPubKey.A.Y.Bytes())
	h.Write(t.receiverPubKey.A.X.Bytes())
	h.Write(t.receiverPubKey.A.Y.Bytes())
	bmsg := h.Sum([]byte{})
	msg.SetBytes(bmsg)

	// verification of the signature
	resSig, err := eddsa.Verify(t.signature, msg, t.senderPubKey)
	if err != nil {
		return false, err
	}
	if !resSig {
		return false, ErrWrongSignature
	}
	return true, nil
}
