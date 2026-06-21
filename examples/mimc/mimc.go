// Package mimc proves knowledge of a MiMC hash preimage: given a public digest,
// prove you know a secret value that hashes to it, without revealing the value.
//
// MiMC is the hash used throughout the rollup reference, because it is cheap to
// evaluate inside an arithmetic circuit.
package mimc

import (
	"github.com/consensys/gnark/frontend"
	gmimc "github.com/consensys/gnark/std/hash/mimc"
)

// Circuit proves MiMC(PreImage) == Hash. PreImage is secret; Hash is public.
type Circuit struct {
	PreImage frontend.Variable // secret preimage
	Hash     frontend.Variable `gnark:",public"` // public MiMC digest
}

// Define constrains the MiMC hash of PreImage to equal the public Hash.
func (c *Circuit) Define(api frontend.API) error {
	h, err := gmimc.NewMiMC(api)
	if err != nil {
		return err
	}
	h.Write(c.PreImage)
	api.AssertIsEqual(c.Hash, h.Sum())
	return nil
}
