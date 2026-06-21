package mimc

import (
	"testing"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	cmimc "github.com/consensys/gnark-crypto/ecc/bn254/fr/mimc"
	"github.com/consensys/gnark/test"
	"github.com/nodebreaker0-0/gnark-rollup-exp/prove"
)

// hashPreimage computes the native (out-of-circuit) MiMC digest of a field
// element, matching the in-circuit hash.
func hashPreimage(x fr.Element) []byte {
	h := cmimc.NewMiMC()
	b := x.Bytes()
	h.Write(b[:])
	return h.Sum(nil)
}

func TestProveVerify(t *testing.T) {
	var pre fr.Element
	pre.SetUint64(35)
	digest := hashPreimage(pre)

	if err := prove.Run(&Circuit{}, &Circuit{PreImage: pre, Hash: digest}); err != nil {
		t.Fatalf("expected proof to verify: %v", err)
	}
}

func TestConstraints(t *testing.T) {
	field := ecc.BN254.ScalarField()

	var pre fr.Element
	pre.SetUint64(35)
	digest := hashPreimage(pre)

	if err := test.IsSolved(&Circuit{}, &Circuit{PreImage: pre, Hash: digest}, field); err != nil {
		t.Fatalf("valid preimage should solve: %v", err)
	}

	var wrong fr.Element
	wrong.SetUint64(36)
	if err := test.IsSolved(&Circuit{}, &Circuit{PreImage: pre, Hash: hashPreimage(wrong)}, field); err == nil {
		t.Fatal("mismatched digest should not solve, but it did")
	}
}
