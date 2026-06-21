// Package eddsa verifies an EdDSA signature inside a circuit: prove that a
// signature over a message is valid under a public key, all as constraints.
//
// This is the building block the rollup uses to check that each transfer was
// authorized by the sender, without trusting the prover.
package eddsa

import (
	tedwards "github.com/consensys/gnark-crypto/ecc/twistededwards"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/std/algebra/native/twistededwards"
	"github.com/consensys/gnark/std/hash/mimc"
	"github.com/consensys/gnark/std/signature/eddsa"
)

// Circuit verifies that Signature is a valid EdDSA signature of Message under
// PublicKey, using MiMC as the hash. All three are public inputs.
//
// curveID selects the embedded twisted Edwards curve and must be set on both the
// circuit passed to compilation and every assignment; use New to get a circuit
// configured for BN254.
type Circuit struct {
	curveID   tedwards.ID
	PublicKey eddsa.PublicKey   `gnark:",public"`
	Signature eddsa.Signature   `gnark:",public"`
	Message   frontend.Variable `gnark:",public"`
}

// New returns a Circuit configured for the BN254 embedded curve.
func New() *Circuit {
	return &Circuit{curveID: tedwards.BN254}
}

// Define verifies the signature within the constraint system.
func (c *Circuit) Define(api frontend.API) error {
	curve, err := twistededwards.NewEdCurve(api, c.curveID)
	if err != nil {
		return err
	}
	h, err := mimc.NewMiMC(api)
	if err != nil {
		return err
	}
	return eddsa.Verify(curve, c.Signature, c.Message, c.PublicKey, &h)
}
